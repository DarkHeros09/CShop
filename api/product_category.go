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

type createProductCategoryRequest struct {
	CategoryName     string `json:"category_name" binding:"required,alphanum"`
	ParentCategoryID int64  `json:"parent_category_id" binding:"required,min=1"`
}

func (server *Server) createProductCategory(ctx *gin.Context) {
	var req createProductCategoryRequest

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

	arg := db.CreateProductCategoryParams{
		ParentCategoryID: null.IntFromPtr(&req.ParentCategoryID),
		CategoryName:     req.CategoryName,
	}

	productCategory, err := server.store.CreateProductCategory(ctx, arg)
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

	ctx.JSON(http.StatusOK, productCategory)
}

//////////////* Get API //////////////

type getProductCategoryRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getProductCategory(ctx *gin.Context) {
	var req getProductCategoryRequest

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	productCategory, err := server.store.GetProductCategory(ctx, req.ID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, productCategory)
}

//////////////* List API //////////////

type listProductCategoriesRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server) listProductCategories(ctx *gin.Context) {
	var req listProductCategoriesRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListProductCategoriesParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}
	productCategorys, err := server.store.ListProductCategories(ctx, arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// requestETag := ctx.GetHeader("If-None-Match")
	// generatedETag := etag.Generate([]byte(fmt.Sprint(productCategorys)), true)

	// if requestETag == generatedETag {
	// 	ctx.JSON(http.StatusNotModified, nil)

	// } else {
	// 	ctx.Header("ETag", generatedETag)
	// 	ctx.JSON(http.StatusOK, productCategorys)
	// }

	ctx.JSON(http.StatusOK, productCategorys)

}

//////////////* Update API //////////////

type updateProductCategoryUriRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type updateProductCategoryJsonRequest struct {
	CategoryName     string `json:"category_name" binding:"omitempty,required"`
	ParentCategoryID int64  `json:"parent_category_id" binding:"omitempty,required,min=1"`
}

func (server *Server) updateProductCategory(ctx *gin.Context) {
	var uri updateProductCategoryUriRequest
	var req updateProductCategoryJsonRequest

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

	arg := db.UpdateProductCategoryParams{
		ID:               uri.ID,
		CategoryName:     req.CategoryName,
		ParentCategoryID: null.IntFromPtr(&req.ParentCategoryID),
	}

	productCategory, err := server.store.UpdateProductCategory(ctx, arg)
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
	ctx.JSON(http.StatusOK, productCategory)
}

//////////////* Delete API //////////////

type deleteProductCategoryUriRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type deleteProductCategoryJsonRequest struct {
	ParentCategoryID int64 `json:"parent_category_id" binding:"required,min=1"`
}

func (server *Server) deleteProductCategory(ctx *gin.Context) {
	var uri deleteProductCategoryUriRequest
	var req deleteProductCategoryJsonRequest

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

	arg := db.DeleteProductCategoryParams{
		ID:               uri.ID,
		ParentCategoryID: null.IntFromPtr(&req.ParentCategoryID),
	}

	err := server.store.DeleteProductCategory(ctx, arg)
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
