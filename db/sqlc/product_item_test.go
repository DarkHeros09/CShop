package db

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/cshop/v3/util"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func createRandomProductItem(t *testing.T) ProductItem {
	product := createRandomProduct(t)
	productSize := createRandomProductSize(t)
	productColor := createRandomProductColor(t)
	productImage := createRandomProductImage(t)
	arg := CreateProductItemParams{
		ProductID:  product.ID,
		ProductSku: util.RandomInt(100, 300),
		QtyInStock: int32(util.RandomInt(0, 100)),
		SizeID:     productSize.ID,
		ImageID:    productImage.ID,
		ColorID:    productColor.ID,
		// ProductImage: util.RandomURL(),
		Price:  util.RandomDecimalString(1, 100),
		Active: true,
	}

	productItem, err := testStore.CreateProductItem(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, productItem)

	require.Equal(t, arg.ProductID, productItem.ProductID)
	require.Equal(t, arg.ProductSku, productItem.ProductSku)
	require.Equal(t, arg.QtyInStock, productItem.QtyInStock)
	// require.Equal(t, arg.ProductImage, productItem.ProductImage)
	require.Equal(t, arg.SizeID, productItem.SizeID)
	require.Equal(t, arg.ColorID, productItem.ColorID)
	require.Equal(t, arg.ImageID, productItem.ImageID)
	require.Equal(t, arg.Price, productItem.Price)
	require.Equal(t, arg.Active, productItem.Active)
	require.NotEmpty(t, productItem.CreatedAt)
	require.True(t, productItem.UpdatedAt.IsZero())
	require.True(t, productItem.Active)

	if util.RandomBool() {
		promotion := createRandomPromotion(t)

		rand := util.RandomInt(1, 3)

		switch rand {
		case 1:

			arg1 := CreateCategoryPromotionParams{
				CategoryID:             product.CategoryID,
				PromotionID:            promotion.ID,
				CategoryPromotionImage: null.StringFrom(util.RandomPromotionURL()),
				Active:                 util.RandomBool(),
			}

			categoryPromotion, err := testStore.CreateCategoryPromotion(context.Background(), arg1)
			require.NoError(t, err)
			require.NotEmpty(t, categoryPromotion)

			require.Equal(t, arg1.CategoryID, categoryPromotion.CategoryID)
			require.Equal(t, arg1.PromotionID, categoryPromotion.PromotionID)
			require.Equal(t, arg1.Active, categoryPromotion.Active)

		case 2:

			arg1 := CreateBrandPromotionParams{
				BrandID:             product.BrandID,
				PromotionID:         promotion.ID,
				BrandPromotionImage: null.StringFrom(util.RandomPromotionURL()),
				Active:              util.RandomBool(),
			}

			brandPromotion, err := testStore.CreateBrandPromotion(context.Background(), arg1)
			require.NoError(t, err)
			require.NotEmpty(t, brandPromotion)

			require.Equal(t, arg1.BrandID, brandPromotion.BrandID)
			require.Equal(t, arg1.PromotionID, brandPromotion.PromotionID)
			require.Equal(t, arg1.Active, brandPromotion.Active)

		case 3:
			arg1 := CreateProductPromotionParams{
				ProductID:             product.ID,
				PromotionID:           promotion.ID,
				ProductPromotionImage: null.StringFrom(util.RandomPromotionURL()),
				Active:                util.RandomBool(),
			}

			productPromotion, err := testStore.CreateProductPromotion(context.Background(), arg1)
			require.NoError(t, err)
			require.NotEmpty(t, productPromotion)

			require.Equal(t, arg1.ProductID, productPromotion.ProductID)
			require.Equal(t, arg1.PromotionID, productPromotion.PromotionID)
			require.Equal(t, arg1.Active, productPromotion.Active)
		}
	}

	return productItem
}

