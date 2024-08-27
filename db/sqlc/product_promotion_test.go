package db

import (
	"context"
	"testing"

	"github.com/cshop/v3/util"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func createRandomProductPromotion(t *testing.T) ProductPromotion {
	product := createRandomProduct(t)
	promotion := createRandomPromotion(t)
	arg := CreateProductPromotionParams{
		ProductID:             product.ID,
		PromotionID:           promotion.ID,
		ProductPromotionImage: null.StringFrom(util.RandomPromotionURL()),
		Active:                util.RandomBool(),
	}

	productPromotion, err := testStore.CreateProductPromotion(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, productPromotion)

	require.Equal(t, arg.ProductID, productPromotion.ProductID)
	require.Equal(t, arg.PromotionID, productPromotion.PromotionID)
	require.Equal(t, arg.Active, productPromotion.Active)

	return productPromotion
}

func adminCreateRandomProductPromotion(t *testing.T) ProductPromotion {
	admin := createRandomAdmin(t)
	product := createRandomProduct(t)
	promotion := createRandomPromotion(t)
	arg := AdminCreateProductPromotionParams{
		AdminID:               admin.ID,
		ProductID:             product.ID,
		PromotionID:           promotion.ID,
		ProductPromotionImage: null.StringFrom(util.RandomPromotionURL()),
		Active:                util.RandomBool(),
	}

	productPromotion, err := testStore.AdminCreateProductPromotion(context.Background(), arg)
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
func TestAdminCreateProductPromotion(t *testing.T) {
	adminCreateRandomProductPromotion(t)
}

func TestGetProductPromotion(t *testing.T) {
	productPromotion1 := createRandomProductPromotion(t)

	arg := GetProductPromotionParams{
		ProductID:   productPromotion1.ProductID,
		PromotionID: productPromotion1.PromotionID,
	}
	productPromotion2, err := testStore.GetProductPromotion(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, productPromotion2)

	require.Equal(t, productPromotion1.ProductID, productPromotion2.ProductID)
	require.Equal(t, productPromotion1.PromotionID, productPromotion2.PromotionID)
	require.Equal(t, productPromotion1.Active, productPromotion2.Active)
}

func TestUpdateProductPromotionActive(t *testing.T) {
	productPromotion1 := createRandomProductPromotion(t)
	arg := UpdateProductPromotionParams{
		PromotionID: productPromotion1.PromotionID,
		Active:      null.BoolFrom(!productPromotion1.Active),
		ProductID:   productPromotion1.ProductID,
	}

	productPromotion2, err := testStore.UpdateProductPromotion(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, productPromotion2)

	require.Equal(t, productPromotion1.ProductID, productPromotion2.ProductID)
	require.Equal(t, productPromotion1.PromotionID, productPromotion2.PromotionID)
	require.NotEqual(t, productPromotion1.Active, productPromotion2.Active)
}

func TestDeleteProductPromotion(t *testing.T) {
	productPromotion1 := createRandomProductPromotion(t)
	arg := DeleteProductPromotionParams{
		ProductID:   productPromotion1.ProductID,
		PromotionID: productPromotion1.PromotionID,
	}
	err := testStore.DeleteProductPromotion(context.Background(), arg)

	require.NoError(t, err)

	arg1 := GetProductPromotionParams{
		ProductID:   productPromotion1.ProductID,
		PromotionID: productPromotion1.PromotionID,
	}
	productPromotion2, err := testStore.GetProductPromotion(context.Background(), arg1)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, productPromotion2)

}

func TestListProductPromotions(t *testing.T) {
	for i := 0; i < 5; i++ {
		createRandomProductPromotion(t)
	}
	arg := ListProductPromotionsParams{
		Limit:  5,
		Offset: 0,
	}

	productPromotions, err := testStore.ListProductPromotions(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, productPromotions)

	for _, productPromotion := range productPromotions {
		require.NotEmpty(t, productPromotion)
	}

}

func TestListProductPromotionsWithImages(t *testing.T) {

	for i := 0; i < 5; i++ {
		createRandomProductPromotion(t)
	}

	ProductPromotions, err := testStore.ListProductPromotionsWithImages(context.Background())

	require.NoError(t, err)
	require.NotEmpty(t, ProductPromotions)

	for _, ProductPromotion := range ProductPromotions {
		require.NotEmpty(t, ProductPromotion)
	}

}
