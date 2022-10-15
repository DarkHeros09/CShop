package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/cshop/v3/util"
	"github.com/stretchr/testify/require"
)

func createRandomShoppingCartItem(t *testing.T) ShoppingCartItem {
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

	return shoppingCartItem
}

func TestCreateShoppingCartItem(t *testing.T) {
	createRandomShoppingCartItem(t)
}

func TestGetShoppingCartItem(t *testing.T) {
	shoppingCartItem1 := createRandomShoppingCartItem(t)

	shoppingCartItem2, err := testQueires.GetShoppingCartItem(context.Background(), shoppingCartItem1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, shoppingCartItem2)

	require.Equal(t, shoppingCartItem1.ShoppingCartID, shoppingCartItem2.ShoppingCartID)
	require.Equal(t, shoppingCartItem1.ProductItemID, shoppingCartItem2.ProductItemID)
	require.Equal(t, shoppingCartItem1.Qty, shoppingCartItem2.Qty)
}

func TestUpdateShoppingCartItem(t *testing.T) {
	shoppingCart := createRandomShoppingCart(t)
	shoppingCartItem1 := createRandomShoppingCartItem(t)
	arg := UpdateShoppingCartItemParams{
		ShoppingCartID: sql.NullInt64{
			Int64: shoppingCart.ID,
			Valid: true,
		},
		ProductItemID: sql.NullInt64{},
		ID:            shoppingCartItem1.ID,
	}

	shoppingCartItem2, err := testQueires.UpdateShoppingCartItem(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, shoppingCartItem2)

	require.NotEqual(t, shoppingCartItem1.ShoppingCartID, shoppingCartItem2.ShoppingCartID)
	require.Equal(t, shoppingCartItem1.ProductItemID, shoppingCartItem2.ProductItemID)
	require.Equal(t, shoppingCartItem1.Qty, shoppingCartItem2.Qty)
}

func TestDeleteShoppingCartItem(t *testing.T) {
	shoppingCartItem1 := createRandomShoppingCartItem(t)

	err := testQueires.DeleteShoppingCartItem(context.Background(), shoppingCartItem1.ID)

	require.NoError(t, err)

	shoppingCartItem2, err := testQueires.GetShoppingCartItem(context.Background(), shoppingCartItem1.ID)

	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, shoppingCartItem2)

}

func TestListShoppingCartItemes(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomShoppingCartItem(t)
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
