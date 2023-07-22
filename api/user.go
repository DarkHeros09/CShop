package api

import (
	"errors"
	"time"

	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/token"
	"github.com/cshop/v3/util"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
)

//////////////* Create API //////////////

type createUserRequest struct {
	//alphanum
	Username  string `json:"username" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=6"`
	Telephone int32  `json:"telephone" validate:"required,numeric,min=910000000,max=929999999"`
}

type userResponse struct {
	UserID         int64  `json:"id"`
	Username       string `json:"username"`
	Email          string `json:"email"`
	Telephone      int32  `json:"telephone"`
	ShoppingCartID int64  `json:"cart_id"`
	WishListID     int64  `json:"wish_id"`
}

func newUserResponse(user db.User) userResponse {
	return userResponse{
		UserID:    user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Telephone: user.Telephone,
	}
}

func newUserWithCartResponse(user db.CreateUserWithCartAndWishListRow) userResponse {
	return userResponse{
		UserID:         user.ID,
		Username:       user.Username,
		Email:          user.Email,
		Telephone:      user.Telephone,
		ShoppingCartID: user.ShoppingCartID,
		WishListID:     user.WishListID,
	}
}

type createUserResponse struct {
	UserSessionID         uuid.UUID    `json:"user_session_id"`
	AccessToken           string       `json:"access_token"`
	AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at"`
	RefreshToken          string       `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at"`
	User                  userResponse `json:"user"`
}

