package api

import (
	"errors"
	"time"

	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/token"
	"github.com/cshop/v3/util"
	"github.com/gofiber/fiber/v2"
	"github.com/guregu/null"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

//////////////* Create API //////////////

type createPromotionParamsRequest struct {
	AdminID int64 `params:"admin_id" validate:"required,min=1"`
}
type createPromotionJsonRequest struct {
	Name         string    `json:"name" validate:"required,alphanum"`
	Description  string    `json:"description" validate:"required"`
	DiscountRate int64     `json:"discount_rate" validate:"required,min=1"`
	Active       bool      `json:"active" validate:"boolean"`
	StartDate    time.Time `json:"start_date" validate:"required"`
	EndDate      time.Time `json:"end_date" validate:"required"`
}

func (server *Server) createPromotion(ctx *fiber.Ctx) error {
	var params createPromotionParamsRequest
	var req createPromotionJsonRequest

	if err := ctx.ParamsParser(&params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := util.ValidateStruct(params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := ctx.BodyParser(&req); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := util.ValidateStruct(req); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.AdminPayload)
	if authPayload.AdminID != params.AdminID || authPayload.TypeID != 1 || !authPayload.Active {
		err := errors.New("account unauthorized")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.CreatePromotionParams{
		Name:         req.Name,
		Description:  req.Description,
		DiscountRate: req.DiscountRate,
		Active:       req.Active,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
	}

	promotion, err := server.store.CreatePromotion(ctx.Context(), arg)
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

	ctx.Status(fiber.StatusOK).JSON(promotion)
	return nil
}

//////////////* Get API //////////////

type getPromotionParamsRequest struct {
	ID int64 `params:"promotion_id" validate:"required,min=1"`
}

func (server *Server) getPromotion(ctx *fiber.Ctx) error {
	var params getPromotionParamsRequest

	if err := ctx.ParamsParser(&params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := util.ValidateStruct(params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	promotion, err := server.store.GetPromotion(ctx.Context(), params.ID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}
	ctx.Status(fiber.StatusOK).JSON(promotion)
	return nil
}

//////////////* List API //////////////

type listPromotionsQueryRequest struct {
	PageID   int32 `query:"page_id" validate:"required,min=1"`
	PageSize int32 `query:"page_size" validate:"required,min=5,max=10"`
}

func (server *Server) listPromotions(ctx *fiber.Ctx) error {
	var query listPromotionsQueryRequest

	if err := ctx.QueryParser(&query); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := util.ValidateStruct(query); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	arg := db.ListPromotionsParams{
		Limit:  query.PageSize,
		Offset: (query.PageID - 1) * query.PageSize,
	}
	promotions, err := server.store.ListPromotions(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(promotions)
	return nil

}

//////////////* Update API //////////////

type updatePromotionParamsRequest struct {
	AdminID     int64 `params:"admin_id" validate:"required,min=1"`
	PromotionID int64 `params:"promotion_id" validate:"required,min=1"`
}

type updatePromotionJsonRequest struct {
	Name         string    `json:"name" validate:"omitempty,required,alphanum"`
	Description  string    `json:"description" validate:"omitempty,required"`
	DiscountRate int64     `json:"discount_rate" validate:"omitempty,required,min=1"`
	Active       bool      `json:"active" validate:"omitempty,required,boolean"`
	StartDate    time.Time `json:"start_date" validate:"omitempty,required"`
	EndDate      time.Time `json:"end_date" validate:"omitempty,required"`
}

func (server *Server) updatePromotion(ctx *fiber.Ctx) error {
	var params updatePromotionParamsRequest
	var req updatePromotionJsonRequest

	if err := ctx.ParamsParser(&params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := util.ValidateStruct(params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := ctx.BodyParser(&req); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := util.ValidateStruct(req); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.AdminPayload)
	if authPayload.AdminID != params.AdminID || authPayload.TypeID != 1 || !authPayload.Active {
		err := errors.New("account unauthorized")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.UpdatePromotionParams{
		ID:           params.PromotionID,
		Name:         null.StringFromPtr(&req.Name),
		Description:  null.StringFromPtr(&req.Description),
		DiscountRate: null.IntFromPtr(&req.DiscountRate),
		Active:       null.BoolFromPtr(&req.Active),
		StartDate:    null.TimeFromPtr(&req.StartDate),
		EndDate:      null.TimeFromPtr(&req.EndDate),
	}

	promotion, err := server.store.UpdatePromotion(ctx.Context(), arg)
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
	ctx.Status(fiber.StatusOK).JSON(promotion)
	return nil
}

//////////////* Delete API //////////////

type deletePromotionParamsRequest struct {
	AdminID     int64 `params:"admin_id" validate:"required,min=1"`
	PromotionID int64 `params:"promotion_id" validate:"required,min=1"`
}

func (server *Server) deletePromotion(ctx *fiber.Ctx) error {
	var params deletePromotionParamsRequest

	if err := ctx.ParamsParser(&params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := util.ValidateStruct(params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.AdminPayload)
	if authPayload.AdminID != params.AdminID || authPayload.TypeID != 1 || !authPayload.Active {
		err := errors.New("account unauthorized")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	err := server.store.DeletePromotion(ctx.Context(), params.PromotionID)
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
