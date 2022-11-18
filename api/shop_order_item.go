package api

import (
	"errors"
	"net/http"

	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/token"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
)

//////////////* Get API //////////////

type getShopOrderItemUriRequest struct {
	ShopOrderID int64 `uri:"shop-order-id" binding:"required,min=1"`
}

type getShopOrderItemJsonRequest struct {
	UserID int64 `json:"user_id" binding:"required,min=1"`
}

func (server *Server) getShopOrderItem(ctx *gin.Context) {
	var uri getShopOrderItemUriRequest
	var req getShopOrderItemJsonRequest

	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.GetShopOrderItemByUserIDOrderIDParams{
		UserID:  req.UserID,
		OrderID: uri.ShopOrderID,
	}

	shopOrderItem, err := server.store.GetShopOrderItemByUserIDOrderID(ctx, arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.UserPayload)
	if authPayload.UserID != shopOrderItem.UserID.Int64 {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, shopOrderItem)
}

//////////////* List API //////////////

type listShopOrderItemsQueryRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

type listShopOrderItemsJsonRequest struct {
	UserID int64 `json:"user_id" binding:"required,min=1"`
}

func (server *Server) listShopOrderItems(ctx *gin.Context) {
	var query listShopOrderItemsQueryRequest
	var req listShopOrderItemsJsonRequest

	if err := ctx.ShouldBindQuery(&query); err != nil {
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

	arg := db.ListShopOrderItemsByUserIDParams{
		UserID: authPayload.UserID,
		Limit:  query.PageSize,
		Offset: (query.PageID - 1) * query.PageSize,
	}
	shopOrderItems, err := server.store.ListShopOrderItemsByUserID(ctx, arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, shopOrderItems)
}
