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

type createProductItemRequest struct {
	ProductID    int64  `json:"product_id" binding:"required,min=1"`
	ProductSKU   int64  `json:"product_sku" binding:"required"`
	QtyInStock   int32  `json:"qty_in_stock" binding:"required"`
	ProductImage string `json:"product_image" binding:"required,url"`
	Price        string `json:"price" binding:"required"`
	Active       bool   `json:"active" binding:"boolean"`
}

func (server *Server) createProductItem(ctx *gin.Context) {
	var req createProductItemRequest

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

	arg := db.CreateProductItemParams{
		ProductID:    req.ProductID,
		ProductSku:   req.ProductSKU,
		QtyInStock:   req.QtyInStock,
		ProductImage: req.ProductImage,
		Price:        req.Price,
		Active:       req.Active,
	}

	productItem, err := server.store.CreateProductItem(ctx, arg)
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

	ctx.JSON(http.StatusOK, productItem)
}

//////////////* Get API //////////////

type getProductItemRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getProductItem(ctx *gin.Context) {
	var req getProductItemRequest

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	productItem, err := server.store.GetProductItem(ctx, req.ID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, productItem)
}

//////////////* List API //////////////

type listProductItemsRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server) listProductItems(ctx *gin.Context) {
	var req listProductItemsRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListProductItemsParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}
	productItems, err := server.store.ListProductItems(ctx, arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// requestETag := ctx.GetHeader("If-None-Match")
	// generatedETag := etag.Generate([]byte(fmt.Sprint(productItems)), true)

	// if requestETag == generatedETag {
	// 	ctx.JSON(http.StatusNotModified, nil)

	// } else {
	// 	ctx.Header("ETag", generatedETag)
	// 	ctx.JSON(http.StatusOK, productItems)
	// }

	ctx.JSON(http.StatusOK, productItems)

}

//////////////* Update API //////////////

type updateProductItemUriRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type updateProductItemJsonRequest struct {
	ProductID    int64  `json:"product_id" binding:"omitempty,omitempty,required,min=1"`
	ProductSKU   int64  `json:"product_sku" binding:"omitempty,required"`
	QtyInStock   int64  `json:"qty_in_stock" binding:"omitempty,required"`
	ProductImage string `json:"product_image" binding:"omitempty,required,url"`
	Price        string `json:"price" binding:"omitempty,required"`
	Active       bool   `json:"active" binding:"boolean"`
}

func (server *Server) updateProductItem(ctx *gin.Context) {
	var uri updateProductItemUriRequest
	var req updateProductItemJsonRequest

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

	arg := db.UpdateProductItemParams{
		ProductID:    null.IntFromPtr(&req.ProductID),
		ProductSku:   null.IntFromPtr(&req.ProductSKU),
		QtyInStock:   null.IntFromPtr(&req.QtyInStock),
		ProductImage: null.StringFromPtr(&req.ProductImage),
		Price:        null.StringFromPtr(&req.Price),
		Active:       null.BoolFromPtr(&req.Active),
		ID:           uri.ID,
	}

	productItem, err := server.store.UpdateProductItem(ctx, arg)
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
	ctx.JSON(http.StatusOK, productItem)
}

//////////////* Delete API //////////////

type deleteProductItemRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) deleteProductItem(ctx *gin.Context) {
	var req deleteProductItemRequest

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

	err := server.store.DeleteProductItem(ctx, req.ID)
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
