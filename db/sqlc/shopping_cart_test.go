package db

import (
	"context"
	"testing"

	"github.com/guregu/null"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/require"
)

func createRandomShoppingCart(t *testing.T) ShoppingCart {
	user1 := createRandomUser(t)

	shoppingCart, err := testQueires.CreateShoppingCart(context.Background(), user1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, shoppingCart)

	require.Equal(t, user1.ID, shoppingCart.UserID)

	return shoppingCart
}

func TestCreateShoppingCart(t *testing.T) {
	createRandomShoppingCart(t)
}

func TestGetShoppingCart(t *testing.T) {
	shoppingCart1 := createRandomShoppingCart(t)

	shoppingCart2, err := testQueires.GetShoppingCart(context.Background(), shoppingCart1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, shoppingCart2)

	require.Equal(t, shoppingCart1.UserID, shoppingCart2.UserID)
}

func TestGetShoppingCartByUser(t *testing.T) {
	shoppingCart1 := createRandomShoppingCart(t)

	shoppingCart2, err := testQueires.GetShoppingCartByUserID(context.Background(), shoppingCart1.UserID)
	require.NoError(t, err)
	require.NotEmpty(t, shoppingCart2)

	require.Equal(t, shoppingCart1.UserID, shoppingCart2.UserID)
}

func TestUpdateShoppingCart(t *testing.T) {
	shoppingCart1 := createRandomShoppingCart(t)
	// user := createRandomUser(t)
	arg := UpdateShoppingCartParams{
		UserID: null.IntFromPtr(&shoppingCart1.UserID),
		ID:     shoppingCart1.ID,
	}

	shoppingCart2, err := testQueires.UpdateShoppingCart(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, shoppingCart2)

	require.Equal(t, shoppingCart1.UserID, shoppingCart2.UserID)
}

func TestDeleteShoppingCart(t *testing.T) {
	shoppingCart1 := createRandomShoppingCart(t)

	err := testQueires.DeleteShoppingCart(context.Background(), shoppingCart1.ID)

	require.NoError(t, err)

	shoppingCart2, err := testQueires.GetShoppingCart(context.Background(), shoppingCart1.ID)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, shoppingCart2)

}

func TestListShoppingCarts(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomShoppingCart(t)
	}
	arg := ListShoppingCartsParams{
		Limit:  5,
		Offset: 0,
	}

	shoppingCarts, err := testQueires.ListShoppingCarts(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, shoppingCarts, 5)

	for _, shoppingCart := range shoppingCarts {
		require.NotEmpty(t, shoppingCart)

	}
}
