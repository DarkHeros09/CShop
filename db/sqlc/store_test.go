package db

import (
	"context"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/cshop/v3/util"
	"github.com/quagmt/udecimal"
	"github.com/stretchr/testify/require"
)

// func TestSize(t *testing.T) {
// 	println(unsafe.Sizeof(FinishedPurchaseTxParams{}))
// }

func TestFinishedPurchaseTx(t *testing.T) {

	// store := NewStore(testDB)

	// run n concurrent purchases transaction
	n := 1
	// Qty := int32(5)
	var userAddress Address
	var listUsersAddress []Address
	var productItem *ProductItem
	// var listProductItem []ProductItem
	var paymentType PaymentType
	var listPaymentType []PaymentType
	var shippingMethod ShippingMethod
	var listShippingMethod []ShippingMethod
	var orderStatus OrderStatus
	var listOrderStatus []OrderStatus
	var shoppingCart *ShoppingCart
	// var listShoppingCart []ShoppingCart
	var shoppingCartItem *ShoppingCartItem
	var listShoppingCartItem []*ShoppingCartItem
	var paymentMethod *PaymentMethod
	var listPaymentMethod []*PaymentMethod
	var err error
	var price udecimal.Decimal
	// var listPrice []udecimal.Decimal
	var totalPrice string
	// var listTotalPrice []string

	errs := make(chan error)
	results := make(chan *FinishedPurchaseTxResult)
	var lock sync.Mutex

	for i := 0; i < n; i++ {
		userAddress = createRandomAddressWithUser(t)
		listUsersAddress = append(listUsersAddress, userAddress)
		paymentType = createRandomPaymentType(t)
		listPaymentType = append(listPaymentType, paymentType)
		shippingMethod = createRandomShippingMethod(t)
		listShippingMethod = append(listShippingMethod, shippingMethod)
		orderStatus = createRandomOrderStatus(t)
		listOrderStatus = append(listOrderStatus, orderStatus)

		paymentMethod, err = testStore.CreatePaymentMethod(context.Background(), CreatePaymentMethodParams{
			UserID:        userAddress.UserID,
			PaymentTypeID: paymentType.ID,
			Provider:      util.RandomString(5),
		})
		if err != nil {
			log.Fatal("err is: ", err)
		}
		listPaymentMethod = append(listPaymentMethod, paymentMethod)

		shoppingCart, err = testStore.CreateShoppingCart(context.Background(), userAddress.UserID)
		if err != nil {
			log.Fatal("err is: ", err)
		}
		price = udecimal.Zero
		for x := 0; x < n; x++ {
			product := createRandomProduct(t)
			image := createRandomProductImage(t)
			color := createRandomProductColor(t)
			productItem, err = testStore.CreateProductItem(context.Background(), CreateProductItemParams{
				ProductID: product.ID,
				// SizeID:     size.ID,
				ImageID:    image.ID,
				ColorID:    color.ID,
				ProductSku: util.RandomInt(5, 100),
				// QtyInStock: 50,
				Price:  util.RandomDecimalString(1, 100),
				Active: true,
			})
			if err != nil {
				log.Fatal("err is: ", err)
			}
			size := createRandomProductSizeWithItemID(t, productItem.ID)
			// listProductItem = append(listProductItem, productItem)

			shoppingCartItem, err = testStore.CreateShoppingCartItem(context.Background(), CreateShoppingCartItemParams{
				ShoppingCartID: shoppingCart.ID,
				ProductItemID:  size.ProductItemID,
				SizeID:         size.ID,
				Qty:            1,
			})
			if err != nil {
				log.Fatal("err is: ", err)
			}
			listShoppingCartItem = append(listShoppingCartItem, shoppingCartItem)

			price, err = udecimal.Parse(productItem.Price)
			if err != nil {
				log.Fatal("err is: ", err)
			}

			totalPrice += price.String()
			// time.Sleep(3 * time.Second)
		}
		go func() {
			lock.Lock()
			result, err := testStore.FinishedPurchaseTx(context.Background(), FinishedPurchaseTxParams{
				UserID:           userAddress.UserID,
				AddressID:        userAddress.ID,
				PaymentTypeID:    paymentType.ID,
				ShoppingCartID:   shoppingCart.ID,
				ShippingMethodID: shippingMethod.ID,
				OrderStatusID:    orderStatus.ID,
				OrderTotal:       totalPrice,
			})
			// fmt.Println("FAIL: ", err)
			// time.Sleep(1 * time.Second)
			errs <- err
			results <- result
			lock.Unlock()
		}()

		// check results
		var resultList []*FinishedPurchaseTxResult
		// time.Sleep(1 * time.Second)
		for z := 0; z < n; z++ {
			err := <-errs
			require.NoError(t, err)

			result := <-results
			require.NotEmpty(t, result)
			resultList = append(resultList, result)
			// check finishedPurchase/ ShopOrder
			finishedShopOrderID := resultList[z].ShopOrderID
			require.NotEmpty(t, finishedShopOrderID)
			finishedShopOrder, err := testStore.GetShopOrder(context.Background(), finishedShopOrderID)
			require.NoError(t, err)
			require.NotEmpty(t, finishedShopOrder)
			require.Equal(t, listUsersAddress[z].UserID, finishedShopOrder.UserID)
			require.Equal(t, listUsersAddress[z].ID, finishedShopOrder.ShippingAddressID.Int64)
			// require.Equal(t, listPaymentMethod[z].ID, finishedShopOrder.PaymentMethodID)
			require.Equal(t, listShippingMethod[z].ID, finishedShopOrder.ShippingMethodID)
			require.Equal(t, listOrderStatus[z].ID, finishedShopOrder.OrderStatusID.Int64)

			_, err = testStore.GetShopOrder(context.Background(), finishedShopOrder.ID)
			require.NoError(t, err)

			// check ProductItem Updated Quantity
			newProductSizeID := resultList[z].UpdatedProductSizeID
			require.NotEmpty(t, newProductSizeID)
			size, err := testStore.GetProductSize(context.Background(), newProductSizeID)
			require.NoError(t, err)
			newProductItemID := size.ProductItemID
			require.NotEmpty(t, newProductItemID)
			newProductItem, err := testStore.GetProductItem(context.Background(), newProductItemID)
			require.NotEmpty(t, newProductItem)
			// require.NotEqual(t, listProductItem[z].QtyInStock, newProductItem.QtyInStock)
			// require.Equal(t, listProductItem[z].QtyInStock-listShoppingCartItem[z].Qty, newProductItem.QtyInStock)

			// check ShoppingCart, and ShopOrder
			argF := ListShopOrderItemsByUserIDOrderIDParams{
				OrderID: finishedShopOrder.ID,
				UserID:  finishedShopOrder.UserID,
				// Limit:   10,
				// Offset:  0,
			}
			finishedShopOrderItems, err := testStore.ListShopOrderItemsByUserIDOrderID(context.Background(), argF)
			require.NotEmpty(t, finishedShopOrderItems)
			require.NoError(t, err)
			println(len(finishedShopOrderItems))
			println(len(listShoppingCartItem))
			for y := 0; y < len(finishedShopOrderItems); y++ {
				require.Equal(t, listShoppingCartItem[z].ProductItemID, finishedShopOrderItems[y].ProductItemID)
				require.Equal(t, listShoppingCartItem[z].Qty, finishedShopOrderItems[y].Quantity)
			}

			arg1 := GetShoppingCartItemByUserIDCartIDParams{
				UserID: shoppingCart.UserID,
				ID:     shoppingCart.ID,
			}

			DeletedShopCartItem, err := testStore.GetShoppingCartItemByUserIDCartID(context.Background(), arg1)
			require.NoError(t, err)
			require.Empty(t, DeletedShopCartItem)
		}
	}
}

