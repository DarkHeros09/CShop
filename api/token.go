package api

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

type renewAccessTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type renewAccessTokenResponse struct {
	UserSessionID        string    `json:"user_session_id"`
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}

func (server *Server) renewAccessToken(ctx *fiber.Ctx) error {
	req := &renewAccessTokenRequest{}

	if err := parseAndValidate(ctx, Input{req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	refreshPayload, err := server.tokenMaker.VerifyTokenForUser(req.RefreshToken)
	if err != nil {
		if err.Error() == "token has expired" {
			err = fmt.Errorf("refresh token has expired")
		}
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	userSession, err := server.store.GetUserSession(ctx.Context(), refreshPayload.ID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	if userSession.IsBlocked {
		err := fmt.Errorf("blocked session")
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	if userSession.UserID != refreshPayload.UserID {
		err := fmt.Errorf("incorrect session user")
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	if userSession.RefreshToken != req.RefreshToken {
		err := fmt.Errorf("mismatched session token")
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	if time.Now().After(userSession.ExpiresAt) {
		err := fmt.Errorf("expired session")
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	accessToken, accessPayload, err := server.tokenMaker.CreateTokenForUser(
		refreshPayload.UserID,
		refreshPayload.Username,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	rsp := renewAccessTokenResponse{
		UserSessionID:        userSession.ID.String(),
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessPayload.ExpiredAt,
	}
	ctx.Status(fiber.StatusOK).JSON(rsp)
	return nil
}

//* Admin Access Token

type renewAccessTokenAdminRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type renewAccessTokenAdminResponse struct {
	AdminSessionID       string    `json:"admin_session_id"`
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}

func (server *Server) renewAccessTokenAdmin(ctx *fiber.Ctx) error {
	req := &renewAccessTokenAdminRequest{}

	if err := parseAndValidate(ctx, Input{req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	refreshPayload, err := server.tokenMaker.VerifyTokenForAdmin(req.RefreshToken)
	if err != nil {
		if err.Error() == "token has expired" {
			err = fmt.Errorf("refresh token has expired")
		}
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	userSession, err := server.store.GetAdminSession(ctx.Context(), refreshPayload.ID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	if userSession.IsBlocked {
		err := fmt.Errorf("blocked session")
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	if userSession.AdminID != refreshPayload.AdminID {
		err := fmt.Errorf("incorrect session user")
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	if userSession.RefreshToken != req.RefreshToken {
		err := fmt.Errorf("mismatched session token")
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	if time.Now().After(userSession.ExpiresAt) {
		err := fmt.Errorf("expired session")
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	accessToken, accessPayload, err := server.tokenMaker.CreateTokenForAdmin(
		refreshPayload.AdminID,
		refreshPayload.Username,
		refreshPayload.TypeID,
		refreshPayload.Active,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	rsp := renewAccessTokenAdminResponse{
		AdminSessionID:       userSession.ID.String(),
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessPayload.ExpiredAt,
	}
	ctx.Status(fiber.StatusOK).JSON(rsp)
	return nil
}
