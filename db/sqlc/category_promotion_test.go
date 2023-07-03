package db

import (
	"context"
	"testing"

	"github.com/cshop/v3/util"
	"github.com/guregu/null"
	"github.com/jackc/pgx/v4"
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

	arg := GetCategoryPromotionParams{
		CategoryID:  CategoryPromotion1.CategoryID,
		PromotionID: CategoryPromotion1.PromotionID,
	}
	CategoryPromotion2, err := testQueires.GetCategoryPromotion(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, CategoryPromotion2)

	require.Equal(t, CategoryPromotion1.CategoryID, CategoryPromotion2.CategoryID)
	require.Equal(t, CategoryPromotion1.PromotionID, CategoryPromotion2.PromotionID)
	require.Equal(t, CategoryPromotion1.Active, CategoryPromotion2.Active)
}

func TestUpdateCategoryPromotionActive(t *testing.T) {
	// categoryPromotionChan := make(chan CategoryPromotion)
	// go func() {
	categoryPromotion1 := createRandomCategoryPromotion(t)

	// 	categoryPromotionChan <- categoryPromotion1

	// }()
	// categoryPromotion1 := <-categoryPromotionChan
	arg := UpdateCategoryPromotionParams{
		PromotionID: categoryPromotion1.PromotionID,
		Active:      null.BoolFrom(!categoryPromotion1.Active),
		CategoryID:  categoryPromotion1.CategoryID,
	}

	categoryPromotion2, err := testQueires.UpdateCategoryPromotion(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, categoryPromotion2)

	require.Equal(t, categoryPromotion1.CategoryID, categoryPromotion2.CategoryID)
	require.Equal(t, categoryPromotion1.PromotionID, categoryPromotion2.PromotionID)
	require.NotEqual(t, categoryPromotion1.Active, categoryPromotion2.Active)
}

func TestDeleteCategoryPromotion(t *testing.T) {
	CategoryPromotion1 := createRandomCategoryPromotion(t)
	arg1 := DeleteCategoryPromotionParams{
		CategoryID:  CategoryPromotion1.CategoryID,
		PromotionID: CategoryPromotion1.PromotionID,
	}
	err := testQueires.DeleteCategoryPromotion(context.Background(), arg1)

	require.NoError(t, err)

	arg := GetCategoryPromotionParams{
		CategoryID:  CategoryPromotion1.CategoryID,
		PromotionID: CategoryPromotion1.PromotionID,
	}

	CategoryPromotion2, err := testQueires.GetCategoryPromotion(context.Background(), arg)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
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
