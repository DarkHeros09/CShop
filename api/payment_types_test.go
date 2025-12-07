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
	"github.com/guregu/null/v6"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestAdminCreatePaymentTypeAPI(t *testing.T) {
	admin, _ := randomPaymentTypeSuperAdmin(t)
	paymentType := randomPaymentType()

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
				"value":     paymentType.Value,
				"is_active": paymentType.IsActive,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminCreatePaymentTypeParams{
					AdminID:  admin.ID,
					Value:    paymentType.Value,
					IsActive: paymentType.IsActive,
				}

				store.EXPECT().
					AdminCreatePaymentType(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(paymentType, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchPaymentType(t, rsp.Body, paymentType)
			},
		},
		{
			name:    "NoAuthorization",
			AdminID: admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminCreatePaymentTypeParams{
					AdminID:  admin.ID,
					Value:    paymentType.Value,
					IsActive: paymentType.IsActive,
				}

				store.EXPECT().
					AdminCreatePaymentType(gomock.Any(), gomock.Eq(arg)).
					Times(0)

			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:    "Unauthorized",
			AdminID: admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, false, time.Minute)
			},
			body: fiber.Map{
				"value":     paymentType.Value,
				"is_active": paymentType.IsActive,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminCreatePaymentTypeParams{
					AdminID:  admin.ID,
					Value:    paymentType.Value,
					IsActive: paymentType.IsActive,
				}

				store.EXPECT().
					AdminCreatePaymentType(gomock.Any(), gomock.Eq(arg)).
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
				"value":     paymentType.Value,
				"is_active": paymentType.IsActive,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					AdminCreatePaymentType(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:    "InvalidID",
			AdminID: admin.ID,
			body: fiber.Map{
				"id": "invalid",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					AdminCreatePaymentType(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/admin/v1/admins/%d/payment-types", tc.AdminID)
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

func TestListPaymentTypeAPI(t *testing.T) {
	n := 5
	paymentTypes := make([]*db.PaymentType, n)
	user, _ := randomPTUser(t)
	paymentType1 := createRandomPaymentType()
	paymentType2 := createRandomPaymentType()

	paymentTypes = append(paymentTypes, paymentType1, paymentType2)

	testCases := []struct {
		name          string
		UserID        int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(rsp *http.Response)
	}{
		{
			name:   "OK",
			UserID: user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					ListPaymentTypes(gomock.Any()).
					Times(1).
					Return(paymentTypes, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchPaymentTypes(t, rsp.Body, paymentTypes)
			},
		},
		{
			name:   "InternalError",
			UserID: user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListPaymentTypes(gomock.Any()).
					Times(1).
					Return(nil, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
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

			url := fmt.Sprintf("/usr/v1/users/%d/payment-types", tc.UserID)
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.userTokenMaker)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
		})
	}
}

func TestAdminListPaymentTypeAPI(t *testing.T) {
	admin, _ := randomPaymentTypeSuperAdmin(t)
	n := 5
	paymentTypes := make([]*db.PaymentType, n)
	paymentType1 := createRandomPaymentType()
	paymentType2 := createRandomPaymentType()

	paymentTypes = append(paymentTypes, paymentType1, paymentType2)

	testCases := []struct {
		name          string
		AdminID       int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(rsp *http.Response)
	}{
		{
			name:    "OK",
			AdminID: admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					AdminListPaymentTypes(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(paymentTypes, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchPaymentTypes(t, rsp.Body, paymentTypes)
			},
		},
		{
			name:    "InternalError",
			AdminID: admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					AdminListPaymentTypes(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(nil, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
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

			url := fmt.Sprintf("/admin/v1/admins/%d/payment-types", tc.AdminID)
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.adminTokenMaker)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
		})
	}
}

func TestUpdatePaymentTypeAPI(t *testing.T) {
	admin, _ := randomPaymentTypeSuperAdmin(t)
	paymentType := randomPaymentType()

	testCases := []struct {
		name          string
		body          fiber.Map
		paymentTypeID int64
		AdminID       int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rsp *http.Response)
	}{
		{
			name:          "OK",
			paymentTypeID: paymentType.ID,
			AdminID:       admin.ID,
			body: fiber.Map{
				"value":     paymentType.Value,
				"is_active": paymentType.IsActive,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminUpdatePaymentTypeParams{
					AdminID:  admin.ID,
					Value:    null.StringFromPtr(&paymentType.Value),
					IsActive: null.BoolFromPtr(&paymentType.IsActive),
					ID:       paymentType.ID,
				}

				store.EXPECT().
					AdminUpdatePaymentType(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(paymentType, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:          "Unauthorized",
			paymentTypeID: paymentType.ID,
			AdminID:       admin.ID,
			body: fiber.Map{
				"value":     paymentType.Value,
				"is_active": paymentType.IsActive,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, false, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminUpdatePaymentTypeParams{
					AdminID:  admin.ID,
					Value:    null.StringFromPtr(&paymentType.Value),
					IsActive: null.BoolFromPtr(&paymentType.IsActive),
					ID:       paymentType.ID,
				}

				store.EXPECT().
					AdminUpdatePaymentType(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:          "NoAuthorization",
			paymentTypeID: paymentType.ID,
			AdminID:       admin.ID,
			body: fiber.Map{
				"value":     paymentType.Value,
				"is_active": paymentType.IsActive,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminUpdatePaymentTypeParams{
					AdminID:  admin.ID,
					Value:    null.StringFromPtr(&paymentType.Value),
					IsActive: null.BoolFromPtr(&paymentType.IsActive),
					ID:       paymentType.ID,
				}

				store.EXPECT().
					AdminUpdatePaymentType(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:          "InternalError",
			paymentTypeID: paymentType.ID,
			AdminID:       admin.ID,
			body: fiber.Map{
				"value":     paymentType.Value,
				"is_active": paymentType.IsActive,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminUpdatePaymentTypeParams{
					AdminID:  admin.ID,
					Value:    null.StringFromPtr(&paymentType.Value),
					IsActive: null.BoolFromPtr(&paymentType.IsActive),
					ID:       paymentType.ID,
				}
				store.EXPECT().
					AdminUpdatePaymentType(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(nil, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:          "InvalidID",
			paymentTypeID: 0,
			AdminID:       admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					AdminUpdatePaymentType(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/admin/v1/admins/%d/payment-types/%d", tc.AdminID, tc.paymentTypeID)
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

// func TestAdminDeletePaymentTypeAPI(t *testing.T) {
// 	admin, _ := randomPaymentTypeSuperAdmin(t)
// 	paymentType := randomPaymentType()

// 	testCases := []struct {
// 		name          string
// 		paymentTypeID int64
// 		AdminID       int64
// 		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
// 		buildStub     func(store *mockdb.MockStore)
// 		checkResponse func(t *testing.T, rsp *http.Response)
// 	}{
// 		{
// 			name:          "OK",
// 			paymentTypeID: paymentType.ID,
// 			AdminID:       admin.ID,
// 			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
// 				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
// 			},
// 			buildStub: func(store *mockdb.MockStore) {

// 				arg := db.AdminDeletePaymentTypeParams{
// 					AdminID: admin.ID,
// 					ID:      paymentType.ID,
// 				}
// 				store.EXPECT().
// 					AdminDeletePaymentType(gomock.Any(), gomock.Eq(arg)).
// 					Times(1).
// 					Return(nil)
// 			},
// 			checkResponse: func(t *testing.T, rsp *http.Response) {
// 				require.Equal(t, http.StatusOK, rsp.StatusCode)
// 			},
// 		},
// 		{
// 			name:          "Unauthorized",
// 			paymentTypeID: paymentType.ID,
// 			AdminID:       admin.ID,
// 			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
// 				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, 2, admin.Active, time.Minute)
// 			},
// 			buildStub: func(store *mockdb.MockStore) {
// 				arg := db.AdminDeletePaymentTypeParams{
// 					AdminID: admin.ID,
// 					ID:      paymentType.ID,
// 				}
// 				store.EXPECT().
// 					AdminDeletePaymentType(gomock.Any(), gomock.Eq(arg)).
// 					Times(0)
// 			},
// 			checkResponse: func(t *testing.T, rsp *http.Response) {
// 				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
// 			},
// 		},
// 		{
// 			name:          "NoAuthorization",
// 			paymentTypeID: paymentType.ID,
// 			AdminID:       admin.ID,
// 			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
// 			},
// 			buildStub: func(store *mockdb.MockStore) {
// 				arg := db.AdminDeletePaymentTypeParams{
// 					AdminID: admin.ID,
// 					ID:      paymentType.ID,
// 				}
// 				store.EXPECT().
// 					AdminDeletePaymentType(gomock.Any(), gomock.Eq(arg)).
// 					Times(0)
// 			},
// 			checkResponse: func(t *testing.T, rsp *http.Response) {
// 				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
// 			},
// 		},
// 		{
// 			name:          "NotFound",
// 			paymentTypeID: paymentType.ID,
// 			AdminID:       admin.ID,
// 			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
// 				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
// 			},
// 			buildStub: func(store *mockdb.MockStore) {
// 				arg := db.AdminDeletePaymentTypeParams{
// 					AdminID: admin.ID,
// 					ID:      paymentType.ID,
// 				}
// 				store.EXPECT().
// 					AdminDeletePaymentType(gomock.Any(), gomock.Eq(arg)).
// 					Times(1).
// 					Return(pgx.ErrNoRows)
// 			},
// 			checkResponse: func(t *testing.T, rsp *http.Response) {
// 				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
// 			},
// 		},
// 		{
// 			name:          "InternalError",
// 			paymentTypeID: paymentType.ID,
// 			AdminID:       admin.ID,
// 			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
// 				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
// 			},
// 			buildStub: func(store *mockdb.MockStore) {
// 				arg := db.AdminDeletePaymentTypeParams{
// 					AdminID: admin.ID,
// 					ID:      paymentType.ID,
// 				}
// 				store.EXPECT().
// 					AdminDeletePaymentType(gomock.Any(), gomock.Eq(arg)).
// 					Times(1).
// 					Return(pgx.ErrTxClosed)
// 			},
// 			checkResponse: func(t *testing.T, rsp *http.Response) {
// 				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
// 			},
// 		},
// 		{
// 			name:          "InvalidID",
// 			paymentTypeID: 0,
// 			AdminID:       admin.ID,
// 			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
// 				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
// 			},
// 			buildStub: func(store *mockdb.MockStore) {
// 				store.EXPECT().
// 					AdminDeletePaymentType(gomock.Any(), gomock.Any()).
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
// 			worker := mockwk.NewMockTaskDistributor(ctrl)
// 			ik := mockik.NewMockImageKitManagement(ctrl)
// 			mailSender := mockemail.NewMockEmailSender(ctrl)

// 			// build stubs
// 			tc.buildStub(store)

// 			// start test server and send request
// 			server := newTestServer(t, store, worker, ik, mailSender)
// 			//recorder := httptest.NewRecorder()

// 			url := fmt.Sprintf("/admin/v1/admins/%d/payment-types/%d", tc.AdminID, tc.paymentTypeID)
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

func randomPTUser(t *testing.T) (user db.User, password string) {
	password = util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	user = db.User{
		ID:       util.RandomMoney(),
		Username: util.RandomUser(),
		Email:    util.RandomEmail(),
		Password: hashedPassword,
		// Telephone: int32(util.RandomInt(910000000, 929999999)),
	}
	return
}

func randomPaymentType() *db.PaymentType {
	return &db.PaymentType{
		ID:       util.RandomInt(1, 1000),
		Value:    util.RandomUser(),
		IsActive: util.RandomBool(),
	}
}

func requireBodyMatchPaymentTypes(t *testing.T, body io.ReadCloser, paymentTypes []*db.PaymentType) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotPaymentTypes []*db.PaymentType
	err = json.Unmarshal(data, &gotPaymentTypes)
	require.NoError(t, err)
	require.Equal(t, paymentTypes, gotPaymentTypes)
}

func requireBodyMatchPaymentType(t *testing.T, body io.ReadCloser, paymentType *db.PaymentType) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotPaymentType *db.PaymentType
	err = json.Unmarshal(data, &gotPaymentType)
	require.NoError(t, err)
	require.Equal(t, paymentType, gotPaymentType)
}

func randomPaymentTypeSuperAdmin(t *testing.T) (admin *db.Admin, password string) {
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
