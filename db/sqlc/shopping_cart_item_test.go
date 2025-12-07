package db

import (
	"context"
	"testing"

	"github.com/cshop/v3/util"
	"github.com/guregu/null/v6"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func createRandomShoppingCartItem(t *testing.T) (ShoppingCartItem, ShoppingCart) {
	shoppingCart := createRandomShoppingCart(t)
	// productItem := createRandomProductItem(t)
	size := createRandomProductSize(t)
	// var shoppingCartItems []ShoppingCartItem
	arg := CreateShoppingCartItemParams{
		ShoppingCartID: shoppingCart.ID,
		ProductItemID:  size.ProductItemID,
		SizeID:         size.ID,
		Qty:            int32(util.RandomInt(0, 10)),
	}

	// result := testStore.CreateShoppingCartItem(context.Background(), arg)

	// result.Query(func(i int, sci []ShoppingCartItem, err error) {
	// 	require.NoError(t, err)
	// 	require.NotEmpty(t, sci)
	// 	require.Equal(t, arg[i].ShoppingCartID, sci[i].ShoppingCartID)
	// 	require.Equal(t, arg[i].ProductItemID, sci[i].ProductItemID)
	// 	require.Equal(t, arg[i].Qty, sci[i].Qty)
	// 	shoppingCartItems = sci
	// })

	shoppingCartItem, err := testStore.CreateShoppingCartItem(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, shoppingCartItem)

	require.Equal(t, arg.ShoppingCartID, shoppingCartItem.ShoppingCartID)
	require.Equal(t, arg.ProductItemID, shoppingCartItem.ProductItemID)
	require.Equal(t, arg.Qty, shoppingCartItem.Qty)

	return *shoppingCartItem, shoppingCart
}

func TestCreateShoppingCartItem(t *testing.T) {
	createRandomShoppingCartItem(t)
}

func TestGetShoppingCartItem(t *testing.T) {
	shoppingCartItem1, _ := createRandomShoppingCartItem(t)

	shoppingCartItem2, err := testStore.GetShoppingCartItem(context.Background(), shoppingCartItem1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, shoppingCartItem2)

	require.Equal(t, shoppingCartItem1.ShoppingCartID, shoppingCartItem2.ShoppingCartID)
	require.Equal(t, shoppingCartItem1.ProductItemID, shoppingCartItem2.ProductItemID)
	require.Equal(t, shoppingCartItem1.Qty, shoppingCartItem2.Qty)
}

// func TestGetShoppingCartItemByCartID(t *testing.T) {
// 	shoppingCartItem1, shoppingCart1 := createRandomShoppingCartItem(t)

// 	arg := GetShoppingCartItemByUserIDCartIDParams{
// 		UserID: shoppingCart1.UserID,
// 		ID:     shoppingCartItem1.ShoppingCartID,
// 	}

// 	shoppingCartItem2, err := testStore.GetShoppingCartItemByUserIDCartID(context.Background(), arg)
// 	require.NoError(t, err)
// 	require.NotEmpty(t, shoppingCartItem2)

// 	require.Equal(t, shoppingCartItem1.ShoppingCartID, shoppingCartItem2.ShoppingCartID)
// 	require.Equal(t, shoppingCartItem1.ProductItemID, shoppingCartItem2.ProductItemID)
// 	require.Equal(t, shoppingCartItem1.Qty, shoppingCartItem2.Qty)
// }

func TestUpdateShoppingCartItem(t *testing.T) {

	shoppingCartItem1, _ := createRandomShoppingCartItem(t)
	arg := UpdateShoppingCartItemParams{
		Qty:            null.IntFrom(int64(shoppingCartItem1.Qty) + 1),
		ID:             shoppingCartItem1.ID,
		ShoppingCartID: shoppingCartItem1.ShoppingCartID,
	}

	shoppingCartItem2, err := testStore.UpdateShoppingCartItem(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, shoppingCartItem2)

	require.Equal(t, shoppingCartItem1.ShoppingCartID, shoppingCartItem2.ShoppingCartID)
	require.Equal(t, shoppingCartItem1.ProductItemID, shoppingCartItem2.ProductItemID)
	require.NotEqual(t, shoppingCartItem1.Qty, shoppingCartItem2.Qty)
}

func TestDeleteShoppingCartItem(t *testing.T) {

	shoppingCartItem1, shoppingCart := createRandomShoppingCartItem(t)

	arg := DeleteShoppingCartItemParams{
		ShoppingCartItemID: shoppingCartItem1.ID,
		UserID:             shoppingCart.UserID,
		ShoppingCartID:     shoppingCart.ID,
	}

	err := testStore.DeleteShoppingCartItem(context.Background(), arg)

	require.NoError(t, err)

	shoppingCartItem2, err := testStore.GetShoppingCartItem(context.Background(), shoppingCartItem1.ID)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, shoppingCartItem2)

}

