package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/cshop/v3/util"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func createRandomShopOrder(t *testing.T) ShopOrder {
	paymentMethod := createRandomPaymentMethod(t)
	user, err := testStore.GetUser(context.Background(), paymentMethod.UserID)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	address := createRandomAddress(user, t)
	shippingMethod := createRandomShippingMethod(t)
	orderStatus := createRandomOrderStatus(t)
	arg := CreateShopOrderParams{
		TrackNumber: util.GenerateTrackNumber(),
		UserID:      paymentMethod.UserID,
		// PaymentMethodID:   paymentMethod.ID,
		ShippingAddressID: null.IntFrom(address.ID),
		OrderTotal:        util.RandomDecimalString(1, 100),
		ShippingMethodID:  shippingMethod.ID,
		OrderStatusID:     null.IntFromPtr(&orderStatus.ID),
		AddressName:       util.RandomUser(),
		AddressTelephone:  util.GenerateRandomValidPhoneNumber(),
		AddressLine:       util.RandomUser(),
		AddressRegion:     util.RandomUser(),
		AddressCity:       util.RandomUser(),
	}

	shopOrder, err := testStore.CreateShopOrder(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, shopOrder)

	require.Equal(t, arg.UserID, shopOrder.UserID)
	// // require.Equal(t, arg.PaymentMethodID, shopOrder.PaymentMethodID)
	require.Equal(t, arg.ShippingAddressID, shopOrder.ShippingAddressID)
	require.Equal(t, arg.OrderTotal, shopOrder.OrderTotal)
	require.Equal(t, arg.ShippingMethodID, shopOrder.ShippingMethodID)
	require.Equal(t, arg.OrderStatusID, shopOrder.OrderStatusID)

	return shopOrder
}
func createRandomShopOrderForListV2(t *testing.T) ShopOrder {
	var shopOrder ShopOrder
	var err error
	paymentMethod := createRandomPaymentMethod(t)

	user, err := testStore.GetUser(context.Background(), paymentMethod.UserID)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	address := createRandomAddress(user, t)
	shippingMethod := createRandomShippingMethod(t)
	orderStatus := createRandomOrderStatus(t)
	arg := CreateShopOrderParams{
		TrackNumber: util.GenerateTrackNumber(),
		UserID:      paymentMethod.UserID, PaymentTypeID: util.RandomMoney(),
		// PaymentMethodID:   paymentMethod.ID,
		ShippingAddressID: null.IntFrom(address.ID),
		OrderTotal:        util.RandomDecimalString(1, 100),
		ShippingMethodID:  shippingMethod.ID,
		OrderStatusID:     null.IntFromPtr(&orderStatus.ID),
	}

	for i := 0; i < 30; i++ {
		shopOrder, err = testStore.CreateShopOrder(context.Background(), arg)
		require.NoError(t, err)
		require.NotEmpty(t, shopOrder)

		require.Equal(t, arg.UserID, shopOrder.UserID)
		// // require.Equal(t, arg.PaymentMethodID, shopOrder.PaymentMethodID)
		require.Equal(t, arg.ShippingAddressID, shopOrder.ShippingAddressID)
		require.Equal(t, arg.OrderTotal, shopOrder.OrderTotal)
		require.Equal(t, arg.ShippingMethodID, shopOrder.ShippingMethodID)
		require.Equal(t, arg.OrderStatusID, shopOrder.OrderStatusID)
	}

	return shopOrder
}
func TestCreateShopOrder(t *testing.T) {
	createRandomShopOrder(t)
}
func TestCreateShopOrderForListV2(t *testing.T) {
	createRandomShopOrderForListV2(t)
}

func TestGetShopOrder(t *testing.T) {
	shopOrder1 := createRandomShopOrder(t)
	shopOrder2, err := testStore.GetShopOrder(context.Background(), shopOrder1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, shopOrder2)

	require.Equal(t, shopOrder1.ID, shopOrder2.ID)
	require.Equal(t, shopOrder1.UserID, shopOrder2.UserID)
	// // require.Equal(t, shopOrder1.PaymentMethodID, shopOrder2.PaymentMethodID)
	require.Equal(t, shopOrder1.ShippingAddressID, shopOrder2.ShippingAddressID)
	require.Equal(t, shopOrder1.OrderTotal, shopOrder2.OrderTotal)
	require.Equal(t, shopOrder1.ShippingMethodID, shopOrder2.ShippingMethodID)
	require.Equal(t, shopOrder1.OrderStatusID, shopOrder2.OrderStatusID)
}

