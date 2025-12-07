package api

import (
	"errors"
	"time"

	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/util"
	"github.com/gofiber/fiber/v2"
	"github.com/guregu/null/v6"
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

	if err := server.parseAndValidate(ctx, Input{req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	refreshPayload, err := server.userTokenMaker.VerifyTokenForUser(req.RefreshToken)
	if err != nil {
		if err.Error() == util.TokenHasExpired {
			err = errors.New("refresh token has expired")

			userSession, _ := server.store.GetUserSession(ctx.Context(), refreshPayload.ID)

			argUpdate := db.UpdateUserSessionParams{
				IsBlocked:    null.BoolFrom(true),
				ID:           userSession.ID,
				UserID:       userSession.UserID,
				RefreshToken: userSession.RefreshToken,
			}

			server.store.UpdateUserSession(ctx.Context(), argUpdate)

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
		err := errors.New("blocked session")
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	if userSession.UserID != refreshPayload.UserID {
		err := errors.New("incorrect session user")
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	if userSession.RefreshToken != req.RefreshToken {
		err := errors.New("mismatched session token")
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	// if time.Now().After(userSession.ExpiresAt) {
	// 	err := errors.New("expired session")
	// 	ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	// 	return nil
	// }

	accessToken, accessPayload, err := server.userTokenMaker.CreateTokenForUser(
		refreshPayload.UserID,
		refreshPayload.Username,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		println("err: ", err)
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

type renewRefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type renewRefreshTokenResponse struct {
	UserSessionID         string    `json:"user_session_id"`
	AccessToken           string    `json:"access_token"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshToken          string    `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
}

func (server *Server) renewRefreshToken(ctx *fiber.Ctx) error {
	req := &renewRefreshTokenRequest{}

	if err := server.parseAndValidate(ctx, Input{req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	refreshPayload, err := server.userTokenMaker.VerifyTokenForUser(req.RefreshToken)
	if err != nil {
		if err.Error() == util.TokenHasExpired {
			err = errors.New("refresh token has expired")

			userSession, _ := server.store.GetUserSession(ctx.Context(), refreshPayload.ID)

			argUpdate := db.UpdateUserSessionParams{
				IsBlocked:    null.BoolFrom(true),
				ID:           userSession.ID,
				UserID:       userSession.UserID,
				RefreshToken: userSession.RefreshToken,
			}

			server.store.UpdateUserSession(ctx.Context(), argUpdate)
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
		err := errors.New("blocked session")
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	if userSession.UserID != refreshPayload.UserID {
		err := errors.New("incorrect session user")
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	if userSession.RefreshToken != req.RefreshToken {
		err := errors.New("mismatched session token")
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	// if time.Now().After(userSession.ExpiresAt) {
	// 	err := errors.New("expired session")
	// 	ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	// 	return nil
	// }

	argUpdate := db.UpdateUserSessionParams{
		IsBlocked:    null.BoolFrom(true),
		ID:           userSession.ID,
		UserID:       userSession.UserID,
		RefreshToken: userSession.RefreshToken,
	}

	_, err = server.store.UpdateUserSession(ctx.Context(), argUpdate)
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	accessToken, accessPayload, err := server.userTokenMaker.CreateTokenForUser(
		refreshPayload.UserID,
		refreshPayload.Username,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	newRefreshToken, newRefreshPayload, err := server.userTokenMaker.CreateTokenForUser(
		refreshPayload.UserID,
		refreshPayload.Username,
		server.config.RefreshTokenDuration,
	)
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	argCreate := db.CreateUserSessionParams{
		ID:           newRefreshPayload.ID,
		UserID:       userSession.UserID,
		RefreshToken: newRefreshToken,
		UserAgent:    string(ctx.Context().UserAgent()),
		ClientIp:     ctx.IP(),
		ExpiresAt:    newRefreshPayload.ExpiredAt,
	}

	newUserSession, err := server.store.CreateUserSession(ctx.Context(), argCreate)
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	rsp := renewRefreshTokenResponse{
		UserSessionID:         newUserSession.ID.String(),
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          newRefreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
	}
	ctx.Status(fiber.StatusOK).JSON(rsp)
	return nil
}

//* Admin Access Token

type renewAccessTokenForAdminRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type renewAccessTokenForAdminResponse struct {
	AdminSessionID       string    `json:"admin_session_id"`
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}

func (server *Server) renewAccessTokenForAdmin(ctx *fiber.Ctx) error {
	req := &renewAccessTokenForAdminRequest{}

	if err := server.parseAndValidate(ctx, Input{req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	refreshPayload, err := server.adminTokenMaker.VerifyTokenForAdmin(req.RefreshToken)
	if err != nil {
		if err.Error() == util.TokenHasExpired {
			err = errors.New("refresh token has expired")
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
		err := errors.New("blocked session")
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	if userSession.AdminID != refreshPayload.AdminID {
		err := errors.New("incorrect session user")
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	if userSession.RefreshToken != req.RefreshToken {
		err := errors.New("mismatched session token")
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	if time.Now().After(userSession.ExpiresAt) {
		err := errors.New("expired session")
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	accessToken, accessPayload, err := server.adminTokenMaker.CreateTokenForAdmin(
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

	rsp := renewAccessTokenForAdminResponse{
		AdminSessionID:       userSession.ID.String(),
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessPayload.ExpiredAt,
	}
	ctx.Status(fiber.StatusOK).JSON(rsp)
	return nil
}

type renewRefreshTokenForAdminRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type renewRefreshTokenForAdminResponse struct {
	AdminSessionID        string    `json:"admin_session_id"`
	AccessToken           string    `json:"access_token"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshToken          string    `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
}

func (server *Server) renewRefreshTokenForAdmin(ctx *fiber.Ctx) error {
	req := &renewRefreshTokenForAdminRequest{}

	if err := server.parseAndValidate(ctx, Input{req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	refreshPayload, err := server.adminTokenMaker.VerifyTokenForAdmin(req.RefreshToken)
	if err != nil {
		if err.Error() == util.TokenHasExpired {
			err = errors.New("refresh token has expired")
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
		err := errors.New("blocked session")
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	if userSession.AdminID != refreshPayload.AdminID {
		err := errors.New("incorrect session user")
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	if userSession.RefreshToken != req.RefreshToken {
		err := errors.New("mismatched session token")
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	if time.Now().After(userSession.ExpiresAt) {
		err := errors.New("expired session")
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	argUpdate := db.UpdateAdminSessionParams{
		IsBlocked:    null.BoolFrom(true),
		ID:           userSession.ID,
		AdminID:      userSession.AdminID,
		RefreshToken: userSession.RefreshToken,
	}

	_, err = server.store.UpdateAdminSession(ctx.Context(), argUpdate)
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	accessToken, accessPayload, err := server.adminTokenMaker.CreateTokenForAdmin(
		refreshPayload.AdminID,
		refreshPayload.Username,
		refreshPayload.TypeID,
		refreshPayload.Active,
		server.config.RefreshTokenDuration,
	)
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	newRefreshToken, newRefreshPayload, err := server.adminTokenMaker.CreateTokenForAdmin(
		refreshPayload.AdminID,
		refreshPayload.Username,
		refreshPayload.TypeID,
		refreshPayload.Active,
		server.config.RefreshTokenDuration,
	)
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	argCreate := db.CreateAdminSessionParams{
		ID:           newRefreshPayload.ID,
		AdminID:      userSession.AdminID,
		RefreshToken: newRefreshToken,
		AdminAgent:   string(ctx.Context().UserAgent()),
		ClientIp:     ctx.IP(),
		ExpiresAt:    newRefreshPayload.ExpiredAt,
	}

	newUserSession, err := server.store.CreateAdminSession(ctx.Context(), argCreate)
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	rsp := renewRefreshTokenForAdminResponse{
		AdminSessionID:        newUserSession.ID.String(),
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          newRefreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
	}
	ctx.Status(fiber.StatusOK).JSON(rsp)
	return nil
}
