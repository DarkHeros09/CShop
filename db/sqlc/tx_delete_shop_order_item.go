package db

import (
	"context"

	"github.com/guregu/null/v5"
	"github.com/shopspring/decimal"
)

// DeleteShopOrderItemTx contains the input parameters of the purchase transaction
type DeleteShopOrderItemTxParams struct {
	AdminID         int64 `json:"admin_id"`
	ShopOrderItemID int64 `json:"shop_order_item_id"`
}

/*
DeleteShopOrderItemTx performs a shop order item delete from DB, and update the new total price in shop order table
*/
func (store *SQLStore) DeleteShopOrderItemTx(ctx context.Context, arg DeleteShopOrderItemTxParams) error {

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		arg := DeleteShopOrderItemParams{
			AdminID: arg.AdminID,
			ID:      arg.ShopOrderItemID,
		}

		deletedShopOrderItem, err := q.DeleteShopOrderItem(ctx, arg)
		if err != nil {
			return err
		}

		shopOrder, err := q.GetShopOrder(ctx, deletedShopOrderItem.OrderID)
		if err != nil {
			return err
		}

		deletedShopOrderItemPrice, err := decimal.NewFromString(deletedShopOrderItem.Price)
		if err != nil {
			return err
		}

		deletedShopOrderItemDiscount := decimal.NewFromInt(int64(deletedShopOrderItem.Discount))

		shopOrderTotalPrice, err := decimal.NewFromString(shopOrder.OrderTotal)
		if err != nil {
			return err
		}

		discountDecimal := decimal.NewFromInt(1).Sub(deletedShopOrderItemDiscount.Div(decimal.NewFromInt(100)))

		newTotalPrice := shopOrderTotalPrice.Sub(deletedShopOrderItemPrice.Mul(decimal.NewFromInt(int64(deletedShopOrderItem.Quantity))).Mul(discountDecimal))

		_, err = q.UpdateShopOrder(ctx, UpdateShopOrderParams{
			AdminID:    arg.AdminID,
			ID:         deletedShopOrderItem.OrderID,
			OrderTotal: null.StringFrom(newTotalPrice.StringFixed(2)),
		})
		if err != nil {
			return err
		}

		return nil
	})

	return err
}
