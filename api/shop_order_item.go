package api

import (
	"errors"

	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/token"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
)

//////////////* Get API //////////////

type getShopOrderItemParamsRequest struct {
	UserID      int64 `params:"id" validate:"required,min=1"`
	ShopOrderID int64 `params:"orderId" validate:"required,min=1"`
}

func (server *Server) getShopOrderItems(ctx *fiber.Ctx) error {
	params := &getShopOrderItemParamsRequest{}

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

	arg := db.ListShopOrderItemsByUserIDOrderIDParams{
		UserID:  params.UserID,
		OrderID: params.ShopOrderID,
	}

	shopOrderItem, err := server.store.ListShopOrderItemsByUserIDOrderID(ctx.Context(), arg)
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
	params := &listShopOrderItemsParamsRequest{}
	query := &listShopOrderItemsQueryRequest{}

	if err := server.parseAndValidate(ctx, Input{params: params, query: query}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationUserPayloadKey).(*token.UserPayload)
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

//////////////* Admin Get API //////////////

type adminGetShopOrderItemParamsRequest struct {
	AdminID     int64 `params:"adminId" validate:"required,min=1"`
	ShopOrderID int64 `params:"orderId" validate:"required,min=1"`
}

type adminGetShopOrderItemJsonRequest struct {
	UserID int64 `json:"user_id" validate:"required,min=1"`
}

func (server *Server) getShopOrderItemsForAdmin(ctx *fiber.Ctx) error {
	params := &adminGetShopOrderItemParamsRequest{}
	req := &adminGetShopOrderItemJsonRequest{}

	if err := server.parseAndValidate(ctx, Input{params: params, req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationAdminPayloadKey).(*token.AdminPayload)
	if authPayload.AdminID != params.AdminID || authPayload.TypeID != 1 || !authPayload.Active {
		err := errors.New("account unauthorized")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.ListShopOrderItemsByUserIDOrderIDParams{
		UserID:  req.UserID,
		OrderID: params.ShopOrderID,
	}

	shopOrderItem, err := server.store.ListShopOrderItemsByUserIDOrderID(ctx.Context(), arg)
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

//////////////* Delete API //////////////

type deleteShopOrderItemParamsRequest struct {
	AdminID         int64 `params:"adminId" validate:"required,min=1"`
	ShopOrderItemID int64 `params:"id" validate:"required,min=1"`
}

func (server *Server) deleteShopOrderItem(ctx *fiber.Ctx) error {
	params := &deleteShopOrderItemParamsRequest{}

	if err := server.parseAndValidate(ctx, Input{params: params}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationAdminPayloadKey).(*token.AdminPayload)
	if authPayload.AdminID != params.AdminID || authPayload.TypeID != 1 || !authPayload.Active {
		err := errors.New("account unauthorized")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.DeleteShopOrderItemTxParams{
		AdminID:         authPayload.AdminID,
		ShopOrderItemID: params.ShopOrderItemID,
	}

	err := server.store.DeleteShopOrderItemTx(ctx.Context(), arg)
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
