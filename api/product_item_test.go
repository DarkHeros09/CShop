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

func TestGetProductItemAPI(t *testing.T) {
	productItem := randomProductItem()

	testCases := []struct {
		name          string
		productItemID int64
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:          "OK",
			productItemID: productItem.ID,
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetProductItem(gomock.Any(), gomock.Eq(productItem.ID)).
					Times(1).
					Return(productItem, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchProductItem(t, recorder.Body, productItem)
			},
		},

		{
			name:          "NotFound",
			productItemID: productItem.ID,
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetProductItem(gomock.Any(), gomock.Eq(productItem.ID)).
					Times(1).
					Return(db.ProductItem{}, pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:          "InternalError",
			productItemID: productItem.ID,
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetProductItem(gomock.Any(), gomock.Eq(productItem.ID)).
					Times(1).
					Return(db.ProductItem{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:          "InvalidID",
			productItemID: 0,
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetProductItem(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/product-items/%d", tc.productItemID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			//check response
			tc.checkResponse(t, recorder)
		})

	}

}

func TestCreateProductItemAPI(t *testing.T) {
	admin, _ := randomProductItemSuperAdmin(t)
	productItem := randomProductItem()

	testCases := []struct {
		name          string
		body          gin.H
		productItemID int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"product_id":    productItem.ProductID,
				"product_sku":   productItem.ProductSku,
				"qty_in_stock":  productItem.QtyInStock,
				"product_image": productItem.ProductImage,
				"price":         productItem.Price,
				"active":        productItem.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateProductItemParams{
					ProductID:    productItem.ProductID,
					ProductSku:   productItem.ProductSku,
					QtyInStock:   productItem.QtyInStock,
					ProductImage: productItem.ProductImage,
					Price:        productItem.Price,
					Active:       productItem.Active,
				}

				store.EXPECT().
					CreateProductItem(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(productItem, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchProductItem(t, recorder.Body, productItem)
			},
		},
		{
			name:          "NoAuthorization",
			productItemID: productItem.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateProductItemParams{
					ProductID:    productItem.ProductID,
					ProductSku:   productItem.ProductSku,
					QtyInStock:   productItem.QtyInStock,
					ProductImage: productItem.ProductImage,
					Price:        productItem.Price,
					Active:       productItem.Active,
				}

				store.EXPECT().
					CreateProductItem(gomock.Any(), gomock.Eq(arg)).
					Times(0)

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:          "Unauthorized",
			productItemID: productItem.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, false, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateProductItemParams{
					ProductID:    productItem.ProductID,
					ProductSku:   productItem.ProductSku,
					QtyInStock:   productItem.QtyInStock,
					ProductImage: productItem.ProductImage,
					Price:        productItem.Price,
					Active:       productItem.Active,
				}

				store.EXPECT().
					CreateProductItem(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"product_id":    productItem.ProductID,
				"product_sku":   productItem.ProductSku,
				"qty_in_stock":  productItem.QtyInStock,
				"product_image": productItem.ProductImage,
				"price":         productItem.Price,
				"active":        productItem.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateProductItem(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.ProductItem{}, pgx.ErrTxClosed)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidID",
			body: gin.H{
				"product_id":    0,
				"product_sku":   productItem.ProductSku,
				"qty_in_stock":  productItem.QtyInStock,
				"product_image": productItem.ProductImage,
				"price":         productItem.Price,
				"active":        productItem.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateProductItem(gomock.Any(), gomock.Any()).
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

			url := "/product-items"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestListProductItemsAPI(t *testing.T) {
	n := 5
	productItems := make([]db.ProductItem, n)
	for i := 0; i < n; i++ {
		productItems[i] = randomProductItem()
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
				arg := db.ListProductItemsParams{
					Limit:  int32(n),
					Offset: 0,
				}

				store.EXPECT().
					ListProductItems(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(productItems, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchProductItems(t, recorder.Body, productItems)
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
					ListProductItems(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.ProductItem{}, pgx.ErrTxClosed)
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
					ListProductItems(gomock.Any(), gomock.Any()).
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
					ListProductItems(gomock.Any(), gomock.Any()).
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

			url := "/product-items"
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

func TestUpdateProductItemAPI(t *testing.T) {
	admin, _ := randomProductItemSuperAdmin(t)
	productItem := randomProductItem()

	testCases := []struct {
		name          string
		body          gin.H
		productItemID int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:          "OK",
			productItemID: productItem.ID,
			body: gin.H{
				"product_id":    productItem.ProductID,
				"product_sku":   productItem.ProductSku,
				"qty_in_stock":  productItem.QtyInStock,
				"product_image": productItem.ProductImage,
				"price":         "1000",
				"active":        productItem.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateProductItemParams{
					ProductID:    productItem.ProductID,
					ProductSku:   null.IntFrom(productItem.ProductSku),
					QtyInStock:   null.IntFrom(int64(productItem.QtyInStock)),
					ProductImage: null.StringFrom(productItem.ProductImage),
					Price:        null.StringFrom("1000"),
					Active:       null.BoolFromPtr(&productItem.Active),
					ID:           productItem.ID,
				}

				store.EXPECT().
					UpdateProductItem(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(productItem, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:          "Unauthorized",
			productItemID: productItem.ID,
			body: gin.H{
				"product_id":    productItem.ProductID,
				"product_sku":   productItem.ProductSku,
				"qty_in_stock":  productItem.QtyInStock,
				"product_image": productItem.ProductImage,
				"price":         "1000",
				"active":        productItem.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, false, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateProductItemParams{
					ProductID:    productItem.ProductID,
					ProductSku:   null.IntFrom(productItem.ProductSku),
					QtyInStock:   null.IntFrom(int64(productItem.QtyInStock)),
					ProductImage: null.StringFrom(productItem.ProductImage),
					Price:        null.StringFrom("1000"),
					Active:       null.BoolFromPtr(&productItem.Active),
					ID:           productItem.ID,
				}

				store.EXPECT().
					UpdateProductItem(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:          "NoAuthorization",
			productItemID: productItem.ID,
			body: gin.H{
				"product_id":    productItem.ProductID,
				"product_sku":   productItem.ProductSku,
				"qty_in_stock":  productItem.QtyInStock,
				"product_image": productItem.ProductImage,
				"price":         "1000",
				"active":        productItem.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateProductItemParams{
					ProductID:    productItem.ProductID,
					ProductSku:   null.IntFrom(productItem.ProductSku),
					QtyInStock:   null.IntFrom(int64(productItem.QtyInStock)),
					ProductImage: null.StringFrom(productItem.ProductImage),
					Price:        null.StringFrom("1000"),
					Active:       null.BoolFromPtr(&productItem.Active),
					ID:           productItem.ID,
				}

				store.EXPECT().
					UpdateProductItem(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:          "InternalError",
			productItemID: productItem.ID,
			body: gin.H{
				"product_id":    productItem.ProductID,
				"product_sku":   productItem.ProductSku,
				"qty_in_stock":  productItem.QtyInStock,
				"product_image": productItem.ProductImage,
				"price":         "1000",
				"active":        productItem.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateProductItemParams{
					ProductID:    productItem.ProductID,
					ProductSku:   null.IntFrom(productItem.ProductSku),
					QtyInStock:   null.IntFrom(int64(productItem.QtyInStock)),
					ProductImage: null.StringFrom(productItem.ProductImage),
					Price:        null.StringFrom("1000"),
					Active:       null.BoolFromPtr(&productItem.Active),
					ID:           productItem.ID,
				}
				store.EXPECT().
					UpdateProductItem(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.ProductItem{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:          "InvalidID",
			productItemID: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateProductItem(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/product-items/%d", tc.productItemID)
			request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestDeleteProductItemAPI(t *testing.T) {
	admin, _ := randomProductItemSuperAdmin(t)
	productItem := randomProductItem()

	testCases := []struct {
		name          string
		body          gin.H
		productItemID int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:          "OK",
			productItemID: productItem.ID,
			body: gin.H{
				"id": productItem.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteProductItem(gomock.Any(), gomock.Eq(productItem.ID)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:          "Unauthorized",
			productItemID: productItem.ID,
			body: gin.H{
				"id": productItem.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, 2, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteProductItem(gomock.Any(), gomock.Eq(productItem.ID)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:          "NoAuthorization",
			productItemID: productItem.ID,
			body: gin.H{
				"id": productItem.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteProductItem(gomock.Any(), gomock.Eq(productItem.ID)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:          "NotFound",
			productItemID: productItem.ID,
			body: gin.H{
				"id": productItem.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteProductItem(gomock.Any(), gomock.Eq(productItem.ID)).
					Times(1).
					Return(pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:          "InternalError",
			productItemID: productItem.ID,
			body: gin.H{
				"id": productItem.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteProductItem(gomock.Any(), gomock.Eq(productItem.ID)).
					Times(1).
					Return(pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:          "InvalidID",
			productItemID: 0,
			body: gin.H{
				"id": 0,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteProductItem(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/product-items/%d", tc.productItemID)
			request, err := http.NewRequest(http.MethodDelete, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			//check response
			tc.checkResponse(t, recorder)
		})

	}

}

func randomProductItemSuperAdmin(t *testing.T) (admin db.Admin, password string) {
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

func randomProductItem() db.ProductItem {
	return db.ProductItem{
		ID:           util.RandomInt(1, 1000),
		ProductID:    util.RandomMoney(),
		ProductSku:   util.RandomMoney(),
		QtyInStock:   int32(util.RandomMoney()),
		ProductImage: util.RandomURL(),
		Price:        util.RandomDecimalString(0, 1000),
		Active:       util.RandomBool(),
	}
}

func requireBodyMatchProductItem(t *testing.T, body *bytes.Buffer, productItem db.ProductItem) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotProductItem db.ProductItem
	err = json.Unmarshal(data, &gotProductItem)
	require.NoError(t, err)
	require.Equal(t, productItem, gotProductItem)
}

func requireBodyMatchProductItems(t *testing.T, body *bytes.Buffer, productItems []db.ProductItem) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotProductItems []db.ProductItem
	err = json.Unmarshal(data, &gotProductItems)
	require.NoError(t, err)
	require.Equal(t, productItems, gotProductItems)
}
