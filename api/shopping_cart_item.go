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

type createShoppingCartItemRequest struct {
	UserID        int64 `json:"user_id" binding:"required,min=1"`
	ProductItemID int64 `json:"product_item_id" binding:"required,min=1"`
	QTY           int32 `json:"qty" binding:"required,min=1"`
}

func (server *Server) createShoppingCartItem(ctx *gin.Context) {
	var req createShoppingCartItemRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	shoppingCart, err := server.store.GetShoppingCartByUserID(ctx, req.UserID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.UserPayload)
	if authPayload.UserID != shoppingCart.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	arg := db.CreateShoppingCartItemParams{
		ShoppingCartID: shoppingCart.ID,
		ProductItemID:  req.ProductItemID,
		Qty:            req.QTY,
	}

	shoppingCartItem, err := server.store.CreateShoppingCartItem(ctx, arg)
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

	ctx.JSON(http.StatusOK, shoppingCartItem)
}

//////////////* Get API //////////////

type getShoppingCartItemUriRequest struct {
	ShoppingCartID int64 `uri:"shopping-cart-id" binding:"required,min=1"`
}

type getShoppingCartItemJsonRequest struct {
	UserID int64 `json:"user_id" binding:"required,min=1"`
}

func (server *Server) getShoppingCartItem(ctx *gin.Context) {
	var uri getShoppingCartItemUriRequest
	var req getShoppingCartItemJsonRequest

	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.GetShoppingCartItemByUserIDCartIDParams{
		UserID:         req.UserID,
		ShoppingCartID: uri.ShoppingCartID,
	}

	shoppingCartItem, err := server.store.GetShoppingCartItemByUserIDCartID(ctx, arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.UserPayload)
	if authPayload.UserID != shoppingCartItem.UserID.Int64 {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, shoppingCartItem)
}

//////////////* List API //////////////

type listShoppingCartItemsRequest struct {
	UserID int64 `json:"user_id" binding:"required,min=1"`
}

func (server *Server) listShoppingCartItems(ctx *gin.Context) {
	var req listShoppingCartItemsRequest

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
	ShoppingCartItems, err := server.store.ListShoppingCartItemsByUserID(ctx, authPayload.UserID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, ShoppingCartItems)
}

// ////////////* UPDATE API //////////////
type updateShoppingCartItemUriRequest struct {
	ShoppingCartID int64 `uri:"shopping-cart-id" binding:"required,min=1"`
}

type updateShoppingCartItemJsonRequest struct {
	ID            int64 `json:"id" binding:"required,min=1"`
	UserID        int64 `json:"user_id" binding:"required,min=1"`
	ProductItemID int64 `json:"product_item_id" binding:"omitempty,required"`
	QTY           int64 `json:"qty" binding:"omitempty,required"`
}

func (server *Server) updateShoppingCartItem(ctx *gin.Context) {
	var uri updateShoppingCartItemUriRequest
	var req updateShoppingCartItemJsonRequest

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

	arg := db.UpdateShoppingCartItemParams{
		ID:             req.ID,
		ShoppingCartID: uri.ShoppingCartID,
		ProductItemID:  null.IntFromPtr(&req.ProductItemID),
		Qty:            null.IntFromPtr(&req.QTY),
	}

	shoppingCart, err := server.store.UpdateShoppingCartItem(ctx, arg)
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

	ctx.JSON(http.StatusOK, shoppingCart)
}

// ////////////* Delete API //////////////
type deleteShoppingCartItemUriRequest struct {
	ShoppingCartItemID int64 `uri:"shopping-cart-item-id" binding:"required,min=1"`
}

type deleteShoppingCartItemJsonRequest struct {
	UserID int64 `json:"user_id" binding:"required,min=1"`
}

func (server *Server) deleteShoppingCartItem(ctx *gin.Context) {
	var uri deleteShoppingCartItemUriRequest
	var req deleteShoppingCartItemJsonRequest

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

	arg := db.DeleteShoppingCartItemParams{
		ID:     uri.ShoppingCartItemID,
		UserID: authPayload.UserID,
	}

	err := server.store.DeleteShoppingCartItem(ctx, arg)
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

// ////////////* Delete All API //////////////

type deleteShoppingCartItemAllJsonRequest struct {
	UserID int64 `json:"user_id" binding:"required,min=1"`
}

func (server *Server) deleteShoppingCartItemAllByUser(ctx *gin.Context) {
	var req deleteShoppingCartItemAllJsonRequest

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

	_, err := server.store.DeleteShoppingCartItemAllByUser(ctx, authPayload.UserID)
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

// ////////////* Finish Purshase API //////////////

type finishPurshaseJsonRequest struct {
	UserID           int64  `json:"user_id" binding:"required,min=1"`
	UserAddressID    int64  `json:"user_address_id" binding:"required,min=1"`
	PaymentMethodID  int64  `json:"payment_method_id" binding:"required,min=1"`
	ShoppingCartID   int64  `json:"shopping_cart_id" binding:"required,min=1"`
	ShippingMethodID int64  `json:"shipping_method_id" binding:"required,min=1"`
	OrderStatusID    int64  `json:"order_status_id" binding:"required,min=1"`
	OrderTotal       string `json:"order_total" binding:"required"`
}

func (server *Server) finishPurchase(ctx *gin.Context) {
	var req finishPurshaseJsonRequest

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

	arg := db.FinishedPurchaseTxParams{
		UserID:           authPayload.UserID,
		UserAddressID:    req.UserAddressID,
		PaymentMethodID:  req.PaymentMethodID,
		ShoppingCartID:   req.ShoppingCartID,
		ShippingMethodID: req.ShippingMethodID,
		OrderStatusID:    req.OrderStatusID,
		OrderTotal:       req.OrderTotal,
	}

	finishedPurchase, err := server.store.FinishedPurchaseTx(ctx, arg)
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

	ctx.JSON(http.StatusOK, finishedPurchase)
}
