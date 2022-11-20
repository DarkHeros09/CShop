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

type createUserReviewRequest struct {
	UserID           int64 `json:"user_id" binding:"required,min=1"`
	OrderedProductID int64 `json:"ordered_product_id" binding:"required,min=1"`
	RatingValue      int32 `json:"rating_value" binding:"required,min=0,max=5"`
}

func (server *Server) createUserReview(ctx *gin.Context) {
	var req createUserReviewRequest

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

	arg := db.CreateUserReviewParams{
		UserID:           authPayload.UserID,
		OrderedProductID: req.OrderedProductID,
		RatingValue:      req.RatingValue,
	}

	userReview, err := server.store.CreateUserReview(ctx, arg)
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

	ctx.JSON(http.StatusOK, userReview)
}

//////////////* Get API //////////////

type getUserReviewUriRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type getUserReviewJsonRequest struct {
	UserID int64 `json:"user_id" binding:"required,min=1"`
}

func (server *Server) getUserReview(ctx *gin.Context) {
	var uri getUserReviewUriRequest
	var req getUserReviewJsonRequest

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

	arg := db.GetUserReviewParams{
		ID:     uri.ID,
		UserID: authPayload.UserID,
	}
	userReview, err := server.store.GetUserReview(ctx, arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if userReview.UserID != authPayload.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, userReview)
}

//////////////* List API //////////////

type listUserReviewsRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server) listUserReviews(ctx *gin.Context) {
	var req listUserReviewsRequest

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.UserPayload)

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	arg := db.ListUserReviewsParams{
		UserID: authPayload.UserID,
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}
	userReviews, err := server.store.ListUserReviews(ctx, arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, userReviews)
}

// ////////////* UPDATE API //////////////
type updateUserReviewUriRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type updateUserReviewJsonRequest struct {
	UserID           int64 `json:"user_id" binding:"required,min=1"`
	OrderedProductID int64 `json:"ordered_product_id" binding:"omitempty,required,min=1"`
	RatingValue      int64 `json:"rating_value" binding:"omitempty,required,min=0,max=5"`
}

func (server *Server) updateUserReview(ctx *gin.Context) {
	var uri updateUserReviewUriRequest
	var req updateUserReviewJsonRequest

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

	arg1 := db.UpdateUserReviewParams{
		UserID:           authPayload.UserID,
		OrderedProductID: null.IntFromPtr(&req.OrderedProductID),
		RatingValue:      null.IntFromPtr(&req.RatingValue),
		ID:               uri.ID,
	}

	userReview, err := server.store.UpdateUserReview(ctx, arg1)
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

	ctx.JSON(http.StatusOK, userReview)
}

// ////////////* Delete API //////////////
type deleteUserReviewUriRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type deleteUserReviewJsonRequest struct {
	UserID int64 `json:"user_id" binding:"required,min=1"`
}

func (server *Server) deleteUserReview(ctx *gin.Context) {
	var uri deleteUserReviewUriRequest
	var req deleteUserReviewJsonRequest

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

	arg := db.DeleteUserReviewParams{
		ID:     uri.ID,
		UserID: req.UserID,
	}

	_, err := server.store.DeleteUserReview(ctx, arg)
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
