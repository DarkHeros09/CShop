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
type createVariationOptionParamsRequest struct {
	AdminID int64 `params:"adminId" validate:"required,min=1"`
}
type createVariationOptionJsonRequest struct {
	VariationID int64  `json:"variation_id" validate:"required,min=1"`
	Value       string `json:"value" validate:"required,alphanum"`
}

func (server *Server) createVariationOption(ctx *fiber.Ctx) error {
	params := &createVariationOptionParamsRequest{}
	req := &createVariationOptionJsonRequest{}

	if err := parseAndValidate(ctx, Input{params: params, req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.AdminPayload)
	if authPayload.AdminID != params.AdminID || authPayload.TypeID != 1 || !authPayload.Active {
		err := errors.New("account unauthorized")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.CreateVariationOptionParams{
		VariationID: null.IntFrom(req.VariationID),
		Value:       req.Value,
	}

	variationOption, err := server.store.CreateVariationOption(ctx.Context(), arg)
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

	ctx.Status(fiber.StatusOK).JSON(variationOption)
	return nil
}

//////////////* Get API //////////////

type getVariationOptionParamsRequest struct {
	ID int64 `params:"id" validate:"required,min=1"`
}

func (server *Server) getVariationOption(ctx *fiber.Ctx) error {
	params := &getVariationOptionParamsRequest{}

	if err := parseAndValidate(ctx, Input{params: params}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	variationOption, err := server.store.GetVariationOption(ctx.Context(), params.ID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(variationOption)
	return nil
}

//////////////* List API //////////////

type listVariationOptionsQueryRequest struct {
	PageID   int32 `query:"page_id" validate:"required,min=1"`
	PageSize int32 `query:"page_size" validate:"required,min=5,max=10"`
}

func (server *Server) listVariationOptions(ctx *fiber.Ctx) error {
	query := &listVariationOptionsQueryRequest{}

	if err := parseAndValidate(ctx, Input{query: query}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	arg := db.ListVariationOptionsParams{
		Limit:  query.PageSize,
		Offset: (query.PageID - 1) * query.PageSize,
	}
	variationOptions, err := server.store.ListVariationOptions(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(variationOptions)
	return nil
}

//////////////* Update API //////////////

type updateVariationOptionParamsRequest struct {
	AdminID int64 `params:"adminId" validate:"required,min=1"`
	ID      int64 `params:"id" validate:"required,min=1"`
}

type updateVariationOptionJsonRequest struct {
	VariationID int64  `json:"variation_id" validate:"omitempty,required,min=1"`
	Value       string `json:"value" validate:"omitempty,required"`
}

func (server *Server) updateVariationOption(ctx *fiber.Ctx) error {
	params := &updateVariationOptionParamsRequest{}
	req := &updateVariationOptionJsonRequest{}

	if err := parseAndValidate(ctx, Input{params: params, req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.AdminPayload)
	if authPayload.AdminID != params.AdminID || authPayload.TypeID != 1 || !authPayload.Active {
		err := errors.New("account unauthorized")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.UpdateVariationOptionParams{
		ID:          params.ID,
		Value:       null.StringFromPtr(&req.Value),
		VariationID: null.IntFromPtr(&req.VariationID),
	}

	variationOption, err := server.store.UpdateVariationOption(ctx.Context(), arg)
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
	ctx.Status(fiber.StatusOK).JSON(variationOption)
	return nil
}

//////////////* Delete API //////////////

type deleteVariationOptionParamsRequest struct {
	AdminID int64 `params:"adminId" validate:"required,min=1"`
	ID      int64 `params:"id" validate:"required,min=1"`
}

func (server *Server) deleteVariationOption(ctx *fiber.Ctx) error {
	params := &deleteVariationOptionParamsRequest{}

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

	err := server.store.DeleteVariationOption(ctx.Context(), params.ID)
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
