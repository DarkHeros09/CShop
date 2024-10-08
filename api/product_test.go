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
	"github.com/cshop/v3/token"
	"github.com/cshop/v3/util"
	mockwk "github.com/cshop/v3/worker/mock"
	"github.com/gofiber/fiber/v2"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetProductAPI(t *testing.T) {
	product := randomProduct()

	testCases := []struct {
		name          string
		productID     int64
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rsp *http.Response)
	}{
		{
			name:      "OK",
			productID: product.ID,
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetProduct(gomock.Any(), gomock.Eq(product.ID)).
					Times(1).
					Return(product, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchProduct(t, rsp.Body, product)
			},
		},

		{
			name:      "NotFound",
			productID: product.ID,
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetProduct(gomock.Any(), gomock.Eq(product.ID)).
					Times(1).
					Return(db.Product{}, pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
			},
		},
		{
			name:      "InternalError",
			productID: product.ID,
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetProduct(gomock.Any(), gomock.Eq(product.ID)).
					Times(1).
					Return(db.Product{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:      "InvalidID",
			productID: 0,
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetProduct(gomock.Any(), gomock.Any()).
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

			// build stubs
			tc.buildStub(store)

			// start test server and send request
			server := newTestServer(t, store, worker, ik)
			//recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/api/v1/products/%d", tc.productID)
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(t, rsp)
		})

	}

}