func adminCreateRandomProductItem(t *testing.T) ProductItem {
	admin := createRandomAdmin(t)
	product := createRandomProduct(t)
	productSize := createRandomProductSize(t)
	productColor := createRandomProductColor(t)
	productImage := createRandomProductImage(t)
	arg := AdminCreateProductItemParams{
		AdminID:    admin.ID,
		ProductID:  product.ID,
		ProductSku: util.RandomInt(100, 300),
		QtyInStock: int32(util.RandomInt(0, 100)),
		SizeID:     productSize.ID,
		ImageID:    productImage.ID,
		ColorID:    productColor.ID,
		// ProductImage: util.RandomURL(),
		Price:  util.RandomDecimalString(1, 100),
		Active: true,
	}

	productItem, err := testStore.AdminCreateProductItem(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, productItem)

	require.Equal(t, arg.ProductID, productItem.ProductID)
	require.Equal(t, arg.ProductSku, productItem.ProductSku)
	require.Equal(t, arg.QtyInStock, productItem.QtyInStock)
	// require.Equal(t, arg.ProductImage, productItem.ProductImage)
	require.Equal(t, arg.SizeID, productItem.SizeID)
	require.Equal(t, arg.ColorID, productItem.ColorID)
	require.Equal(t, arg.ImageID, productItem.ImageID)
	require.Equal(t, arg.Price, productItem.Price)
	require.Equal(t, arg.Active, productItem.Active)
	require.NotEmpty(t, productItem.CreatedAt)
	require.True(t, productItem.UpdatedAt.IsZero())
	require.True(t, productItem.Active)

	if util.RandomBool() {
		promotion := createRandomPromotion(t)

		rand := util.RandomInt(1, 3)

		switch rand {
		case 1:

			arg1 := CreateCategoryPromotionParams{
				CategoryID:             product.CategoryID,
				PromotionID:            promotion.ID,
				CategoryPromotionImage: null.StringFrom(util.RandomPromotionURL()),
				Active:                 util.RandomBool(),
			}

			categoryPromotion, err := testStore.CreateCategoryPromotion(context.Background(), arg1)
			require.NoError(t, err)
			require.NotEmpty(t, categoryPromotion)

			require.Equal(t, arg1.CategoryID, categoryPromotion.CategoryID)
			require.Equal(t, arg1.PromotionID, categoryPromotion.PromotionID)
			require.Equal(t, arg1.Active, categoryPromotion.Active)

		case 2:

			arg1 := CreateBrandPromotionParams{
				BrandID:             product.BrandID,
				PromotionID:         promotion.ID,
				BrandPromotionImage: null.StringFrom(util.RandomPromotionURL()),
				Active:              util.RandomBool(),
			}

			brandPromotion, err := testStore.CreateBrandPromotion(context.Background(), arg1)
			require.NoError(t, err)
			require.NotEmpty(t, brandPromotion)

			require.Equal(t, arg1.BrandID, brandPromotion.BrandID)
			require.Equal(t, arg1.PromotionID, brandPromotion.PromotionID)
			require.Equal(t, arg1.Active, brandPromotion.Active)

		case 3:
			arg1 := CreateProductPromotionParams{
				ProductID:             product.ID,
				PromotionID:           promotion.ID,
				ProductPromotionImage: null.StringFrom(util.RandomPromotionURL()),
				Active:                util.RandomBool(),
			}

			productPromotion, err := testStore.CreateProductPromotion(context.Background(), arg1)
			require.NoError(t, err)
			require.NotEmpty(t, productPromotion)

			require.Equal(t, arg1.ProductID, productPromotion.ProductID)
			require.Equal(t, arg1.PromotionID, productPromotion.PromotionID)
			require.Equal(t, arg1.Active, productPromotion.Active)
		}
	}

	return productItem
}
func TestCreateProductItem(t *testing.T) {
	createRandomProductItem(t)
}

func TestAdminCreateProductItem(t *testing.T) {
	adminCreateRandomProductItem(t)
}

func TestGetProductItem(t *testing.T) {
	productItem1 := createRandomProductItem(t)

	productItem2, err := testStore.GetProductItem(context.Background(), productItem1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, productItem2)

	require.Equal(t, productItem1.ProductID, productItem2.ProductID)
	require.Equal(t, productItem1.ProductSku, productItem2.ProductSku)
	require.Equal(t, productItem1.QtyInStock, productItem2.QtyInStock)
	// require.Equal(t, productItem1.ProductImage, productItem2.ProductImage)
	require.Equal(t, productItem1.SizeID, productItem2.SizeID)
	require.Equal(t, productItem1.ColorID, productItem2.ColorID)
	require.Equal(t, productItem1.ImageID, productItem2.ImageID)
	require.Equal(t, productItem1.Price, productItem2.Price)
	require.Equal(t, productItem1.Active, productItem2.Active)
	require.Equal(t, productItem1.CreatedAt, productItem2.CreatedAt)
	require.Equal(t, productItem1.UpdatedAt, productItem2.UpdatedAt)
	require.True(t, productItem2.Active)

}

