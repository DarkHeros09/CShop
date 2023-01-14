package api

import (
	"errors"

	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/token"
	"github.com/cshop/v3/util"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v4"
)

//////////////* Get API //////////////

type getShopOrderItemParamsRequest struct {
	UserID      int64 `params:"id" validate:"required,min=1"`
	ShopOrderID int64 `params:"orderId" validate:"required,min=1"`
}

func (server *Server) getShopOrderItem(ctx *fiber.Ctx) error {
	var params getShopOrderItemParamsRequest

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

	arg := db.GetShopOrderItemByUserIDOrderIDParams{
		UserID:  params.UserID,
		OrderID: params.ShopOrderID,
	}

	shopOrderItem, err := server.store.GetShopOrderItemByUserIDOrderID(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(shopOrderItem)
	return nil
}

//////////////* List API //////////////

type listShopOrderItemsParamsRequest struct {
	UserID int64 `params:"id" validate:"required,min=1"`
}
type listShopOrderItemsQueryRequest struct {
	PageID   int32 `query:"page_id" validate:"required,min=1"`
	PageSize int32 `query:"page_size" validate:"required,min=5,max=10"`
}

func (server *Server) listShopOrderItems(ctx *fiber.Ctx) error {
	var params listShopOrderItemsParamsRequest
	var query listShopOrderItemsQueryRequest

	if err := ctx.ParamsParser(&params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := util.ValidateStruct(params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := ctx.QueryParser(&query); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}
	if err := util.ValidateStruct(query); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.ListShopOrderItemsByUserIDParams{
		UserID: authPayload.UserID,
		Limit:  query.PageSize,
		Offset: (query.PageID - 1) * query.PageSize,
	}
	shopOrderItems, err := server.store.ListShopOrderItemsByUserID(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}
	ctx.Status(fiber.StatusOK).JSON(shopOrderItems)
	return nil
}
