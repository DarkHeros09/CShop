package db

import (
	"context"
	"math/rand/v2"
	"testing"

	"github.com/guregu/null/v5"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func createRandomOrderStatus(t *testing.T) OrderStatus {
	var err error
	orderStatuses := []string{"تحت الإجراء", "تم التسليم", "ملغي"}
	var orderStatus OrderStatus
	var orderStatusAll []OrderStatus
	orderStatuseList, err := testStore.ListOrderStatuses(context.Background() /* arg*/)

	require.NoError(t, err)
	if len(orderStatuseList) != 3 {

		for i := 0; i < len(orderStatuses); i++ {
			orderStatus, err = testStore.CreateOrderStatus(context.Background(), orderStatuses[i])
			require.NoError(t, err)
			require.NotEmpty(t, orderStatus)

			orderStatusAll = append(orderStatusAll, orderStatus)
		}
		// select rand int from 0 - 2

		r := rand.IntN(len(orderStatuses))
		// fmt.Println("THIS IS THE RANDOM NUMBER IN", r)
		return orderStatusAll[r]
	}
	// ln := len(orderStatuseList)
	// fmt.Println("LENGTH NUMBER OUT", ln)
	r := rand.IntN(len(orderStatuseList))
	// fmt.Println("THIS IS THE RANDOM NUMBER OUT", r)
	return orderStatuseList[r]
}
func TestCreateOrderStatus(t *testing.T) {
	createRandomOrderStatus(t)
}

func TestGetOrderStatus(t *testing.T) {
	orderStatus1 := createRandomOrderStatus(t)
	orderStatus2, err := testStore.GetOrderStatus(context.Background(), orderStatus1.ID)

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
		Status: null.StringFrom(orderStatus1.Status),
	}

	orderStatus2, err := testStore.UpdateOrderStatus(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, orderStatus2)

	require.Equal(t, orderStatus1.ID, orderStatus2.ID)
	require.Equal(t, orderStatus1.Status, orderStatus2.Status)
}

func TestDeleteOrderStatus(t *testing.T) {
	for {
		orderStatus1 := createRandomOrderStatus(t)
		if orderStatus1.Status == "ملغي" {
			err := testStore.DeleteOrderStatus(context.Background(), orderStatus1.ID)

			require.NoError(t, err)

			orderStatus2, err := testStore.GetOrderStatus(context.Background(), orderStatus1.ID)

			require.Error(t, err)
			require.EqualError(t, err, pgx.ErrNoRows.Error())
			require.Empty(t, orderStatus2)
			break
		}
	}

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

	orderStatuses, err := testStore.ListOrderStatuses(context.Background() /* arg*/)

	require.NoError(t, err)
	require.NotEmpty(t, orderStatuses)

	for _, orderStatus := range orderStatuses {
		require.NotEmpty(t, orderStatus)
	}

}

func TestAdminListOrderStatuses(t *testing.T) {
	admin := createRandomAdmin(t)
	orderStatuses1 := []string{"تحت الإجراء", "تم التسليم", "ملغي"}
	for i := 0; i < len(orderStatuses1); i++ {
		createRandomOrderStatus(t)
	}
	// arg := ListOrderStatusesParams{
	// 	Limit:  int32(len(orderStatuses1)),
	// 	Offset: int32(len(orderStatuses1)),
	// }

	orderStatuses, err := testStore.AdminListOrderStatuses(context.Background(), admin.ID)

	require.NoError(t, err)
	require.NotEmpty(t, orderStatuses)

	for _, orderStatus := range orderStatuses {
		require.NotEmpty(t, orderStatus)
	}

}
