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
type createVariationParamsRequest struct {
	AdminID int64 `params:"adminId" validate:"required,min=1"`
}
type createVariationJsonRequest struct {
	Name       string `json:"name" validate:"required,alphanum"`
	CategoryID int64  `json:"category_id" validate:"required,min=1"`
}

func (server *Server) createVariation(ctx *fiber.Ctx) error {
	params := &createVariationParamsRequest{}
	req := &createVariationJsonRequest{}

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

	arg := db.CreateVariationParams{
		CategoryID: req.CategoryID,
		Name:       req.Name,
	}

	variation, err := server.store.CreateVariation(ctx.Context(), arg)
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

	ctx.Status(fiber.StatusOK).JSON(variation)
	return nil
}

//////////////* Get API //////////////

type getVariationParamsRequest struct {
	VariationID int64 `params:"variationId" validate:"required,min=1"`
}

func (server *Server) getVariation(ctx *fiber.Ctx) error {
	params := &getVariationParamsRequest{}

	if err := parseAndValidate(ctx, Input{params: params}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	variation, err := server.store.GetVariation(ctx.Context(), params.VariationID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(variation)
	return nil
}

//////////////* List API //////////////

type listVariationsQueryRequest struct {
	PageID   int32 `query:"page_id" validate:"required,min=1"`
	PageSize int32 `query:"page_size" validate:"required,min=5,max=10"`
}

func (server *Server) listVariations(ctx *fiber.Ctx) error {
	query := &listVariationsQueryRequest{}

	if err := parseAndValidate(ctx, Input{query: query}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	arg := db.ListVariationsParams{
		Limit:  query.PageSize,
		Offset: (query.PageID - 1) * query.PageSize,
	}
	variations, err := server.store.ListVariations(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(variations)
	return nil

}

//////////////* Update API //////////////

type updateVariationParamsRequest struct {
	AdminID     int64 `params:"adminId" validate:"required,min=1"`
	VariationID int64 `params:"variationId" validate:"required,min=1"`
}

type updateVariationJsonRequest struct {
	Name       string `json:"name" validate:"omitempty,required"`
	CategoryID int64  `json:"category_id" validate:"omitempty,required,min=1"`
}

func (server *Server) updateVariation(ctx *fiber.Ctx) error {
	params := &updateVariationParamsRequest{}
	req := &updateVariationJsonRequest{}

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

	arg := db.UpdateVariationParams{
		ID:         params.VariationID,
		Name:       null.StringFromPtr(&req.Name),
		CategoryID: null.IntFromPtr(&req.CategoryID),
	}

	variation, err := server.store.UpdateVariation(ctx.Context(), arg)
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
	ctx.Status(fiber.StatusOK).JSON(variation)
	return nil
}

//////////////* Delete API //////////////

type deleteVariationParamsRequest struct {
	AdminID     int64 `params:"adminId" validate:"required,min=1"`
	VariationID int64 `params:"variationId" validate:"required,min=1"`
}

func (server *Server) deleteVariation(ctx *fiber.Ctx) error {
	params := &deleteVariationParamsRequest{}

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

	err := server.store.DeleteVariation(ctx.Context(), params.VariationID)
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
