package db

import (
	"context"
	"testing"

	"github.com/cshop/v3/util"
	"github.com/guregu/null"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/require"
)

func createRandomProductPromotion(t *testing.T) ProductPromotion {
	product := createRandomProduct(t)
	promotion := createRandomPromotion(t)
	arg := CreateProductPromotionParams{
		ProductID:   product.ID,
		PromotionID: promotion.ID,
		Active:      util.RandomBool(),
	}

	productPromotion, err := testQueires.CreateProductPromotion(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, productPromotion)

	require.Equal(t, arg.ProductID, productPromotion.ProductID)
	require.Equal(t, arg.PromotionID, productPromotion.PromotionID)
	require.Equal(t, arg.Active, productPromotion.Active)

	return productPromotion
}
func TestCreateProductPromotion(t *testing.T) {
	createRandomProductPromotion(t)
}

func TestGetProductPromotion(t *testing.T) {
	productPromotion1 := createRandomProductPromotion(t)
	productPromotion2, err := testQueires.GetProductPromotion(context.Background(), productPromotion1.ProductID)

	require.NoError(t, err)
	require.NotEmpty(t, productPromotion2)

	require.Equal(t, productPromotion1.ProductID, productPromotion2.ProductID)
	require.Equal(t, productPromotion1.PromotionID, productPromotion2.PromotionID)
	require.Equal(t, productPromotion1.Active, productPromotion2.Active)
}

func TestUpdateProductPromotionActive(t *testing.T) {
	productPromotion1 := createRandomProductPromotion(t)
	arg := UpdateProductPromotionParams{
		PromotionID: null.Int{},
		Active:      null.BoolFrom(!productPromotion1.Active),
		ProductID:   productPromotion1.ProductID,
	}

	productPromotion2, err := testQueires.UpdateProductPromotion(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, productPromotion2)

	require.Equal(t, productPromotion1.ProductID, productPromotion2.ProductID)
	require.Equal(t, productPromotion1.PromotionID, productPromotion2.PromotionID)
	require.NotEqual(t, productPromotion1.Active, productPromotion2.Active)
}

func TestDeleteProductPromotion(t *testing.T) {
	productPromotion1 := createRandomProductPromotion(t)
	err := testQueires.DeleteProductPromotion(context.Background(), productPromotion1.ProductID)

	require.NoError(t, err)

	productPromotion2, err := testQueires.GetProductPromotion(context.Background(), productPromotion1.ProductID)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, productPromotion2)

}

func TestListProductPromotions(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomProductPromotion(t)
	}
	arg := ListProductPromotionsParams{
		Limit:  5,
		Offset: 5,
	}

	productPromotions, err := testQueires.ListProductPromotions(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, productPromotions)

	for _, productPromotion := range productPromotions {
		require.NotEmpty(t, productPromotion)
	}

}