func TestGetProductItemForUpdate(t *testing.T) {
	productItem1 := createRandomProductItem(t)

	productItem2, err := testStore.GetProductItemForUpdate(context.Background(), productItem1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, productItem2)

	require.Equal(t, productItem1.ProductID, productItem2.ProductID)
	require.Equal(t, productItem1.ProductSku, productItem2.ProductSku)
	require.Equal(t, productItem1.QtyInStock, productItem2.QtyInStock)
	// require.Equal(t, productItem1.ProductImage, productItem2.ProductImage)
	require.Equal(t, productItem1.SizeID, productItem2.SizeID)
	require.Equal(t, productItem1.ColorID, productItem2.ColorID)
	require.Equal(t, productItem1.ImageID, productItem2.ImageID)
	require.Equal(t, productItem1.Price, productItem2.Price)
	require.Equal(t, productItem1.Active, productItem2.Active)
	require.Equal(t, productItem1.CreatedAt, productItem2.CreatedAt)
	require.Equal(t, productItem1.UpdatedAt, productItem2.UpdatedAt)
	require.True(t, productItem2.Active)

}

func TestGetProductItemForUpdateWithPromotion(t *testing.T) {
	productItem1 := createRandomProductItem(t)

	productItem2, err := testStore.GetProductItemWithPromotions(context.Background(), productItem1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, productItem2)

	require.Equal(t, productItem1.ProductID, productItem2.ProductID)
	require.Equal(t, productItem1.ProductSku, productItem2.ProductSku)
	require.Equal(t, productItem1.QtyInStock, productItem2.QtyInStock)
	// require.Equal(t, productItem1.ProductImage, productItem2.ProductImage)
	require.Equal(t, productItem1.SizeID, productItem2.SizeID)
	require.Equal(t, productItem1.ColorID, productItem2.ColorID)
	require.Equal(t, productItem1.ImageID, productItem2.ImageID)
	require.Equal(t, productItem1.Price, productItem2.Price)
	require.Equal(t, productItem1.Active, productItem2.Active)
	require.Equal(t, productItem1.CreatedAt, productItem2.CreatedAt)
	require.Equal(t, productItem1.UpdatedAt, productItem2.UpdatedAt)
	require.True(t, productItem2.Active)

}

func TestUpdateProductItemQtyAndPriceAndActive(t *testing.T) {
	productItem1 := createRandomProductItem(t)
	arg := UpdateProductItemParams{
		ProductID:  productItem1.ProductID,
		ProductSku: null.Int{},
		QtyInStock: null.IntFrom(util.RandomInt(1, 500)),
		// ProductImage: null.String{},
		Price:  null.StringFrom(util.RandomDecimalString(1, 100)),
		Active: null.BoolFrom(!productItem1.Active),
		ID:     productItem1.ID,
	}

	productItem2, err := testStore.UpdateProductItem(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, productItem2)

	require.Equal(t, productItem1.ProductID, productItem2.ProductID)
	require.Equal(t, productItem1.ProductSku, productItem2.ProductSku)
	require.NotEqual(t, productItem1.QtyInStock, productItem2.QtyInStock)
	// require.Equal(t, productItem1.ProductImage, productItem2.ProductImage)
	require.Equal(t, productItem1.SizeID, productItem2.SizeID)
	require.Equal(t, productItem1.ColorID, productItem2.ColorID)
	require.Equal(t, productItem1.ImageID, productItem2.ImageID)
	require.NotEqual(t, productItem1.Price, productItem2.Price)
	require.NotEqual(t, productItem1.Active, productItem2.Active)
	require.False(t, productItem2.Active)
	require.WithinDuration(t, productItem1.CreatedAt, productItem2.CreatedAt, time.Second)
	require.NotEqual(t, productItem1.UpdatedAt, productItem2.UpdatedAt)
}

func TestAdminUpdateProductItemQtyAndPriceAndActive(t *testing.T) {
	admin := createRandomAdmin(t)
	productItem1 := createRandomProductItem(t)
	arg := AdminUpdateProductItemParams{
		AdminID:    admin.ID,
		ProductID:  productItem1.ProductID,
		ProductSku: null.Int{},
		QtyInStock: null.IntFrom(util.RandomInt(1, 500)),
		// ProductImage: null.String{},
		Price:  null.StringFrom(util.RandomDecimalString(1, 100)),
		Active: null.BoolFrom(!productItem1.Active),
		ID:     productItem1.ID,
	}

	productItem2, err := testStore.AdminUpdateProductItem(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, productItem2)

	require.Equal(t, productItem1.ProductID, productItem2.ProductID)
	require.Equal(t, productItem1.ProductSku, productItem2.ProductSku)
	require.NotEqual(t, productItem1.QtyInStock, productItem2.QtyInStock)
	// require.Equal(t, productItem1.ProductImage, productItem2.ProductImage)
	require.Equal(t, productItem1.SizeID, productItem2.SizeID)
	require.Equal(t, productItem1.ColorID, productItem2.ColorID)
	require.Equal(t, productItem1.ImageID, productItem2.ImageID)
	require.NotEqual(t, productItem1.Price, productItem2.Price)
	require.NotEqual(t, productItem1.Active, productItem2.Active)
	require.False(t, productItem2.Active)
	require.WithinDuration(t, productItem1.CreatedAt, productItem2.CreatedAt, time.Second)
	require.NotEqual(t, productItem1.UpdatedAt, productItem2.UpdatedAt)
}

