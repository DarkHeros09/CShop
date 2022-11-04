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

func TestCreatePaymentMethodAPI(t *testing.T) {
	user, _ := randomPMUser(t)
	paymentType := createRandomPaymentType(t)
	paymentMethod := createRandomPaymentMethod(t, user, paymentType)

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
				"user_id":           user.ID,
				"payment_method_id": paymentMethod.ID,
				"provider":          paymentMethod.Provider,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.CreatePaymentMethodParams{
					UserID:        user.ID,
					PaymentTypeID: paymentMethod.ID,
					Provider:      paymentMethod.Provider,
				}

				store.EXPECT().
					CreatePaymentMethod(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(paymentMethod, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchPaymentMethod(t, recorder.Body, paymentMethod)
			},
		},
		{
			name: "NoAuthorization",
			body: gin.H{
				"user_id":           user.ID,
				"payment_method_id": paymentMethod.ID,
				"provider":          paymentMethod.Provider,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreatePaymentMethod(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"user_id":           user.ID,
				"payment_method_id": paymentMethod.ID,
				"provider":          paymentMethod.Provider,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreatePaymentMethodParams{
					UserID:        user.ID,
					PaymentTypeID: paymentMethod.ID,
					Provider:      paymentMethod.Provider,
				}

				store.EXPECT().
					CreatePaymentMethod(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.PaymentMethod{}, pgx.ErrTxClosed)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidUserID",
			body: gin.H{
				"user_id":           0,
				"payment_method_id": paymentMethod.ID,
				"provider":          paymentMethod.Provider,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreatePaymentMethod(gomock.Any(), gomock.Any()).
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

			url := "/users/payment-method"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestGetPaymentMethodAPI(t *testing.T) {
	user, _ := randomPMUser(t)
	paymentType := createRandomPaymentType(t)
	paymentMethod := createRandomPaymentMethod(t, user, paymentType)

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
			ID:   paymentMethod.ID,
			body: gin.H{
				"user_id": paymentMethod.UserID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.GetPaymentMethodParams{
					ID:     paymentMethod.ID,
					UserID: paymentMethod.UserID,
				}
				store.EXPECT().
					GetPaymentMethod(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(paymentMethod, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchPaymentMethod(t, recorder.Body, paymentMethod)
			},
		},
		{
			name: "NoAuthorization",
			ID:   paymentMethod.ID,
			body: gin.H{
				"user_id": paymentMethod.UserID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetPaymentMethod(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "NotFound",
			ID:   paymentMethod.ID,
			body: gin.H{
				"user_id": paymentMethod.UserID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.GetPaymentMethodParams{
					ID:     paymentMethod.ID,
					UserID: paymentMethod.UserID,
				}
				store.EXPECT().
					GetPaymentMethod(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.PaymentMethod{}, pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "InternalError",
			ID:   paymentMethod.ID,
			body: gin.H{
				"user_id": paymentMethod.UserID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.GetPaymentMethodParams{
					ID:     paymentMethod.ID,
					UserID: paymentMethod.UserID,
				}
				store.EXPECT().
					GetPaymentMethod(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.PaymentMethod{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidID",
			ID:   0,
			body: gin.H{
				"user_id": paymentMethod.UserID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetPaymentMethod(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/users/payment-method/%d", tc.ID)
			request, err := http.NewRequest(http.MethodGet, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			//check response
			tc.checkResponse(t, recorder)
		})

	}

}

func TestListPaymentMethodAPI(t *testing.T) {
	n := 5
	paymentMethods := make([]db.PaymentMethod, n)
	user, _ := randomPMUser(t)
	paymentType := createRandomPaymentType(t)
	paymentMethod1 := createRandomPaymentMethod(t, user, paymentType)
	paymentMethod2 := createRandomPaymentMethod(t, user, paymentType)
	paymentMethod3 := createRandomPaymentMethod(t, user, paymentType)

	paymentMethods = append(paymentMethods, paymentMethod1, paymentMethod2, paymentMethod3)

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
				arg := db.ListPaymentMethodsParams{
					UserID: user.ID,
					Limit:  int32(n),
					Offset: 0,
				}

				store.EXPECT().
					ListPaymentMethods(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(paymentMethods, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchPaymentMethodss(t, recorder.Body, paymentMethods)
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
					ListPaymentMethods(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.PaymentMethod{}, pgx.ErrTxClosed)
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
					ListPaymentMethods(gomock.Any(), gomock.Any()).
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
					ListPaymentMethods(gomock.Any(), gomock.Any()).
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

			url := "/users/payment-method"
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

func TestUpdatePaymentMethodAPI(t *testing.T) {
	user, _ := randomPMUser(t)
	paymentType := createRandomPaymentType(t)
	paymentMethod := createRandomPaymentMethod(t, user, paymentType)

	testCases := []struct {
		name          string
		ID            int64
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			ID:   paymentMethod.ID,
			body: gin.H{
				"user_id":         paymentMethod.UserID,
				"payment_type_id": paymentMethod.PaymentTypeID,
				"provider":        "new address",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.UpdatePaymentMethodParams{
					UserID:        null.IntFromPtr(&paymentMethod.UserID),
					PaymentTypeID: null.IntFromPtr(&paymentMethod.PaymentTypeID),
					Provider:      null.StringFrom("new address"),
					ID:            paymentMethod.ID,
				}

				store.EXPECT().
					UpdatePaymentMethod(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(paymentMethod, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "NoAuthorization",
			ID:   paymentMethod.ID,
			body: gin.H{
				"user_id":         paymentMethod.UserID,
				"payment_type_id": paymentMethod.PaymentTypeID,
				"provider":        "new address",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdatePaymentMethod(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InternalError",
			ID:   paymentMethod.ID,
			body: gin.H{
				"user_id":         paymentMethod.UserID,
				"payment_type_id": paymentMethod.PaymentTypeID,
				"provider":        "new address",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdatePaymentMethodParams{
					UserID:        null.IntFromPtr(&paymentMethod.UserID),
					PaymentTypeID: null.IntFromPtr(&paymentMethod.PaymentTypeID),
					Provider:      null.StringFrom("new address"),
					ID:            paymentMethod.ID,
				}

				store.EXPECT().
					UpdatePaymentMethod(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.PaymentMethod{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidUserID",
			ID:   paymentMethod.ID,
			body: gin.H{
				"user_id":         0,
				"payment_type_id": paymentMethod.PaymentTypeID,
				"provider":        "new address",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdatePaymentMethod(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/users/payment-method/%d", tc.ID)
			request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}

}

func TestDeletePaymentMethodAPI(t *testing.T) {
	user, _ := randomPMUser(t)
	paymentType := createRandomPaymentType(t)
	paymentMethod := createRandomPaymentMethod(t, user, paymentType)

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
			ID:   paymentMethod.ID,
			body: gin.H{
				"user_id": paymentMethod.UserID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeletePaymentMethodParams{
					ID:     paymentMethod.ID,
					UserID: paymentMethod.UserID,
				}

				store.EXPECT().
					DeletePaymentMethod(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(paymentMethod, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "NotFound",
			ID:   paymentMethod.ID,
			body: gin.H{
				"user_id": paymentMethod.UserID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeletePaymentMethodParams{
					ID:     paymentMethod.ID,
					UserID: paymentMethod.UserID,
				}

				store.EXPECT().
					DeletePaymentMethod(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.PaymentMethod{}, pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "InternalError",
			ID:   paymentMethod.ID,
			body: gin.H{
				"user_id": paymentMethod.UserID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeletePaymentMethodParams{
					ID:     paymentMethod.ID,
					UserID: paymentMethod.UserID,
				}

				store.EXPECT().
					DeletePaymentMethod(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.PaymentMethod{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidID",
			ID:   0,
			body: gin.H{
				"user_id": paymentMethod.UserID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeletePaymentMethod(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/users/payment-method/%d", tc.ID)
			request, err := http.NewRequest(http.MethodDelete, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			//check response
			tc.checkResponse(t, recorder)
		})

	}

}

func randomPMUser(t *testing.T) (user db.User, password string) {
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

func createRandomPaymentMethod(t *testing.T, user db.User, paymentType db.PaymentType) (paymentMethod db.PaymentMethod) {
	paymentMethod = db.PaymentMethod{
		ID:            util.RandomMoney(),
		UserID:        user.ID,
		PaymentTypeID: paymentType.ID,
		Provider:      util.RandomUser(),
		IsDefault:     util.RandomBool(),
	}
	return
}
func createRandomPaymentType(t *testing.T) (paymentType db.PaymentType) {
	paymentType = db.PaymentType{
		ID:    util.RandomMoney(),
		Value: util.RandomUser(),
	}
	return
}

func requireBodyMatchPaymentMethod(t *testing.T, body *bytes.Buffer, paymentMethod db.PaymentMethod) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotPaymentMethod db.PaymentMethod
	err = json.Unmarshal(data, &gotPaymentMethod)

	require.NoError(t, err)
	require.Equal(t, paymentMethod.UserID, gotPaymentMethod.UserID)
	require.Equal(t, paymentMethod.ID, gotPaymentMethod.ID)
	require.Equal(t, paymentMethod.PaymentTypeID, gotPaymentMethod.PaymentTypeID)
	require.Equal(t, paymentMethod.IsDefault, gotPaymentMethod.IsDefault)
	require.Equal(t, paymentMethod.Provider, gotPaymentMethod.Provider)
}

func requireBodyMatchPaymentMethodss(t *testing.T, body *bytes.Buffer, paymentMethods []db.PaymentMethod) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotPaymentMethods []db.PaymentMethod
	err = json.Unmarshal(data, &gotPaymentMethods)
	require.NoError(t, err)
	require.Equal(t, paymentMethods, gotPaymentMethods)
}
