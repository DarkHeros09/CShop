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

func TestGetVariationOptionAPI(t *testing.T) {
	variationOption := randomVariationOption()

	testCases := []struct {
		name              string
		VariationOptionID int64
		buildStub         func(store *mockdb.MockStore)
		checkResponse     func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:              "OK",
			VariationOptionID: variationOption.ID,
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetVariationOption(gomock.Any(), gomock.Eq(variationOption.ID)).
					Times(1).
					Return(variationOption, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchVariationOption(t, recorder.Body, variationOption)
			},
		},

		{
			name:              "NotFound",
			VariationOptionID: variationOption.ID,
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetVariationOption(gomock.Any(), gomock.Eq(variationOption.ID)).
					Times(1).
					Return(db.VariationOption{}, pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:              "InternalError",
			VariationOptionID: variationOption.ID,
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetVariationOption(gomock.Any(), gomock.Eq(variationOption.ID)).
					Times(1).
					Return(db.VariationOption{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:              "InvalidID",
			VariationOptionID: 0,
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetVariationOption(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/variation-options/%d", tc.VariationOptionID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			//check response
			tc.checkResponse(t, recorder)
		})

	}

}

func TestCreateVariationOptionAPI(t *testing.T) {
	admin, _ := randomVariationOptionSuperAdmin(t)
	variationOption := randomVariationOption()

	testCases := []struct {
		name              string
		body              gin.H
		VariationOptionID int64
		setupAuth         func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs        func(store *mockdb.MockStore)
		checkResponse     func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"value":        variationOption.Value,
				"variation_id": variationOption.VariationID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateVariationOptionParams{
					VariationID: variationOption.VariationID,
					Value:       variationOption.Value,
				}

				store.EXPECT().
					CreateVariationOption(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(variationOption, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchVariationOption(t, recorder.Body, variationOption)
			},
		},
		{
			name:              "NoAuthorization",
			VariationOptionID: variationOption.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateVariationOptionParams{
					VariationID: variationOption.VariationID,
					Value:       variationOption.Value,
				}

				store.EXPECT().
					CreateVariationOption(gomock.Any(), gomock.Eq(arg)).
					Times(0)

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:              "Unauthorized",
			VariationOptionID: variationOption.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, false, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateVariationOptionParams{
					VariationID: variationOption.VariationID,
					Value:       variationOption.Value,
				}

				store.EXPECT().
					CreateVariationOption(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"value":        variationOption.Value,
				"variation_id": variationOption.VariationID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateVariationOption(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.VariationOption{}, pgx.ErrTxClosed)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidID",
			body: gin.H{
				"id": "invalid",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateVariationOption(gomock.Any(), gomock.Any()).
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

			url := "/variation-options"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestListVariationOptionsAPI(t *testing.T) {
	n := 5
	VariationOptions := make([]db.VariationOption, n)
	for i := 0; i < n; i++ {
		VariationOptions[i] = randomVariationOption()
	}

	type Query struct {
		pageID   int
		pageSize int
	}

	testCases := []struct {
		name          string
		query         Query
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			query: Query{
				pageID:   1,
				pageSize: n,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListVariationOptionsParams{
					Limit:  int32(n),
					Offset: 0,
				}

				store.EXPECT().
					ListVariationOptions(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(VariationOptions, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchVariationOptions(t, recorder.Body, VariationOptions)
			},
		},
		{
			name: "InternalError",
			query: Query{
				pageID:   1,
				pageSize: n,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListVariationOptions(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.VariationOption{}, pgx.ErrTxClosed)
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
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListVariationOptions(gomock.Any(), gomock.Any()).
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
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListVariationOptions(gomock.Any(), gomock.Any()).
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

			url := "/variation-options"
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			// Add query parameters to request URL
			q := request.URL.Query()
			q.Add("page_id", fmt.Sprintf("%d", tc.query.pageID))
			q.Add("page_size", fmt.Sprintf("%d", tc.query.pageSize))
			request.URL.RawQuery = q.Encode()

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestUpdateVariationOptionAPI(t *testing.T) {
	admin, _ := randomVariationOptionSuperAdmin(t)
	variationOption := randomVariationOption()

	testCases := []struct {
		name              string
		body              gin.H
		VariationOptionID int64
		setupAuth         func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs        func(store *mockdb.MockStore)
		checkResponse     func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:              "OK",
			VariationOptionID: variationOption.ID,
			body: gin.H{
				"value":        "new name",
				"variation_id": variationOption.VariationID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateVariationOptionParams{
					Value:       null.StringFrom("new name"),
					ID:          variationOption.ID,
					VariationID: null.IntFrom(variationOption.VariationID),
				}

				store.EXPECT().
					UpdateVariationOption(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(variationOption, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:              "Unauthorized",
			VariationOptionID: variationOption.ID,
			body: gin.H{
				"value":        "new name",
				"variation_id": variationOption.VariationID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, false, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateVariationOptionParams{
					Value:       null.StringFrom("new name"),
					ID:          variationOption.ID,
					VariationID: null.IntFrom(variationOption.VariationID),
				}

				store.EXPECT().
					UpdateVariationOption(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:              "NoAuthorization",
			VariationOptionID: variationOption.ID,
			body: gin.H{
				"value":        "new name",
				"variation_id": variationOption.VariationID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateVariationOptionParams{
					Value:       null.StringFrom("new name"),
					ID:          variationOption.ID,
					VariationID: null.IntFrom(variationOption.VariationID),
				}

				store.EXPECT().
					UpdateVariationOption(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:              "InternalError",
			VariationOptionID: variationOption.ID,
			body: gin.H{
				"value":        "new name",
				"variation_id": variationOption.VariationID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateVariationOptionParams{
					Value:       null.StringFrom("new name"),
					ID:          variationOption.ID,
					VariationID: null.IntFrom(variationOption.VariationID),
				}
				store.EXPECT().
					UpdateVariationOption(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.VariationOption{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:              "InvalidID",
			VariationOptionID: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateVariationOption(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/variation-options/%d", tc.VariationOptionID)
			request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestDeleteVariationOptionAPI(t *testing.T) {
	admin, _ := randomVariationOptionSuperAdmin(t)
	variationOption := randomVariationOption()

	testCases := []struct {
		name              string
		body              gin.H
		VariationOptionID int64
		setupAuth         func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub         func(store *mockdb.MockStore)
		checkResponse     func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:              "OK",
			VariationOptionID: variationOption.ID,
			body: gin.H{
				"variation_id": variationOption.VariationID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {

				store.EXPECT().
					DeleteVariationOption(gomock.Any(), gomock.Eq(variationOption.ID)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:              "Unauthorized",
			VariationOptionID: variationOption.ID,
			body: gin.H{
				"variation_id": variationOption.VariationID,
				"id":           variationOption.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, 2, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {

				store.EXPECT().
					DeleteVariationOption(gomock.Any(), gomock.Eq(variationOption.ID)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:              "NoAuthorization",
			VariationOptionID: variationOption.ID,
			body: gin.H{
				"variation_id": variationOption.VariationID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStub: func(store *mockdb.MockStore) {

				store.EXPECT().
					DeleteVariationOption(gomock.Any(), gomock.Eq(variationOption.ID)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:              "NotFound",
			VariationOptionID: variationOption.ID,
			body: gin.H{
				"variation_id": variationOption.VariationID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {

				store.EXPECT().
					DeleteVariationOption(gomock.Any(), gomock.Eq(variationOption.ID)).
					Times(1).
					Return(pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:              "InternalError",
			VariationOptionID: variationOption.ID,
			body: gin.H{
				"variation_id": variationOption.VariationID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {

				store.EXPECT().
					DeleteVariationOption(gomock.Any(), gomock.Eq(variationOption.ID)).
					Times(1).
					Return(pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:              "InvalidID",
			VariationOptionID: 0,
			body: gin.H{
				"variation_id": 0,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteVariationOption(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/variation-options/%d", tc.VariationOptionID)
			request, err := http.NewRequest(http.MethodDelete, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			//check response
			tc.checkResponse(t, recorder)
		})

	}

}

func randomVariationOptionSuperAdmin(t *testing.T) (admin db.Admin, password string) {
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

func randomVariationOption() db.VariationOption {
	return db.VariationOption{
		ID:          util.RandomInt(1, 1000),
		VariationID: util.RandomMoney(),
		Value:       util.RandomUser(),
	}
}

func requireBodyMatchVariationOption(t *testing.T, body *bytes.Buffer, VariationOption db.VariationOption) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotVariationOption db.VariationOption
	err = json.Unmarshal(data, &gotVariationOption)
	require.NoError(t, err)
	require.Equal(t, VariationOption, gotVariationOption)
}

func requireBodyMatchVariationOptions(t *testing.T, body *bytes.Buffer, variationOptions []db.VariationOption) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotVariationOptions []db.VariationOption
	err = json.Unmarshal(data, &gotVariationOptions)
	require.NoError(t, err)
	require.Equal(t, variationOptions, gotVariationOptions)
}
