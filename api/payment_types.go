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

// //////////////* Admin Create Payment Type API ////////////

type createPaymentTypeParamsRequest struct {
	AdminID int64 `params:"adminId" validate:"required,min=1"`
}

type createPaymentTypeJsonRequest struct {
	Value    string `json:"value" validate:"required,alphanumunicode"`
	IsActive bool   `json:"is_active" validate:"boolean"`
}

func (server *Server) createPaymentType(ctx *fiber.Ctx) error {
	params := &createPaymentTypeParamsRequest{}
	req := &createPaymentTypeJsonRequest{}

	if err := parseAndValidate(ctx, Input{params: params, req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationAdminPayloadKey).(*token.AdminPayload)
	if authPayload.AdminID != params.AdminID || authPayload.TypeID != 1 || !authPayload.Active {
		err := errors.New("account unauthorized")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.AdminCreatePaymentTypeParams{
		AdminID:  authPayload.AdminID,
		Value:    req.Value,
		IsActive: req.IsActive,
	}

	product, err := server.store.AdminCreatePaymentType(ctx.Context(), arg)
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

	ctx.Status(fiber.StatusOK).JSON(product)
	return nil
}

// //////////////* List API //////////////

type adminListPaymentTypesParamsRequest struct {
	AdminID int64 `params:"adminId" validate:"required,min=1"`
}

func (server *Server) adminListPaymentTypes(ctx *fiber.Ctx) error {
	params := &adminListPaymentTypesParamsRequest{}

	if err := parseAndValidate(ctx, Input{params: params}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationAdminPayloadKey).(*token.AdminPayload)
	if authPayload.AdminID != params.AdminID || authPayload.TypeID != 1 || !authPayload.Active {
		err := errors.New("account unauthorized")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	paymentTypes, err := server.store.AdminListPaymentTypes(ctx.Context(), authPayload.AdminID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}
	ctx.Status(fiber.StatusOK).JSON(paymentTypes)
	return nil
}

//////////////* Update API //////////////

type updatePaymentTypeParamsRequest struct {
	AdminID       int64 `params:"adminId" validate:"required,min=1"`
	PaymentTypeID int64 `params:"paymentTypeId" validate:"required,min=1"`
}

type updatePaymentTypeJsonRequest struct {
	Value    *string `json:"value" validate:"omitempty,required,alphanumunicode"`
	IsActive *bool   `json:"is_active" validate:"omitempty,required,boolean"`
}

func (server *Server) updatePaymentType(ctx *fiber.Ctx) error {
	params := &updatePaymentTypeParamsRequest{}
	req := &updatePaymentTypeJsonRequest{}

	if err := parseAndValidate(ctx, Input{params: params, req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationAdminPayloadKey).(*token.AdminPayload)
	if authPayload.AdminID != params.AdminID || authPayload.TypeID != 1 || !authPayload.Active {
		err := errors.New("account unauthorized")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.AdminUpdatePaymentTypeParams{
		AdminID:  authPayload.AdminID,
		ID:       params.PaymentTypeID,
		Value:    null.StringFromPtr(req.Value),
		IsActive: null.BoolFromPtr(req.IsActive),
	}

	product, err := server.store.AdminUpdatePaymentType(ctx.Context(), arg)
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
	ctx.Status(fiber.StatusOK).JSON(product)
	return nil
}

//////////////* Delete API //////////////

type deletePaymentTypeParamsRequest struct {
	AdminID       int64 `params:"adminId" validate:"required,min=1"`
	PaymentTypeID int64 `params:"paymentTypeId" validate:"required,min=1"`
}

func (server *Server) deletePaymentType(ctx *fiber.Ctx) error {
	params := &deletePaymentTypeParamsRequest{}

	if err := parseAndValidate(ctx, Input{params: params}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationAdminPayloadKey).(*token.AdminPayload)
	if authPayload.AdminID != params.AdminID || authPayload.TypeID != 1 || !authPayload.Active {
		err := errors.New("account unauthorized")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.AdminDeletePaymentTypeParams{
		AdminID: authPayload.AdminID,
		ID:      params.PaymentTypeID,
	}

	err := server.store.AdminDeletePaymentType(ctx.Context(), arg)
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

// //////////////* List API //////////////

type listPaymentTypesParamsRequest struct {
	UserID int64 `params:"id" validate:"required,min=1"`
}

func (server *Server) listPaymentTypes(ctx *fiber.Ctx) error {
	params := &listPaymentTypesParamsRequest{}

	if err := parseAndValidate(ctx, Input{params: params}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationUserPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	paymentTypes, err := server.store.ListPaymentTypes(ctx.Context())
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}
	ctx.Status(fiber.StatusOK).JSON(paymentTypes)
	return nil
}
