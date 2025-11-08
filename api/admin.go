package api

import (
	"errors"
	"time"

	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/token"
	"github.com/cshop/v3/util"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
)

type adminResponse struct {
	AdminID  int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	TypeID   int64  `json:"type_id"`
}

func newAdminResponse(admin db.Admin) adminResponse {
	return adminResponse{
		AdminID:  admin.ID,
		Username: admin.Username,
		Email:    admin.Email,
		TypeID:   admin.TypeID,
	}
}

// //////////////* Login API //////////////

type loginAdminRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type loginAdminResponse struct {
	AdminSessionID        string        `json:"admin_session_id"`
	AccessToken           string        `json:"access_token"`
	AccessTokenExpiresAt  time.Time     `json:"access_token_expires_at"`
	RefreshToken          string        `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time     `json:"refresh_token_expires_at"`
	Admin                 adminResponse `json:"admin"`
}

func (server *Server) loginAdmin(ctx *fiber.Ctx) error {
	req := &loginAdminRequest{}

	if err := server.parseAndValidate(ctx, Input{req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	admin, err := server.store.GetAdminByEmail(ctx.Context(), req.Email)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	if admin == nil {
		ctx.Status(fiber.StatusNotFound).JSON(pgx.ErrNoRows)
		return nil
	}

	if !admin.Active {
		err := errors.New("account unauthorized")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	err = util.CheckPassword(req.Password, admin.Password)
	if err != nil {
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	accessToken, accessPayload, err := server.adminTokenMaker.CreateTokenForAdmin(
		admin.ID,
		admin.Username,
		admin.TypeID,
		admin.Active,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	refreshToken, refreshPayload, err := server.adminTokenMaker.CreateTokenForAdmin(
		admin.ID,
		admin.Username,
		admin.TypeID,
		admin.Active,
		server.config.RefreshTokenDuration,
	)
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	arg := db.CreateAdminSessionParams{
		ID:           refreshPayload.ID,
		AdminID:      admin.ID,
		RefreshToken: refreshToken,
		AdminAgent:   string(ctx.Context().UserAgent()),
		ClientIp:     ctx.IP(),
		ExpiresAt:    refreshPayload.ExpiredAt,
	}

	adminSession, err := server.store.CreateAdminSession(ctx.Context(), arg)
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	rsp := loginAdminResponse{
		AdminSessionID:        adminSession.ID.String(),
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
		Admin:                 newAdminResponse(*admin),
	}
	ctx.Status(fiber.StatusOK).JSON(rsp)
	return nil
}

// //////////////* Logout API //////////////

type logoutAdminParamsRequest struct {
	AdminID int64 `params:"id" validate:"required,min=1"`
}

type logoutAdminJsonRequest struct {
	AdminSessionID string `json:"admin_session_id" validate:"required"`
	RefreshToken   string `json:"refresh_token" validate:"required"`
}

func (server *Server) logoutAdmin(ctx *fiber.Ctx) error {
	params := &logoutAdminParamsRequest{}
	req := &logoutAdminJsonRequest{}

	if err := server.parseAndValidate(ctx, Input{params: params, req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}
	adminSessionID, err := uuid.Parse(req.AdminSessionID)
	if err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationAdminPayloadKey).(*token.AdminPayload)
	if params.AdminID != authPayload.AdminID {
		err := errors.New("account doesn't belong to the authenticated admin")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.UpdateAdminSessionParams{
		ID:           adminSessionID,
		AdminID:      authPayload.AdminID,
		RefreshToken: req.RefreshToken,
		IsBlocked:    null.BoolFrom(true),
	}

	_, err = server.store.UpdateAdminSession(ctx.Context(), arg)
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

	ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
	return nil
}
