package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
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

func TestCreateUserAddessAPI(t *testing.T) {
	var defaultAddressId int64
	user, _ := randomUAUser(t)
	address := createRandomAddress(user)
	if user.DefaultAddressID.Valid {
		defaultAddressId = user.DefaultAddressID.Int64
	} else {
		defaultAddressId = address.ID
	}
	userAddress := createRandomUserAddress(user, address, defaultAddressId)
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
				"name":         address.Name,
				"telephone":    address.Telephone,
				"address_line": address.AddressLine,
				"region":       address.Region,
				"city":         address.City,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(user, nil)

				arg1 := db.CreateAddressParams{
					UserID:      user.ID,
					Name:        address.Name,
					Telephone:   address.Telephone,
					AddressLine: address.AddressLine,
					Region:      address.Region,
					City:        address.City,
				}

				store.EXPECT().
					CreateAddress(gomock.Any(), gomock.Eq(arg1)).
					Times(1).
					Return(address, nil)

				if !user.DefaultAddressID.Valid && user.DefaultAddressID.Int64 == 0 {
					store.EXPECT().
						UpdateUser(gomock.Any(), gomock.Any()).
						Times(1)
				}

			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchUserAddressForCreate(t, rsp.Body, userAddress)
			},
		},
		{
			name: "NoAuthorization",
			ID:   user.ID,
			body: fiber.Map{
				"name":         address.Name,
				"telephone":    address.Telephone,
				"address_line": address.AddressLine,
				"region":       address.Region,
				"city":         address.City,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					CreateAddress(gomock.Any(), gomock.Any()).
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
				"name":         address.Name,
				"telephone":    address.Telephone,
				"address_line": address.AddressLine,
				"region":       address.Region,
				"city":         address.City,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, pgx.ErrTxClosed)

				arg1 := db.CreateAddressParams{
					UserID:      user.ID,
					Name:        address.Name,
					Telephone:   address.Telephone,
					AddressLine: address.AddressLine,
					Region:      address.Region,
					City:        address.City,
				}

				store.EXPECT().
					CreateAddress(gomock.Any(), gomock.Eq(arg1)).
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
				"telephone":    address.Telephone,
				"address_line": address.AddressLine,
				"region":       address.Region,
				"city":         address.City,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					CreateAddress(gomock.Any(), gomock.Any()).
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
	address := createRandomAddress(user)
	// userAddress := createRandomUserAddress(user, address)
	userAddressWithAddress := createRandomUserAddressWithAddressForGet(address)

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
			AddressID: address.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.GetUserAddressParams{
					UserID: user.ID,
					ID:     address.ID,
				}
				store.EXPECT().
					GetUserAddress(gomock.Any(), gomock.Eq(arg)).
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
			AddressID: address.ID,
			ID:        user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserAddress(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:      "NotFound",
			ID:        user.ID,
			AddressID: address.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.GetUserAddressParams{
					UserID: user.ID,
					ID:     address.ID,
				}
				store.EXPECT().
					GetUserAddress(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(nil, pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
			},
		},
		{
			name:      "InternalError",
			ID:        user.ID,
			AddressID: address.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.GetUserAddressParams{
					UserID: user.ID,
					ID:     address.ID,
				}
				store.EXPECT().
					GetUserAddress(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(nil, pgx.ErrTxClosed)
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
					GetUserAddress(gomock.Any(), gomock.Any()).
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
	addresses := make([]*db.Address, n)
	userAddresses := make([]*db.ListAddressesByUserIDRow, n)
	addressesID := make([]int64, n)
	user, _ := randomUAUser(t)
	address1 := createRandomAddress(user)
	address2 := createRandomAddress(user)
	address3 := createRandomAddress(user)
	userAddress1 := createRandomAddressForList(user, address1)
	userAddress2 := createRandomAddressForList(user, address2)
	userAddress3 := createRandomAddressForList(user, address3)
	userAddresses = append(userAddresses, userAddress1[0], userAddress2[0], userAddress3[0])

	addresses = append(addresses, address1, address2, address3)
	// userAddresses = append(userAddresses, userAddress1, userAddress2, userAddress3)
	addressesID = append(addressesID, address1.ID, address2.ID, address3.ID)

	// finalRsp := createFinalRspForListAddress(userAddresses, addresses)

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

				store.EXPECT().
					ListAddressesByUserID(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(userAddresses, nil)

			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchListUserAddresses(t, rsp.Body, addresses)
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
					ListAddressesByUserID(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, pgx.ErrTxClosed)
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
					ListAddressesByUserID(gomock.Any(), gomock.Any()).
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
					ListAddressesByUserID(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/usr/v1/users/%d/addresses", tc.ID)
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			// Add query parameters to request URL
			q := request.URL.Query()
			q.Add("page_id", strconv.Itoa(tc.query.pageID))
			q.Add("page_size", strconv.Itoa(tc.query.pageSize))
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
	address := createRandomAddress(user)
	// userAddress := createRandomUserAddress(user, address)

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
			AddressID: address.ID,
			ID:        user.ID,
			body: fiber.Map{
				"default_address_id": user.DefaultAddressID.Int64,
				"address_line":       "new address",
				"region":             "new region",
				"city":               "new city",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.UpdateUserParams{
					ID:               user.ID,
					DefaultAddressID: user.DefaultAddressID,
				}

				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(user, nil)

				arg1 := db.UpdateAddressParams{
					ID:          address.ID,
					UserID:      user.ID,
					AddressLine: null.StringFrom("new address"),
					Region:      null.StringFrom("new region"),
					City:        null.StringFrom("new city"),
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
			AddressID: address.ID,
			ID:        user.ID,
			body: fiber.Map{

				"default_address_id": user.DefaultAddressID.Int64,
				"address_line":       "new address",
				"region":             "new region",
				"city":               "new city",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
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
			AddressID: address.ID,
			ID:        user.ID,
			body: fiber.Map{

				"default_address_id": user.DefaultAddressID.Int64,
				"address_line":       "new address",
				"region":             "new region",
				"city":               "new city",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateUserParams{
					ID:               user.ID,
					DefaultAddressID: user.DefaultAddressID,
				}

				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(nil, pgx.ErrTxClosed)

				arg1 := db.UpdateAddressParams{
					ID:          address.ID,
					UserID:      user.ID,
					AddressLine: null.StringFrom("new address"),
					Region:      null.StringFrom("new region"),
					City:        null.StringFrom("new city"),
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
					UpdateUser(gomock.Any(), gomock.Any()).
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
			ik := mockik.NewMockImageKitManagement(ctrl)
			mailSender := mockemail.NewMockEmailSender(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store, worker, ik, mailSender)
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
	address := createRandomAddress(user)
	// userAddress := createRandomUserAddress(user, address)

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
			AddressID: address.ID,
			ID:        user.ID,

			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteUserAddressParams{
					ID:     address.ID,
					UserID: user.ID,
				}

				store.EXPECT().
					DeleteUserAddress(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(address, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:      "NotFound",
			AddressID: address.ID,
			ID:        user.ID,

			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteUserAddressParams{
					UserID: user.ID,
					ID:     address.ID,
				}
				store.EXPECT().
					DeleteUserAddress(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(nil, pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
			},
		},
		{
			name:      "InternalError",
			AddressID: address.ID,
			ID:        user.ID,

			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteUserAddressParams{
					ID:     address.ID,
					UserID: user.ID,
				}
				store.EXPECT().
					DeleteUserAddress(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(nil, pgx.ErrTxClosed)
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
			ik := mockik.NewMockImageKitManagement(ctrl)
			mailSender := mockemail.NewMockEmailSender(ctrl)

			// build stubs
			tc.buildStub(store)

			// start test server and send request
			server := newTestServer(t, store, worker, ik, mailSender)
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

func randomUAUser(t *testing.T) (user *db.User, password string) {
	password = util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	user = &db.User{
		ID:               util.RandomMoney(),
		Username:         util.RandomUser(),
		Email:            util.RandomEmail(),
		Password:         hashedPassword,
		DefaultAddressID: null.IntFrom(util.RandomMoney()),
		// Telephone: int32(util.RandomInt(910000000, 929999999)),
	}
	return
}

func createRandomAddress(user *db.User) (address *db.Address) {
	address = &db.Address{
		ID:          util.RandomMoney(),
		UserID:      user.ID,
		Name:        util.RandomUser(),
		Telephone:   util.GenerateRandomValidPhoneNumber(),
		AddressLine: util.RandomUser(),
		Region:      util.RandomUser(),
		City:        util.RandomUser(),
	}
	return
}

func createRandomAddressForList(user *db.User, address *db.Address) (addressList []*db.ListAddressesByUserIDRow) {
	return []*db.ListAddressesByUserIDRow{
		{
			ID:               address.ID,
			UserID:           address.UserID,
			Name:             address.Name,
			Telephone:        address.Telephone,
			AddressLine:      address.AddressLine,
			Region:           address.Region,
			City:             address.City,
			DefaultAddressID: user.DefaultAddressID,
		},
	}
}

func createRandomUserAddress(user *db.User, address *db.Address, defaultAddresId int64) (userAddress userAddressResponse) {
	userAddress = userAddressResponse{
		ID:               address.ID,
		UserID:           user.ID,
		Name:             address.Name,
		Telephone:        address.Telephone,
		AddressLine:      address.AddressLine,
		Region:           address.Region,
		City:             address.City,
		DefaultAddressID: defaultAddresId,
	}
	return
}

// func createRandomUserAddressWithAddress(userAddress db.UserAddress, address db.Address) (userAddressWithAddress db.CreateUserAddressWithAddressRow) {
// 	userAddressWithAddress = db.CreateUserAddressWithAddressRow{
// 		UserID:      user.ID,
// 		AddressID:   address.ID,
// 		AddressLine: address.AddressLine,
// 		Region:      address.Region,
// 		City:        address.City,
// 	}
// 	return
// }

func createRandomUserAddressWithAddressForGet(address *db.Address) (userAddressWithAddress *db.GetUserAddressRow) {
	userAddressWithAddress = &db.GetUserAddressRow{
		UserID:      address.UserID,
		ID:          address.ID,
		AddressLine: address.AddressLine,
		Region:      address.Region,
		City:        address.City,
	}
	return
}

func requireBodyMatchUserAddressForCreate(t *testing.T, body io.ReadCloser, address userAddressResponse) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotAddress userAddressResponse
	err = json.Unmarshal(data, &gotAddress)
	require.NoError(t, err)

	// var gotUserAddress db.UserAddress
	// err = json.Unmarshal(data, &gotUserAddress)

	require.NoError(t, err)
	require.Equal(t, address.ID, gotAddress.ID)
	require.Equal(t, address.UserID, gotAddress.UserID)
	require.Equal(t, address.Name, gotAddress.Name)
	require.Equal(t, address.Telephone, gotAddress.Telephone)
	require.Equal(t, address.AddressLine, gotAddress.AddressLine)
	require.Equal(t, address.Region, gotAddress.Region)
	require.Equal(t, address.City, gotAddress.City)
}

func requireBodyMatchUserAddressForGet(t *testing.T, body io.ReadCloser, userAddress *db.GetUserAddressRow) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotUserAddress *db.GetUserAddressRow
	err = json.Unmarshal(data, &gotUserAddress)

	require.NoError(t, err)
	require.Equal(t, userAddress.UserID, gotUserAddress.UserID)
	require.Equal(t, userAddress.ID, gotUserAddress.ID)
	require.Equal(t, userAddress.AddressLine, gotUserAddress.AddressLine)
	require.Equal(t, userAddress.Region, gotUserAddress.Region)
	require.Equal(t, userAddress.City, gotUserAddress.City)
}

// func requireBodyMatchUserAddresses(t *testing.T, body io.ReadCloser, userAddresses []db.UserAddress) {
// 	data, err := io.ReadAll(body)
// 	require.NoError(t, err)

// 	var gotUserAddresses []db.UserAddress
// 	err = json.Unmarshal(data, &gotUserAddresses)
// 	require.NoError(t, err)
// 	require.Equal(t, userAddresses, gotUserAddresses)
// }

func requireBodyMatchListUserAddresses(t *testing.T, body io.ReadCloser, userAddresses []*db.Address) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotUserAddresses []*db.Address
	err = json.Unmarshal(data, &gotUserAddresses)
	require.NoError(t, err)
	require.Equal(t, userAddresses, gotUserAddresses)
}

// func createFinalRspForListAddress(addresses []db.Address) (rsp []db.Address) {
// 	for i := 0; i < len(addresses); i++ {
// 		rsp = append(rsp, listUserAddressResponse{
// 			UserID:         userAddresses[i].UserID,
// 			AddressID:      userAddresses[i].AddressID,
// 			DefaultAddress: null.IntFromPtr(&userAddresses[i].DefaultAddress.Int64),
// 			AddressLine:    addresses[i].AddressLine,
// 			Region:         addresses[i].Region,
// 			City:           addresses[i].City,
// 		})
// 	}
// 	return
// }
