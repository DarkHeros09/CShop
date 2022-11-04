package db

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/cshop/v3/util"
	"github.com/guregu/null"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestFinishedPurchaseTx(t *testing.T) {
	store := NewStore(testDB)

	// run n concurrent purchases transaction
	n := 1
	Qty := int32(5)
	var userAddress UserAddress
	var listUsersAddress []UserAddress
	var productItem ProductItem
	var listProductItem []ProductItem
	var paymentType PaymentType
	var listPaymentType []PaymentType
	var shippingMethod ShippingMethod
	var listShippingMethod []ShippingMethod
	var orderStatus OrderStatus
	var listOrderStatus []OrderStatus
	var shoppingCart ShoppingCart
	// var listShoppingCart []ShoppingCart
	var shoppingCartItem ShoppingCartItem
	var listShoppingCartItem []ShoppingCartItem
	var paymentMethod PaymentMethod
	var listPaymentMethod []PaymentMethod
	var err error
	var price decimal.Decimal
	// var listPrice []decimal.Decimal
	var totalPrice string
	// var listTotalPrice []string

	errs := make(chan error)
	results := make(chan FinishedPurchaseTxResult)

	for i := 0; i < n; i++ {
		userAddress = createRandomUserAddress(t)
		listUsersAddress = append(listUsersAddress, userAddress)
		paymentType = createRandomPaymentType(t)
		listPaymentType = append(listPaymentType, paymentType)
		shippingMethod = createRandomShippingMethod(t)
		listShippingMethod = append(listShippingMethod, shippingMethod)
		orderStatus = createRandomOrderStatus(t)
		listOrderStatus = append(listOrderStatus, orderStatus)

		paymentMethod, err = store.CreatePaymentMethod(context.Background(), CreatePaymentMethodParams{
			UserID:        userAddress.UserID,
			PaymentTypeID: paymentType.ID,
			Provider:      util.RandomString(5),
		})
		if err != nil {
			log.Fatal("err is: ", err)
		}
		listPaymentMethod = append(listPaymentMethod, paymentMethod)

		shoppingCart, err = store.CreateShoppingCart(context.Background(), userAddress.UserID)
		if err != nil {
			log.Fatal("err is: ", err)
		}
		price = decimal.Zero
		for x := 0; x < n; x++ {
			product := createRandomProduct(t)
			productItem, err = store.CreateProductItem(context.Background(), CreateProductItemParams{
				ProductID:    product.ID,
				ProductSku:   util.RandomInt(5, 100),
				QtyInStock:   50,
				ProductImage: util.RandomString(5),
				Price:        util.RandomDecimalString(1, 100),
				Active:       true,
			})
			if err != nil {
				log.Fatal("err is: ", err)
			}
			listProductItem = append(listProductItem, productItem)

			shoppingCartItem, err = store.CreateShoppingCartItem(context.Background(), CreateShoppingCartItemParams{
				ShoppingCartID: shoppingCart.ID,
				ProductItemID:  productItem.ID,
				Qty:            Qty,
			})
			if err != nil {
				log.Fatal("err is: ", err)
			}
			listShoppingCartItem = append(listShoppingCartItem, shoppingCartItem)

			price, err = decimal.NewFromString(productItem.Price)
			if err != nil {
				log.Fatal("err is: ", err)
			}

			totalPrice += price.String()
			time.Sleep(3 * time.Second)
		}
		go func() {
			result, err := store.FinishedPurchaseTx(context.Background(), FinishedPurchaseTxParams{
				UserAddress: UserAddress{
					UserID:         userAddress.UserID,
					AddressID:      userAddress.AddressID,
					DefaultAddress: null.IntFromPtr(&userAddress.DefaultAddress.Int64),
				},
				PaymentMethod: PaymentMethod{
					ID:            paymentMethod.ID,
					UserID:        paymentMethod.UserID,
					PaymentTypeID: paymentMethod.PaymentTypeID,
					Provider:      util.RandomString(5),
					IsDefault:     true,
				},
				ShoppingCart: ShoppingCart{
					ID:     shoppingCart.ID,
					UserID: shoppingCart.UserID,
				},

				ShippingMethod: ShippingMethod{
					ID:    shippingMethod.ID,
					Name:  shippingMethod.Name,
					Price: shippingMethod.Price,
				},
				OrderStatus: OrderStatus{
					ID:     orderStatus.ID,
					Status: orderStatus.Status,
				},
				OrderTotal: totalPrice,
			})
			// time.Sleep(1 * time.Second)
			errs <- err
			results <- result
		}()
	}

	// check results
	var resultList []FinishedPurchaseTxResult
	// time.Sleep(1 * time.Second)
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)
		resultList = append(resultList, result)
		// check finishedPurchase/ ShopOrder
		finishedShopOrder := resultList[i].ShopOrder
		require.NotEmpty(t, finishedShopOrder)
		require.Equal(t, listUsersAddress[i].UserID, finishedShopOrder.UserID)
		require.Equal(t, listUsersAddress[i].AddressID, finishedShopOrder.ShippingAddressID)
		require.Equal(t, listPaymentMethod[i].ID, finishedShopOrder.PaymentMethodID)
		require.Equal(t, listShippingMethod[i].ID, finishedShopOrder.ShippingMethodID)
		require.Equal(t, listOrderStatus[i].ID, finishedShopOrder.OrderStatusID)

		_, err = testQueires.GetShopOrder(context.Background(), finishedShopOrder.ID)
		require.NoError(t, err)

		// check ProductItem Updated Quantity
		newProductItem := resultList[i].UpdatedProductItem
		require.NotEmpty(t, newProductItem)
		require.NotEqual(t, listProductItem[i].QtyInStock, newProductItem.QtyInStock)
		require.Equal(t, listProductItem[i].QtyInStock-listShoppingCartItem[i].Qty, newProductItem.QtyInStock)

		//check ShoppingCart, and ShopOrder
		finishedShopOrderItems, err := testQueires.ListShopOrderItemsByOrderID(context.Background(), finishedShopOrder.ID)
		require.NotEmpty(t, finishedShopOrderItems)
		require.NoError(t, err)
		for _, finishedShopOrderItem := range finishedShopOrderItems {
			require.Equal(t, listShoppingCartItem[i].ProductItemID, finishedShopOrderItem.ProductItemID)
			require.Equal(t, listShoppingCartItem[i].Qty, finishedShopOrderItem.Quantity)
		}

		arg1 := GetShoppingCartItemByUserIDCartIDParams{
			UserID:         shoppingCart.UserID,
			ShoppingCartID: shoppingCart.ID,
		}

		DeletedShopCartItem, err := testQueires.GetShoppingCartItemByUserIDCartID(context.Background(), arg1)
		require.Error(t, err)
		require.Empty(t, DeletedShopCartItem)
	}
}

