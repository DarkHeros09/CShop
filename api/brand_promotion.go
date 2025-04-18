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

type createBrandPromotionParamsRequest struct {
	AdminID int64 `params:"adminId" validate:"required,min=1"`
}

type createBrandPromotionJsonRequest struct {
	BrandID             int64  `json:"brand_id" validate:"required,min=1"`
	PromotionID         int64  `json:"promotion_id" validate:"required,min=1"`
	BrandPromotionImage string `json:"brand_promotion_image" validate:"required,http_url"`
	Active              bool   `json:"active" validate:"boolean"`
}

func (server *Server) createBrandPromotion(ctx *fiber.Ctx) error {
	params := &createBrandPromotionParamsRequest{}
	req := &createBrandPromotionJsonRequest{}

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

	arg := db.AdminCreateBrandPromotionParams{
		AdminID:             authPayload.AdminID,
		BrandID:             req.BrandID,
		PromotionID:         req.PromotionID,
		BrandPromotionImage: null.StringFromPtr(&req.BrandPromotionImage),
		Active:              req.Active,
	}

	brandPromotion, err := server.store.AdminCreateBrandPromotion(ctx.Context(), arg)
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

	ctx.Status(fiber.StatusOK).JSON(brandPromotion)
	return nil
}

//////////////* Get API //////////////

type getBrandPromotionParamsRequest struct {
	PromotionID int64 `params:"promotionId" validate:"required,min=1"`
	BrandID     int64 `params:"brandId" validate:"required,min=1"`
}

func (server *Server) getBrandPromotion(ctx *fiber.Ctx) error {
	params := &getBrandPromotionParamsRequest{}

	if err := server.parseAndValidate(ctx, Input{params: params}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	arg := db.GetBrandPromotionParams{
		BrandID:     params.BrandID,
		PromotionID: params.PromotionID,
	}

	brandPromotion, err := server.store.GetBrandPromotion(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}
	ctx.Status(fiber.StatusOK).JSON(brandPromotion)
	return nil
}

//////////////* List API //////////////

type listBrandPromotionsQueryRequest struct {
	PageID   int32 `query:"page_id" validate:"required,min=1"`
	PageSize int32 `query:"page_size" validate:"required,min=5,max=10"`
}

func (server *Server) listBrandPromotions(ctx *fiber.Ctx) error {
	query := &listBrandPromotionsQueryRequest{}

	if err := server.parseAndValidate(ctx, Input{query: query}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	arg := db.ListBrandPromotionsParams{
		Limit:  query.PageSize,
		Offset: (query.PageID - 1) * query.PageSize,
	}
	brandPromotions, err := server.store.ListBrandPromotions(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(brandPromotions)
	return nil

}

//////////////* List API with Images //////////////

func (server *Server) listBrandPromotionsWithImages(ctx *fiber.Ctx) error {

	brandPromotions, err := server.store.ListBrandPromotionsWithImages(ctx.Context())
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(brandPromotions)
	return nil

}

//////////////* Admin List API with Images //////////////

type adminListBrandPromotionParamsRequest struct {
	AdminID int64 `params:"adminId" validate:"required,min=1"`
}

func (server *Server) listBrandPromotionsForAdmins(ctx *fiber.Ctx) error {
	params := &adminListBrandPromotionParamsRequest{}

	if err := server.parseAndValidate(ctx, Input{params: params}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationAdminPayloadKey).(*token.AdminPayload)
	if authPayload.AdminID != params.AdminID || authPayload.TypeID != 1 || !authPayload.Active {
		err := errors.New("account unauthorized")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	brandPromotions, err := server.store.AdminListBrandPromotions(ctx.Context(), authPayload.AdminID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(brandPromotions)
	return nil

}

//////////////* Update API //////////////

type updateBrandPromotionParamsRequest struct {
	AdminID     int64 `params:"adminId" validate:"required,min=1"`
	BrandID     int64 `params:"brandId" validate:"required,min=1"`
	PromotionID int64 `params:"promotionId" validate:"required,min=1"`
}

type updateBrandPromotionJsonRequest struct {
	BrandPromotionImage *string `json:"brand_promotion_image" validate:"omitempty,required,url"`
	Active              *bool   `json:"active" validate:"omitempty,required,boolean"`
}

func (server *Server) updateBrandPromotion(ctx *fiber.Ctx) error {
	params := &updateBrandPromotionParamsRequest{}
	req := &updateBrandPromotionJsonRequest{}

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

	arg := db.AdminUpdateBrandPromotionParams{
		AdminID:             authPayload.AdminID,
		BrandID:             params.BrandID,
		PromotionID:         params.PromotionID,
		BrandPromotionImage: null.StringFromPtr(req.BrandPromotionImage),
		Active:              null.BoolFromPtr(req.Active),
	}

	brandPromotion, err := server.store.AdminUpdateBrandPromotion(ctx.Context(), arg)
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
	ctx.Status(fiber.StatusOK).JSON(brandPromotion)
	return nil
}

//////////////* Delete API //////////////

type deleteBrandPromotionParamsRequest struct {
	AdminID     int64 `params:"adminId" validate:"required,min=1"`
	PromotionID int64 `params:"promotionId" validate:"required,min=1"`
	BrandID     int64 `params:"brandId" validate:"required,min=1"`
}

func (server *Server) deleteBrandPromotion(ctx *fiber.Ctx) error {
	params := &deleteBrandPromotionParamsRequest{}

	if err := server.parseAndValidate(ctx, Input{params: params}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationAdminPayloadKey).(*token.AdminPayload)
	if authPayload.AdminID != params.AdminID || authPayload.TypeID != 1 || !authPayload.Active {
		err := errors.New("account unauthorized")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.DeleteBrandPromotionParams{
		BrandID:     params.BrandID,
		PromotionID: params.PromotionID,
	}

	err := server.store.DeleteBrandPromotion(ctx.Context(), arg)
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
