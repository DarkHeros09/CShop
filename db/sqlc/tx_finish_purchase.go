package db

import (
	"context"
	"time"

	"github.com/cshop/v3/util"
	"github.com/guregu/null/v5"
	"github.com/pkg/errors"
)

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
			TrackNumber:       trackNumber,
			UserID:            arg.UserID,
			PaymentTypeID:     arg.PaymentTypeID,
			ShippingAddressID: arg.UserAddressID,
			OrderTotal:        arg.OrderTotal,
			ShippingMethodID:  arg.ShippingMethodID,
			OrderStatusID:     null.IntFrom(arg.OrderStatusID),
		})
		if err != nil {
			return err
		}

		shippingMethod, err := q.GetShippingMethod(ctx, arg.ShippingMethodID)
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

			productItemAfterUpdate, err := q.GetProductItemWithPromotions(ctx, shopCartItems[i].ProductItemID)
			if err != nil {
				return err
			}

			bestDiscount := discount(productItemAfterUpdate)

			// discountValue := decimal.NewFromInt(bestDiscount)

			// discountDecimal := decimal.NewFromInt(1).Sub(discountValue.Div(decimal.NewFromInt(100)))

			// price, err := decimal.NewFromString(productItem.Price)
			// if err != nil {
			// 	return err
			// }

			// bestPrice := price.Mul(decimal.NewFromInt(int64(shopCartItems[i].Qty))).Mul(discountDecimal)

			createdShopOrderItem, err := q.CreateShopOrderItem(ctx, CreateShopOrderItemParams{
				ProductItemID:       shopCartItems[i].ProductItemID,
				OrderID:             createdShopOrder.ID,
				Quantity:            shopCartItems[i].Qty,
				Price:               productItem.Price,
				Discount:            int32(bestDiscount),
				ShippingMethodPrice: shippingMethod.Price,
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

func discount(productItem GetProductItemWithPromotionsRow) int64 {
	var promos []Promotion

	if productItem.ProductPromoDiscountRate.Valid {
		promos = append(promos, Promotion{
			ID:           productItem.ProductPromoID.Int64,
			Name:         productItem.ProductPromoName.String,
			Description:  productItem.ProductPromoDescription.String,
			DiscountRate: productItem.ProductPromoDiscountRate.Int64,
			Active:       productItem.ProductPromoActive,
			StartDate:    productItem.ProductPromoStartDate.Time,
			EndDate:      productItem.ProductPromoEndDate.Time,
		})
	}

	if productItem.CategoryPromoDiscountRate.Valid {
		promos = append(promos, Promotion{
			ID:           productItem.CategoryPromoID.Int64,
			Name:         productItem.CategoryPromoName.String,
			Description:  productItem.CategoryPromoDescription.String,
			DiscountRate: productItem.CategoryPromoDiscountRate.Int64,
			Active:       productItem.CategoryPromoActive,
			StartDate:    productItem.CategoryPromoStartDate.Time,
			EndDate:      productItem.CategoryPromoEndDate.Time,
		})
	}

	if productItem.BrandPromoDiscountRate.Valid {
		promos = append(promos, Promotion{
			ID:           productItem.BrandPromoID.Int64,
			Name:         productItem.BrandPromoName.String,
			Description:  productItem.BrandPromoDescription.String,
			DiscountRate: productItem.BrandPromoDiscountRate.Int64,
			Active:       productItem.BrandPromoActive,
			StartDate:    productItem.BrandPromoStartDate.Time,
			EndDate:      productItem.BrandPromoEndDate.Time,
		})
	}

	var validPromos []Promotion
	now := time.Now()

	for _, promo := range promos {
		if promo.Active &&
			now.After(promo.StartDate) && now.Before(promo.EndDate) {
			validPromos = append(validPromos, promo)
		}
	}

	if len(validPromos) == 0 {
		return 0
	}

	bestPromo := validPromos[0]
	for _, promo := range validPromos {
		if promo.DiscountRate > bestPromo.DiscountRate {
			bestPromo = promo
		}
	}

	if bestPromo.DiscountRate > 0 {
		return bestPromo.DiscountRate
	}

	return 0
}