func TestFinishedPurchaseTxFailedNotEnoughStock(t *testing.T) {

	// store := NewStore(testDB)

	// run n concurrent purchases transaction
	n := 1
	// Qty := int32(5)
	var userAddress Address
	var listUsersAddress []Address
	var productItem *ProductItem
	var listProductItem []ProductItem
	var paymentType PaymentType
	var listPaymentType []PaymentType
	var shippingMethod ShippingMethod
	var listShippingMethod []ShippingMethod
	var orderStatus OrderStatus
	var listOrderStatus []OrderStatus
	var shoppingCart *ShoppingCart
	// var listShoppingCart []ShoppingCart
	var shoppingCartItem *ShoppingCartItem
	var listShoppingCartItem []*ShoppingCartItem
	var paymentMethod *PaymentMethod
	var listPaymentMethod []*PaymentMethod
	var err error
	var price udecimal.Decimal
	// var listPrice []udecimal.Decimal
	var totalPrice string
	// var listTotalPrice []string

	// errs := make(chan error)
	// results := make(chan FinishedPurchaseTxResult)

	for i := 0; i < n; i++ {
		userAddress = createRandomAddressWithUser(t)
		listUsersAddress = append(listUsersAddress, userAddress)
		paymentType = createRandomPaymentType(t)
		listPaymentType = append(listPaymentType, paymentType)
		shippingMethod = createRandomShippingMethod(t)
		listShippingMethod = append(listShippingMethod, shippingMethod)
		orderStatus = createRandomOrderStatus(t)
		listOrderStatus = append(listOrderStatus, orderStatus)

		paymentMethod, err = testStore.CreatePaymentMethod(context.Background(), CreatePaymentMethodParams{
			UserID:        userAddress.UserID,
			PaymentTypeID: paymentType.ID,
			Provider:      util.RandomString(5),
		})
		if err != nil {
			log.Fatal("err is: ", err)
		}
		listPaymentMethod = append(listPaymentMethod, paymentMethod)

		shoppingCart, err = testStore.CreateShoppingCart(context.Background(), userAddress.UserID)
		if err != nil {
			log.Fatal("err is: ", err)
		}
		price = udecimal.Zero
		for x := 0; x < n; x++ {
			product := createRandomProduct(t)
			size := createRandomProductSizeWithQTY(t, 4)
			image := createRandomProductImage(t)
			color := createRandomProductColor(t)
			productItem, err = testStore.CreateProductItem(context.Background(), CreateProductItemParams{
				ProductID: product.ID,
				// SizeID:     size.ID,
				ImageID:    image.ID,
				ColorID:    color.ID,
				ProductSku: util.RandomInt(5, 100),
				// QtyInStock: 4,
				Price:  util.RandomDecimalString(1, 100),
				Active: true,
			})
			if err != nil {
				log.Fatal("err is: ", err)
			}
			listProductItem = append(listProductItem, *productItem)

			shoppingCartItem, err = testStore.CreateShoppingCartItem(context.Background(), CreateShoppingCartItemParams{
				ShoppingCartID: shoppingCart.ID,
				ProductItemID:  size.ProductItemID,
				SizeID:         size.ID,
				Qty:            size.Qty,
			})
			if err != nil {
				log.Fatal("err is: ", err)
			}
			listShoppingCartItem = append(listShoppingCartItem, shoppingCartItem)

			price, err = udecimal.Parse(productItem.Price)
			if err != nil {
				log.Fatal("err is: ", err)
			}

			totalPrice += price.String()
			time.Sleep(3 * time.Second)
		}
		// go func() {
		result, err := testStore.FinishedPurchaseTx(context.Background(), FinishedPurchaseTxParams{
			UserID:           userAddress.UserID,
			AddressID:        userAddress.ID,
			PaymentTypeID:    paymentType.ID,
			ShoppingCartID:   shoppingCart.ID,
			ShippingMethodID: shippingMethod.ID,
			OrderStatusID:    orderStatus.ID,
			OrderTotal:       totalPrice,
		})
		// time.Sleep(1 * time.Second)
		// 	errs <- err
		// 	results <- result
		// }()
		// }

		// check results
		// time.Sleep(1 * time.Second)
		// for i := 0; i < n; i++ {
		// err := <-errs
		require.Error(t, err)
		require.EqualError(t, err, "Not Enough Qty in Stock")

		// result := <-results
		require.Empty(t, result)
	}
}

