package db

import (
	"context"
	"fmt"

	"github.com/cshop/v3/util"
	"github.com/guregu/null"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

// Store provides all functions to execute db queries and transactions
type Store interface {
	Querier
	FinishedPurchaseTx(ctx context.Context, arg FinishedPurchaseTxParams) (FinishedPurchaseTxResult, error)
}

// Store provides all functions to execute db queries and transactions
type SQLStore struct {
	*Queries
	db *pgxpool.Pool
}

// NewStore creates a new Store
func NewStore(db *pgxpool.Pool) Store {
	return &SQLStore{
		db:      db,
		Queries: New(db),
	}
}

// execTx executes a function within a database transaction
// method starts with lower case to not be exported so external packages can't call it directly
// we will provide an exported function for each specific transaction
func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.Begin(ctx)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("transaction error: %v, rollback error: %v", err, rbErr)
		}
		return err
	}
	return tx.Commit(ctx)

}

// FinishedPurchaseTx contains the input parameters of the purchase transaction
type FinishedPurchaseTxParams struct {
	UserID        int64 `json:"user_id"`
	UserAddressID int64 `json:"user_address_id"`
	// PaymentMethodID  int64  `json:"payment_method_id"`
	PaymentTypeID    int64  `json:"payment_type_id"`
	ShoppingCartID   int64  `json:"shopping_cart_id"`
	ShippingMethodID int64  `json:"shipping_method_id"`
	OrderStatusID    int64  `json:"order_status_id"`
	OrderTotal       string `json:"order_total"`
}

// FinishedPurchaseTxResult is the result of the purchase transaction
type FinishedPurchaseTxResult struct {
	UpdatedProductItemID int64 `json:"product_item_id"`
	ShopOrderID          int64 `json:"shop_order_id"`
	ShopOrderItemID      int64 `json:"shop_order_item_id"`
}

/*
	FinishedPurchaseTx performs a product transfer from products DB to the user's shop_cart_item

once the payments is finished successfully it creates ShopOrderItem record,
substract from/update the product DB, adds the products to the users' shop_order_item DB,
and update products quantity within a single database transaction.
*/
func (store *SQLStore) FinishedPurchaseTx(ctx context.Context, arg FinishedPurchaseTxParams) (FinishedPurchaseTxResult, error) {
	var result FinishedPurchaseTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		shopCartItems, err := q.ListShoppingCartItemsByCartID(ctx, arg.ShoppingCartID)
		if err != nil {
			return err
		}

		trackNumber := util.GenerateTrackNumber()

		// argPM := GetPaymentMethodParams{
		// 	UserID:        arg.UserID,
		// 	PaymentTypeID: arg.PaymentTypeID,
		// }

		// paymentMethod, err := q.GetPaymentMethod(ctx, argPM)
		// if err != nil {
		// 	return err
		// }

		createdShopOrder, err := q.CreateShopOrder(ctx, CreateShopOrderParams{
			TrackNumber: trackNumber,
			UserID:      arg.UserID,
			// PaymentMethodID:   paymentMethod.ID,
			ShippingAddressID: arg.UserAddressID,
			OrderTotal:        arg.OrderTotal,
			ShippingMethodID:  arg.ShippingMethodID,
			OrderStatusID:     null.IntFrom(arg.OrderStatusID),
		})
		if err != nil {
			return err
		}

		for i := 0; i < len(shopCartItems); i++ {

			productItem, err := q.GetProductItemForUpdate(ctx, shopCartItems[i].ProductItemID)
			if err != nil {
				return err
			}

			if productItem.QtyInStock <= shopCartItems[i].Qty && productItem.QtyInStock > 0 {
				return errors.New("Not Enough Qty in Stock")
			}

			if productItem.QtyInStock <= 0 {
				return errors.New("Stock is Empty")
			}

			result.ShopOrderID = createdShopOrder.ID

			updatedProductItem, err := q.UpdateProductItem(ctx, UpdateProductItemParams{
				ID:         productItem.ID,
				ProductID:  productItem.ProductID,
				QtyInStock: null.IntFrom(int64(productItem.QtyInStock - shopCartItems[i].Qty)),
			})
			if err != nil {
				return err
			}
			result.UpdatedProductItemID = updatedProductItem.ID

			createdShopOrderItem, err := q.CreateShopOrderItem(ctx, CreateShopOrderItemParams{
				ProductItemID: shopCartItems[i].ProductItemID,
				OrderID:       createdShopOrder.ID,
				Quantity:      shopCartItems[i].Qty,
				Price:         productItem.Price,
			})
			if err != nil {
				return err
			}
			result.ShopOrderItemID = createdShopOrderItem.ID
		}
		_, err = q.DeleteShoppingCartItemAllByUser(ctx, DeleteShoppingCartItemAllByUserParams{
			UserID:         arg.UserID,
			ShoppingCartID: arg.ShoppingCartID,
		})
		if err != nil {
			return err
		}

		return nil
	})

	return result, err
}
