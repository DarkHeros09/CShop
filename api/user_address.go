package api

import (
	"errors"

	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/token"
	"github.com/cshop/v3/util"
	"github.com/gofiber/fiber/v2"
	"github.com/guregu/null/v6"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
)

//////////////* Create API //////////////

type createUserAddressParamsRequest struct {
	UserID int64 `params:"id" validate:"required,min=1"`
}
type createUserAddressJsonRequest struct {
	Name        string `json:"name" validate:"required"`
	Telephone   string `json:"telephone" validate:"required,custom_phone_number"`
	AddressLine string `json:"address_line" validate:"required"`
	Region      string `json:"region" validate:"required"`
	City        string `json:"city" validate:"required"`
}

type userAddressResponse struct {
	ID               int64  `json:"id"`
	UserID           int64  `json:"user_id"`
	DefaultAddressID int64  `json:"default_address_id"`
	Telephone        string `json:"telephone"`
	Name             string `json:"name"`
	AddressLine      string `json:"address_line"`
	Region           string `json:"region"`
	City             string `json:"city"`
}

func newUserAddressResponseForCreate(address *db.Address, defualtAddressId int64) userAddressResponse {
	return userAddressResponse{
		ID:               address.ID,
		UserID:           address.UserID,
		Telephone:        address.Telephone,
		Name:             address.Name,
		AddressLine:      address.AddressLine,
		Region:           address.Region,
		City:             address.City,
		DefaultAddressID: defualtAddressId,
	}
}

