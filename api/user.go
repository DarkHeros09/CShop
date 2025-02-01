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

//////////////* Create API //////////////

type createUserRequest struct {
	//alphanumunicode
	Username string `json:"username" validate:"required,alphanumunicode"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	// Telephone int32  `json:"telephone" validate:"required,numeric,min=910000000,max=929999999"`
	// FcmToken  string `json:"fcm_token" validate:"omitempty,required"`
	// DeviceID  string `json:"device_id" validate:"omitempty,required"`
}

type userResponse struct {
	UserID   int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	// Telephone      int32  `json:"telephone"`
	ShoppingCartID  int64 `json:"cart_id"`
	WishListID      int64 `json:"wish_id"`
	IsBlocked       bool  `json:"is_blocked"`
	IsEmailVerified bool  `json:"is_email_verified"`
}

func newUserResponse(user db.User) userResponse {
	return userResponse{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		// Telephone: user.Telephone,
	}
}

func newUserWithCartResponse(user db.CreateUserWithCartAndWishListRow) userResponse {
	return userResponse{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		// Telephone:      user.Telephone,
		ShoppingCartID: user.ShoppingCartID,
		WishListID:     user.WishListID,
	}
}

type createUserResponse struct {
	UserSessionID         string       `json:"user_session_id"`
	AccessToken           string       `json:"access_token"`
	AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at"`
	RefreshToken          string       `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at"`
	User                  userResponse `json:"user"`
}