func TestUpdateShopOrderOrderTotal(t *testing.T) {
	admin := createRandomAdmin(t)
	shopOrder1 := createRandomShopOrder(t)
	arg := UpdateShopOrderParams{
		AdminID:    admin.ID,
		OrderTotal: null.StringFrom(util.RandomDecimalString(1, 100)),
		ID:         shopOrder1.ID,
	}

	shopOrder2, err := testStore.UpdateShopOrder(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, shopOrder2)

	require.Equal(t, shopOrder1.ID, shopOrder2.ID)
	require.Equal(t, shopOrder1.UserID, shopOrder2.UserID)
	// // require.Equal(t, shopOrder1.PaymentMethodID, shopOrder2.PaymentMethodID)
	require.Equal(t, shopOrder1.ShippingAddressID, shopOrder2.ShippingAddressID)
	require.NotEqual(t, shopOrder1.OrderTotal, shopOrder2.OrderTotal)
	require.Equal(t, shopOrder1.ShippingMethodID, shopOrder2.ShippingMethodID)
	require.Equal(t, shopOrder1.OrderStatusID, shopOrder2.OrderStatusID)
}
func TestUpdateShopOrderOrderStatus(t *testing.T) {
	admin := createRandomAdmin(t)
	for {
		shopOrder1 := createRandomShopOrder(t)
		if shopOrder1.OrderStatusID.Int64 != 2 {
			arg := UpdateShopOrderParams{
				AdminID:       admin.ID,
				OrderStatusID: null.IntFrom(2),
				ID:            shopOrder1.ID,
			}

			shopOrder2, err := testStore.UpdateShopOrder(context.Background(), arg)
			require.NoError(t, err)
			require.NotEmpty(t, shopOrder2)

			require.Equal(t, shopOrder1.ID, shopOrder2.ID)
			require.Equal(t, shopOrder1.UserID, shopOrder2.UserID)
			// require.Equal(t, shopOrder1.PaymentMethodID, shopOrder2.PaymentMethodID)
			require.Equal(t, shopOrder1.ShippingAddressID, shopOrder2.ShippingAddressID)
			require.Equal(t, shopOrder1.OrderTotal, shopOrder2.OrderTotal)
			require.Equal(t, shopOrder1.ShippingMethodID, shopOrder2.ShippingMethodID)
			require.NotEqual(t, shopOrder1.OrderStatusID, shopOrder2.OrderStatusID)
			require.NotEqual(t, shopOrder1.OrderStatusID, shopOrder2.OrderStatusID)
			require.NotZero(t, shopOrder2.CompletedAt)
			require.NotEqual(t, shopOrder1.CompletedAt, shopOrder2.CompletedAt)
			break
		}
	}
}

func TestDeleteShopOrder(t *testing.T) {
	shopOrder1 := createRandomShopOrder(t)
	err := testStore.DeleteShopOrder(context.Background(), shopOrder1.ID)

	require.NoError(t, err)

	shopOrder2, err := testStore.GetShopOrder(context.Background(), shopOrder1.ID)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, shopOrder2)

}

func TestListShopOrders(t *testing.T) {
	t.Parallel()
	// var wg sync.WaitGroup
	// wg.Add(10)
	for i := 0; i < 10; i++ {
		// go func(i int) {
		createRandomShopOrder(t)
		// wg.Done()
		// 	}(i)
	}
	// wg.Wait()
	arg := ListShopOrdersParams{
		Limit:  5,
		Offset: 0,
	}

	shopOrders, err := testStore.ListShopOrders(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, shopOrders)

	for _, shopOrder := range shopOrders {
		require.NotEmpty(t, shopOrder)
	}

}

func TestListShopOrdersByUserID(t *testing.T) {
	t.Parallel()
	// var wg sync.WaitGroup
	var shopOrder ShopOrder
	// wg.Add(10)
	for i := 0; i < 10; i++ {
		// go func(i int) {
		shopOrder = createRandomShopOrder(t)
		// wg.Done()
		// }(i)
	}
	// wg.Wait()
	arg := ListShopOrdersByUserIDParams{
		UserID: shopOrder.UserID,
		Limit:  5,
		Offset: 0,
	}

	shopOrders, err := testStore.ListShopOrdersByUserID(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, shopOrders)

	for _, shopOrder := range shopOrders {
		require.NotEmpty(t, shopOrder)
	}

}

