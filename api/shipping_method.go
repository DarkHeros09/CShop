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

//////////////* Create API //////////////

type createShippingMethodParamsRequest struct {
	UserID int64 `params:"id" validate:"required,min=1"`
}
type createShippingMethodJsonRequest struct {
	Name  string `json:"name" validate:"required"`
	Price string `json:"price" validate:"required"`
}

func (server *Server) createShippingMethod(ctx *fiber.Ctx) error {
	params := &createShippingMethodParamsRequest{}
	req := &createShippingMethodJsonRequest{}

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

	arg := db.CreateShippingMethodParams{
		Name:  req.Name,
		Price: req.Price,
	}

	shippingMethod, err := server.store.CreateShippingMethod(ctx.Context(), arg)
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

	ctx.Status(fiber.StatusOK).JSON(shippingMethod)
	return nil
}

// //////////////* Get API //////////////

type getShippingMethodParamsRequest struct {
	UserID           int64 `params:"id" validate:"required,min=1"`
	ShippingMethodID int64 `params:"methodId" validate:"required,min=1"`
}

func (server *Server) getShippingMethod(ctx *fiber.Ctx) error {
	params := &getShippingMethodParamsRequest{}

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

	arg := db.GetShippingMethodByUserIDParams{
		ID:     params.ShippingMethodID,
		UserID: params.UserID,
	}

	shippingMethod, err := server.store.GetShippingMethodByUserID(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(shippingMethod)
	return nil
}

// //////////////* List API //////////////

type listShippingMethodsParamsRequest struct {
	UserID int64 `params:"id" validate:"required,min=1"`
}

// type listShippingMethodsQueryRequest struct {
// 	PageID   int32 `query:"page_id" validate:"required,min=1"`
// 	PageSize int32 `query:"page_size" validate:"required,min=5,max=10"`
// }

func (server *Server) listShippingMethods(ctx *fiber.Ctx) error {
	params := &listShippingMethodsParamsRequest{}
	// query := &listShippingMethodsQueryRequest{}

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
	// arg := db.ListShippingMethodsByUserIDParams{
	// 	UserID: authPayload.UserID,
	// Limit:  query.PageSize,
	// Offset: (query.PageID - 1) * query.PageSize,
	// }
	shippingMethods, err := server.store.ListShippingMethods(ctx.Context())
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}
	ctx.Status(fiber.StatusOK).JSON(shippingMethods)
	return nil
}

// //////////////* UPDATE API ///////////////
type updateShippingMethodParamsRequest struct {
	UserID           int64 `params:"id" validate:"required,min=1"`
	ShippingMethodID int64 `params:"methodId" validate:"required,min=1"`
}

type updateShippingMethodJsonRequest struct {
	Name  string `json:"name" validate:"omitempty,required"`
	Price string `json:"price" validate:"omitempty,required"`
}

func (server *Server) updateShippingMethod(ctx *fiber.Ctx) error {
	params := &updateShippingMethodParamsRequest{}
	req := &updateShippingMethodJsonRequest{}

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

	arg := db.UpdateShippingMethodParams{
		Name:  null.StringFromPtr(&req.Name),
		Price: null.StringFromPtr(&req.Price),
		ID:    params.ShippingMethodID,
	}

	shippingMethod, err := server.store.UpdateShippingMethod(ctx.Context(), arg)
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

	ctx.Status(fiber.StatusOK).JSON(shippingMethod)
	return nil
}

// ////////////* Delete API //////////////

type deleteShippingMethodParamsRequest struct {
	AdminID          int64 `params:"adminId" validate:"required,min=1"`
	ShippingMethodID int64 `params:"methodId" validate:"required,min=1"`
}

func (server *Server) deleteShippingMethod(ctx *fiber.Ctx) error {
	params := &deleteShippingMethodParamsRequest{}

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

	err := server.store.DeleteShippingMethod(ctx.Context(), params.ShippingMethodID)
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
