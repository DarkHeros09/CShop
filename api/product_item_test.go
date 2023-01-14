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
	"github.com/gofiber/fiber/v2"
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
		checkResponse func(t *testing.T, rsp *http.Response)
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
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchProductItem(t, rsp.Body, productItem)
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
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
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
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
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

			// build stubs
			tc.buildStub(store)

			// start test server and send request
			server := newTestServer(t, store)
			//recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/api/v1/product-items/%d", tc.productItemID)
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(t, rsp)
		})

	}

}

func TestCreateProductItemAPI(t *testing.T) {
	admin, _ := randomProductItemSuperAdmin(t)
	productItem := randomProductItem()

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
				"product_id":     productItem.ProductID,
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
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchProductItem(t, rsp.Body, productItem)
			},
		},
		{
			name:    "NoAuthorization",
			AdminID: admin.ID,
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
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:    "Unauthorized",
			AdminID: admin.ID,
			body: fiber.Map{
				"product_id":     productItem.ProductID,
				"product_sku":   productItem.ProductSku,
				"qty_in_stock":  productItem.QtyInStock,
				"product_image": productItem.ProductImage,
				"price":         productItem.Price,
				"active":        productItem.Active,
			},
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
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:    "InternalError",
			AdminID: admin.ID,
			body: fiber.Map{
				"product_id":     productItem.ProductID,
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
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:    "InvalidID",
			AdminID: admin.ID,
			body: fiber.Map{
				"product_id":     0,
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
			tc.buildStubs(store)

			server := newTestServer(t, store)
			//recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/api/admin/%d/v1/product-items", tc.AdminID)
			request, err := http.NewRequest(fiber.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
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
		checkResponse func(rsp *http.Response)
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
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchProductItems(t, rsp.Body, productItems)
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
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
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
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusBadRequest, rsp.StatusCode)
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
			tc.buildStubs(store)

			server := newTestServer(t, store)
			//recorder := httptest.NewRecorder()

			url := "/api/v1/product-items"
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			// Add query parameters to request URL
			q := request.URL.Query()
			q.Add("page_id", fmt.Sprintf("%d", tc.query.pageID))
			q.Add("page_size", fmt.Sprintf("%d", tc.query.pageSize))
			request.URL.RawQuery = q.Encode()

			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
		})
	}
}

func TestUpdateProductItemAPI(t *testing.T) {
	admin, _ := randomProductItemSuperAdmin(t)
	productItem := randomProductItem()

	testCases := []struct {
		name          string
		body          fiber.Map
		productItemID int64
		AdminID       int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rsp *http.Response)
	}{
		{
			name:          "OK",
			productItemID: productItem.ID,
			AdminID:       admin.ID,
			body: fiber.Map{
				"product_id":     productItem.ProductID,
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
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:          "Unauthorized",
			productItemID: productItem.ID,
			AdminID:       admin.ID,
			body: fiber.Map{
				"product_id":     productItem.ProductID,
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
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:          "NoAuthorization",
			productItemID: productItem.ID,
			AdminID:       admin.ID,
			body: fiber.Map{
				"product_id":     productItem.ProductID,
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
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:          "InternalError",
			productItemID: productItem.ID,
			AdminID:       admin.ID,
			body: fiber.Map{
				"product_id":     productItem.ProductID,
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
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:          "InvalidID",
			productItemID: 0,
			AdminID:       admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateProductItem(gomock.Any(), gomock.Any()).
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
			tc.buildStubs(store)

			server := newTestServer(t, store)
			//recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/api/admin/%d/v1/product-items/%d", tc.AdminID, tc.productItemID)
			request, err := http.NewRequest(fiber.MethodPut, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(t, rsp)
		})
	}
}

func TestDeleteProductItemAPI(t *testing.T) {
	admin, _ := randomProductItemSuperAdmin(t)
	productItem := randomProductItem()

	testCases := []struct {
		name          string
		productItemID int64
		AdminID       int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rsp *http.Response)
	}{
		{
			name:          "OK",
			productItemID: productItem.ID,
			AdminID:       admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteProductItem(gomock.Any(), gomock.Eq(productItem.ID)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:          "Unauthorized",
			productItemID: productItem.ID,
			AdminID:       admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, 2, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteProductItem(gomock.Any(), gomock.Eq(productItem.ID)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:          "NoAuthorization",
			productItemID: productItem.ID,
			AdminID:       admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteProductItem(gomock.Any(), gomock.Eq(productItem.ID)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:          "NotFound",
			productItemID: productItem.ID,
			AdminID:       admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteProductItem(gomock.Any(), gomock.Eq(productItem.ID)).
					Times(1).
					Return(pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
			},
		},
		{
			name:          "InternalError",
			productItemID: productItem.ID,
			AdminID:       admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteProductItem(gomock.Any(), gomock.Eq(productItem.ID)).
					Times(1).
					Return(pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:          "InvalidID",
			productItemID: 0,
			AdminID:       admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteProductItem(gomock.Any(), gomock.Any()).
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

			// build stubs
			tc.buildStub(store)

			// start test server and send request
			server := newTestServer(t, store)
			//recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/api/admin/%d/v1/product-items/%d", tc.AdminID, tc.productItemID)
			request, err := http.NewRequest(fiber.MethodDelete, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(t, rsp)
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

func requireBodyMatchProductItem(t *testing.T, body io.ReadCloser, productItem db.ProductItem) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotProductItem db.ProductItem
	err = json.Unmarshal(data, &gotProductItem)
	require.NoError(t, err)
	require.Equal(t, productItem, gotProductItem)
}

func requireBodyMatchProductItems(t *testing.T, body io.ReadCloser, productItems []db.ProductItem) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotProductItems []db.ProductItem
	err = json.Unmarshal(data, &gotProductItems)
	require.NoError(t, err)
	require.Equal(t, productItems, gotProductItems)
}
