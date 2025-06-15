package api

import (
	"errors"

	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/token"
	"github.com/gofiber/fiber/v2"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
)

// //////////////* Create API //////////////

// type createShoppingCartItemParamsRequest struct {
// 	UserID         int64 `params:"id" validate:"required,min=1"`
// 	ShoppingCartID int64 `params:"cartId" validate:"required,min=1"`
// }

// type data struct {
// 	ProductItemID int64 `json:"product_item_id" validate:"required,min=1"`
// 	QTY           int32 `json:"qty" validate:"required,min=1"`
// }

// type createShoppingCartItemsRequest struct {
// 	ShopCartItem []data `json:"data" validate:"required,dive,required"`
// }

// func (server *Server) createShoppingCartItem(ctx *fiber.Ctx) error {
// 	params := &createShoppingCartItemParamsRequest{}
// 	req := &createShoppingCartItemsRequest{}
// 	var arg []db.CreateShoppingCartItemParams
// 	var shoppingCartItems []db.ShoppingCartItem
// 	var err1 error

// 	if err := server.parseAndValidate(ctx, Input{params: params, req: req}); err != nil {
// 		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
// 		return nil
// 	}

// 	authPayload := ctx.Locals(authorizationUserPayloadKey).(*token.UserPayload)
// 	if authPayload.UserID != params.UserID {
// 		err := errors.New("account deosn't belong to the authenticated user")
// 		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
// 		return nil
// 	}

// 	for i := 0; i < len(req.ShopCartItem); i++ {
// 		arg = append(arg, db.CreateShoppingCartItemParams{
// 			ShoppingCartID: params.ShoppingCartID,
// 			ProductItemID:  req.ShopCartItem[i].ProductItemID,
// 			Qty:            req.ShopCartItem[i].QTY,
// 		})
// 	}

// 	result := server.store.CreateShoppingCartItem(ctx.Context(), arg)

// 	result.Query(func(i int, sci []db.ShoppingCartItem, err error) {
// 		err1 = err
// 		shoppingCartItems = append(shoppingCartItems, sci...)
// 	})

// 	if err1 != nil {
// 		if pqErr, ok := err1.(*pgconn.PgError); ok {
// 			switch pqErr.Message {
// 			case "foreign_key_violation", "unique_violation":
// 				ctx.Status(fiber.StatusForbidden).JSON(errorResponse(err1))
// 				return nil
// 			}
// 		}
// 		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err1))
// 		return nil
// 	}
// 	ctx.Status(fiber.StatusOK).JSON(shoppingCartItems)
// 	return nil
// }

//////////////* Create API //////////////

type createShoppingCartItemParamsRequest struct {
	UserID         int64 `params:"id" validate:"required,min=1"`
	ShoppingCartID int64 `params:"cartId" validate:"required,min=1"`
}
type createShoppingCartItemRequest struct {
	ProductItemID int64 `json:"product_item_id" validate:"required,min=1"`
	SizeID        int64 `json:"size_id" validate:"required,min=1"`
	QTY           int32 `json:"qty" validate:"required,min=1"`
}

