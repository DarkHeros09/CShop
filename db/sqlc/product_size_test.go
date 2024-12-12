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
	productItem := createRandomProductItem(t)
	arg := CreateProductSizeParams{
		ProductItemID: productItem.ID,
		SizeValue:     util.RandomSize(),
		Qty:           int32(util.RandomInt(1, 100)),
	}
	productSize, err := testStore.CreateProductSize(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, productSize)

	require.Equal(t, arg.SizeValue, productSize.SizeValue)

	return productSize
}

func createRandomProductSizeWithItemID(t *testing.T, itemID int64) ProductSize {
	// productItem := createRandomProductItem(t)
	arg := CreateProductSizeParams{
		ProductItemID: itemID,
		SizeValue:     util.RandomSize(),
		Qty:           int32(util.RandomInt(2, 100)),
	}
	productSize, err := testStore.CreateProductSize(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, productSize)

	require.Equal(t, arg.SizeValue, productSize.SizeValue)

	return productSize
}

func createRandomProductSizeWithQTY(t *testing.T, qty int32) ProductSize {
	productItem := createRandomProductItem(t)
	arg := CreateProductSizeParams{
		ProductItemID: productItem.ID,
		SizeValue:     util.RandomSize(),
		Qty:           qty,
	}
	productSize, err := testStore.CreateProductSize(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, productSize)

	require.Equal(t, arg.SizeValue, productSize.SizeValue)

	return productSize
}

func adminCreateRandomProductSize(t *testing.T) ProductSize {
	admin := createRandomAdmin(t)
	productItem := createRandomProductItem(t)
	arg := AdminCreateProductSizeParams{
		AdminID:       admin.ID,
		ProductItemID: productItem.ID,
		SizeValue:     util.RandomSize(),
		Qty:           int32(util.RandomInt(1, 100)),
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

func TestListProductSizesForProductItem(t *testing.T) {
	productItem := createRandomProductItem(t)
	for i := 0; i < 5; i++ {
		createRandomProductSizeWithItemID(t, productItem.ID)
	}
	productSize2, err := testStore.ListProductSizesByProductItemID(context.Background(), productItem.ID)

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
		ID:            productSize1.ID,
		ProductItemID: productSize1.ProductItemID,
		SizeValue:     updatedSize,
	}

	productSize2, err := testStore.UpdateProductSize(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, productSize2)

	require.Equal(t, productSize1.ID, productSize2.ID)
	require.NotEqual(t, productSize1.SizeValue, productSize2.SizeValue)
}

func TestAdminUpdateProductSize(t *testing.T) {
	admin := createRandomAdmin(t)
	productSize1 := createRandomProductSize(t)
	updatedSizeValue := util.RandomSize()

	if productSize1.SizeValue == updatedSizeValue {
		updatedSizeValue = "purple"
	}

	arg := AdminUpdateProductSizeParams{
		AdminID:       admin.ID,
		ID:            productSize1.ID,
		SizeValue:     null.StringFrom(updatedSizeValue),
		ProductItemID: productSize1.ProductItemID,
	}

	productSize2, err := testStore.AdminUpdateProductSize(context.Background(), arg)
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

func TestDeleteProductSizeByProductItemID(t *testing.T) {
	productSize1 := createRandomProductSize(t)
	err := testStore.DeleteProductSizeByProductItemID(context.Background(), productSize1.ProductItemID)

	require.NoError(t, err)

	productSize2, err := testStore.GetProductSize(context.Background(), productSize1.ID)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, productSize2)

}