func TestFinishedPurchaseTxFailedNotEnoughStock(t *testing.T) {
	store := NewStore(testDB)

	// run n concurrent purchases transaction
	n := 1
	Qty := int32(5)
	var userAddress UserAddress
	var listUsersAddress []UserAddress
	var productItem ProductItem
	var listProductItem []ProductItem
	var paymentType PaymentType
	var listPaymentType []PaymentType
	var shippingMethod ShippingMethod
	var listShippingMethod []ShippingMethod
	var orderStatus OrderStatus
	var listOrderStatus []OrderStatus
	var shoppingCart ShoppingCart
	// var listShoppingCart []ShoppingCart
	var shoppingCartItem ShoppingCartItem
	var listShoppingCartItem []ShoppingCartItem
	var paymentMethod PaymentMethod
	var listPaymentMethod []PaymentMethod
	var err error
	var price decimal.Decimal
	// var listPrice []decimal.Decimal
	var totalPrice string
	// var listTotalPrice []string

	errs := make(chan error)
	results := make(chan FinishedPurchaseTxResult)

	for i := 0; i < n; i++ {
		userAddress = createRandomUserAddress(t)
		listUsersAddress = append(listUsersAddress, userAddress)
		paymentType = createRandomPaymentType(t)
		listPaymentType = append(listPaymentType, paymentType)
		shippingMethod = createRandomShippingMethod(t)
		listShippingMethod = append(listShippingMethod, shippingMethod)
		orderStatus = createRandomOrderStatus(t)
		listOrderStatus = append(listOrderStatus, orderStatus)

		paymentMethod, err = store.CreatePaymentMethod(context.Background(), CreatePaymentMethodParams{
			UserID:        userAddress.UserID,
			PaymentTypeID: paymentType.ID,
			Provider:      util.RandomString(5),
		})
		if err != nil {
			log.Fatal("err is: ", err)
		}
		listPaymentMethod = append(listPaymentMethod, paymentMethod)

		shoppingCart, err = store.CreateShoppingCart(context.Background(), userAddress.UserID)
		if err != nil {
			log.Fatal("err is: ", err)
		}
		price = decimal.Zero
		for x := 0; x < n; x++ {
			product := createRandomProduct(t)
			productItem, err = store.CreateProductItem(context.Background(), CreateProductItemParams{
				ProductID:    product.ID,
				ProductSku:   util.RandomInt(5, 100),
				QtyInStock:   4,
				ProductImage: util.RandomString(5),
				Price:        util.RandomDecimalString(1, 100),
				Active:       true,
			})
			if err != nil {
				log.Fatal("err is: ", err)
			}
			listProductItem = append(listProductItem, productItem)

			shoppingCartItem, err = store.CreateShoppingCartItem(context.Background(), CreateShoppingCartItemParams{
				ShoppingCartID: shoppingCart.ID,
				ProductItemID:  productItem.ID,
				Qty:            Qty,
			})
			if err != nil {
				log.Fatal("err is: ", err)
			}
			listShoppingCartItem = append(listShoppingCartItem, shoppingCartItem)

			price, err = decimal.NewFromString(productItem.Price)
			if err != nil {
				log.Fatal("err is: ", err)
			}

			totalPrice += price.String()
			time.Sleep(3 * time.Second)
		}
		go func() {
			result, err := store.FinishedPurchaseTx(context.Background(), FinishedPurchaseTxParams{
				UserAddress: UserAddress{
					UserID:         userAddress.UserID,
					AddressID:      userAddress.AddressID,
					DefaultAddress: null.IntFromPtr(&userAddress.DefaultAddress.Int64),
				},
				PaymentMethod: PaymentMethod{
					ID:            paymentMethod.ID,
					UserID:        paymentMethod.UserID,
					PaymentTypeID: paymentMethod.PaymentTypeID,
					Provider:      util.RandomString(5),
					IsDefault:     true,
				},
				ShoppingCart: ShoppingCart{
					ID:     shoppingCart.ID,
					UserID: shoppingCart.UserID,
				},

				ShippingMethod: ShippingMethod{
					ID:    shippingMethod.ID,
					Name:  shippingMethod.Name,
					Price: shippingMethod.Price,
				},
				OrderStatus: OrderStatus{
					ID:     orderStatus.ID,
					Status: orderStatus.Status,
				},
				OrderTotal: totalPrice,
			})
			// time.Sleep(1 * time.Second)
			errs <- err
			results <- result
		}()
	}

	// check results
	// time.Sleep(1 * time.Second)
	for i := 0; i < n; i++ {
		err := <-errs
		require.Error(t, err)
		require.EqualError(t, err, "Not Enough Qty in Stock")

		result := <-results
		require.Empty(t, result)
	}
}