func (server *Server) createShoppingCartItem(ctx *fiber.Ctx) error {
	params := &createShoppingCartItemParamsRequest{}
	req := &createShoppingCartItemRequest{}

	if err := server.parseAndValidate(ctx, Input{params: params, req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationUserPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.CreateShoppingCartItemParams{
		ShoppingCartID: params.ShoppingCartID,
		ProductItemID:  req.ProductItemID,
		Qty:            req.QTY,
		SizeID:         req.SizeID,
	}

	shoppingCartItem, err := server.store.CreateShoppingCartItem(ctx.Context(), arg)
	if err != nil {
		if pqErr, ok := err.(*pgconn.PgError); ok {
			switch pqErr.Message {
			case "foreign_key_violation", "unique_violation":
				ctx.Status(fiber.StatusForbidden).JSON(errorResponse(err))
				return nil
			}
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(shoppingCartItem)
	return nil
}

// ////////////* Get API //////////////
type getShoppingCartItemParamsRequest struct {
	UserID         int64 `params:"id" validate:"required,min=1"`
	ShoppingCartID int64 `params:"cartId" validate:"required,min=1"`
}

func (server *Server) getShoppingCartItem(ctx *fiber.Ctx) error {
	params := &getShoppingCartItemParamsRequest{}

	if err := server.parseAndValidate(ctx, Input{params: params}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationUserPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.GetShoppingCartItemByUserIDCartIDParams{
		UserID: params.UserID,
		ID:     params.ShoppingCartID,
	}

	shoppingCartItem, err := server.store.GetShoppingCartItemByUserIDCartID(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(shoppingCartItem)
	return nil
}

//////////////* List API //////////////

type listShoppingCartItemsParamsRequest struct {
	UserID int64 `params:"id" validate:"required,min=1"`
}

type listShoppingCartItemsResponse struct {
	// ShopCartItems []db.ListShoppingCartItemsByUserIDRow `json:"shop_cart_items"`
	ID             null.Int  `json:"id"`
	ShoppingCartID null.Int  `json:"shopping_cart_id"`
	CreatedAt      null.Time `json:"created_at"`
	UpdatedAt      null.Time `json:"updated_at"`
	// ProductItems  []db.ListProductItemsByIDsRow         `json:"product_items"`
	ProductItemID             null.Int    `json:"product_item_id"`
	Name                      null.String `json:"name"`
	SizeID                    null.Int    `json:"size_id"`
	SizeValue                 null.String `json:"size_value"`
	SizeQty                   null.Int32  `json:"size_qty"`
	Color                     null.String `json:"color"`
	Qty                       null.Int    `json:"qty"`
	ProductID                 int64       `json:"product_id"`
	ProductImage              string      `json:"product_image"`
	Price                     string      `json:"price"`
	Active                    bool        `json:"active"`
	CategoryPromoID           null.Int    `json:"category_promo_id"`
	CategoryPromoName         null.String `json:"category_promo_name"`
	CategoryPromoDescription  null.String `json:"category_promo_description"`
	CategoryPromoDiscountRate null.Int    `json:"category_promo_discount_rate"`
	CategoryPromoActive       bool        `json:"category_promo_active"`
	CategoryPromoStartDate    null.Time   `json:"category_promo_start_date"`
	CategoryPromoEndDate      null.Time   `json:"category_promo_end_date"`
	BrandPromoID              null.Int    `json:"brand_promo_id"`
	BrandPromoName            null.String `json:"brand_promo_name"`
	BrandPromoDescription     null.String `json:"brand_promo_description"`
	BrandPromoDiscountRate    null.Int    `json:"brand_promo_discount_rate"`
	BrandPromoActive          bool        `json:"brand_promo_active"`
	BrandPromoStartDate       null.Time   `json:"brand_promo_start_date"`
	BrandPromoEndDate         null.Time   `json:"brand_promo_end_date"`
	ProductPromoID            null.Int    `json:"product_promo_id"`
	ProductPromoName          null.String `json:"product_promo_name"`
	ProductPromoDescription   null.String `json:"product_promo_description"`
	ProductPromoDiscountRate  null.Int    `json:"product_promo_discount_rate"`
	ProductPromoActive        bool        `json:"product_promo_active"`
	ProductPromoStartDate     null.Time   `json:"product_promo_start_date"`
	ProductPromoEndDate       null.Time   `json:"product_promo_end_date"`
}

func newlistShoppingCartItemsResponse(shopCartItems []db.ListShoppingCartItemsByUserIDRow, productItems []db.ListProductItemsByIDsRow, productsSizes []db.ProductSize) []listShoppingCartItemsResponse {
	rsp := make([]listShoppingCartItemsResponse, len(productItems))
	for i := 0; i < len(productItems); i++ {
		for j := 0; j < len(shopCartItems); j++ {
			for k := 0; k < len(productsSizes); k++ {
				if productItems[i].ID == shopCartItems[j].ProductItemID.Int64 && productItems[i].ID == productsSizes[k].ProductItemID {
					rsp[i] = listShoppingCartItemsResponse{
						ID:             shopCartItems[j].ID,
						ShoppingCartID: shopCartItems[j].ShoppingCartID,
						CreatedAt:      shopCartItems[j].CreatedAt,
						UpdatedAt:      shopCartItems[j].UpdatedAt,
						ProductItemID:  shopCartItems[j].ProductItemID,
						Name:           productItems[i].Name,
						Qty:            shopCartItems[j].Qty,
						ProductID:      productItems[i].ProductID,
						// ProductImage:   productItems[i].ProductImage,
						ProductImage:              productItems[i].ProductImage1.String,
						SizeID:                    null.IntFromPtr(&productsSizes[k].ID),
						SizeValue:                 null.StringFromPtr(&productsSizes[k].SizeValue),
						SizeQty:                   null.Int32FromPtr(&productsSizes[k].Qty),
						Color:                     productItems[i].ColorValue,
						Price:                     productItems[i].Price,
						Active:                    productItems[i].Active,
						CategoryPromoID:           productItems[i].CategoryPromoID,
						CategoryPromoName:         productItems[i].CategoryPromoName,
						CategoryPromoDescription:  productItems[i].CategoryPromoDescription,
						CategoryPromoDiscountRate: productItems[i].CategoryPromoDiscountRate,
						CategoryPromoActive:       productItems[i].CategoryPromoActive,
						CategoryPromoStartDate:    productItems[i].CategoryPromoStartDate,
						CategoryPromoEndDate:      productItems[i].CategoryPromoEndDate,
						BrandPromoID:              productItems[i].BrandPromoID,
						BrandPromoName:            productItems[i].BrandPromoName,
						BrandPromoDescription:     productItems[i].BrandPromoDescription,
						BrandPromoDiscountRate:    productItems[i].BrandPromoDiscountRate,
						BrandPromoActive:          productItems[i].BrandPromoActive,
						BrandPromoStartDate:       productItems[i].BrandPromoStartDate,
						BrandPromoEndDate:         productItems[i].BrandPromoEndDate,
						ProductPromoID:            productItems[i].ProductPromoID,
						ProductPromoName:          productItems[i].ProductPromoName,
						ProductPromoDescription:   productItems[i].ProductPromoDescription,
						ProductPromoDiscountRate:  productItems[i].ProductPromoDiscountRate,
						ProductPromoActive:        productItems[i].ProductPromoActive,
						ProductPromoStartDate:     productItems[i].ProductPromoStartDate,
						ProductPromoEndDate:       productItems[i].ProductPromoEndDate,
					}
				}
			}
		}
	}

	return rsp
}

func (server *Server) listShoppingCartItems(ctx *fiber.Ctx) error {
	params := &listShoppingCartItemsParamsRequest{}

	if err := server.parseAndValidate(ctx, Input{params: params}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationUserPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}
	shoppingCartItems, err := server.store.ListShoppingCartItemsByUserID(ctx.Context(), authPayload.UserID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	productsItemsIds := make([]int64, len(shoppingCartItems))
	for i := 0; i < len(shoppingCartItems); i++ {
		productsItemsIds[i] = shoppingCartItems[i].ProductItemID.Int64
	}

	productItems, err := server.store.ListProductItemsByIDs(ctx.Context(), productsItemsIds)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	productsSizesIds := make([]int64, len(shoppingCartItems))
	for i := 0; i < len(shoppingCartItems); i++ {
		productsSizesIds[i] = shoppingCartItems[i].SizeID.Int64
	}

	productsSizes, err := server.store.ListProductSizesByIDs(ctx.Context(), productsSizesIds)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	// listShoppingCartItemsChan := make(chan []listShoppingCartItemsResponse, len(productItems))

	// go func() {
	// 	rsp := newlistShoppingCartItemsResponse(shoppingCartItems, productItems)
	// 	listShoppingCartItemsChan <- rsp
	// 	}()

	// 	rsp := <-listShoppingCartItemsChan
	rsp := newlistShoppingCartItemsResponse(shoppingCartItems, productItems, productsSizes)
	ctx.Status(fiber.StatusOK).JSON(rsp)
	return nil
}

// ////////////* UPDATE API //////////////
type updateShoppingCartItemParamsRequest struct {
	UserID             int64 `params:"id" validate:"required,min=1"`
	ShoppingCartID     int64 `params:"cartId" validate:"required,min=1"`
	ShoppingCartItemID int64 `params:"itemId" validate:"required,min=1"`
}

type updateShoppingCartItemJsonRequest struct {
	ProductItemID *int64 `json:"product_item_id" validate:"omitempty,required,min=1"`
	SizeID        *int64 `json:"size_id" validate:"omitempty,required,min=1"`
	QTY           *int64 `json:"qty" validate:"omitempty,required"`
}

func (server *Server) updateShoppingCartItem(ctx *fiber.Ctx) error {
	params := &updateShoppingCartItemParamsRequest{}
	req := &updateShoppingCartItemJsonRequest{}

	if err := server.parseAndValidate(ctx, Input{params: params, req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationUserPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.UpdateShoppingCartItemParams{
		ID:             params.ShoppingCartItemID,
		ShoppingCartID: params.ShoppingCartID,
		ProductItemID:  null.IntFromPtr(req.ProductItemID),
		SizeID:         null.IntFromPtr(req.SizeID),
		Qty:            null.IntFromPtr(req.QTY),
	}

	shoppingCart, err := server.store.UpdateShoppingCartItem(ctx.Context(), arg)
	if err != nil {
		if pqErr, ok := err.(*pgconn.PgError); ok {
			switch pqErr.Message {
			case "foreign_key_violation", "unique_violation":
				ctx.Status(fiber.StatusForbidden).JSON(errorResponse(err))
				return nil
			}
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(shoppingCart)
	return nil
}

// ////////////* Delete API //////////////
type deleteShoppingCartItemParamsRequest struct {
	UserID             int64 `params:"id" validate:"required,min=1"`
	ShoppingCartID     int64 `params:"cartId" validate:"required,min=1"`
	ShoppingCartItemID int64 `params:"itemId" validate:"required,min=1"`
}

func (server *Server) deleteShoppingCartItem(ctx *fiber.Ctx) error {
	params := &deleteShoppingCartItemParamsRequest{}

	if err := server.parseAndValidate(ctx, Input{params: params}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationUserPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.DeleteShoppingCartItemParams{
		UserID:             authPayload.UserID,
		ShoppingCartID:     params.ShoppingCartID,
		ShoppingCartItemID: params.ShoppingCartItemID,
	}

	err := server.store.DeleteShoppingCartItem(ctx.Context(), arg)
	if err != nil {
		if pqErr, ok := err.(*pgconn.PgError); ok {
			switch pqErr.Message {
			case "foreign_key_violation", "unique_violation":
				ctx.Status(fiber.StatusForbidden).JSON(errorResponse(err))
				return nil
			}
		} else if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
	return nil
}

// ////////////* Delete All API //////////////

type deleteShoppingCartItemAllParamsRequest struct {
	UserID         int64 `params:"id" validate:"required,min=1"`
	ShoppingCartID int64 `params:"cartId" validate:"required,min=1"`
}

func (server *Server) deleteShoppingCartItemAllByUser(ctx *fiber.Ctx) error {
	params := &deleteShoppingCartItemAllParamsRequest{}

	if err := server.parseAndValidate(ctx, Input{params: params}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationUserPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.DeleteShoppingCartItemAllByUserParams{
		UserID:         params.UserID,
		ShoppingCartID: params.ShoppingCartID,
	}

	_, err := server.store.DeleteShoppingCartItemAllByUser(ctx.Context(), arg)
	if err != nil {
		if pqErr, ok := err.(*pgconn.PgError); ok {
			switch pqErr.Message {
			case "foreign_key_violation", "unique_violation":
				ctx.Status(fiber.StatusForbidden).JSON(errorResponse(err))
				return nil
			}
		} else if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
	return nil
}

// ////////////* Finish Purshase API //////////////
type finishPurshaseParamsRequest struct {
	UserID         int64 `params:"id" validate:"required,min=1"`
	ShoppingCartID int64 `params:"cartId" validate:"required,min=1"`
}
type finishPurshaseJsonRequest struct {
	UserAddressID    int64  `json:"user_address_id" validate:"required,min=1"`
	PaymentTypeID    int64  `json:"payment_type_id" validate:"required,min=1"`
	ShippingMethodID int64  `json:"shipping_method_id" validate:"required,min=1"`
	OrderStatusID    int64  `json:"order_status_id" validate:"required,min=1"`
	OrderTotal       string `json:"order_total" validate:"required"`
}

func (server *Server) finishPurchase(ctx *fiber.Ctx) error {
	params := &finishPurshaseParamsRequest{}
	req := &finishPurshaseJsonRequest{}

	if err := server.parseAndValidate(ctx, Input{params: params, req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationUserPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.FinishedPurchaseTxParams{
		UserID:           authPayload.UserID,
		AddressID:        req.UserAddressID,
		PaymentTypeID:    req.PaymentTypeID,
		ShoppingCartID:   params.ShoppingCartID,
		ShippingMethodID: req.ShippingMethodID,
		OrderStatusID:    req.OrderStatusID,
		OrderTotal:       req.OrderTotal,
	}

	finishedPurchase, err := server.store.FinishedPurchaseTx(ctx.Context(), arg)
	if err != nil {
		if pqErr, ok := err.(*pgconn.PgError); ok {
			switch pqErr.Message {
			case "foreign_key_violation", "unique_violation":
				ctx.Status(fiber.StatusForbidden).JSON(errorResponse(err))
				return nil
			}
		} else if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(finishedPurchase)
	return nil
}
