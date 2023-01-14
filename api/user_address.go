package api

import (
	"errors"

	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/token"
	"github.com/cshop/v3/util"
	"github.com/gofiber/fiber/v2"
	"github.com/guregu/null"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

//////////////* Create API //////////////

type createUserAddressParamsRequest struct {
	UserID int64 `params:"id" validate:"required,min=1"`
}
type createUserAddressJsonRequest struct {
	AddressLine    string   `json:"address_line" validate:"required"`
	Region         string   `json:"region" validate:"required"`
	City           string   `json:"city" validate:"required"`
	DefaultAddress null.Int `json:"default_address" validate:"omitempty,required"`
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

func (server *Server) createUserAddress(ctx *fiber.Ctx) error {
	var params createUserAddressParamsRequest
	var req createUserAddressJsonRequest

	if err := ctx.ParamsParser(&params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := util.ValidateStruct(params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := ctx.BodyParser(&req); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := util.ValidateStruct(req); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.CreateUserAddressWithAddressParams{
		UserID:         authPayload.UserID,
		AddressLine:    req.AddressLine,
		Region:         req.Region,
		City:           req.City,
		DefaultAddress: req.DefaultAddress,
	}

	userAddress, err := server.store.CreateUserAddressWithAddress(ctx.Context(), arg)
	if err != nil {
		if pqErr, ok := err.(*pgconn.PgError); ok {
			switch pqErr.Message {
			case "foreign_key_violation", "unique_violation":
				ctx.Status(fiber.StatusForbidden).JSON(errorResponse(err))
				return nil
			}
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	rsp := newUserAddressResponseForCreate(userAddress)

	ctx.Status(fiber.StatusOK).JSON(rsp)
	return nil
}

//////////////* Get API //////////////

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

type getUserAddressParamsRequest struct {
	UserID    int64 `params:"id" validate:"required,min=1"`
	AddressID int64 `params:"addressId" validate:"required,min=1"`
}

func (server *Server) getUserAddress(ctx *fiber.Ctx) error {
	var params getUserAddressParamsRequest

	if err := ctx.ParamsParser(&params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := util.ValidateStruct(params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.GetUserAddressWithAddressParams{
		UserID:    authPayload.UserID,
		AddressID: params.AddressID,
	}
	userAddress, err := server.store.GetUserAddressWithAddress(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	rsp := newUserAddressResponseForGet(userAddress)

	ctx.Status(fiber.StatusOK).JSON(rsp)
	return nil
}

//////////////* List API //////////////

type listUserAddressParamsRequest struct {
	UserID int64 `params:"id" validate:"required,min=1"`
}
type listUserAddressesQueryRequest struct {
	PageID   int32 `query:"page_id" validate:"required,min=1"`
	PageSize int32 `query:"page_size" validate:"required,min=5,max=10"`
}

func (server *Server) listUserAddresses(ctx *fiber.Ctx) error {
	var params listUserAddressParamsRequest
	var query listUserAddressesQueryRequest

	if err := ctx.ParamsParser(&params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := util.ValidateStruct(params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := ctx.QueryParser(&query); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := util.ValidateStruct(query); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.ListUserAddressesParams{
		UserID: authPayload.UserID,
		Limit:  query.PageSize,
		Offset: (query.PageID - 1) * query.PageSize,
	}
	userAddresses, err := server.store.ListUserAddresses(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}
	ctx.Status(fiber.StatusOK).JSON(userAddresses)
	return nil
}

// ////////////* UPDATE API //////////////
type updateUserAddressParamsRequest struct {
	UserID    int64 `params:"id" validate:"required,min=1"`
	AddressID int64 `params:"addressId" validate:"required,min=1"`
}

type updateUserAddressJsonRequest struct {
	AddressLine    string `json:"address_line" validate:"omitempty,required"`
	City           string `json:"city" validate:"omitempty,required"`
	Region         string `json:"region" validate:"omitempty,required"`
	DefaultAddress int64  `json:"default_address" validate:"omitempty,required,min=1"`
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

func (server *Server) updateUserAddress(ctx *fiber.Ctx) error {
	var params updateUserAddressParamsRequest
	var req updateUserAddressJsonRequest

	if err := ctx.ParamsParser(&params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := util.ValidateStruct(params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := ctx.BodyParser(&req); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := util.ValidateStruct(req); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg1 := db.UpdateUserAddressParams{
		UserID:         authPayload.UserID,
		AddressID:      params.AddressID,
		DefaultAddress: null.IntFromPtr(&req.DefaultAddress),
	}

	userAddress, err := server.store.UpdateUserAddress(ctx.Context(), arg1)
	if err != nil {
		if pqErr, ok := err.(*pgconn.PgError); ok {
			switch pqErr.Message {
			case "foreign_key_violation", "unique_violation":
				ctx.Status(fiber.StatusForbidden).JSON(errorResponse(err))
				return nil
			}
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	arg2 := db.UpdateAddressParams{
		AddressLine: null.StringFromPtr(&req.AddressLine),
		Region:      null.StringFromPtr(&req.Region),
		City:        null.StringFromPtr(&req.City),
		ID:          userAddress.AddressID,
	}

	address, err := server.store.UpdateAddress(ctx.Context(), arg2)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	rsp := newUserAddressResponseForUpdate(address, userAddress)

	ctx.Status(fiber.StatusOK).JSON(rsp)
	return nil
}

// ////////////* Delete API //////////////
type deleteUserAddressParamsRequest struct {
	UserID    int64 `params:"id" validate:"required,min=1"`
	AddressID int64 `params:"addressId" validate:"required,min=1"`
}

func (server *Server) deleteUserAddress(ctx *fiber.Ctx) error {
	var params deleteUserAddressParamsRequest

	if err := ctx.ParamsParser(&params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	if err := util.ValidateStruct(params); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.DeleteUserAddressParams{
		UserID:    authPayload.UserID,
		AddressID: params.AddressID,
	}

	_, err := server.store.DeleteUserAddress(ctx.Context(), arg)
	if err != nil {
		if pqErr, ok := err.(*pgconn.PgError); ok {
			switch pqErr.Message {
			case "foreign_key_violation", "unique_violation":
				ctx.Status(fiber.StatusForbidden).JSON(errorResponse(err))
				return nil
			}
		} else if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
	return nil
}
