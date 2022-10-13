package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Store provides all functions to execute db queries and transactions
type Store interface {
	Querier
	FinishedPurchaseTx(ctx context.Context, arg FinishedPurchaseTxParams) (FinishedPurchaseTxResult, error)
}

// Store provides all functions to execute db queries and transactions
type SQLStore struct {
	*Queries
	db *sql.DB
}

// NewStore creates a new Store
func NewStore(db *sql.DB) Store {
	return &SQLStore{
		db:      db,
		Queries: New(db),
	}
}

// execTx executes a function within a database transaction
// method starts with lower case to not be exported so external packages can't call it directly
// we will provide an exported function for each specific transaction
func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction error: %v, rollback error: %v", err, rbErr)
		}
		return err
	}
	return tx.Commit()

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

		shopCartItems, err := q.ListShoppingCartItemsByCartID(ctx, arg.ShoppingCart.ID)
		if err != nil {
			return err
		}

		fmt.Println("user: ", arg.UserAddress.UserID)
		for _, shopCartItem := range shopCartItems {

			productItem, err := q.GetProductItemForUpdate(ctx, shopCartItem.ProductItemID)
			if err != nil {
				return err
			}

			result.UpdatedProductItem, err = q.UpdateProductItem(ctx, UpdateProductItemParams{
				ProductID: sql.NullInt64{
					Int64: productItem.ProductID,
					Valid: true,
				},
				ProductSku: sql.NullInt64{},
				QtyInStock: sql.NullInt32{
					Int32: productItem.QtyInStock - shopCartItem.Qty,
					Valid: true,
				},
				ProductImage: sql.NullString{},
				Price:        sql.NullString{},
				Active:       sql.NullBool{},
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

		return nil
	})

	return result, err
}