func TestAdminCreateProductAPI(t *testing.T) {
	admin, _ := randomPromotionSuperAdmin(t)
	product := randomProduct()

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
				"name":        product.Name,
				"description": product.Description,
				"category_id": product.CategoryID,
				"brand_id":    product.BrandID,
				// "product_image": product.ProductImage,
				"active": product.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminCreateProductParams{
					AdminID:     admin.ID,
					CategoryID:  product.CategoryID,
					BrandID:     product.BrandID,
					Name:        product.Name,
					Description: product.Description,
					// ProductImage: product.ProductImage,
					Active: product.Active,
				}

				store.EXPECT().
					AdminCreateProduct(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(product, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchProduct(t, rsp.Body, product)
			},
		},
		{
			name:    "NoAuthorization",
			AdminID: admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminCreateProductParams{
					AdminID:     admin.ID,
					CategoryID:  product.CategoryID,
					BrandID:     product.BrandID,
					Name:        product.Name,
					Description: product.Description,
					// ProductImage: product.ProductImage,
					Active: product.Active,
				}

				store.EXPECT().
					AdminCreateProduct(gomock.Any(), gomock.Eq(arg)).
					Times(0)

			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:    "Unauthorized",
			AdminID: admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, false, time.Minute)
			},
			body: fiber.Map{
				"name":        product.Name,
				"description": product.Description,
				"category_id": product.CategoryID,
				"brand_id":    product.BrandID,
				// "product_image": product.ProductImage,
				"active": product.Active,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminCreateProductParams{
					AdminID:     admin.ID,
					CategoryID:  product.CategoryID,
					BrandID:     product.BrandID,
					Name:        product.Name,
					Description: product.Description,
					// ProductImage: product.ProductImage,
					Active: product.Active,
				}

				store.EXPECT().
					AdminCreateProduct(gomock.Any(), gomock.Eq(arg)).
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
				"name":        product.Name,
				"description": product.Description,
				"category_id": product.CategoryID,
				"brand_id":    product.BrandID,
				// "product_image": product.ProductImage,
				"active": product.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					AdminCreateProduct(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Product{}, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:    "InvalidID",
			AdminID: admin.ID,
			body: fiber.Map{
				"id": "invalid",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					AdminCreateProduct(gomock.Any(), gomock.Any()).
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
			tc.buildStubs(store)

			server := newTestServer(t, store, worker, ik)
			//recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/admin/v1/admins/%d/products", tc.AdminID)
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

func TestListProductsAPI(t *testing.T) {
	n := 5
	products := make([]db.ListProductsRow, n)
	for i := 0; i < n; i++ {
		products[i] = randomProductForList()
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
				arg := db.ListProductsParams{
					Limit:  int32(n),
					Offset: 0,
				}

				store.EXPECT().
					ListProducts(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(products, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchProductsList(t, rsp.Body, products)
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
					ListProducts(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.ListProductsRow{}, pgx.ErrTxClosed)
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
					ListProducts(gomock.Any(), gomock.Any()).
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
					ListProducts(gomock.Any(), gomock.Any()).
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
			tc.buildStubs(store)

			server := newTestServer(t, store, worker, ik)
			//recorder := httptest.NewRecorder()

			url := "/api/v1/products"
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

func TestUpdateProductAPI(t *testing.T) {
	admin, _ := randomPromotionSuperAdmin(t)
	product := randomProduct()

	testCases := []struct {
		name          string
		body          fiber.Map
		productID     int64
		AdminID       int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rsp *http.Response)
	}{
		{
			name:      "OK",
			productID: product.ID,
			AdminID:   admin.ID,
			body: fiber.Map{
				"name":          product.Name,
				"description":   "new product Description",
				"category_id":   product.CategoryID,
				"product_image": "https://newproduct.com/ProductImage",
				"active":        product.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminUpdateProductParams{
					AdminID:     admin.ID,
					CategoryID:  null.IntFromPtr(&product.CategoryID),
					Name:        null.StringFromPtr(&product.Name),
					Description: null.StringFrom("new product Description"),
					// ProductImage: null.StringFrom("https://newproduct.com/ProductImage"),
					Active: null.BoolFromPtr(&product.Active),
					ID:     product.ID,
				}

				store.EXPECT().
					AdminUpdateProduct(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(product, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:      "Unauthorized",
			productID: product.ID,
			AdminID:   admin.ID,
			body: fiber.Map{
				"name":          product.Name,
				"description":   "new product Description",
				"category_id":   product.CategoryID,
				"product_image": "https://newproduct.com/ProductImage",
				"active":        product.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, false, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminUpdateProductParams{
					AdminID:     admin.ID,
					CategoryID:  null.IntFromPtr(&product.CategoryID),
					Name:        null.StringFromPtr(&product.Name),
					Description: null.StringFrom("new product Description"),
					// ProductImage: null.StringFrom("https://newproduct.com/ProductImage"),
					Active: null.BoolFromPtr(&product.Active),
					ID:     product.ID,
				}

				store.EXPECT().
					AdminUpdateProduct(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:      "NoAuthorization",
			productID: product.ID,
			AdminID:   admin.ID,
			body: fiber.Map{
				"name":          product.Name,
				"description":   "new product Description",
				"category_id":   product.CategoryID,
				"product_image": "https://newproduct.com/ProductImage",
				"active":        product.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminUpdateProductParams{
					AdminID:     admin.ID,
					CategoryID:  null.IntFromPtr(&product.CategoryID),
					Name:        null.StringFromPtr(&product.Name),
					Description: null.StringFrom("new product Description"),
					// ProductImage: null.StringFrom("https://newproduct.com/ProductImage"),
					Active: null.BoolFromPtr(&product.Active),
					ID:     product.ID,
				}

				store.EXPECT().
					AdminUpdateProduct(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:      "InternalError",
			productID: product.ID,
			AdminID:   admin.ID,
			body: fiber.Map{
				"name":          product.Name,
				"description":   "new product Description",
				"category_id":   product.CategoryID,
				"product_image": "https://newproduct.com/ProductImage",
				"active":        product.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminUpdateProductParams{
					AdminID:     admin.ID,
					CategoryID:  null.IntFromPtr(&product.CategoryID),
					Name:        null.StringFromPtr(&product.Name),
					Description: null.StringFrom("new product Description"),
					// ProductImage: null.StringFrom("https://newproduct.com/ProductImage"),
					Active: null.BoolFromPtr(&product.Active),
					ID:     product.ID,
				}
				store.EXPECT().
					AdminUpdateProduct(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.Product{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:      "InvalidID",
			productID: 0,
			AdminID:   admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					AdminUpdateProduct(gomock.Any(), gomock.Any()).
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
			tc.buildStubs(store)

			server := newTestServer(t, store, worker, ik)
			//recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/admin/v1/admins/%d/products/%d", tc.AdminID, tc.productID)
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

func TestAdminDeleteProductAPI(t *testing.T) {
	admin, _ := randomPromotionSuperAdmin(t)
	product := randomProduct()

	testCases := []struct {
		name          string
		productID     int64
		AdminID       int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rsp *http.Response)
	}{
		{
			name:      "OK",
			productID: product.ID,
			AdminID:   admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {

				arg := db.AdminDeleteProductParams{
					AdminID: admin.ID,
					ID:      product.ID,
				}
				store.EXPECT().
					AdminDeleteProduct(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:      "Unauthorized",
			productID: product.ID,
			AdminID:   admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, 2, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.AdminDeleteProductParams{
					AdminID: admin.ID,
					ID:      product.ID,
				}
				store.EXPECT().
					AdminDeleteProduct(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:      "NoAuthorization",
			productID: product.ID,
			AdminID:   admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.AdminDeleteProductParams{
					AdminID: admin.ID,
					ID:      product.ID,
				}
				store.EXPECT().
					AdminDeleteProduct(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:      "NotFound",
			productID: product.ID,
			AdminID:   admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.AdminDeleteProductParams{
					AdminID: admin.ID,
					ID:      product.ID,
				}
				store.EXPECT().
					AdminDeleteProduct(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
			},
		},
		{
			name:      "InternalError",
			productID: product.ID,
			AdminID:   admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.AdminDeleteProductParams{
					AdminID: admin.ID,
					ID:      product.ID,
				}
				store.EXPECT().
					AdminDeleteProduct(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:      "InvalidID",
			productID: 0,
			AdminID:   admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					AdminDeleteProduct(gomock.Any(), gomock.Any()).
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

			// build stubs
			tc.buildStub(store)

			// start test server and send request
			server := newTestServer(t, store, worker, ik)
			//recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/admin/v1/admins/%d/products/%d", tc.AdminID, tc.productID)
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

func TestListProductsV2API(t *testing.T) {
	n := 10
	products := make([]db.ListProductsV2Row, n)
	for i := 0; i < n; i++ {
		products[i] = randomListProductV2()
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
				store.EXPECT().
					ListProductsV2(gomock.Any(), gomock.Eq(int32(10))).
					Times(1).
					Return(products, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchListProductsV2(t, rsp.Body, products)
			},
		},
		{
			name: "InternalError",
			query: Query{
				Limit: 10,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListProductsV2(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.ListProductsV2Row{}, pgx.ErrTxClosed)
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
					ListProductsV2(gomock.Any(), gomock.Any()).
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
			tc.buildStubs(store)

			server := newTestServer(t, store, worker, ik)
			//recorder := httptest.NewRecorder()

			url := "/api/v1/products-v2"
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

func TestListProductsNextPageAPI(t *testing.T) {
	n := 10
	products1 := make([]db.ListProductsNextPageRow, n)
	products2 := make([]db.ListProductsNextPageRow, n)
	for i := 0; i < n; i++ {
		products1[i] = randomRestProductList()
	}
	for i := 0; i < n; i++ {
		products2[i] = randomRestProductList()
	}

	type Query struct {
		ProductCursor int
		Limit         int
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
				Limit:         n,
				ProductCursor: int(products1[len(products1)-1].ID),
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.ListProductsNextPageParams{
					Limit: int32(n),
					ID:    products1[len(products1)-1].ID,
				}

				store.EXPECT().
					ListProductsNextPage(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(products2, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchListProductsNextPage(t, rsp.Body, products2)
			},
		},
		{
			name: "InternalError",
			query: Query{
				Limit:         10,
				ProductCursor: int(products1[len(products1)-1].ID),
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					ListProductsNextPage(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.ListProductsNextPageRow{}, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name: "InvalidLimit",
			query: Query{
				Limit:         11,
				ProductCursor: int(products1[len(products1)-1].ID),
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListProductsNextPage(gomock.Any(), gomock.Any()).
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
			tc.buildStubs(store)

			server := newTestServer(t, store, worker, ik)
			//recorder := httptest.NewRecorder()

			url := "/api/v1/products-next-page"
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			// Add query parameters to request URL
			q := request.URL.Query()
			q.Add("limit", fmt.Sprintf("%d", tc.query.Limit))
			q.Add("product_cursor", fmt.Sprintf("%d", tc.query.ProductCursor))
			request.URL.RawQuery = q.Encode()

			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
		})
	}
}

func TestSearchProductsAPI(t *testing.T) {
	n := 10
	q := "abc"
	products := make([]db.SearchProductsRow, n)
	for i := 0; i < n; i++ {
		products[i] = randomSearchProducts()
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
				arg := db.SearchProductsParams{
					Limit: int32(n),
					Query: q,
				}
				store.EXPECT().
					SearchProducts(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(products, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchSearchProducts(t, rsp.Body, products)
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
					SearchProducts(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.SearchProductsRow{}, pgx.ErrTxClosed)
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
					SearchProducts(gomock.Any(), gomock.Any()).
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
			tc.buildStubs(store)

			server := newTestServer(t, store, worker, ik)
			//recorder := httptest.NewRecorder()

			url := "/api/v1/search-products"
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

func TestSearchProductsNextPageAPI(t *testing.T) {
	n := 10
	q := "abc"
	products1 := make([]db.SearchProductsNextPageRow, n)
	products2 := make([]db.SearchProductsNextPageRow, n)
	for i := 0; i < n; i++ {
		products1[i] = randomSearchProductNextPage()
	}
	for i := 0; i < n; i++ {
		products2[i] = randomSearchProductNextPage()
	}

	type Query struct {
		ProductCursor int
		Limit         int
		Query         string
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
				Limit:         n,
				Query:         q,
				ProductCursor: int(products1[len(products1)-1].ID),
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.SearchProductsNextPageParams{
					Limit:     int32(n),
					ProductID: products1[len(products1)-1].ID,
					Query:     q,
				}

				store.EXPECT().
					SearchProductsNextPage(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(products2, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchSearchProductsNextPage(t, rsp.Body, products2)
			},
		},
		{
			name: "InternalError",
			query: Query{
				Limit:         10,
				ProductCursor: int(products1[len(products1)-1].ID),
				Query:         q,
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					SearchProductsNextPage(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.SearchProductsNextPageRow{}, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name: "InvalidLimit",
			query: Query{
				Limit:         11,
				ProductCursor: int(products1[len(products1)-1].ID),
				Query:         q,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					SearchProductsNextPage(gomock.Any(), gomock.Any()).
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
			tc.buildStubs(store)

			server := newTestServer(t, store, worker, ik)
			//recorder := httptest.NewRecorder()

			url := "/api/v1/search-products-next-page"
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			// Add query parameters to request URL
			q := request.URL.Query()
			q.Add("limit", fmt.Sprintf("%d", tc.query.Limit))
			q.Add("product_item_cursor", fmt.Sprintf("%d", tc.query.ProductCursor))
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

func randomPromotionSuperAdmin(t *testing.T) (admin db.Admin, password string) {
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

func randomProduct() db.Product {
	return db.Product{
		ID:          util.RandomInt(1, 1000),
		CategoryID:  util.RandomInt(1, 1000),
		BrandID:     util.RandomInt(1, 1000),
		Name:        util.RandomUser(),
		Description: util.RandomUser(),
		// ProductImage: util.RandomURL(),
		Active: util.RandomBool(),
	}
}
func randomProductForList() db.ListProductsRow {
	return db.ListProductsRow{
		ID:          util.RandomInt(1, 1000),
		CategoryID:  util.RandomInt(1, 500),
		Name:        util.RandomUser(),
		Description: util.RandomUser(),
		// ProductImage: util.RandomURL(),
		Active:     util.RandomBool(),
		TotalCount: util.RandomMoney(),
	}
}

func requireBodyMatchProduct(t *testing.T, body io.ReadCloser, product db.Product) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotProduct db.Product
	err = json.Unmarshal(data, &gotProduct)
	require.NoError(t, err)
	require.Equal(t, product, gotProduct)
}

func requireBodyMatchProductsList(t *testing.T, body io.ReadCloser, products []db.ListProductsRow) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotProducts []db.ListProductsRow
	err = json.Unmarshal(data, &gotProducts)
	require.NoError(t, err)
	require.Equal(t, products, gotProducts)
}

func randomSearchProducts() db.SearchProductsRow {
	return db.SearchProductsRow{
		ID:          util.RandomInt(1, 1000),
		Name:        "abc" + util.RandomUser(),
		Description: util.RandomUser(),
		CategoryID:  util.RandomInt(1, 1000),
		BrandID:     util.RandomInt(1, 1000),
		Active:      util.RandomBool(),
	}
}

func randomSearchProductNextPage() db.SearchProductsNextPageRow {
	return db.SearchProductsNextPageRow{
		ID:          util.RandomInt(1, 1000),
		Name:        "abc" + util.RandomUser(),
		Description: util.RandomUser(),
		CategoryID:  util.RandomInt(1, 1000),
		BrandID:     util.RandomInt(1, 1000),
		Active:      util.RandomBool(),
	}
}

func requireBodyMatchSearchProducts(t *testing.T, body io.ReadCloser, products []db.SearchProductsRow) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotProducts []db.SearchProductsRow
	err = json.Unmarshal(data, &gotProducts)
	require.NoError(t, err)
	require.Equal(t, products, gotProducts)
}

func requireBodyMatchSearchProductsNextPage(t *testing.T, body io.ReadCloser, products []db.SearchProductsNextPageRow) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotProducts []db.SearchProductsNextPageRow
	err = json.Unmarshal(data, &gotProducts)
	require.NoError(t, err)
	require.Equal(t, products, gotProducts)
}

func requireBodyMatchListProductsV2(t *testing.T, body io.ReadCloser, products []db.ListProductsV2Row) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotProducts []db.ListProductsV2Row
	err = json.Unmarshal(data, &gotProducts)
	require.NoError(t, err)
	require.Equal(t, products, gotProducts)
}

func requireBodyMatchListProductsNextPage(t *testing.T, body io.ReadCloser, products []db.ListProductsNextPageRow) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotProducts []db.ListProductsNextPageRow
	err = json.Unmarshal(data, &gotProducts)
	require.NoError(t, err)
	require.Equal(t, products, gotProducts)
}

func randomListProductV2() db.ListProductsV2Row {
	return db.ListProductsV2Row{
		ID:          util.RandomInt(1, 1000),
		CategoryID:  util.RandomInt(1, 500),
		BrandID:     util.RandomInt(1, 500),
		Name:        util.RandomUser(),
		Description: util.RandomUser(),
		// ProductImage: util.RandomURL(),
		Active: util.RandomBool(),
	}
}

func randomRestProductList() db.ListProductsNextPageRow {
	return db.ListProductsNextPageRow{
		ID:          util.RandomInt(1, 1000),
		CategoryID:  util.RandomInt(1, 500),
		BrandID:     util.RandomInt(1, 500),
		Name:        util.RandomUser(),
		Description: util.RandomUser(),
		// ProductImage: util.RandomURL(),
		Active: util.RandomBool(),
	}
}
