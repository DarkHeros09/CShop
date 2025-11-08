package db

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/cshop/v3/util"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func createRandomBrandPromotion(t *testing.T) BrandPromotion {
	brand := createRandomProductBrand(t)
	promotion := createRandomPromotion(t)

	fmt.Println(brand.ID)
	fmt.Println(promotion.ID)
	arg := CreateBrandPromotionParams{
		BrandID:             brand.ID,
		PromotionID:         promotion.ID,
		BrandPromotionImage: null.StringFrom(util.RandomPromotionURL()),
		Active:              util.RandomBool(),
	}

	brandPromotion, err := testStore.CreateBrandPromotion(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, brandPromotion)

	require.Equal(t, arg.BrandID, brandPromotion.BrandID)
	require.Equal(t, arg.PromotionID, brandPromotion.PromotionID)
	require.Equal(t, arg.Active, brandPromotion.Active)

	return *brandPromotion
}
func adminCreateRandomBrandPromotion(t *testing.T) BrandPromotion {
	admin := createRandomAdmin(t)
	brand := createRandomProductBrand(t)
	promotion := createRandomPromotion(t)

	fmt.Println(brand.ID)
	fmt.Println(promotion.ID)
	arg := AdminCreateBrandPromotionParams{
		AdminID:             admin.ID,
		BrandID:             brand.ID,
		PromotionID:         promotion.ID,
		BrandPromotionImage: null.StringFrom(util.RandomPromotionURL()),
		Active:              util.RandomBool(),
	}

	brandPromotion, err := testStore.AdminCreateBrandPromotion(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, brandPromotion)

	require.Equal(t, arg.BrandID, brandPromotion.BrandID)
	require.Equal(t, arg.PromotionID, brandPromotion.PromotionID)
	require.Equal(t, arg.Active, brandPromotion.Active)

	return *brandPromotion
}
func TestCreateBrandPromotion(t *testing.T) {
	createRandomBrandPromotion(t)
}

func TestAdminCreateBrandPromotion(t *testing.T) {
	adminCreateRandomBrandPromotion(t)
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

func TestAdminUpdateBrandPromotionActive(t *testing.T) {
	// brandPromotionChan := make(chan BrandPromotion)
	// go func() {
	admin := createRandomAdmin(t)
	brandPromotion1 := createRandomBrandPromotion(t)

	// 	brandPromotionChan <- brandPromotion1

	// }()
	// brandPromotion1 := <-brandPromotionChan
	arg := AdminUpdateBrandPromotionParams{
		AdminID:     admin.ID,
		PromotionID: brandPromotion1.PromotionID,
		Active:      null.BoolFrom(!brandPromotion1.Active),
		BrandID:     brandPromotion1.BrandID,
	}

	brandPromotion2, err := testStore.AdminUpdateBrandPromotion(context.Background(), arg)
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

func TestListBrandPromotionsWithImages(t *testing.T) {
	exit := false
	for {
		p := createRandomBrandPromotion(t)
		if p.Active && p.BrandPromotionImage.Valid {
			brandPromotions, _ := testStore.ListBrandPromotionsWithImages(context.Background())
			for _, brandPromotion := range brandPromotions {
				if brandPromotion.StartDate.Unix() <= time.Now().Unix() && brandPromotion.EndDate.Unix() >= time.Now().Unix() && len(brandPromotions) > 0 {
					exit = true
					break
				}
			}
			if exit {
				break
			}
		}
	}

	brandPromotions, err := testStore.ListBrandPromotionsWithImages(context.Background())

	require.NoError(t, err)
	require.NotEmpty(t, brandPromotions)

	for _, brandPromotion := range brandPromotions {
		require.NotEmpty(t, brandPromotion)
	}

}

func TestAdminListBrandPromotions(t *testing.T) {
	admin := createRandomAdmin(t)
	for i := 0; i < 5; i++ {
		createRandomBrandPromotion(t)
	}

	brandPromotions, err := testStore.AdminListBrandPromotions(context.Background(), admin.ID)

	require.NoError(t, err)
	require.NotEmpty(t, brandPromotions)

	for _, brandPromotion := range brandPromotions {
		require.NotEmpty(t, brandPromotion)
	}

}
