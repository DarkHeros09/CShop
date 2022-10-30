package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mockdb "github.com/cshop/v3/db/mock"
	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/token"
	"github.com/cshop/v3/util"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/guregu/null"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/require"
)

func TestCreateUserAddessAPI(t *testing.T) {
	user, _ := randomUAUser(t)
	address := createRandomAddress(t)
	userAddress := createRandomUserAddress(t, user, address)
	userAddressWithAddress := createRandomUserAddressWithAddress(t, userAddress, address)

	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"user_id":      user.ID,
				"address_line": address.AddressLine,
				"region":       address.Region,
				"city":         address.City,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.CreateUserAddressWithAddressParams{
					AddressLine: address.AddressLine,
					Region:      address.Region,
					City:        address.City,
					UserID:      user.ID,
				}

				store.EXPECT().
					CreateUserAddressWithAddress(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(userAddressWithAddress, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchUserAddressForCreate(t, recorder.Body, userAddressWithAddress)
			},
		},
		{
			name: "NoAuthorization",
			body: gin.H{
				"user_id":      user.ID,
				"address_line": address.AddressLine,
				"region":       address.Region,
				"city":         address.City,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUserAddressWithAddress(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"user_id":      user.ID,
				"address_line": address.AddressLine,
				"region":       address.Region,
				"city":         address.City,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateUserAddressWithAddressParams{
					AddressLine: address.AddressLine,
					Region:      address.Region,
					City:        address.City,
					UserID:      user.ID,
				}

				store.EXPECT().
					CreateUserAddressWithAddress(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.CreateUserAddressWithAddressRow{}, pgx.ErrTxClosed)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidUserID",
			body: gin.H{
				"user_id":      0,
				"address_line": address.AddressLine,
				"region":       address.Region,
				"city":         address.City,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUserAddressWithAddress(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/users/addresses"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
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
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			ID:   userAddress.AddressID,
			body: gin.H{
				"id": userAddress.AddressID,
			},
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchUserAddressForGet(t, recorder.Body, userAddressWithAddress)
			},
		},
		{
			name: "NoAuthorization",
			ID:   userAddress.AddressID,
			body: gin.H{
				"id": userAddress.AddressID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserAddressWithAddress(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "NotFound",
			ID:   userAddress.AddressID,
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "InternalError",
			ID:   userAddress.AddressID,
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidID",
			ID:   0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserAddressWithAddress(gomock.Any(), gomock.Any()).
					Times(0)

			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t) // no need to call defer ctrl.finish() in 1.6V

			store := mockdb.NewMockStore(ctrl)
			// build stubs
			tc.buildStub(store)

			// start test server and send request
			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/users/addresses/%d", tc.ID)
			request, err := http.NewRequest(http.MethodGet, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			//check response
			tc.checkResponse(t, recorder)
		})

	}

}

func TestListUsersAddressAPI(t *testing.T) {
	n := 5
	userAddresses := make([]db.UserAddress, n)
	user, _ := randomUAUser(t)
	address1 := createRandomAddress(t)
	address2 := createRandomAddress(t)
	address3 := createRandomAddress(t)
	userAddress1 := createRandomUserAddress(t, user, address1)
	userAddress2 := createRandomUserAddress(t, user, address2)
	userAddress3 := createRandomUserAddress(t, user, address3)

	userAddresses = append(userAddresses, userAddress1, userAddress2, userAddress3)

	type Query struct {
		pageID   int
		pageSize int
	}

	testCases := []struct {
		name          string
		query         Query
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
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
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchUserAddresses(t, recorder.Body, userAddresses)
			},
		},
		{
			name: "InternalError",
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
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidPageID",
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
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidPageSize",
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
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := "/users/addresses"
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			// Add query parameters to request URL
			q := request.URL.Query()
			q.Add("page_id", fmt.Sprintf("%d", tc.query.pageID))
			q.Add("page_size", fmt.Sprintf("%d", tc.query.pageSize))
			request.URL.RawQuery = q.Encode()

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestUpdateUserAddressAPI(t *testing.T) {
	user, _ := randomUAUser(t)
	address := createRandomAddress(t)
	userAddress := createRandomUserAddress(t, user, address)

	testCases := []struct {
		name          string
		UserID        int64
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:   "OK",
			UserID: user.ID,
			body: gin.H{
				"address_id":      userAddress.AddressID,
				"default_address": userAddress.DefaultAddress,
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:   "NoAuthorization",
			UserID: user.ID,
			body: gin.H{
				"address_id":      userAddress.AddressID,
				"default_address": userAddress.DefaultAddress,
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:   "InternalError",
			UserID: user.ID,
			body: gin.H{
				"address_id":      userAddress.AddressID,
				"default_address": userAddress.DefaultAddress,
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:   "InvalidUserID",
			UserID: 0,
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/users/addresses/%d", tc.UserID)
			request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestDeleteUserAddressAPI(t *testing.T) {
	user, _ := randomUAUser(t)
	address := createRandomAddress(t)
	userAddress := createRandomUserAddress(t, user, address)

	testCases := []struct {
		name          string
		UserID        int64
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:   "OK",
			UserID: userAddress.UserID,
			body: gin.H{
				"address_id": userAddress.AddressID,
			},
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
					Return(nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:   "NotFound",
			UserID: userAddress.UserID,
			body: gin.H{
				"address_id": userAddress.AddressID,
			},
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
					Return(pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:   "InternalError",
			UserID: userAddress.UserID,
			body: gin.H{
				"address_id": userAddress.AddressID,
			},
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
					Return(pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:   "InvalidID",
			UserID: 0,
			body: gin.H{
				"address_id": userAddress.AddressID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteUserAddress(gomock.Any(), gomock.Any()).
					Times(0)

			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t) // no need to call defer ctrl.finish() in 1.6V

			store := mockdb.NewMockStore(ctrl)

			// build stubs
			tc.buildStub(store)

			// start test server and send request
			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/users/addresses/%d", tc.UserID)
			request, err := http.NewRequest(http.MethodDelete, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			//check response
			tc.checkResponse(t, recorder)
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

func requireBodyMatchUserAddressForCreate(t *testing.T, body *bytes.Buffer, userAddress db.CreateUserAddressWithAddressRow) {
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

func requireBodyMatchUserAddressForGet(t *testing.T, body *bytes.Buffer, userAddress db.GetUserAddressWithAddressRow) {
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

func requireBodyMatchUserAddresses(t *testing.T, body *bytes.Buffer, userAddresses []db.UserAddress) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotUserAddresses []db.UserAddress
	err = json.Unmarshal(data, &gotUserAddresses)
	require.NoError(t, err)
	require.Equal(t, userAddresses, gotUserAddresses)
}
