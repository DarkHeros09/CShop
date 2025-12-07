package db

import (
	"context"

	"github.com/guregu/null/v6"
	"github.com/quagmt/udecimal"
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

		deletedShopOrderItemPrice, err := udecimal.Parse(deletedShopOrderItem.Price)
		if err != nil {
			return err
		}

		deletedShopOrderItemDiscount, err := udecimal.NewFromInt64(int64(deletedShopOrderItem.Discount), 2)
		if err != nil {
			return err
		}

		shopOrderTotalPrice, err := udecimal.Parse(shopOrder.OrderTotal)
		if err != nil {
			return err
		}

		diviBy100, err := udecimal.NewFromInt64(100, 2)
		if err != nil {
			return err
		}

		one, err := udecimal.NewFromInt64(1, 2)
		if err != nil {
			return err
		}

		divResult, err := deletedShopOrderItemDiscount.Div(diviBy100)
		if err != nil {
			return err
		}

		discountDecimal := one.Sub(divResult)

		qunt, err := udecimal.NewFromInt64(int64(deletedShopOrderItem.Quantity), 2)
		if err != nil {
			return err
		}

		newTotalPrice := shopOrderTotalPrice.Sub(deletedShopOrderItemPrice.Mul(qunt).Mul(discountDecimal))

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
