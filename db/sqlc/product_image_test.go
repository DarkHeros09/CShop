package db

import (
	"context"
	"testing"

	"github.com/cshop/v3/util"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func createRandomProductImage(t *testing.T) ProductImage {
	arg := CreateProductImageParams{
		ProductImage1: util.RandomURL(),
		ProductImage2: util.RandomURL(),
		ProductImage3: util.RandomURL(),
	}
	productImage, err := testStore.CreateProductImage(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, productImage)

	require.Equal(t, arg.ProductImage1, productImage.ProductImage1)
	require.Equal(t, arg.ProductImage2, productImage.ProductImage2)
	require.Equal(t, arg.ProductImage3, productImage.ProductImage3)

	return productImage
}
func TestCreateProductImage(t *testing.T) {
	createRandomProductImage(t)
}

func TestGetProductImage(t *testing.T) {
	productImage1 := createRandomProductImage(t)
	productImage2, err := testStore.GetProductImage(context.Background(), productImage1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, productImage2)

	require.Equal(t, productImage1.ID, productImage2.ID)
	require.Equal(t, productImage1.ProductImage1, productImage2.ProductImage1)
	require.Equal(t, productImage1.ProductImage2, productImage2.ProductImage2)
	require.Equal(t, productImage1.ProductImage3, productImage2.ProductImage3)
}

func TestUpdateProductImage(t *testing.T) {
	productImage1 := createRandomProductImage(t)
	arg := UpdateProductImageParams{
		ID:            productImage1.ID,
		ProductImage1: null.StringFrom(productImage1.ProductImage3),
		ProductImage2: null.StringFrom(productImage1.ProductImage1),
		ProductImage3: null.StringFrom(productImage1.ProductImage2),
	}

	productImage2, err := testStore.UpdateProductImage(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, productImage2)

	require.Equal(t, productImage1.ID, productImage2.ID)
	require.NotEqual(t, productImage1, productImage2)
	// require.NotEqual(t, productImage1.ProductImage2, productImage2.ProductImage2)
	// require.NotEqual(t, productImage1.ProductImage3, productImage2.ProductImage3)
}

func TestDeleteProductImage(t *testing.T) {
	productImage1 := createRandomProductImage(t)
	err := testStore.DeleteProductImage(context.Background(), productImage1.ID)

	require.NoError(t, err)

	productImage2, err := testStore.GetProductImage(context.Background(), productImage1.ID)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, productImage2)

}
