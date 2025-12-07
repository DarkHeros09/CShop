package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	mockdb "github.com/cshop/v3/db/mock"
	db "github.com/cshop/v3/db/sqlc"
	mockik "github.com/cshop/v3/image/mock"
	mockemail "github.com/cshop/v3/mail/mock"
	"github.com/cshop/v3/token"
	"github.com/cshop/v3/util"
	mockwk "github.com/cshop/v3/worker/mock"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/guregu/null/v6"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestLoginAdminAPI(t *testing.T) {

	admin, password := randomAdminLogin(t)
	adminSession := randomAdminSession(admin)

	testCases := []struct {
		name          string
		body          fiber.Map
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(rsp *http.Response)
	}{
		{
			name: "OK",
			body: fiber.Map{
				"email":    admin.Email,
				"password": password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAdminByEmail(gomock.Any(), gomock.Eq(admin.Email)).
					Times(1).
					Return(admin, nil)

				store.EXPECT().
					CreateAdminSession(gomock.Any(), gomock.Any()).
					Times(1).Return(adminSession, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name: "AdminNotFound",
			body: fiber.Map{
				"email":    "NotFound@NotFound.com",
				"password": password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAdminByEmail(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, pgx.ErrNoRows)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
			},
		},
		{
			name: "IncorrectPassword",
			body: fiber.Map{
				"email":    admin.Email,
				"password": "incorrect",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAdminByEmail(gomock.Any(), gomock.Eq(admin.Email)).
					Times(1).
					Return(nil, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
			},
		},
		{
			name: "InternalError",
			body: fiber.Map{
				"email":    admin.Email,
				"password": password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAdminByEmail(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, pgx.ErrTxClosed)
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
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAdminByEmail(gomock.Any(), gomock.Any()).
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
			tc.buildStubs(store)

			server := newTestServer(t, store, worker, ik, mailSender)
			// //recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/api/v1/admins/login"
			request, err := http.NewRequest(fiber.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			go tc.checkResponse(rsp)
		})
	}
}

func TestLogoutAdminAPI(t *testing.T) {
	adminSession, adminLogin := randomAdminLogout(t)

	testCases := []struct {
		name          string
		body          fiber.Map
		AdminID       int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(rsp *http.Response)
	}{
		{
			name:    "OK",
			AdminID: adminLogin.ID,
			body: fiber.Map{
				"admin_session_id": adminSession.ID,
				"refresh_token":    adminSession.RefreshToken,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, adminLogin.ID, adminLogin.Username, adminLogin.TypeID, adminLogin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.UpdateAdminSessionParams{
					IsBlocked:    null.BoolFrom(true),
					ID:           adminSession.ID,
					AdminID:      adminSession.AdminID,
					RefreshToken: adminSession.RefreshToken,
				}
				store.EXPECT().
					UpdateAdminSession(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(adminSession, nil)

			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:    "NoAuthorization",
			AdminID: adminLogin.ID,
			body: fiber.Map{
				"admin_session_id": adminSession.ID,
				"refresh_token":    adminSession.RefreshToken,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateAdminSession(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:    "InternalError",
			AdminID: adminLogin.ID,
			body: fiber.Map{
				"admin_session_id": adminSession.ID,
				"refresh_token":    adminSession.RefreshToken,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, adminLogin.ID, adminLogin.Username, adminLogin.TypeID, adminLogin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateAdminSession(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:    "InvalidSessionID",
			AdminID: adminLogin.ID,
			body: fiber.Map{
				"admin_session_id": 0,
				"refresh_token":    adminSession.RefreshToken,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, adminLogin.ID, adminLogin.Username, adminLogin.TypeID, adminLogin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateAdminSession(gomock.Any(), gomock.Any()).
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

			// url := "/api/v1/admin/logout"
			url := fmt.Sprintf("/admin/v1/admins/%d/logout", tc.AdminID)
			request, err := http.NewRequest(fiber.MethodDelete, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.adminTokenMaker)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
		})
	}
}

func randomAdminLogin(t *testing.T) (user *db.Admin, password string) {
	password = util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	user = &db.Admin{
		ID:       util.RandomMoney(),
		Username: util.RandomUser(),
		Password: hashedPassword,
		Active:   true,
		Email:    util.RandomEmail(),
		TypeID:   1,
	}
	return
}

func randomAdminLogout(t *testing.T) (user *db.AdminSession, userLog *db.Admin) {
	uuid1, err := uuid.NewRandom()
	require.NoError(t, err)

	userLog, _ = randomAdminLogin(t)

	user = &db.AdminSession{
		ID:           uuid1,
		AdminID:      userLog.ID,
		RefreshToken: util.RandomString(5),
		AdminAgent:   util.RandomString(5),
		ClientIp:     util.RandomString(5),
		IsBlocked:    false,
	}
	return
}

func randomAdminSession(admin *db.Admin) *db.AdminSession {
	uuid1, _ := uuid.NewRandom()
	return &db.AdminSession{
		ID:           uuid1,
		AdminID:      admin.ID,
		RefreshToken: util.RandomString(5),
		AdminAgent:   util.RandomString(5),
		ClientIp:     util.RandomString(5),
		IsBlocked:    false,
	}
}