func (server *Server) createUser(ctx *fiber.Ctx) error {
	req := &createUserRequest{}

	if err := server.parseAndValidate(ctx, Input{req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	arg := db.CreateUserWithCartAndWishListParams{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
		// Telephone: req.Telephone,
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

	accessToken, accessPayload, err := server.userTokenMaker.CreateTokenForUser(
		user.ID,
		user.Username,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	refreshToken, refreshPayload, err := server.userTokenMaker.CreateTokenForUser(
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

	// arg2 := db.CreateNotificationParams{
	// 	UserID:   user.ID,
	// 	DeviceID: null.StringFromPtr(&req.DeviceID),
	// 	FcmToken: null.StringFromPtr(&req.FcmToken),
	// }

	// _, err = server.store.CreateNotification(ctx.Context(), arg2)
	// if err != nil {
	// 	ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
	// 	return nil
	// }

	createdUser := newUserWithCartResponse(user)
	rsp := createUserResponse{
		UserSessionID:         userSession.ID.String(),
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
		User:                  createdUser,
	}

	ctx.Status(fiber.StatusOK).JSON(rsp)
	return nil
}

//////////////* SignUpV2 API //////////////

func (server *Server) signUp(ctx *fiber.Ctx) error {
	req := &createUserRequest{}

	if err := server.parseAndValidate(ctx, Input{req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	// check if user already exists
	checkUser, err := server.store.GetVerifyEmailByEmail(ctx.Context(), req.Email)
	if err != nil {
		if err != pgx.ErrNoRows {
			ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
			return nil
		}
	}

	if checkUser.IsEmailVerified {
		ctx.Status(fiber.StatusConflict).JSON(errorResponse(errors.New("email already exists")))
		return nil
	}

	if !checkUser.IsUsed && time.Now().Before(checkUser.ExpiredAt) {
		ctx.Status(fiber.StatusConflict).JSON(errorResponse(errors.New("email already exists")))
		return nil
	}

	err = server.store.DeleteUserByEmailNotVerified(ctx.Context(), checkUser.Email)
	if err != nil {
		if pqErr, ok := err.(*pgconn.PgError); ok {
			switch pqErr.Message {
			case "foreign_key_violation", "unique_violation":
				ctx.Status(fiber.StatusForbidden).JSON(errorResponse(err))
				return nil
			}
		} else if err != pgx.ErrNoRows {
			ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
			return nil
		}
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	arg := db.SignUpTxParams{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
		// Telephone: req.Telephone,
	}

	user, err := server.store.SignUpTx(ctx.Context(), arg)
	if err != nil {
		if pqErr, ok := err.(*pgconn.PgError); ok {
			switch pqErr.Message {
			case "foreign_key_violation", "unique_violation":
				ctx.Status(fiber.StatusForbidden).JSON(errorResponse(err))
				return err
			}
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	// send email
	subject := "Verify your email"

	content := "Please verify your email by entering the following code in the mobile app: " + user.SecretCode

	to := []string{user.Email}

	err = server.sender.SendEmail(
		subject,
		content,
		to,
		nil, nil, nil,
	)
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	createdUser := userResponse{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		// Telephone: user.Telephone,
		IsBlocked:       user.IsBlocked,
		IsEmailVerified: user.IsEmailVerified,
	}

	ctx.Status(fiber.StatusOK).JSON(createdUser)
	return nil
}

// //////////////* Verify OTP API //////////////

type verifyOTPJsonRequest struct {
	Email string `json:"email" validate:"required,email"`
	OTP   string `json:"otp" validate:"required,numeric,len=6"`
}

func (server *Server) verifyOTP(ctx *fiber.Ctx) error {
	req := &verifyOTPJsonRequest{}

	if err := server.parseAndValidate(ctx, Input{req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	arg := db.UpdateVerifyEmailParams{
		Email:      req.Email,
		SecretCode: req.OTP,
	}

	user, err := server.store.UpdateVerifyEmail(ctx.Context(), arg)
	if err != nil {
		if pqErr, ok := err.(*pgconn.PgError); ok {
			switch pqErr.Message {
			case "foreign_key_violation", "unique_violation":
				ctx.Status(fiber.StatusForbidden).JSON(errorResponse(err))
				return nil
			}
		}
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	accessToken, accessPayload, err := server.userTokenMaker.CreateTokenForUser(
		user.ID,
		user.Username,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	refreshToken, refreshPayload, err := server.userTokenMaker.CreateTokenForUser(
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

	userResponse := userResponse{
		UserID:          user.ID,
		Username:        user.Username,
		Email:           user.Email,
		IsBlocked:       user.IsBlocked,
		IsEmailVerified: user.IsEmailVerified,
		ShoppingCartID:  user.ShoppingCartID,
		WishListID:      user.WishListID,
	}

	rsp := createUserResponse{
		UserSessionID:         userSession.ID.String(),
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
		User:                  userResponse,
	}
	ctx.Status(fiber.StatusOK).JSON(rsp)
	return nil
}

// //////////////* Resend OTP API //////////////

type resendOTPJsonRequest struct {
	Email string `json:"email" validate:"required,email"`
}

func (server *Server) resendOTP(ctx *fiber.Ctx) error {
	req := &resendOTPJsonRequest{}

	if err := server.parseAndValidate(ctx, Input{req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	var secretCode string

	// check if user already exists
	checkUser, err := server.store.GetVerifyEmailByEmail(ctx.Context(), req.Email)
	if err != nil {
		if err != pgx.ErrNoRows {
			ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
			return nil
		}
	}

	if checkUser.IsEmailVerified {
		ctx.Status(fiber.StatusConflict).JSON(errorResponse(errors.New("email already verified")))
		return nil
	}

	if !checkUser.IsUsed && time.Now().Before(checkUser.ExpiredAt) {
		secretCode = checkUser.SecretCode
	} else {
		secretCode = util.GenerateOTP()

		arg := db.CreateVerifyEmailParams{
			UserID:     checkUser.UserID,
			SecretCode: secretCode,
		}
		_, err := server.store.CreateVerifyEmail(ctx.Context(), arg)
		if err != nil {
			ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
			return nil
		}
	}

	// send email
	subject := "Verify your email"

	content := "Please verify your email by entering the following code in the mobile app: " + secretCode

	to := []string{req.Email}

	err = server.sender.SendEmail(
		subject,
		content,
		to,
		nil, nil, nil,
	)
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "OTP sent successfully"})
	return nil
}

// ////////////* Reset Password Request API Mobile //////////////
type resetPasswordRequestJsonRequest struct {
	Email string `json:"email" validate:"required,email"`
}

func (server *Server) resetPasswordRequest(ctx *fiber.Ctx) error {
	req := &resetPasswordRequestJsonRequest{}

	var secretCode string
	if err := server.parseAndValidate(ctx, Input{req: req}); err != nil {
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
		ctx.Status(fiber.StatusForbidden).JSON(errorResponse(err))
		return nil
	} else {
		secretCode = util.GenerateOTP()

		arg := db.CreateResetPasswordParams{
			UserID:     user.ID,
			SecretCode: secretCode,
		}
		_, err := server.store.CreateResetPassword(ctx.Context(), arg)
		if err != nil {
			ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
			return nil
		}
	}

	// send email
	subject := "Reset Your Password"

	content := `Dear ` + user.Username + `,

	You have requested to reset your account password. Please use the One-Time Password (OTP) below to complete the password reset process:
	
	Your OTP: ` + secretCode + `
	
	For security reasons, this code is valid for 10 minutes only. If you did not request a password reset, you can safely ignore this email.
	
	If you need further assistance, please contact our support team.
	
	Best regards,  
	Classic Shop`

	to := []string{req.Email}

	err = server.sender.SendEmail(
		subject,
		content,
		to,
		nil, nil, nil,
	)
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "OTP sent successfully"})
	return nil
}

// //////////////* Verify Reset Password Request OTP Mobile API //////////////

type verifyResetPasswordOTPJsonRequest struct {
	Email string `json:"email" validate:"required,email"`
	OTP   string `json:"otp" validate:"required,numeric,len=6"`
}

func (server *Server) verifyResetPasswordOTP(ctx *fiber.Ctx) error {
	req := &verifyResetPasswordOTPJsonRequest{}

	if err := server.parseAndValidate(ctx, Input{req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	resetPassword, err := server.store.GetResetPasswordsByEmail(ctx.Context(), req.Email)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	if resetPassword.IsBlockedUser {
		err := errors.New("account unauthorized")
		ctx.Status(fiber.StatusForbidden).JSON(errorResponse(err))
		return nil
	}

	arg := db.UpdateResetPasswordParams{
		ID:         resetPassword.ID,
		SecretCode: req.OTP,
	}

	_, err = server.store.UpdateResetPassword(ctx.Context(), arg)
	if err != nil {
		if pqErr, ok := err.(*pgconn.PgError); ok {
			switch pqErr.Message {
			case "foreign_key_violation", "unique_violation":
				ctx.Status(fiber.StatusForbidden).JSON(errorResponse(err))
				return nil
			}
		}
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
	return nil
}

// //////////////* Resend Reset Password Request OTP Mobile API //////////////

type resendResetPasswordOTPJsonRequest struct {
	Email string `json:"email" validate:"required,email"`
}

func (server *Server) resendResetPasswordOTP(ctx *fiber.Ctx) error {
	req := &resendResetPasswordOTPJsonRequest{}

	if err := server.parseAndValidate(ctx, Input{req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	var secretCode string
	// check if user already exists
	checkUser, err := server.store.GetResetPasswordsByEmail(ctx.Context(), req.Email)
	if err != nil {
		if err != pgx.ErrNoRows {
			ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
			return nil
		}
	}

	if checkUser.IsBlockedUser {
		err := errors.New("account unauthorized")
		ctx.Status(fiber.StatusForbidden).JSON(errorResponse(err))
		return nil
	}

	if !checkUser.IsUsed && time.Now().Before(checkUser.ExpiredAt) {
		secretCode = checkUser.SecretCode
	} else {
		secretCode = util.GenerateOTP()

		arg := db.CreateResetPasswordParams{
			UserID:     checkUser.UserID,
			SecretCode: secretCode,
		}
		_, err := server.store.CreateResetPassword(ctx.Context(), arg)
		if err != nil {
			ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
			return nil
		}
	}

	// send email
	subject := "Reset Your Password"

	content := `Dear ` + checkUser.Username + `,

	You have requested to reset your account password. Please use the One-Time Password (OTP) below to complete the password reset process:
	
	Your OTP: ` + secretCode + `
	
	For security reasons, this code is valid for 10 minutes only. If you did not request a password reset, you can safely ignore this email.
	
	If you need further assistance, please contact our support team.
	
	Best regards,  
	Classic Shop`

	to := []string{req.Email}

	err = server.sender.SendEmail(
		subject,
		content,
		to,
		nil, nil, nil,
	)
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "OTP sent successfully"})
	return nil
}

//////////////* Reset Password Approved API //////////////

type resetPasswordApprovedRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	OTP      string `json:"otp" validate:"required,numeric,len=6"`
}

func (server *Server) resetPasswordApproved(ctx *fiber.Ctx) error {
	req := &resetPasswordApprovedRequest{}

	if err := server.parseAndValidate(ctx, Input{req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	lastUsedPasswordReset, err := server.store.GetLastUsedResetPassword(ctx.Context(), req.Email)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	if lastUsedPasswordReset.SecretCode == req.OTP {

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
			ctx.Status(fiber.StatusForbidden).JSON(errorResponse(err))
			return nil
		}

		hashedPassword, err := util.HashPassword(req.Password)
		if err != nil {
			ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
			return nil
		}

		// taskPayload := &worker.PayloadSendResetPassword{
		// 	Email: getUser.Email,
		// }

		// opts := []asynq.Option{
		// 	asynq.MaxRetry(10),
		// 	asynq.ProcessIn(10 * time.Second),
		// 	asynq.Queue(worker.QueueCritical),
		// }

		// err = server.taskDistributor.DistributeTaskSendResetPassword(ctx.Context(), taskPayload, opts...)
		// if err != nil {
		// 	ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		// 	return nil
		// }

		arg2 := db.UpdateUserParams{
			ID:       getUser.ID,
			Password: null.StringFromPtr(&hashedPassword),
		}

		_, err = server.store.UpdateUser(ctx.Context(), arg2)
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

		// //! i think it should be empty response
		// rsp := newUserResponse(user)
		// ctx.Status(fiber.StatusOK).JSON(rsp)
		ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
		return nil
	}
	ctx.Status(fiber.StatusBadRequest)
	return nil
}

// ////////////* Change Password API //////////////
type changePasswordParamsRequest struct {
	UserID int64 `params:"id" validate:"required,min=1"`
}
type changePasswordJsonRequest struct {
	OldPassword string `json:"old_password" validate:"required,min=6"`
	NewPassword string `json:"new_password" validate:"required,min=6"`
}

func (server *Server) changePassword(ctx *fiber.Ctx) error {
	params := &changePasswordParamsRequest{}
	req := &changePasswordJsonRequest{}

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

	if user.IsBlocked || !user.IsEmailVerified {
		err := errors.New("account unauthorized")
		ctx.Status(fiber.StatusForbidden).JSON(errorResponse(err))
		return nil

	}

	err = util.CheckPassword(req.OldPassword, user.Password)
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	newHashedPassword, err := util.HashPassword(req.NewPassword)
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	arg := db.UpdateUserPasswordParams{
		ID:          authPayload.UserID,
		Oldpassword: user.Password,
		Newpassword: newHashedPassword,
	}

	_, err = server.store.UpdateUserPassword(ctx.Context(), arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.Status(fiber.StatusNotFound).JSON(errorResponse(err))
			return nil
		}
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
	return nil
}

//////////////* Get API //////////////

type getUserParamsRequest struct {
	UserID int64 `params:"id" validate:"required,min=1"`
}

func (server *Server) getUser(ctx *fiber.Ctx) error {
	params := &getUserParamsRequest{}

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
		ctx.Status(fiber.StatusForbidden).JSON(errorResponse(err))
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

	if err := server.parseAndValidate(ctx, Input{params: params, query: query}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationAdminPayloadKey).(*token.AdminPayload)
	if authPayload.AdminID != params.AdminID || authPayload.TypeID != 1 || !authPayload.Active {
		err := errors.New("account unauthorized")
		ctx.Status(fiber.StatusForbidden).JSON(errorResponse(err))
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
	// Telephone      *int64 `json:"telephone" validate:"omitempty,required,numeric,min=910000000,max=929999999"`
	DefaultPayment *int64 `json:"default_payment" validate:"omitempty,required"`
}

func (server *Server) updateUser(ctx *fiber.Ctx) error {
	params := &updateUserParamsRequest{}
	req := &updateUserJsonRequest{}

	if err := server.parseAndValidate(ctx, Input{params: params, req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationUserPayloadKey).(*token.UserPayload)
	if params.UserID != authPayload.UserID {
		err := errors.New("account doesn't belong to the authenticated user")
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	arg := db.UpdateUserParams{
		ID: authPayload.UserID,
		// Telephone:      null.IntFromPtr(req.Telephone),
		DefaultPayment: null.IntFromPtr(req.DefaultPayment),
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

	if err := server.parseAndValidate(ctx, Input{params: params}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationAdminPayloadKey).(*token.AdminPayload)
	if authPayload.AdminID != params.AdminID || authPayload.TypeID != 1 || !authPayload.Active {
		err := errors.New("account unauthorized")
		ctx.Status(fiber.StatusForbidden).JSON(errorResponse(err))
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

// type loginUserResponse struct {
// 	UserSessionID         string       `json:"user_session_id"`
// 	AccessToken           string       `json:"access_token"`
// 	AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at"`
// 	RefreshToken          string       `json:"refresh_token"`
// 	RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at"`
// 	User                  userResponse `json:"user"`
// }

func newUserLoginResponse(user db.GetUserByEmailRow) userResponse {
	return userResponse{
		UserID:          user.ID,
		Username:        user.Username,
		Email:           user.Email,
		IsBlocked:       user.IsBlocked,
		IsEmailVerified: user.IsEmailVerified,
		// Telephone:      user.Telephone,
		ShoppingCartID: user.ShopCartID.Int64,
		WishListID:     user.WishListID.Int64,
	}
}

func (server *Server) loginUser(ctx *fiber.Ctx) error {
	req := &loginUserRequest{}

	if err := server.parseAndValidate(ctx, Input{req: req}); err != nil {
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
		ctx.Status(fiber.StatusForbidden).JSON(errorResponse(err))
		return nil
	}

	err = util.CheckPassword(req.Password, user.Password)
	if err != nil {
		ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse(err))
		return nil
	}

	if !user.IsEmailVerified {
		var secretCode string

		// check if user already exists
		checkUser, err := server.store.GetVerifyEmailByEmail(ctx.Context(), req.Email)
		if err != nil {
			if err != pgx.ErrNoRows {
				ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
				return nil
			}
		}

		if !checkUser.IsUsed && time.Now().Before(checkUser.ExpiredAt) {
			secretCode = checkUser.SecretCode
		} else {
			secretCode = util.GenerateOTP()

			arg := db.CreateVerifyEmailParams{
				UserID:     checkUser.UserID,
				SecretCode: secretCode,
			}
			_, err := server.store.CreateVerifyEmail(ctx.Context(), arg)
			if err != nil {
				ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
				return nil
			}
		}

		// send email
		subject := "Verify your email"

		content := "Please verify your email by entering the following code in the mobile app: " + secretCode

		to := []string{req.Email}

		err = server.sender.SendEmail(
			subject,
			content,
			to,
			nil, nil, nil,
		)
		if err != nil {
			ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
			return nil
		}

		createdUser := userResponse{
			UserID:   user.ID,
			Username: user.Username,
			Email:    user.Email,
			// Telephone: user.Telephone,
			IsBlocked:       user.IsBlocked,
			IsEmailVerified: user.IsEmailVerified,
		}

		ctx.Status(fiber.StatusPreconditionFailed).JSON(createdUser)
		return nil
	}

	accessToken, accessPayload, err := server.userTokenMaker.CreateTokenForUser(
		user.ID,
		user.Username,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError).JSON(errorResponse(err))
		return nil
	}

	refreshToken, refreshPayload, err := server.userTokenMaker.CreateTokenForUser(
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

	rsp := createUserResponse{
		UserSessionID:         userSession.ID.String(),
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

	if err := server.parseAndValidate(ctx, Input{params: params, req: req}); err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}
	userSessionID, err := uuid.Parse(req.UserSessionID)
	if err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(errorResponse(err))
		return nil
	}

	authPayload := ctx.Locals(authorizationUserPayloadKey).(*token.UserPayload)
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
