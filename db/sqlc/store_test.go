package db

import (
	"context"
	"fmt"
	"log"
	"testing"

	"cshop.com/v2/util"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestFinishedPurchaseTx(t *testing.T) {
	store := NewStore(testDB)

	// run n concurrent purchases transaction
	n := 1
	Qty := int32(5)
	var userAddress UserAddress
	var productItem ProductItem
	var paymentType PaymentType
	var shippingMethod ShippingMethod
	var orderStatus OrderStatus
	var shoppingCart ShoppingCart
	var shoppingCartItem ShoppingCartItem
	var paymentMethod PaymentMethod
	var err error
	var price decimal.Decimal
	var totalPrice string

	errs := make(chan error)
	results := make(chan FinishedPurchaseTxResult)

	for i := 0; i < n; i++ {
		userAddress = createRandomUserAddress(t)
		paymentType = createRandomPaymentType(t)
		shippingMethod = createRandomShippingMethod(t)
		orderStatus = createRandomOrderStatus(t)

		paymentMethod, err = store.CreatePaymentMethod(context.Background(), CreatePaymentMethodParams{
			UserID:        userAddress.UserID,
			PaymentTypeID: int32(paymentType.ID),
			Provider:      util.RandomString(5),
			IsDefault:     true,
		})
		shoppingCart, err = store.CreateShoppingCart(context.Background(), userAddress.UserID)

		for i := 0; i < n; i++ {
			product := createRandomProduct(t)
			productItem, err = store.CreateProductItem(context.Background(), CreateProductItemParams{
				ProductID:    product.ID,
				ProductSku:   util.RandomInt(5, 100),
				QtyInStock:   50,
				ProductImage: util.RandomString(5),
				Price:        fmt.Sprint(util.RandomMoney()),
				Active:       true,
			})
			shoppingCartItem, err = store.CreateShoppingCartItem(context.Background(), CreateShoppingCartItemParams{
				ShoppingCartID: shoppingCart.ID,
				ProductItemID:  productItem.ID,
				Qty:            Qty,
			})
			if err != nil {
				log.Fatal("err is: ", err)
			}

			price, err = decimal.NewFromString(productItem.Price)
			if err != nil {
				log.Fatal("err is: ", err)
			}

			totalPrice += price.String()
		}
		go func() {
			result, err := store.FinishedPurchaseTx(context.Background(), FinishedPurchaseTxParams{
				UserAddress: UserAddress{
					UserID:    userAddress.UserID,
					AddressID: userAddress.AddressID,
					IsDefault: userAddress.IsDefault,
				},
				PaymentMethod: PaymentMethod{
					ID:            paymentMethod.ID,
					UserID:        paymentMethod.UserID,
					PaymentTypeID: paymentMethod.PaymentTypeID,
					Provider:      util.RandomString(5),
					IsDefault:     true,
				},
				ProductItem: ProductItem{
					ID:           productItem.ID,
					ProductID:    productItem.ProductID,
					ProductSku:   productItem.ProductSku,
					QtyInStock:   productItem.QtyInStock,
					ProductImage: util.RandomString(5),
					Price:        fmt.Sprint(util.RandomDecimal(0, 100)),
					Active:       true,
				},
				ShoppingCart: ShoppingCart{
					ID:     shoppingCart.ID,
					UserID: shoppingCart.UserID,
				},
				ShoppingCartItem: ShoppingCartItem{
					ID:             shoppingCartItem.ID,
					ShoppingCartID: shoppingCartItem.ShoppingCartID,
					ProductItemID:  shoppingCartItem.ProductItemID,
					Qty:            shoppingCartItem.Qty,
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

			errs <- err
			results <- result
		}()
	}

	// check results
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		// check finishedPurchase/ ShopOrder
		finishedShopOrder := result.ShopOrder
		require.NotEmpty(t, finishedShopOrder)
		require.Equal(t, userAddress.UserID, finishedShopOrder.UserID)
		require.Equal(t, userAddress.AddressID, finishedShopOrder.ShippingAddressID)
		require.Equal(t, paymentMethod.ID, finishedShopOrder.PaymentMethodID)
		require.Equal(t, shippingMethod.ID, finishedShopOrder.ShippingMethodID)
		require.Equal(t, orderStatus.ID, finishedShopOrder.OrderStatusID)

		_, err = store.GetShopOrder(context.Background(), finishedShopOrder.ID)
		require.NoError(t, err)

		// check ProductItem Updated Quantity
		newProductItem := result.ProductItem
		require.NotEmpty(t, newProductItem)
		require.NotEqual(t, productItem.QtyInStock, newProductItem.QtyInStock)
		require.Equal(t, productItem.QtyInStock-shoppingCartItem.Qty, newProductItem.QtyInStock)

		//check ShoppingCart, and ShopOrder
		finishedShopOrderItem := result.ShopOrderItem
		require.Equal(t, shoppingCartItem.ProductItemID, finishedShopOrderItem.ProductItemID)
		require.Equal(t, shoppingCartItem.Qty, finishedShopOrderItem.Quantity)

	}

}