func TestDeleteProductItem(t *testing.T) {
	productItem1 := createRandomProductItem(t)
	err := testStore.DeleteProductItem(context.Background(), productItem1.ID)

	require.NoError(t, err)

	productItem2, err := testStore.GetProductItem(context.Background(), productItem1.ID)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, productItem2)

}

func TestListProductItems(t *testing.T) {
	t.Parallel()
	for i := 0; i < 5; i++ {
		createRandomProductItem(t)
	}
	arg := ListProductItemsParams{
		Limit:  5,
		Offset: 0,
	}

	productItems, err := testStore.ListProductItems(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, productItems)

	for _, productItem := range productItems {
		require.NotEmpty(t, productItem)
	}

}

func TestListProductItemsByIDs(t *testing.T) {
	var productsIds []int64

	for i := 0; i < 5; i++ {
		pi := createRandomProductItem(t)
		productsIds = append(productsIds, pi.ID)
	}

	fmt.Println("ProductsIDS", productsIds)

	productItems, err := testStore.ListProductItemsByIDs(context.Background(), productsIds)
	require.NoError(t, err)
	require.NotEmpty(t, productItems)

	for i, productItem := range productItems {
		require.NotEmpty(t, productItem)
		require.Equal(t, productItems[i].ID, productsIds[i])
	}

}

