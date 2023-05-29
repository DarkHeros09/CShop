package api

import (
	"errors"

	"github.com/cshop/v3/token"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v4"
)

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

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.UserPayload)
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
