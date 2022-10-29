package db

import (
	"context"
	"testing"

	"github.com/cshop/v3/util"
	"github.com/guregu/null"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/require"
)

func createRandomOrderStatus(t *testing.T) OrderStatus {
	orderStatus, err := testQueires.CreateOrderStatus(context.Background(), util.RandomString(5))
	require.NoError(t, err)
	require.NotEmpty(t, orderStatus)

	return orderStatus
}
func TestCreateOrderStatus(t *testing.T) {
	createRandomOrderStatus(t)
}

func TestGetOrderStatus(t *testing.T) {
	orderStatus1 := createRandomOrderStatus(t)
	orderStatus2, err := testQueires.GetOrderStatus(context.Background(), orderStatus1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, orderStatus2)

	require.Equal(t, orderStatus1.ID, orderStatus2.ID)
	require.Equal(t, orderStatus1.Status, orderStatus2.Status)
}

func TestUpdateOrderStatusNameAndPrice(t *testing.T) {
	orderStatus1 := createRandomOrderStatus(t)
	arg := UpdateOrderStatusParams{
		ID:     orderStatus1.ID,
		Status: null.StringFrom(util.RandomString(5)),
	}

	orderStatus2, err := testQueires.UpdateOrderStatus(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, orderStatus2)

	require.Equal(t, orderStatus1.ID, orderStatus2.ID)
	require.NotEqual(t, orderStatus1.Status, orderStatus2.Status)
}

func TestDeleteOrderStatus(t *testing.T) {
	orderStatus1 := createRandomOrderStatus(t)
	err := testQueires.DeleteOrderStatus(context.Background(), orderStatus1.ID)

	require.NoError(t, err)

	orderStatus2, err := testQueires.GetOrderStatus(context.Background(), orderStatus1.ID)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, orderStatus2)

}

func TestListOrderStatuses(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomOrderStatus(t)
	}
	arg := ListOrderStatusesParams{
		Limit:  5,
		Offset: 5,
	}

	orderStatuses, err := testQueires.ListOrderStatuses(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, orderStatuses)

	for _, orderStatus := range orderStatuses {
		require.NotEmpty(t, orderStatus)
	}

}
