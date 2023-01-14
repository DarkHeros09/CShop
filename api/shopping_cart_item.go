package api

import (
	"errors"

	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/token"
	"github.com/cshop/v3/util"
	"github.com/gofiber/fiber/v2"
	"github.com/guregu/null"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

//////////////* Create API //////////////

type createShoppingCartItemParamsRequest struct {
	UserID         int64 `params:"id" validate:"required,min=1"`
	ShoppingCartID int64 `params:"cartId" validate:"required,min=1"`
}
type createShoppingCartItemRequest struct {
	ProductItemID int64 `json:"product_item_id" validate:"required,min=1"`
	QTY           int32 `json:"qty" validate:"required,min=1"`
}

func (server *Server) createShoppingCartItem(ctx *fiber.Ctx) error {
	var params createShoppingCartItemParamsRequest
	var req createShoppingCartItemRequest

	if err := ctx.ParamsParser(&params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := util.ValidateStruct(params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := ctx.BodyParser(&req); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := util.ValidateStruct(req); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.CreateShoppingCartItemParams{
		ShoppingCartID: params.ShoppingCartID,
		ProductItemID:  req.ProductItemID,
		Qty:            req.QTY,
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
	var params getShoppingCartItemParamsRequest

	if err := ctx.ParamsParser(&params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := util.ValidateStruct(params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.GetShoppingCartItemByUserIDCartIDParams{
		UserID:         params.UserID,
		ShoppingCartID: params.ShoppingCartID,
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

func (server *Server) listShoppingCartItems(ctx *fiber.Ctx) error {
	var params listShoppingCartItemsParamsRequest

	if err := ctx.ParamsParser(&params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := util.ValidateStruct(params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.UserPayload)
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
	ctx.Status(fiber.StatusOK).JSON(shoppingCartItems)
	return nil
}

// ////////////* UPDATE API //////////////
type updateShoppingCartItemParamsRequest struct {
	UserID             int64 `params:"id" validate:"required,min=1"`
	ShoppingCartID     int64 `params:"cartId" validate:"required,min=1"`
	ShoppingCartItemID int64 `params:"itemId" validate:"required,min=1"`
}

type updateShoppingCartItemJsonRequest struct {
	ProductItemID int64 `json:"product_item_id" validate:"omitempty,required"`
	QTY           int64 `json:"qty" validate:"omitempty,required"`
}

func (server *Server) updateShoppingCartItem(ctx *fiber.Ctx) error {
	var params updateShoppingCartItemParamsRequest
	var req updateShoppingCartItemJsonRequest

	if err := ctx.ParamsParser(&params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := util.ValidateStruct(params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := ctx.BodyParser(&req); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := util.ValidateStruct(req); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.UpdateShoppingCartItemParams{
		ID:             params.ShoppingCartItemID,
		ShoppingCartID: params.ShoppingCartID,
		ProductItemID:  null.IntFromPtr(&req.ProductItemID),
		Qty:            null.IntFromPtr(&req.QTY),
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
	var params deleteShoppingCartItemParamsRequest

	if err := ctx.ParamsParser(&params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := util.ValidateStruct(params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.UserPayload)
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
	var params deleteShoppingCartItemAllParamsRequest

	if err := ctx.ParamsParser(&params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := util.ValidateStruct(params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.UserPayload)
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
	PaymentMethodID  int64  `json:"payment_method_id" validate:"required,min=1"`
	ShippingMethodID int64  `json:"shipping_method_id" validate:"required,min=1"`
	OrderStatusID    int64  `json:"order_status_id" validate:"required,min=1"`
	OrderTotal       string `json:"order_total" validate:"required"`
}

func (server *Server) finishPurchase(ctx *fiber.Ctx) error {
	var params finishPurshaseParamsRequest
	var req finishPurshaseJsonRequest

	if err := ctx.ParamsParser(&params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := util.ValidateStruct(params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := ctx.BodyParser(&req); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := util.ValidateStruct(req); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.FinishedPurchaseTxParams{
		UserID:           authPayload.UserID,
		UserAddressID:    req.UserAddressID,
		PaymentMethodID:  req.PaymentMethodID,
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
