package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"

	mockdb "github.com/cshop/v3/db/mock"
	db "github.com/cshop/v3/db/sqlc"
	mockik "github.com/cshop/v3/image/mock"
	mockemail "github.com/cshop/v3/mail/mock"
	"github.com/cshop/v3/token"
	"github.com/cshop/v3/util"
	mockwk "github.com/cshop/v3/worker/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type eqCreateUserParamsMatcher struct {
	arg      db.CreateUserWithCartAndWishListParams
	password string
}

func (e eqCreateUserParamsMatcher) Matches(x interface{}) bool {
	arg, ok := x.(db.CreateUserWithCartAndWishListParams)
	if !ok {
		return false
	}

	err := util.CheckPassword(e.password, arg.Password)
	if err != nil {
		return false
	}

	e.arg.Password = arg.Password
	return reflect.DeepEqual(e.arg, arg)
}

func (e eqCreateUserParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.arg, e.password)
}

func EqCreateUserParamsMatcher(arg db.CreateUserWithCartAndWishListParams, password string) gomock.Matcher {
	return eqCreateUserParamsMatcher{arg, password}
}

func addAccess(
	t *testing.T,
	tokenMaker token.Maker,
	userID int64,
	username string,
	duration time.Duration,
) (string, *token.UserPayload, error) {
	token, payload, err := tokenMaker.CreateTokenForUser(userID, username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, payload)
	return token, payload, err
}

