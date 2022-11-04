package db

import (
	"context"
	"testing"

	"github.com/cshop/v3/util"
	"github.com/guregu/null"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/require"
)

func createRandomShoppingCartItem(t *testing.T) (ShoppingCartItem, ShoppingCart) {
	shoppingCart := createRandomShoppingCart(t)
	productItem := createRandomProductItem(t)

	arg := CreateShoppingCartItemParams{
		ShoppingCartID: shoppingCart.ID,
		ProductItemID:  productItem.ID,
		Qty:            int32(util.RandomInt(0, 10)),
	}
	shoppingCartItem, err := testQueires.CreateShoppingCartItem(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, shoppingCartItem)

	require.Equal(t, arg.ShoppingCartID, shoppingCartItem.ShoppingCartID)
	require.Equal(t, arg.ProductItemID, shoppingCartItem.ProductItemID)
	require.Equal(t, arg.Qty, shoppingCartItem.Qty)

	return shoppingCartItem, shoppingCart
}

func TestCreateShoppingCartItem(t *testing.T) {
	createRandomShoppingCartItem(t)
}

func TestGetShoppingCartItem(t *testing.T) {
	shoppingCartItem1, _ := createRandomShoppingCartItem(t)

	shoppingCartItem2, err := testQueires.GetShoppingCartItem(context.Background(), shoppingCartItem1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, shoppingCartItem2)

	require.Equal(t, shoppingCartItem1.ShoppingCartID, shoppingCartItem2.ShoppingCartID)
	require.Equal(t, shoppingCartItem1.ProductItemID, shoppingCartItem2.ProductItemID)
	require.Equal(t, shoppingCartItem1.Qty, shoppingCartItem2.Qty)
}

func TestGetShoppingCartItemByCartID(t *testing.T) {
	shoppingCartItem1, shoppingCart1 := createRandomShoppingCartItem(t)

	arg := GetShoppingCartItemByUserIDCartIDParams{
		UserID:         shoppingCart1.UserID,
		ShoppingCartID: shoppingCartItem1.ShoppingCartID,
	}

	shoppingCartItem2, err := testQueires.GetShoppingCartItemByUserIDCartID(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, shoppingCartItem2)

	require.Equal(t, shoppingCartItem1.ShoppingCartID, shoppingCartItem2.ShoppingCartID)
	require.Equal(t, shoppingCartItem1.ProductItemID, shoppingCartItem2.ProductItemID)
	require.Equal(t, shoppingCartItem1.Qty, shoppingCartItem2.Qty)
}

func TestUpdateShoppingCartItem(t *testing.T) {
	shoppingCartItem1, _ := createRandomShoppingCartItem(t)
	arg := UpdateShoppingCartItemParams{
		ProductItemID:  null.Int{},
		Qty:            null.Int{},
		ID:             shoppingCartItem1.ID,
		ShoppingCartID: shoppingCartItem1.ShoppingCartID,
	}

	shoppingCartItem2, err := testQueires.UpdateShoppingCartItem(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, shoppingCartItem2)

	require.Equal(t, shoppingCartItem1.ShoppingCartID, shoppingCartItem2.ShoppingCartID)
	require.Equal(t, shoppingCartItem1.ProductItemID, shoppingCartItem2.ProductItemID)
	require.Equal(t, shoppingCartItem1.Qty, shoppingCartItem2.Qty)
}

func TestDeleteShoppingCartItem(t *testing.T) {
	shoppingCartItem1, shoppingCart := createRandomShoppingCartItem(t)

	arg := DeleteShoppingCartItemParams{
		ID:     shoppingCartItem1.ID,
		UserID: shoppingCart.UserID,
	}

	err := testQueires.DeleteShoppingCartItem(context.Background(), arg)

	require.NoError(t, err)

	shoppingCartItem2, err := testQueires.GetShoppingCartItem(context.Background(), shoppingCartItem1.ID)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, shoppingCartItem2)

}

func TestListShoppingCartItemes(t *testing.T) {
	shoppingCart := createRandomShoppingCart(t)
	productItem := createRandomProductItem(t)
	for i := 0; i < 10; i++ {
		arg := CreateShoppingCartItemParams{
			ShoppingCartID: shoppingCart.ID,
			ProductItemID:  productItem.ID,
			Qty:            int32(util.RandomInt(0, 10)),
		}
		testQueires.CreateShoppingCartItem(context.Background(), arg)
	}

	arg := ListShoppingCartItemsParams{
		Limit:  5,
		Offset: 0,
	}

	shoppingCartItems, err := testQueires.ListShoppingCartItems(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, shoppingCartItems, 5)

	for _, shoppingCartItem := range shoppingCartItems {
		require.NotEmpty(t, shoppingCartItem)

	}
}

func TestListShoppingCartItemsByCartID(t *testing.T) {
	shoppingCart := createRandomShoppingCart(t)
	productItem := createRandomProductItem(t)
	for i := 0; i < 10; i++ {
		arg := CreateShoppingCartItemParams{
			ShoppingCartID: shoppingCart.ID,
			ProductItemID:  productItem.ID,
			Qty:            int32(util.RandomInt(0, 10)),
		}
		testQueires.CreateShoppingCartItem(context.Background(), arg)
	}

	shoppingCartItems, err := testQueires.ListShoppingCartItemsByCartID(context.Background(), shoppingCart.ID)
	require.NoError(t, err)
	require.Len(t, shoppingCartItems, 10)

	for _, shoppingCartItem := range shoppingCartItems {
		require.NotEmpty(t, shoppingCartItem)

	}
}

func TestListShoppingCartItemsByUserID(t *testing.T) {
	shoppingCart := createRandomShoppingCart(t)
	productItem := createRandomProductItem(t)
	for i := 0; i < 10; i++ {
		arg := CreateShoppingCartItemParams{
			ShoppingCartID: shoppingCart.ID,
			ProductItemID:  productItem.ID,
			Qty:            int32(util.RandomInt(0, 10)),
		}
		testQueires.CreateShoppingCartItem(context.Background(), arg)
	}

	shoppingCartItems, err := testQueires.ListShoppingCartItemsByUserID(context.Background(), shoppingCart.UserID)
	require.NoError(t, err)
	require.Len(t, shoppingCartItems, 10)

	for _, shoppingCartItem := range shoppingCartItems {
		require.NotEmpty(t, shoppingCartItem)

	}
}

func TestDeleteALLShoppingCartItemes(t *testing.T) {
	shoppingCart := createRandomShoppingCart(t)
	productItem := createRandomProductItem(t)
	for i := 0; i < 10; i++ {
		arg := CreateShoppingCartItemParams{
			ShoppingCartID: shoppingCart.ID,
			ProductItemID:  productItem.ID,
			Qty:            int32(util.RandomInt(0, 10)),
		}
		testQueires.CreateShoppingCartItem(context.Background(), arg)
	}

	shoppingCartItem1, err := testQueires.DeleteShoppingCartItemAllByUser(context.Background(), shoppingCart.UserID)
	require.NoError(t, err)
	require.NotEmpty(t, shoppingCartItem1)

	shoppingCartItem2, err := testQueires.DeleteShoppingCartItemAllByUser(context.Background(), shoppingCart.UserID)
	// require.Error(t, err)
	require.Empty(t, shoppingCartItem2)
	// require.EqualError(t, err, pgx.ErrNoRows.Error())
}
