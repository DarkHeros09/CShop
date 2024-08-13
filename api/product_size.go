package api

import (
	"errors"

	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/token"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
)

// ////////////* Create API //////////////
type createProductSizeParamsRequest struct {
	AdminID int64 `params:"adminId" validate:"required,min=1"`
}

type createProductSizeJsonRequest struct {
	SizeValue string `json:"size_value" validate:"required,alphanum"`
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