func (server *Server) createUserAddress(ctx *fiber.Ctx) error {
	params := &createUserAddressParamsRequest{}
	req := &createUserAddressJsonRequest{}

	if err := server.parseAndValidate(ctx, Input{params: params, req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationUserPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	user, err := server.store.GetUser(ctx.Context(), authPayload.UserID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	arg1 := db.CreateAddressParams{
		UserID:      user.ID,
		Name:        req.Name,
		Telephone:   req.Telephone,
		AddressLine: req.AddressLine,
		Region:      req.Region,
		City:        req.City,
	}

	address, err := server.store.CreateAddress(ctx.Context(), arg1)
	if err != nil {
		if pqErr, ok := err.(*pgconn.PgError); ok {
			switch pqErr.Message {
			case util.ForeignKeyViolation, util.UniqueViolation:
				ctx.Status(fiber.StatusForbidden).JSON(errorResponse(err))
				return nil
			}
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	var defaultAddressId int64

	if user.DefaultAddressID.Valid && user.DefaultAddressID.Int64 != 0 {
		defaultAddressId = user.DefaultAddressID.Int64
	} else {
		arg2 := db.UpdateUserParams{
			ID:               authPayload.UserID,
			DefaultAddressID: null.IntFromPtr(&address.ID),
		}

		userAddress, err := server.store.UpdateUser(ctx.Context(), arg2)
		if err != nil {
			if pqErr, ok := err.(*pgconn.PgError); ok {
				switch pqErr.Message {
				case util.ForeignKeyViolation, util.UniqueViolation:
					ctx.Status(fiber.StatusForbidden).JSON(errorResponse(err))
					return nil
				}
			}
			ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
			return nil
		}

		defaultAddressId = userAddress.DefaultAddressID.Int64
	}

	rsp := newUserAddressResponseForCreate(address, defaultAddressId)

	ctx.Status(fiber.StatusOK).JSON(rsp)
	return nil
}

//////////////* Get API //////////////

func newUserAddressResponseForGet(address *db.GetUserAddressRow) userAddressResponse {
	return userAddressResponse{
		ID:               address.ID,
		UserID:           address.UserID,
		DefaultAddressID: address.DefaultAddressID.Int64,
		AddressLine:      address.AddressLine,
		Region:           address.Region,
		City:             address.City,
	}
}

type getUserAddressParamsRequest struct {
	UserID    int64 `params:"id" validate:"required,min=1"`
	AddressID int64 `params:"addressId" validate:"required,min=1"`
}

func (server *Server) getUserAddress(ctx *fiber.Ctx) error {
	params := &getUserAddressParamsRequest{}

	if err := server.parseAndValidate(ctx, Input{params: params}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationUserPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.GetUserAddressParams{
		UserID: authPayload.UserID,
		ID:     params.AddressID,
	}
	userAddress, err := server.store.GetUserAddress(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	if userAddress == nil {
		ctx.Status(fiber.StatusNotFound).JSON(pgx.ErrNoRows)
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
	params := &listUserAddressParamsRequest{}
	query := &listUserAddressesQueryRequest{}

	if err := server.parseAndValidate(ctx, Input{params: params, query: query}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationUserPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	userAddresses, err := server.store.ListAddressesByUserID(ctx.Context(), authPayload.UserID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	if userAddresses == nil {
		ctx.Status(fiber.StatusNotFound).JSON(pgx.ErrNoRows)
		return nil
	}

	// rsp := newlistUserAddressResponse(userAddresses, addresses)
	ctx.Status(fiber.StatusOK).JSON(userAddresses)
	return nil
}

// ////////////* UPDATE API //////////////
type updateUserAddressParamsRequest struct {
	UserID    int64 `params:"id" validate:"required,min=1"`
	AddressID int64 `params:"addressId" validate:"required,min=1"`
}

type updateUserAddressJsonRequest struct {
	Name             *string `json:"name" validate:"omitempty,required"`
	Telephone        *string `json:"telephone" validate:"omitempty,required"`
	AddressLine      *string `json:"address_line" validate:"omitempty,required"`
	City             *string `json:"city" validate:"omitempty,required"`
	Region           *string `json:"region" validate:"omitempty,required"`
	DefaultAddressID *int64  `json:"default_address_id" validate:"omitempty,required"`
}

func newUserAddressResponseForUpdate(user *db.User, address *db.Address) userAddressResponse {
	return userAddressResponse{
		ID:               address.ID,
		UserID:           address.UserID,
		Name:             address.Name,
		Telephone:        address.Telephone,
		DefaultAddressID: user.DefaultAddressID.Int64,
		AddressLine:      address.AddressLine,
		Region:           address.Region,
		City:             address.City,
	}
}

func (server *Server) updateUserAddress(ctx *fiber.Ctx) error {
	params := &updateUserAddressParamsRequest{}
	req := &updateUserAddressJsonRequest{}

	if err := server.parseAndValidate(ctx, Input{params: params, req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationUserPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg1 := db.UpdateUserParams{
		ID:               authPayload.UserID,
		DefaultAddressID: null.IntFromPtr(req.DefaultAddressID),
	}

	user, err := server.store.UpdateUser(ctx.Context(), arg1)
	if err != nil {
		if pqErr, ok := err.(*pgconn.PgError); ok {
			switch pqErr.Message {
			case util.ForeignKeyViolation, util.UniqueViolation:
				ctx.Status(fiber.StatusForbidden).JSON(errorResponse(err))
				return nil
			}
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	arg2 := db.UpdateAddressParams{
		ID:          params.AddressID,
		UserID:      authPayload.UserID,
		Name:        null.StringFromPtr(req.Name),
		Telephone:   null.StringFromPtr(req.Telephone),
		AddressLine: null.StringFromPtr(req.AddressLine),
		Region:      null.StringFromPtr(req.Region),
		City:        null.StringFromPtr(req.City),
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

	rsp := newUserAddressResponseForUpdate(user, address)

	ctx.Status(fiber.StatusOK).JSON(rsp)
	return nil
}

// ////////////* Delete API //////////////
type deleteUserAddressParamsRequest struct {
	UserID    int64 `params:"id" validate:"required,min=1"`
	AddressID int64 `params:"addressId" validate:"required,min=1"`
}

func (server *Server) deleteUserAddress(ctx *fiber.Ctx) error {
	params := &deleteUserAddressParamsRequest{}

	if err := server.parseAndValidate(ctx, Input{params: params}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationUserPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.DeleteUserAddressParams{
		UserID: authPayload.UserID,
		ID:     params.AddressID,
	}

	_, err := server.store.DeleteUserAddress(ctx.Context(), arg)
	if err != nil {
		if pqErr, ok := err.(*pgconn.PgError); ok {
			switch pqErr.Message {
			case util.ForeignKeyViolation, util.UniqueViolation:
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
