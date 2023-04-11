package db

import (
	"context"
	"testing"

	"github.com/cshop/v3/util"
	"github.com/guregu/null"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/require"
)

func createRandomShoppingCartItem(t *testing.T) ([]ShoppingCartItem, ShoppingCart) {
	shoppingCart := createRandomShoppingCart(t)
	productItem := createRandomProductItem(t)
	var shoppingCartItems []ShoppingCartItem
	arg := []CreateShoppingCartItemParams{
		{ShoppingCartID: shoppingCart.ID,
			ProductItemID: productItem.ID,
			Qty:           int32(util.RandomInt(0, 10)),
		}}

	result := testQueires.CreateShoppingCartItem(context.Background(), arg)

	result.Query(func(i int, sci []ShoppingCartItem, err error) {
		require.NoError(t, err)
		require.NotEmpty(t, sci)
		require.Equal(t, arg[i].ShoppingCartID, sci[i].ShoppingCartID)
		require.Equal(t, arg[i].ProductItemID, sci[i].ProductItemID)
		require.Equal(t, arg[i].Qty, sci[i].Qty)
		shoppingCartItems = sci
	})

	// shoppingCartItem, err := testQueires.CreateShoppingCartItem(context.Background(), arg)
	// require.NoError(t, err)
	// require.NotEmpty(t, shoppingCartItem)

	// require.Equal(t, arg.ShoppingCartID, shoppingCartItem.ShoppingCartID)
	// require.Equal(t, arg.ProductItemID, shoppingCartItem.ProductItemID)
	// require.Equal(t, arg.Qty, shoppingCartItem.Qty)

	return shoppingCartItems, shoppingCart
}

func TestCreateShoppingCartItem(t *testing.T) {
	createRandomShoppingCartItem(t)
}

func TestGetShoppingCartItem(t *testing.T) {
	shoppingCartItem1, _ := createRandomShoppingCartItem(t)

	shoppingCartItem2, err := testQueires.GetShoppingCartItem(context.Background(), shoppingCartItem1[0].ID)
	require.NoError(t, err)
	require.NotEmpty(t, shoppingCartItem2)

	require.Equal(t, shoppingCartItem1[0].ShoppingCartID, shoppingCartItem2.ShoppingCartID)
	require.Equal(t, shoppingCartItem1[0].ProductItemID, shoppingCartItem2.ProductItemID)
	require.Equal(t, shoppingCartItem1[0].Qty, shoppingCartItem2.Qty)
}

func TestGetShoppingCartItemByCartID(t *testing.T) {
	shoppingCartItem1, shoppingCart1 := createRandomShoppingCartItem(t)

	arg := GetShoppingCartItemByUserIDCartIDParams{
		UserID: shoppingCart1.UserID,
		ID:     shoppingCartItem1[0].ShoppingCartID,
	}

	shoppingCartItem2, err := testQueires.GetShoppingCartItemByUserIDCartID(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, shoppingCartItem2)

	require.Equal(t, shoppingCartItem1[0].ShoppingCartID, shoppingCartItem2[0].ShoppingCartID)
	require.Equal(t, shoppingCartItem1[0].ProductItemID, shoppingCartItem2[0].ProductItemID)
	require.Equal(t, shoppingCartItem1[0].Qty, shoppingCartItem2[0].Qty)
}

func TestUpdateShoppingCartItem(t *testing.T) {
	shoppingCartItem1, _ := createRandomShoppingCartItem(t)
	arg := UpdateShoppingCartItemParams{
		ProductItemID:  null.Int{},
		Qty:            null.Int{},
		ID:             shoppingCartItem1[0].ID,
		ShoppingCartID: shoppingCartItem1[0].ShoppingCartID,
	}

	shoppingCartItem2, err := testQueires.UpdateShoppingCartItem(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, shoppingCartItem2)

	require.Equal(t, shoppingCartItem1[0].ShoppingCartID, shoppingCartItem2.ShoppingCartID)
	require.Equal(t, shoppingCartItem1[0].ProductItemID, shoppingCartItem2.ProductItemID)
	require.Equal(t, shoppingCartItem1[0].Qty, shoppingCartItem2.Qty)
}

func TestDeleteShoppingCartItem(t *testing.T) {
	shoppingCartItem1, shoppingCart := createRandomShoppingCartItem(t)

	arg := DeleteShoppingCartItemParams{
		ShoppingCartItemID: shoppingCartItem1[0].ID,
		UserID:             shoppingCart.UserID,
		ShoppingCartID:     shoppingCart.ID,
	}

	err := testQueires.DeleteShoppingCartItem(context.Background(), arg)

	require.NoError(t, err)

	shoppingCartItem2, err := testQueires.GetShoppingCartItem(context.Background(), shoppingCartItem1[0].ID)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, shoppingCartItem2)

}