func TestFinishedPurchaseTxFailedEmptyStock(t *testing.T) {
	store := NewStore(testDB)

	// run n concurrent purchases transaction
	n := 1
	Qty := int32(5)
	var userAddress UserAddress
	var listUsersAddress []UserAddress
	var productItem ProductItem
	var listProductItem []ProductItem
	var paymentType PaymentType
	var listPaymentType []PaymentType
	var shippingMethod ShippingMethod
	var listShippingMethod []ShippingMethod
	var orderStatus OrderStatus
	var listOrderStatus []OrderStatus
	var shoppingCart ShoppingCart
	// var listShoppingCart []ShoppingCart
	var shoppingCartItem ShoppingCartItem
	var listShoppingCartItem []ShoppingCartItem
	var paymentMethod PaymentMethod
	var listPaymentMethod []PaymentMethod
	var err error
	var price decimal.Decimal
	// var listPrice []decimal.Decimal
	var totalPrice string
	// var listTotalPrice []string

	errs := make(chan error)
	results := make(chan FinishedPurchaseTxResult)

	for i := 0; i < n; i++ {
		userAddress = createRandomUserAddress(t)
		listUsersAddress = append(listUsersAddress, userAddress)
		paymentType = createRandomPaymentType(t)
		listPaymentType = append(listPaymentType, paymentType)
		shippingMethod = createRandomShippingMethod(t)
		listShippingMethod = append(listShippingMethod, shippingMethod)
		orderStatus = createRandomOrderStatus(t)
		listOrderStatus = append(listOrderStatus, orderStatus)

		paymentMethod, err = store.CreatePaymentMethod(context.Background(), CreatePaymentMethodParams{
			UserID:        userAddress.UserID,
			PaymentTypeID: paymentType.ID,
			Provider:      util.RandomString(5),
		})
		if err != nil {
			log.Fatal("err is: ", err)
		}
		listPaymentMethod = append(listPaymentMethod, paymentMethod)

		shoppingCart, err = store.CreateShoppingCart(context.Background(), userAddress.UserID)
		if err != nil {
			log.Fatal("err is: ", err)
		}
		price = decimal.Zero
		for x := 0; x < n; x++ {
			product := createRandomProduct(t)
			productItem, err = store.CreateProductItem(context.Background(), CreateProductItemParams{
				ProductID:    product.ID,
				ProductSku:   util.RandomInt(5, 100),
				QtyInStock:   0,
				ProductImage: util.RandomString(5),
				Price:        util.RandomDecimalString(1, 100),
				Active:       true,
			})
			if err != nil {
				log.Fatal("err is: ", err)
			}
			listProductItem = append(listProductItem, productItem)

			shoppingCartItem, err = store.CreateShoppingCartItem(context.Background(), CreateShoppingCartItemParams{
				ShoppingCartID: shoppingCart.ID,
				ProductItemID:  productItem.ID,
				Qty:            Qty,
			})
			if err != nil {
				log.Fatal("err is: ", err)
			}
			listShoppingCartItem = append(listShoppingCartItem, shoppingCartItem)

			price, err = decimal.NewFromString(productItem.Price)
			if err != nil {
				log.Fatal("err is: ", err)
			}

			totalPrice += price.String()
			time.Sleep(3 * time.Second)
		}
		go func() {
			result, err := store.FinishedPurchaseTx(context.Background(), FinishedPurchaseTxParams{
				UserAddress: UserAddress{
					UserID:         userAddress.UserID,
					AddressID:      userAddress.AddressID,
					DefaultAddress: null.IntFromPtr(&userAddress.DefaultAddress.Int64),
				},
				PaymentMethod: PaymentMethod{
					ID:            paymentMethod.ID,
					UserID:        paymentMethod.UserID,
					PaymentTypeID: paymentMethod.PaymentTypeID,
					Provider:      util.RandomString(5),
					IsDefault:     true,
				},
				ShoppingCart: ShoppingCart{
					ID:     shoppingCart.ID,
					UserID: shoppingCart.UserID,
				},

				ShippingMethod: ShippingMethod{
					ID:    shippingMethod.ID,
					Name:  shippingMethod.Name,
					Price: shippingMethod.Price,
				},
				OrderStatus: OrderStatus{
					ID:     orderStatus.ID,
					Status: orderStatus.Status,
				},
				OrderTotal: totalPrice,
			})
			// time.Sleep(1 * time.Second)
			errs <- err
			results <- result
		}()
	}

	// check results
	// time.Sleep(1 * time.Second)
	for i := 0; i < n; i++ {
		err := <-errs
		require.Error(t, err)
		require.EqualError(t, err, "Stock is Empty")

		result := <-results
		require.Empty(t, result)
	}
}
