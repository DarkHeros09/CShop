package api

import (
	"errors"
	"log"
	"time"

	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/token"
	"github.com/gofiber/fiber/v2"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
)

//////////////* Create API //////////////

type createPromotionParamsRequest struct {
	AdminID int64 `params:"adminId" validate:"required,min=1"`
}
type createPromotionJsonRequest struct {
	Name         string `json:"name" validate:"required,alphanum"`
	Description  string `json:"description" validate:"required"`
	DiscountRate int64  `json:"discount_rate" validate:"required,min=1"`
	Active       bool   `json:"active" validate:"boolean"`
	StartDate    string `json:"start_date" validate:"required"`
	EndDate      string `json:"end_date" validate:"required"`
}

const (
	timeLayout = "2006-01-02T15:04:05.000"
)

func (server *Server) createPromotion(ctx *fiber.Ctx) error {
	params := &createPromotionParamsRequest{}
	req := &createPromotionJsonRequest{}

	if err := parseAndValidate(ctx, Input{params: params, req: req}); err != nil {
		log.Fatal(err)
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationAdminPayloadKey).(*token.AdminPayload)
	if authPayload.AdminID != params.AdminID || authPayload.TypeID != 1 || !authPayload.Active {
		err := errors.New("account unauthorized")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}
	startDate, err := time.Parse(timeLayout, req.StartDate)
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}
	endDate, err := time.Parse(timeLayout, req.EndDate)
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	arg := db.AdminCreatePromotionParams{
		AdminID:      authPayload.AdminID,
		Name:         req.Name,
		Description:  req.Description,
		DiscountRate: req.DiscountRate,
		Active:       req.Active,
		StartDate:    startDate,
		EndDate:      endDate,
	}

	promotion, err := server.store.AdminCreatePromotion(ctx.Context(), arg)
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
	ID int64 `params:"promotionId" validate:"required,min=1"`
}

func (server *Server) getPromotion(ctx *fiber.Ctx) error {
	params := &getPromotionParamsRequest{}

	if err := parseAndValidate(ctx, Input{params: params}); err != nil {
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

// type listPromotionsQueryRequest struct {
// 	PageID   int32 `query:"page_id" validate:"required,min=1"`
// 	PageSize int32 `query:"page_size" validate:"required,min=5,max=10"`
// }

func (server *Server) listPromotions(ctx *fiber.Ctx) error {
	// query := &listPromotionsQueryRequest{}

	// if err := parseAndValidate(ctx, Input{query: query}); err != nil {
	// 	ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
	// 	return nil
	// }

	// arg := db.ListPromotionsParams{
	// 	Limit:  query.PageSize,
	// 	Offset: (query.PageID - 1) * query.PageSize,
	// }
	promotions, err := server.store.ListPromotions(ctx.Context())
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

func parseTimeOrNil(layout, value string) (*time.Time, error) {
	parsedTime, err := time.Parse(layout, value)
	if err != nil {
		return nil, err // Return nil on error
	}
	return &parsedTime, nil // Return a pointer to the parsed time
}

type updatePromotionParamsRequest struct {
	AdminID     int64 `params:"adminId" validate:"required,min=1"`
	PromotionID int64 `params:"promotionId" validate:"required,min=1"`
}

type updatePromotionJsonRequest struct {
	Name         *string `json:"name" validate:"omitempty,required,alphanum"`
	Description  *string `json:"description" validate:"omitempty,required"`
	DiscountRate *int64  `json:"discount_rate" validate:"omitempty,required,min=1"`
	Active       *bool   `json:"active" validate:"omitempty,required,boolean"`
	StartDate    *string `json:"start_date" validate:"omitempty,required"`
	EndDate      *string `json:"end_date" validate:"omitempty,required"`
}

func (server *Server) updatePromotion(ctx *fiber.Ctx) error {
	params := &updatePromotionParamsRequest{}
	req := &updatePromotionJsonRequest{}

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

	startDate, err := parseTimeOrNil(timeLayout, *req.StartDate)
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	endDate, err := parseTimeOrNil(timeLayout, *req.EndDate)
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	arg := db.AdminUpdatePromotionParams{
		AdminID:      authPayload.AdminID,
		ID:           params.PromotionID,
		Name:         null.StringFromPtr(req.Name),
		Description:  null.StringFromPtr(req.Description),
		DiscountRate: null.IntFromPtr(req.DiscountRate),
		Active:       null.BoolFromPtr(req.Active),
		StartDate:    null.TimeFromPtr(startDate),
		EndDate:      null.TimeFromPtr(endDate),
	}

	promotion, err := server.store.AdminUpdatePromotion(ctx.Context(), arg)
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
	AdminID     int64 `params:"adminId" validate:"required,min=1"`
	PromotionID int64 `params:"promotionId" validate:"required,min=1"`
}

func (server *Server) deletePromotion(ctx *fiber.Ctx) error {
	params := &deletePromotionParamsRequest{}

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