func TestListShopOrdersByUserIDV2(t *testing.T) {

	shopOrder := createRandomShopOrderForListV2(t)

	fmt.Println(shopOrder.UserID)

	arg := ListShopOrdersByUserIDV2Params{
		UserID: shopOrder.UserID,
		Limit:  10,
	}

	initialSearchResult, err := testStore.ListShopOrdersByUserIDV2(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, initialSearchResult)
	require.Equal(t, len(initialSearchResult), 10)

	arg1 := ListShopOrdersByUserIDNextPageParams{
		UserID:      shopOrder.UserID,
		ShopOrderID: initialSearchResult[len(initialSearchResult)-1].ID,
		Limit:       10,
	}

	secondPage, err := testStore.ListShopOrdersByUserIDNextPage(context.Background(), arg1)
	require.NoError(t, err)
	require.NotEmpty(t, secondPage)
	require.Equal(t, len(initialSearchResult), 10)
	require.Greater(t, initialSearchResult[len(initialSearchResult)-1].ID, secondPage[len(secondPage)-1].ID)

	arg2 := ListShopOrdersByUserIDNextPageParams{
		UserID:      shopOrder.UserID,
		ShopOrderID: secondPage[len(initialSearchResult)-1].ID,
		Limit:       10,
	}

	thirdPage, err := testStore.ListShopOrdersByUserIDNextPage(context.Background(), arg2)
	require.NoError(t, err)
	require.NotEmpty(t, secondPage)
	require.Equal(t, len(initialSearchResult), 10)
	require.Greater(t, secondPage[len(initialSearchResult)-1].ID, thirdPage[len(secondPage)-1].ID)
	require.Greater(t, initialSearchResult[len(initialSearchResult)-1].ID, thirdPage[len(secondPage)-1].ID)
}

func TestGetShopOrdersCountByStatusId(t *testing.T) {
	admin := createRandomAdmin(t)
	shopOrder1 := createRandomShopOrder(t)

	arg := GetShopOrdersCountByStatusIdParams{
		OrderStatusID: shopOrder1.OrderStatusID,
		AdminID:       admin.ID,
	}
	shopOrder2, err := testStore.GetShopOrdersCountByStatusId(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, shopOrder2)
}

func TestGetTotalShopOrder(t *testing.T) {
	admin := createRandomAdmin(t)
	createRandomShopOrder(t)

	shopOrder2, err := testStore.GetTotalShopOrder(context.Background(), admin.ID)

	require.NoError(t, err)
	require.NotEmpty(t, shopOrder2)
}

func TestGetDailyOrderTotal(t *testing.T) {
	admin := createRandomAdmin(t)

	shopOrder2, err := testStore.GetCompletedDailyOrderTotal(context.Background(), admin.ID)

	require.NoError(t, err)
	require.NotEmpty(t, shopOrder2)
}

func TestAdminListShopOrdersV2(t *testing.T) {
	admin := createRandomAdmin(t)
	shopOrder := createRandomShopOrderForListV2(t)

	fmt.Println(shopOrder.UserID)

	arg := AdminListShopOrdersV2Params{
		AdminID: admin.ID,
		Limit:   10,
	}

	initialSearchResult, err := testStore.AdminListShopOrdersV2(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, initialSearchResult)
	require.Equal(t, len(initialSearchResult), 10)

	arg1 := AdminListShopOrdersNextPageParams{
		AdminID:     admin.ID,
		ShopOrderID: initialSearchResult[len(initialSearchResult)-1].ID,
		Limit:       10,
	}

	secondPage, err := testStore.AdminListShopOrdersNextPage(context.Background(), arg1)
	require.NoError(t, err)
	require.NotEmpty(t, secondPage)
	require.Equal(t, len(initialSearchResult), 10)
	require.Greater(t, initialSearchResult[len(initialSearchResult)-1].ID, secondPage[len(secondPage)-1].ID)

	arg2 := AdminListShopOrdersNextPageParams{
		AdminID:     admin.ID,
		ShopOrderID: secondPage[len(initialSearchResult)-1].ID,
		Limit:       10,
	}

	thirdPage, err := testStore.AdminListShopOrdersNextPage(context.Background(), arg2)
	require.NoError(t, err)
	require.NotEmpty(t, secondPage)
	require.Equal(t, len(initialSearchResult), 10)
	require.Greater(t, secondPage[len(initialSearchResult)-1].ID, thirdPage[len(secondPage)-1].ID)
	require.Greater(t, initialSearchResult[len(initialSearchResult)-1].ID, thirdPage[len(secondPage)-1].ID)
}