func TestListProductItemsV2(t *testing.T) {
	for i := 0; i < 30; i++ {
		createRandomProductItem(t)
	}

	arg := ListProductItemsV2Params{
		Limit: 10,
	}

	initialSearchResult, err := testStore.ListProductItemsV2(context.Background(), arg)
	// fmt.Println(initialSearchResult)
	require.NoError(t, err)
	require.NotEmpty(t, initialSearchResult)
	require.Equal(t, 10, len(initialSearchResult))

	arg1 := ListProductItemsNextPageParams{
		Limit:         10,
		ProductItemID: initialSearchResult[len(initialSearchResult)-1].ID,
		ProductID:     initialSearchResult[len(initialSearchResult)-1].ProductID,
	}

	secondPage, err := testStore.ListProductItemsNextPage(context.Background(), arg1)
	fmt.Println(len(secondPage))
	require.NoError(t, err)
	require.NotEmpty(t, secondPage)
	require.Equal(t, 10, len(secondPage))
	require.Greater(t, initialSearchResult[len(initialSearchResult)-1].ID, secondPage[len(secondPage)-1].ID)

	arg2 := ListProductItemsNextPageParams{
		Limit:         10,
		ProductItemID: secondPage[len(secondPage)-1].ID,
		ProductID:     secondPage[len(secondPage)-1].ProductID,
	}

	thirdPage, err := testStore.ListProductItemsNextPage(context.Background(), arg2)
	require.NoError(t, err)
	require.NotEmpty(t, secondPage)
	require.Equal(t, len(initialSearchResult), 10)
	require.Greater(t, secondPage[len(secondPage)-1].ID, thirdPage[len(thirdPage)-1].ID)
	require.Greater(t, initialSearchResult[len(initialSearchResult)-1].ID, thirdPage[len(thirdPage)-1].ID)
}
func TestListProductItemsV2OrderByHighPrice(t *testing.T) {
	t.Parallel()
	for i := 0; i < 30; i++ {
		createRandomProductItem(t)
	}

	arg1 := ListProductItemsV2Params{
		Limit:            10,
		OrderByHighPrice: true,
	}

	initialSearchResult, err := testStore.ListProductItemsV2(context.Background(), arg1)
	// fmt.Println(initialSearchResult)
	require.NoError(t, err)
	require.NotEmpty(t, initialSearchResult)
	require.Equal(t, 10, len(initialSearchResult))
	for i := 0; i < len(initialSearchResult)-1; i++ {
		price1, err := strconv.ParseFloat(initialSearchResult[i].Price, 64)
		require.NoError(t, err)
		price2, err := strconv.ParseFloat(initialSearchResult[i+1].Price, 64)
		require.NoError(t, err)
		require.GreaterOrEqual(t, price1, price2)

	}

	arg2 := ListProductItemsNextPageParams{
		Limit:            10,
		ProductItemID:    initialSearchResult[len(initialSearchResult)-1].ID,
		ProductID:        initialSearchResult[len(initialSearchResult)-1].ProductID,
		OrderByHighPrice: true,
		Price:            null.StringFrom(initialSearchResult[len(initialSearchResult)-1].Price),
	}

	secondPage, err := testStore.ListProductItemsNextPage(context.Background(), arg2)
	fmt.Println(len(secondPage))
	require.NoError(t, err)
	require.NotEmpty(t, secondPage)
	require.Equal(t, 10, len(secondPage))
	require.GreaterOrEqual(t, initialSearchResult[len(initialSearchResult)-1].Price, secondPage[len(secondPage)-1].Price)
	for i := 0; i < len(secondPage)-1; i++ {
		price1, err := strconv.ParseFloat(secondPage[i].Price, 64)
		require.NoError(t, err)
		price2, err := strconv.ParseFloat(secondPage[i+1].Price, 64)
		require.NoError(t, err)
		require.GreaterOrEqual(t, price1, price2)

	}

	arg3 := ListProductItemsNextPageParams{
		Limit:            10,
		ProductItemID:    secondPage[len(secondPage)-1].ID,
		ProductID:        secondPage[len(secondPage)-1].ProductID,
		OrderByHighPrice: true,
		Price:            null.StringFrom(secondPage[len(secondPage)-1].Price),
	}

	thirdPage, err := testStore.ListProductItemsNextPage(context.Background(), arg3)
	require.NoError(t, err)
	require.NotEmpty(t, thirdPage)
	require.Equal(t, len(thirdPage), 10)
	require.GreaterOrEqual(t, secondPage[len(secondPage)-1].Price, thirdPage[len(thirdPage)-1].Price)
	// require.Greater(t, initialSearchResult[len(initialSearchResult)-1].ID, thirdPage[len(thirdPage)-1].ID)
	for i := 0; i < len(thirdPage)-1; i++ {
		price1, err := strconv.ParseFloat(thirdPage[i].Price, 64)
		require.NoError(t, err)
		price2, err := strconv.ParseFloat(thirdPage[i+1].Price, 64)
		require.NoError(t, err)
		require.GreaterOrEqual(t, price1, price2)

	}
}

func TestListProductItemsV2OrderByLowPrice(t *testing.T) {
	t.Parallel()
	for i := 0; i < 30; i++ {
		createRandomProductItem(t)
	}

	arg1 := ListProductItemsV2Params{
		Limit:           10,
		OrderByLowPrice: true,
	}

	initialSearchResult, err := testStore.ListProductItemsV2(context.Background(), arg1)
	// fmt.Println(initialSearchResult)
	require.NoError(t, err)
	require.NotEmpty(t, initialSearchResult)
	require.Equal(t, 10, len(initialSearchResult))
	for i := 0; i < len(initialSearchResult)-1; i++ {
		price1, err := strconv.ParseFloat(initialSearchResult[i].Price, 64)
		require.NoError(t, err)
		price2, err := strconv.ParseFloat(initialSearchResult[i+1].Price, 64)
		require.NoError(t, err)
		require.LessOrEqual(t, price1, price2)

	}

	arg2 := ListProductItemsNextPageParams{
		Limit:           10,
		ProductItemID:   initialSearchResult[len(initialSearchResult)-1].ID,
		ProductID:       initialSearchResult[len(initialSearchResult)-1].ProductID,
		OrderByLowPrice: true,
		Price:           null.StringFrom(initialSearchResult[len(initialSearchResult)-1].Price),
	}

	secondPage, err := testStore.ListProductItemsNextPage(context.Background(), arg2)
	fmt.Println(len(secondPage))
	require.NoError(t, err)
	require.NotEmpty(t, secondPage)
	require.Equal(t, 10, len(secondPage))
	require.LessOrEqual(t, initialSearchResult[len(initialSearchResult)-1].Price, secondPage[len(secondPage)-1].Price)
	for i := 0; i < len(secondPage)-1; i++ {
		price1, err := strconv.ParseFloat(secondPage[i].Price, 64)
		require.NoError(t, err)
		price2, err := strconv.ParseFloat(secondPage[i+1].Price, 64)
		require.NoError(t, err)
		require.LessOrEqual(t, price1, price2)

	}

	arg3 := ListProductItemsNextPageParams{
		Limit:           10,
		ProductItemID:   secondPage[len(secondPage)-1].ID,
		ProductID:       secondPage[len(secondPage)-1].ProductID,
		OrderByLowPrice: true,
		Price:           null.StringFrom(secondPage[len(secondPage)-1].Price),
	}

	thirdPage, err := testStore.ListProductItemsNextPage(context.Background(), arg3)
	require.NoError(t, err)
	require.NotEmpty(t, thirdPage)
	require.Equal(t, len(thirdPage), 10)
	require.LessOrEqual(t, secondPage[len(secondPage)-1].Price, thirdPage[len(thirdPage)-1].Price)
	// require.Greater(t, initialSearchResult[len(initialSearchResult)-1].ID, thirdPage[len(thirdPage)-1].ID)
	for i := 0; i < len(thirdPage)-1; i++ {
		price1, err := strconv.ParseFloat(thirdPage[i].Price, 64)
		require.NoError(t, err)
		price2, err := strconv.ParseFloat(thirdPage[i+1].Price, 64)
		require.NoError(t, err)
		require.LessOrEqual(t, price1, price2)

	}
}

