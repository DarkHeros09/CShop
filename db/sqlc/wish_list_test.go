package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
)

func createRandomWishList(t *testing.T) WishList {
	user1 := createRandomUser(t)

	wishList, err := testQueires.CreateWishList(context.Background(), user1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, wishList)

	require.Equal(t, user1.ID, wishList.UserID)

	return wishList
}

func TestCreateWishList(t *testing.T) {
	createRandomWishList(t)
}

func TestGetWishList(t *testing.T) {
	wishList1 := createRandomWishList(t)

	wishList2, err := testQueires.GetWishList(context.Background(), wishList1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, wishList2)

	require.Equal(t, wishList1.UserID, wishList2.UserID)
}

func TestUpdateWishList(t *testing.T) {
	wishList1 := createRandomWishList(t)
	user := createRandomUser(t)
	arg := UpdateWishListParams{
		UserID: sql.NullInt64{
			Int64: user.ID,
			Valid: false,
		},
		ID: wishList1.ID,
	}

	wishList2, err := testQueires.UpdateWishList(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, wishList2)

	require.Equal(t, wishList1.UserID, wishList2.UserID)
}

func TestDeleteWishList(t *testing.T) {
	wishList1 := createRandomWishList(t)

	err := testQueires.DeleteWishList(context.Background(), wishList1.ID)

	require.NoError(t, err)

	wishList2, err := testQueires.GetWishList(context.Background(), wishList1.ID)

	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, wishList2)

}

func TestListWishLists(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomWishList(t)
	}
	arg := ListWishListsParams{
		Limit:  5,
		Offset: 0,
	}

	wishLists, err := testQueires.ListWishLists(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, wishLists, 5)

	for _, wishList := range wishLists {
		require.NotEmpty(t, wishList)

	}
}
