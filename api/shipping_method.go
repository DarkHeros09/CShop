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

type createShippingMethodRequest struct {
	Name   string `json:"name" binding:"required"`
	Price  string `json:"price" binding:"required"`
	UserID int64  `json:"user_id" binding:"required,min=1"`
}

func (server *Server) createShippingMethod(ctx *gin.Context) {
	var req createShippingMethodRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.UserPayload)
	if authPayload.UserID != req.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	arg := db.CreateShippingMethodParams{
		Name:  req.Name,
		Price: req.Price,
	}

	ShippingMethod, err := server.store.CreateShippingMethod(ctx, arg)
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

	ctx.JSON(http.StatusOK, ShippingMethod)
}

// //////////////* Get API //////////////

type getShippingMethodUriRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type getShippingMethodJsonRequest struct {
	UserID int64 `json:"user_id" binding:"required,min=1"`
}

func (server *Server) getShippingMethod(ctx *gin.Context) {
	var uri getShippingMethodUriRequest
	var req getShippingMethodJsonRequest

	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.GetShippingMethodByUserIDParams{
		ID:     uri.ID,
		UserID: req.UserID,
	}

	ShippingMethod, err := server.store.GetShippingMethodByUserID(ctx, arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.UserPayload)
	if ShippingMethod.UserID.Int64 != authPayload.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, ShippingMethod)
}

// //////////////* List API //////////////

type listShippingMethodesRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server) listShippingMethodes(ctx *gin.Context) {
	var req listShippingMethodesRequest

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.UserPayload)

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	arg := db.ListShippingMethodsByUserIDParams{
		UserID: authPayload.UserID,
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}
	ShippingMethodes, err := server.store.ListShippingMethodsByUserID(ctx, arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, ShippingMethodes)
}

// //////////////* UPDATE API ///////////////
type updateShippingMethodUriRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type updateShippingMethodJsonRequest struct {
	UserID int64  `json:"user_id" binding:"required,min=1"`
	Name   string `json:"name" binding:"omitempty,required"`
	Price  string `json:"price" binding:"omitempty,required"`
}

func (server *Server) updateShippingMethod(ctx *gin.Context) {
	var uri updateShippingMethodUriRequest
	var req updateShippingMethodJsonRequest

	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.UserPayload)
	if authPayload.UserID != req.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	arg := db.UpdateShippingMethodParams{
		Name:  null.StringFromPtr(&req.Name),
		Price: null.StringFromPtr(&req.Price),
		ID:    uri.ID,
	}

	ShippingMethod, err := server.store.UpdateShippingMethod(ctx, arg)
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

	ctx.JSON(http.StatusOK, ShippingMethod)
}

// ////////////* Delete API //////////////

type deleteShippingMethodUriRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) deleteShippingMethod(ctx *gin.Context) {
	var uri deleteShippingMethodUriRequest

	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.AdminPayload)
	if authPayload.AdminID == 0 || authPayload.TypeID != 1 || !authPayload.Active {
		err := errors.New("account unauthorized")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	err := server.store.DeleteShippingMethod(ctx, uri.ID)
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
