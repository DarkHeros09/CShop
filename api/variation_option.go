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

type createVariationOptionRequest struct {
	VariationID int64  `json:"variation_id" binding:"required,min=1"`
	Value       string `json:"value" binding:"required,alphanum"`
}

func (server *Server) createVariationOption(ctx *gin.Context) {
	var req createVariationOptionRequest

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

	arg := db.CreateVariationOptionParams{
		VariationID: req.VariationID,
		Value:       req.Value,
	}

	variationOption, err := server.store.CreateVariationOption(ctx, arg)
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

	ctx.JSON(http.StatusOK, variationOption)
}

//////////////* Get API //////////////

type getVariationOptionRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getVariationOption(ctx *gin.Context) {
	var req getVariationOptionRequest

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	variationOption, err := server.store.GetVariationOption(ctx, req.ID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, variationOption)
}

//////////////* List API //////////////

type listVariationOptionsRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server) listVariationOptions(ctx *gin.Context) {
	var req listVariationOptionsRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListVariationOptionsParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}
	variationOptions, err := server.store.ListVariationOptions(ctx, arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// requestETag := ctx.GetHeader("If-None-Match")
	// generatedETag := etag.Generate([]byte(fmt.Sprint(VariationOptions)), true)

	// if requestETag == generatedETag {
	// 	ctx.JSON(http.StatusNotModified, nil)

	// } else {
	// 	ctx.Header("ETag", generatedETag)
	// 	ctx.JSON(http.StatusOK, VariationOptions)
	// }

	ctx.JSON(http.StatusOK, variationOptions)

}

//////////////* Update API //////////////

type updateVariationOptionUriRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type updateVariationOptionJsonRequest struct {
	Value       string `json:"value" binding:"omitempty,required"`
	VariationID int64  `json:"variation_id" binding:"omitempty,required,min=1"`
}

func (server *Server) updateVariationOption(ctx *gin.Context) {
	var uri updateVariationOptionUriRequest
	var req updateVariationOptionJsonRequest

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

	arg := db.UpdateVariationOptionParams{
		ID:          uri.ID,
		Value:       null.StringFromPtr(&req.Value),
		VariationID: null.IntFromPtr(&req.VariationID),
	}

	variationOption, err := server.store.UpdateVariationOption(ctx, arg)
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
	ctx.JSON(http.StatusOK, variationOption)
}

//////////////* Delete API //////////////

type deleteVariationOptionUriRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type deleteVariationOptionJsonRequest struct {
	VariationID int64 `json:"variation_id" binding:"required,min=1"`
}

func (server *Server) deleteVariationOption(ctx *gin.Context) {
	var uri deleteVariationOptionUriRequest
	var req deleteVariationOptionJsonRequest

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

	err := server.store.DeleteVariationOption(ctx, uri.ID)
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
