package api

import (
	"math/rand"
	"testing"
	"time"

	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/util"
	"github.com/guregu/null/v6"
)

const n = 300

func BenchmarkNewListShoppingCartItems_Old(b *testing.B) {
	shopItems, prodItems, sizeItems := generateTestData(n, n, n) // adjust size as needed

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = newlistShoppingCartItemsResponse_Old(shopItems, prodItems, sizeItems)
	}
}

func BenchmarkNewListShoppingCartItems_Optimized(b *testing.B) {
	shopItems, prodItems, sizeItems := generateTestData(n, n, n)
	// Pre-allocate final slice - NO append
	rsp := make([]*listShoppingCartItemsResponse, len(prodItems))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = newlistShoppingCartItemsResponse_Optimized(shopItems, prodItems, sizeItems, rsp)
	}
}

// ------------------------------------------------------------
// Generate Test Data (same for both benchmarks)
// ------------------------------------------------------------

func generateTestData(nShop, nProd, nSizes int) (
	[]*db.ListShoppingCartItemsByUserIDRow,
	[]*db.ListProductItemsByIDsRow,
	[]*db.ProductSize,
) {
	// Properly seed the RNG
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// --- Shopping cart items ---
	shop := make([]*db.ListShoppingCartItemsByUserIDRow, nShop)
	for i := 0; i < nShop; i++ {
		shop[i] = &db.ListShoppingCartItemsByUserIDRow{
			ID:             null.IntFrom(int64(i + 1)),
			Qty:            null.IntFrom(int64(r.Intn(5) + 1)),
			ProductItemID:  null.IntFrom(int64(i % nProd)), // wrap around product items
			UserID:         int64(r.Intn(1000)),
			ShoppingCartID: null.IntFrom(int64(r.Intn(500))),
			SizeID:         null.IntFrom(int64(i % nSizes)), // wrap around sizes
			CreatedAt:      null.TimeFrom(time.Now().Add(time.Duration(-r.Intn(1000)) * time.Hour)),
			UpdatedAt:      null.TimeFrom(time.Now()),
			SizeQty:        int32(r.Intn(20)),
			SizeValue:      "M",
		}
	}

	// --- Product items ---
	prod := make([]*db.ListProductItemsByIDsRow, nProd)
	for i := 0; i < nProd; i++ {
		prod[i] = &db.ListProductItemsByIDsRow{
			ID:                        int64(i),
			Name:                      null.StringFrom("ProductName_" + util.RandomString(10)),
			ProductID:                 int64(i * 10),
			ColorValue:                null.StringFrom("Red"),
			Price:                     util.RandomDecimalString(0, 1000),
			ProductImage1:             null.StringFrom("img1.png"),
			ProductImage2:             null.StringFrom("ProductName_" + util.RandomString(10)),
			ProductImage3:             null.StringFrom("ProductName_" + util.RandomString(10)),
			Active:                    true,
			QtyInStock:                int64(r.Intn(50)),
			CategoryPromoID:           null.IntFrom(int64(r.Intn(50))),
			CategoryPromoName:         null.StringFrom("ProductName_" + util.RandomString(10)),
			CategoryPromoDescription:  null.StringFrom("ProductName_" + util.RandomString(10)),
			CategoryPromoDiscountRate: null.IntFrom(int64(r.Intn(50))),
			CategoryPromoActive:       util.RandomBool(),
			CategoryPromoStartDate:    null.Time{},
			CategoryPromoEndDate:      null.Time{},
			BrandPromoID:              null.IntFrom(int64(r.Intn(50))),
			BrandPromoName:            null.StringFrom("ProductName_" + util.RandomString(10)),
			BrandPromoDescription:     null.StringFrom("ProductName_" + util.RandomString(10)),
			BrandPromoDiscountRate:    null.IntFrom(int64(r.Intn(50))),
			BrandPromoActive:          util.RandomBool(),
			BrandPromoStartDate:       null.Time{},
			BrandPromoEndDate:         null.Time{},
			ProductPromoID:            null.IntFrom(int64(r.Intn(50))),
			ProductPromoName:          null.StringFrom("ProductName_" + util.RandomString(10)),
			ProductPromoDescription:   null.StringFrom("ProductName_" + util.RandomString(10)),
			ProductPromoDiscountRate:  null.IntFrom(int64(r.Intn(50))),
			ProductPromoActive:        util.RandomBool(),
			ProductPromoStartDate:     null.Time{},
			ProductPromoEndDate:       null.Time{},
		}
	}

	// --- Product sizes ---
	sizes := make([]*db.ProductSize, nSizes)
	for i := 0; i < nSizes; i++ {
		sizes[i] = &db.ProductSize{
			ID:            int64(i),
			ProductItemID: int64(i % nProd), // wrap around products
			SizeValue:     "L",
			Qty:           int32(r.Intn(20)),
		}
	}

	return shop, prod, sizes
}

