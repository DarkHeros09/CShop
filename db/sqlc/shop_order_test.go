package db

import (
	"context"
	"sync"
	"testing"

	"github.com/cshop/v3/util"
	"github.com/guregu/null"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/require"
)

func createRandomShopOrder(t *testing.T) ShopOrder {
	paymentMethod := createRandomPaymentMethod(t)
	address := createRandomAddress(t)
	shippingMethod := createRandomShippingMethod(t)
	orderStatus := createRandomOrderStatus(t)
	arg := CreateShopOrderParams{
		UserID:            paymentMethod.UserID,
		PaymentMethodID:   paymentMethod.ID,
		ShippingAddressID: address.ID,
		OrderTotal:        util.RandomDecimalString(1, 100),
		ShippingMethodID:  shippingMethod.ID,
		OrderStatusID:     null.IntFromPtr(&orderStatus.ID),
	}

	shopOrder, err := testQueires.CreateShopOrder(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, shopOrder)

	require.Equal(t, arg.UserID, shopOrder.UserID)
	require.Equal(t, arg.PaymentMethodID, shopOrder.PaymentMethodID)
	require.Equal(t, arg.ShippingAddressID, shopOrder.ShippingAddressID)
	require.Equal(t, arg.OrderTotal, shopOrder.OrderTotal)
	require.Equal(t, arg.ShippingMethodID, shopOrder.ShippingMethodID)
	require.Equal(t, arg.OrderStatusID, shopOrder.OrderStatusID)

	return shopOrder
}
func TestCreateShopOrder(t *testing.T) {
	createRandomShopOrder(t)
}

func TestGetShopOrder(t *testing.T) {
	shopOrder1 := createRandomShopOrder(t)
	shopOrder2, err := testQueires.GetShopOrder(context.Background(), shopOrder1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, shopOrder2)

	require.Equal(t, shopOrder1.ID, shopOrder2.ID)
	require.Equal(t, shopOrder1.UserID, shopOrder2.UserID)
	require.Equal(t, shopOrder1.PaymentMethodID, shopOrder2.PaymentMethodID)
	require.Equal(t, shopOrder1.ShippingAddressID, shopOrder2.ShippingAddressID)
	require.Equal(t, shopOrder1.OrderTotal, shopOrder2.OrderTotal)
	require.Equal(t, shopOrder1.ShippingMethodID, shopOrder2.ShippingMethodID)
	require.Equal(t, shopOrder1.OrderStatusID, shopOrder2.OrderStatusID)
}

func TestUpdateShopOrderOrderTotal(t *testing.T) {
	shopOrder1 := createRandomShopOrder(t)
	arg := UpdateShopOrderParams{
		UserID:            null.Int{},
		PaymentMethodID:   null.Int{},
		ShippingAddressID: null.Int{},
		OrderTotal:        null.StringFrom(util.RandomDecimalString(1, 100)),
		ShippingMethodID:  null.Int{},
		OrderStatusID:     null.Int{},
		ID:                shopOrder1.ID,
	}

	shopOrder2, err := testQueires.UpdateShopOrder(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, shopOrder2)

	require.Equal(t, shopOrder1.ID, shopOrder2.ID)
	require.Equal(t, shopOrder1.UserID, shopOrder2.UserID)
	require.Equal(t, shopOrder1.PaymentMethodID, shopOrder2.PaymentMethodID)
	require.Equal(t, shopOrder1.ShippingAddressID, shopOrder2.ShippingAddressID)
	require.NotEqual(t, shopOrder1.OrderTotal, shopOrder2.OrderTotal)
	require.Equal(t, shopOrder1.ShippingMethodID, shopOrder2.ShippingMethodID)
	require.Equal(t, shopOrder1.OrderStatusID, shopOrder2.OrderStatusID)
}

func TestDeleteShopOrder(t *testing.T) {
	shopOrder1 := createRandomShopOrder(t)
	err := testQueires.DeleteShopOrder(context.Background(), shopOrder1.ID)

	require.NoError(t, err)

	shopOrder2, err := testQueires.GetShopOrder(context.Background(), shopOrder1.ID)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, shopOrder2)

}

func TestListShopOrders(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func(i int) {
			createRandomShopOrder(t)
			wg.Done()
		}(i)
	}
	wg.Wait()
	arg := ListShopOrdersParams{
		Limit:  5,
		Offset: 0,
	}

	shopOrders, err := testQueires.ListShopOrders(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, shopOrders)

	for _, shopOrder := range shopOrders {
		require.NotEmpty(t, shopOrder)
	}

}
