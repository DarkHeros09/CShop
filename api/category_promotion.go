package api

import (
	"errors"
	"net/http"

	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/token"
	"github.com/gin-gonic/gin"
	"github.com/guregu/null"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

//////////////* Create API //////////////

type createCategoryPromotionRequest struct {
	CategoryID  int64 `json:"category_id" binding:"required,min=1"`
	PromotionID int64 `json:"promotion_id" binding:"required,min=1"`
	Active      bool  `json:"active" binding:"boolean"`
}

func (server *Server) createCategoryPromotion(ctx *gin.Context) {
	var req createCategoryPromotionRequest

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

	arg := db.CreateCategoryPromotionParams{
		CategoryID:  req.CategoryID,
		PromotionID: req.PromotionID,
		Active:      req.Active,
	}

	categoryPromotion, err := server.store.CreateCategoryPromotion(ctx, arg)
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

	ctx.JSON(http.StatusOK, categoryPromotion)
}

//////////////* Get API //////////////

type getCategoryPromotionRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getCategoryPromotion(ctx *gin.Context) {
	var req getCategoryPromotionRequest

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	categoryPromotion, err := server.store.GetCategoryPromotion(ctx, req.ID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, categoryPromotion)
}

//////////////* List API //////////////

type listCategoryPromotionsRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server) listCategoryPromotions(ctx *gin.Context) {
	var req listCategoryPromotionsRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListCategoryPromotionsParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}
	CategoryPromotions, err := server.store.ListCategoryPromotions(ctx, arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// requestETag := ctx.GetHeader("If-None-Match")
	// generatedETag := etag.Generate([]byte(fmt.Sprint(CategoryPromotions)), true)

	// if requestETag == generatedETag {
	// 	ctx.JSON(http.StatusNotModified, nil)

	// } else {
	// 	ctx.Header("ETag", generatedETag)
	// 	ctx.JSON(http.StatusOK, CategoryPromotions)
	// }

	ctx.JSON(http.StatusOK, CategoryPromotions)

}

//////////////* Update API //////////////

type updateCategoryPromotionUriRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type updateCategoryPromotionJsonRequest struct {
	CategoryID  int64 `json:"category_id" binding:"omitempty,required,min=1"`
	PromotionID int64 `json:"promotion_id" binding:"omitempty,required,min=1"`
	Active      bool  `json:"active" binding:"omitempty,required,boolean"`
}

func (server *Server) updateCategoryPromotion(ctx *gin.Context) {
	var uri updateCategoryPromotionUriRequest
	var req updateCategoryPromotionJsonRequest

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

	arg := db.UpdateCategoryPromotionParams{
		CategoryID:  req.CategoryID,
		PromotionID: null.IntFromPtr(&req.PromotionID),
		Active:      null.BoolFromPtr(&req.Active),
	}

	categoryPromotion, err := server.store.UpdateCategoryPromotion(ctx, arg)
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
	ctx.JSON(http.StatusOK, categoryPromotion)
}

//////////////* Delete API //////////////

type deleteCategoryPromotionRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) deleteCategoryPromotion(ctx *gin.Context) {
	var req deleteCategoryPromotionRequest

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

	err := server.store.DeleteCategoryPromotion(ctx, req.ID)
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
