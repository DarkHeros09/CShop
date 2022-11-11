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

type createProductConfigurationRequest struct {
	ProductItemID     int64 `json:"product_item_id" binding:"required,min=1"`
	VariationOptionID int64 `json:"variation_option_id" binding:"required,min=1"`
}

func (server *Server) createProductConfiguration(ctx *gin.Context) {
	var req createProductConfigurationRequest

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

	arg := db.CreateProductConfigurationParams{
		ProductItemID:     req.ProductItemID,
		VariationOptionID: req.VariationOptionID,
	}

	productConfiguration, err := server.store.CreateProductConfiguration(ctx, arg)
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

	ctx.JSON(http.StatusOK, productConfiguration)
}

//////////////* Get API //////////////

type getProductConfigurationRequest struct {
	ProductItemID int64 `uri:"product-item-id" binding:"required,min=1"`
}

func (server *Server) getProductConfiguration(ctx *gin.Context) {
	var req getProductConfigurationRequest

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	productConfiguration, err := server.store.GetProductConfiguration(ctx, req.ProductItemID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, productConfiguration)
}

//////////////* List API //////////////

type listProductConfigurationsRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server) listProductConfigurations(ctx *gin.Context) {
	var req listProductConfigurationsRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListProductConfigurationsParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}
	productConfigurations, err := server.store.ListProductConfigurations(ctx, arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// requestETag := ctx.GetHeader("If-None-Match")
	// generatedETag := etag.Generate([]byte(fmt.Sprint(productConfigurations)), true)

	// if requestETag == generatedETag {
	// 	ctx.JSON(http.StatusNotModified, nil)

	// } else {
	// 	ctx.Header("ETag", generatedETag)
	// 	ctx.JSON(http.StatusOK, productConfigurations)
	// }

	ctx.JSON(http.StatusOK, productConfigurations)

}

//////////////* Update API //////////////

type updateProductConfigurationUriRequest struct {
	ProductItemID int64 `uri:"product-item-id" binding:"required,min=1"`
}

type updateProductConfigurationJsonRequest struct {
	VariationOptionID int64 `json:"variation_option_id" binding:"omitempty,required,min=1"`
}

func (server *Server) updateProductConfiguration(ctx *gin.Context) {
	var uri updateProductConfigurationUriRequest
	var req updateProductConfigurationJsonRequest

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

	arg := db.UpdateProductConfigurationParams{
		VariationOptionID: null.IntFromPtr(&req.VariationOptionID),
		ProductItemID:     uri.ProductItemID,
	}

	productConfiguration, err := server.store.UpdateProductConfiguration(ctx, arg)
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
	ctx.JSON(http.StatusOK, productConfiguration)
}

//////////////* Delete API //////////////

type deleteProductConfigurationRequest struct {
	ProductItemID int64 `uri:"product-item-id" binding:"required,min=1"`
}

func (server *Server) deleteProductConfiguration(ctx *gin.Context) {
	var req deleteProductConfigurationRequest

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

	err := server.store.DeleteProductConfiguration(ctx, req.ProductItemID)
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
