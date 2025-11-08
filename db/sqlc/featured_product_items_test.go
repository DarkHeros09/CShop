package db

import (
	"context"
	"testing"
	"time"

	"github.com/cshop/v3/util"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func adminCreateRandomFeaturedProductItem(t *testing.T) FeaturedProductItem {
	admin := createRandomAdmin(t)
	productItem := createRandomProductItem(t)
	arg := AdminCreateFeaturedProductItemParams{
		ProductItemID: productItem.ID,
		Active:        util.RandomBool(),
		Priority:      null.IntFrom(util.RandomMoney()),
		AdminID:       admin.ID,
		EndDate:       time.Now().Add(time.Hour * 24 * 7),
	}

	featuredProductItem, err := testStore.AdminCreateFeaturedProductItem(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, featuredProductItem)

	require.Equal(t, arg.ProductItemID, featuredProductItem.ProductItemID)
	require.Equal(t, arg.Active, featuredProductItem.Active)
	require.Equal(t, arg.Priority.Int64, featuredProductItem.Priority.Int64)
	require.NotEmpty(t, featuredProductItem.ID)

	return *featuredProductItem
}
func TestCreateFeaturedProductItem(t *testing.T) {
	adminCreateRandomFeaturedProductItem(t)
}

func TestGetFeaturedProductItem(t *testing.T) {
	featuredProductItem1 := adminCreateRandomFeaturedProductItem(t)

	featuredProductItem2, err := testStore.GetFeaturedProductItem(context.Background(), featuredProductItem1.ProductItemID)

	require.NoError(t, err)
	require.NotEmpty(t, featuredProductItem2)

	require.Equal(t, featuredProductItem1.ID, featuredProductItem2.ID)
	require.Equal(t, featuredProductItem1.ProductItemID, featuredProductItem2.ProductItemID)
	require.Equal(t, featuredProductItem1.Active, featuredProductItem2.Active)
	require.Equal(t, featuredProductItem1.StartDate, featuredProductItem2.StartDate)
	require.Equal(t, featuredProductItem1.EndDate, featuredProductItem2.EndDate)
	require.Equal(t, featuredProductItem1.Priority.Int64, featuredProductItem2.Priority.Int64)
}

func TestAdminUpdateFeaturedProductItemActive(t *testing.T) {
	admin := createRandomAdmin(t)
	featuredProductItem1 := adminCreateRandomFeaturedProductItem(t)
	arg := AdminUpdateFeaturedProductItemParams{
		AdminID:       admin.ID,
		Priority:      featuredProductItem1.Priority,
		Active:        null.BoolFrom(!featuredProductItem1.Active),
		StartDate:     null.TimeFrom(featuredProductItem1.StartDate),
		EndDate:       null.TimeFrom(featuredProductItem1.EndDate),
		ProductItemID: featuredProductItem1.ProductItemID,
	}

	featuredProductItem2, err := testStore.AdminUpdateFeaturedProductItem(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, featuredProductItem2)

	require.Equal(t, featuredProductItem1.ID, featuredProductItem2.ID)
	require.Equal(t, featuredProductItem1.ProductItemID, featuredProductItem2.ProductItemID)
	require.NotEqual(t, featuredProductItem1.Active, featuredProductItem2.Active)
	require.Equal(t, featuredProductItem1.StartDate, featuredProductItem2.StartDate)
	require.Equal(t, featuredProductItem1.EndDate, featuredProductItem2.EndDate)
	require.Equal(t, featuredProductItem1.Priority.Int64, featuredProductItem2.Priority.Int64)
}

func TestDeleteFeaturedProductItem(t *testing.T) {
	admin := createRandomAdmin(t)
	featuredProductItem1 := adminCreateRandomFeaturedProductItem(t)
	arg := DeleteFeaturedProductItemParams{
		ProductItemID: featuredProductItem1.ProductItemID,
		AdminID:       admin.ID,
	}
	err := testStore.DeleteFeaturedProductItem(context.Background(), arg)

	require.NoError(t, err)

	featuredProductItem2, err := testStore.GetFeaturedProductItem(context.Background(), featuredProductItem1.ProductItemID)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, featuredProductItem2)

}

func TestListFeaturedProductItems(t *testing.T) {
	for i := 0; i < 5; i++ {
		adminCreateRandomFeaturedProductItem(t)
	}
	arg := ListFeaturedProductItemsParams{
		Limit:  5,
		Offset: 0,
	}

	featuredProductItems, err := testStore.ListFeaturedProductItems(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, featuredProductItems)

	for _, featuredProductItem := range featuredProductItems {
		require.NotEmpty(t, featuredProductItem)
	}

}
