package db

import (
	"context"
	"testing"
	"time"

	"github.com/cshop/v3/util"
	"github.com/guregu/null"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/require"
)

func createRandomProductItem(t *testing.T) ProductItem {
	product := createRandomProduct(t)
	arg := CreateProductItemParams{
		ProductID:    product.ID,
		ProductSku:   util.RandomInt(100, 300),
		QtyInStock:   int32(util.RandomInt(0, 100)),
		ProductImage: util.RandomURL(),
		Price:        util.RandomDecimalString(1, 100),
		Active:       true,
	}

	productItem, err := testQueires.CreateProductItem(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, productItem)

	require.Equal(t, arg.ProductID, productItem.ProductID)
	require.Equal(t, arg.ProductSku, productItem.ProductSku)
	require.Equal(t, arg.QtyInStock, productItem.QtyInStock)
	require.Equal(t, arg.ProductImage, productItem.ProductImage)
	require.Equal(t, arg.Price, productItem.Price)
	require.Equal(t, arg.Active, productItem.Active)
	require.NotEmpty(t, productItem.CreatedAt)
	require.True(t, productItem.UpdatedAt.IsZero())
	require.True(t, productItem.Active)

	return productItem
}
func TestCreateProductItem(t *testing.T) {
	createRandomProductItem(t)
}

func TestGetProductItem(t *testing.T) {
	productItem1 := createRandomProductItem(t)

	productItem2, err := testQueires.GetProductItem(context.Background(), productItem1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, productItem2)

	require.Equal(t, productItem1.ProductID, productItem2.ProductID)
	require.Equal(t, productItem1.ProductSku, productItem2.ProductSku)
	require.Equal(t, productItem1.QtyInStock, productItem2.QtyInStock)
	require.Equal(t, productItem1.ProductImage, productItem2.ProductImage)
	require.Equal(t, productItem1.Price, productItem2.Price)
	require.Equal(t, productItem1.Active, productItem2.Active)
	require.Equal(t, productItem1.CreatedAt, productItem2.CreatedAt)
	require.Equal(t, productItem1.UpdatedAt, productItem2.UpdatedAt)
	require.True(t, productItem2.Active)

}

func TestUpdateProductItemQtyAndPriceAndActive(t *testing.T) {
	productItem1 := createRandomProductItem(t)
	arg := UpdateProductItemParams{
		ProductID:    productItem1.ProductID,
		ProductSku:   null.Int{},
		QtyInStock:   null.IntFrom(util.RandomInt(1, 500)),
		ProductImage: null.String{},
		Price:        null.StringFrom(util.RandomDecimalString(1, 100)),
		Active:       null.BoolFrom(!productItem1.Active),
		ID:           productItem1.ID,
	}

	productItem2, err := testQueires.UpdateProductItem(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, productItem2)

	require.Equal(t, productItem1.ProductID, productItem2.ProductID)
	require.Equal(t, productItem1.ProductSku, productItem2.ProductSku)
	require.NotEqual(t, productItem1.QtyInStock, productItem2.QtyInStock)
	require.Equal(t, productItem1.ProductImage, productItem2.ProductImage)
	require.NotEqual(t, productItem1.Price, productItem2.Price)
	require.NotEqual(t, productItem1.Active, productItem2.Active)
	require.False(t, productItem2.Active)
	require.WithinDuration(t, productItem1.CreatedAt, productItem2.CreatedAt, time.Second)
	require.NotEqual(t, productItem1.UpdatedAt, productItem2.UpdatedAt)
}

func TestDeleteProductItem(t *testing.T) {
	productItem1 := createRandomProductItem(t)
	err := testQueires.DeleteProductItem(context.Background(), productItem1.ID)

	require.NoError(t, err)

	productItem2, err := testQueires.GetProductItem(context.Background(), productItem1.ID)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, productItem2)

}

func TestListProductItems(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomProductItem(t)
	}
	arg := ListProductItemsParams{
		Limit:  5,
		Offset: 5,
	}

	productItems, err := testQueires.ListProductItems(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, productItems)

	for _, productItem := range productItems {
		require.NotEmpty(t, productItem)
	}

}

func TestListProductItemsV2(t *testing.T) {
	for i := 0; i < 30; i++ {
		createRandomProductItem(t)
	}

	initialSearchResult, err := testQueires.ListProductItemsV2(context.Background(), 10)

	require.NoError(t, err)
	require.NotEmpty(t, initialSearchResult)
	require.Equal(t, len(initialSearchResult), 10)

	arg1 := ListProductItemsNextPageParams{
		Limit: 10,
		ID:    initialSearchResult[len(initialSearchResult)-1].ID,
	}

	secondPage, err := testQueires.ListProductItemsNextPage(context.Background(), arg1)
	require.NoError(t, err)
	require.NotEmpty(t, secondPage)
	require.Equal(t, len(initialSearchResult), 10)
	require.Greater(t, initialSearchResult[len(initialSearchResult)-1].ID, secondPage[len(secondPage)-1].ID)

	arg2 := ListProductItemsNextPageParams{
		Limit: 10,
		ID:    secondPage[len(initialSearchResult)-1].ID,
	}

	thirdPage, err := testQueires.ListProductItemsNextPage(context.Background(), arg2)
	require.NoError(t, err)
	require.NotEmpty(t, secondPage)
	require.Equal(t, len(initialSearchResult), 10)
	require.Greater(t, secondPage[len(initialSearchResult)-1].ID, thirdPage[len(secondPage)-1].ID)
	require.Greater(t, initialSearchResult[len(initialSearchResult)-1].ID, thirdPage[len(secondPage)-1].ID)
}

func TestSearchProductItems(t *testing.T) {

	productItem := createRandomProductItem(t)

	product, err := testQueires.GetProduct(context.Background(), productItem.ProductID)

	require.NoError(t, err)
	require.NotEmpty(t, product)

	arg1 := SearchProductItemsParams{
		Limit: 10,
		Query: product.Name,
	}

	searchedProductItem, err := testQueires.SearchProductItems(context.Background(), arg1)

	require.NoError(t, err)
	require.NotEmpty(t, searchedProductItem)
	require.Equal(t, productItem.ID, searchedProductItem[len(searchedProductItem)-1].ID)

	arg2 := SearchProductItemsNextPageParams{
		Limit: 10,
		ID:    searchedProductItem[len(searchedProductItem)-1].ID,
		Query: product.Name,
	}

	searchedRestProductItem, err := testQueires.SearchProductItemsNextPage(context.Background(), arg2)

	require.NoError(t, err)
	require.Empty(t, searchedRestProductItem)
	// require.Equal(t, productItem.ID, searchedProductItem[len(searchedProductItem)-1].ID)
}