func TestListProductItemsV2OrderByNew(t *testing.T) {
	for i := 0; i < 30; i++ {
		createRandomProductItem(t)
	}

	arg1 := ListProductItemsV2Params{
		Limit:      10,
		OrderByNew: true,
	}

	initialSearchResult, err := testStore.ListProductItemsV2(context.Background(), arg1)
	fmt.Println(initialSearchResult[len(initialSearchResult)-1].CreatedAt)
	require.NoError(t, err)
	require.NotEmpty(t, initialSearchResult)
	require.Equal(t, 10, len(initialSearchResult))

	arg2 := ListProductItemsNextPageParams{
		Limit:         10,
		ProductItemID: initialSearchResult[len(initialSearchResult)-1].ID,
		ProductID:     initialSearchResult[len(initialSearchResult)-1].ProductID,
		OrderByNew:    true,
		CreatedAt:     null.TimeFrom(initialSearchResult[len(initialSearchResult)-1].CreatedAt),
	}

	secondPage, err := testStore.ListProductItemsNextPage(context.Background(), arg2)
	fmt.Println(len(secondPage))
	require.NoError(t, err)
	require.NotEmpty(t, secondPage)
	require.Equal(t, 10, len(secondPage))
	require.GreaterOrEqual(t, initialSearchResult[len(initialSearchResult)-1].CreatedAt, secondPage[len(secondPage)-1].CreatedAt)

	arg3 := ListProductItemsNextPageParams{
		Limit:         10,
		ProductItemID: secondPage[len(secondPage)-1].ID,
		ProductID:     secondPage[len(secondPage)-1].ProductID,
		OrderByNew:    true,
		CreatedAt:     null.TimeFrom(secondPage[len(secondPage)-1].CreatedAt),
	}

	thirdPage, err := testStore.ListProductItemsNextPage(context.Background(), arg3)
	require.NoError(t, err)
	require.NotEmpty(t, thirdPage)
	require.Equal(t, len(thirdPage), 10)
	require.GreaterOrEqual(t, secondPage[len(secondPage)-1].CreatedAt, thirdPage[len(thirdPage)-1].CreatedAt)
	// require.Greater(t, initialSearchResult[len(initialSearchResult)-1].ID, thirdPage[len(thirdPage)-1].ID)

}

