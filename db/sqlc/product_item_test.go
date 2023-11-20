package db

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/cshop/v3/util"
	"github.com/guregu/null"
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
func TestCreateProductItem(t *testing.T) {
	createRandomProductItem(t)
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

	require.NoError(t, err)
	require.NotEmpty(t, initialSearchResult)
	require.Equal(t, len(initialSearchResult), 10)

	arg1 := ListProductItemsNextPageParams{
		Limit:         10,
		ProductItemID: initialSearchResult[len(initialSearchResult)-1].ID,
		ProductID:     initialSearchResult[len(initialSearchResult)-1].ProductID,
	}

	secondPage, err := testStore.ListProductItemsNextPage(context.Background(), arg1)
	require.NoError(t, err)
	require.NotEmpty(t, secondPage)
	require.Equal(t, len(secondPage), 10)
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
