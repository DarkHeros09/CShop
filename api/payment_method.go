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

type createPaymentMethodRequest struct {
	UserID        int64  `json:"user_id" binding:"required,min=1"`
	PaymentTypeID int64  `json:"payment_method_id" binding:"required,min=1"`
	Provider      string `json:"provider" binding:"required"`
}

func (server *Server) createPaymentMethod(ctx *gin.Context) {
	var req createPaymentMethodRequest

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

	arg1 := db.CreatePaymentMethodParams{
		UserID:        authPayload.UserID,
		PaymentTypeID: req.PaymentTypeID,
		Provider:      req.Provider,
	}

	paymentMethod, err := server.store.CreatePaymentMethod(ctx, arg1)
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

	ctx.JSON(http.StatusOK, paymentMethod)
}

// //////////////* Get API //////////////

type getPaymentMethodUriRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type getPaymentMethodJsonRequest struct {
	UserID int64 `json:"user_id" binding:"required,min=1"`
}

func (server *Server) getPaymentMethod(ctx *gin.Context) {
	var uri getPaymentMethodUriRequest
	var req getPaymentMethodJsonRequest

	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.GetPaymentMethodParams{
		ID:     uri.ID,
		UserID: req.UserID,
	}

	paymentMethod, err := server.store.GetPaymentMethod(ctx, arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.UserPayload)
	if paymentMethod.UserID != authPayload.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, paymentMethod)
}

// //////////////* List API //////////////

type listPaymentMethodesRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server) listPaymentMethodes(ctx *gin.Context) {
	var req listPaymentMethodesRequest

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.UserPayload)

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	arg := db.ListPaymentMethodsParams{
		UserID: authPayload.UserID,
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}
	paymentMethodes, err := server.store.ListPaymentMethods(ctx, arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, paymentMethodes)
}

// //////////////* UPDATE API ///////////////
type updatePaymentMethodUriRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type updatePaymentMethodJsonRequest struct {
	UserID        int64  `json:"user_id" binding:"required,min=1"`
	PaymentTypeID int64  `json:"payment_type_id" binding:"required,min=1"`
	Provider      string `json:"provider" binding:"required"`
}

func (server *Server) updatePaymentMethod(ctx *gin.Context) {
	var uri updatePaymentMethodUriRequest
	var req updatePaymentMethodJsonRequest

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

	arg := db.UpdatePaymentMethodParams{
		ID:            uri.ID,
		UserID:        null.IntFromPtr(&req.UserID),
		PaymentTypeID: null.IntFromPtr(&req.PaymentTypeID),
		Provider:      null.StringFromPtr(&req.Provider),
	}

	paymentMethod, err := server.store.UpdatePaymentMethod(ctx, arg)
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

	ctx.JSON(http.StatusOK, paymentMethod)
}

// ////////////* Delete API //////////////
type deletePaymentMethodUriRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type deletePaymentMethodJsonRequest struct {
	UserID int64 `json:"user_id" binding:"required,min=1"`
}

func (server *Server) deletePaymentMethod(ctx *gin.Context) {
	var uri deletePaymentMethodUriRequest
	var req deletePaymentMethodJsonRequest

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

	arg := db.DeletePaymentMethodParams{
		ID:     uri.ID,
		UserID: authPayload.UserID,
	}

	_, err := server.store.DeletePaymentMethod(ctx, arg)
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
