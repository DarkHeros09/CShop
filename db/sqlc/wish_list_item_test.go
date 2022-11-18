package db

import (
	"context"
	"testing"

	"github.com/guregu/null"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/require"
)

func createRandomWishListItem(t *testing.T) WishListItem {
	wishList := createRandomWishList(t)
	productItem := createRandomProductItem(t)

	arg := CreateWishListItemParams{
		WishListID:    wishList.ID,
		ProductItemID: productItem.ID,
	}
	wishListItem, err := testQueires.CreateWishListItem(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, wishListItem)

	require.Equal(t, arg.WishListID, wishListItem.WishListID)
	require.Equal(t, arg.ProductItemID, wishListItem.ProductItemID)

	return wishListItem
}

func TestCreateWishListItem(t *testing.T) {
	createRandomWishListItem(t)
}

func TestGetWishListItem(t *testing.T) {
	wishListItem1 := createRandomWishListItem(t)

	wishListItem2, err := testQueires.GetWishListItem(context.Background(), wishListItem1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, wishListItem2)

	require.Equal(t, wishListItem1.WishListID, wishListItem2.WishListID)
	require.Equal(t, wishListItem1.ProductItemID, wishListItem2.ProductItemID)
}

func TestUpdateWishListItem(t *testing.T) {
	wishList := createRandomWishList(t)
	wishListItem1 := createRandomWishListItem(t)
	newProduct := createRandomProductItem(t)
	arg := UpdateWishListItemParams{
		WishListID:    wishList.ID,
		ProductItemID: null.IntFrom(newProduct.ID),
		ID:            wishListItem1.ID,
	}

	wishListItem2, err := testQueires.UpdateWishListItem(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, wishListItem2)

	require.Equal(t, wishListItem1.WishListID, wishListItem2.WishListID)
	require.NotEqual(t, wishListItem1.ProductItemID, wishListItem2.ProductItemID)
}

func TestDeleteWishListItem(t *testing.T) {
	wishListItem1 := createRandomWishListItem(t)

	wishList, err := testQueires.GetWishList(context.Background(), wishListItem1.WishListID)
	require.NoError(t, err)

	arg := DeleteWishListItemParams{
		ID:     wishListItem1.ID,
		UserID: wishList.UserID,
	}

	err = testQueires.DeleteWishListItem(context.Background(), arg)

	require.NoError(t, err)

	wishListItem2, err := testQueires.GetWishListItem(context.Background(), wishListItem1.ID)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, wishListItem2)

}

func TestListWishListItemes(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomWishListItem(t)
	}

	arg := ListWishListItemsParams{
		Limit:  5,
		Offset: 0,
	}

	wishListItems, err := testQueires.ListWishListItems(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, wishListItems, 5)

	for _, wishListItem := range wishListItems {
		require.NotEmpty(t, wishListItem)

	}
}
