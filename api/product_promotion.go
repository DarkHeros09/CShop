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

//////////////* Create API //////////////

type createProductPromotionParamsRequest struct {
	AdminID     int64 `params:"adminId" validate:"required,min=1"`
	PromotionID int64 `params:"promotionId" validate:"required,min=1"`
	ProductID   int64 `params:"productId" validate:"required,min=1"`
}

type createProductPromotionJsonRequest struct {
	Active bool `json:"active" validate:"boolean"`
}

func (server *Server) createProductPromotion(ctx *fiber.Ctx) error {
	params := &createProductPromotionParamsRequest{}
	req := &createProductPromotionJsonRequest{}

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

	arg := db.CreateProductPromotionParams{
		ProductID:   params.ProductID,
		PromotionID: params.PromotionID,
		Active:      req.Active,
	}

	productPromotion, err := server.store.CreateProductPromotion(ctx.Context(), arg)
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

	ctx.Status(fiber.StatusOK).JSON(productPromotion)
	return nil
}

//////////////* Get API //////////////

type getProductPromotionParamsRequest struct {
	PromotionID int64 `params:"promotionId" validate:"required,min=1"`
	ProductID   int64 `params:"productId" validate:"required,min=1"`
}

func (server *Server) getProductPromotion(ctx *fiber.Ctx) error {
	params := &getProductPromotionParamsRequest{}

	if err := parseAndValidate(ctx, Input{params: params}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	arg := db.GetProductPromotionParams{
		ProductID:   params.ProductID,
		PromotionID: params.PromotionID,
	}

	productPromotion, err := server.store.GetProductPromotion(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}
	ctx.Status(fiber.StatusOK).JSON(productPromotion)
	return nil
}

//////////////* List API //////////////

type listProductPromotionsQueryRequest struct {
	PageID   int32 `query:"page_id" validate:"required,min=1"`
	PageSize int32 `query:"page_size" validate:"required,min=5,max=10"`
}

func (server *Server) listProductPromotions(ctx *fiber.Ctx) error {
	query := &listProductPromotionsQueryRequest{}

	if err := parseAndValidate(ctx, Input{query: query}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	arg := db.ListProductPromotionsParams{
		Limit:  query.PageSize,
		Offset: (query.PageID - 1) * query.PageSize,
	}
	productPromotions, err := server.store.ListProductPromotions(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return err
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(productPromotions)
	return nil

}

//////////////* List API with Images //////////////

func (server *Server) listProductPromotionsWithImages(ctx *fiber.Ctx) error {

	productPromotions, err := server.store.ListProductPromotionsWithImages(ctx.Context())
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(productPromotions)
	return nil

}

//////////////* Update API //////////////

type updateProductPromotionParamsRequest struct {
	AdminID     int64 `params:"adminId" validate:"required,min=1"`
	ProductID   int64 `params:"productId" validate:"required,min=1"`
	PromotionID int64 `params:"promotionId" validate:"required,min=1"`
}

type updateProductPromotionJsonRequest struct {
	Active bool `json:"active" validate:"omitempty,required,boolean"`
}

func (server *Server) updateProductPromotion(ctx *fiber.Ctx) error {
	params := &updateProductPromotionParamsRequest{}
	req := &updateProductPromotionJsonRequest{}

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

	arg := db.UpdateProductPromotionParams{
		ProductID:   params.ProductID,
		PromotionID: params.PromotionID,
		Active:      null.BoolFromPtr(&req.Active),
	}

	productPromotion, err := server.store.UpdateProductPromotion(ctx.Context(), arg)
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
	ctx.Status(fiber.StatusOK).JSON(productPromotion)
	return nil
}

//////////////* Delete API //////////////

type deleteProductPromotionParamsRequest struct {
	AdminID     int64 `params:"adminId" validate:"required,min=1"`
	PromotionID int64 `params:"promotionId" validate:"required,min=1"`
	ProductID   int64 `params:"productId" validate:"required,min=1"`
}

func (server *Server) deleteProductPromotion(ctx *fiber.Ctx) error {
	params := &deleteProductPromotionParamsRequest{}

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

	arg := db.DeleteProductPromotionParams{
		ProductID:   params.ProductID,
		PromotionID: params.PromotionID,
	}

	err := server.store.DeleteProductPromotion(ctx.Context(), arg)
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
