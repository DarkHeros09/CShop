package db

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/cshop/v3/util"
	"github.com/stretchr/testify/require"
)

func createRandomProductItem(t *testing.T) ProductItem {
	product := createRandomProduct(t)
	arg := CreateProductItemParams{
		ProductID:    product.ID,
		ProductSku:   util.RandomInt(100, 300),
		QtyInStock:   int32(util.RandomInt(0, 100)),
		ProductImage: util.RandomString(5),
		Price:        fmt.Sprint(util.RandomMoney()),
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
		ProductID:  sql.NullInt64{},
		ProductSku: sql.NullInt64{},
		QtyInStock: sql.NullInt32{
			Int32: int32(util.RandomInt(5, 90)),
			Valid: true,
		},
		ProductImage: sql.NullString{},
		Price: sql.NullString{
			String: fmt.Sprint(util.RandomMoney()),
			Valid:  true,
		},
		Active: sql.NullBool{
			Bool:  !productItem1.Active,
			Valid: true,
		},
		ID: productItem1.ID,
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
	require.EqualError(t, err, sql.ErrNoRows.Error())
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