func TestListProductItemsV2OrderByOld(t *testing.T) {
	for i := 0; i < 30; i++ {
		createRandomProductItem(t)
	}

	arg1 := ListProductItemsV2Params{
		Limit:      10,
		OrderByOld: true,
	}

	initialSearchResult, err := testStore.ListProductItemsV2(context.Background(), arg1)
	fmt.Println(initialSearchResult[len(initialSearchResult)-1].CreatedAt)
	require.NoError(t, err)
	require.NotEmpty(t, initialSearchResult)
	require.Equal(t, 10, len(initialSearchResult))

	arg2 := ListProductItemsNextPageParams{
		Limit:         10,
		ProductItemID: initialSearchResult[len(initialSearchResult)-1].ID,
		ProductID:     initialSearchResult[len(initialSearchResult)-1].ProductID,
		OrderByOld:    true,
		CreatedAt:     null.TimeFrom(initialSearchResult[len(initialSearchResult)-1].CreatedAt),
	}

	secondPage, err := testStore.ListProductItemsNextPage(context.Background(), arg2)
	fmt.Println(len(secondPage))
	require.NoError(t, err)
	require.NotEmpty(t, secondPage)
	require.Equal(t, 10, len(secondPage))
	require.LessOrEqual(t, initialSearchResult[len(initialSearchResult)-1].CreatedAt, secondPage[len(secondPage)-1].CreatedAt)

	arg3 := ListProductItemsNextPageParams{
		Limit:         10,
		ProductItemID: secondPage[len(secondPage)-1].ID,
		ProductID:     secondPage[len(secondPage)-1].ProductID,
		OrderByOld:    true,
		CreatedAt:     null.TimeFrom(secondPage[len(secondPage)-1].CreatedAt),
	}

	thirdPage, err := testStore.ListProductItemsNextPage(context.Background(), arg3)
	require.NoError(t, err)
	require.NotEmpty(t, thirdPage)
	require.Equal(t, len(thirdPage), 10)
	require.LessOrEqual(t, secondPage[len(secondPage)-1].CreatedAt, thirdPage[len(thirdPage)-1].CreatedAt)
	// require.Greater(t, initialSearchResult[len(initialSearchResult)-1].ID, thirdPage[len(thirdPage)-1].ID)

}

func TestSearchProductItems(t *testing.T) {

	productItem := createRandomProductItem(t)

	product, err := testStore.GetProduct(context.Background(), productItem.ProductID)

	require.NoError(t, err)
	require.NotEmpty(t, product)

	arg1 := SearchProductItemsParams{
		Limit: 10,
		Query: product.Name,
	}

	searchedProductItem, err := testStore.SearchProductItems(context.Background(), arg1)

	require.NoError(t, err)
	require.NotEmpty(t, searchedProductItem)
	require.Equal(t, productItem.ID, searchedProductItem[len(searchedProductItem)-1].ID)

	arg2 := SearchProductItemsNextPageParams{
		Limit:         10,
		ProductItemID: searchedProductItem[len(searchedProductItem)-1].ID,
		ProductID:     searchedProductItem[len(searchedProductItem)-1].ProductID,
		Query:         product.Name,
	}

	searchedRestProductItem, err := testStore.SearchProductItemsNextPage(context.Background(), arg2)

	require.NoError(t, err)
	require.Empty(t, searchedRestProductItem)
	// require.Equal(t, productItem.ID, searchedProductItem[len(searchedProductItem)-1].ID)
}

func TestListProductItemsWithPromotion(t *testing.T) {
	t.Parallel()
	var productId int64
	var ok bool = false
	for {
		pi := createRandomProductItem(t)
		for pi.Active {
			arg := ListProductItemsWithPromotionsParams{
				Limit:     10,
				ProductID: pi.ProductID,
			}
			productItem, err := testStore.ListProductItemsWithPromotions(context.Background(), arg)
			require.NoError(t, err)
			if len(productItem) > 0 && productItem[len(productItem)-1].ProductPromoActive == true {
				productId = productItem[len(productItem)-1].ProductID
				ok = true
				break
			}
			break
		}
		if ok {
			break
		}
	}

	arg := ListProductItemsWithPromotionsParams{
		Limit:     10,
		ProductID: productId,
	}

	initialSearchResult, err := testStore.ListProductItemsWithPromotions(context.Background(), arg)
	// fmt.Println(initialSearchResult)
	require.NoError(t, err)
	require.NotEmpty(t, initialSearchResult)
	// require.Equal(t, len(initialSearchResult), 10)

	arg1 := ListProductItemsWithPromotionsNextPageParams{
		Limit:         10,
		ProductItemID: initialSearchResult[len(initialSearchResult)-1].ID,
		ProductID:     initialSearchResult[len(initialSearchResult)-1].ProductID,
	}

	_, err = testStore.ListProductItemsWithPromotionsNextPage(context.Background(), arg1)
	// fmt.Println(secondPage)
	require.NoError(t, err)
	// require.NotEmpty(t, secondPage)
	// require.Equal(t, len(secondPage), 10)
	// require.Greater(t, initialSearchResult[len(initialSearchResult)-1].ID, secondPage[len(secondPage)-1].ID)

}

