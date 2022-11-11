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

func TestGetProductPromotionAPI(t *testing.T) {
	productPromotion := randomProductPromotion()

	testCases := []struct {
		name          string
		productID     int64
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			productID: productPromotion.ProductID,
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetProductPromotion(gomock.Any(), gomock.Eq(productPromotion.ProductID)).
					Times(1).
					Return(productPromotion, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchProductPromotion(t, recorder.Body, productPromotion)
			},
		},

		{
			name:      "NotFound",
			productID: productPromotion.ProductID,
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetProductPromotion(gomock.Any(), gomock.Eq(productPromotion.ProductID)).
					Times(1).
					Return(db.ProductPromotion{}, pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			productID: productPromotion.ProductID,
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetProductPromotion(gomock.Any(), gomock.Eq(productPromotion.ProductID)).
					Times(1).
					Return(db.ProductPromotion{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "InvalidID",
			productID: 0,
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetProductPromotion(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/product-promotions/%d", tc.productID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			//check response
			tc.checkResponse(t, recorder)
		})

	}

}

func TestCreateProductPromotionAPI(t *testing.T) {
	admin, _ := randomProductPromotionSuperAdmin(t)
	productPromotion := randomProductPromotion()

	testCases := []struct {
		name          string
		body          gin.H
		productID     int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"product_id":   productPromotion.ProductID,
				"promotion_id": productPromotion.PromotionID,
				"active":       productPromotion.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateProductPromotionParams{
					ProductID:   productPromotion.ProductID,
					PromotionID: productPromotion.PromotionID,
					Active:      productPromotion.Active,
				}

				store.EXPECT().
					CreateProductPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(productPromotion, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchProductPromotion(t, recorder.Body, productPromotion)
			},
		},
		{
			name:      "NoAuthorization",
			productID: productPromotion.ProductID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateProductPromotionParams{
					ProductID:   productPromotion.ProductID,
					PromotionID: productPromotion.PromotionID,
					Active:      productPromotion.Active,
				}

				store.EXPECT().
					CreateProductPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(0)

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:      "Unauthorized",
			productID: productPromotion.ProductID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, false, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateProductPromotionParams{
					ProductID:   productPromotion.ProductID,
					PromotionID: productPromotion.PromotionID,
					Active:      productPromotion.Active,
				}

				store.EXPECT().
					CreateProductPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"product_id":   productPromotion.ProductID,
				"promotion_id": productPromotion.PromotionID,
				"active":       productPromotion.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateProductPromotion(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.ProductPromotion{}, pgx.ErrTxClosed)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidID",
			body: gin.H{
				"product_id":   "invalid",
				"promotion_id": productPromotion.PromotionID,
				"active":       productPromotion.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateProductPromotion(gomock.Any(), gomock.Any()).
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

			url := "/product-promotions"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestListProductPromotionsAPI(t *testing.T) {
	n := 5
	productPromotions := make([]db.ProductPromotion, n)
	for i := 0; i < n; i++ {
		productPromotions[i] = randomProductPromotion()
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
				arg := db.ListProductPromotionsParams{
					Limit:  int32(n),
					Offset: 0,
				}

				store.EXPECT().
					ListProductPromotions(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(productPromotions, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchProductPromotions(t, recorder.Body, productPromotions)
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
					ListProductPromotions(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.ProductPromotion{}, pgx.ErrTxClosed)
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
					ListProductPromotions(gomock.Any(), gomock.Any()).
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
					ListProductPromotions(gomock.Any(), gomock.Any()).
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

			url := "/product-promotions"
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

func TestUpdateProductPromotionAPI(t *testing.T) {
	admin, _ := randomProductPromotionSuperAdmin(t)
	productPromotion := randomProductPromotion()

	testCases := []struct {
		name          string
		body          gin.H
		productID     int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			productID: productPromotion.ProductID,
			body: gin.H{
				"product_id":   productPromotion.ProductID,
				"promotion_id": productPromotion.PromotionID,
				"active":       productPromotion.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateProductPromotionParams{
					PromotionID: null.IntFromPtr(&productPromotion.PromotionID),
					Active:      null.BoolFromPtr(&productPromotion.Active),
					ProductID:   productPromotion.ProductID,
				}

				store.EXPECT().
					UpdateProductPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(productPromotion, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:      "Unauthorized",
			productID: productPromotion.ProductID,
			body: gin.H{
				"product_id":   productPromotion.ProductID,
				"promotion_id": productPromotion.PromotionID,
				"active":       productPromotion.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, false, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateProductPromotionParams{
					PromotionID: null.IntFromPtr(&productPromotion.PromotionID),
					Active:      null.BoolFromPtr(&productPromotion.Active),
					ProductID:   productPromotion.ProductID,
				}

				store.EXPECT().
					UpdateProductPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:      "NoAuthorization",
			productID: productPromotion.ProductID,
			body: gin.H{
				"product_id":   productPromotion.ProductID,
				"promotion_id": productPromotion.PromotionID,
				"active":       productPromotion.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateProductPromotionParams{
					PromotionID: null.IntFromPtr(&productPromotion.PromotionID),
					Active:      null.BoolFromPtr(&productPromotion.Active),
					ProductID:   productPromotion.ProductID,
				}

				store.EXPECT().
					UpdateProductPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			productID: productPromotion.ProductID,
			body: gin.H{
				"product_id":   productPromotion.ProductID,
				"promotion_id": productPromotion.PromotionID,
				"active":       productPromotion.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateProductPromotionParams{
					PromotionID: null.IntFromPtr(&productPromotion.PromotionID),
					Active:      null.BoolFromPtr(&productPromotion.Active),
					ProductID:   productPromotion.ProductID,
				}
				store.EXPECT().
					UpdateProductPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.ProductPromotion{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "InvalidID",
			productID: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateProductPromotion(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/product-promotions/%d", tc.productID)
			request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestDeleteProductPromotionAPI(t *testing.T) {
	admin, _ := randomProductPromotionSuperAdmin(t)
	productPromotion := randomProductPromotion()

	testCases := []struct {
		name          string
		body          gin.H
		productID     int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			productID: productPromotion.ProductID,
			body: gin.H{
				"id": productPromotion.ProductID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteProductPromotion(gomock.Any(), gomock.Eq(productPromotion.ProductID)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:      "Unauthorized",
			productID: productPromotion.ProductID,
			body: gin.H{
				"id": productPromotion.ProductID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, 2, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteProductPromotion(gomock.Any(), gomock.Eq(productPromotion.ProductID)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:      "NoAuthorization",
			productID: productPromotion.ProductID,
			body: gin.H{
				"id": productPromotion.ProductID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteProductPromotion(gomock.Any(), gomock.Eq(productPromotion.ProductID)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:      "NotFound",
			productID: productPromotion.ProductID,
			body: gin.H{
				"id": productPromotion.ProductID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteProductPromotion(gomock.Any(), gomock.Eq(productPromotion.ProductID)).
					Times(1).
					Return(pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			productID: productPromotion.ProductID,
			body: gin.H{
				"id": productPromotion.ProductID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteProductPromotion(gomock.Any(), gomock.Eq(productPromotion.ProductID)).
					Times(1).
					Return(pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "InvalidID",
			productID: 0,
			body: gin.H{
				"id": 0,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteProductPromotion(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/product-promotions/%d", tc.productID)
			request, err := http.NewRequest(http.MethodDelete, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			//check response
			tc.checkResponse(t, recorder)
		})

	}

}

func randomProductPromotionSuperAdmin(t *testing.T) (admin db.Admin, password string) {
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

func randomProductPromotion() db.ProductPromotion {
	return db.ProductPromotion{
		ProductID:   util.RandomMoney(),
		PromotionID: util.RandomMoney(),
		Active:      util.RandomBool(),
	}
}

func requireBodyMatchProductPromotion(t *testing.T, body *bytes.Buffer, productPromotion db.ProductPromotion) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotProductPromotion db.ProductPromotion
	err = json.Unmarshal(data, &gotProductPromotion)
	require.NoError(t, err)
	require.Equal(t, productPromotion, gotProductPromotion)
}

func requireBodyMatchProductPromotions(t *testing.T, body *bytes.Buffer, productPromotions []db.ProductPromotion) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotProductPromotions []db.ProductPromotion
	err = json.Unmarshal(data, &gotProductPromotions)
	require.NoError(t, err)
	require.Equal(t, productPromotions, gotProductPromotions)
}
