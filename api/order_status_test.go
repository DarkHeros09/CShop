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

func TestCreateOrderStatusAPI(t *testing.T) {
	// user, _ := randomOSUser(t)
	admin, _ := randomOrderStatusSuperAdmin(t)
	orderStatus := createRandomOrderStatusForStatus()

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
				"status": orderStatus.Status,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					CreateOrderStatus(gomock.Any(), gomock.Eq(orderStatus.Status)).
					Times(1).
					Return(orderStatus, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchOrderStatus(t, rsp.Body, orderStatus)
			},
		},
		{
			name:    "NoAuthorization",
			AdminID: admin.ID,
			body: fiber.Map{
				"status": orderStatus.Status,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, false, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateOrderStatus(gomock.Any(), gomock.Any()).
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
				"status": orderStatus.Status,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					CreateOrderStatus(gomock.Any(), gomock.Eq(orderStatus.Status)).
					Times(1).
					Return(db.OrderStatus{}, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:    "InvalidUserID",
			AdminID: 0,
			body: fiber.Map{
				"status": orderStatus.Status,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, 0, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateOrderStatus(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/admin/v1/admins/%d/order-status", tc.AdminID)
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

func TestGetOrderStatusAPI(t *testing.T) {
	user, _ := randomOSUser(t)
	orderStatus := createRandomOrderStatusForGet()

	testCases := []struct {
		name          string
		StatusID      int64
		UserID        int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rsp *http.Response)
	}{
		{
			name:     "OK",
			StatusID: orderStatus.ID,
			UserID:   user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				// arg := db.GetOrderStatusByUserIDParams{
				// 	ID:     orderStatus.ID,
				// 	UserID: user.ID,
				// }
				store.EXPECT().
					GetOrderStatus(gomock.Any(), gomock.Eq(orderStatus.ID)).
					Times(1).
					Return(orderStatus, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchOrderStatusForGet(t, rsp.Body, orderStatus)
			},
		},
		{
			name:     "NoAuthorization",
			StatusID: orderStatus.ID,
			UserID:   user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetOrderStatus(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:     "NotFound",
			StatusID: orderStatus.ID,
			UserID:   user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				// arg := db.GetOrderStatusByUserIDParams{
				// 	ID:     orderStatus.ID,
				// 	UserID: user.ID,
				// }
				store.EXPECT().
					GetOrderStatus(gomock.Any(), gomock.Eq(orderStatus.ID)).
					Times(1).
					Return(db.OrderStatus{}, pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
			},
		},
		{
			name:     "InternalError",
			StatusID: orderStatus.ID,
			UserID:   user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				// arg := db.GetOrderStatusByUserIDParams{
				// 	ID:     orderStatus.ID,
				// 	UserID: user.ID,
				// }
				store.EXPECT().
					GetOrderStatus(gomock.Any(), gomock.Eq(orderStatus.ID)).
					Times(1).
					Return(db.OrderStatus{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:     "InvalidID",
			StatusID: 0,
			UserID:   user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetOrderStatus(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/usr/v1/users/%d/order-status/%d", tc.UserID, tc.StatusID)
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.userTokenMaker)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(t, rsp)
		})

	}

}

func TestListOrderStatusAPI(t *testing.T) {
	n := 5
	orderStatuses := make([]db.ListOrderStatusesByUserIDRow, n)
	user, _ := randomOSUser(t)
	orderStatus1 := createRandomOrderStatusForList(user)
	orderStatus2 := createRandomOrderStatusForList(user)
	orderStatus3 := createRandomOrderStatusForList(user)

	orderStatuses = append(orderStatuses, orderStatus1, orderStatus2, orderStatus3)

	type Query struct {
		pageID   int
		pageSize int
	}

	testCases := []struct {
		name          string
		UserID        int64
		query         Query
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(rsp *http.Response)
	}{
		{
			name:   "OK",
			UserID: user.ID,
			query: Query{
				pageID:   1,
				pageSize: n,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListOrderStatusesByUserIDParams{
					Limit:  int32(n),
					Offset: 0,
					UserID: user.ID,
				}

				store.EXPECT().
					ListOrderStatusesByUserID(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(orderStatuses, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchOrderStatuses(t, rsp.Body, orderStatuses)
			},
		},
		{
			name:   "InternalError",
			UserID: user.ID,
			query: Query{
				pageID:   1,
				pageSize: n,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListOrderStatusesByUserID(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.ListOrderStatusesByUserIDRow{}, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:   "InvalidPageID",
			UserID: user.ID,
			query: Query{
				pageID:   -1,
				pageSize: n,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListOrderStatusesByUserID(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusBadRequest, rsp.StatusCode)
			},
		},
		{
			name:   "InvalidPageSize",
			UserID: user.ID,
			query: Query{
				pageID:   1,
				pageSize: 100000,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListOrderStatusesByUserID(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/usr/v1/users/%d/order-status", tc.UserID)
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			// Add query parameters to request URL
			q := request.URL.Query()
			q.Add("page_id", fmt.Sprintf("%d", tc.query.pageID))
			q.Add("page_size", fmt.Sprintf("%d", tc.query.pageSize))
			request.URL.RawQuery = q.Encode()

			tc.setupAuth(t, request, server.userTokenMaker)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
		})
	}
}

func TestAdminListOrderStatusAPI(t *testing.T) {
	n := 3
	orderStatuses := make([]db.OrderStatus, n)
	admin, _ := randomOrderStatusSuperAdmin(t)
	orderStatus1 := createRandomOrderStatus()
	orderStatus2 := createRandomOrderStatus()
	orderStatus3 := createRandomOrderStatus()

	orderStatuses = append(orderStatuses, orderStatus1, orderStatus2, orderStatus3)

	testCases := []struct {
		name          string
		Admin         int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(rsp *http.Response)
	}{
		{
			name:  "OK",
			Admin: admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					AdminListOrderStatuses(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(orderStatuses, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchOrderStatusesForAdmin(t, rsp.Body, orderStatuses)
			},
		},
		{
			name:  "InternalError",
			Admin: admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					AdminListOrderStatuses(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.OrderStatus{}, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:  "InvalidAdminID",
			Admin: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					AdminListOrderStatuses(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/admin/v1/admins/%d/order-status", tc.Admin)
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

func TestUpdateOrderStatusAPI(t *testing.T) {
	admin, _ := randomOrderStatusSuperAdmin(t)
	orderStatus := createRandomOrderStatusForStatus()

	testCases := []struct {
		name          string
		StatusID      int64
		AdminID       int64
		body          fiber.Map
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rsp *http.Response)
	}{
		{
			name:     "OK",
			StatusID: orderStatus.ID,
			AdminID:  admin.ID,
			body: fiber.Map{
				"status": "new status",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.UpdateOrderStatusParams{
					Status: null.StringFrom("new status"),
					ID:     orderStatus.ID,
				}

				store.EXPECT().
					UpdateOrderStatus(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(orderStatus, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:     "NoAuthorization",
			StatusID: orderStatus.ID,
			AdminID:  admin.ID,
			body: fiber.Map{
				"status": "new status",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateOrderStatus(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:     "InternalError",
			StatusID: orderStatus.ID,
			AdminID:  admin.ID,
			body: fiber.Map{
				"status": "new status",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateOrderStatusParams{
					Status: null.StringFrom("new status"),
					ID:     orderStatus.ID,
				}

				store.EXPECT().
					UpdateOrderStatus(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.OrderStatus{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:     "InvalidUserID",
			StatusID: orderStatus.ID,
			AdminID:  0,
			body: fiber.Map{
				"status": "new status",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, 0, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateOrderStatus(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/admin/v1/admins/%d/order-status/%d", tc.AdminID, tc.StatusID)
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

func TestDeleteOrderStatusAPI(t *testing.T) {
	admin, _ := randomOrderStatusSuperAdmin(t)
	orderStatus := createRandomOrderStatusForStatus()
	userID := util.RandomMoney()

	testCases := []struct {
		name          string
		StatusID      int64
		AdminID       int64
		UserD         int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rsp *http.Response)
	}{
		{
			name:     "OK",
			StatusID: orderStatus.ID,
			AdminID:  admin.ID,
			UserD:    userID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {

				store.EXPECT().
					DeleteOrderStatus(gomock.Any(), gomock.Eq(orderStatus.ID)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:     "NotFound",
			StatusID: orderStatus.ID,
			AdminID:  admin.ID,
			UserD:    userID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {

				store.EXPECT().
					DeleteOrderStatus(gomock.Any(), gomock.Eq(orderStatus.ID)).
					Times(1).
					Return(pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
			},
		},
		{
			name:     "InternalError",
			StatusID: orderStatus.ID,
			AdminID:  admin.ID,
			UserD:    userID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {

				store.EXPECT().
					DeleteOrderStatus(gomock.Any(), gomock.Eq(orderStatus.ID)).
					Times(1).
					Return(pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:     "InvalidID",
			StatusID: 0,
			AdminID:  admin.ID,
			UserD:    userID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, 0, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteOrderStatus(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/admin/v1/admins/%d/order-status/%d", tc.AdminID, tc.StatusID)
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

func randomOSUser(t *testing.T) (user db.User, password string) {
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

func randomOrderStatusSuperAdmin(t *testing.T) (admin db.Admin, password string) {
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

func createRandomOrderStatusForStatus() (orderStatus db.OrderStatus) {
	orderStatus = db.OrderStatus{
		ID:     util.RandomMoney(),
		Status: util.RandomUser(),
	}
	return
}

func createRandomOrderStatusForGet() (orderStatus db.OrderStatus) {
	orderStatus = db.OrderStatus{
		ID:     util.RandomMoney(),
		Status: util.RandomUser(),
		// UserID: null.IntFrom(user.ID),
	}
	return
}

func createRandomOrderStatusForList(user db.User) (orderStatus db.ListOrderStatusesByUserIDRow) {
	orderStatus = db.ListOrderStatusesByUserIDRow{
		ID:     util.RandomMoney(),
		Status: util.RandomUser(),
		UserID: null.IntFrom(user.ID),
	}
	return
}

func requireBodyMatchOrderStatus(t *testing.T, body io.ReadCloser, orderStatus db.OrderStatus) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotOrderStatus db.OrderStatus
	err = json.Unmarshal(data, &gotOrderStatus)

	require.NoError(t, err)
	require.Equal(t, orderStatus.ID, gotOrderStatus.ID)
	require.Equal(t, orderStatus.Status, gotOrderStatus.Status)
}

func requireBodyMatchOrderStatusForGet(t *testing.T, body io.ReadCloser, orderStatus db.OrderStatus) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotOrderStatus db.OrderStatus
	err = json.Unmarshal(data, &gotOrderStatus)

	require.NoError(t, err)
	require.Equal(t, orderStatus.ID, gotOrderStatus.ID)
	require.Equal(t, orderStatus.Status, gotOrderStatus.Status)
	// require.Equal(t, orderStatus.UserID.Int64, gotOrderStatus.UserID.Int64)
}

func requireBodyMatchOrderStatuses(t *testing.T, body io.ReadCloser, orderStatuses []db.ListOrderStatusesByUserIDRow) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotOrderStatuses []db.ListOrderStatusesByUserIDRow
	err = json.Unmarshal(data, &gotOrderStatuses)
	require.NoError(t, err)
	for i := range gotOrderStatuses {
		require.Equal(t, orderStatuses[i].ID, gotOrderStatuses[i].ID)
		require.Equal(t, orderStatuses[i].Status, gotOrderStatuses[i].Status)
		require.Equal(t, orderStatuses[i].UserID.Int64, gotOrderStatuses[i].UserID.Int64)
	}
}

func requireBodyMatchOrderStatusesForAdmin(t *testing.T, body io.ReadCloser, orderStatuses []db.OrderStatus) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotOrderStatuses []db.OrderStatus
	err = json.Unmarshal(data, &gotOrderStatuses)
	require.NoError(t, err)
	for i := range gotOrderStatuses {
		require.Equal(t, orderStatuses[i].ID, gotOrderStatuses[i].ID)
		require.Equal(t, orderStatuses[i].Status, gotOrderStatuses[i].Status)
	}
}
