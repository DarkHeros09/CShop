package db

import (
	"context"
	"sync"
	"testing"

	"github.com/guregu/null/v5"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func createRandomWishList(t *testing.T) WishList {
	t.Helper()
	user1 := createRandomUser(t)

	wishList, err := testStore.CreateWishList(context.Background(), user1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, wishList)

	require.Equal(t, user1.ID, wishList.UserID)

	return *wishList
}

func TestCreateWishList(t *testing.T) {
	t.Parallel()
	createRandomWishList(t)
}

func TestGetWishList(t *testing.T) {
	wishList1 := createRandomWishList(t)

	wishList2, err := testStore.GetWishList(context.Background(), wishList1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, wishList2)

	require.Equal(t, wishList1.UserID, wishList2.UserID)
}

func TestUpdateWishList(t *testing.T) {
	wishList1 := createRandomWishList(t)
	// user := createRandomUser(t)
	arg := UpdateWishListParams{
		UserID: null.IntFromPtr(&wishList1.UserID),
		ID:     wishList1.ID,
	}

	wishList2, err := testStore.UpdateWishList(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, wishList2)

	require.Equal(t, wishList1.UserID, wishList2.UserID)
}

func TestDeleteWishList(t *testing.T) {
	t.Parallel()
	wishList1 := createRandomWishList(t)

	err := testStore.DeleteWishList(context.Background(), wishList1.ID)

	require.NoError(t, err)

	wishList2, err := testStore.GetWishList(context.Background(), wishList1.ID)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, wishList2)

}

func TestListWishLists(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			createRandomWishList(t)
			wg.Done()
		}()
	}
	wg.Wait()
	arg := ListWishListsParams{
		Limit:  5,
		Offset: 0,
	}

	wishLists, err := testStore.ListWishLists(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, wishLists, 5)

	for _, wishList := range wishLists {
		require.NotEmpty(t, wishList)

	}
}
