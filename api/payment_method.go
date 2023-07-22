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

type createPaymentMethodParamsRequest struct {
	UserID int64 `params:"id" validate:"required,min=1"`
}
type createPaymentMethodJsonRequest struct {
	Provider      string `json:"provider" validate:"required"`
	PaymentTypeID int64  `json:"payment_type_id" validate:"required,min=1"`
}

func (server *Server) createPaymentMethod(ctx *fiber.Ctx) error {
	params := &createPaymentMethodParamsRequest{}
	req := &createPaymentMethodJsonRequest{}

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

	arg := db.CreatePaymentMethodParams{
		UserID:        authPayload.UserID,
		PaymentTypeID: req.PaymentTypeID,
		Provider:      req.Provider,
	}

	paymentMethod, err := server.store.CreatePaymentMethod(ctx.Context(), arg)
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

	ctx.Status(fiber.StatusOK).JSON(paymentMethod)
	return nil
}

// //////////////* Get API //////////////

type getPaymentMethodParamsRequest struct {
	// ID     int64 `params:"paymentId" validate:"required,min=1"`
	UserID int64 `params:"id" validate:"required,min=1"`
}

type getPaymentMethodJsonRequest struct {
	PaymentTypeID int64 `json:"payment_type_id" validate:"required,min=1"`
}

func (server *Server) getPaymentMethod(ctx *fiber.Ctx) error {
	params := &getPaymentMethodParamsRequest{}
	req := &getPaymentMethodJsonRequest{}

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

	arg := db.GetPaymentMethodParams{
		// ID:            params.ID,
		UserID:        params.UserID,
		PaymentTypeID: req.PaymentTypeID,
	}

	paymentMethod, err := server.store.GetPaymentMethod(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(paymentMethod)
	return nil
}

// //////////////* List API //////////////

type listPaymentMethodsParamsRequest struct {
	UserID int64 `params:"id" validate:"required,min=1"`
}

type listPaymentMethodsRequest struct {
	PageID   int32 `query:"page_id" validate:"required,min=1"`
	PageSize int32 `query:"page_size" validate:"required,min=5,max=10"`
}

func (server *Server) listPaymentMethods(ctx *fiber.Ctx) error {
	params := &listPaymentMethodsParamsRequest{}
	query := &listPaymentMethodsRequest{}

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
	arg := db.ListPaymentMethodsParams{
		UserID: authPayload.UserID,
		Limit:  query.PageSize,
		Offset: (query.PageID - 1) * query.PageSize,
	}
	paymentMethods, err := server.store.ListPaymentMethods(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}
	ctx.Status(fiber.StatusOK).JSON(paymentMethods)
	return nil
}

// //////////////* UPDATE API ///////////////
type updatePaymentMethodParamsRequest struct {
	ID     int64 `params:"paymentId" validate:"required,min=1"`
	UserID int64 `params:"id" validate:"required,min=1"`
}

type updatePaymentMethodJsonRequest struct {
	PaymentTypeID int64  `json:"payment_type_id" validate:"required,min=1"`
	Provider      string `json:"provider" validate:"required"`
}

func (server *Server) updatePaymentMethod(ctx *fiber.Ctx) error {
	params := &updatePaymentMethodParamsRequest{}
	req := &updatePaymentMethodJsonRequest{}

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

	arg := db.UpdatePaymentMethodParams{
		ID:            params.ID,
		UserID:        null.IntFromPtr(&authPayload.UserID),
		PaymentTypeID: null.IntFromPtr(&req.PaymentTypeID),
		Provider:      null.StringFromPtr(&req.Provider),
	}

	paymentMethod, err := server.store.UpdatePaymentMethod(ctx.Context(), arg)
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

	ctx.Status(fiber.StatusOK).JSON(paymentMethod)
	return nil
}

// ////////////* Delete API //////////////
type deletePaymentMethodParamsRequest struct {
	ID     int64 `params:"paymentId" validate:"required,min=1"`
	UserID int64 `params:"id" validate:"required,min=1"`
}

func (server *Server) deletePaymentMethod(ctx *fiber.Ctx) error {
	params := &deletePaymentMethodParamsRequest{}

	if err := parseAndValidate(ctx, Input{params: params}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.DeletePaymentMethodParams{
		ID:     params.ID,
		UserID: authPayload.UserID,
	}

	_, err := server.store.DeletePaymentMethod(ctx.Context(), arg)
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
