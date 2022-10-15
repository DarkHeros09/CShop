package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/cshop/v3/util"
	"github.com/stretchr/testify/require"
)

func createRandomCategoryPromotion(t *testing.T) CategoryPromotion {
	category := createRandomProductCategory(t)
	promotion := createRandomPromotion(t)
	arg := CreateCategoryPromotionParams{
		CategoryID:  category.ID,
		PromotionID: promotion.ID,
		Active:      util.RandomBool(),
	}

	CategoryPromotion, err := testQueires.CreateCategoryPromotion(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, CategoryPromotion)

	require.Equal(t, arg.CategoryID, CategoryPromotion.CategoryID)
	require.Equal(t, arg.PromotionID, CategoryPromotion.PromotionID)
	require.Equal(t, arg.Active, CategoryPromotion.Active)

	return CategoryPromotion
}
func TestCreateCategoryPromotion(t *testing.T) {
	createRandomCategoryPromotion(t)
}

func TestGetCategoryPromotion(t *testing.T) {
	CategoryPromotion1 := createRandomCategoryPromotion(t)
	CategoryPromotion2, err := testQueires.GetCategoryPromotion(context.Background(), CategoryPromotion1.CategoryID)

	require.NoError(t, err)
	require.NotEmpty(t, CategoryPromotion2)

	require.Equal(t, CategoryPromotion1.CategoryID, CategoryPromotion2.CategoryID)
	require.Equal(t, CategoryPromotion1.PromotionID, CategoryPromotion2.PromotionID)
	require.Equal(t, CategoryPromotion1.Active, CategoryPromotion2.Active)
}

func TestUpdateCategoryPromotionActive(t *testing.T) {
	CategoryPromotion1 := createRandomCategoryPromotion(t)
	arg := UpdateCategoryPromotionParams{
		PromotionID: sql.NullInt64{},
		Active:      sql.NullBool{Bool: !CategoryPromotion1.Active, Valid: true},
		CategoryID:  CategoryPromotion1.CategoryID,
	}

	CategoryPromotion2, err := testQueires.UpdateCategoryPromotion(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, CategoryPromotion2)

	require.Equal(t, CategoryPromotion1.CategoryID, CategoryPromotion2.CategoryID)
	require.Equal(t, CategoryPromotion1.PromotionID, CategoryPromotion2.PromotionID)
	require.NotEqual(t, CategoryPromotion1.Active, CategoryPromotion2.Active)
}

func TestDeleteCategoryPromotion(t *testing.T) {
	CategoryPromotion1 := createRandomCategoryPromotion(t)
	err := testQueires.DeleteCategoryPromotion(context.Background(), CategoryPromotion1.CategoryID)

	require.NoError(t, err)

	CategoryPromotion2, err := testQueires.GetCategoryPromotion(context.Background(), CategoryPromotion1.CategoryID)

	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, CategoryPromotion2)

}

func TestListCategoryPromotions(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomCategoryPromotion(t)
	}
	arg := ListCategoryPromotionsParams{
		Limit:  5,
		Offset: 5,
	}

	CategoryPromotions, err := testQueires.ListCategoryPromotions(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, CategoryPromotions)

	for _, CategoryPromotion := range CategoryPromotions {
		require.NotEmpty(t, CategoryPromotion)
	}

}
