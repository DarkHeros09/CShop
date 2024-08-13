package db

import (
	"context"
	"testing"

	"github.com/cshop/v3/util"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func createRandomProductSize(t *testing.T) ProductSize {
	arg := util.RandomSize()
	productSize, err := testStore.CreateProductSize(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, productSize)

	require.Equal(t, arg, productSize.SizeValue)

	return productSize
}

func adminCreateRandomProductSize(t *testing.T) ProductSize {
	admin := createRandomAdmin(t)
	arg := AdminCreateProductSizeParams{
		AdminID:   admin.ID,
		SizeValue: util.RandomSize(),
	}
	productSize, err := testStore.AdminCreateProductSize(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, productSize)

	require.Equal(t, arg.SizeValue, productSize.SizeValue)

	return productSize
}
func TestCreateProductSize(t *testing.T) {
	createRandomProductSize(t)
}
func TestAdminCreateProductSize(t *testing.T) {
	adminCreateRandomProductSize(t)
}

func TestGetProductSize(t *testing.T) {
	productSize1 := createRandomProductSize(t)
	productSize2, err := testStore.GetProductSize(context.Background(), productSize1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, productSize2)

	require.Equal(t, productSize1.ID, productSize2.ID)
	require.Equal(t, productSize1.SizeValue, productSize2.SizeValue)
}

func TestListProductSizes(t *testing.T) {
	for i := 0; i < 5; i++ {
		createRandomProductSize(t)
	}
	productSize2, err := testStore.ListProductSizes(context.Background())

	require.NoError(t, err)
	require.NotEmpty(t, productSize2)
}

func TestUpdateProductSize(t *testing.T) {
	productSize1 := createRandomProductSize(t)
	updatedSize := null.StringFrom(util.RandomSize())
	if updatedSize.String == productSize1.SizeValue {
		updatedSize = null.StringFrom("XS")
	}
	arg := UpdateProductSizeParams{
		ID:        productSize1.ID,
		SizeValue: updatedSize,
	}

	productSize2, err := testStore.UpdateProductSize(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, productSize2)

	require.Equal(t, productSize1.ID, productSize2.ID)
	require.NotEqual(t, productSize1.SizeValue, productSize2.SizeValue)
}

func TestDeleteProductSize(t *testing.T) {
	productSize1 := createRandomProductSize(t)
	err := testStore.DeleteProductSize(context.Background(), productSize1.ID)

	require.NoError(t, err)

	productSize2, err := testStore.GetProductSize(context.Background(), productSize1.ID)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, productSize2)

}
