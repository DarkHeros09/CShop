package api

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v4"
)

type renewAccessTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type renewAccessTokenResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"accesss_token_expires_at"`
}

func (server *Server) renewAccessToken(ctx *fiber.Ctx) error {
	req := &renewAccessTokenRequest{}

	if err := parseAndValidate(ctx, Input{req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	refreshPayload, err := server.tokenMaker.VerifyTokenForUser(req.RefreshToken)
	if err != nil {
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
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessPayload.ExpiredAt,
	}
	ctx.Status(fiber.StatusOK).JSON(rsp)
	return nil
}
