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

type createProductPromotionRequest struct {
	ProductID   int64 `json:"product_id" binding:"required,min=1"`
	PromotionID int64 `json:"promotion_id" binding:"required,min=1"`
	Active      bool  `json:"active" binding:"boolean"`
}

func (server *Server) createProductPromotion(ctx *gin.Context) {
	var req createProductPromotionRequest

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

	arg := db.CreateProductPromotionParams{
		ProductID:   req.ProductID,
		PromotionID: req.PromotionID,
		Active:      req.Active,
	}

	productPromotion, err := server.store.CreateProductPromotion(ctx, arg)
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

	ctx.JSON(http.StatusOK, productPromotion)
}

//////////////* Get API //////////////

type getProductPromotionUriRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type getProductPromotionJsonRequest struct {
	ProductID int64 `json:"product_id" binding:"required,min=1"`
}

func (server *Server) getProductPromotion(ctx *gin.Context) {
	var uri getProductPromotionUriRequest
	var req getProductPromotionJsonRequest

	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.GetProductPromotionParams{
		ProductID:   req.ProductID,
		PromotionID: uri.ID,
	}

	productPromotion, err := server.store.GetProductPromotion(ctx, arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, productPromotion)
}

//////////////* List API //////////////

type listProductPromotionsRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server) listProductPromotions(ctx *gin.Context) {
	var req listProductPromotionsRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListProductPromotionsParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}
	productPromotions, err := server.store.ListProductPromotions(ctx, arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// requestETag := ctx.GetHeader("If-None-Match")
	// generatedETag := etag.Generate([]byte(fmt.Sprint(ProductPromotions)), true)

	// if requestETag == generatedETag {
	// 	ctx.JSON(http.StatusNotModified, nil)

	// } else {
	// 	ctx.Header("ETag", generatedETag)
	// 	ctx.JSON(http.StatusOK, ProductPromotions)
	// }

	ctx.JSON(http.StatusOK, productPromotions)

}

//////////////* Update API //////////////

type updateProductPromotionUriRequest struct {
	PromotionID int64 `uri:"id" binding:"required,min=1"`
}

type updateProductPromotionJsonRequest struct {
	ProductID int64 `json:"product_id" binding:"required,min=1"`
	Active    bool  `json:"active" binding:"omitempty,required,boolean"`
}

func (server *Server) updateProductPromotion(ctx *gin.Context) {
	var uri updateProductPromotionUriRequest
	var req updateProductPromotionJsonRequest

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

	arg := db.UpdateProductPromotionParams{
		ProductID:   req.ProductID,
		PromotionID: uri.PromotionID,
		Active:      null.BoolFromPtr(&req.Active),
	}

	productPromotion, err := server.store.UpdateProductPromotion(ctx, arg)
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
	ctx.JSON(http.StatusOK, productPromotion)
}

//////////////* Delete API //////////////

type deleteProductPromotionUriRequest struct {
	PromotionID int64 `uri:"id" binding:"required,min=1"`
}
type deleteProductPromotionJsonRequest struct {
	ProductID int64 `json:"product_id" binding:"required,min=1"`
}

func (server *Server) deleteProductPromotion(ctx *gin.Context) {
	var uri deleteProductPromotionUriRequest
	var req deleteProductPromotionJsonRequest

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

	arg := db.DeleteProductPromotionParams{
		ProductID:   req.ProductID,
		PromotionID: uri.PromotionID,
	}

	err := server.store.DeleteProductPromotion(ctx, arg)
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
