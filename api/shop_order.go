package api

import (
	"errors"
	"fmt"
	"math"

	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/token"
	"github.com/gofiber/fiber/v2"
	"github.com/guregu/null"
	"github.com/jackc/pgx/v5"
)

//////////////* List API //////////////

type listShopOrdersParamsRequest struct {
	UserID int64 `params:"id" validate:"required,min=1"`
}
type listShopOrdersQueryRequest struct {
	PageID   int32 `query:"page_id" validate:"required,min=1"`
	PageSize int32 `query:"page_size" validate:"required,min=5,max=10"`
}

func (server *Server) listShopOrders(ctx *fiber.Ctx) error {
	params := &listShopOrdersParamsRequest{}
	query := &listShopOrdersQueryRequest{}

	if err := parseAndValidate(ctx, Input{params: params, query: query}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.ListShopOrdersByUserIDParams{
		UserID: authPayload.UserID,
		Limit:  query.PageSize,
		Offset: (query.PageID - 1) * query.PageSize,
	}
	shopOrders, err := server.store.ListShopOrdersByUserID(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}
	ctx.Status(fiber.StatusOK).JSON(shopOrders)
	return nil
}

//////////////* Pagination List API //////////////

type listShopOrdersV2ParamsRequest struct {
	UserID int64 `params:"id" validate:"required,min=1"`
}

type listShopOrdersV2QueryRequest struct {
	Limit       int32       `query:"limit" validate:"required,min=5,max=10"`
	OrderStatus null.String `query:"order_status" validate:"omitempty,required"`
}

func (server *Server) listShopOrdersV2(ctx *fiber.Ctx) error {
	params := &listShopOrdersV2ParamsRequest{}
	query := &listShopOrdersV2QueryRequest{}
	var maxPage int64

	if err := parseAndValidate(ctx, Input{params: params, query: query}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.ListShopOrdersByUserIDV2Params{
		UserID:      authPayload.UserID,
		Limit:       query.Limit,
		OrderStatus: query.OrderStatus,
	}
	shopOrders, err := server.store.ListShopOrdersByUserIDV2(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	if len(shopOrders) != 0 {
		maxPage = int64(math.Ceil(float64(shopOrders[0].TotalCount) / float64(query.Limit)))
	} else {
		maxPage = 0
	}

	ctx.Set("Max-Page", fmt.Sprint(maxPage))
	ctx.Status(fiber.StatusOK).JSON(shopOrders)
	return nil
}

type listShopOrdersNextPageParamsRequest struct {
	UserID int64 `params:"id" validate:"required,min=1"`
}

type listShopOrdersNextPageQueryRequest struct {
	Cursor      int64       `query:"cursor" validate:"required,min=1"`
	Limit       int32       `query:"limit" validate:"required,min=5,max=10"`
	OrderStatus null.String `query:"order_status" validate:"omitempty,required"`
}

func (server *Server) listShopOrdersNextPage(ctx *fiber.Ctx) error {
	params := &listShopOrdersNextPageParamsRequest{}
	query := &listShopOrdersNextPageQueryRequest{}
	var maxPage int64

	if err := parseAndValidate(ctx, Input{params: params, query: query}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.ListShopOrdersByUserIDNextPageParams{
		UserID:      authPayload.UserID,
		ShopOrderID: query.Cursor,
		Limit:       query.Limit,
		OrderStatus: query.OrderStatus,
	}
	shopOrders, err := server.store.ListShopOrdersByUserIDNextPage(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	if len(shopOrders) != 0 {
		maxPage = int64(math.Ceil(float64(shopOrders[0].TotalCount) / float64(query.Limit)))
	} else {
		maxPage = 0
	}

	ctx.Set("Max-Page", fmt.Sprint(maxPage))
	ctx.Status(fiber.StatusOK).JSON(shopOrders)
	return nil
}
