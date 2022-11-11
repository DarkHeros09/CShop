package api

import (
	"errors"
	"net/http"
	"time"

	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/token"
	"github.com/gin-gonic/gin"
	"github.com/guregu/null"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

//////////////* Create API //////////////

type createPromotionRequest struct {
	Name         string    `json:"name" binding:"required,alphanum"`
	Description  string    `json:"description" binding:"required"`
	DiscountRate int64     `json:"discount_rate" binding:"required,min=1"`
	Active       bool      `json:"active" binding:"boolean"`
	StartDate    time.Time `json:"start_date" binding:"required"`
	EndDate      time.Time `json:"end_date" binding:"required"`
}

func (server *Server) createPromotion(ctx *gin.Context) {
	var req createPromotionRequest

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.AdminPayload)
	if authPayload.AdminID == 0 || authPayload.TypeID != 1 || !authPayload.Active {
		err := errors.New("account unauthorized")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CreatePromotionParams{
		Name:         req.Name,
		Description:  req.Description,
		DiscountRate: req.DiscountRate,
		Active:       req.Active,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
	}

	promotion, err := server.store.CreatePromotion(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pgconn.PgError); ok {
			switch pqErr.Message {
			case "foreign_key_violation", "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, promotion)
}

//////////////* Get API //////////////

type getPromotionRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getPromotion(ctx *gin.Context) {
	var req getPromotionRequest

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	promotion, err := server.store.GetPromotion(ctx, req.ID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, promotion)
}

//////////////* List API //////////////

type listPromotionsRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server) listPromotions(ctx *gin.Context) {
	var req listPromotionsRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListPromotionsParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}
	promotions, err := server.store.ListPromotions(ctx, arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// requestETag := ctx.GetHeader("If-None-Match")
	// generatedETag := etag.Generate([]byte(fmt.Sprint(Promotions)), true)

	// if requestETag == generatedETag {
	// 	ctx.JSON(http.StatusNotModified, nil)

	// } else {
	// 	ctx.Header("ETag", generatedETag)
	// 	ctx.JSON(http.StatusOK, Promotions)
	// }

	ctx.JSON(http.StatusOK, promotions)

}

//////////////* Update API //////////////

type updatePromotionUriRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type updatePromotionJsonRequest struct {
	Name         string    `json:"name" binding:"omitempty,required,alphanum"`
	Description  string    `json:"description" binding:"omitempty,required"`
	DiscountRate int64     `json:"discount_rate" binding:"omitempty,required,min=1"`
	Active       bool      `json:"active" binding:"omitempty,required,boolean"`
	StartDate    time.Time `json:"start_date" binding:"omitempty,required"`
	EndDate      time.Time `json:"end_date" binding:"omitempty,required"`
}

func (server *Server) updatePromotion(ctx *gin.Context) {
	var uri updatePromotionUriRequest
	var req updatePromotionJsonRequest

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.AdminPayload)
	if authPayload.AdminID == 0 || authPayload.TypeID != 1 || !authPayload.Active {
		err := errors.New("account unauthorized")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.UpdatePromotionParams{
		ID:           uri.ID,
		Name:         null.StringFromPtr(&req.Name),
		Description:  null.StringFromPtr(&req.Description),
		DiscountRate: null.IntFromPtr(&req.DiscountRate),
		Active:       null.BoolFromPtr(&req.Active),
		StartDate:    null.TimeFromPtr(&req.StartDate),
		EndDate:      null.TimeFromPtr(&req.EndDate),
	}

	promotion, err := server.store.UpdatePromotion(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pgconn.PgError); ok {
			switch pqErr.Message {
			case "foreign_key_violation", "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, promotion)
}

//////////////* Delete API //////////////

type deletePromotionRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) deletePromotion(ctx *gin.Context) {
	var req deletePromotionRequest

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.AdminPayload)
	if authPayload.AdminID == 0 || authPayload.TypeID != 1 || !authPayload.Active {
		err := errors.New("account unauthorized")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err := server.store.DeletePromotion(ctx, req.ID)
	if err != nil {
		if pqErr, ok := err.(*pgconn.PgError); ok {
			switch pqErr.Message {
			case "foreign_key_violation", "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		} else if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}
