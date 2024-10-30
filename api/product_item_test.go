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
	"github.com/guregu/null/v5"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
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
			worker := mockwk.NewMockTaskDistributor(ctrl)
			ik := mockik.NewMockImageKitManagement(ctrl)
			mailSender := mockemail.NewMockEmailSender(ctrl)

			// build stubs
			tc.buildStub(store)

			// start test server and send request
			server := newTestServer(t, store, worker, ik, mailSender)
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
				arg := db.AdminCreateProductItemParams{
					AdminID:    admin.ID,
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
					AdminCreateProductItem(gomock.Any(), gomock.Eq(arg)).
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
				arg := db.AdminCreateProductItemParams{
					AdminID:    admin.ID,
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
					AdminCreateProductItem(gomock.Any(), gomock.Eq(arg)).
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
				arg := db.AdminCreateProductItemParams{
					AdminID:    admin.ID,
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
					AdminCreateProductItem(gomock.Any(), gomock.Eq(arg)).
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
					AdminCreateProductItem(gomock.Any(), gomock.Any()).
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
					AdminCreateProductItem(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/admin/v1/admins/%d/product-items", tc.AdminID)
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
			worker := mockwk.NewMockTaskDistributor(ctrl)
			ik := mockik.NewMockImageKitManagement(ctrl)
			mailSender := mockemail.NewMockEmailSender(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store, worker, ik, mailSender)
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

func TestListProductItemsWithBestSalesAPI(t *testing.T) {
	n := 10
	productItems := make([]db.ListProductItemsWithBestSalesRow, n)
	for i := 0; i < n; i++ {
		productItems[i] = randomListProductItemWithBestSales()
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
				arg := int32(n)
				store.EXPECT().
					ListProductItemsWithBestSales(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(productItems, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchListProductItemsWithBestSales(t, rsp.Body, productItems)
			},
		},
		{
			name: "InternalError",
			query: Query{
				Limit: 10,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListProductItemsWithBestSales(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.ListProductItemsWithBestSalesRow{}, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name: "InvalidLimit",
			query: Query{
				Limit: 0,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListProductItemsWithBestSales(gomock.Any(), gomock.Any()).
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

			url := "/api/v1/products-best-sellers"
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
		productItems1[i] = randomRestProductItemList()
	}
	for i := 0; i < n; i++ {
		productItems2[i] = randomRestProductItemList()
	}

	type Query struct {
		ProductItemCursor int
		ProductCursor     int
		Limit             int
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
				Limit:             n,
				ProductItemCursor: int(productItems1[len(productItems1)-1].ID),
				ProductCursor:     int(productItems1[len(productItems1)-1].ProductID),
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.ListProductItemsNextPageParams{
					Limit:            int32(n),
					ProductItemID:    productItems1[len(productItems1)-1].ID,
					ProductID:        productItems1[len(productItems1)-1].ProductID,
					CategoryID:       null.IntFromPtr(nil),
					BrandID:          null.IntFromPtr(nil),
					ColorID:          null.IntFromPtr(nil),
					SizeID:           null.IntFromPtr(nil),
					IsNew:            null.BoolFromPtr(nil),
					IsPromoted:       null.BoolFromPtr(nil),
					IsFeatured:       null.BoolFromPtr(nil),
					IsQtyLimited:     null.BoolFromPtr(nil),
					OrderByLowPrice:  null.BoolFromPtr(nil),
					OrderByHighPrice: null.BoolFromPtr(nil),
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
				Limit:             10,
				ProductItemCursor: int(productItems1[len(productItems1)-1].ID),
				ProductCursor:     int(productItems1[len(productItems1)-1].ProductID),
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
				Limit:             11,
				ProductItemCursor: int(productItems1[len(productItems1)-1].ID),
				ProductCursor:     int(productItems1[len(productItems1)-1].ProductID),
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
			worker := mockwk.NewMockTaskDistributor(ctrl)
			ik := mockik.NewMockImageKitManagement(ctrl)
			mailSender := mockemail.NewMockEmailSender(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store, worker, ik, mailSender)
			//recorder := httptest.NewRecorder()

			url := "/api/v1/product-items-next-page"
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			// Add query parameters to request URL
			q := request.URL.Query()
			q.Add("limit", fmt.Sprintf("%d", tc.query.Limit))
			q.Add("product_item_cursor", fmt.Sprintf("%d", tc.query.ProductItemCursor))
			q.Add("product_cursor", fmt.Sprintf("%d", tc.query.ProductCursor))
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
			worker := mockwk.NewMockTaskDistributor(ctrl)
			ik := mockik.NewMockImageKitManagement(ctrl)
			mailSender := mockemail.NewMockEmailSender(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store, worker, ik, mailSender)
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
		ProductItemCursor int
		ProductCursor     int
		Limit             int
		Query             string
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
				Limit:             n,
				Query:             q,
				ProductItemCursor: int(productItems1[len(productItems1)-1].ID),
				ProductCursor:     int(productItems1[len(productItems1)-1].ProductID),
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.SearchProductItemsNextPageParams{
					Limit:         int32(n),
					ProductItemID: productItems1[len(productItems1)-1].ID,
					ProductID:     productItems1[len(productItems1)-1].ProductID,
					Query:         q,
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
				Limit:             10,
				ProductItemCursor: int(productItems1[len(productItems1)-1].ID),
				ProductCursor:     int(productItems1[len(productItems1)-1].ProductID),
				Query:             q,
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
				Limit:             11,
				ProductItemCursor: int(productItems1[len(productItems1)-1].ID),
				ProductCursor:     int(productItems1[len(productItems1)-1].ProductID),
				Query:             q,
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
			worker := mockwk.NewMockTaskDistributor(ctrl)
			ik := mockik.NewMockImageKitManagement(ctrl)
			mailSender := mockemail.NewMockEmailSender(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store, worker, ik, mailSender)
			//recorder := httptest.NewRecorder()

			url := "/api/v1/search-product-items-next-page"
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			// Add query parameters to request URL
			q := request.URL.Query()
			q.Add("limit", fmt.Sprintf("%d", tc.query.Limit))
			q.Add("product_item_cursor", fmt.Sprintf("%d", tc.query.ProductItemCursor))
			q.Add("product_cursor", fmt.Sprintf("%d", tc.query.ProductCursor))
			q.Add("query", tc.query.Query)
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
					Limit:            10,
					CategoryID:       null.IntFromPtr(nil),
					BrandID:          null.IntFromPtr(nil),
					ColorID:          null.IntFromPtr(nil),
					SizeID:           null.IntFromPtr(nil),
					IsNew:            null.BoolFromPtr(nil),
					IsPromoted:       null.BoolFromPtr(nil),
					IsFeatured:       null.BoolFromPtr(nil),
					IsQtyLimited:     null.BoolFromPtr(nil),
					OrderByLowPrice:  null.BoolFromPtr(nil),
					OrderByHighPrice: null.BoolFromPtr(nil),
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
			worker := mockwk.NewMockTaskDistributor(ctrl)
			ik := mockik.NewMockImageKitManagement(ctrl)
			mailSender := mockemail.NewMockEmailSender(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store, worker, ik, mailSender)
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

func TestListProductItemsWithPromotionsAPI(t *testing.T) {
	n := 10
	productItems := make([]db.ListProductItemsWithPromotionsRow, n)
	for i := 0; i < n; i++ {
		productItems[i] = randomListProductItemWithPromotions()
	}

	type Query struct {
		ProductID int64
		Limit     int
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
				ProductID: productItems[len(productItems)-1].ProductID,
				Limit:     n,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListProductItemsWithPromotionsParams{
					ProductID: productItems[len(productItems)-1].ProductID,
					Limit:     int32(n),
				}
				store.EXPECT().
					ListProductItemsWithPromotions(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(productItems, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchListProductItemsWithPromotions(t, rsp.Body, productItems)
			},
		},
		{
			name: "InternalError",
			query: Query{
				ProductID: productItems[len(productItems)-1].ProductID,
				Limit:     10,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListProductItemsWithPromotions(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.ListProductItemsWithPromotionsRow{}, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name: "InvalidLimit",
			query: Query{
				ProductID: productItems[len(productItems)-1].ProductID,
				Limit:     11,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListProductItemsWithPromotions(gomock.Any(), gomock.Any()).
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

			url := "/api/v1/product-items-with-promotions"
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			// Add query parameters to request URL
			q := request.URL.Query()
			q.Add("limit", fmt.Sprintf("%d", tc.query.Limit))
			q.Add("product_id", fmt.Sprintf("%d", tc.query.ProductID))
			request.URL.RawQuery = q.Encode()

			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
		})
	}
}

func TestListProductItemsWithPromotionsNextPageAPI(t *testing.T) {
	n := 10
	productItems1 := make([]db.ListProductItemsWithPromotionsNextPageRow, n)
	productItems2 := make([]db.ListProductItemsWithPromotionsNextPageRow, n)
	for i := 0; i < n; i++ {
		productItems1[i] = randomRestProductItemWithPromotions()
	}
	for i := 0; i < n; i++ {
		productItems2[i] = randomRestProductItemWithPromotions()
	}

	type Query struct {
		ProductItemCursor int
		ProductCursor     int
		Limit             int
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
				Limit:             n,
				ProductItemCursor: int(productItems1[len(productItems1)-1].ID),
				ProductCursor:     int(productItems1[len(productItems1)-1].ProductID),
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.ListProductItemsWithPromotionsNextPageParams{
					Limit:         int32(n),
					ProductItemID: productItems1[len(productItems1)-1].ID,
					ProductID:     productItems1[len(productItems1)-1].ProductID,
				}

				store.EXPECT().
					ListProductItemsWithPromotionsNextPage(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(productItems2, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchListProductItemsWithPromotionsNextPage(t, rsp.Body, productItems2)
			},
		},
		{
			name: "InternalError",
			query: Query{
				Limit:             10,
				ProductItemCursor: int(productItems1[len(productItems1)-1].ID),
				ProductCursor:     int(productItems1[len(productItems1)-1].ProductID),
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					ListProductItemsWithPromotionsNextPage(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.ListProductItemsWithPromotionsNextPageRow{}, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name: "InvalidLimit",
			query: Query{
				Limit:             11,
				ProductItemCursor: int(productItems1[len(productItems1)-1].ID),
				ProductCursor:     int(productItems1[len(productItems1)-1].ProductID),
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListProductItemsWithPromotionsNextPage(gomock.Any(), gomock.Any()).
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

			url := "/api/v1/product-items-with-promotions-next-page"
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			// Add query parameters to request URL
			q := request.URL.Query()
			q.Add("limit", fmt.Sprintf("%d", tc.query.Limit))
			q.Add("product_item_cursor", fmt.Sprintf("%d", tc.query.ProductItemCursor))
			q.Add("product_cursor", fmt.Sprintf("%d", tc.query.ProductCursor))
			request.URL.RawQuery = q.Encode()

			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
		})
	}
}

func TestListProductItemsWithBrandPromotionsAPI(t *testing.T) {
	n := 10
	productItems := make([]db.ListProductItemsWithBrandPromotionsRow, n)
	for i := 0; i < n; i++ {
		productItems[i] = randomListProductItemWithBrandPromotions()
	}

	type Query struct {
		BrandID int64
		Limit   int
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
				BrandID: productItems[len(productItems)-1].BrandID,
				Limit:   n,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListProductItemsWithBrandPromotionsParams{
					BrandID: productItems[len(productItems)-1].BrandID,
					Limit:   int32(n),
				}
				store.EXPECT().
					ListProductItemsWithBrandPromotions(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(productItems, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchListProductItemsWithBrandPromotions(t, rsp.Body, productItems)
			},
		},
		{
			name: "InternalError",
			query: Query{
				BrandID: productItems[len(productItems)-1].BrandID,
				Limit:   10,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListProductItemsWithBrandPromotions(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.ListProductItemsWithBrandPromotionsRow{}, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name: "InvalidLimit",
			query: Query{
				BrandID: productItems[len(productItems)-1].BrandID,
				Limit:   11,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListProductItemsWithBrandPromotions(gomock.Any(), gomock.Any()).
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

			url := "/api/v1/product-items-with-brand-promotions"
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			// Add query parameters to request URL
			q := request.URL.Query()
			q.Add("limit", fmt.Sprintf("%d", tc.query.Limit))
			q.Add("brand_id", fmt.Sprintf("%d", tc.query.BrandID))
			request.URL.RawQuery = q.Encode()

			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
		})
	}
}

func TestListProductItemsWithBrandPromotionsNextPageAPI(t *testing.T) {
	n := 10
	productItems1 := make([]db.ListProductItemsWithBrandPromotionsNextPageRow, n)
	productItems2 := make([]db.ListProductItemsWithBrandPromotionsNextPageRow, n)
	for i := 0; i < n; i++ {
		productItems1[i] = randomRestProductItemWithBrandPromotions()
	}
	for i := 0; i < n; i++ {
		productItems2[i] = randomRestProductItemWithBrandPromotions()
	}

	type Query struct {
		ProductItemCursor int
		ProductCursor     int
		Limit             int
		BrandID           int64
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
				Limit:             n,
				ProductItemCursor: int(productItems1[len(productItems1)-1].ID),
				ProductCursor:     int(productItems1[len(productItems1)-1].ProductID),
				BrandID:           productItems1[len(productItems1)-1].BrandID,
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.ListProductItemsWithBrandPromotionsNextPageParams{
					Limit:         int32(n),
					ProductItemID: productItems1[len(productItems1)-1].ID,
					ProductID:     productItems1[len(productItems1)-1].ProductID,
					BrandID:       productItems1[len(productItems1)-1].BrandID,
				}

				store.EXPECT().
					ListProductItemsWithBrandPromotionsNextPage(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(productItems2, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchListProductItemsWithBrandPromotionsNextPage(t, rsp.Body, productItems2)
			},
		},
		{
			name: "InternalError",
			query: Query{
				Limit:             10,
				ProductItemCursor: int(productItems1[len(productItems1)-1].ID),
				ProductCursor:     int(productItems1[len(productItems1)-1].ProductID),
				BrandID:           productItems1[len(productItems1)-1].BrandID,
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					ListProductItemsWithBrandPromotionsNextPage(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.ListProductItemsWithBrandPromotionsNextPageRow{}, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name: "InvalidLimit",
			query: Query{
				Limit:             11,
				ProductItemCursor: int(productItems1[len(productItems1)-1].ID),
				ProductCursor:     int(productItems1[len(productItems1)-1].ProductID),
				BrandID:           productItems1[len(productItems1)-1].BrandID,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListProductItemsWithBrandPromotionsNextPage(gomock.Any(), gomock.Any()).
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

			url := "/api/v1/product-items-with-brand-promotions-next-page"
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			// Add query parameters to request URL
			q := request.URL.Query()
			q.Add("limit", fmt.Sprintf("%d", tc.query.Limit))
			q.Add("brand_id", fmt.Sprintf("%d", tc.query.BrandID))
			q.Add("product_item_cursor", fmt.Sprintf("%d", tc.query.ProductItemCursor))
			q.Add("product_cursor", fmt.Sprintf("%d", tc.query.ProductCursor))
			request.URL.RawQuery = q.Encode()

			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
		})
	}
}

func TestListProductItemsWithCategoryPromotionsAPI(t *testing.T) {
	n := 10
	productItems := make([]db.ListProductItemsWithCategoryPromotionsRow, n)
	for i := 0; i < n; i++ {
		productItems[i] = randomListProductItemWithCategoryPromotions()
	}

	type Query struct {
		CategoryID int64
		Limit      int
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
				CategoryID: productItems[len(productItems)-1].CategoryID,
				Limit:      n,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListProductItemsWithCategoryPromotionsParams{
					CategoryID: productItems[len(productItems)-1].CategoryID,
					Limit:      int32(n),
				}
				store.EXPECT().
					ListProductItemsWithCategoryPromotions(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(productItems, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchListProductItemsWithCategoryPromotions(t, rsp.Body, productItems)
			},
		},
		{
			name: "InternalError",
			query: Query{
				CategoryID: productItems[len(productItems)-1].CategoryID,
				Limit:      10,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListProductItemsWithCategoryPromotions(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.ListProductItemsWithCategoryPromotionsRow{}, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name: "InvalidLimit",
			query: Query{
				CategoryID: productItems[len(productItems)-1].CategoryID,
				Limit:      11,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListProductItemsWithCategoryPromotions(gomock.Any(), gomock.Any()).
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

			url := "/api/v1/product-items-with-category-promotions"
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			// Add query parameters to request URL
			q := request.URL.Query()
			q.Add("limit", fmt.Sprintf("%d", tc.query.Limit))
			q.Add("category_id", fmt.Sprintf("%d", tc.query.CategoryID))
			request.URL.RawQuery = q.Encode()

			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
		})
	}
}

func TestListProductItemsWithCategoryPromotionsNextPageAPI(t *testing.T) {
	n := 10
	productItems1 := make([]db.ListProductItemsWithCategoryPromotionsNextPageRow, n)
	productItems2 := make([]db.ListProductItemsWithCategoryPromotionsNextPageRow, n)
	for i := 0; i < n; i++ {
		productItems1[i] = randomRestProductItemWithCategoryPromotions()
	}
	for i := 0; i < n; i++ {
		productItems2[i] = randomRestProductItemWithCategoryPromotions()
	}

	type Query struct {
		CategoryID        int64
		ProductItemCursor int
		ProductCursor     int
		Limit             int
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
				Limit:             n,
				ProductItemCursor: int(productItems1[len(productItems1)-1].ID),
				ProductCursor:     int(productItems1[len(productItems1)-1].ProductID),
				CategoryID:        productItems1[len(productItems1)-1].CategoryID,
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.ListProductItemsWithCategoryPromotionsNextPageParams{
					Limit:         int32(n),
					ProductItemID: productItems1[len(productItems1)-1].ID,
					ProductID:     productItems1[len(productItems1)-1].ProductID,
					CategoryID:    productItems1[len(productItems1)-1].CategoryID,
				}

				store.EXPECT().
					ListProductItemsWithCategoryPromotionsNextPage(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(productItems2, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchListProductItemsWithCategoryPromotionsNextPage(t, rsp.Body, productItems2)
			},
		},
		{
			name: "InternalError",
			query: Query{
				Limit:             10,
				ProductItemCursor: int(productItems1[len(productItems1)-1].ID),
				ProductCursor:     int(productItems1[len(productItems1)-1].ProductID),
				CategoryID:        productItems1[len(productItems1)-1].CategoryID,
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					ListProductItemsWithCategoryPromotionsNextPage(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.ListProductItemsWithCategoryPromotionsNextPageRow{}, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name: "InvalidLimit",
			query: Query{
				Limit:             11,
				ProductItemCursor: int(productItems1[len(productItems1)-1].ID),
				ProductCursor:     int(productItems1[len(productItems1)-1].ProductID),
				CategoryID:        productItems1[len(productItems1)-1].CategoryID,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListProductItemsWithCategoryPromotionsNextPage(gomock.Any(), gomock.Any()).
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

			url := "/api/v1/product-items-with-category-promotions-next-page"
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			// Add query parameters to request URL
			q := request.URL.Query()
			q.Add("limit", fmt.Sprintf("%d", tc.query.Limit))
			q.Add("category_id", fmt.Sprintf("%d", tc.query.CategoryID))
			q.Add("product_item_cursor", fmt.Sprintf("%d", tc.query.ProductItemCursor))
			q.Add("product_cursor", fmt.Sprintf("%d", tc.query.ProductCursor))
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
				arg := db.AdminUpdateProductItemParams{
					AdminID:    admin.ID,
					ProductID:  productItem.ProductID,
					ProductSku: null.IntFrom(productItem.ProductSku),
					QtyInStock: null.IntFrom(int64(productItem.QtyInStock)),
					// ProductImage: null.StringFrom(productItem.ProductImage),
					Price:  null.StringFrom("1000"),
					Active: null.BoolFromPtr(&productItem.Active),
					ID:     productItem.ID,
				}

				store.EXPECT().
					AdminUpdateProductItem(gomock.Any(), gomock.Eq(arg)).
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
				arg := db.AdminUpdateProductItemParams{
					AdminID:    admin.ID,
					ProductID:  productItem.ProductID,
					ProductSku: null.IntFrom(productItem.ProductSku),
					QtyInStock: null.IntFrom(int64(productItem.QtyInStock)),
					// ProductImage: null.StringFrom(productItem.ProductImage),
					Price:  null.StringFrom("1000"),
					Active: null.BoolFromPtr(&productItem.Active),
					ID:     productItem.ID,
				}

				store.EXPECT().
					AdminUpdateProductItem(gomock.Any(), gomock.Eq(arg)).
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
				arg := db.AdminUpdateProductItemParams{
					AdminID:    admin.ID,
					ProductID:  productItem.ProductID,
					ProductSku: null.IntFrom(productItem.ProductSku),
					QtyInStock: null.IntFrom(int64(productItem.QtyInStock)),
					// ProductImage: null.StringFrom(productItem.ProductImage),
					Price:  null.StringFrom("1000"),
					Active: null.BoolFromPtr(&productItem.Active),
					ID:     productItem.ID,
				}

				store.EXPECT().
					AdminUpdateProductItem(gomock.Any(), gomock.Eq(arg)).
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
				arg := db.AdminUpdateProductItemParams{
					AdminID:    admin.ID,
					ProductID:  productItem.ProductID,
					ProductSku: null.IntFrom(productItem.ProductSku),
					QtyInStock: null.IntFrom(int64(productItem.QtyInStock)),
					// ProductImage: null.StringFrom(productItem.ProductImage),
					Price:  null.StringFrom("1000"),
					Active: null.BoolFromPtr(&productItem.Active),
					ID:     productItem.ID,
				}
				store.EXPECT().
					AdminUpdateProductItem(gomock.Any(), gomock.Eq(arg)).
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
					AdminUpdateProductItem(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/admin/v1/admins/%d/product-items/%d", tc.AdminID, tc.productItemID)
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
			worker := mockwk.NewMockTaskDistributor(ctrl)
			ik := mockik.NewMockImageKitManagement(ctrl)
			mailSender := mockemail.NewMockEmailSender(ctrl)

			// build stubs
			tc.buildStub(store)

			// start test server and send request
			server := newTestServer(t, store, worker, ik, mailSender)
			//recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/admin/v1/admins/%d/product-items/%d", tc.AdminID, tc.productItemID)
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
		Name:          util.RandomUser(),
	}
}

func randomListProductItemWithBestSales() db.ListProductItemsWithBestSalesRow {
	return db.ListProductItemsWithBestSalesRow{
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
		Name:          util.RandomUser(),
		TotalSold:     null.IntFrom(util.RandomMoney()),
	}
}

func randomRestProductItemList() db.ListProductItemsNextPageRow {
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
		Name:          util.RandomUser(),
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
		Name:          "abc" + util.RandomUser(),
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
		Name:          "abc" + util.RandomUser(),
	}
}

func randomListProductItemWithPromotions() db.ListProductItemsWithPromotionsRow {
	return db.ListProductItemsWithPromotionsRow{
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
		Name:          util.RandomUser(),
	}
}

func randomRestProductItemWithPromotions() db.ListProductItemsWithPromotionsNextPageRow {
	return db.ListProductItemsWithPromotionsNextPageRow{
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
		Name:          util.RandomUser(),
	}
}

func randomListProductItemWithBrandPromotions() db.ListProductItemsWithBrandPromotionsRow {
	return db.ListProductItemsWithBrandPromotionsRow{
		BrandID:       util.RandomInt(1, 1000),
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
		Name:          util.RandomUser(),
	}
}

func randomRestProductItemWithBrandPromotions() db.ListProductItemsWithBrandPromotionsNextPageRow {
	return db.ListProductItemsWithBrandPromotionsNextPageRow{
		ID:            util.RandomInt(1, 1000),
		BrandID:       util.RandomInt(1, 1000),
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
		Name:          util.RandomUser(),
	}
}

func randomListProductItemWithCategoryPromotions() db.ListProductItemsWithCategoryPromotionsRow {
	return db.ListProductItemsWithCategoryPromotionsRow{
		ID:            util.RandomInt(1, 1000),
		CategoryID:    util.RandomInt(1, 1000),
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
		Name:          util.RandomUser(),
	}
}

func randomRestProductItemWithCategoryPromotions() db.ListProductItemsWithCategoryPromotionsNextPageRow {
	return db.ListProductItemsWithCategoryPromotionsNextPageRow{
		ID:            util.RandomInt(1, 1000),
		CategoryID:    util.RandomInt(1, 1000),
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
		Name:          util.RandomUser(),
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
func requireBodyMatchListProductItemsWithBestSales(t *testing.T, body io.ReadCloser, productItems []db.ListProductItemsWithBestSalesRow) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotProductItems []db.ListProductItemsWithBestSalesRow
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

func requireBodyMatchListProductItemsWithPromotions(t *testing.T, body io.ReadCloser, productItems []db.ListProductItemsWithPromotionsRow) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotProductItems []db.ListProductItemsWithPromotionsRow
	err = json.Unmarshal(data, &gotProductItems)
	require.NoError(t, err)
	require.Equal(t, productItems, gotProductItems)
}

func requireBodyMatchListProductItemsWithPromotionsNextPage(t *testing.T, body io.ReadCloser, productItems []db.ListProductItemsWithPromotionsNextPageRow) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotProductItems []db.ListProductItemsWithPromotionsNextPageRow
	err = json.Unmarshal(data, &gotProductItems)
	require.NoError(t, err)
	require.Equal(t, productItems, gotProductItems)
}

func requireBodyMatchListProductItemsWithBrandPromotions(t *testing.T, body io.ReadCloser, productItems []db.ListProductItemsWithBrandPromotionsRow) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotProductItems []db.ListProductItemsWithBrandPromotionsRow
	err = json.Unmarshal(data, &gotProductItems)
	require.NoError(t, err)
	require.Equal(t, productItems, gotProductItems)
}

func requireBodyMatchListProductItemsWithBrandPromotionsNextPage(t *testing.T, body io.ReadCloser, productItems []db.ListProductItemsWithBrandPromotionsNextPageRow) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotProductItems []db.ListProductItemsWithBrandPromotionsNextPageRow
	err = json.Unmarshal(data, &gotProductItems)
	require.NoError(t, err)
	require.Equal(t, productItems, gotProductItems)
}

func requireBodyMatchListProductItemsWithCategoryPromotions(t *testing.T, body io.ReadCloser, productItems []db.ListProductItemsWithCategoryPromotionsRow) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotProductItems []db.ListProductItemsWithCategoryPromotionsRow
	err = json.Unmarshal(data, &gotProductItems)
	require.NoError(t, err)
	require.Equal(t, productItems, gotProductItems)
}

func requireBodyMatchListProductItemsWithCategoryPromotionsNextPage(t *testing.T, body io.ReadCloser, productItems []db.ListProductItemsWithCategoryPromotionsNextPageRow) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotProductItems []db.ListProductItemsWithCategoryPromotionsNextPageRow
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
