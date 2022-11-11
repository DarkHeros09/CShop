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

type createVariationRequest struct {
	Name       string `json:"name" binding:"required,alphanum"`
	CategoryID int64  `json:"category_id" binding:"required,min=1"`
}

func (server *Server) createVariation(ctx *gin.Context) {
	var req createVariationRequest

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

	arg := db.CreateVariationParams{
		CategoryID: req.CategoryID,
		Name:       req.Name,
	}

	variation, err := server.store.CreateVariation(ctx, arg)
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

	ctx.JSON(http.StatusOK, variation)
}

//////////////* Get API //////////////

type getVariationRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getVariation(ctx *gin.Context) {
	var req getVariationRequest

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	variation, err := server.store.GetVariation(ctx, req.ID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, variation)
}

//////////////* List API //////////////

type listVariationsRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server) listVariations(ctx *gin.Context) {
	var req listVariationsRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListVariationsParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}
	variations, err := server.store.ListVariations(ctx, arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// requestETag := ctx.GetHeader("If-None-Match")
	// generatedETag := etag.Generate([]byte(fmt.Sprint(Variations)), true)

	// if requestETag == generatedETag {
	// 	ctx.JSON(http.StatusNotModified, nil)

	// } else {
	// 	ctx.Header("ETag", generatedETag)
	// 	ctx.JSON(http.StatusOK, Variations)
	// }

	ctx.JSON(http.StatusOK, variations)

}

//////////////* Update API //////////////

type updateVariationUriRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type updateVariationJsonRequest struct {
	Name       string `json:"name" binding:"omitempty,required"`
	CategoryID int64  `json:"category_id" binding:"omitempty,required,min=1"`
}

func (server *Server) updateVariation(ctx *gin.Context) {
	var uri updateVariationUriRequest
	var req updateVariationJsonRequest

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

	arg := db.UpdateVariationParams{
		ID:         uri.ID,
		Name:       null.StringFromPtr(&req.Name),
		CategoryID: null.IntFromPtr(&req.CategoryID),
	}

	variation, err := server.store.UpdateVariation(ctx, arg)
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
	ctx.JSON(http.StatusOK, variation)
}

//////////////* Delete API //////////////

type deleteVariationUriRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type deleteVariationJsonRequest struct {
	CategoryID int64 `json:"category_id" binding:"required,min=1"`
}

func (server *Server) deleteVariation(ctx *gin.Context) {
	var uri deleteVariationUriRequest
	var req deleteVariationJsonRequest

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

	err := server.store.DeleteVariation(ctx, uri.ID)
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
