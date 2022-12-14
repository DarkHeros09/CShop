package db

import (
	"context"
	"testing"

	"github.com/guregu/null"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/require"
)

func createRandomProductConfiguration(t *testing.T) ProductConfiguration {
	productItem := createRandomProductItem(t)
	variationOption := createRandomVariationOption(t)
	arg := CreateProductConfigurationParams{
		ProductItemID:     productItem.ID,
		VariationOptionID: variationOption.ID,
	}

	productConfiguration, err := testQueires.CreateProductConfiguration(context.Background(), arg)
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
	productConfiguration2, err := testQueires.GetProductConfiguration(context.Background(), productConfiguration1.ProductItemID)

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

	productConfiguration2, err := testQueires.UpdateProductConfiguration(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, productConfiguration2)

	require.Equal(t, productConfiguration1.ProductItemID, productConfiguration2.ProductItemID)
	require.NotEqual(t, productConfiguration1.VariationOptionID, productConfiguration2.VariationOptionID)
}

func TestDeleteProductConfiguration(t *testing.T) {
	productConfiguration1 := createRandomProductConfiguration(t)
	err := testQueires.DeleteProductConfiguration(context.Background(), productConfiguration1.ProductItemID)

	require.NoError(t, err)

	productConfiguration2, err := testQueires.GetProductConfiguration(context.Background(), productConfiguration1.ProductItemID)

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

	productConfigurations, err := testQueires.ListProductConfigurations(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, productConfigurations)

	for _, productConfiguration := range productConfigurations {
		require.NotEmpty(t, productConfiguration)
	}

}
