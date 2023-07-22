package db

import (
	"context"
	"testing"

	"github.com/cshop/v3/util"
	"github.com/guregu/null"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func createRandomVariation(t *testing.T) Variation {
	category := createRandomProductCategory(t)
	arg := CreateVariationParams{
		CategoryID: category.ID,
		Name:       util.RandomString(5),
	}

	variation, err := testStore.CreateVariation(context.Background(), arg)
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
	t.Parallel()
	variation1 := createRandomVariation(t)
	variation2, err := testStore.GetVariation(context.Background(), variation1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, variation2)

	require.Equal(t, variation1.ID, variation2.ID)
	require.Equal(t, variation1.CategoryID, variation2.CategoryID)
	require.Equal(t, variation1.Name, variation2.Name)
}

func TestUpdateVariationNameAndCategoryID(t *testing.T) {
	t.Parallel()
	variation1 := createRandomVariation(t)
	// category := createRandomProductCategory(t)
	arg := UpdateVariationParams{
		ID:   variation1.ID,
		Name: null.StringFrom(util.RandomString(5)),
		// CategoryID: null.IntFromPtr(&category.ID),
		CategoryID: null.IntFromPtr(&variation1.CategoryID),
	}

	variation2, err := testStore.UpdateVariation(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, variation2)

	require.Equal(t, variation1.ID, variation2.ID)
	require.Equal(t, variation1.CategoryID, variation2.CategoryID)
	require.NotEqual(t, variation1.Name, variation2.Name)
}

func TestDeleteVariation(t *testing.T) {
	t.Parallel()
	variation1 := createRandomVariation(t)
	err := testStore.DeleteVariation(context.Background(), variation1.CategoryID)

	require.NoError(t, err)

	variation2, err := testStore.GetVariation(context.Background(), variation1.CategoryID)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, variation2)

}

func TestListVariations(t *testing.T) {
	t.Parallel()
	for i := 0; i < 10; i++ {
		createRandomVariation(t)
	}
	arg := ListVariationsParams{
		Limit:  5,
		Offset: 0,
	}

	variations, err := testStore.ListVariations(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, variations)

	for _, variation := range variations {
		require.NotEmpty(t, variation)
	}

}
