package api

import (
	"errors"

	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/token"
	"github.com/gofiber/fiber/v2"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
)

//////////////* Create API //////////////

type createNotificationParamsRequest struct {
	UserID int64 `params:"id" validate:"required,min=1"`
}
type createNotificationRequest struct {
	DeviceID        string `json:"device_id" validate:"required"`
	FcmToken        string `json:"fcm_token" validate:"required"`
	DeliveryUpdates bool   `json:"delivery_updates" validate:"required"`
}

func (server *Server) createNotification(ctx *fiber.Ctx) error {
	params := &createNotificationParamsRequest{}
	req := &createNotificationRequest{}

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

	arg := db.CreateNotificationParams{
		UserID:          authPayload.UserID,
		DeviceID:        null.StringFromPtr(&req.DeviceID),
		FcmToken:        null.StringFromPtr(&req.FcmToken),
		DeliveryUpdates: req.DeliveryUpdates,
	}

	notification, err := server.store.CreateNotification(ctx.Context(), arg)
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

	ctx.Status(fiber.StatusOK).JSON(notification)
	return nil
}

// ////////////* Get API //////////////
type getNotificationParamsRequest struct {
	UserID   int64  `params:"id" validate:"required,min=1"`
	DeviceID string `params:"deviceId" validate:"required"`
}

func (server *Server) getNotification(ctx *fiber.Ctx) error {
	params := &getNotificationParamsRequest{}

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

	arg := db.GetNotificationParams{
		UserID:   params.UserID,
		DeviceID: null.StringFromPtr(&params.DeviceID),
	}

	notification, err := server.store.GetNotification(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(notification)
	return nil
}

// ////////////* UPDATE API //////////////
type updateNotificationParamsRequest struct {
	UserID   int64  `params:"id" validate:"required,min=1"`
	DeviceID string `params:"deviceId" validate:"required"`
}

type updateNotificationJsonRequest struct {
	FcmToken        *string `json:"fcm_token" validate:"required"`
	DeliveryUpdates *bool   `json:"delivery_updates" validate:"required"`
}

func (server *Server) updateNotification(ctx *fiber.Ctx) error {
	params := &updateNotificationParamsRequest{}
	req := &updateNotificationJsonRequest{}

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

	arg := db.UpdateNotificationParams{
		FcmToken:        null.StringFromPtr(req.FcmToken),
		UserID:          authPayload.UserID,
		DeviceID:        null.StringFromPtr(&params.DeviceID),
		DeliveryUpdates: null.BoolFromPtr(req.DeliveryUpdates),
	}

	notification, err := server.store.UpdateNotification(ctx.Context(), arg)
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

	ctx.Status(fiber.StatusOK).JSON(notification)
	return nil
}

// ////////////* Delete API //////////////
type deleteNotificationParamsRequest struct {
	UserID   int64  `params:"id" validate:"required,min=1"`
	DeviceID string `params:"deviceId" validate:"required"`
}

func (server *Server) deleteNotification(ctx *fiber.Ctx) error {
	params := &deleteNotificationParamsRequest{}

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

	arg := db.DeleteNotificationParams{
		UserID:   authPayload.UserID,
		DeviceID: null.StringFromPtr(&params.DeviceID),
	}

	_, err := server.store.DeleteNotification(ctx.Context(), arg)
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

// ////////////* Delete All API //////////////

type deleteNotificationAllParamsRequest struct {
	UserID int64 `params:"id" validate:"required,min=1"`
}

func (server *Server) deleteNotificationAllByUser(ctx *fiber.Ctx) error {
	params := &deleteNotificationAllParamsRequest{}

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

	err := server.store.DeleteNotificationAllByUser(ctx.Context(), authPayload.UserID)
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
