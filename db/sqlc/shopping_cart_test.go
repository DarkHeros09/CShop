package db

import (
	"context"
	"testing"

	"github.com/guregu/null/v5"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func createRandomShoppingCart(t *testing.T) ShoppingCart {

	// shoppingCartChan := make(chan ShoppingCart)

	// go func() {
	user1 := createRandomUser(t)
	shoppingCart, err := testStore.CreateShoppingCart(context.Background(), user1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, shoppingCart)

	require.Equal(t, user1.ID, shoppingCart.UserID)

	// shoppingCartChan <- shoppingCart
	// }()

	// shoppingCart := <-shoppingCartChan

	return shoppingCart
}

func TestCreateShoppingCart(t *testing.T) {
	createRandomShoppingCart(t)
}

func TestGetShoppingCart(t *testing.T) {
	t.Parallel()
	shoppingCart1 := createRandomShoppingCart(t)

	shoppingCart2, err := testStore.GetShoppingCart(context.Background(), shoppingCart1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, shoppingCart2)

	require.Equal(t, shoppingCart1.UserID, shoppingCart2.UserID)
}

func TestGetShoppingCartByUser(t *testing.T) {
	shoppingCart1 := createRandomShoppingCart(t)

	arg := GetShoppingCartByUserIDCartIDParams{
		UserID: shoppingCart1.UserID,
		ID:     shoppingCart1.ID,
	}

	shoppingCart2, err := testStore.GetShoppingCartByUserIDCartID(context.Background(), arg)
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

	shoppingCart2, err := testStore.UpdateShoppingCart(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, shoppingCart2)

	require.Equal(t, shoppingCart1.UserID, shoppingCart2.UserID)
}

func TestDeleteShoppingCart(t *testing.T) {
	shoppingCart1 := createRandomShoppingCart(t)

	err := testStore.DeleteShoppingCart(context.Background(), shoppingCart1.ID)

	require.NoError(t, err)

	shoppingCart2, err := testStore.GetShoppingCart(context.Background(), shoppingCart1.ID)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, shoppingCart2)

}

func TestListShoppingCarts(t *testing.T) {
	// var wg sync.WaitGroup
	// wg.Add(10)
	for i := 0; i < 5; i++ {
		// go func() {
		createRandomShoppingCart(t)
		// 	wg.Done()
		// }()
	}

	// wg.Wait()
	arg := ListShoppingCartsParams{
		Limit:  5,
		Offset: 0,
	}

	shoppingCarts, err := testStore.ListShoppingCarts(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, shoppingCarts, 5)

	for _, shoppingCart := range shoppingCarts {
		require.NotEmpty(t, shoppingCart)

	}
}
