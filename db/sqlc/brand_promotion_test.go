package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/cshop/v3/util"
	"github.com/guregu/null"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func createRandomBrandPromotion(t *testing.T) BrandPromotion {
	brand := createRandomProductBrand(t)
	promotion := createRandomPromotion(t)

	fmt.Println(brand.ID)
	fmt.Println(promotion.ID)
	arg := CreateBrandPromotionParams{
		BrandID:     brand.ID,
		PromotionID: promotion.ID,
		Active:      util.RandomBool(),
	}

	brandPromotion, err := testStore.CreateBrandPromotion(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, brandPromotion)

	require.Equal(t, arg.BrandID, brandPromotion.BrandID)
	require.Equal(t, arg.PromotionID, brandPromotion.PromotionID)
	require.Equal(t, arg.Active, brandPromotion.Active)

	return brandPromotion
}
func TestCreateBrandPromotion(t *testing.T) {
	createRandomBrandPromotion(t)
}

func TestGetBrandPromotion(t *testing.T) {
	brandPromotion1 := createRandomBrandPromotion(t)

	arg := GetBrandPromotionParams{
		BrandID:     brandPromotion1.BrandID,
		PromotionID: brandPromotion1.PromotionID,
	}
	brandPromotion2, err := testStore.GetBrandPromotion(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, brandPromotion2)

	require.Equal(t, brandPromotion1.BrandID, brandPromotion2.BrandID)
	require.Equal(t, brandPromotion1.PromotionID, brandPromotion2.PromotionID)
	require.Equal(t, brandPromotion1.Active, brandPromotion2.Active)
}

func TestUpdateBrandPromotionActive(t *testing.T) {
	// brandPromotionChan := make(chan BrandPromotion)
	// go func() {
	brandPromotion1 := createRandomBrandPromotion(t)

	// 	brandPromotionChan <- brandPromotion1

	// }()
	// brandPromotion1 := <-brandPromotionChan
	arg := UpdateBrandPromotionParams{
		PromotionID: brandPromotion1.PromotionID,
		Active:      null.BoolFrom(!brandPromotion1.Active),
		BrandID:     brandPromotion1.BrandID,
	}

	brandPromotion2, err := testStore.UpdateBrandPromotion(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, brandPromotion2)

	require.Equal(t, brandPromotion1.BrandID, brandPromotion2.BrandID)
	require.Equal(t, brandPromotion1.PromotionID, brandPromotion2.PromotionID)
	require.NotEqual(t, brandPromotion1.Active, brandPromotion2.Active)
}

func TestDeleteBrandPromotion(t *testing.T) {
	BrandPromotion1 := createRandomBrandPromotion(t)
	arg1 := DeleteBrandPromotionParams{
		BrandID:     BrandPromotion1.BrandID,
		PromotionID: BrandPromotion1.PromotionID,
	}
	err := testStore.DeleteBrandPromotion(context.Background(), arg1)

	require.NoError(t, err)

	arg := GetBrandPromotionParams{
		BrandID:     BrandPromotion1.BrandID,
		PromotionID: BrandPromotion1.PromotionID,
	}

	BrandPromotion2, err := testStore.GetBrandPromotion(context.Background(), arg)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, BrandPromotion2)

}

func TestListBrandPromotions(t *testing.T) {
	for i := 0; i < 5; i++ {
		createRandomBrandPromotion(t)
	}
	arg := ListBrandPromotionsParams{
		Limit:  5,
		Offset: 0,
	}

	BrandPromotions, err := testStore.ListBrandPromotions(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, BrandPromotions)

	for _, BrandPromotion := range BrandPromotions {
		require.NotEmpty(t, BrandPromotion)
	}

}
