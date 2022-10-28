package api

import (
	"database/sql"
	"errors"
	"net/http"

	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/token"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"gopkg.in/guregu/null.v4"
)

type createUserAddressRequest struct {
	AddressLine    string   `json:"address_line" binding:"required"`
	Region         string   `json:"region" binding:"required"`
	City           string   `json:"city" binding:"required"`
	DefaultAddress null.Int `json:"default_address" binding:"omitempty,required"`
}

type userAddressResponse struct {
	UserID         int64  `json:"user_id"`
	AddressID      int64  `json:"address_id"`
	DefaultAddress int64  `json:"default_address"`
	AddressLine    string `json:"address_line"`
	Region         string `json:"region"`
	City           string `json:"city"`
}

func newUserAddressResponseForCreate(address db.CreateUserAddressWithAddressRow) userAddressResponse {
	return userAddressResponse{
		UserID:         address.UserID,
		AddressID:      address.AddressID,
		DefaultAddress: address.DefaultAddress.Int64,
		AddressLine:    address.AddressLine,
		Region:         address.Region,
		City:           address.City,
	}
}

func (server *Server) createUserAddress(ctx *gin.Context) {
	var req createUserAddressRequest

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.UserPayload)

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg1 := db.CreateUserAddressWithAddressParams{
		UserID:         authPayload.UserID,
		AddressLine:    req.AddressLine,
		Region:         req.Region,
		City:           req.City,
		DefaultAddress: req.DefaultAddress.NullInt64,
	}

	userAddress, err := server.store.CreateUserAddressWithAddress(ctx, arg1)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "foreign_key_violation", "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := newUserAddressResponseForCreate(userAddress)

	ctx.JSON(http.StatusOK, rsp)
}

func newUserAddressResponseForGet(address db.GetUserAddressWithAddressRow) userAddressResponse {
	return userAddressResponse{
		UserID:         address.UserID,
		AddressID:      address.AddressID,
		DefaultAddress: address.DefaultAddress.Int64,
		AddressLine:    address.AddressLine,
		Region:         address.Region,
		City:           address.City,
	}
}

type getUserAddressRequest struct {
	AddressID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getUserAddress(ctx *gin.Context) {
	var req getUserAddressRequest

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.UserPayload)

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	arg := db.GetUserAddressWithAddressParams{
		UserID:    authPayload.UserID,
		AddressID: req.AddressID,
	}
	userAddress, err := server.store.GetUserAddressWithAddress(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if userAddress.UserID != authPayload.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	rsp := newUserAddressResponseForGet(userAddress)

	ctx.JSON(http.StatusOK, rsp)
}

type listUserAddressesRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server) listUserAddresses(ctx *gin.Context) {
	var req listUserAddressesRequest

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.UserPayload)

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	arg := db.ListUserAddressesParams{
		UserID: authPayload.UserID,
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}
	userAddresses, err := server.store.ListUserAddresses(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, userAddresses)
}

func newUserAddressResponseForUpdate(address db.Address, userAddress db.UserAddress) userAddressResponse {
	return userAddressResponse{
		UserID:         userAddress.UserID,
		AddressID:      userAddress.AddressID,
		DefaultAddress: userAddress.DefaultAddress.Int64,
		AddressLine:    address.AddressLine,
		Region:         address.Region,
		City:           address.City,
	}
}

type updateUserAddressRequest struct {
	UserID         int64  `uri:"user_id" binding:"required,min=1"`
	AddressID      int64  `json:"address_id" binding:"required,min=1"`
	AddressLine    string `json:"address_line" binding:"omitempty,required"`
	City           string `json:"city" binding:"omitempty,required"`
	Region         string `json:"region" binding:"omitempty,required"`
	DefaultAddress int64  `json:"default_address" binding:"omitempty,required,min=1"`
}

func (server *Server) updateUserAddress(ctx *gin.Context) {
	var req updateUserAddressRequest

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.UserPayload)

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	arg1 := db.UpdateUserAddressParams{
		UserID:    authPayload.UserID,
		AddressID: req.AddressID,
		DefaultAddress: sql.NullInt64{
			Int64: req.DefaultAddress,
			Valid: req.DefaultAddress != 0,
		},
	}

	userAddress, err := server.store.UpdateUserAddress(ctx, arg1)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if userAddress.UserID != authPayload.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	arg2 := db.UpdateAddressParams{
		AddressLine: sql.NullString{
			String: req.AddressLine,
			Valid:  req.AddressLine != "",
		},
		Region: sql.NullString{
			String: req.Region,
			Valid:  req.Region != "",
		},
		City: sql.NullString{
			String: req.City,
			Valid:  req.City != "",
		},
		ID: userAddress.AddressID,
	}

	address, err := server.store.UpdateAddress(ctx, arg2)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := newUserAddressResponseForUpdate(address, userAddress)

	ctx.JSON(http.StatusOK, rsp)
}

type deleteUserAddressRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) deleteUserAddress(ctx *gin.Context) {
	var req deleteUserAddressRequest

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.UserPayload)

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg1 := db.DeleteUserAddressParams{
		UserID:    authPayload.UserID,
		AddressID: req.ID,
	}

	err := server.store.DeleteUserAddress(ctx, arg1)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "foreign_key_violation", "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		} else if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}
