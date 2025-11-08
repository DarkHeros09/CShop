package api

import (
	"errors"

	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/token"
	"github.com/cshop/v3/util"
	"github.com/gofiber/fiber/v2"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
)

// ////////////* Create API //////////////
type createProductColorParamsRequest struct {
	AdminID int64 `params:"adminId" validate:"required,min=1"`
}

type createProductColorJsonRequest struct {
	ColorValue string `json:"color_value" validate:"required,alphanumunicode_space"`
}

func (server *Server) createProductColor(ctx *fiber.Ctx) error {
	params := &createProductColorParamsRequest{}
	req := &createProductColorJsonRequest{}

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

	arg := db.AdminCreateProductColorParams{
		AdminID:    authPayload.AdminID,
		ColorValue: req.ColorValue,
	}

	productColor, err := server.store.AdminCreateProductColor(ctx.Context(), arg)
	if err != nil {
		if pqErr, ok := err.(*pgconn.PgError); ok {
			switch pqErr.Message {
			case util.ForeignKeyViolation, util.UniqueViolation:
				ctx.Status(fiber.StatusForbidden).JSON(errorResponse(err))
				return nil
			}
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(productColor)
	return nil
}

// ////////////* List API //////////////

func (server *Server) listProductColors(ctx *fiber.Ctx) error {

	productColors, err := server.store.ListProductColors(ctx.Context())
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	if productColors == nil {
		ctx.Status(fiber.StatusNotFound).JSON(pgx.ErrNoRows)
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(productColors)
	return nil
}

//////////////* Update API //////////////

type updateProductColorParamsRequest struct {
	AdminID int64 `params:"adminId" validate:"required,min=1"`
	ID      int64 `params:"id" validate:"required,min=1"`
}

type updateProductColorJsonRequest struct {
	Color *string `json:"color" validate:"omitempty,required,alphaunicode"`
}

func (server *Server) updateProductColor(ctx *fiber.Ctx) error {
	params := &updateProductColorParamsRequest{}
	req := &updateProductColorJsonRequest{}

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

	arg := db.AdminUpdateProductColorParams{
		AdminID:    authPayload.AdminID,
		ID:         params.ID,
		ColorValue: null.StringFromPtr(req.Color),
	}

	color, err := server.store.AdminUpdateProductColor(ctx.Context(), arg)
	if err != nil {
		if pqErr, ok := err.(*pgconn.PgError); ok {
			switch pqErr.Message {
			case util.ForeignKeyViolation, util.UniqueViolation:
				ctx.Status(fiber.StatusForbidden).JSON(errorResponse(err))
				return nil
			}
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}
	ctx.Status(fiber.StatusOK).JSON(color)
	return nil
}
