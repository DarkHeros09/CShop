package db

import (
	"context"
	"testing"

	"github.com/guregu/null"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func createRandomProductConfiguration(t *testing.T) ProductConfiguration {
	productItem := createRandomProductItem(t)
	variationOption := createRandomVariationOption(t)
	arg := CreateProductConfigurationParams{
		ProductItemID:     productItem.ID,
		VariationOptionID: variationOption.ID,
	}

	productConfiguration, err := testStore.CreateProductConfiguration(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, productConfiguration)

	require.Equal(t, arg.ProductItemID, productConfiguration.ProductItemID)
	require.Equal(t, arg.VariationOptionID, productConfiguration.VariationOptionID)

	return productConfiguration
}
func TestCreateProductConfiguration(t *testing.T) {
	createRandomProductConfiguration(t)
}

func TestGetProductConfiguration(t *testing.T) {
	productConfiguration1 := createRandomProductConfiguration(t)

	arg := GetProductConfigurationParams{
		ProductItemID:     productConfiguration1.ProductItemID,
		VariationOptionID: productConfiguration1.VariationOptionID,
	}
	productConfiguration2, err := testStore.GetProductConfiguration(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, productConfiguration2)

	require.Equal(t, productConfiguration1.ProductItemID, productConfiguration2.ProductItemID)
	require.Equal(t, productConfiguration1.VariationOptionID, productConfiguration2.VariationOptionID)
}

func TestUpdateProductConfiguration(t *testing.T) {
	productConfiguration1 := createRandomProductConfiguration(t)
	variationOption := createRandomVariationOption(t)
	arg := UpdateProductConfigurationParams{
		VariationOptionID: null.IntFromPtr(&variationOption.ID),
		ProductItemID:     productConfiguration1.ProductItemID,
	}

	productConfiguration2, err := testStore.UpdateProductConfiguration(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, productConfiguration2)

	require.Equal(t, productConfiguration1.ProductItemID, productConfiguration2.ProductItemID)
	require.NotEqual(t, productConfiguration1.VariationOptionID, productConfiguration2.VariationOptionID)
}

func TestDeleteProductConfiguration(t *testing.T) {
	productConfiguration1 := createRandomProductConfiguration(t)
	arg1 := DeleteProductConfigurationParams{
		ProductItemID:     productConfiguration1.ProductItemID,
		VariationOptionID: productConfiguration1.VariationOptionID,
	}
	err := testStore.DeleteProductConfiguration(context.Background(), arg1)

	require.NoError(t, err)
	arg := GetProductConfigurationParams{
		ProductItemID:     productConfiguration1.ProductItemID,
		VariationOptionID: productConfiguration1.VariationOptionID,
	}
	productConfiguration2, err := testStore.GetProductConfiguration(context.Background(), arg)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, productConfiguration2)

}

func TestListProductConfigurations(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomProductConfiguration(t)
	}
	arg := ListProductConfigurationsParams{
		Limit:  5,
		Offset: 5,
	}

	productConfigurations, err := testStore.ListProductConfigurations(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, productConfigurations)

	for _, productConfiguration := range productConfigurations {
		require.NotEmpty(t, productConfiguration)
	}

}
