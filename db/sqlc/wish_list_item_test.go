package db

import (
	"context"
	"sync"
	"testing"

	"github.com/guregu/null"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func createRandomWishListItem(t *testing.T) WishListItem {
	t.Helper()
	wishList := createRandomWishList(t)
	productItem := createRandomProductItem(t)

	arg := CreateWishListItemParams{
		WishListID:    wishList.ID,
		ProductItemID: productItem.ID,
	}
	wishListItem, err := testStore.CreateWishListItem(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, wishListItem)

	require.Equal(t, arg.WishListID, wishListItem.WishListID)
	require.Equal(t, arg.ProductItemID, wishListItem.ProductItemID)

	return wishListItem
}

func TestCreateWishListItem(t *testing.T) {
	t.Parallel()
	createRandomWishListItem(t)
}

func TestGetWishListItem(t *testing.T) {
	t.Parallel()
	wishListItem1 := createRandomWishListItem(t)

	wishListItem2, err := testStore.GetWishListItem(context.Background(), wishListItem1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, wishListItem2)

	require.Equal(t, wishListItem1.WishListID, wishListItem2.WishListID)
	require.Equal(t, wishListItem1.ProductItemID, wishListItem2.ProductItemID)
}

func TestGetWishListItemByUserIDWishID(t *testing.T) {
	t.Parallel()
	wishListItem1 := createRandomWishListItem(t)

	wishList, err := testStore.GetWishList(context.Background(), wishListItem1.WishListID)
	require.NoError(t, err)

	arg := GetWishListItemByUserIDCartIDParams{
		UserID:     wishList.UserID,
		ID:         wishListItem1.ID,
		WishListID: wishListItem1.WishListID,
	}
	wishListItem2, err := testStore.GetWishListItemByUserIDCartID(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, wishListItem2)

	require.Equal(t, wishListItem1.WishListID, wishListItem2.WishListID)
	require.Equal(t, wishListItem1.ProductItemID, wishListItem2.ProductItemID)
}

func TestUpdateWishListItem(t *testing.T) {
	t.Parallel()
	wishListItem := createRandomWishListItem(t)
	newProduct := createRandomProductItem(t)
	arg := UpdateWishListItemParams{
		ProductItemID: null.IntFrom(newProduct.ID),
		ID:            wishListItem.ID,
		WishListID:    wishListItem.WishListID,
	}

	wishListItem2, err := testStore.UpdateWishListItem(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, wishListItem2)

	require.Equal(t, wishListItem.WishListID, wishListItem2.WishListID)
	require.NotEqual(t, wishListItem.ProductItemID, wishListItem2.ProductItemID)
}

func TestDeleteWishListItem(t *testing.T) {
	t.Parallel()
	wishListItem1 := createRandomWishListItem(t)

	wishList, err := testStore.GetWishList(context.Background(), wishListItem1.WishListID)
	require.NoError(t, err)

	arg := DeleteWishListItemParams{
		ID:         wishListItem1.ID,
		WishListID: wishList.ID,
	}

	err = testStore.DeleteWishListItem(context.Background(), arg)

	require.NoError(t, err)

	wishListItem2, err := testStore.GetWishListItem(context.Background(), wishListItem1.ID)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, wishListItem2)

}

func TestDeleteWishListItemAll(t *testing.T) {
	t.Parallel()
	wishListItem1 := createRandomWishListItem(t)

	_, err := testStore.DeleteWishListItemAll(context.Background(), wishListItem1.WishListID)

	require.NoError(t, err)

	wishListItem2, err := testStore.GetWishListItem(context.Background(), wishListItem1.ID)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, wishListItem2)

}

func TestListWishListItemes(t *testing.T) {
	t.Parallel()
	var wg sync.WaitGroup
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go func() {
			createRandomWishListItem(t)
			wg.Done()
		}()
	}
	wg.Wait()
	arg := ListWishListItemsParams{
		Limit:  5,
		Offset: 0,
	}

	wishListItems, err := testStore.ListWishListItems(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, wishListItems, 5)

	for _, wishListItem := range wishListItems {
		require.NotEmpty(t, wishListItem)

	}
}