func TestListShoppingCartItemes(t *testing.T) {

	shoppingCart := createRandomShoppingCart(t)
	// var wg sync.WaitGroup

	// wg.Add(10)

	for i := 0; i < 5; i++ {
		// go func(i int) {
		// 	fmt.Println("LOOP: ", i)
		productItem := createRandomProductItem(t)
		Qty := int32(util.RandomInt(5, 10))
		arg := CreateShoppingCartItemParams{
			ShoppingCartID: shoppingCart.ID,
			ProductItemID:  productItem.ID,
			Qty:            Qty,
		}
		// cartItemsChan := make(chan ShoppingCartItem)
		// errChan := make(chan error)
		// go func() {
		cartItems, err := testStore.CreateShoppingCartItem(context.Background(), arg)
		// cartItemsChan <- cartItems
		// errChan <- err
		// wg.Done()
		// }()
		// cartItems := <-cartItemsChan
		// err := <-errChan
		require.NoError(t, err)
		require.NotEmpty(t, cartItems)
		// wg.Done()

		// }(i)

	}

	// wg.Wait()

	arg := ListShoppingCartItemsParams{
		Limit:  5,
		Offset: 0,
	}
	shoppingCartItemsChan := make(chan []*ShoppingCartItem)
	errChan := make(chan error)
	go func() {
		shoppingCartItems, err := testStore.ListShoppingCartItems(context.Background(), arg)
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
	for i := 0; i < 5; i++ {
		// time.Sleep(time.Millisecond * 500)
		productItem := createRandomProductItem(t)
		arg := CreateShoppingCartItemParams{
			ShoppingCartID: shoppingCart.ID,
			ProductItemID:  productItem.ID,
			Qty:            int32(util.RandomInt(5, 10)),
		}
		testStore.CreateShoppingCartItem(context.Background(), arg)

	}

	shoppingCartItems, err := testStore.ListShoppingCartItemsByCartID(context.Background(), shoppingCart.ID)
	require.NoError(t, err)
	require.Len(t, shoppingCartItems, 5)

	for _, shoppingCartItem := range shoppingCartItems {
		require.NotEmpty(t, shoppingCartItem)

	}
}

func TestListShoppingCartItemsByUserID(t *testing.T) {

	// var shoppingCartItems []ListShoppingCartItemsByUserIDRow
	// var totalShoppingCartItems []ListShoppingCartItemsByUserIDRow
	// shoppingCartItemsChan := make(chan []ListShoppingCartItemsByUserIDRow)
	// errChan := make(chan error)
	// var shoppingCart ShoppingCart
	shoppingCart := createRandomShoppingCart(t)
	for i := 0; i < 5; i++ {
		// time.Sleep(time.Millisecond * 500)
		productItem := createRandomProductItem(t)
		arg := CreateShoppingCartItemParams{
			ShoppingCartID: shoppingCart.ID,
			ProductItemID:  productItem.ID,
			Qty:            int32(util.RandomInt(5, 10)),
		}
		testStore.CreateShoppingCartItem(context.Background(), arg)

	}

	// go func() {
	// 	for i := 0; i < 5; i++ {
	// 		productItem := createRandomProductItem(t)
	// 		arg := []CreateShoppingCartItemParams{
	// 			{ShoppingCartID: shoppingCart.ID,
	// 				ProductItemID: productItem.ID,
	// 				Qty:           int32(util.RandomInt(1, 10))},
	// 		}
	// 		result := testStore.CreateShoppingCartItem(context.Background(), arg)
	// 		result.Query(func(i int, sci []ShoppingCartItem, err error) {

	// 			require.Equal(t, arg[i].ProductItemID, sci[i].ProductItemID)
	// 			require.NoError(t, err)
	// 		})

	// 		// totalShoppingCartItems = append(totalShoppingCartItems, shoppingCartItems...)
	// 	}
	// }()
	// go func() {
	shoppingCartItems, err := testStore.ListShoppingCartItemsByUserID(context.Background(), shoppingCart.UserID)
	// shoppingCartItemsChan <- shoppingCartItems
	// errChan <- err

	// }()
	// shoppingCartItems := <-shoppingCartItemsChan
	// err := <-errChan

	require.NoError(t, err)
	// require.Len(t, totalShoppingCartItems, 10)

	for _, shoppingCartItem := range shoppingCartItems {
		require.NotEmpty(t, shoppingCartItem)

	}
}

func TestDeleteALLShoppingCartItemes(t *testing.T) {

	shoppingCart := createRandomShoppingCart(t)
	for i := 0; i < 5; i++ {
		productItem := createRandomProductItem(t)
		arg := CreateShoppingCartItemParams{
			ShoppingCartID: shoppingCart.ID,
			ProductItemID:  productItem.ID,
			Qty:            int32(util.RandomInt(1, 10)),
		}
		testStore.CreateShoppingCartItem(context.Background(), arg)

	}

	arg1 := DeleteShoppingCartItemAllByUserParams{
		UserID:         shoppingCart.UserID,
		ShoppingCartID: shoppingCart.ID,
	}
	shoppingCartItem1, err := testStore.DeleteShoppingCartItemAllByUser(context.Background(), arg1)
	require.NoError(t, err)
	require.NotEmpty(t, shoppingCartItem1)

	// shoppingCartItem2, err := testStore.DeleteShoppingCartItemAllByUser(context.Background(), arg1)
	// require.Error(t, err)
	// require.Empty(t, shoppingCartItem2)
	// require.EqualError(t, err, pgx.ErrNoRows.Error())
}
