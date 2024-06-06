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
	"github.com/cshop/v3/token"
	"github.com/cshop/v3/util"
	mockwk "github.com/cshop/v3/worker/mock"
	"github.com/gofiber/fiber/v2"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCreateUserAddessAPI(t *testing.T) {
	user, _ := randomUAUser(t)
	address := createRandomAddress(t)
	userAddress := createRandomUserAddress(t, user, address)
	// userAddressWithAddress := createRandomUserAddressWithAddress(t, userAddress, address)

	testCases := []struct {
		name          string
		body          fiber.Map
		ID            int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(rsp *http.Response)
	}{
		{
			name: "OK",
			ID:   user.ID,
			body: fiber.Map{
				"address_line": address.AddressLine,
				"region":       address.Region,
				"city":         address.City,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg1 := db.CreateAddressParams{
					AddressLine: address.AddressLine,
					Region:      address.Region,
					City:        address.City,
				}

				store.EXPECT().
					CreateAddress(gomock.Any(), gomock.Eq(arg1)).
					Times(1).
					Return(address, nil)

				store.EXPECT().
					CheckUserAddressDefaultAddress(gomock.Any(), gomock.Eq(user.ID)).
					Times(1)

				arg2 := db.CreateUserAddressParams{
					UserID:         user.ID,
					AddressID:      address.ID,
					DefaultAddress: null.IntFrom(address.ID),
				}

				store.EXPECT().
					CreateUserAddress(gomock.Any(), gomock.Eq(arg2)).
					Times(1).
					Return(userAddress, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchUserAddressForCreate(t, rsp.Body, address, userAddress)
			},
		},
		{
			name: "NoAuthorization",
			ID:   user.ID,
			body: fiber.Map{
				"address_line": address.AddressLine,
				"region":       address.Region,
				"city":         address.City,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAddress(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					CreateUserAddress(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name: "InternalError",
			ID:   user.ID,
			body: fiber.Map{
				"address_line": address.AddressLine,
				"region":       address.Region,
				"city":         address.City,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg1 := db.CreateAddressParams{
					AddressLine: address.AddressLine,
					Region:      address.Region,
					City:        address.City,
				}

				store.EXPECT().
					CreateAddress(gomock.Any(), gomock.Eq(arg1)).
					Times(1).
					Return(db.Address{}, pgx.ErrTxClosed)

				store.EXPECT().
					CreateUserAddress(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name: "InvalidUserID",
			ID:   0,
			body: fiber.Map{
				"address_line": address.AddressLine,
				"region":       address.Region,
				"city":         address.City,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAddress(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					CreateUserAddress(gomock.Any(), gomock.Any()).
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
			tc.buildStubs(store)

			server := newTestServer(t, store, worker)
			// //recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/usr/v1/users/%d/addresses", tc.ID)
			request, err := http.NewRequest(fiber.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.userTokenMaker)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
		})
	}
}

func TestGetUserAddressAPI(t *testing.T) {
	user, _ := randomUAUser(t)
	address := createRandomAddress(t)
	userAddress := createRandomUserAddress(t, user, address)
	userAddressWithAddress := createRandomUserAddressWithAddressForGet(t, userAddress, address)

	testCases := []struct {
		name          string
		ID            int64
		AddressID     int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rsp *http.Response)
	}{
		{
			name:      "OK",
			ID:        user.ID,
			AddressID: userAddress.AddressID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.GetUserAddressWithAddressParams{
					UserID:    userAddress.UserID,
					AddressID: userAddress.AddressID,
				}
				store.EXPECT().
					GetUserAddressWithAddress(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(userAddressWithAddress, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchUserAddressForGet(t, rsp.Body, userAddressWithAddress)
			},
		},
		{
			name:      "NoAuthorization",
			AddressID: userAddress.AddressID,
			ID:        user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserAddressWithAddress(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:      "NotFound",
			ID:        user.ID,
			AddressID: userAddress.AddressID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.GetUserAddressWithAddressParams{
					UserID:    userAddress.UserID,
					AddressID: userAddress.AddressID,
				}
				store.EXPECT().
					GetUserAddressWithAddress(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.GetUserAddressWithAddressRow{}, pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
			},
		},
		{
			name:      "InternalError",
			ID:        user.ID,
			AddressID: userAddress.AddressID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.GetUserAddressWithAddressParams{
					UserID:    userAddress.UserID,
					AddressID: userAddress.AddressID,
				}
				store.EXPECT().
					GetUserAddressWithAddress(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.GetUserAddressWithAddressRow{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:      "InvalidID",
			ID:        0,
			AddressID: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserAddressWithAddress(gomock.Any(), gomock.Any()).
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
			// build stubs
			tc.buildStub(store)

			// start test server and send request
			server := newTestServer(t, store, worker)
			// //recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/usr/v1/users/%d/addresses/%d", tc.ID, tc.AddressID)
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.userTokenMaker)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request, -1)
			require.NoError(t, err)
			//check response
			tc.checkResponse(t, rsp)
		})

	}

}

func TestListUsersAddressAPI(t *testing.T) {
	n := 5
	addresses := make([]db.Address, n)
	userAddresses := make([]db.UserAddress, n)
	addressesID := make([]int64, n)
	user, _ := randomUAUser(t)
	address1 := createRandomAddress(t)
	address2 := createRandomAddress(t)
	address3 := createRandomAddress(t)
	userAddress1 := createRandomUserAddress(t, user, address1)
	userAddress2 := createRandomUserAddress(t, user, address2)
	userAddress3 := createRandomUserAddress(t, user, address3)

	addresses = append(addresses, address1, address2, address3)
	userAddresses = append(userAddresses, userAddress1, userAddress2, userAddress3)
	addressesID = append(addressesID, address1.ID, address2.ID, address3.ID)

	finalRsp := createFinalRspForListAddress(t, userAddresses, addresses)

	type Query struct {
		pageID   int
		pageSize int
	}

	testCases := []struct {
		name          string
		query         Query
		ID            int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rsp *http.Response)
	}{
		{
			name: "OK",
			ID:   user.ID,
			query: Query{
				pageID:   1,
				pageSize: n,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListUserAddressesParams{
					UserID: user.ID,
					Limit:  int32(n),
					Offset: 0,
				}

				store.EXPECT().
					ListUserAddresses(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(userAddresses, nil)

				store.EXPECT().
					ListAddressesByID(gomock.Any(), gomock.Eq(addressesID)).
					Times(1).
					Return(addresses, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchListUserAddresses(t, rsp.Body, finalRsp)
			},
		},
		{
			name: "InternalError",
			ID:   user.ID,
			query: Query{
				pageID:   1,
				pageSize: n,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListUserAddresses(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.UserAddress{}, pgx.ErrTxClosed)

				store.EXPECT().
					ListAddressesByID(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name: "InvalidPageID",
			ID:   user.ID,
			query: Query{
				pageID:   -1,
				pageSize: n,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListUserAddresses(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					ListAddressesByID(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusBadRequest, rsp.StatusCode)
			},
		},
		{
			name: "InvalidPageSize",
			ID:   user.ID,
			query: Query{
				pageID:   1,
				pageSize: 100000,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListUserAddresses(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					ListAddressesByID(gomock.Any(), gomock.Any()).
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
			tc.buildStubs(store)

			server := newTestServer(t, store, worker)
			// //recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/usr/v1/users/%d/addresses", tc.ID)
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			// Add query parameters to request URL
			q := request.URL.Query()
			q.Add("page_id", fmt.Sprintf("%d", tc.query.pageID))
			q.Add("page_size", fmt.Sprintf("%d", tc.query.pageSize))
			request.URL.RawQuery = q.Encode()

			tc.setupAuth(t, request, server.userTokenMaker)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request, -1)
			require.NoError(t, err)
			tc.checkResponse(t, rsp)
		})
	}
}

func TestUpdateUserAddressAPI(t *testing.T) {
	user, _ := randomUAUser(t)
	address := createRandomAddress(t)
	userAddress := createRandomUserAddress(t, user, address)

	testCases := []struct {
		name          string
		AddressID     int64
		body          fiber.Map
		ID            int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rsp *http.Response)
	}{
		{
			name:      "OK",
			AddressID: userAddress.AddressID,
			ID:        user.ID,
			body: fiber.Map{
				"default_address": userAddress.DefaultAddress.Int64,
				"address_line":    "new address",
				"region":          "new region",
				"city":            "new city",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.UpdateUserAddressParams{
					DefaultAddress: null.IntFromPtr(userAddress.DefaultAddress.Ptr()),
					UserID:         userAddress.UserID,
					AddressID:      userAddress.AddressID,
				}

				store.EXPECT().
					UpdateUserAddress(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(userAddress, nil)

				arg1 := db.UpdateAddressParams{
					AddressLine: null.StringFrom("new address"),
					Region:      null.StringFrom("new region"),
					City:        null.StringFrom("new city"),
					ID:          userAddress.AddressID,
				}

				store.EXPECT().
					UpdateAddress(gomock.Any(), gomock.Eq(arg1)).
					Times(1).
					Return(address, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:      "NoAuthorization",
			AddressID: userAddress.AddressID,
			ID:        user.ID,
			body: fiber.Map{

				"default_address": userAddress.DefaultAddress.Int64,
				"address_line":    "new address",
				"region":          "new region",
				"city":            "new city",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateUserAddress(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					UpdateAddress(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:      "InternalError",
			AddressID: userAddress.AddressID,
			ID:        user.ID,
			body: fiber.Map{

				"default_address": userAddress.DefaultAddress.Int64,
				"address_line":    "new address",
				"region":          "new region",
				"city":            "new city",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateUserAddressParams{
					DefaultAddress: null.IntFromPtr(userAddress.DefaultAddress.Ptr()),
					UserID:         userAddress.UserID,
					AddressID:      userAddress.AddressID,
				}

				store.EXPECT().
					UpdateUserAddress(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.UserAddress{}, pgx.ErrTxClosed)

				arg1 := db.UpdateAddressParams{
					AddressLine: null.StringFrom("new address"),
					Region:      null.StringFrom("new region"),
					City:        null.StringFrom("new city"),
					ID:          userAddress.AddressID,
				}

				store.EXPECT().
					UpdateAddress(gomock.Any(), gomock.Eq(arg1)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:      "InvalidUserID",
			AddressID: 0,
			ID:        user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateUserAddress(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					UpdateAddress(gomock.Any(), gomock.Any()).
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
			tc.buildStubs(store)

			server := newTestServer(t, store, worker)
			// //recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/usr/v1/users/%d/addresses/%d", tc.ID, tc.AddressID)
			request, err := http.NewRequest(fiber.MethodPut, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.userTokenMaker)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(t, rsp)
		})
	}
}

func TestDeleteUserAddressAPI(t *testing.T) {
	user, _ := randomUAUser(t)
	address := createRandomAddress(t)
	userAddress := createRandomUserAddress(t, user, address)

	testCases := []struct {
		name          string
		AddressID     int64
		ID            int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rsp *http.Response)
	}{
		{
			name:      "OK",
			AddressID: userAddress.AddressID,
			ID:        user.ID,

			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteUserAddressParams{
					UserID:    userAddress.UserID,
					AddressID: userAddress.AddressID,
				}

				store.EXPECT().
					DeleteUserAddress(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(userAddress, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:      "NotFound",
			AddressID: userAddress.AddressID,
			ID:        user.ID,

			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteUserAddressParams{
					UserID:    userAddress.UserID,
					AddressID: userAddress.AddressID,
				}
				store.EXPECT().
					DeleteUserAddress(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.UserAddress{}, pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
			},
		},
		{
			name:      "InternalError",
			AddressID: userAddress.AddressID,
			ID:        user.ID,

			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteUserAddressParams{
					UserID:    userAddress.UserID,
					AddressID: userAddress.AddressID,
				}
				store.EXPECT().
					DeleteUserAddress(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.UserAddress{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:      "InvalidID",
			AddressID: 0,

			ID: user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteUserAddress(gomock.Any(), gomock.Any()).
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

			// build stubs
			tc.buildStub(store)

			// start test server and send request
			server := newTestServer(t, store, worker)
			// //recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/usr/v1/users/%d/addresses/%d", tc.ID, tc.AddressID)
			request, err := http.NewRequest(fiber.MethodDelete, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.userTokenMaker)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			//check response
			tc.checkResponse(t, rsp)
		})

	}

}

func randomUAUser(t *testing.T) (user db.User, password string) {
	password = util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	user = db.User{
		ID:        util.RandomMoney(),
		Username:  util.RandomUser(),
		Email:     util.RandomEmail(),
		Password:  hashedPassword,
		Telephone: int32(util.RandomInt(910000000, 929999999)),
	}
	return
}

func createRandomAddress(t *testing.T) (address db.Address) {
	address = db.Address{
		ID:          util.RandomMoney(),
		AddressLine: util.RandomUser(),
		Region:      util.RandomUser(),
		City:        util.RandomUser(),
	}
	return
}

func createRandomUserAddress(t *testing.T, user db.User, address db.Address) (userAddress db.UserAddress) {
	userAddress = db.UserAddress{
		DefaultAddress: null.IntFromPtr(&address.ID),
		UserID:         user.ID,
		AddressID:      address.ID,
	}
	return
}

func createRandomUserAddressWithAddress(t *testing.T, userAddress db.UserAddress, address db.Address) (userAddressWithAddress db.CreateUserAddressWithAddressRow) {
	userAddressWithAddress = db.CreateUserAddressWithAddressRow{
		UserID:      userAddress.UserID,
		AddressID:   userAddress.AddressID,
		AddressLine: address.AddressLine,
		Region:      address.Region,
		City:        address.City,
	}
	return
}

func createRandomUserAddressWithAddressForGet(t *testing.T, userAddress db.UserAddress, address db.Address) (userAddressWithAddress db.GetUserAddressWithAddressRow) {
	userAddressWithAddress = db.GetUserAddressWithAddressRow{
		UserID:      userAddress.UserID,
		AddressID:   userAddress.AddressID,
		AddressLine: address.AddressLine,
		Region:      address.Region,
		City:        address.City,
	}
	return
}

func requireBodyMatchUserAddressForCreate(t *testing.T, body io.ReadCloser, address db.Address, userAddress db.UserAddress) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotAddress db.Address
	err = json.Unmarshal(data, &gotAddress)

	var gotUserAddress db.UserAddress
	err = json.Unmarshal(data, &gotUserAddress)

	require.NoError(t, err)
	require.Equal(t, userAddress.UserID, gotUserAddress.UserID)
	require.Equal(t, userAddress.AddressID, gotUserAddress.AddressID)
	require.Equal(t, address.AddressLine, gotAddress.AddressLine)
	require.Equal(t, address.Region, gotAddress.Region)
	require.Equal(t, address.City, gotAddress.City)
}

func requireBodyMatchUserAddressForGet(t *testing.T, body io.ReadCloser, userAddress db.GetUserAddressWithAddressRow) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotUserAddress db.CreateUserAddressWithAddressRow
	err = json.Unmarshal(data, &gotUserAddress)

	require.NoError(t, err)
	require.Equal(t, userAddress.UserID, gotUserAddress.UserID)
	require.Equal(t, userAddress.AddressID, gotUserAddress.AddressID)
	require.Equal(t, userAddress.AddressLine, gotUserAddress.AddressLine)
	require.Equal(t, userAddress.Region, gotUserAddress.Region)
	require.Equal(t, userAddress.City, gotUserAddress.City)
}

func requireBodyMatchUserAddresses(t *testing.T, body io.ReadCloser, userAddresses []db.UserAddress) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotUserAddresses []db.UserAddress
	err = json.Unmarshal(data, &gotUserAddresses)
	require.NoError(t, err)
	require.Equal(t, userAddresses, gotUserAddresses)
}

func requireBodyMatchListUserAddresses(t *testing.T, body io.ReadCloser, userAddresses []listUserAddressResponse) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotUserAddresses []listUserAddressResponse
	err = json.Unmarshal(data, &gotUserAddresses)
	require.NoError(t, err)
	require.Equal(t, userAddresses, gotUserAddresses)
}

func createFinalRspForListAddress(t *testing.T, userAddresses []db.UserAddress, addresses []db.Address) (rsp []listUserAddressResponse) {
	for i := 0; i < len(addresses); i++ {
		rsp = append(rsp, listUserAddressResponse{
			UserID:         userAddresses[i].UserID,
			AddressID:      userAddresses[i].AddressID,
			DefaultAddress: null.IntFromPtr(&userAddresses[i].DefaultAddress.Int64),
			AddressLine:    addresses[i].AddressLine,
			Region:         addresses[i].Region,
			City:           addresses[i].City,
		})
	}
	return
}
