package api

import (
	"errors"

	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/token"
	"github.com/gofiber/fiber/v2"
	"github.com/guregu/null"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

//////////////* Create API //////////////

type createCategoryPromotionParamsRequest struct {
	AdminID     int64 `params:"adminId" validate:"required,min=1"`
	CategoryID  int64 `params:"categoryId" validate:"required,min=1"`
	PromotionID int64 `params:"promotionId" validate:"required,min=1"`
}

type createCategoryPromotionJsonRequest struct {
	Active bool `json:"active" validate:"boolean"`
}

func (server *Server) createCategoryPromotion(ctx *fiber.Ctx) error {
	params := &createCategoryPromotionParamsRequest{}
	req := &createCategoryPromotionJsonRequest{}

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

	arg := db.CreateCategoryPromotionParams{
		CategoryID:  params.CategoryID,
		PromotionID: params.PromotionID,
		Active:      req.Active,
	}

	categoryPromotion, err := server.store.CreateCategoryPromotion(ctx.Context(), arg)
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

	ctx.Status(fiber.StatusOK).JSON(categoryPromotion)
	return nil
}

//////////////* Get API //////////////

type getCategoryPromotionParamsRequest struct {
	PromotionID int64 `params:"promotionId" validate:"required,min=1"`
	CategoryID  int64 `params:"categoryId" validate:"required,min=1"`
}

func (server *Server) getCategoryPromotion(ctx *fiber.Ctx) error {
	params := &getCategoryPromotionParamsRequest{}

	if err := parseAndValidate(ctx, Input{params: params}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	arg := db.GetCategoryPromotionParams{
		CategoryID:  params.CategoryID,
		PromotionID: params.PromotionID,
	}

	categoryPromotion, err := server.store.GetCategoryPromotion(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}
	ctx.Status(fiber.StatusOK).JSON(categoryPromotion)
	return nil
}

//////////////* List API //////////////

type listCategoryPromotionsQueryRequest struct {
	PageID   int32 `query:"page_id" validate:"required,min=1"`
	PageSize int32 `query:"page_size" validate:"required,min=5,max=10"`
}

func (server *Server) listCategoryPromotions(ctx *fiber.Ctx) error {
	query := &listCategoryPromotionsQueryRequest{}

	if err := parseAndValidate(ctx, Input{query: query}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	arg := db.ListCategoryPromotionsParams{
		Limit:  query.PageSize,
		Offset: (query.PageID - 1) * query.PageSize,
	}
	CategoryPromotions, err := server.store.ListCategoryPromotions(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(CategoryPromotions)
	return nil

}

//////////////* Update API //////////////

type updateCategoryPromotionParamsRequest struct {
	AdminID     int64 `params:"adminId" validate:"required,min=1"`
	CategoryID  int64 `params:"categoryId" validate:"required,min=1"`
	PromotionID int64 `params:"promotionId" validate:"required,min=1"`
}

type updateCategoryPromotionJsonRequest struct {
	Active bool `json:"active" validate:"omitempty,required,boolean"`
}

func (server *Server) updateCategoryPromotion(ctx *fiber.Ctx) error {
	params := &updateCategoryPromotionParamsRequest{}
	req := &updateCategoryPromotionJsonRequest{}

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

	arg := db.UpdateCategoryPromotionParams{
		CategoryID:  params.CategoryID,
		PromotionID: params.PromotionID,
		Active:      null.BoolFromPtr(&req.Active),
	}

	categoryPromotion, err := server.store.UpdateCategoryPromotion(ctx.Context(), arg)
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
	ctx.Status(fiber.StatusOK).JSON(categoryPromotion)
	return nil
}

//////////////* Delete API //////////////

type deleteCategoryPromotionParamsRequest struct {
	AdminID     int64 `params:"adminId" validate:"required,min=1"`
	PromotionID int64 `params:"promotionId" validate:"required,min=1"`
	CategoryID  int64 `params:"categoryId" validate:"required,min=1"`
}

func (server *Server) deleteCategoryPromotion(ctx *fiber.Ctx) error {
	params := &deleteCategoryPromotionParamsRequest{}

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

	arg := db.DeleteCategoryPromotionParams{
		CategoryID:  params.CategoryID,
		PromotionID: params.PromotionID,
	}

	err := server.store.DeleteCategoryPromotion(ctx.Context(), arg)
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
