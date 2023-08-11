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
	"github.com/jackc/pgx/v5"
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
				"product_id":   productItem.ProductID,
				"product_sku":  productItem.ProductSku,
				"qty_in_stock": productItem.QtyInStock,
				"size_id":      productItem.SizeID,
				"color_id":     productItem.ColorID,
				"image_id":     productItem.ImageID,
				// "product_image": productItem.ProductImage,
				"price":  productItem.Price,
				"active": productItem.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateProductItemParams{
					ProductID:  productItem.ProductID,
					ProductSku: productItem.ProductSku,
					QtyInStock: productItem.QtyInStock,
					SizeID:     productItem.SizeID,
					ColorID:    productItem.ColorID,
					ImageID:    productItem.ImageID,
					Price:      productItem.Price,
					Active:     productItem.Active,
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
					ProductID:  productItem.ProductID,
					ProductSku: productItem.ProductSku,
					QtyInStock: productItem.QtyInStock,
					SizeID:     productItem.SizeID,
					ColorID:    productItem.ColorID,
					ImageID:    productItem.ImageID,
					Price:      productItem.Price,
					Active:     productItem.Active,
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
				"product_id":   productItem.ProductID,
				"product_sku":  productItem.ProductSku,
				"qty_in_stock": productItem.QtyInStock,
				"size_id":      productItem.SizeID,
				"color_id":     productItem.ColorID,
				"image_id":     productItem.ImageID,
				"price":        productItem.Price,
				"active":       productItem.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, false, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateProductItemParams{
					ProductID:  productItem.ProductID,
					ProductSku: productItem.ProductSku,
					QtyInStock: productItem.QtyInStock,
					SizeID:     productItem.SizeID,
					ColorID:    productItem.ColorID,
					ImageID:    productItem.ImageID,
					Price:      productItem.Price,
					Active:     productItem.Active,
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
				"product_id":   productItem.ProductID,
				"product_sku":  productItem.ProductSku,
				"qty_in_stock": productItem.QtyInStock,
				"size_id":      productItem.SizeID,
				"color_id":     productItem.ColorID,
				"image_id":     productItem.ImageID,
				"price":        productItem.Price,
				"active":       productItem.Active,
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
				"product_id":   0,
				"product_sku":  productItem.ProductSku,
				"qty_in_stock": productItem.QtyInStock,
				"size_id":      productItem.SizeID,
				"color_id":     productItem.ColorID,
				"image_id":     productItem.ImageID,
				"price":        productItem.Price,
				"active":       productItem.Active,
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

			url := fmt.Sprintf("/admin/%d/v1/product-items", tc.AdminID)
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
	productItems := make([]db.ListProductItemsRow, n)
	for i := 0; i < n; i++ {
		productItems[i] = randomProductItemList()
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
				requireBodyMatchProductItemsList(t, rsp.Body, productItems)
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
					Return([]db.ListProductItemsRow{}, pgx.ErrTxClosed)
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

func TestListProductItemsV2API(t *testing.T) {
	n := 10
	productItems := make([]db.ListProductItemsV2Row, n)
	for i := 0; i < n; i++ {
		productItems[i] = randomListProductItemV2()
	}

	type Query struct {
		Limit int
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
				Limit: n,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListProductItemsV2Params{
					Limit:      10,
					CategoryID: null.IntFromPtr(nil),
					BrandID:    null.IntFromPtr(nil),
					ColorID:    null.IntFromPtr(nil),
					SizeID:     null.IntFromPtr(nil),
					IsNew:      null.BoolFromPtr(nil),
					IsPromoted: null.BoolFromPtr(nil),
				}
				store.EXPECT().
					ListProductItemsV2(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(productItems, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchListProductItemsV2(t, rsp.Body, productItems)
			},
		},
		{
			name: "InternalError",
			query: Query{
				Limit: 10,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListProductItemsV2(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.ListProductItemsV2Row{}, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name: "InvalidLimit",
			query: Query{
				Limit: 11,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListProductItemsV2(gomock.Any(), gomock.Any()).
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

			url := "/api/v1/product-items-v2"
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			// Add query parameters to request URL
			q := request.URL.Query()
			q.Add("limit", fmt.Sprintf("%d", tc.query.Limit))
			request.URL.RawQuery = q.Encode()

			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
		})
	}
}

func TestListProductItemsNextPageAPI(t *testing.T) {
	n := 10
	productItems1 := make([]db.ListProductItemsNextPageRow, n)
	productItems2 := make([]db.ListProductItemsNextPageRow, n)
	for i := 0; i < n; i++ {
		productItems1[i] = randomRestProductItemSearch()
	}
	for i := 0; i < n; i++ {
		productItems2[i] = randomRestProductItemSearch()
	}

	type Query struct {
		Cursor int
		Limit  int
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
				Limit:  n,
				Cursor: int(productItems1[len(productItems1)-1].ID),
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.ListProductItemsNextPageParams{
					Limit:      int32(n),
					ID:         productItems1[len(productItems1)-1].ID,
					CategoryID: null.IntFromPtr(nil),
					BrandID:    null.IntFromPtr(nil),
					ColorID:    null.IntFromPtr(nil),
					SizeID:     null.IntFromPtr(nil),
					IsNew:      null.BoolFromPtr(nil),
					IsPromoted: null.BoolFromPtr(nil),
				}

				store.EXPECT().
					ListProductItemsNextPage(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(productItems2, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchListProductItemsNextPage(t, rsp.Body, productItems2)
			},
		},
		{
			name: "InternalError",
			query: Query{
				Limit:  10,
				Cursor: int(productItems1[len(productItems1)-1].ID),
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					ListProductItemsNextPage(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.ListProductItemsNextPageRow{}, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name: "InvalidLimit",
			query: Query{
				Limit:  11,
				Cursor: int(productItems1[len(productItems1)-1].ID),
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListProductItemsNextPage(gomock.Any(), gomock.Any()).
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

			url := "/api/v1/product-items-next-page"
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			// Add query parameters to request URL
			q := request.URL.Query()
			q.Add("limit", fmt.Sprintf("%d", tc.query.Limit))
			q.Add("cursor", fmt.Sprintf("%d", tc.query.Cursor))
			request.URL.RawQuery = q.Encode()

			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
		})
	}
}
func TestSearchProductItemsAPI(t *testing.T) {
	n := 10
	q := "abc"
	productItems := make([]db.SearchProductItemsRow, n)
	for i := 0; i < n; i++ {
		productItems[i] = randomSearchProductItems()
	}

	type Query struct {
		Limit int
		Query string
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
				Limit: n,
				Query: q,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.SearchProductItemsParams{
					Limit: int32(n),
					Query: q,
				}
				store.EXPECT().
					SearchProductItems(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(productItems, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchSearchProductItems(t, rsp.Body, productItems)
			},
		},
		{
			name: "InternalError",
			query: Query{
				Limit: n,
				Query: q,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					SearchProductItems(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.SearchProductItemsRow{}, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name: "InvalidLimit",
			query: Query{
				Limit: 11,
				Query: q,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					SearchProductItems(gomock.Any(), gomock.Any()).
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

			url := "/api/v1/search-product-items"
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			// Add query parameters to request URL
			q := request.URL.Query()
			q.Add("limit", fmt.Sprintf("%d", tc.query.Limit))
			q.Add("query", tc.query.Query)
			request.URL.RawQuery = q.Encode()

			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
		})
	}
}

func TestSearchProductItemsNextPageAPI(t *testing.T) {
	n := 10
	q := "abc"
	productItems1 := make([]db.SearchProductItemsNextPageRow, n)
	productItems2 := make([]db.SearchProductItemsNextPageRow, n)
	for i := 0; i < n; i++ {
		productItems1[i] = randomSearchProductItemNextPage()
	}
	for i := 0; i < n; i++ {
		productItems2[i] = randomSearchProductItemNextPage()
	}

	type Query struct {
		Cursor int
		Limit  int
		Query  string
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
				Limit:  n,
				Query:  q,
				Cursor: int(productItems1[len(productItems1)-1].ID),
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.SearchProductItemsNextPageParams{
					Limit: int32(n),
					ID:    productItems1[len(productItems1)-1].ID,
					Query: q,
				}

				store.EXPECT().
					SearchProductItemsNextPage(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(productItems2, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchSearchProductItemsNextPage(t, rsp.Body, productItems2)
			},
		},
		{
			name: "InternalError",
			query: Query{
				Limit:  10,
				Cursor: int(productItems1[len(productItems1)-1].ID),
				Query:  q,
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					SearchProductItemsNextPage(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.SearchProductItemsNextPageRow{}, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name: "InvalidLimit",
			query: Query{
				Limit:  11,
				Cursor: int(productItems1[len(productItems1)-1].ID),
				Query:  q,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					SearchProductItemsNextPage(gomock.Any(), gomock.Any()).
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

			url := "/api/v1/search-product-items-next-page"
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			// Add query parameters to request URL
			q := request.URL.Query()
			q.Add("limit", fmt.Sprintf("%d", tc.query.Limit))
			q.Add("cursor", fmt.Sprintf("%d", tc.query.Cursor))
			q.Add("query", tc.query.Query)
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
				"product_id":   productItem.ProductID,
				"product_sku":  productItem.ProductSku,
				"qty_in_stock": productItem.QtyInStock,
				// "product_image": productItem.ProductImage,
				"size_id":  productItem.SizeID,
				"color_id": productItem.ColorID,
				"image_id": productItem.ImageID,
				"price":    "1000",
				"active":   productItem.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateProductItemParams{
					ProductID:  productItem.ProductID,
					ProductSku: null.IntFrom(productItem.ProductSku),
					QtyInStock: null.IntFrom(int64(productItem.QtyInStock)),
					// ProductImage: null.StringFrom(productItem.ProductImage),
					Price:  null.StringFrom("1000"),
					Active: null.BoolFromPtr(&productItem.Active),
					ID:     productItem.ID,
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
				"product_id":   productItem.ProductID,
				"product_sku":  productItem.ProductSku,
				"qty_in_stock": productItem.QtyInStock,
				// "product_image": productItem.ProductImage,
				"size_id":  productItem.SizeID,
				"color_id": productItem.ColorID,
				"image_id": productItem.ImageID,
				"price":    "1000",
				"active":   productItem.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, false, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateProductItemParams{
					ProductID:  productItem.ProductID,
					ProductSku: null.IntFrom(productItem.ProductSku),
					QtyInStock: null.IntFrom(int64(productItem.QtyInStock)),
					// ProductImage: null.StringFrom(productItem.ProductImage),
					Price:  null.StringFrom("1000"),
					Active: null.BoolFromPtr(&productItem.Active),
					ID:     productItem.ID,
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
				"product_id":   productItem.ProductID,
				"product_sku":  productItem.ProductSku,
				"qty_in_stock": productItem.QtyInStock,
				// "product_image": productItem.ProductImage,
				"size_id":  productItem.SizeID,
				"color_id": productItem.ColorID,
				"image_id": productItem.ImageID,
				"price":    "1000",
				"active":   productItem.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateProductItemParams{
					ProductID:  productItem.ProductID,
					ProductSku: null.IntFrom(productItem.ProductSku),
					QtyInStock: null.IntFrom(int64(productItem.QtyInStock)),
					// ProductImage: null.StringFrom(productItem.ProductImage),
					Price:  null.StringFrom("1000"),
					Active: null.BoolFromPtr(&productItem.Active),
					ID:     productItem.ID,
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
				"product_id":   productItem.ProductID,
				"product_sku":  productItem.ProductSku,
				"qty_in_stock": productItem.QtyInStock,
				// "product_image": productItem.ProductImage,
				"size_id":  productItem.SizeID,
				"color_id": productItem.ColorID,
				"image_id": productItem.ImageID,
				"price":    "1000",
				"active":   productItem.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateProductItemParams{
					ProductID:  productItem.ProductID,
					ProductSku: null.IntFrom(productItem.ProductSku),
					QtyInStock: null.IntFrom(int64(productItem.QtyInStock)),
					// ProductImage: null.StringFrom(productItem.ProductImage),
					Price:  null.StringFrom("1000"),
					Active: null.BoolFromPtr(&productItem.Active),
					ID:     productItem.ID,
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

			url := fmt.Sprintf("/admin/%d/v1/product-items/%d", tc.AdminID, tc.productItemID)
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

			url := fmt.Sprintf("/admin/%d/v1/product-items/%d", tc.AdminID, tc.productItemID)
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
		ID:         util.RandomInt(1, 1000),
		ProductID:  util.RandomMoney(),
		ProductSku: util.RandomMoney(),
		QtyInStock: int32(util.RandomMoney()),
		SizeID:     util.RandomMoney(),
		ImageID:    util.RandomMoney(),
		ColorID:    util.RandomMoney(),
		Price:      util.RandomDecimalString(0, 1000),
		Active:     util.RandomBool(),
	}
}

func randomProductItemNew() db.ListProductItemsV2Row {
	return db.ListProductItemsV2Row{
		ID:            util.RandomInt(1, 1000),
		ProductID:     util.RandomMoney(),
		ProductSku:    util.RandomMoney(),
		QtyInStock:    int32(util.RandomMoney()),
		SizeID:        util.RandomMoney(),
		SizeValue:     null.StringFrom(util.RandomUser()),
		ImageID:       util.RandomMoney(),
		ProductImage1: null.StringFrom(util.RandomURL()),
		ProductImage2: null.StringFrom(util.RandomURL()),
		ProductImage3: null.StringFrom(util.RandomURL()),
		ColorID:       util.RandomMoney(),
		ColorValue:    null.StringFrom(util.RandomUser()),
		Price:         util.RandomDecimalString(0, 1000),
		Active:        util.RandomBool(),
	}
}

func randomProductItemList() db.ListProductItemsRow {
	return db.ListProductItemsRow{
		ID:            util.RandomInt(1, 1000),
		ProductID:     util.RandomMoney(),
		ProductSku:    util.RandomMoney(),
		QtyInStock:    int32(util.RandomMoney()),
		SizeValue:     null.StringFrom(util.RandomUser()),
		ColorValue:    null.StringFrom(util.RandomUser()),
		ProductImage1: null.StringFrom(util.RandomURL()),
		ProductImage2: null.StringFrom(util.RandomURL()),
		ProductImage3: null.StringFrom(util.RandomURL()),
		Price:         util.RandomDecimalString(0, 1000),
		Active:        util.RandomBool(),
		Name:          null.StringFrom(util.RandomUser()),
		TotalCount:    util.RandomMoney(),
	}
}

func randomListProductItemV2() db.ListProductItemsV2Row {
	return db.ListProductItemsV2Row{
		ID:            util.RandomInt(1, 1000),
		ProductID:     util.RandomMoney(),
		ProductSku:    util.RandomMoney(),
		QtyInStock:    int32(util.RandomMoney()),
		SizeValue:     null.StringFrom(util.RandomUser()),
		ColorValue:    null.StringFrom(util.RandomUser()),
		ProductImage1: null.StringFrom(util.RandomURL()),
		ProductImage2: null.StringFrom(util.RandomURL()),
		ProductImage3: null.StringFrom(util.RandomURL()),
		Price:         util.RandomDecimalString(0, 1000),
		Active:        util.RandomBool(),
		Name:          null.StringFrom(util.RandomUser()),
	}
}

func randomRestProductItemSearch() db.ListProductItemsNextPageRow {
	return db.ListProductItemsNextPageRow{
		ID:            util.RandomInt(1, 1000),
		ProductID:     util.RandomMoney(),
		ProductSku:    util.RandomMoney(),
		QtyInStock:    int32(util.RandomMoney()),
		SizeValue:     null.StringFrom(util.RandomUser()),
		ColorValue:    null.StringFrom(util.RandomUser()),
		ProductImage1: null.StringFrom(util.RandomURL()),
		ProductImage2: null.StringFrom(util.RandomURL()),
		ProductImage3: null.StringFrom(util.RandomURL()),
		Price:         util.RandomDecimalString(0, 1000),
		Active:        util.RandomBool(),
		Name:          null.StringFrom(util.RandomUser()),
	}
}

func randomSearchProductItems() db.SearchProductItemsRow {
	return db.SearchProductItemsRow{
		ID:            util.RandomInt(1, 1000),
		ProductID:     util.RandomMoney(),
		ProductSku:    util.RandomMoney(),
		QtyInStock:    int32(util.RandomMoney()),
		SizeValue:     null.StringFrom(util.RandomUser()),
		ColorValue:    null.StringFrom(util.RandomUser()),
		ProductImage1: null.StringFrom(util.RandomURL()),
		ProductImage2: null.StringFrom(util.RandomURL()),
		ProductImage3: null.StringFrom(util.RandomURL()),
		Price:         util.RandomDecimalString(0, 1000),
		Active:        util.RandomBool(),
		Name:          null.StringFrom("abc" + util.RandomUser()),
	}
}

func randomSearchProductItemNextPage() db.SearchProductItemsNextPageRow {
	return db.SearchProductItemsNextPageRow{
		ID:            util.RandomInt(1, 1000),
		ProductID:     util.RandomMoney(),
		ProductSku:    util.RandomMoney(),
		QtyInStock:    int32(util.RandomMoney()),
		SizeValue:     null.StringFrom(util.RandomUser()),
		ColorValue:    null.StringFrom(util.RandomUser()),
		ProductImage1: null.StringFrom(util.RandomURL()),
		ProductImage2: null.StringFrom(util.RandomURL()),
		ProductImage3: null.StringFrom(util.RandomURL()),
		Price:         util.RandomDecimalString(0, 1000),
		Active:        util.RandomBool(),
		Name:          null.StringFrom("abc" + util.RandomUser()),
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

func requireBodyMatchProductItemsList(t *testing.T, body io.ReadCloser, productItems []db.ListProductItemsRow) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotProductItems []db.ListProductItemsRow
	err = json.Unmarshal(data, &gotProductItems)
	require.NoError(t, err)
	require.Equal(t, productItems, gotProductItems)
}

func requireBodyMatchListProductItemsV2(t *testing.T, body io.ReadCloser, productItems []db.ListProductItemsV2Row) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotProductItems []db.ListProductItemsV2Row
	err = json.Unmarshal(data, &gotProductItems)
	require.NoError(t, err)
	require.Equal(t, productItems, gotProductItems)
}

func requireBodyMatchListProductItemsNextPage(t *testing.T, body io.ReadCloser, productItems []db.ListProductItemsNextPageRow) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotProductItems []db.ListProductItemsNextPageRow
	err = json.Unmarshal(data, &gotProductItems)
	require.NoError(t, err)
	require.Equal(t, productItems, gotProductItems)
}

func requireBodyMatchSearchProductItems(t *testing.T, body io.ReadCloser, productItems []db.SearchProductItemsRow) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotProductItems []db.SearchProductItemsRow
	err = json.Unmarshal(data, &gotProductItems)
	require.NoError(t, err)
	require.Equal(t, productItems, gotProductItems)
}

func requireBodyMatchSearchProductItemsNextPage(t *testing.T, body io.ReadCloser, productItems []db.SearchProductItemsNextPageRow) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotProductItems []db.SearchProductItemsNextPageRow
	err = json.Unmarshal(data, &gotProductItems)
	require.NoError(t, err)
	require.Equal(t, productItems, gotProductItems)
}
