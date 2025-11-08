package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
	"github.com/guregu/null/v5"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCreateAppPolicyAPI(t *testing.T) {
	admin, _ := randomAppPolicySuperAdmin(t)
	appPolicy := createRandomAppPolicy()

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
			AdminID: admin.ID,
			body: fiber.Map{
				"policy": appPolicy.Policy.String,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.CreateAppPolicyParams{
					AdminID: admin.ID,
					Policy:  appPolicy.Policy,
				}

				store.EXPECT().
					CreateAppPolicy(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(appPolicy, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchAppPolicy(t, rsp.Body, appPolicy)
			},
		},
		{
			name:    "NoAuthorization",
			AdminID: admin.ID,
			body: fiber.Map{
				"policy": appPolicy.Policy.String,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					CreateAppPolicy(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:    "InternalError",
			AdminID: admin.ID,
			body: fiber.Map{
				"policy": appPolicy.Policy.String,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.CreateAppPolicyParams{
					AdminID: admin.ID,
					Policy:  appPolicy.Policy,
				}

				store.EXPECT().
					CreateAppPolicy(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(nil, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:    "InvalidAdminID",
			AdminID: 0,
			body: fiber.Map{
				"policy": appPolicy.Policy.String,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, 0, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					CreateAppPolicy(gomock.Any(), gomock.Any()).
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
			//recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/admin/v1/admins/%d/app-policy", tc.AdminID)
			request, err := http.NewRequest(fiber.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.adminTokenMaker)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
		})
	}
}

func TestGetAppPolicyAPI(t *testing.T) {
	appPolicy := createRandomAppPolicy()

	testCases := []struct {
		name          string
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(rsp *http.Response)
	}{
		{
			name: "OK",
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					GetAppPolicy(gomock.Any()).
					Times(1).
					Return(appPolicy, nil)

			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchAppPolicy(t, rsp.Body, appPolicy)
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
			//recorder := httptest.NewRecorder()

			url := "/api/v1/app-policy"
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
		})
	}
}

func TestUpdateAppPolicyAPI(t *testing.T) {
	admin, _ := randomAppPolicySuperAdmin(t)
	appPolicy := createRandomAppPolicy()

	testCases := []struct {
		name          string
		AdminID       int64
		ID            int64
		body          fiber.Map
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rsp *http.Response)
	}{
		{
			name:    "OK",
			AdminID: admin.ID,
			ID:      appPolicy.ID,
			body: fiber.Map{
				"policy": appPolicy.Policy.String,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.UpdateAppPolicyParams{
					ID:      appPolicy.ID,
					AdminID: admin.ID,
					Policy:  appPolicy.Policy,
				}

				store.EXPECT().
					UpdateAppPolicy(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(appPolicy, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:    "NoAuthorization",
			AdminID: admin.ID,
			ID:      appPolicy.ID,
			body: fiber.Map{
				"policy": appPolicy.Policy.String,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateAppPolicy(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:    "InternalError",
			AdminID: admin.ID,
			ID:      appPolicy.ID,
			body: fiber.Map{
				"policy": appPolicy.Policy.String,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateAppPolicyParams{
					ID:      appPolicy.ID,
					AdminID: admin.ID,
					Policy:  appPolicy.Policy,
				}

				store.EXPECT().
					UpdateAppPolicy(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(nil, pgx.ErrTxClosed)

			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name: "InvalidID",
			ID:   0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateAppPolicy(gomock.Any(), gomock.Any()).
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
			//recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/admin/v1/admins/%d/app-policy/%d", tc.AdminID, tc.ID)
			request, err := http.NewRequest(fiber.MethodPut, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.adminTokenMaker)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(t, rsp)
		})
	}
}

func TestDeleteAppPolicyAPI(t *testing.T) {
	admin, _ := randomAppPolicySuperAdmin(t)
	appPolicy := createRandomAppPolicy()

	testCases := []struct {
		name          string
		ID            int64
		AdminID       int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rsp *http.Response)
	}{
		{
			name:    "OK",
			AdminID: admin.ID,
			ID:      appPolicy.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteAppPolicyParams{
					AdminID: admin.ID,
					ID:      appPolicy.ID,
				}

				store.EXPECT().
					DeleteAppPolicy(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(appPolicy, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:    "NotFound",
			AdminID: admin.ID,
			ID:      appPolicy.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteAppPolicyParams{
					AdminID: admin.ID,
					ID:      appPolicy.ID,
				}

				store.EXPECT().
					DeleteAppPolicy(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(nil, pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
			},
		},
		{
			name:    "InternalError",
			AdminID: admin.ID,
			ID:      appPolicy.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteAppPolicyParams{
					AdminID: admin.ID,
					ID:      appPolicy.ID,
				}

				store.EXPECT().
					DeleteAppPolicy(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(nil, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:    "InvalidID",
			AdminID: admin.ID,
			ID:      0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteAppPolicy(gomock.Any(), gomock.Any()).
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
			mailSender := mockemail.NewMockEmailSender(ctrl)

			// build stubs
			tc.buildStub(store)

			// start test server and send request
			server := newTestServer(t, store, worker, ik, mailSender)
			//recorder := httptest.NewRecorder()

			// // Marshal body data to JSON
			// data, err := json.Marshal(tc.body)
			// require.NoError(t, err)

			url := fmt.Sprintf("/admin/v1/admins/%d/app-policy/%d", tc.AdminID, tc.ID)
			request, err := http.NewRequest(fiber.MethodDelete, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.adminTokenMaker)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(t, rsp)
		})

	}

}

// func TestDeleteAppPolicyAllByUserAPI(t *testing.T) {
// 	admin, _ := randomAppPolicyUser(t)
// 	shoppingCart := createRandomShoppingCart(t, admin)
// 	appPolicy := createRandomAppPolicy( shoppingCart)

// 	var AppPolicyList []db.AppPolicy
// 	AppPolicyList = append(AppPolicyList, appPolicy...)

// 	testCases := []struct {
// 		name           string
// 		AdminID         int64
// 		ID int64
// 		setupAuth      func(t *testing.T, request *http.Request, tokenMaker token.Maker)
// 		buildStub      func(store *mockdb.MockStore)
// 		checkResponse  func(t *testing.T, rsp *http.Response)
// 	}{
// 		{
// 			name:           "OK",
// 			AdminID:         admin.ID,
// // 			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
// 				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, time.Minute)
// 			},
// 			buildStub: func(store *mockdb.MockStore) {

// 				arg := db.DeleteAppPolicyAllByUserParams{
// 					AdminID:         admin.ID,
// 		// 				}
// 				store.EXPECT().
// 					DeleteAppPolicyAllByUser(gomock.Any(), gomock.Eq(arg)).
// 					Times(1).
// 					Return(appPolicyList, nil)
// 			},
// 			checkResponse: func(t *testing.T, rsp *http.Response) {
// 				require.Equal(t, http.StatusOK, rsp.StatusCode)
// 			},
// 		},
// 		{
// 			name:           "NotFound",
// 			AdminID:         admin.ID,
// // 			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
// 				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, time.Minute)
// 			},
// 			buildStub: func(store *mockdb.MockStore) {
// 				arg := db.DeleteAppPolicyAllByUserParams{
// 					AdminID:         admin.ID,
// 		// 				}
// 				store.EXPECT().
// 					DeleteAppPolicyAllByUser(gomock.Any(), gomock.Eq(arg)).
// 					Times(1).
// 					Return([]db.AppPolicy{}, pgx.ErrNoRows)
// 			},
// 			checkResponse: func(t *testing.T, rsp *http.Response) {
// 				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
// 			},
// 		},
// 		{
// 			name:           "InternalError",
// 			AdminID:         admin.ID,
// // 			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
// 				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, time.Minute)
// 			},
// 			buildStub: func(store *mockdb.MockStore) {
// 				arg := db.DeleteAppPolicyAllByUserParams{
// 					AdminID:         admin.ID,
// 		// 				}
// 				store.EXPECT().
// 					DeleteAppPolicyAllByUser(gomock.Any(), gomock.Eq(arg)).
// 					Times(1).
// 					Return([]db.AppPolicy{}, pgx.ErrTxClosed)
// 			},
// 			checkResponse: func(t *testing.T, rsp *http.Response) {
// 				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
// 			},
// 		},
// 		{
// 			name:           "InvalidID",
// 			AdminID:         0,
// // 			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
// 				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, admin.Username, time.Minute)
// 			},
// 			buildStub: func(store *mockdb.MockStore) {
// 				store.EXPECT().
// 					DeleteAppPolicy(gomock.Any(), gomock.Any()).
// 					Times(0)

// 			},
// 			checkResponse: func(t *testing.T, rsp *http.Response) {
// 				require.Equal(t, http.StatusBadRequest, rsp.StatusCode)
// 			},
// 		},
// 	}

// 	for i := range testCases {
// 		tc := testCases[i]

// 		t.Run(tc.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t) // no need to call defer ctrl.finish() in 1.6V

// 			store := mockdb.NewMockStore(ctrl)
// worker := mockwk.NewMockTaskDistributor(ctrl)

// 			// build stubs
// 			tc.buildStub(store)

// 			// start test server and send request
// 			server := newTestServer(t, store, worker,ik)
// 			//recorder := httptest.NewRecorder()

// 			url := fmt.Sprintf("/admin/v1/admins/%d/app-policy/%d", tc.AdminID, tc.ID)
// 			request, err := http.NewRequest(fiber.MethodDelete, url, nil)
// 			require.NoError(t, err)

// 			tc.setupAuth(t, request, server.adminTokenMaker)
// 			request.Header.Set("Content-Type", "application/json")

// 			rsp, err := server.router.Test(request)
// 			require.NoError(t, err)
// 			tc.checkResponse(t, rsp)
// 		})

// 	}

// }

func randomAppPolicyUser(t *testing.T) (admin db.User, password string) {
	password = util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	admin = db.User{
		ID:       util.RandomMoney(),
		Username: util.RandomUser(),
		Password: hashedPassword,
		// Telephone: int32(util.RandomInt(7, 7)),
		Email: util.RandomEmail(),
	}
	return
}
func randomAppPolicySuperAdmin(t *testing.T) (admin *db.Admin, password string) {
	password = util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	admin = &db.Admin{
		ID:       util.RandomMoney(),
		Username: util.RandomUser(),
		Email:    util.RandomEmail(),
		Password: hashedPassword,
		Active:   true,
		TypeID:   1,
	}
	return
}

func createRandomAppPolicy() (AppPolicy *db.AppPolicy) {
	AppPolicy = &db.AppPolicy{
		ID:     util.RandomMoney(),
		Policy: null.StringFrom(util.RandomString(10)),
	}
	return
}

func requireBodyMatchAppPolicy(t *testing.T, body io.ReadCloser, appPolicy *db.AppPolicy) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotAppPolicy db.AppPolicy
	err = json.Unmarshal(data, &gotAppPolicy)

	require.NoError(t, err)
	require.Equal(t, appPolicy.Policy, gotAppPolicy.Policy)
	require.Equal(t, appPolicy.CreatedAt, gotAppPolicy.CreatedAt)
	require.Equal(t, appPolicy.CreatedAt, gotAppPolicy.CreatedAt)
}
