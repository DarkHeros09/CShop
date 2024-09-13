package db

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/cshop/v3/util"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func createRandomPromotion(t *testing.T) Promotion {
	arg := CreatePromotionParams{
		Name:         util.RandomString(6),
		Description:  util.RandomString(6),
		DiscountRate: util.RandomInt(1, 90),
		Active:       util.RandomBool(),
		StartDate:    util.RandomStartDate(),
		EndDate:      util.RandomEndDate(),
	}

	promotion, err := testStore.CreatePromotion(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, promotion)

	require.Equal(t, arg.Name, promotion.Name)
	require.Equal(t, arg.Description, promotion.Description)
	require.Equal(t, arg.DiscountRate, promotion.DiscountRate)
	require.Equal(t, arg.Active, promotion.Active)
	require.WithinDuration(t, arg.StartDate, promotion.StartDate, time.Second)
	require.WithinDuration(t, arg.EndDate, promotion.EndDate, time.Second)

	return promotion
}
func adminCreateRandomPromotion(t *testing.T) Promotion {
	admin := createRandomAdmin(t)
	arg := AdminCreatePromotionParams{
		AdminID:      admin.ID,
		Name:         util.RandomString(6),
		Description:  util.RandomString(6),
		DiscountRate: util.RandomInt(1, 90),
		Active:       util.RandomBool(),
		StartDate:    util.RandomStartDate(),
		EndDate:      util.RandomEndDate(),
	}

	promotion, err := testStore.AdminCreatePromotion(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, promotion)

	require.Equal(t, arg.Name, promotion.Name)
	require.Equal(t, arg.Description, promotion.Description)
	require.Equal(t, arg.DiscountRate, promotion.DiscountRate)
	require.Equal(t, arg.Active, promotion.Active)
	require.WithinDuration(t, arg.StartDate, promotion.StartDate, time.Second)
	require.WithinDuration(t, arg.EndDate, promotion.EndDate, time.Second)

	return promotion
}
func TestCreatePromotion(t *testing.T) {
	createRandomPromotion(t)
}
func TestAdminCreatePromotion(t *testing.T) {
	adminCreateRandomPromotion(t)
}

func TestGetPromotion(t *testing.T) {
	promotion1 := createRandomPromotion(t)
	promotion2, err := testStore.GetPromotion(context.Background(), promotion1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, promotion2)

	require.Equal(t, promotion1.ID, promotion2.ID)
	require.Equal(t, promotion1.Name, promotion2.Name)
	require.Equal(t, promotion1.Description, promotion2.Description)
	require.Equal(t, promotion1.DiscountRate, promotion2.DiscountRate)
	require.Equal(t, promotion1.Active, promotion2.Active)
	require.WithinDuration(t, promotion1.StartDate, promotion2.StartDate, time.Second)
	require.WithinDuration(t, promotion1.EndDate, promotion2.EndDate, time.Second)
}

func TestUpdatePromotionName(t *testing.T) {
	promotion1 := createRandomPromotion(t)
	arg := UpdatePromotionParams{
		Name:         null.StringFrom(util.RandomString(5)),
		Description:  null.String{},
		DiscountRate: null.Int{},
		Active:       null.Bool{},
		StartDate:    null.Time{},
		EndDate:      null.Time{},
		ID:           promotion1.ID,
	}

	promotion2, err := testStore.UpdatePromotion(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, promotion2)

	require.Equal(t, promotion1.ID, promotion2.ID)
	require.NotEqual(t, promotion1.Name, promotion2.Name)
	require.Equal(t, promotion1.Description, promotion2.Description)
	require.Equal(t, promotion1.DiscountRate, promotion2.DiscountRate)
	require.Equal(t, promotion1.Active, promotion2.Active)
	require.WithinDuration(t, promotion1.StartDate, promotion2.StartDate, time.Second)
	require.WithinDuration(t, promotion1.EndDate, promotion2.EndDate, time.Second)
	require.NotEqual(t, promotion1.StartDate, promotion2.EndDate)
}

func TestUpdatePromotionDiscriptionAndDiscountRate(t *testing.T) {
	promotion1 := createRandomPromotion(t)
	arg := UpdatePromotionParams{
		Description:  null.StringFrom(util.RandomString(5)),
		DiscountRate: null.IntFrom(util.RandomInt(1, 99)),
		ID:           promotion1.ID,
	}

	promotion2, err := testStore.UpdatePromotion(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, promotion2)

	require.Equal(t, promotion1.ID, promotion2.ID)
	require.Equal(t, promotion1.Name, promotion2.Name)
	require.NotEqual(t, promotion1.Description, promotion2.Description)
	require.NotEqual(t, promotion1.DiscountRate, promotion2.DiscountRate)
	require.Equal(t, promotion1.Active, promotion2.Active)
	require.WithinDuration(t, promotion1.StartDate, promotion2.StartDate, time.Second)
	require.WithinDuration(t, promotion1.EndDate, promotion2.EndDate, time.Second)
	require.NotEqual(t, promotion1.StartDate, promotion2.EndDate)

}

func TestAdminUpdatePromotion(t *testing.T) {
	admin := createRandomAdmin(t)
	promotion1 := createRandomPromotion(t)
	arg := AdminUpdatePromotionParams{
		AdminID:      admin.ID,
		Description:  null.StringFrom(util.RandomString(5)),
		DiscountRate: null.IntFrom(util.RandomInt(1, 99)),
		ID:           promotion1.ID,
	}

	promotion2, err := testStore.AdminUpdatePromotion(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, promotion2)

	require.Equal(t, promotion1.ID, promotion2.ID)
	require.Equal(t, promotion1.Name, promotion2.Name)
	require.NotEqual(t, promotion1.Description, promotion2.Description)
	require.NotEqual(t, promotion1.DiscountRate, promotion2.DiscountRate)
	require.Equal(t, promotion1.Active, promotion2.Active)
	require.WithinDuration(t, promotion1.StartDate, promotion2.StartDate, time.Second)
	require.WithinDuration(t, promotion1.EndDate, promotion2.EndDate, time.Second)
	require.NotEqual(t, promotion1.StartDate, promotion2.EndDate)

}

func TestDeletePromotion(t *testing.T) {
	promotion1 := createRandomPromotion(t)
	err := testStore.DeletePromotion(context.Background(), promotion1.ID)

	require.NoError(t, err)

	promotion2, err := testStore.GetPromotion(context.Background(), promotion1.ID)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, promotion2)

}

func TestListPromotions(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			createRandomPromotion(t)
			wg.Done()
		}()
	}

	wg.Wait()
	// arg := ListPromotionsParams{
	// 	Limit:  5,
	// 	Offset: 5,
	// }

	promotions, err := testStore.ListPromotions(context.Background())

	require.NoError(t, err)
	require.NotEmpty(t, promotions)

	for _, promotion := range promotions {
		require.NotEmpty(t, promotion)
	}

}