func (server *Server) createUser(ctx *fiber.Ctx) error {
	req := &createUserRequest{}

	if err := parseAndValidate(ctx, Input{req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	arg := db.CreateUserWithCartAndWishListParams{
		Username:  req.Username,
		Email:     req.Email,
		Password:  hashedPassword,
		Telephone: req.Telephone,
	}

	user, err := server.store.CreateUserWithCartAndWishList(ctx.Context(), arg)
	if err != nil {
		if pqErr, ok := err.(*pgconn.PgError); ok {
			switch pqErr.Message {
			case "foreign_key_violation", "unique_violation":
				ctx.Status(fiber.StatusForbidden).JSON(errorResponse(err))
				return err
			}
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return err
	}

	accessToken, accessPayload, err := server.tokenMaker.CreateTokenForUser(
		user.ID,
		user.Username,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	refreshToken, refreshPayload, err := server.tokenMaker.CreateTokenForUser(
		user.ID,
		user.Username,
		server.config.RefreshTokenDuration,
	)
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	arg1 := db.CreateUserSessionParams{
		ID:           refreshPayload.ID,
		UserID:       user.ID,
		RefreshToken: refreshToken,
		UserAgent:    string(ctx.Context().UserAgent()),
		ClientIp:     ctx.IP(),
		ExpiresAt:    refreshPayload.ExpiredAt,
	}

	userSession, err := server.store.CreateUserSession(ctx.Context(), arg1)
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	createdUser := newUserWithCartResponse(user)
	rsp := createUserResponse{
		UserSessionID:         userSession.ID,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
		User:                  createdUser,
	}

	ctx.Status(fiber.StatusOK).JSON(rsp)
	return nil
}

//////////////* Reset Password API //////////////

type resetPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

func (server *Server) resetPassword(ctx *fiber.Ctx) error {
	req := &resetPasswordRequest{}

	if err := parseAndValidate(ctx, Input{req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	getUser, err := server.store.GetUserByEmail(ctx.Context(), req.Email)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	if getUser.IsBlocked {
		err := errors.New("account unauthorized")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	newPassword, err := util.GeneratePassword(10, 3, 2, false, false)
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	hashedPassword, err := util.HashPassword(newPassword)
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	err = util.EmailSend(getUser.Email, newPassword, server.config.GmailRandomPassword)
	if err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	arg := db.UpdateUserParams{
		ID:       getUser.ID,
		Password: null.StringFromPtr(&hashedPassword),
	}

	user, err := server.store.UpdateUser(ctx.Context(), arg)
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

	//! i think it should be empty response
	rsp := newUserResponse(user)
	ctx.Status(fiber.StatusOK).JSON(rsp)
	return nil
}

//////////////* Get API //////////////

type getUserParamsRequest struct {
	UserID int64 `params:"id" validate:"required,min=1"`
}

func (server *Server) getUser(ctx *fiber.Ctx) error {
	params := &getUserParamsRequest{}

	if err := parseAndValidate(ctx, Input{params: params}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.UserPayload)
	if authPayload.UserID != params.UserID {
		err := errors.New("account deosn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	user, err := server.store.GetUser(ctx.Context(), params.UserID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	if user.IsBlocked {
		err := errors.New("account unauthorized")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	rsp := newUserResponse(user)
	ctx.Status(fiber.StatusOK).JSON(rsp)
	return nil
}

// //////////////* List API //////////////

type listUsersParamsRequest struct {
	AdminID int64 `params:"adminId" validate:"required,min=1"`
}

type listUsersQueryRequest struct {
	PageID   int32 `query:"page_id" validate:"required,min=1"`
	PageSize int32 `query:"page_size" validate:"required,min=5,max=10"`
}

func (server *Server) listUsers(ctx *fiber.Ctx) error {
	params := &listUsersParamsRequest{}
	query := &listUsersQueryRequest{}

	if err := parseAndValidate(ctx, Input{params: params, query: query}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.AdminPayload)
	if authPayload.AdminID != params.AdminID || authPayload.TypeID != 1 || !authPayload.Active {
		err := errors.New("account unauthorized")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.ListUsersParams{
		Limit:  query.PageSize,
		Offset: (query.PageID - 1) * query.PageSize,
	}

	users, err := server.store.ListUsers(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}
	ctx.Status(fiber.StatusOK).JSON(users)
	return nil

}

// //////////////* Update API //////////////

type updateUserParamsRequest struct {
	UserID int64 `params:"id" validate:"required,min=1"`
}
type updateUserJsonRequest struct {
	Telephone      int64 `json:"telephone" validate:"omitempty,required,numeric,min=910000000,max=929999999"`
	DefaultPayment int64 `json:"default_payment" validate:"omitempty,required"`
}

func (server *Server) updateUser(ctx *fiber.Ctx) error {
	params := &updateUserParamsRequest{}
	req := &updateUserJsonRequest{}

	if err := parseAndValidate(ctx, Input{params: params, req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.UserPayload)
	if params.UserID != authPayload.UserID {
		err := errors.New("account doesn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.UpdateUserParams{
		ID:             authPayload.UserID,
		Telephone:      null.IntFromPtr(&req.Telephone),
		DefaultPayment: null.IntFromPtr(&req.DefaultPayment),
	}

	user, err := server.store.UpdateUser(ctx.Context(), arg)
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

	rsp := newUserResponse(user)
	ctx.Status(fiber.StatusOK).JSON(rsp)
	return nil
}

// //////////////* Delete API //////////////

type deleteUserParamsRequest struct {
	UserID  int64 `params:"id" validate:"required,min=1"`
	AdminID int64 `params:"adminId" validate:"required,min=1"`
}

func (server *Server) deleteUser(ctx *fiber.Ctx) error {
	params := &deleteUserParamsRequest{}

	if err := parseAndValidate(ctx, Input{params: params}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.AdminPayload)
	if authPayload.AdminID != params.AdminID || authPayload.TypeID != 1 || !authPayload.Active {
		err := errors.New("account unauthorized")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}
	_, err := server.store.DeleteUser(ctx.Context(), params.UserID)
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

// //////////////* Login API //////////////

type loginUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type loginUserResponse struct {
	UserSessionID         uuid.UUID    `json:"user_session_id"`
	AccessToken           string       `json:"access_token"`
	AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at"`
	RefreshToken          string       `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at"`
	User                  userResponse `json:"user"`
}

func newUserLoginResponse(user db.GetUserByEmailRow) userResponse {
	return userResponse{
		UserID:         user.ID,
		Username:       user.Username,
		Email:          user.Email,
		Telephone:      user.Telephone,
		ShoppingCartID: user.ShopCartID.Int64,
		WishListID:     user.WishListID.Int64,
	}
}

func (server *Server) loginUser(ctx *fiber.Ctx) error {
	req := &loginUserRequest{}

	if err := parseAndValidate(ctx, Input{req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	user, err := server.store.GetUserByEmail(ctx.Context(), req.Email)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	if user.IsBlocked {
		err := errors.New("account unauthorized")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	err = util.CheckPassword(req.Password, user.Password)
	if err != nil {
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	accessToken, accessPayload, err := server.tokenMaker.CreateTokenForUser(
		user.ID,
		user.Username,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	refreshToken, refreshPayload, err := server.tokenMaker.CreateTokenForUser(
		user.ID,
		user.Username,
		server.config.RefreshTokenDuration,
	)
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	arg := db.CreateUserSessionParams{
		ID:           refreshPayload.ID,
		UserID:       user.ID,
		RefreshToken: refreshToken,
		UserAgent:    string(ctx.Context().UserAgent()),
		ClientIp:     ctx.IP(),
		ExpiresAt:    refreshPayload.ExpiredAt,
	}

	userSession, err := server.store.CreateUserSession(ctx.Context(), arg)
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	rsp := loginUserResponse{
		UserSessionID:         userSession.ID,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
		User:                  newUserLoginResponse(user),
	}
	ctx.Status(fiber.StatusOK).JSON(rsp)
	return nil
}

// //////////////* Logout API //////////////

type logoutUserParamsRequest struct {
	UserID int64 `params:"id" validate:"required,min=1"`
}

type logoutUserJsonRequest struct {
	UserSessionID string `json:"user_session_id" validate:"required"`
	RefreshToken  string `json:"refresh_token" validate:"required"`
}

func (server *Server) logoutUser(ctx *fiber.Ctx) error {
	params := &logoutUserParamsRequest{}
	req := &logoutUserJsonRequest{}

	if err := parseAndValidate(ctx, Input{params: params, req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}
	userSessionID, err := uuid.Parse(req.UserSessionID)
	if err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationPayloadKey).(*token.UserPayload)
	if params.UserID != authPayload.UserID {
		err := errors.New("account doesn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.UpdateUserSessionParams{
		ID:           userSessionID,
		UserID:       authPayload.UserID,
		RefreshToken: req.RefreshToken,
		IsBlocked:    null.BoolFrom(true),
	}

	_, err = server.store.UpdateUserSession(ctx.Context(), arg)
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

	ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
	return nil
}
