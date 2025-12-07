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

func TestCreateShippingMethodAPI(t *testing.T) {
	admin, _ := randomShippingMethodSuperAdmin(t)
	shippingMethod := createRandomShippingMethodForMethod()

	testCases := []struct {
		name          string
		AdminID       int64
		body          fiber.Map
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(rsp *http.Response)
	}{
		{
			name:    "OK",
			AdminID: admin.ID,
			body: fiber.Map{
				"name":  shippingMethod.Name,
				"price": shippingMethod.Price,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.AdminCreateShippingMethodParams{
					AdminID: admin.ID,
					Name:    shippingMethod.Name,
					Price:   shippingMethod.Price,
				}

				store.EXPECT().
					AdminCreateShippingMethod(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(shippingMethod, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchShippingMethod(t, rsp.Body, shippingMethod)
			},
		},
		{
			name:    "NoAuthorization",
			AdminID: admin.ID,
			body: fiber.Map{
				"name":  shippingMethod.Name,
				"price": shippingMethod.Price,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					AdminCreateShippingMethod(gomock.Any(), gomock.Any()).
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
				"name":  shippingMethod.Name,
				"price": shippingMethod.Price,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.AdminCreateShippingMethodParams{
					AdminID: admin.ID,
					Name:    shippingMethod.Name,
					Price:   shippingMethod.Price,
				}

				store.EXPECT().
					AdminCreateShippingMethod(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(nil, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:    "InvalidUserID",
			AdminID: 0,
			body: fiber.Map{
				"name":  0,
				"price": "",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, 0, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					AdminCreateShippingMethod(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/admin/v1/admins/%d/shipping-method", tc.AdminID)
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

func TestGetShippingMethodAPI(t *testing.T) {
	user, _ := randomSMUser(t)
	shippingMethod := createRandomShippingMethodForGet(user)

	testCases := []struct {
		name          string
		ID            int64
		UserID        int64
		body          fiber.Map
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rsp *http.Response)
	}{
		{
			name:   "OK",
			ID:     shippingMethod.ID,
			UserID: user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.GetShippingMethodByUserIDParams{
					ID:     shippingMethod.ID,
					UserID: user.ID,
				}
				store.EXPECT().
					GetShippingMethodByUserID(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(shippingMethod, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchShippingMethodForGet(t, rsp.Body, shippingMethod)
			},
		},
		{
			name:   "NoAuthorization",
			ID:     shippingMethod.ID,
			UserID: user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetShippingMethodByUserID(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:   "NotFound",
			ID:     shippingMethod.ID,
			UserID: user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.GetShippingMethodByUserIDParams{
					ID:     shippingMethod.ID,
					UserID: user.ID,
				}
				store.EXPECT().
					GetShippingMethodByUserID(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(nil, pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
			},
		},
		{
			name:   "InternalError",
			ID:     shippingMethod.ID,
			UserID: user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.GetShippingMethodByUserIDParams{
					ID:     shippingMethod.ID,
					UserID: user.ID,
				}
				store.EXPECT().
					GetShippingMethodByUserID(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(nil, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:   "InvalidID",
			ID:     0,
			UserID: user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetShippingMethodByUserID(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/usr/v1/users/%d/shipping-method/%d", tc.UserID, tc.ID)
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

func TestListShippingMethodAPI(t *testing.T) {
	n := 5
	shippingMethods := make([]*db.ShippingMethod, n)
	user, _ := randomSMUser(t)
	for i := 0; i < len(shippingMethods); i++ {
		shippingMethods[i] = createRandomShippingMethodForList()
	}

	// type Query struct {
	// 	pageID   int
	// 	pageSize int
	// }

	testCases := []struct {
		name string
		// query         Query
		UserID        int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(rsp *http.Response)
	}{
		{
			name:   "OK",
			UserID: user.ID,
			// query: Query{
			// 	pageID:   1,
			// 	pageSize: n,
			// },
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// arg := db.ListShippingMethodsParams{
				// 	Limit:  int32(n),
				// 	Offset: 0,
				// 	UserID: user.ID,
				// }

				store.EXPECT().
					ListShippingMethods(gomock.Any()).
					Times(1).
					Return(shippingMethods, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchShippingMethods(t, rsp.Body, shippingMethods)
			},
		},
		{
			name:   "InternalError",
			UserID: user.ID,
			// query: Query{
			// 	pageID:   1,
			// 	pageSize: n,
			// },
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListShippingMethods(gomock.Any()).
					Times(1).
					Return(nil, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		// {
		// 	name:   "InvalidPageID",
		// 	UserID: user.ID,
		// 	// query: Query{
		// 	// 	pageID:   -1,
		// 	// 	pageSize: n,
		// 	// },
		// 	setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
		// 		addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
		// 	},
		// 	buildStubs: func(store *mockdb.MockStore) {
		// 		store.EXPECT().
		// 			ListShippingMethods(gomock.Any()).
		// 			Times(0)
		// 	},
		// 	checkResponse: func(rsp *http.Response) {
		// 		require.Equal(t, http.StatusBadRequest, rsp.StatusCode)
		// 	},
		// },
		// {
		// 	name:   "InvalidPageSize",
		// 	UserID: user.ID,
		// 	// query: Query{
		// 	// 	pageID:   1,
		// 	// 	pageSize: 100000,
		// 	// },
		// 	setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
		// 		addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
		// 	},
		// 	buildStubs: func(store *mockdb.MockStore) {
		// 		store.EXPECT().
		// 			ListShippingMethods(gomock.Any()).
		// 			Times(0)
		// 	},
		// 	checkResponse: func(rsp *http.Response) {
		// 		require.Equal(t, http.StatusBadRequest, rsp.StatusCode)
		// 	},
		// },
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

			url := fmt.Sprintf("/usr/v1/users/%d/shipping-method", tc.UserID)
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			// Add query parameters to request URL
			// q := request.URL.Query()
			// q.Add("page_id", strconv.Itoa(tc.query.pageID))
			// q.Add("page_size", strconv.Itoa(tc.query.pageSize))
			// request.URL.RawQuery = q.Encode()

			tc.setupAuth(t, request, server.userTokenMaker)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
		})
	}
}

func TestAdminListShippingMethodAPI(t *testing.T) {
	admin, _ := randomShippingMethodSuperAdmin(t)
	n := 5
	shippingMethods := make([]*db.ShippingMethod, n)
	for i := 0; i < len(shippingMethods); i++ {
		shippingMethods[i] = createRandomShippingMethodForList()
	}

	// type Query struct {
	// 	pageID   int
	// 	pageSize int
	// }

	testCases := []struct {
		name string
		// query         Query
		AdminID       int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(rsp *http.Response)
	}{
		{
			name:    "OK",
			AdminID: admin.ID,
			// query: Query{
			// 	pageID:   1,
			// 	pageSize: n,
			// },
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// arg := db.ListShippingMethodsParams{
				// 	Limit:  int32(n),
				// 	Offset: 0,
				// 	AdminID : admin.id,
				// }

				store.EXPECT().
					ListShippingMethods(gomock.Any()).
					Times(1).
					Return(shippingMethods, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchShippingMethods(t, rsp.Body, shippingMethods)
			},
		},
		{
			name:    "InternalError",
			AdminID: admin.ID,
			// query: Query{
			// 	pageID:   1,
			// 	pageSize: n,
			// },
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListShippingMethods(gomock.Any()).
					Times(1).
					Return(nil, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		// {
		// 	name:   "InvalidPageID",
		// 	AdminID : admin.id,
		// 	// query: Query{
		// 	// 	pageID:   -1,
		// 	// 	pageSize: n,
		// 	// },
		// 	setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
		// 		addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
		// 	},
		// 	buildStubs: func(store *mockdb.MockStore) {
		// 		store.EXPECT().
		// 			ListShippingMethods(gomock.Any()).
		// 			Times(0)
		// 	},
		// 	checkResponse: func(rsp *http.Response) {
		// 		require.Equal(t, http.StatusBadRequest, rsp.StatusCode)
		// 	},
		// },
		// {
		// 	name:   "InvalidPageSize",
		// 	AdminID : admin.id,
		// 	// query: Query{
		// 	// 	pageID:   1,
		// 	// 	pageSize: 100000,
		// 	// },
		// 	setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
		// 		addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
		// 	},
		// 	buildStubs: func(store *mockdb.MockStore) {
		// 		store.EXPECT().
		// 			ListShippingMethods(gomock.Any()).
		// 			Times(0)
		// 	},
		// 	checkResponse: func(rsp *http.Response) {
		// 		require.Equal(t, http.StatusBadRequest, rsp.StatusCode)
		// 	},
		// },
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

			url := fmt.Sprintf("/admin/v1/admins/%d/shipping-method", tc.AdminID)
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			// Add query parameters to request URL
			// q := request.URL.Query()
			// q.Add("page_id", strconv.Itoa(tc.query.pageID))
			// q.Add("page_size", strconv.Itoa(tc.query.pageSize))
			// request.URL.RawQuery = q.Encode()

			tc.setupAuth(t, request, server.adminTokenMaker)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
		})
	}
}

func TestUpdateShippingMethodAPI(t *testing.T) {
	admin, _ := randomShippingMethodSuperAdmin(t)
	shippingMethod := createRandomShippingMethodForMethod()

	testCases := []struct {
		name          string
		ID            int64
		AdminID       int64
		body          fiber.Map
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rsp *http.Response)
	}{
		{
			name:    "OK",
			ID:      shippingMethod.ID,
			AdminID: admin.ID,
			body: fiber.Map{
				"name":  "new status",
				"price": shippingMethod.Price,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.AdminUpdateShippingMethodParams{
					AdminID: admin.ID,
					Name:    null.StringFrom("new status"),
					Price:   null.StringFrom(shippingMethod.Price),
					ID:      shippingMethod.ID,
				}

				store.EXPECT().
					AdminUpdateShippingMethod(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(shippingMethod, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:    "NoAuthorization",
			ID:      shippingMethod.ID,
			AdminID: admin.ID,
			body: fiber.Map{
				"name":  "new status",
				"price": shippingMethod.Price,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					AdminUpdateShippingMethod(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:    "InternalError",
			ID:      shippingMethod.ID,
			AdminID: admin.ID,
			body: fiber.Map{
				"name":  "new status",
				"price": shippingMethod.Price,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminUpdateShippingMethodParams{
					AdminID: admin.ID,
					Name:    null.StringFrom("new status"),
					Price:   null.StringFrom(shippingMethod.Price),
					ID:      shippingMethod.ID,
				}

				store.EXPECT().
					AdminUpdateShippingMethod(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(nil, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:    "InvalidUserID",
			ID:      shippingMethod.ID,
			AdminID: 0,
			body: fiber.Map{
				"name":  "new status",
				"price": shippingMethod.Price,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, 0, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					AdminUpdateShippingMethod(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/admin/v1/admins/%d/shipping-method/%d", tc.AdminID, tc.ID)
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

func TestDeleteShippingMethodAPI(t *testing.T) {
	admin, _ := randomShippingMethodSuperAdmin(t)
	shippingMethod := createRandomShippingMethodForMethod()

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
			ID:      shippingMethod.ID,
			AdminID: admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {

				store.EXPECT().
					DeleteShippingMethod(gomock.Any(), gomock.Eq(shippingMethod.ID)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:    "NotFound",
			ID:      shippingMethod.ID,
			AdminID: admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {

				store.EXPECT().
					DeleteShippingMethod(gomock.Any(), gomock.Eq(shippingMethod.ID)).
					Times(1).
					Return(pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
			},
		},
		{
			name:    "InternalError",
			ID:      shippingMethod.ID,
			AdminID: admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {

				store.EXPECT().
					DeleteShippingMethod(gomock.Any(), gomock.Eq(shippingMethod.ID)).
					Times(1).
					Return(pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:    "InvalidID",
			ID:      0,
			AdminID: admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, 0, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteShippingMethod(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/admin/v1/admins/%d/shipping-method/%d", tc.AdminID, tc.ID)
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

func randomSMUser(t *testing.T) (user db.User, password string) {
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

func randomShippingMethodSuperAdmin(t *testing.T) (admin db.Admin, password string) {
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

func createRandomShippingMethodForMethod() (ShippingMethod *db.ShippingMethod) {
	ShippingMethod = &db.ShippingMethod{
		ID:    util.RandomMoney(),
		Name:  util.RandomUser(),
		Price: util.RandomDecimalString(1, 1000),
	}
	return
}

func createRandomShippingMethodForGet(user db.User) (ShippingMethod *db.GetShippingMethodByUserIDRow) {
	ShippingMethod = &db.GetShippingMethodByUserIDRow{
		ID:     util.RandomMoney(),
		Name:   util.RandomUser(),
		Price:  util.RandomDecimalString(1, 1000),
		UserID: null.IntFrom(user.ID),
	}
	return
}

func createRandomShippingMethodForList() (ShippingMethod *db.ShippingMethod) {
	ShippingMethod = &db.ShippingMethod{
		ID:    util.RandomMoney(),
		Name:  util.RandomUser(),
		Price: util.RandomDecimalString(1, 1000),
		// UserID: null.IntFrom(user.ID),
	}
	return
}

func requireBodyMatchShippingMethod(t *testing.T, body io.ReadCloser, shippingMethod *db.ShippingMethod) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotShippingMethod *db.ShippingMethod
	err = json.Unmarshal(data, &gotShippingMethod)

	require.NoError(t, err)
	require.Equal(t, shippingMethod.ID, gotShippingMethod.ID)
	require.Equal(t, shippingMethod.Name, gotShippingMethod.Name)
	require.Equal(t, shippingMethod.Price, gotShippingMethod.Price)
}

func requireBodyMatchShippingMethodForGet(t *testing.T, body io.ReadCloser, shippingMethod *db.GetShippingMethodByUserIDRow) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotShippingMethod *db.GetShippingMethodByUserIDRow
	err = json.Unmarshal(data, &gotShippingMethod)

	require.NoError(t, err)
	require.Equal(t, shippingMethod.ID, gotShippingMethod.ID)
	require.Equal(t, shippingMethod.Name, gotShippingMethod.Name)
	require.Equal(t, shippingMethod.Price, gotShippingMethod.Price)
	require.Equal(t, shippingMethod.UserID.Int64, gotShippingMethod.UserID.Int64)
}

func requireBodyMatchShippingMethods(t *testing.T, body io.ReadCloser, shippingMethods []*db.ShippingMethod) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotShippingMethods []*db.ShippingMethod
	err = json.Unmarshal(data, &gotShippingMethods)
	require.NoError(t, err)
	for i := range gotShippingMethods {
		require.Equal(t, shippingMethods[i].ID, gotShippingMethods[i].ID)
		require.Equal(t, shippingMethods[i].Name, gotShippingMethods[i].Name)
		require.Equal(t, shippingMethods[i].Price, gotShippingMethods[i].Price)
		// require.Equal(t, shippingMethods[i].UserID.Int64, gotShippingMethods[i].UserID.Int64)
	}
}
