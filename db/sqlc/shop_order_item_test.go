package db

import (
	"context"
	"testing"

	"github.com/cshop/v3/util"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func createRandomShopOrderItem(t *testing.T) (ShopOrderItem, ShopOrder) {
	shopOrder := createRandomShopOrder(t)
	productItem := createRandomProductItem(t)
	arg := CreateShopOrderItemParams{
		ProductItemID:       productItem.ID,
		OrderID:             shopOrder.ID,
		Quantity:            int32(util.RandomInt(0, 100)),
		Price:               util.RandomDecimalString(1, 200),
		Discount:            int32(util.RandomInt(0, 90)),
		ShippingMethodPrice: util.RandomDecimalString(1, 100),
	}

	shopOrderItem, err := testStore.CreateShopOrderItem(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, shopOrderItem)

	require.Equal(t, arg.ProductItemID, shopOrderItem.ProductItemID)
	require.Equal(t, arg.OrderID, shopOrderItem.OrderID)
	require.Equal(t, arg.Quantity, shopOrderItem.Quantity)
	require.Equal(t, arg.Price, shopOrderItem.Price)
	require.Equal(t, arg.Discount, shopOrderItem.Discount)
	require.Equal(t, arg.ShippingMethodPrice, shopOrderItem.ShippingMethodPrice)

	return *shopOrderItem, shopOrder
}
func TestCreateShopOrderItem(t *testing.T) {
	createRandomShopOrderItem(t)
}

func TestGetShopOrderItem(t *testing.T) {
	t.Parallel()
	shopOrderItem1, _ := createRandomShopOrderItem(t)
	shopOrderItem2, err := testStore.GetShopOrderItem(context.Background(), shopOrderItem1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, shopOrderItem2)

	require.Equal(t, shopOrderItem1.ID, shopOrderItem2.ID)
	require.Equal(t, shopOrderItem1.ProductItemID, shopOrderItem2.ProductItemID)
	require.Equal(t, shopOrderItem1.OrderID, shopOrderItem2.OrderID)
	require.Equal(t, shopOrderItem1.Quantity, shopOrderItem2.Quantity)
	require.Equal(t, shopOrderItem1.Price, shopOrderItem2.Price)
	require.Equal(t, shopOrderItem1.Discount, shopOrderItem2.Discount)
}

func TestUpdateShopOrderItemOrderTotal(t *testing.T) {
	shopOrderItem1, _ := createRandomShopOrderItem(t)
	arg := UpdateShopOrderItemParams{
		ProductItemID: shopOrderItem1.ProductItemID,
		OrderID:       shopOrderItem1.OrderID,
		Quantity:      null.Int{},
		Price:         null.StringFrom(util.RandomDecimalString(1, 100)),
		ID:            shopOrderItem1.ID,
	}

	shopOrderItem2, err := testStore.UpdateShopOrderItem(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, shopOrderItem2)

	require.Equal(t, shopOrderItem1.ID, shopOrderItem2.ID)
	require.Equal(t, shopOrderItem1.ProductItemID, shopOrderItem2.ProductItemID)
	require.Equal(t, shopOrderItem1.OrderID, shopOrderItem2.OrderID)
	require.Equal(t, shopOrderItem1.Quantity, shopOrderItem2.Quantity)
	require.NotEqual(t, shopOrderItem1.Price, shopOrderItem2.Price)
	require.Equal(t, shopOrderItem1.Discount, shopOrderItem2.Discount)
}

func TestDeleteShopOrderItem(t *testing.T) {
	admin := createRandomAdmin(t)
	shopOrderItem1, _ := createRandomShopOrderItem(t)

	arg := DeleteShopOrderItemParams{
		AdminID: admin.ID,
		ID:      shopOrderItem1.ID,
	}

	deletedShopOrderItem, err := testStore.DeleteShopOrderItem(context.Background(), arg)

	require.NotEmpty(t, deletedShopOrderItem)
	require.NoError(t, err)

	shopOrderItem2, err := testStore.GetShopOrderItem(context.Background(), shopOrderItem1.ID)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, shopOrderItem2)

}

func TestListShopOrderItemsByUserIDOrderID(t *testing.T) {
	// var wg sync.WaitGroup
	// wg.Add(10)
	var shopOrderItem ShopOrderItem
	var shopOrder ShopOrder
	for i := 0; i < 5; i++ {
		// go func() {
		shopOrderItem, shopOrder = createRandomShopOrderItem(t)
		// wg.Done()
		// }()
	}
	// wg.Wait()
	arg := ListShopOrderItemsByUserIDOrderIDParams{
		OrderID: shopOrderItem.OrderID,
		UserID:  shopOrder.UserID,
	}

	shopOrderItems, err := testStore.ListShopOrderItemsByUserIDOrderID(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, shopOrderItems)

	for _, shopOrderItem := range shopOrderItems {
		require.NotEmpty(t, shopOrderItem)
	}

}

func TestListShopOrderItems(t *testing.T) {
	t.Parallel()
	// var wg sync.WaitGroup
	// wg.Add(10)
	for i := 0; i < 5; i++ {
		// go func() {
		createRandomShopOrderItem(t)
		// 	wg.Done()
		// }()
	}
	// wg.Wait()
	arg := ListShopOrderItemsParams{
		Limit:  5,
		Offset: 0,
	}

	shopOrderItems, err := testStore.ListShopOrderItems(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, shopOrderItems)

	for _, shopOrderItem := range shopOrderItems {
		require.NotEmpty(t, shopOrderItem)
	}

}
