package db

import (
	"context"
	"database/sql"
	"testing"

	"cshop.com/v2/util"
	"github.com/stretchr/testify/require"
)

func createRandomVariation(t *testing.T) Variation {
	category := createRandomProductCategory(t)
	arg := CreateVariationParams{
		CategoryID: category.ID,
		Name:       util.RandomString(5),
	}

	variation, err := testQueires.CreateVariation(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, variation)

	require.Equal(t, arg.CategoryID, variation.CategoryID)
	require.Equal(t, arg.Name, variation.Name)

	return variation
}
func TestCreateVariation(t *testing.T) {
	createRandomVariation(t)
}

func TestGetVariation(t *testing.T) {
	variation1 := createRandomVariation(t)
	variation2, err := testQueires.GetVariation(context.Background(), variation1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, variation2)

	require.Equal(t, variation1.ID, variation2.ID)
	require.Equal(t, variation1.CategoryID, variation2.CategoryID)
	require.Equal(t, variation1.Name, variation2.Name)
}

func TestUpdateVariationNameAndCategoryID(t *testing.T) {
	variation1 := createRandomVariation(t)
	category := createRandomProductCategory(t)
	arg := UpdateVariationParams{
		ID: variation1.ID,
		Name: sql.NullString{
			String: util.RandomString(5),
			Valid:  true,
		},
		CategoryID: sql.NullInt64{
			Int64: category.ID,
			Valid: true,
		},
	}

	variation2, err := testQueires.UpdateVariation(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, variation2)

	require.Equal(t, variation1.ID, variation2.ID)
	require.NotEqual(t, variation1.CategoryID, variation2.CategoryID)
	require.NotEqual(t, variation1.Name, variation2.Name)
}

func TestDeleteVariation(t *testing.T) {
	variation1 := createRandomVariation(t)
	err := testQueires.DeleteVariation(context.Background(), variation1.CategoryID)

	require.NoError(t, err)

	variation2, err := testQueires.GetVariation(context.Background(), variation1.CategoryID)

	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, variation2)

}

func TestListVariations(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomVariation(t)
	}
	arg := ListVariationsParams{
		Limit:  5,
		Offset: 5,
	}

	variations, err := testQueires.ListVariations(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, variations)

	for _, variation := range variations {
		require.NotEmpty(t, variation)
	}

}
