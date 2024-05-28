package db

import (
	"context"
	"testing"

	"github.com/cshop/v3/util"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func createRandomVariationOption(t *testing.T) VariationOption {
	variation := createRandomVariation(t)
	arg := CreateVariationOptionParams{
		VariationID: null.IntFrom(variation.ID),
		Value:       util.RandomString(5),
	}

	variationOption, err := testStore.CreateVariationOption(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, variation)

	require.Equal(t, arg.VariationID, variationOption.VariationID)
	require.Equal(t, arg.Value, variationOption.Value)

	return variationOption
}
func TestCreateVariationOption(t *testing.T) {
	createRandomVariationOption(t)
}

func TestGetVariationOption(t *testing.T) {
	variationOption1 := createRandomVariationOption(t)
	variationOption2, err := testStore.GetVariationOption(context.Background(), variationOption1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, variationOption2)

	require.Equal(t, variationOption1.ID, variationOption2.ID)
	require.Equal(t, variationOption1.VariationID, variationOption2.VariationID)
	require.Equal(t, variationOption1.Value, variationOption2.Value)
}

func TestUpdateVariationOptionValue(t *testing.T) {
	variationOption1 := createRandomVariationOption(t)
	arg := UpdateVariationOptionParams{
		ID:          variationOption1.ID,
		Value:       null.StringFrom(util.RandomString(5)),
		VariationID: null.Int{},
	}

	variationOption2, err := testStore.UpdateVariationOption(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, variationOption2)

	require.Equal(t, variationOption1.ID, variationOption2.ID)
	require.Equal(t, variationOption1.VariationID, variationOption2.VariationID)
	require.NotEqual(t, variationOption1.Value, variationOption2.Value)
}

func TestDeleteVariationOption(t *testing.T) {
	variationOption1 := createRandomVariationOption(t)
	err := testStore.DeleteVariationOption(context.Background(), variationOption1.ID)

	require.NoError(t, err)

	variationOption2, err := testStore.GetVariationOption(context.Background(), variationOption1.ID)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, variationOption2)

}

func TestListVariationOptions(t *testing.T) {
	t.Parallel()
	for i := 0; i < 10; i++ {
		createRandomVariationOption(t)
	}
	arg := ListVariationOptionsParams{
		Limit:  5,
		Offset: 0,
	}

	variationOptions, err := testStore.ListVariationOptions(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, variationOptions)

	for _, variationOption := range variationOptions {
		require.NotEmpty(t, variationOption)
	}

}
