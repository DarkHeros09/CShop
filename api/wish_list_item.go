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

type createWishListItemRequest struct {
	UserID        int64 `json:"user_id" binding:"required,min=1"`
	ProductItemID int64 `json:"product_item_id" binding:"required,min=1"`
}

func (server *Server) createWishListItem(ctx *gin.Context) {
	var req createWishListItemRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	WishList, err := server.store.GetWishListByUserID(ctx, req.UserID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.UserPayload)
	if authPayload.UserID != WishList.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	arg := db.CreateWishListItemParams{
		WishListID:    WishList.ID,
		ProductItemID: req.ProductItemID,
	}

	WishListItem, err := server.store.CreateWishListItem(ctx, arg)
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

	ctx.JSON(http.StatusOK, WishListItem)
}

//////////////* Get API //////////////

type getWishListItemUriRequest struct {
	WishListID int64 `uri:"wish-list-id" binding:"required,min=1"`
}

type getWishListItemJsonRequest struct {
	UserID int64 `json:"user_id" binding:"required,min=1"`
}

func (server *Server) getWishListItem(ctx *gin.Context) {
	var uri getWishListItemUriRequest
	var req getWishListItemJsonRequest

	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.GetWishListItemByUserIDCartIDParams{
		UserID:     req.UserID,
		WishListID: uri.WishListID,
	}

	WishListItem, err := server.store.GetWishListItemByUserIDCartID(ctx, arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.UserPayload)
	if authPayload.UserID != WishListItem.UserID.Int64 {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, WishListItem)
}

//////////////* List API //////////////

type listWishListItemsRequest struct {
	UserID int64 `json:"user_id" binding:"required,min=1"`
}

func (server *Server) listWishListItems(ctx *gin.Context) {
	var req listWishListItemsRequest

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
	WishListItems, err := server.store.ListWishListItemsByUserID(ctx, authPayload.UserID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, WishListItems)
}

// ////////////* UPDATE API //////////////
type updateWishListItemUriRequest struct {
	WishListID int64 `uri:"wish-list-id" binding:"required,min=1"`
}

type updateWishListItemJsonRequest struct {
	ID            int64 `json:"id" binding:"required,min=1"`
	UserID        int64 `json:"user_id" binding:"required,min=1"`
	ProductItemID int64 `json:"product_item_id" binding:"omitempty,required"`
}

func (server *Server) updateWishListItem(ctx *gin.Context) {
	var uri updateWishListItemUriRequest
	var req updateWishListItemJsonRequest

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

	arg := db.UpdateWishListItemParams{
		ID:            req.ID,
		WishListID:    uri.WishListID,
		ProductItemID: null.IntFromPtr(&req.ProductItemID),
	}

	WishList, err := server.store.UpdateWishListItem(ctx, arg)
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

	ctx.JSON(http.StatusOK, WishList)
}

// ////////////* Delete API //////////////
type deleteWishListItemUriRequest struct {
	WishListItemID int64 `uri:"wish-list-item-id" binding:"required,min=1"`
}

type deleteWishListItemJsonRequest struct {
	UserID int64 `json:"user_id" binding:"required,min=1"`
}

func (server *Server) deleteWishListItem(ctx *gin.Context) {
	var uri deleteWishListItemUriRequest
	var req deleteWishListItemJsonRequest

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

	arg := db.DeleteWishListItemParams{
		ID:     uri.WishListItemID,
		UserID: authPayload.UserID,
	}

	err := server.store.DeleteWishListItem(ctx, arg)
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

type deleteWishListItemAllJsonRequest struct {
	UserID int64 `json:"user_id" binding:"required,min=1"`
}

func (server *Server) deleteWishListItemAllByUser(ctx *gin.Context) {
	var req deleteWishListItemAllJsonRequest

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

	_, err := server.store.DeleteWishListItemAllByUser(ctx, authPayload.UserID)
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
