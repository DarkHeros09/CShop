package api

import (
	"errors"

	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/token"
	"github.com/gofiber/fiber/v2"
	"github.com/guregu/null"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
)

// ////////////* Create API //////////////
type createOrderStatusParamsRequest struct {
	UserID int64 `params:"id" validate:"required,min=1"`
}

type createOrderStatusJsonRequest struct {
	Status string `json:"status" validate:"required"`
}

func (server *Server) createOrderStatus(ctx *fiber.Ctx) error {
	params := &createOrderStatusParamsRequest{}
	req := &createOrderStatusJsonRequest{}

	if err := parseAndValidate(ctx, Input{params: params, req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.UserPayload)
	if params.UserID != authPayload.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	orderStatus, err := server.store.CreateOrderStatus(ctx.Context(), req.Status)
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

	ctx.Status(fiber.StatusOK).JSON(orderStatus)
	return nil
}

// //////////////* Get API //////////////

type getOrderStatusParamsRequest struct {
	UserID   int64 `params:"id" validate:"required,min=1"`
	StatusID int64 `params:"statusId" validate:"required,min=1"`
}

func (server *Server) getOrderStatus(ctx *fiber.Ctx) error {
	params := &getOrderStatusParamsRequest{}

	if err := parseAndValidate(ctx, Input{params: params}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.UserPayload)
	if params.UserID != authPayload.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	// arg := db.GetOrderStatusByUserIDParams{
	// 	ID:     params.StatusID,
	// 	UserID: params.UserID,
	// }

	orderStatus, err := server.store.GetOrderStatus(ctx.Context(), params.StatusID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(orderStatus)
	return nil
}

// //////////////* List API //////////////
type listOrderStatusParamsRequest struct {
	UserID int64 `params:"id" validate:"required,min=1"`
}
type listOrderStatusQueryRequest struct {
	PageID   int32 `query:"page_id" validate:"required,min=1"`
	PageSize int32 `query:"page_size" validate:"required,min=5,max=10"`
}

func (server *Server) listOrderStatuses(ctx *fiber.Ctx) error {
	params := &listOrderStatusParamsRequest{}
	query := &listOrderStatusQueryRequest{}

	if err := parseAndValidate(ctx, Input{params: params, query: query}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.UserPayload)
	if params.UserID != authPayload.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.ListOrderStatusesByUserIDParams{
		UserID: authPayload.UserID,
		Limit:  query.PageSize,
		Offset: (query.PageID - 1) * query.PageSize,
	}
	orderStatuses, err := server.store.ListOrderStatusesByUserID(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}
	ctx.Status(fiber.StatusOK).JSON(orderStatuses)
	return nil
}

// //////////////* UPDATE API ///////////////
type updateOrderStatusParamsRequest struct {
	UserID   int64 `params:"id" validate:"required,min=1"`
	StatusID int64 `params:"statusId" validate:"required,min=1"`
}

type updateOrderStatusJsonRequest struct {
	Status string `json:"status" validate:"omitempty,required"`
}

func (server *Server) updateOrderStatus(ctx *fiber.Ctx) error {
	params := &updateOrderStatusParamsRequest{}
	req := &updateOrderStatusJsonRequest{}

	if err := parseAndValidate(ctx, Input{params: params, req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.UpdateOrderStatusParams{
		Status: null.StringFromPtr(&req.Status),
		ID:     params.StatusID,
	}

	orderStatus, err := server.store.UpdateOrderStatus(ctx.Context(), arg)
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

	ctx.Status(fiber.StatusOK).JSON(orderStatus)
	return nil
}

// ////////////* Delete API //////////////

type deleteOrderStatusParamsRequest struct {
	StatusID int64 `params:"statusId" validate:"required,min=1"`
	AdminID  int64 `params:"adminId" validate:"required,min=1"`
}

func (server *Server) deleteOrderStatus(ctx *fiber.Ctx) error {
	params := &deleteOrderStatusParamsRequest{}

	if err := parseAndValidate(ctx, Input{params: params}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.AdminPayload)
	if authPayload.AdminID != params.AdminID || authPayload.TypeID != 1 || !authPayload.Active {
		err := errors.New("account unauthorized")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	err := server.store.DeleteOrderStatus(ctx.Context(), params.StatusID)
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
