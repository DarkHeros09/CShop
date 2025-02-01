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

type createAppPolicyParamsRequest struct {
	AdminID int64 `params:"adminId" validate:"required,min=1"`
}
type createAppPolicyRequest struct {
	Policy string `json:"policy" validate:"required"`
}

func (server *Server) createAppPolicy(ctx *fiber.Ctx) error {
	params := &createAppPolicyParamsRequest{}
	req := &createAppPolicyRequest{}

	if err := server.parseAndValidate(ctx, Input{params: params, req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationAdminPayloadKey).(*token.AdminPayload)
	if authPayload.AdminID != params.AdminID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.CreateAppPolicyParams{
		AdminID: authPayload.AdminID,
		Policy:  null.StringFromPtr(&req.Policy),
	}

	appPolicy, err := server.store.CreateAppPolicy(ctx.Context(), arg)
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

	ctx.Status(fiber.StatusOK).JSON(appPolicy)
	return nil
}

// ////////////* Get API //////////////

func (server *Server) getAppPolicy(ctx *fiber.Ctx) error {

	appPolicy, err := server.store.GetAppPolicy(ctx.Context())
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(appPolicy)
	return nil
}

// ////////////* UPDATE API //////////////
type updateAppPolicyParamsRequest struct {
	AdminID int64 `params:"adminId" validate:"required,min=1"`
	ID      int64 `params:"id" validate:"required,min=1"`
}

type updateAppPolicyJsonRequest struct {
	Policy *string `json:"policy" validate:"required"`
}

func (server *Server) updateAppPolicy(ctx *fiber.Ctx) error {
	params := &updateAppPolicyParamsRequest{}
	req := &updateAppPolicyJsonRequest{}

	if err := server.parseAndValidate(ctx, Input{params: params, req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationAdminPayloadKey).(*token.AdminPayload)
	if authPayload.AdminID != params.AdminID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.UpdateAppPolicyParams{
		ID:      params.ID,
		AdminID: authPayload.AdminID,
		Policy:  null.StringFromPtr(req.Policy),
	}

	appPolicy, err := server.store.UpdateAppPolicy(ctx.Context(), arg)
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

	ctx.Status(fiber.StatusOK).JSON(appPolicy)
	return nil
}

// ////////////* Delete API //////////////
type deleteAppPolicyParamsRequest struct {
	AdminID int64 `params:"adminId" validate:"required,min=1"`
	ID      int64 `params:"id" validate:"required,min=1"`
}

func (server *Server) deleteAppPolicy(ctx *fiber.Ctx) error {
	params := &deleteAppPolicyParamsRequest{}

	if err := server.parseAndValidate(ctx, Input{params: params}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationAdminPayloadKey).(*token.AdminPayload)
	if authPayload.AdminID != params.AdminID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.DeleteAppPolicyParams{
		AdminID: authPayload.AdminID,
		ID:      params.ID,
	}

	_, err := server.store.DeleteAppPolicy(ctx.Context(), arg)
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