func TestCreateUserAPI(t *testing.T) {
	t.Parallel()
	userChan := make(chan db.CreateUserWithCartAndWishListRow)
	passwordChan := make(chan string)

	go func() {
		user, password := randomUserWithCartAndWishList(t)
		userChan <- user
		passwordChan <- password
	}()
	user := <-userChan
	password := <-passwordChan
	// user, password := randomUserWithCartAndWishList(t)
	finalRsp := createUserResponse{
		User: newUserResponse(db.User{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Password: password,
			// Telephone:      user.Telephone,
			IsBlocked:      user.IsBlocked,
			DefaultPayment: user.DefaultPayment,
		})}

	testCases := []struct {
		name          string
		body          fiber.Map
		buildStubs    func(store *mockdb.MockStore, tokenMaker token.Maker)
		checkResponse func(rsp *http.Response)
	}{
		{
			name: "OK",
			body: fiber.Map{
				"username": user.Username,
				"email":    user.Email,
				"password": password,
				// "telephone": user.Telephone,
				// "fcm_token": util.RandomString(32),
				// "device_id": util.RandomString(32),
			},
			buildStubs: func(store *mockdb.MockStore, tokenMaker token.Maker) {

				arg := db.CreateUserWithCartAndWishListParams{
					Username: user.Username,
					Email:    user.Email,
					// // Telephone: user.Telephone,
				}

				store.EXPECT().
					CreateUserWithCartAndWishList(gomock.Any(), EqCreateUserParamsMatcher(arg, password)).
					Times(1).
					Return(user, nil)
				accessToken, accessPayload, err := addAccess(t, tokenMaker, user.ID, user.Username, time.Duration(time.Duration.Minutes(30)))
				require.NoError(t, err)
				refreshToken, refreshPayload, err := addAccess(t, tokenMaker, user.ID, user.Username, time.Duration(time.Duration.Hours(720)))
				require.NoError(t, err)

				userSession := db.UserSession{
					ID: uuid.New(),
					// UserID:       user.ID,
					// RefreshToken: refreshToken,
					// UserAgent:    "TestAgent",
					// ClientIp:     "TestIP",
					// IsBlocked:    false,
					// ExpiresAt:    accessPayload.ExpiredAt,
				}

				finalRsp = createUserResponse{
					UserSessionID:         userSession.ID.String(),
					AccessToken:           accessToken,
					AccessTokenExpiresAt:  accessPayload.ExpiredAt,
					RefreshToken:          refreshToken,
					RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
					User: newUserResponse(db.User{
						ID:       user.ID,
						Username: user.Username,
						Email:    user.Email,
						Password: password,
						// // Telephone:      user.Telephone,
						IsBlocked:      user.IsBlocked,
						DefaultPayment: user.DefaultPayment,
					}),
				}

				// store.EXPECT().
				// 	CreateUserWithCartAndWishList(gomock.Any(), gomock.Any()).
				// 	Times(1)

				store.EXPECT().CreateUserSession(gomock.Any(), gomock.Any()).
					Times(1)

				// store.EXPECT().CreateNotification(gomock.Any(), gomock.Any()).
				// 	Times(1)

			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchUserForCreate(t, rsp.Body, finalRsp)
			},
		},
		{
			name: "InternalError",
			body: fiber.Map{
				"username": user.Username,
				"email":    user.Email,
				"password": password,
				// "telephone": user.Telephone,
				// "fcm_token": util.RandomString(32),
				// "device_id": util.RandomString(32),
			},
			buildStubs: func(store *mockdb.MockStore, tokenMaker token.Maker) {
				arg := db.CreateUserWithCartAndWishListParams{
					Username: user.Username,
					Email:    user.Email,
					// // Telephone: user.Telephone,
				}

				store.EXPECT().
					CreateUserWithCartAndWishList(gomock.Any(), EqCreateUserParamsMatcher(arg, password)).
					Times(1).
					Return(db.CreateUserWithCartAndWishListRow{}, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name: "DuplicateUsername",
			body: fiber.Map{
				"username": user.Username,
				"email":    user.Email,
				"password": password,
				// "telephone": user.Telephone,
				// "fcm_token": util.RandomString(32),
				// "device_id": util.RandomString(32),
			},
			buildStubs: func(store *mockdb.MockStore, tokenMaker token.Maker) {
				arg := db.CreateUserWithCartAndWishListParams{
					Username: user.Username,
					Email:    user.Email,
					// // Telephone: user.Telephone,
				}

				store.EXPECT().
					CreateUserWithCartAndWishList(gomock.Any(), EqCreateUserParamsMatcher(arg, password)).
					Times(1).
					Return(db.CreateUserWithCartAndWishListRow{},
						&pgconn.PgError{
							Code:    "23505",
							Message: "unique_violation"},
					)
			},
			checkResponse: func(rsp *http.Response) {
				//! should return forbiddenStatus
				// fix url path
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		// {
		// 	name: "InvalidUsername",
		// 	body: fiber.Map{
		// 		"username":  "invalid-user#1",
		// 		"email":     user.Email,
		// 		"password":  password,
		// "telephone": user.Telephone,
		// 	},
		// 	buildStubs: func(store *mockdb.MockStore, tokenMaker token.Maker) {
		// 		arg := db.CreateUserWithCartAndWishListParams{
		// 			Username:  user.Username,
		// 			Email:     user.Email,
		// Telephone: user.Telephone,
		// 		}

		// 		store.EXPECT().
		// 			CreateUserWithCartAndWishList(gomock.Any(), EqCreateUserParamsMatcher(arg, password)).
		// 			Times(0)
		// 	},
		// 	checkResponse: func(rsp *http.Response) {
		// 		require.Equal(t, http.StatusBadRequest, rsp.StatusCode)
		// 	},
		// },
		{
			name: "InvalidEmail",
			body: fiber.Map{
				"username": user.Username,
				"email":    "invalid-user#1",
				"password": password,
				// "telephone": user.Telephone,
				// "fcm_token": util.RandomString(32),
				// "device_id": util.RandomString(32),
			},

			buildStubs: func(store *mockdb.MockStore, tokenMaker token.Maker) {
				arg := db.CreateUserWithCartAndWishListParams{
					Username: user.Username,
					Email:    user.Email,
					// // Telephone: user.Telephone,
				}

				store.EXPECT().
					CreateUserWithCartAndWishList(gomock.Any(), EqCreateUserParamsMatcher(arg, password)).
					Times(0)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusBadRequest, rsp.StatusCode)
			},
		},
		{
			name: "TooShortPassword",
			body: fiber.Map{
				"username": user.Username,
				"email":    user.Email,
				"password": "123",
				// "telephone": user.Telephone,
				// "fcm_token": util.RandomString(32),
				// "device_id": util.RandomString(32),
			},
			buildStubs: func(store *mockdb.MockStore, tokenMaker token.Maker) {
				arg := db.CreateUserWithCartAndWishListParams{
					Username: user.Username,
					Email:    user.Email,
					// // Telephone: user.Telephone,
				}

				store.EXPECT().
					CreateUserWithCartAndWishList(gomock.Any(), EqCreateUserParamsMatcher(arg, password)).
					Times(0)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusBadRequest, rsp.StatusCode)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			store := mockdb.NewMockStore(ctrl)
			worker := mockwk.NewMockTaskDistributor(ctrl)
			ik := mockik.NewMockImageKitManagement(ctrl)
			mailSender := mockemail.NewMockEmailSender(ctrl)

			server := newTestServer(t, store, worker, ik, mailSender)
			tc.buildStubs(store, server.userTokenMaker)

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/api/v1/users"
			request, err := http.NewRequest(fiber.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
		})

	}
}

type eqSignUpTxParamsMatcher struct {
	arg      db.SignUpTxParams
	password string
}

func (e eqSignUpTxParamsMatcher) Matches(x interface{}) bool {
	arg, ok := x.(db.SignUpTxParams)
	if !ok {
		return false
	}

	err := util.CheckPassword(e.password, arg.Password)
	if err != nil {
		return false
	}

	e.arg.Password = arg.Password
	return reflect.DeepEqual(e.arg, arg)
}

func (e eqSignUpTxParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.arg, e.password)
}

func EqSignUpTxParamsMatcher(arg db.SignUpTxParams, password string) gomock.Matcher {
	return eqSignUpTxParamsMatcher{arg, password}
}

func TestSignUpAPI(t *testing.T) {
	t.Parallel()
	userChan := make(chan db.SignUpTxResult)
	passwordChan := make(chan string)
	verifyEmailChan := make(chan db.GetVerifyEmailByEmailRow)

	go func() {
		user, password := randomSignUpV2User()
		userChan <- user
		passwordChan <- password
		verifyEmailChan <- randomVerifyEmail(user)
	}()
	user := <-userChan
	password := <-passwordChan
	verifyEmail := <-verifyEmailChan
	// user, password := randomUserWithCartAndWishList(t)
	finalRsp := userResponse{
		UserID:          user.ID,
		Username:        user.Username,
		Email:           user.Email,
		IsBlocked:       user.IsBlocked,
		IsEmailVerified: user.IsEmailVerified,
	}

	testCases := []struct {
		name          string
		body          fiber.Map
		buildStubs    func(store *mockdb.MockStore, sender *mockemail.MockEmailSender, tokenMaker token.Maker)
		checkResponse func(rsp *http.Response)
	}{
		{
			name: "OK",
			body: fiber.Map{
				"username": user.Username,
				"email":    user.Email,
				"password": password,
				// "telephone": user.Telephone,
				// "fcm_token": util.RandomString(32),
				// "device_id": util.RandomString(32),
			},
			buildStubs: func(store *mockdb.MockStore, sender *mockemail.MockEmailSender, tokenMaker token.Maker) {
				store.EXPECT().GetVerifyEmailByEmail(gomock.Any(), user.Email).
					Times(1).Return(verifyEmail, nil)

				store.EXPECT().DeleteUserByEmailNotVerified(gomock.Any(), user.Email).Times(1)

				arg := db.SignUpTxParams{
					Username: user.Username,
					Email:    user.Email,
					// Password: password,
					// // Telephone: user.Telephone,
				}

				store.EXPECT().
					SignUpTx(gomock.Any(), EqSignUpTxParamsMatcher(arg, password)).
					Times(1).
					Return(user, nil)

				sender.EXPECT().SendEmail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchUserForSignUp(t, rsp.Body, finalRsp)
			},
		},
		{
			name: "InternalError",
			body: fiber.Map{
				"username": user.Username,
				"email":    user.Email,
				"password": password,
				// "telephone": user.Telephone,
				// "fcm_token": util.RandomString(32),
				// "device_id": util.RandomString(32),
			},
			buildStubs: func(store *mockdb.MockStore, sender *mockemail.MockEmailSender, tokenMaker token.Maker) {
				store.EXPECT().GetVerifyEmailByEmail(gomock.Any(), user.Email).
					Times(1).Return(db.GetVerifyEmailByEmailRow{}, pgx.ErrTxClosed)

				store.EXPECT().DeleteUserByEmailNotVerified(gomock.Any(), user.Email).Times(0)

				arg := db.SignUpTxParams{
					Username: user.Username,
					Email:    user.Email,
					// // Telephone: user.Telephone,
				}

				store.EXPECT().
					SignUpTx(gomock.Any(), EqSignUpTxParamsMatcher(arg, password)).
					Times(0)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name: "DuplicateUsername",
			body: fiber.Map{
				"username": user.Username,
				"email":    user.Email,
				"password": password,
				// "telephone": user.Telephone,
				// "fcm_token": util.RandomString(32),
				// "device_id": util.RandomString(32),
			},
			buildStubs: func(store *mockdb.MockStore, sender *mockemail.MockEmailSender, tokenMaker token.Maker) {

				store.EXPECT().GetVerifyEmailByEmail(gomock.Any(), user.Email).
					Times(1).Return(verifyEmail, nil)

				store.EXPECT().DeleteUserByEmailNotVerified(gomock.Any(), user.Email).Times(1)

				arg := db.SignUpTxParams{
					Username: user.Username,
					Email:    user.Email,
					Password: password,
					// // Telephone: user.Telephone,
				}

				store.EXPECT().
					SignUpTx(gomock.Any(), EqSignUpTxParamsMatcher(arg, password)).
					Times(1).
					Return(db.SignUpTxResult{},
						&pgconn.PgError{
							Code:    "23505",
							Message: "unique_violation"},
					)
			},
			checkResponse: func(rsp *http.Response) {
				//! should return forbiddenStatus
				// fix url path
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name: "InvalidUsername",
			body: fiber.Map{
				"username": 1111,
				"email":    user.Email,
				"password": password,
			},
			buildStubs: func(store *mockdb.MockStore, sender *mockemail.MockEmailSender, tokenMaker token.Maker) {
				store.EXPECT().GetVerifyEmailByEmail(gomock.Any(), user.Email).
					Times(0)

			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusBadRequest, rsp.StatusCode)
			},
		},
		{
			name: "InvalidEmail",
			body: fiber.Map{
				"username": user.Username,
				"email":    "invalid-user#1",
				"password": password,
				// "telephone": user.Telephone,
				// "fcm_token": util.RandomString(32),
				// "device_id": util.RandomString(32),
			},

			buildStubs: func(store *mockdb.MockStore, sender *mockemail.MockEmailSender, tokenMaker token.Maker) {
				store.EXPECT().GetVerifyEmailByEmail(gomock.Any(), user.Email).
					Times(0)
				arg := db.SignUpTxParams{
					Username: user.Username,
					Email:    user.Email,
					// // Telephone: user.Telephone,
				}

				store.EXPECT().
					SignUpTx(gomock.Any(), EqSignUpTxParamsMatcher(arg, password)).
					Times(0)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusBadRequest, rsp.StatusCode)
			},
		},
		{
			name: "TooShortPassword",
			body: fiber.Map{
				"username": user.Username,
				"email":    user.Email,
				"password": "123",
				// "telephone": user.Telephone,
				// "fcm_token": util.RandomString(32),
				// "device_id": util.RandomString(32),
			},
			buildStubs: func(store *mockdb.MockStore, sender *mockemail.MockEmailSender, tokenMaker token.Maker) {
				store.EXPECT().GetVerifyEmailByEmail(gomock.Any(), user.Email).
					Times(0)
				arg := db.SignUpTxParams{
					Username: user.Username,
					Email:    user.Email,
					// // Telephone: user.Telephone,
				}

				store.EXPECT().
					SignUpTx(gomock.Any(), EqSignUpTxParamsMatcher(arg, password)).
					Times(0)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusBadRequest, rsp.StatusCode)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			store := mockdb.NewMockStore(ctrl)
			worker := mockwk.NewMockTaskDistributor(ctrl)
			ik := mockik.NewMockImageKitManagement(ctrl)
			mailSender := mockemail.NewMockEmailSender(ctrl)

			server := newTestServer(t, store, worker, ik, mailSender)
			tc.buildStubs(store, mailSender, server.userTokenMaker)

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/api/v1/users/signup"
			request, err := http.NewRequest(fiber.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
		})

	}
}
func TestVerifyOTPAPI(t *testing.T) {
	t.Parallel()
	userChan := make(chan db.SignUpTxResult)
	passwordChan := make(chan string)
	verifyEmailChan := make(chan db.GetVerifyEmailByEmailRow)

	go func() {
		user, password := randomSignUpV2User()
		userChan <- user
		passwordChan <- password
		verifyEmailChan <- randomVerifyEmail(user)
	}()
	user := <-userChan
	password := <-passwordChan
	verifyEmail := <-verifyEmailChan
	updateVerifyEmail := randomUpdateVerifyEmail(user, verifyEmail)
	// user, password := randomUserWithCartAndWishList(t)
	finalRsp := createUserResponse{
		User: newUserWithCartResponse(db.CreateUserWithCartAndWishListRow{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Password: password,
			// Telephone:      user.Telephone,
			IsBlocked:       user.IsBlocked,
			IsEmailVerified: user.IsEmailVerified,
			// DefaultPayment: user.DefaultPayment,
			ShoppingCartID: updateVerifyEmail.ShoppingCartID,
			WishListID:     updateVerifyEmail.WishListID,
		})}

	testCases := []struct {
		name          string
		body          fiber.Map
		buildStubs    func(store *mockdb.MockStore, sender *mockemail.MockEmailSender, tokenMaker token.Maker)
		checkResponse func(rsp *http.Response)
	}{
		{
			name: "OK",
			body: fiber.Map{
				"email": user.Email,
				"otp":   verifyEmail.SecretCode,
			},
			buildStubs: func(store *mockdb.MockStore, sender *mockemail.MockEmailSender, tokenMaker token.Maker) {

				arg := db.UpdateVerifyEmailParams{
					Email:      user.Email,
					SecretCode: verifyEmail.SecretCode,
				}

				store.EXPECT().UpdateVerifyEmail(gomock.Any(), arg).
					Times(1).Return(updateVerifyEmail, nil)

				accessToken, accessPayload, err := addAccess(t, tokenMaker, user.ID, user.Username, time.Duration(time.Duration.Minutes(30)))
				require.NoError(t, err)
				refreshToken, refreshPayload, err := addAccess(t, tokenMaker, user.ID, user.Username, time.Duration(time.Duration.Hours(720)))
				require.NoError(t, err)

				userSession := db.UserSession{
					ID: uuid.New(),
					// UserID:       user.ID,
					// RefreshToken: refreshToken,
					// UserAgent:    "TestAgent",
					// ClientIp:     "TestIP",
					// IsBlocked:    false,
					// ExpiresAt:    accessPayload.ExpiredAt,
				}

				finalRsp = createUserResponse{
					UserSessionID:         userSession.ID.String(),
					AccessToken:           accessToken,
					AccessTokenExpiresAt:  accessPayload.ExpiredAt,
					RefreshToken:          refreshToken,
					RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
					User: newUserWithCartResponse(db.CreateUserWithCartAndWishListRow{
						ID:       user.ID,
						Username: user.Username,
						Email:    user.Email,
						Password: password,
						// Telephone:      user.Telephone,
						IsBlocked:       user.IsBlocked,
						IsEmailVerified: user.IsEmailVerified,
						// DefaultPayment: user.DefaultPayment,
						ShoppingCartID: updateVerifyEmail.ShoppingCartID,
						WishListID:     updateVerifyEmail.WishListID,
					}),
				}

				store.EXPECT().CreateUserSession(gomock.Any(), gomock.Any()).
					Times(1)

			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchUserForCreate(t, rsp.Body, finalRsp)
			},
		},
		{
			name: "InternalError",
			body: fiber.Map{
				"email": user.Email,
				"otp":   verifyEmail.SecretCode,
			},
			buildStubs: func(store *mockdb.MockStore, sender *mockemail.MockEmailSender, tokenMaker token.Maker) {
				arg := db.UpdateVerifyEmailParams{
					Email:      user.Email,
					SecretCode: verifyEmail.SecretCode,
				}

				store.EXPECT().UpdateVerifyEmail(gomock.Any(), arg).
					Times(1).Return(db.UpdateVerifyEmailRow{}, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name: "InvalidEmail",
			body: fiber.Map{
				"email": "1111",
				"otp":   verifyEmail.SecretCode,
			},

			buildStubs: func(store *mockdb.MockStore, sender *mockemail.MockEmailSender, tokenMaker token.Maker) {
				store.EXPECT().UpdateVerifyEmail(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusBadRequest, rsp.StatusCode)
			},
		},
		{
			name: "TooShortOTP",
			body: fiber.Map{
				"email": user.Email,
				"otp":   "12345",
			},
			buildStubs: func(store *mockdb.MockStore, sender *mockemail.MockEmailSender, tokenMaker token.Maker) {
				store.EXPECT().UpdateVerifyEmail(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusBadRequest, rsp.StatusCode)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			store := mockdb.NewMockStore(ctrl)
			worker := mockwk.NewMockTaskDistributor(ctrl)
			ik := mockik.NewMockImageKitManagement(ctrl)
			mailSender := mockemail.NewMockEmailSender(ctrl)

			server := newTestServer(t, store, worker, ik, mailSender)
			tc.buildStubs(store, mailSender, server.userTokenMaker)

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/api/v1/users/verify-otp"
			request, err := http.NewRequest(fiber.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
		})

	}
}

func TestResendOTPAPI(t *testing.T) {
	t.Parallel()
	userChan := make(chan db.SignUpTxResult)
	// passwordChan := make(chan string)
	verifyEmailChan := make(chan db.GetVerifyEmailByEmailRow)

	go func() {
		user, _ := randomSignUpV2User()
		userChan <- user
		// passwordChan <- password
		verifyEmailChan <- randomVerifyEmail(user)
	}()
	user := <-userChan
	// password := <-passwordChan
	verifyEmail := <-verifyEmailChan
	// user, password := randomUserWithCartAndWishList(t)

	testCases := []struct {
		name          string
		body          fiber.Map
		buildStubs    func(store *mockdb.MockStore, sender *mockemail.MockEmailSender, tokenMaker token.Maker)
		checkResponse func(rsp *http.Response)
	}{
		{
			name: "OK",
			body: fiber.Map{
				"email": user.Email,
			},
			buildStubs: func(store *mockdb.MockStore, sender *mockemail.MockEmailSender, tokenMaker token.Maker) {
				store.EXPECT().GetVerifyEmailByEmail(gomock.Any(), user.Email).
					Times(1).Return(verifyEmail, nil)

				store.EXPECT().CreateVerifyEmail(gomock.Any(), gomock.Any()).
					Times(1)

				sender.EXPECT().SendEmail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchResendOTP(t, rsp.Body, fiber.Map{"message": "OTP sent successfully"})
			},
		},
		{
			name: "InternalError",
			body: fiber.Map{
				"email": user.Email,
			},
			buildStubs: func(store *mockdb.MockStore, sender *mockemail.MockEmailSender, tokenMaker token.Maker) {
				store.EXPECT().GetVerifyEmailByEmail(gomock.Any(), user.Email).
					Times(1).Return(db.GetVerifyEmailByEmailRow{}, pgx.ErrTxClosed)

				sender.EXPECT().SendEmail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name: "InvalidEmail",
			body: fiber.Map{
				"email": "invalid-user#1",
			},

			buildStubs: func(store *mockdb.MockStore, sender *mockemail.MockEmailSender, tokenMaker token.Maker) {
				store.EXPECT().GetVerifyEmailByEmail(gomock.Any(), user.Email).
					Times(0)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusBadRequest, rsp.StatusCode)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			store := mockdb.NewMockStore(ctrl)
			worker := mockwk.NewMockTaskDistributor(ctrl)
			ik := mockik.NewMockImageKitManagement(ctrl)
			mailSender := mockemail.NewMockEmailSender(ctrl)

			server := newTestServer(t, store, worker, ik, mailSender)
			tc.buildStubs(store, mailSender, server.userTokenMaker)

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/api/v1/users/resend-otp"
			request, err := http.NewRequest(fiber.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
		})

	}
}

func TestResetPasswordRequestAPI(t *testing.T) {
	t.Parallel()
	userChan := make(chan db.GetUserByEmailRow)
	// passwordChan := make(chan string)
	// verifyEmailChan := make(chan db.GetVerifyEmailByEmailRow)

	go func() {
		user, _ := randomUserLogin(t)
		userChan <- user
		// passwordChan <- password
		// verifyEmailChan <- randomVerifyEmail(user)
	}()
	user := <-userChan
	// password := <-passwordChan
	// verifyEmail := <-verifyEmailChan
	// user, password := randomUserWithCartAndWishList(t)

	testCases := []struct {
		name          string
		body          fiber.Map
		buildStubs    func(store *mockdb.MockStore, sender *mockemail.MockEmailSender, tokenMaker token.Maker)
		checkResponse func(rsp *http.Response)
	}{
		{
			name: "OK",
			body: fiber.Map{
				"email": user.Email,
			},
			buildStubs: func(store *mockdb.MockStore, sender *mockemail.MockEmailSender, tokenMaker token.Maker) {
				store.EXPECT().GetUserByEmail(gomock.Any(), user.Email).
					Times(1).Return(user, nil)

				store.EXPECT().CreateResetPassword(gomock.Any(), gomock.Any()).
					Times(1)

				sender.EXPECT().SendEmail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchResendOTP(t, rsp.Body, fiber.Map{"message": "OTP sent successfully"})
			},
		},
		{
			name: "InternalError",
			body: fiber.Map{
				"email": user.Email,
			},
			buildStubs: func(store *mockdb.MockStore, sender *mockemail.MockEmailSender, tokenMaker token.Maker) {
				store.EXPECT().GetUserByEmail(gomock.Any(), user.Email).
					Times(1).Return(db.GetUserByEmailRow{}, pgx.ErrTxClosed)

				sender.EXPECT().SendEmail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name: "InvalidEmail",
			body: fiber.Map{
				"email": "invalid-user#1",
			},

			buildStubs: func(store *mockdb.MockStore, sender *mockemail.MockEmailSender, tokenMaker token.Maker) {
				store.EXPECT().GetUserByEmail(gomock.Any(), user.Email).
					Times(0)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusBadRequest, rsp.StatusCode)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			store := mockdb.NewMockStore(ctrl)
			worker := mockwk.NewMockTaskDistributor(ctrl)
			ik := mockik.NewMockImageKitManagement(ctrl)
			mailSender := mockemail.NewMockEmailSender(ctrl)

			server := newTestServer(t, store, worker, ik, mailSender)
			tc.buildStubs(store, mailSender, server.userTokenMaker)

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/api/v1/users/reset-password-request"
			request, err := http.NewRequest(fiber.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
		})

	}
}

func TestVerifyResetPasswordOTPAPI(t *testing.T) {
	t.Parallel()
	userChan := make(chan db.GetUserByEmailRow)
	// passwordChan := make(chan string)
	resetPasswordChan := make(chan db.GetResetPasswordsByEmailRow)

	go func() {
		user, _ := randomUserLogin(t)
		userChan <- user
		// passwordChan <- password
		resetPasswordChan <- randomResetPasswordOTP(user)
	}()
	user := <-userChan
	// password := <-passwordChan
	resetPassword := <-resetPasswordChan
	updateResetPassword := randomUpdateResetPasswordOTP(resetPassword)
	// user, password := randomUserWithCartAndWishList(t)

	testCases := []struct {
		name          string
		body          fiber.Map
		buildStubs    func(store *mockdb.MockStore, sender *mockemail.MockEmailSender, tokenMaker token.Maker)
		checkResponse func(rsp *http.Response)
	}{
		{
			name: "OK",
			body: fiber.Map{
				"email": user.Email,
				"otp":   resetPassword.SecretCode,
			},
			buildStubs: func(store *mockdb.MockStore, sender *mockemail.MockEmailSender, tokenMaker token.Maker) {
				store.EXPECT().GetResetPasswordsByEmail(gomock.Any(), user.Email).
					Times(1).Return(resetPassword, nil)

				arg := db.UpdateResetPasswordParams{
					ID:         resetPassword.ID,
					SecretCode: resetPassword.SecretCode,
				}

				store.EXPECT().UpdateResetPassword(gomock.Any(), arg).
					Times(1).Return(updateResetPassword, nil)

			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name: "InternalError",
			body: fiber.Map{
				"email": user.Email,
				"otp":   resetPassword.SecretCode,
			},
			buildStubs: func(store *mockdb.MockStore, sender *mockemail.MockEmailSender, tokenMaker token.Maker) {
				// arg := db.UpdateVerifyEmailParams{
				// 	Email:      user.Email,
				// 	SecretCode: resetPassword.SecretCode,
				// }

				store.EXPECT().GetResetPasswordsByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).Return(db.GetResetPasswordsByEmailRow{}, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name: "InvalidEmail",
			body: fiber.Map{
				"email": "1111",
				"otp":   resetPassword.SecretCode,
			},

			buildStubs: func(store *mockdb.MockStore, sender *mockemail.MockEmailSender, tokenMaker token.Maker) {
				store.EXPECT().GetResetPasswordsByEmail(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusBadRequest, rsp.StatusCode)
			},
		},
		{
			name: "TooShortOTP",
			body: fiber.Map{
				"email": user.Email,
				"otp":   "12345",
			},
			buildStubs: func(store *mockdb.MockStore, sender *mockemail.MockEmailSender, tokenMaker token.Maker) {
				store.EXPECT().GetResetPasswordsByEmail(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusBadRequest, rsp.StatusCode)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			store := mockdb.NewMockStore(ctrl)
			worker := mockwk.NewMockTaskDistributor(ctrl)
			ik := mockik.NewMockImageKitManagement(ctrl)
			mailSender := mockemail.NewMockEmailSender(ctrl)

			server := newTestServer(t, store, worker, ik, mailSender)
			tc.buildStubs(store, mailSender, server.userTokenMaker)

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/api/v1/users/verify-password-reset-otp"
			request, err := http.NewRequest(fiber.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
		})

	}
}

func TestResendPasswordResetOTPAPI(t *testing.T) {
	t.Parallel()
	userChan := make(chan db.GetUserByEmailRow)
	// passwordChan := make(chan string)
	resetPasswordChan := make(chan db.GetResetPasswordsByEmailRow)

	go func() {
		user, _ := randomUserLogin(t)
		userChan <- user
		// passwordChan <- password
		resetPasswordChan <- randomResetPasswordOTP(user)
	}()
	user := <-userChan
	// password := <-passwordChan
	resetPassword := <-resetPasswordChan
	// user, password := randomUserWithCartAndWishList(t)

	testCases := []struct {
		name          string
		body          fiber.Map
		buildStubs    func(store *mockdb.MockStore, sender *mockemail.MockEmailSender, tokenMaker token.Maker)
		checkResponse func(rsp *http.Response)
	}{
		{
			name: "OK",
			body: fiber.Map{
				"email": user.Email,
			},
			buildStubs: func(store *mockdb.MockStore, sender *mockemail.MockEmailSender, tokenMaker token.Maker) {
				store.EXPECT().GetResetPasswordsByEmail(gomock.Any(), user.Email).
					Times(1).Return(resetPassword, nil)

				store.EXPECT().CreateResetPassword(gomock.Any(), gomock.Any()).
					Times(1)

				sender.EXPECT().SendEmail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchResendOTP(t, rsp.Body, fiber.Map{"message": "OTP sent successfully"})
			},
		},
		{
			name: "InternalError",
			body: fiber.Map{
				"email": user.Email,
			},
			buildStubs: func(store *mockdb.MockStore, sender *mockemail.MockEmailSender, tokenMaker token.Maker) {
				store.EXPECT().GetResetPasswordsByEmail(gomock.Any(), user.Email).
					Times(1).Return(db.GetResetPasswordsByEmailRow{}, pgx.ErrTxClosed)

				sender.EXPECT().SendEmail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name: "InvalidEmail",
			body: fiber.Map{
				"email": "invalid-user#1",
			},

			buildStubs: func(store *mockdb.MockStore, sender *mockemail.MockEmailSender, tokenMaker token.Maker) {
				store.EXPECT().GetResetPasswordsByEmail(gomock.Any(), user.Email).
					Times(0)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusBadRequest, rsp.StatusCode)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			store := mockdb.NewMockStore(ctrl)
			worker := mockwk.NewMockTaskDistributor(ctrl)
			ik := mockik.NewMockImageKitManagement(ctrl)
			mailSender := mockemail.NewMockEmailSender(ctrl)

			server := newTestServer(t, store, worker, ik, mailSender)
			tc.buildStubs(store, mailSender, server.userTokenMaker)

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/api/v1/users/resend-password-reset-otp"
			request, err := http.NewRequest(fiber.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
		})

	}
}

/////////

type eqUpdateUserParamsMatcher struct {
	arg      db.UpdateUserParams
	password string
}

func (e eqUpdateUserParamsMatcher) Matches(x interface{}) bool {
	arg, ok := x.(db.UpdateUserParams)
	if !ok {
		return false
	}

	err := util.CheckPassword(e.password, arg.Password.String)
	if err != nil {
		return false
	}

	e.arg.Password.String = arg.Password.String
	return reflect.DeepEqual(e.arg, arg)
}

func (e eqUpdateUserParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.arg.Password.String, e.password)
}

func EqUpdateUserParamsMatcher(arg db.UpdateUserParams, password string) gomock.Matcher {
	return eqUpdateUserParamsMatcher{arg, password}
}

func TestResetPasswordApprovedAPI(t *testing.T) {
	t.Parallel()
	user, _ := randomUserLogin(t)
	newPassword := "newpassword"
	passwordReset := randomResetPassword(user)

	testCases := []struct {
		name          string
		body          fiber.Map
		buildStubs    func(store *mockdb.MockStore, sender *mockemail.MockEmailSender, tokenMaker token.Maker)
		checkResponse func(rsp *http.Response)
	}{
		{
			name: "OK",
			body: fiber.Map{
				"email":    user.Email,
				"password": newPassword,
				"otp":      passwordReset.SecretCode,
			},
			buildStubs: func(store *mockdb.MockStore, sender *mockemail.MockEmailSender, tokenMaker token.Maker) {

				store.EXPECT().GetLastUsedResetPassword(gomock.Any(), user.Email).Times(1).Return(passwordReset, nil)

				if true {
					store.EXPECT().GetUserByEmail(gomock.Any(), user.Email).
						Times(1).Return(user, nil)

					arg := db.UpdateUserParams{
						ID:       user.ID,
						Password: null.StringFrom(newPassword),
					}

					store.EXPECT().UpdateUser(gomock.Any(), EqUpdateUserParamsMatcher(arg, newPassword)).
						Times(1)
				}

			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name: "InternalError",
			body: fiber.Map{
				"email":    user.Email,
				"password": newPassword,
				"otp":      passwordReset.SecretCode,
			},
			buildStubs: func(store *mockdb.MockStore, sender *mockemail.MockEmailSender, tokenMaker token.Maker) {
				store.EXPECT().GetLastUsedResetPassword(gomock.Any(), gomock.Any()).Times(1).Return(db.ResetPassword{}, pgx.ErrTxClosed)

				store.EXPECT().GetUserByEmail(gomock.Any(), user.Email).
					Times(0)
					// .Return(db.GetUserByEmailRow{}, pgx.ErrTxClosed)

				store.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name: "InvalidEmail",
			body: fiber.Map{
				"email":    "invalid-user#1",
				"password": newPassword,
				"otp":      passwordReset.SecretCode,
			},

			buildStubs: func(store *mockdb.MockStore, sender *mockemail.MockEmailSender, tokenMaker token.Maker) {
				store.EXPECT().GetLastUsedResetPassword(gomock.Any(), gomock.Any()).Times(0)
				store.EXPECT().GetUserByEmail(gomock.Any(), user.Email).
					Times(0)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusBadRequest, rsp.StatusCode)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			store := mockdb.NewMockStore(ctrl)
			worker := mockwk.NewMockTaskDistributor(ctrl)
			ik := mockik.NewMockImageKitManagement(ctrl)
			mailSender := mockemail.NewMockEmailSender(ctrl)

			server := newTestServer(t, store, worker, ik, mailSender)
			tc.buildStubs(store, mailSender, server.userTokenMaker)

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/api/v1/users/reset-password-approved"
			request, err := http.NewRequest(fiber.MethodPut, url, bytes.NewReader(data))
			require.NoError(t, err)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
		})

	}
}

func TestLoginUserAPI(t *testing.T) {
	t.Parallel()
	user, password := randomUserLogin(t)
	user2, password2 := randomUserLoginNotVerified(t)

	testCases := []struct {
		name          string
		body          fiber.Map
		buildStubs    func(store *mockdb.MockStore, sender *mockemail.MockEmailSender)
		checkResponse func(rsp *http.Response)
	}{
		{
			name: "OK",
			body: fiber.Map{
				"email":    user.Email,
				"password": password,
			},
			buildStubs: func(store *mockdb.MockStore, sender *mockemail.MockEmailSender) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)

				store.EXPECT().
					GetVerifyEmailByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(0)

				store.EXPECT().
					CreateUserSession(gomock.Any(), gomock.Any()).
					Times(1)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name: "NeedsVerification",
			body: fiber.Map{
				"email":    user2.Email,
				"password": password2,
			},
			buildStubs: func(store *mockdb.MockStore, sender *mockemail.MockEmailSender) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user2.Email)).
					Times(1).
					Return(user2, nil)

				store.EXPECT().
					GetVerifyEmailByEmail(gomock.Any(), gomock.Eq(user2.Email)).
					Times(1)

				store.EXPECT().CreateVerifyEmail(gomock.Any(), gomock.Any()).Times(1)

				sender.EXPECT().SendEmail(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

				store.EXPECT().
					CreateUserSession(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusPreconditionFailed, rsp.StatusCode)
			},
		},
		{
			name: "UserNotFound",
			body: fiber.Map{
				"email":    "NotFound@NotFound.com",
				"password": password,
			},
			buildStubs: func(store *mockdb.MockStore, sender *mockemail.MockEmailSender) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.GetUserByEmailRow{}, pgx.ErrNoRows)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
			},
		},
		{
			name: "IncorrectPassword",
			body: fiber.Map{
				"email":    user.Email,
				"password": "incorrect",
			},
			buildStubs: func(store *mockdb.MockStore, sender *mockemail.MockEmailSender) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name: "InternalError",
			body: fiber.Map{
				"email":    user.Email,
				"password": password,
			},
			buildStubs: func(store *mockdb.MockStore, sender *mockemail.MockEmailSender) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.GetUserByEmailRow{}, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name: "InvalidUsername",
			body: fiber.Map{
				"email":    "invalid-email#1",
				"password": password,
			},
			buildStubs: func(store *mockdb.MockStore, sender *mockemail.MockEmailSender) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusBadRequest, rsp.StatusCode)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		// t.Run(tc.name, func(t *testing.T) {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			store := mockdb.NewMockStore(ctrl)
			worker := mockwk.NewMockTaskDistributor(ctrl)
			ik := mockik.NewMockImageKitManagement(ctrl)
			mailSender := mockemail.NewMockEmailSender(ctrl)
			tc.buildStubs(store, mailSender)

			server := newTestServer(t, store, worker, ik, mailSender)
			// //recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/api/v1/users/login"
			request, err := http.NewRequest(fiber.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			go tc.checkResponse(rsp)
		})
	}
}

func TestLogoutUserAPI(t *testing.T) {
	userSession, userLogin := randomUserLogout(t)

	testCases := []struct {
		name          string
		body          fiber.Map
		UserID        int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(rsp *http.Response)
	}{
		{
			name:   "OK",
			UserID: userLogin.ID,
			body: fiber.Map{
				"user_session_id": userSession.ID,
				"refresh_token":   userSession.RefreshToken,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, userLogin.ID, userLogin.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.UpdateUserSessionParams{
					IsBlocked:    null.BoolFrom(true),
					ID:           userSession.ID,
					UserID:       userSession.UserID,
					RefreshToken: userSession.RefreshToken,
				}
				store.EXPECT().
					UpdateUserSession(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(userSession, nil)

			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:   "NoAuthorization",
			UserID: userLogin.ID,
			body: fiber.Map{
				"user_session_id": userSession.ID,
				"refresh_token":   userSession.RefreshToken,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateUserSession(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:   "InternalError",
			UserID: userLogin.ID,
			body: fiber.Map{
				"user_session_id": userSession.ID,
				"refresh_token":   userSession.RefreshToken,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, userLogin.ID, userLogin.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateUserSession(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.UserSession{}, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:   "InvalidSessionID",
			UserID: userLogin.ID,
			body: fiber.Map{
				"user_session_id": 0,
				"refresh_token":   userSession.RefreshToken,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, userLogin.ID, userLogin.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateUserSession(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusBadRequest, rsp.StatusCode)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			store := mockdb.NewMockStore(ctrl)
			worker := mockwk.NewMockTaskDistributor(ctrl)
			ik := mockik.NewMockImageKitManagement(ctrl)
			mailSender := mockemail.NewMockEmailSender(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store, worker, ik, mailSender)
			// //recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			// url := "/api/v1/users/logout"
			url := fmt.Sprintf("/usr/v1/users/%d/logout", tc.UserID)
			request, err := http.NewRequest(fiber.MethodDelete, url, bytes.NewReader(data))
			require.NoError(t, err)

			request.Header.Set("Content-Type", "application/json")

			tc.setupAuth(t, request, server.userTokenMaker)

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
		})
	}
}

func TestGetUserAPI(t *testing.T) {
	user, _ := randomUser(t)
	require.NotEmpty(t, user)
	require.NotZero(t, user.ID)

	testCases := []struct {
		name          string
		UserID        int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rsp *http.Response)
	}{
		{
			name:   "OK",
			UserID: user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {

				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchUser(t, rsp.Body, user)
			},
		},
		{
			name:   "NoAuthorization",
			UserID: user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:   "UnauthorizedUser",
			UserID: user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, "unauthorizedUser", time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.ID)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:   "NotFound",
			UserID: user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(db.User{}, pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
			},
		},
		{
			name:   "InternalError",
			UserID: user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(db.User{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:   "InvalidID",
			UserID: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(0)

			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusBadRequest, rsp.StatusCode)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t) // no need to call defer ctrl.finish() in 1.6V

			store := mockdb.NewMockStore(ctrl)
			worker := mockwk.NewMockTaskDistributor(ctrl)
			ik := mockik.NewMockImageKitManagement(ctrl)
			emailSender := mockemail.NewMockEmailSender(ctrl)
			// build stubs
			tc.buildStub(store)

			// start test server and send request
			server := newTestServer(t, store, worker, ik, emailSender)
			//recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/usr/v1/users/%d", tc.UserID)
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)
			request.Header.Set("Content-Type", "application/json")

			tc.setupAuth(t, request, server.userTokenMaker)

			rsp, err := server.router.Test(request, -1)
			require.NoError(t, err)
			//check response
			tc.checkResponse(t, rsp)
		})

	}

}

func TestUpdateUserAPI(t *testing.T) {
	user, _ := randomUser(t)

	testCases := []struct {
		name          string
		UserID        int64
		body          fiber.Map
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rsp *http.Response)
	}{
		{
			name:   "OK",
			UserID: user.ID,
			body: fiber.Map{
				// "telephone":       user.Telephone,
				"default_payment": user.DefaultPayment,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateUserParams{
					ID: user.ID,
					// // Telephone:      null.IntFrom(int64(user.Telephone)),
					DefaultPayment: null.IntFrom(user.DefaultPayment.Int64),
				}

				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:   "NoAuthorization",
			UserID: user.ID,
			body:   fiber.Map{
				// "telephone": user.Telephone,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:   "InternalError",
			UserID: user.ID,
			body: fiber.Map{
				// "telephone":       user.Telephone,
				"default_payment": user.DefaultPayment,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateUserParams{
					ID: user.ID,
					// // Telephone:      null.IntFrom(int64(user.Telephone)),
					DefaultPayment: null.IntFrom(user.DefaultPayment.Int64),
				}
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.User{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:   "InvalidID",
			UserID: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusBadRequest, rsp.StatusCode)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			store := mockdb.NewMockStore(ctrl)
			worker := mockwk.NewMockTaskDistributor(ctrl)
			ik := mockik.NewMockImageKitManagement(ctrl)
			mailSender := mockemail.NewMockEmailSender(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store, worker, ik, mailSender)
			// //recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/usr/v1/users/%d", tc.UserID)
			request, err := http.NewRequest(fiber.MethodPut, url, bytes.NewReader(data))
			require.NoError(t, err)
			request.Header.Set("Content-Type", "application/json")

			tc.setupAuth(t, request, server.userTokenMaker)

			rsp, err := server.router.Test(request, -1)
			require.NoError(t, err)

			tc.checkResponse(t, rsp)
		})
	}
}

func TestListUsersAPI(t *testing.T) {
	n := 5
	users := make([]db.User, n)
	for i := 0; i < n; i++ {
		users[i], _ = randomUser(t)
	}
	admin, _ := randomSuperAdmin(t)

	type Query struct {
		pageID   int
		pageSize int
	}

	testCases := []struct {
		name          string
		query         Query
		AdminID       int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rsp *http.Response)
	}{
		{
			name:    "OK",
			AdminID: admin.ID,
			query: Query{
				pageID:   1,
				pageSize: n,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListUsersParams{
					Limit:  int32(n),
					Offset: 0,
				}

				store.EXPECT().
					ListUsers(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(users, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchUsers(t, rsp.Body, users)
			},
		},
		{
			name:    "Unauthorized",
			AdminID: admin.ID,
			query: Query{
				pageID:   1,
				pageSize: n,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, false, time.Minute)

			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListUsers(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusForbidden, rsp.StatusCode)
			},
		},
		{
			name:    "No authorization",
			AdminID: admin.ID,
			query: Query{
				pageID:   1,
				pageSize: n,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListUsers(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:    "InternalError",
			AdminID: admin.ID,
			query: Query{
				pageID:   1,
				pageSize: n,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListUsers(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.User{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name: "InvalidPageID",
			query: Query{
				pageID:   -1,
				pageSize: n,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListUsers(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusBadRequest, rsp.StatusCode)
			},
		},
		{
			name:    "InvalidPageSize",
			AdminID: admin.ID,
			query: Query{
				pageID:   1,
				pageSize: 100000,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListUsers(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusBadRequest, rsp.StatusCode)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			store := mockdb.NewMockStore(ctrl)
			worker := mockwk.NewMockTaskDistributor(ctrl)
			ik := mockik.NewMockImageKitManagement(ctrl)
			mailSender := mockemail.NewMockEmailSender(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store, worker, ik, mailSender)
			// //recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/admin/v1/admins/%d/users", tc.AdminID)
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			request.Header.Set("Content-Type", "application/json")

			// Add query parameters to request URL
			q := request.URL.Query()
			q.Add("page_id", fmt.Sprintf("%d", tc.query.pageID))
			q.Add("page_size", fmt.Sprintf("%d", tc.query.pageSize))
			request.URL.RawQuery = q.Encode()

			tc.setupAuth(t, request, server.adminTokenMaker)
			rsp, err := server.router.Test(request, -1)
			require.NoError(t, err)

			tc.checkResponse(t, rsp)
		})
	}
}

func TestDeleteUserAPI(t *testing.T) {
	user, _ := randomUser(t)
	admin, _ := randomSuperAdmin(t)

	testCases := []struct {
		name          string
		UserID        int64
		AdminID       int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rsp *http.Response)
	}{
		{
			name:    "OK",
			UserID:  user.ID,
			AdminID: admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteUser(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:    "Unauthorized",
			UserID:  user.ID,
			AdminID: admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:    "NotFound",
			UserID:  user.ID,
			AdminID: admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
			},
		},
		{
			name:    "InternalError",
			UserID:  user.ID,
			AdminID: admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:    "InvalidID",
			UserID:  0,
			AdminID: admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteUser(gomock.Any(), gomock.Any()).
					Times(0)

			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusBadRequest, rsp.StatusCode)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t) // no need to call defer ctrl.finish() in 1.6V

			store := mockdb.NewMockStore(ctrl)
			worker := mockwk.NewMockTaskDistributor(ctrl)
			ik := mockik.NewMockImageKitManagement(ctrl)
			emailSender := mockemail.NewMockEmailSender(ctrl)

			// build stubs
			tc.buildStub(store)

			// start test server and send request
			server := newTestServer(t, store, worker, ik, emailSender)
			//recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/admin/v1/admins/%d/users/%d", tc.AdminID, tc.UserID)
			request, err := http.NewRequest(fiber.MethodDelete, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.adminTokenMaker)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			//check response
			tc.checkResponse(t, rsp)
		})

	}

}

func randomUser(t *testing.T) (user db.User, password string) {
	password = util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	user = db.User{
		ID:       util.RandomMoney(),
		Username: util.RandomUser(),
		Password: hashedPassword,
		// Telephone:      int32(util.RandomInt(910000000, 929999999)),
		IsBlocked:      false,
		Email:          util.RandomEmail(),
		DefaultPayment: null.IntFrom(util.RandomMoney()),
	}
	return
}

func randomUserLogin(t *testing.T) (user db.GetUserByEmailRow, password string) {
	password = util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	user = db.GetUserByEmailRow{
		ID:       util.RandomMoney(),
		Username: util.RandomUser(),
		Password: hashedPassword,
		// Telephone: int32(util.RandomInt(910000000, 929999999)),
		IsEmailVerified: true,
		IsBlocked:       false,
		Email:           util.RandomEmail(),
	}
	return
}

func randomUserLoginNotVerified(t *testing.T) (user db.GetUserByEmailRow, password string) {
	password = util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	user = db.GetUserByEmailRow{
		ID:       util.RandomMoney(),
		Username: util.RandomUser(),
		Password: hashedPassword,
		// Telephone: int32(util.RandomInt(910000000, 929999999)),
		IsEmailVerified: false,
		IsBlocked:       false,
		Email:           util.RandomEmail(),
	}
	return
}

func randomUserLogout(t *testing.T) (user db.UserSession, userLog db.GetUserByEmailRow) {
	uuid1, err := uuid.NewRandom()
	require.NoError(t, err)

	userLog, _ = randomUserLogin(t)

	user = db.UserSession{
		ID:           uuid1,
		UserID:       userLog.ID,
		RefreshToken: util.RandomString(5),
		UserAgent:    util.RandomString(5),
		ClientIp:     util.RandomString(5),
		IsBlocked:    false,
	}
	return
}

func randomUserWithCartAndWishList(t *testing.T) (user db.CreateUserWithCartAndWishListRow, password string) {
	password = util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	user = db.CreateUserWithCartAndWishListRow{
		ID:       util.RandomMoney(),
		Username: util.RandomUser(),
		Password: hashedPassword,
		// Telephone: int32(util.RandomInt(910000000, 929999999)),

		Email: util.RandomEmail(),
	}
	return
}

func randomSignUpV2User() (user db.SignUpTxResult, password string) {
	password = util.RandomString(6)
	// hashedPassword, err := util.HashPassword(password)
	// require.NoError(t, err)

	user = db.SignUpTxResult{
		ID:       util.RandomMoney(),
		Username: util.RandomUser(),
		// Password: hashedPassword,
		// Telephone: int32(util.RandomInt(910000000, 929999999)),

		Email:           util.RandomEmail(),
		IsBlocked:       false,
		IsEmailVerified: false,
		SecretCode:      util.GenerateOTP(),
	}
	return
}
func randomVerifyEmail(signUpUser db.SignUpTxResult) (user db.GetVerifyEmailByEmailRow) {

	user = db.GetVerifyEmailByEmailRow{
		ID:              util.RandomMoney(),
		UserID:          null.IntFrom(signUpUser.ID),
		IsUsed:          false,
		ExpiredAt:       time.Now().Add(-time.Hour),
		Email:           signUpUser.Email,
		IsBlocked:       signUpUser.IsBlocked,
		IsEmailVerified: signUpUser.IsEmailVerified,
		SecretCode:      util.GenerateOTP(),
	}
	return
}
func randomVerifyPasswordResetOTP(signUpUser db.GetUserByEmailRow) (user db.GetVerifyEmailByEmailRow) {

	user = db.GetVerifyEmailByEmailRow{
		ID:              util.RandomMoney(),
		UserID:          null.IntFrom(signUpUser.ID),
		IsUsed:          false,
		ExpiredAt:       time.Now().Add(-time.Hour),
		Email:           signUpUser.Email,
		IsBlocked:       signUpUser.IsBlocked,
		IsEmailVerified: signUpUser.IsEmailVerified,
		SecretCode:      util.GenerateOTP(),
	}
	return
}
func randomResetPasswordOTP(signUpUser db.GetUserByEmailRow) (user db.GetResetPasswordsByEmailRow) {

	user = db.GetResetPasswordsByEmailRow{
		ID:              util.RandomMoney(),
		UserID:          signUpUser.ID,
		IsUsed:          false,
		ExpiredAt:       time.Now().Add(-time.Hour),
		Email:           signUpUser.Email,
		IsBlockedUser:   signUpUser.IsBlocked,
		IsEmailVerified: signUpUser.IsEmailVerified,
		SecretCode:      util.GenerateOTP(),
	}
	return
}

func randomResetPassword(signUpUser db.GetUserByEmailRow) (user db.ResetPassword) {

	user = db.ResetPassword{
		ID:         util.RandomMoney(),
		UserID:     signUpUser.ID,
		IsUsed:     false,
		ExpiredAt:  time.Now().Add(-time.Hour),
		SecretCode: util.GenerateOTP(),
	}
	return
}

func randomUpdateVerifyEmail(signUpUser db.SignUpTxResult, verifyEmail db.GetVerifyEmailByEmailRow) (user db.UpdateVerifyEmailRow) {

	user = db.UpdateVerifyEmailRow{
		ID:              verifyEmail.UserID.Int64,
		Username:        signUpUser.Username,
		Email:           signUpUser.Email,
		IsBlocked:       signUpUser.IsBlocked,
		IsEmailVerified: signUpUser.IsEmailVerified,
		WishListID:      util.RandomMoney(),
		ShoppingCartID:  util.RandomMoney(),
	}
	return
}
func randomUpdateVerifyPasswordResetOTP(signUpUser db.GetUserByEmailRow, verifyEmail db.GetVerifyEmailByEmailRow) (user db.UpdateVerifyEmailRow) {

	user = db.UpdateVerifyEmailRow{
		ID:              verifyEmail.UserID.Int64,
		Username:        signUpUser.Username,
		Email:           signUpUser.Email,
		IsBlocked:       signUpUser.IsBlocked,
		IsEmailVerified: signUpUser.IsEmailVerified,
		WishListID:      util.RandomMoney(),
		ShoppingCartID:  util.RandomMoney(),
	}
	return
}
func randomUpdateResetPasswordOTP(verifyEmail db.GetResetPasswordsByEmailRow) (user db.ResetPassword) {

	user = db.ResetPassword{
		ID:         verifyEmail.ID,
		UserID:     verifyEmail.UserID,
		SecretCode: verifyEmail.SecretCode,
		IsUsed:     verifyEmail.IsUsed,
	}
	return
}

func randomSuperAdmin(t *testing.T) (admin db.Admin, password string) {
	password = util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	admin = db.Admin{
		ID:       util.RandomMoney(),
		Username: util.RandomUser(),
		Email:    util.RandomEmail(),
		Password: hashedPassword,
		Active:   true,
		TypeID:   1,
	}
	return
}

func requireBodyMatchUser(t *testing.T, body io.Reader, user db.User) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotUser db.User
	err = json.Unmarshal(data, &gotUser)

	require.NoError(t, err)
	require.Equal(t, user.Username, gotUser.Username)
	// // require.Equal(t, user.Telephone, gotUser.Telephone)
	require.Equal(t, user.Email, gotUser.Email)
	require.Empty(t, gotUser.Password)
}

func requireBodyMatchUserForCreate(t *testing.T, body io.Reader, user createUserResponse) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotUser createUserResponse
	err = json.Unmarshal(data, &gotUser)

	require.NoError(t, err)
	require.Equal(t, user.User, gotUser.User)
	// require.Empty(t, gotUser.Password)
}
func requireBodyMatchUserForSignUp(t *testing.T, body io.Reader, user userResponse) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotUser userResponse
	err = json.Unmarshal(data, &gotUser)

	require.NoError(t, err)
	require.Equal(t, user.UserID, gotUser.UserID)
	require.Equal(t, user.Username, gotUser.Username)
	require.Equal(t, user.Email, gotUser.Email)
	require.Equal(t, user.IsBlocked, gotUser.IsBlocked)
	require.Equal(t, user.IsEmailVerified, gotUser.IsEmailVerified)
	// require.Empty(t, gotUser.Password)
}
func requireBodyMatchResendOTP(t *testing.T, body io.Reader, message map[string]interface{}) {
	// match body of fiber map[string]interface{}
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	// var gotUser map[string]interface{}
	err = json.Unmarshal(data, &message)

	require.NoError(t, err)
	require.Equal(t, "OTP sent successfully", message["message"])
}

// func requireBodyMatchUserForCreate(t *testing.T, body io.Reader, user db.CreateUserWithCartAndWishListRow) {
// 	data, err := io.ReadAll(body)
// 	require.NoError(t, err)

// 	var gotUser db.CreateUserWithCartAndWishListRow
// 	err = json.Unmarshal(data, &gotUser)

// 	require.NoError(t, err)
// 	require.Equal(t, user.Username, gotUser.Username)
// require.Equal(t, user.Telephone, gotUser.Telephone)
// 	require.Equal(t, user.Email, gotUser.Email)
// 	require.Empty(t, gotUser.Password)
// }

func requireBodyMatchUsers(t *testing.T, body io.ReadCloser, users []db.User) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotUsers []db.User
	err = json.Unmarshal(data, &gotUsers)
	require.NoError(t, err)
	require.Equal(t, users, gotUsers)
}