func newlistShoppingCartItemsResponse_Old(
	shopCartItems []*db.ListShoppingCartItemsByUserIDRow,
	productItems []*db.ListProductItemsByIDsRow,
	productsSizes []*db.ProductSize,
) []*listShoppingCartItemsResponse {
	rsp := make([]*listShoppingCartItemsResponse, len(productItems))
	for i := 0; i < len(productItems); i++ {
		for j := 0; j < len(shopCartItems); j++ {
			for k := 0; k < len(productsSizes); k++ {
				if productItems[i].ID == shopCartItems[j].ProductItemID.Int64 && productItems[i].ID == productsSizes[k].ProductItemID {
					rsp[i] = &listShoppingCartItemsResponse{
						ID:             shopCartItems[j].ID,
						ShoppingCartID: shopCartItems[j].ShoppingCartID,
						CreatedAt:      shopCartItems[j].CreatedAt,
						UpdatedAt:      shopCartItems[j].UpdatedAt,
						ProductItemID:  shopCartItems[j].ProductItemID,
						Name:           productItems[i].Name,
						Qty:            shopCartItems[j].Qty,
						ProductID:      productItems[i].ProductID,
						// ProductImage:   productItems[i].ProductImage,
						ProductImage:              productItems[i].ProductImage1.String,
						SizeID:                    null.IntFromPtr(&productsSizes[k].ID),
						SizeValue:                 null.StringFromPtr(&productsSizes[k].SizeValue),
						SizeQty:                   null.Int32FromPtr(&productsSizes[k].Qty),
						Color:                     productItems[i].ColorValue,
						Price:                     productItems[i].Price,
						Active:                    productItems[i].Active,
						CategoryPromoID:           productItems[i].CategoryPromoID,
						CategoryPromoName:         productItems[i].CategoryPromoName,
						CategoryPromoDescription:  productItems[i].CategoryPromoDescription,
						CategoryPromoDiscountRate: productItems[i].CategoryPromoDiscountRate,
						CategoryPromoActive:       productItems[i].CategoryPromoActive,
						CategoryPromoStartDate:    productItems[i].CategoryPromoStartDate,
						CategoryPromoEndDate:      productItems[i].CategoryPromoEndDate,
						BrandPromoID:              productItems[i].BrandPromoID,
						BrandPromoName:            productItems[i].BrandPromoName,
						BrandPromoDescription:     productItems[i].BrandPromoDescription,
						BrandPromoDiscountRate:    productItems[i].BrandPromoDiscountRate,
						BrandPromoActive:          productItems[i].BrandPromoActive,
						BrandPromoStartDate:       productItems[i].BrandPromoStartDate,
						BrandPromoEndDate:         productItems[i].BrandPromoEndDate,
						ProductPromoID:            productItems[i].ProductPromoID,
						ProductPromoName:          productItems[i].ProductPromoName,
						ProductPromoDescription:   productItems[i].ProductPromoDescription,
						ProductPromoDiscountRate:  productItems[i].ProductPromoDiscountRate,
						ProductPromoActive:        productItems[i].ProductPromoActive,
						ProductPromoStartDate:     productItems[i].ProductPromoStartDate,
						ProductPromoEndDate:       productItems[i].ProductPromoEndDate,
					}
				}
			}
		}
	}

	return rsp
}

func newlistShoppingCartItemsResponse_Optimized(
	shopCartItems []*db.ListShoppingCartItemsByUserIDRow,
	productItems []*db.ListProductItemsByIDsRow,
	productSizes []*db.ProductSize,
	rsp []*listShoppingCartItemsResponse,
) []*listShoppingCartItemsResponse {
	// Build lookup maps O(n)
	cartMap := make(map[int64]*db.ListShoppingCartItemsByUserIDRow, len(shopCartItems))
	for i := 0; i < len(shopCartItems); i++ {
		sc := shopCartItems[i]
		cartMap[sc.ProductItemID.Int64] = sc
	}

	productMap := make(map[int64]*db.ListProductItemsByIDsRow, len(productItems))
	for i := 0; i < len(productItems); i++ {
		p := productItems[i]
		productMap[p.ID] = p
	}

	sizeMap := make(map[int64]*db.ProductSize, len(productSizes))
	for i := 0; i < len(productSizes); i++ {
		s := productSizes[i]
		sizeMap[s.ProductItemID] = s
	}

	// Fill with classic for-loop
	for i := 0; i < len(productItems); i++ {
		p := productItems[i]
		pid := p.ID

		sc := cartMap[pid]

		s := sizeMap[pid]

		rsp[i] = &listShoppingCartItemsResponse{
			ID:             sc.ID,
			ShoppingCartID: sc.ShoppingCartID,
			CreatedAt:      sc.CreatedAt,
			UpdatedAt:      sc.UpdatedAt,
			ProductItemID:  sc.ProductItemID,
			Name:           p.Name,
			Qty:            sc.Qty,
			ProductID:      p.ProductID,
			ProductImage:   p.ProductImage1.String,

			SizeID:    null.IntFromPtr(&s.ID),
			SizeValue: null.StringFromPtr(&s.SizeValue),
			SizeQty:   null.Int32FromPtr(&s.Qty),

			Color:  p.ColorValue,
			Price:  p.Price,
			Active: p.Active,

			CategoryPromoID:           p.CategoryPromoID,
			CategoryPromoName:         p.CategoryPromoName,
			CategoryPromoDescription:  p.CategoryPromoDescription,
			CategoryPromoDiscountRate: p.CategoryPromoDiscountRate,
			CategoryPromoActive:       p.CategoryPromoActive,
			CategoryPromoStartDate:    p.CategoryPromoStartDate,
			CategoryPromoEndDate:      p.CategoryPromoEndDate,

			BrandPromoID:           p.BrandPromoID,
			BrandPromoName:         p.BrandPromoName,
			BrandPromoDescription:  p.BrandPromoDescription,
			BrandPromoDiscountRate: p.BrandPromoDiscountRate,
			BrandPromoActive:       p.BrandPromoActive,
			BrandPromoStartDate:    p.BrandPromoStartDate,
			BrandPromoEndDate:      p.BrandPromoEndDate,

			ProductPromoID:           p.ProductPromoID,
			ProductPromoName:         p.ProductPromoName,
			ProductPromoDescription:  p.ProductPromoDescription,
			ProductPromoDiscountRate: p.ProductPromoDiscountRate,
			ProductPromoActive:       p.ProductPromoActive,
			ProductPromoStartDate:    p.ProductPromoStartDate,
			ProductPromoEndDate:      p.ProductPromoEndDate,
		}
	}

	return rsp
}