func TestFinishedPurchaseTxFailedEmptyStock(t *testing.T) {

	// store := NewStore(testDB)

	// run n concurrent purchases transaction
	n := 3
	// Qty := int32(5)
	var userAddress Address
	var listUsersAddress []Address
	var productItem *ProductItem
	var listProductItem []ProductItem
	var paymentType PaymentType
	var listPaymentType []PaymentType
	var shippingMethod ShippingMethod
	var listShippingMethod []ShippingMethod
	var orderStatus OrderStatus
	var listOrderStatus []OrderStatus
	var shoppingCart *ShoppingCart
	// var listShoppingCart []ShoppingCart
	var shoppingCartItem *ShoppingCartItem
	var listShoppingCartItem []*ShoppingCartItem
	var paymentMethod *PaymentMethod
	listPaymentMethod := make([]*PaymentMethod, n)
	var err error
	var price udecimal.Decimal
	// var listPrice []udecimal.Decimal
	var totalPrice string
	// var listTotalPrice []string
	var result *FinishedPurchaseTxResult

	// errs := make(chan error)
	// results := make(chan FinishedPurchaseTxResult)

	for i := 0; i < n; i++ {
		userAddress = createRandomAddressWithUser(t)
		listUsersAddress = append(listUsersAddress, userAddress)
		paymentType = createRandomPaymentType(t)
		listPaymentType = append(listPaymentType, paymentType)
		shippingMethod = createRandomShippingMethod(t)
		listShippingMethod = append(listShippingMethod, shippingMethod)
		orderStatus = createRandomOrderStatus(t)
		listOrderStatus = append(listOrderStatus, orderStatus)

		paymentMethod, err = testStore.CreatePaymentMethod(context.Background(), CreatePaymentMethodParams{
			UserID:        userAddress.UserID,
			PaymentTypeID: paymentType.ID,
			Provider:      util.RandomString(5),
		})
		if err != nil {
			log.Fatal("err is: ", err)
		}
		listPaymentMethod[i] = paymentMethod

		shoppingCart, err = testStore.CreateShoppingCart(context.Background(), userAddress.UserID)
		if err != nil {
			log.Fatal("err is: ", err)
		}
		price = udecimal.Zero
		for x := 0; x < n; x++ {
			product := createRandomProduct(t)
			size := createRandomProductSizeWithQTY(t, 0)
			image := createRandomProductImage(t)
			color := createRandomProductColor(t)
			productItem, err = testStore.CreateProductItem(context.Background(), CreateProductItemParams{
				ProductID: product.ID,
				// SizeID:     size.ID,
				ImageID:    image.ID,
				ColorID:    color.ID,
				ProductSku: util.RandomInt(5, 100),
				// QtyInStock: 0,
				Price:  util.RandomDecimalString(1, 100),
				Active: true,
			})
			if err != nil {
				log.Fatal("err is: ", err)
			}
			listProductItem = append(listProductItem, *productItem)

			shoppingCartItem, err = testStore.CreateShoppingCartItem(context.Background(), CreateShoppingCartItemParams{
				ShoppingCartID: shoppingCart.ID,
				ProductItemID:  size.ProductItemID,
				SizeID:         size.ID,
				Qty:            size.Qty,
			})
			if err != nil {
				log.Fatal("err is: ", err)
			}
			listShoppingCartItem = append(listShoppingCartItem, shoppingCartItem)

			price, err = udecimal.Parse(productItem.Price)
			if err != nil {
				log.Fatal("err is: ", err)
			}

			totalPrice += price.String()
			// time.Sleep(3 * time.Second)
		}
		// go func() {
		result, err = testStore.FinishedPurchaseTx(context.Background(), FinishedPurchaseTxParams{
			UserID:           userAddress.UserID,
			AddressID:        userAddress.ID,
			PaymentTypeID:    paymentType.ID,
			ShoppingCartID:   shoppingCart.ID,
			ShippingMethodID: shippingMethod.ID,
			OrderStatusID:    orderStatus.ID,
			OrderTotal:       totalPrice,
		})
		// time.Sleep(1 * time.Second)
		// errs <- err
		// results <- result
		// }()
	}

	// check results
	// time.Sleep(1 * time.Second)
	for i := 0; i < n; i++ {
		// err := <-errs
		require.Error(t, err)
		require.EqualError(t, err, "Stock is Empty")

		// result := <-results
		require.Empty(t, result)
	}
}

func TestDeleteShopOrderItemTx(t *testing.T) {
	admin := createRandomAdmin(t)
	shopOrderItem, shopOrder := createRandomShopOrderItem(t)

	err := testStore.DeleteShopOrderItemTx(context.Background(), DeleteShopOrderItemTxParams{
		ShopOrderItemID: shopOrderItem.ID,
		AdminID:         admin.ID,
	})

	require.NoError(t, err)

	deletedShopOrderItem, err := testStore.GetShopOrderItem(context.Background(), shopOrderItem.ID)

	require.Error(t, err)
	require.Empty(t, deletedShopOrderItem)

	updatedShopOrder, err := testStore.GetShopOrder(context.Background(), shopOrder.ID)

	require.NoError(t, err)
	require.NotEqual(t, updatedShopOrder, shopOrder)
	// require.NotEqual(t, updatedShopOrder.OrderTotal, shopOrder.OrderTotal)

}
