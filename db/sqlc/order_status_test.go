package db

import (
	"context"
	"testing"

	"github.com/guregu/null"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/require"
)

func createRandomOrderStatus(t *testing.T) OrderStatus {
	var orderStatus OrderStatus
	var err error
	orderStatuses := []string{"تحت الإجراء", "تم التسليم", "ملغي"}
	for i := 0; i < len(orderStatuses); i++ {
		orderStatus, err = testQueires.CreateOrderStatus(context.Background(), orderStatuses[i])
		require.NoError(t, err)
		require.NotEmpty(t, orderStatus)
	}
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
		ID: orderStatus1.ID,
		//last value from orderStatus1 otherwise will get duplicate keys becuase of unique constraint
		Status: null.StringFrom("ملغي"),
	}

	orderStatus2, err := testQueires.UpdateOrderStatus(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, orderStatus2)

	require.Equal(t, orderStatus1.ID, orderStatus2.ID)
	require.Equal(t, orderStatus1.Status, orderStatus2.Status)
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
	orderStatuses1 := []string{"تحت الإجراء", "تم التسليم", "ملغي"}
	for i := 0; i < len(orderStatuses1); i++ {
		createRandomOrderStatus(t)
	}
	// arg := ListOrderStatusesParams{
	// 	Limit:  int32(len(orderStatuses1)),
	// 	Offset: int32(len(orderStatuses1)),
	// }

	orderStatuses, err := testQueires.ListOrderStatuses(context.Background() /* arg*/)

	require.NoError(t, err)
	require.NotEmpty(t, orderStatuses)

	for _, orderStatus := range orderStatuses {
		require.NotEmpty(t, orderStatus)
	}

}
