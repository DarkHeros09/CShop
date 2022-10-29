package db

import (
	"context"
	"testing"

	"github.com/cshop/v3/util"
	"github.com/guregu/null"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/require"
)

func createRandomShopOrderItem(t *testing.T) ShopOrderItem {
	shopOrder := createRandomShopOrder(t)
	productItem := createRandomProductItem(t)
	arg := CreateShopOrderItemParams{
		ProductItemID: productItem.ID,
		OrderID:       shopOrder.ID,
		Quantity:      int32(util.RandomInt(0, 100)),
		Price:         util.RandomDecimalString(1, 100),
	}

	shopOrderItem, err := testQueires.CreateShopOrderItem(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, shopOrderItem)

	require.Equal(t, arg.ProductItemID, shopOrderItem.ProductItemID)
	require.Equal(t, arg.OrderID, shopOrderItem.OrderID)
	require.Equal(t, arg.Quantity, shopOrderItem.Quantity)
	require.Equal(t, arg.Price, shopOrderItem.Price)

	return shopOrderItem
}
func TestCreateShopOrderItem(t *testing.T) {
	createRandomShopOrderItem(t)
}

func TestGetShopOrderItem(t *testing.T) {
	shopOrderItem1 := createRandomShopOrderItem(t)
	shopOrderItem2, err := testQueires.GetShopOrderItem(context.Background(), shopOrderItem1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, shopOrderItem2)

	require.Equal(t, shopOrderItem1.ID, shopOrderItem2.ID)
	require.Equal(t, shopOrderItem1.ProductItemID, shopOrderItem2.ProductItemID)
	require.Equal(t, shopOrderItem1.OrderID, shopOrderItem2.OrderID)
	require.Equal(t, shopOrderItem1.Quantity, shopOrderItem2.Quantity)
	require.Equal(t, shopOrderItem1.Price, shopOrderItem2.Price)
}

func TestUpdateShopOrderItemOrderTotal(t *testing.T) {
	shopOrderItem1 := createRandomShopOrderItem(t)
	arg := UpdateShopOrderItemParams{
		ProductItemID: null.Int{},
		OrderID:       null.Int{},
		Quantity:      null.Int{},
		Price:         null.StringFrom(util.RandomDecimalString(1, 100)),
		ID:            shopOrderItem1.ID,
	}

	shopOrderItem2, err := testQueires.UpdateShopOrderItem(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, shopOrderItem2)

	require.Equal(t, shopOrderItem1.ID, shopOrderItem2.ID)
	require.Equal(t, shopOrderItem1.ProductItemID, shopOrderItem2.ProductItemID)
	require.Equal(t, shopOrderItem1.OrderID, shopOrderItem2.OrderID)
	require.Equal(t, shopOrderItem1.Quantity, shopOrderItem2.Quantity)
	require.NotEqual(t, shopOrderItem1.Price, shopOrderItem2.Price)
}

func TestDeleteShopOrderItem(t *testing.T) {
	shopOrderItem1 := createRandomShopOrderItem(t)
	err := testQueires.DeleteShopOrderItem(context.Background(), shopOrderItem1.ID)

	require.NoError(t, err)

	shopOrderItem2, err := testQueires.GetShopOrderItem(context.Background(), shopOrderItem1.ID)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, shopOrderItem2)

}

func TestListShopOrderItems(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomShopOrderItem(t)
	}
	arg := ListShopOrderItemsParams{
		Limit:  5,
		Offset: 5,
	}

	shopOrderItems, err := testQueires.ListShopOrderItems(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, shopOrderItems)

	for _, shopOrderItem := range shopOrderItems {
		require.NotEmpty(t, shopOrderItem)
	}

}
