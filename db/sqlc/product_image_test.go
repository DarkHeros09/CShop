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

func adminCreateRandomProductImages(t *testing.T) ProductImage {
	admin := createRandomAdmin(t)
	arg := AdminCreateProductImagesParams{
		AdminID:       admin.ID,
		ProductImage1: util.RandomURL(),
		ProductImage2: util.RandomURL(),
		ProductImage3: util.RandomURL(),
	}
	productImage, err := testStore.AdminCreateProductImages(context.Background(), arg)
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

func TestAdminCreateProductImage(t *testing.T) {
	adminCreateRandomProductImages(t)
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

func TestListProductImagesV2(t *testing.T) {
	for i := 0; i < 30; i++ {
		createRandomProductImage(t)
	}

	limit := 10

	initialSearchResult, err := testStore.ListProductImagesV2(context.Background(), int32(limit))
	// fmt.Println(initialSearchResult)
	require.NoError(t, err)
	require.NotEmpty(t, initialSearchResult)
	require.Equal(t, len(initialSearchResult), 10)

	arg1 := ListProductImagesNextPageParams{
		Limit: 10,
		ID:    initialSearchResult[len(initialSearchResult)-1].ID,
	}

	secondPage, err := testStore.ListProductImagesNextPage(context.Background(), arg1)
	// fmt.Println(secondPage)
	require.NoError(t, err)
	require.NotEmpty(t, secondPage)
	require.Equal(t, len(secondPage), 10)
	require.Greater(t, initialSearchResult[len(initialSearchResult)-1].ID, secondPage[len(secondPage)-1].ID)

	arg2 := ListProductImagesNextPageParams{
		Limit: 10,
		ID:    secondPage[len(secondPage)-1].ID,
	}

	thirdPage, err := testStore.ListProductImagesNextPage(context.Background(), arg2)
	require.NoError(t, err)
	require.NotEmpty(t, secondPage)
	require.Equal(t, len(initialSearchResult), 10)
	require.Greater(t, secondPage[len(secondPage)-1].ID, thirdPage[len(thirdPage)-1].ID)
	require.Greater(t, initialSearchResult[len(initialSearchResult)-1].ID, thirdPage[len(thirdPage)-1].ID)
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
