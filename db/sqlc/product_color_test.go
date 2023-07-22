package db

import (
	"context"
	"testing"

	"github.com/cshop/v3/util"
	"github.com/guregu/null"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func createRandomProductColor(t *testing.T) ProductColor {
	arg := util.RandomColor()
	productColor, err := testStore.CreateProductColor(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, productColor)

	require.Equal(t, arg, productColor.ColorValue)

	return productColor
}
func TestCreateProductColor(t *testing.T) {
	createRandomProductColor(t)
}

func TestGetProductColor(t *testing.T) {
	productColor1 := createRandomProductColor(t)
	productColor2, err := testStore.GetProductColor(context.Background(), productColor1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, productColor2)

	require.Equal(t, productColor1.ID, productColor2.ID)
	require.Equal(t, productColor1.ColorValue, productColor2.ColorValue)
}

func TestUpdateProductColor(t *testing.T) {
	productColor1 := createRandomProductColor(t)
	updatedColorValue := util.RandomColor()

	if productColor1.ColorValue == updatedColorValue {
		updatedColorValue = "purple"
	}

	arg := UpdateProductColorParams{
		ID:         productColor1.ID,
		ColorValue: null.StringFrom(updatedColorValue),
	}

	productColor2, err := testStore.UpdateProductColor(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, productColor2)

	require.Equal(t, productColor1.ID, productColor2.ID)
	require.NotEqual(t, productColor1.ColorValue, productColor2.ColorValue)
}

func TestDeleteProductColor(t *testing.T) {
	productColor1 := createRandomProductColor(t)
	err := testStore.DeleteProductColor(context.Background(), productColor1.ID)

	require.NoError(t, err)

	productColor2, err := testStore.GetProductColor(context.Background(), productColor1.ID)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, productColor2)

}