func TestListShoppingCartItemes(t *testing.T) {
	shoppingCart := createRandomShoppingCart(t)
	nChan := make(chan int32)
	go func() {
		for i := 0; i < 5; i++ {
			productItem := createRandomProductItem(t)
			arg := []CreateShoppingCartItemParams{
				{ShoppingCartID: shoppingCart.ID,
					ProductItemID: productItem.ID,
					Qty:           int32(util.RandomInt(5, 10))},
			}
			go testQueires.CreateShoppingCartItem(context.Background(), arg)

		}
		nChan <- 5
	}()

	n := <-nChan
	arg := ListShoppingCartItemsParams{
		Limit:  n,
		Offset: 0,
	}
	shoppingCartItemsChan := make(chan []ShoppingCartItem)
	errChan := make(chan error)
	go func() {
		shoppingCartItems, err := testQueires.ListShoppingCartItems(context.Background(), arg)
		shoppingCartItemsChan <- shoppingCartItems
		errChan <- err
	}()
	shoppingCartItems := <-shoppingCartItemsChan
	err := <-errChan
	require.NoError(t, err)
	require.Len(t, shoppingCartItems, 5)

	for _, shoppingCartItem := range shoppingCartItems {
		require.NotEmpty(t, shoppingCartItem)

	}
}

func TestListShoppingCartItemsByCartID(t *testing.T) {
	shoppingCart := createRandomShoppingCart(t)
	for i := 0; i < 10; i++ {
		// time.Sleep(time.Millisecond * 500)
		productItem := createRandomProductItem(t)
		arg := []CreateShoppingCartItemParams{
			{ShoppingCartID: shoppingCart.ID,
				ProductItemID: productItem.ID,
				Qty:           int32(util.RandomInt(5, 10))},
		}
		result := testQueires.CreateShoppingCartItem(context.Background(), arg)
		result.Query(func(i int, sci []ShoppingCartItem, err error) {
			require.NoError(t, err)
		})
	}

	shoppingCartItems, err := testQueires.ListShoppingCartItemsByCartID(context.Background(), shoppingCart.ID)
	require.NoError(t, err)
	require.Len(t, shoppingCartItems, 10)

	for _, shoppingCartItem := range shoppingCartItems {
		require.NotEmpty(t, shoppingCartItem)

	}
}

func TestListShoppingCartItemsByUserID(t *testing.T) {
	// var shoppingCartItems []ListShoppingCartItemsByUserIDRow
	// var totalShoppingCartItems []ListShoppingCartItemsByUserIDRow
	shoppingCartItemsChan := make(chan []ListShoppingCartItemsByUserIDRow)
	errChan := make(chan error)
	// var shoppingCart ShoppingCart
	shoppingCart := createRandomShoppingCart(t)

	// go func() {
	// 	for i := 0; i < 5; i++ {
	// 		productItem := createRandomProductItem(t)
	// 		arg := []CreateShoppingCartItemParams{
	// 			{ShoppingCartID: shoppingCart.ID,
	// 				ProductItemID: productItem.ID,
	// 				Qty:           int32(util.RandomInt(1, 10))},
	// 		}
	// 		result := testQueires.CreateShoppingCartItem(context.Background(), arg)
	// 		result.Query(func(i int, sci []ShoppingCartItem, err error) {

	// 			require.Equal(t, arg[i].ProductItemID, sci[i].ProductItemID)
	// 			require.NoError(t, err)
	// 		})

	// 		// totalShoppingCartItems = append(totalShoppingCartItems, shoppingCartItems...)
	// 	}
	// }()
	go func() {
		shoppingCartItems, err := testQueires.ListShoppingCartItemsByUserID(context.Background(), shoppingCart.UserID)
		shoppingCartItemsChan <- shoppingCartItems
		errChan <- err

	}()
	shoppingCartItems := <-shoppingCartItemsChan
	err := <-errChan

	require.NoError(t, err)
	// require.Len(t, totalShoppingCartItems, 10)

	for _, shoppingCartItem := range shoppingCartItems {
		require.NotEmpty(t, shoppingCartItem)

	}
}

func TestDeleteALLShoppingCartItemes(t *testing.T) {
	shoppingCart := createRandomShoppingCart(t)
	for i := 0; i < 10; i++ {
		productItem := createRandomProductItem(t)
		arg := []CreateShoppingCartItemParams{
			{ShoppingCartID: shoppingCart.ID,
				ProductItemID: productItem.ID,
				Qty:           int32(util.RandomInt(1, 10))},
		}
		result := testQueires.CreateShoppingCartItem(context.Background(), arg)
		result.Query(func(i int, sci []ShoppingCartItem, err error) {

			require.Equal(t, arg[i].ProductItemID, sci[i].ProductItemID)
			require.NoError(t, err)
		})
	}

	arg1 := DeleteShoppingCartItemAllByUserParams{
		UserID:         shoppingCart.UserID,
		ShoppingCartID: shoppingCart.ID,
	}
	shoppingCartItem1, err := testQueires.DeleteShoppingCartItemAllByUser(context.Background(), arg1)
	require.NoError(t, err)
	require.NotEmpty(t, shoppingCartItem1)

	// shoppingCartItem2, err := testQueires.DeleteShoppingCartItemAllByUser(context.Background(), arg1)
	// require.Error(t, err)
	// require.Empty(t, shoppingCartItem2)
	// require.EqualError(t, err, pgx.ErrNoRows.Error())
}