func TestListProductItemsWithBrandPromotion(t *testing.T) {
	t.Parallel()
	var brandId int64
	for {
		pi := createRandomProductItem(t)
		p, err := testStore.GetProduct(context.Background(), pi.ProductID)
		require.NoError(t, err)
		if p.Active {
			arg := ListProductItemsWithBrandPromotionsParams{
				Limit:   10,
				BrandID: p.BrandID,
			}
			productItem, err := testStore.ListProductItemsWithBrandPromotions(context.Background(), arg)
			require.NoError(t, err)
			if len(productItem) > 0 && productItem[len(productItem)-1].BrandPromoActive {
				brandId = productItem[len(productItem)-1].BrandID
				break
			}
		}
	}

	arg := ListProductItemsWithBrandPromotionsParams{
		Limit:   10,
		BrandID: brandId,
	}

	initialSearchResult, err := testStore.ListProductItemsWithBrandPromotions(context.Background(), arg)
	// fmt.Println(initialSearchResult)
	require.NoError(t, err)
	require.NotEmpty(t, initialSearchResult)
	// require.Equal(t, len(initialSearchResult), 10)

	arg1 := ListProductItemsWithBrandPromotionsNextPageParams{
		Limit:         10,
		ProductItemID: initialSearchResult[len(initialSearchResult)-1].ID,
		ProductID:     initialSearchResult[len(initialSearchResult)-1].ProductID,
	}

	_, err = testStore.ListProductItemsWithBrandPromotionsNextPage(context.Background(), arg1)
	// fmt.Println(secondPage)
	require.NoError(t, err)
	// require.NotEmpty(t, secondPage)
	// require.Equal(t, len(secondPage), 10)
	// require.Greater(t, initialSearchResult[len(initialSearchResult)-1].ID, secondPage[len(secondPage)-1].ID)

}

func TestListProductItemsWithCategoryPromotion(t *testing.T) {
	t.Parallel()
	var categoryId int64
	var isOK = true
	for isOK {
		pi := createRandomProductItem(t)
		p, err := testStore.GetProduct(context.Background(), pi.ProductID)
		require.NoError(t, err)
		if p.Active {
			arg := ListProductItemsWithCategoryPromotionsParams{
				Limit:      10,
				CategoryID: p.CategoryID,
			}
			productItem, err := testStore.ListProductItemsWithCategoryPromotions(context.Background(), arg)
			require.NoError(t, err)
			if len(productItem) > 0 && productItem[len(productItem)-1].CategoryPromoActive == true {
				categoryId = productItem[len(productItem)-1].CategoryID
				isOK = false
			}
		}
	}

	arg := ListProductItemsWithCategoryPromotionsParams{
		Limit:      10,
		CategoryID: categoryId,
	}

	initialSearchResult, err := testStore.ListProductItemsWithCategoryPromotions(context.Background(), arg)
	// fmt.Println(initialSearchResult)
	require.NoError(t, err)
	require.NotEmpty(t, initialSearchResult)
	// require.Equal(t, len(initialSearchResult), 10)

	arg1 := ListProductItemsWithCategoryPromotionsNextPageParams{
		Limit:         10,
		ProductItemID: initialSearchResult[len(initialSearchResult)-1].ID,
		ProductID:     initialSearchResult[len(initialSearchResult)-1].ProductID,
	}

	_, err = testStore.ListProductItemsWithCategoryPromotionsNextPage(context.Background(), arg1)
	// fmt.Println(secondPage)
	require.NoError(t, err)
	// require.NotEmpty(t, secondPage)
	// require.Equal(t, len(secondPage), 10)
	// require.Greater(t, initialSearchResult[len(initialSearchResult)-1].ID, secondPage[len(secondPage)-1].ID)

}

func TestGetActiveProductItems(t *testing.T) {
	admin := createRandomAdmin(t)

	productItem2, err := testStore.GetActiveProductItems(context.Background(), admin.ID)
	require.NoError(t, err)
	require.NotEmpty(t, productItem2)
}

func TestGetTotalProductItems(t *testing.T) {
	admin := createRandomAdmin(t)

	productItem2, err := testStore.GetTotalProductItems(context.Background(), admin.ID)
	require.NoError(t, err)
	require.NotEmpty(t, productItem2)
}

func TestListProductItemsWithBestSales(t *testing.T) {
	for i := 0; i < 30; i++ {
		createRandomProductItem(t)
	}

	arg := int32(20)
	initialSearchResult, err := testStore.ListProductItemsWithBestSales(context.Background(), arg)
	// fmt.Println(initialSearchResult)
	require.NoError(t, err)
	require.NotEmpty(t, initialSearchResult)
	require.Equal(t, len(initialSearchResult), 20)
}
