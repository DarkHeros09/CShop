package db

import (
	"context"
	"fmt"

	"github.com/guregu/null"
	"github.com/jackc/pgx/v4"
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
	db *pgx.Conn
}

// NewStore creates a new Store
func NewStore(db *pgx.Conn) Store {
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
	UserAddress    UserAddress    `json:"user_address"`
	PaymentMethod  PaymentMethod  `json:"payment_method"`
	ShoppingCart   ShoppingCart   `json:"shopping_cart"`
	ShippingMethod ShippingMethod `json:"shipping_method"`
	OrderStatus    OrderStatus    `json:"order_status"`
	OrderTotal     string         `json:"order_total"`
}

// FinishedPurchaseTxResult is the result of the purchase transaction
type FinishedPurchaseTxResult struct {
	UpdatedProductItem ProductItem   `json:"product_item"`
	OrderStatus        OrderStatus   `json:"order_status"`
	ShopOrder          ShopOrder     `json:"shop_order"`
	ShopOrderItem      ShopOrderItem `json:"shop_order_item"`
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

		shopCartItems, err := q.ListShoppingCartItemsByCartID(ctx, arg.ShoppingCart.ID)
		if err != nil {
			return err
		}

		for _, shopCartItem := range shopCartItems {

			productItem, err := q.GetProductItemForUpdate(ctx, shopCartItem.ProductItemID)
			if err != nil {
				return err
			}

			if productItem.QtyInStock <= shopCartItem.Qty && productItem.QtyInStock > 0 {
				return errors.New("Not Enough Qty in Stock")
			}

			if productItem.QtyInStock <= 0 {
				return errors.New("Stock is Empty")
			}

			result.ShopOrder, err = q.CreateShopOrder(ctx, CreateShopOrderParams{
				UserID:            arg.UserAddress.UserID,
				PaymentMethodID:   arg.PaymentMethod.ID,
				ShippingAddressID: arg.UserAddress.AddressID,
				OrderTotal:        arg.OrderTotal,
				ShippingMethodID:  arg.ShippingMethod.ID,
				OrderStatusID:     arg.OrderStatus.ID,
			})
			if err != nil {
				return err
			}

			result.UpdatedProductItem, err = q.UpdateProductItem(ctx, UpdateProductItemParams{
				ProductID:    null.IntFromPtr(&productItem.ProductID),
				ProductSku:   null.Int{},
				QtyInStock:   null.IntFrom(int64(productItem.QtyInStock - shopCartItem.Qty)),
				ProductImage: null.String{},
				Price:        null.String{},
				Active:       null.Bool{},
				ID:           productItem.ID,
			})
			if err != nil {
				return err
			}
			result.ShopOrderItem, err = q.CreateShopOrderItem(ctx, CreateShopOrderItemParams{
				ProductItemID: shopCartItem.ProductItemID,
				OrderID:       result.ShopOrder.ID,
				Quantity:      shopCartItem.Qty,
				Price:         productItem.Price,
			})
			if err != nil {
				return err
			}
		}
		_, err = q.DeleteShoppingCartItemAllByUser(ctx, arg.UserAddress.UserID)
		if err != nil {
			return err
		}

		return nil
	})

	return result, err
}
