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

// ////////////* Create API //////////////
type createProductSizeParamsRequest struct {
	AdminID int64 `params:"adminId" validate:"required,min=1"`
}

type createProductSizeJsonRequest struct {
	SizeValue string `json:"size_value" validate:"required,alphanumunicode"`
}

func (server *Server) createProductSize(ctx *fiber.Ctx) error {
	params := &createProductSizeParamsRequest{}
	req := &createProductSizeJsonRequest{}

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

	arg := db.AdminCreateProductSizeParams{
		AdminID:   authPayload.AdminID,
		SizeValue: req.SizeValue,
	}

	productSize, err := server.store.AdminCreateProductSize(ctx.Context(), arg)
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

	ctx.Status(fiber.StatusOK).JSON(productSize)
	return nil
}

// ////////////* List API //////////////

func (server *Server) listProductSizes(ctx *fiber.Ctx) error {

	productSizes, err := server.store.ListProductSizes(ctx.Context())
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(productSizes)
	return nil
}

//////////////* Update API //////////////

type updateProductSizeParamsRequest struct {
	AdminID int64 `params:"adminId" validate:"required,min=1"`
	ID      int64 `params:"id" validate:"required,min=1"`
}

type updateProductSizeJsonRequest struct {
	Size *string `json:"size" validate:"omitempty,required,alphaunicode"`
}

func (server *Server) updateProductSize(ctx *fiber.Ctx) error {
	params := &updateProductSizeParamsRequest{}
	req := &updateProductSizeJsonRequest{}

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

	arg := db.AdminUpdateProductSizeParams{
		AdminID:   authPayload.AdminID,
		ID:        params.ID,
		SizeValue: null.StringFromPtr(req.Size),
	}

	size, err := server.store.AdminUpdateProductSize(ctx.Context(), arg)
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
	ctx.Status(fiber.StatusOK).JSON(size)
	return nil
}
