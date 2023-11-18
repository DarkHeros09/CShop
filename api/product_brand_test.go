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
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetProductBrandAPI(t *testing.T) {
	productBrand := randomProductBrand()

	testCases := []struct {
		name           string
		productBrandID int64
		buildStub      func(store *mockdb.MockStore)
		checkResponse  func(t *testing.T, rsp *http.Response)
	}{
		{
			name:           "OK",
			productBrandID: productBrand.ID,
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetProductBrand(gomock.Any(), gomock.Eq(productBrand.ID)).
					Times(1).
					Return(productBrand, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchProductBrand(t, rsp.Body, productBrand)
			},
		},

		{
			name:           "NotFound",
			productBrandID: productBrand.ID,
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetProductBrand(gomock.Any(), gomock.Eq(productBrand.ID)).
					Times(1).
					Return(db.ProductBrand{}, pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
			},
		},
		{
			name:           "InternalError",
			productBrandID: productBrand.ID,
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetProductBrand(gomock.Any(), gomock.Eq(productBrand.ID)).
					Times(1).
					Return(db.ProductBrand{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:           "InvalidID",
			productBrandID: 0,
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetProductBrand(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/api/v1/brands/%d", tc.productBrandID)
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(t, rsp)
		})

	}

}

func TestCreateProductBrandAPI(t *testing.T) {
	admin, _ := randomPBrandSuperAdmin(t)
	productBrand := randomProductBrand()

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
				"brand_name":  productBrand.BrandName,
				"brand_image": productBrand.BrandImage,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.CreateProductBrandParams{
					BrandName:  productBrand.BrandName,
					BrandImage: productBrand.BrandImage,
				}

				store.EXPECT().
					CreateProductBrand(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(productBrand, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchProductBrand(t, rsp.Body, productBrand)
			},
		},
		{
			name:    "NoAuthorization",
			AdminID: admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateProductBrandParams{
					BrandName:  productBrand.BrandName,
					BrandImage: productBrand.BrandImage,
				}

				store.EXPECT().
					CreateProductBrand(gomock.Any(), gomock.Eq(arg)).
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
				"brand_name":  productBrand.BrandName,
				"brand_image": productBrand.BrandImage,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateProductBrandParams{
					BrandName:  productBrand.BrandName,
					BrandImage: productBrand.BrandImage,
				}
				store.EXPECT().
					CreateProductBrand(gomock.Any(), gomock.Eq(arg)).
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
				"brand_name":  productBrand.BrandName,
				"brand_image": productBrand.BrandImage,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateProductBrand(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.ProductBrand{}, pgx.ErrTxClosed)
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
					CreateProductBrand(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/admin/%d/v1/brands", tc.AdminID)
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

func TestListProductBrandsAPI(t *testing.T) {
	n := 5
	productBrands := make([]db.ProductBrand, n)
	for i := 0; i < n; i++ {
		productBrands[i] = randomProductBrand()
	}

	type Query struct {
		pageID   int
		pageSize int
	}

	testCases := []struct {
		name string
		// query         Query
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(rsp *http.Response)
	}{
		{
			name: "OK",
			// query: Query{
			// 	pageID:   1,
			// 	pageSize: n,
			// },
			buildStubs: func(store *mockdb.MockStore) {
				// arg := db.ListProductBrandsParams{
				// 	Limit:  int32(n),
				// 	Offset: 0,
				// }

				store.EXPECT().
					ListProductBrands(gomock.Any()).
					Times(1).
					Return(productBrands, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchProductBrands(t, rsp.Body, productBrands)
			},
		},
		{
			name: "InternalError",
			// query: Query{
			// 	pageID:   1,
			// 	pageSize: n,
			// },
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListProductBrands(gomock.Any()).
					Times(1).
					Return([]db.ProductBrand{}, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		// {
		// 	name: "InvalidPageID",
		// 	query: Query{
		// 		pageID:   -1,
		// 		pageSize: n,
		// 	},
		// 	buildStubs: func(store *mockdb.MockStore) {
		// 		store.EXPECT().
		// 			ListProductBrands(gomock.Any(), gomock.Any()).
		// 			Times(0)
		// 	},
		// 	checkResponse: func(rsp *http.Response) {
		// 		require.Equal(t, http.StatusBadRequest, rsp.StatusCode)
		// 	},
		// },
		// {
		// 	name: "InvalidPageSize",
		// 	query: Query{
		// 		pageID:   1,
		// 		pageSize: 100000,
		// 	},
		// 	buildStubs: func(store *mockdb.MockStore) {
		// 		store.EXPECT().
		// 			ListProductBrands(gomock.Any(), gomock.Any()).
		// 			Times(0)
		// 	},
		// 	checkResponse: func(rsp *http.Response) {
		// 		require.Equal(t, http.StatusBadRequest, rsp.StatusCode)
		// 	},
		// },
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			//recorder := httptest.NewRecorder()

			url := "/api/v1/brands"
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			// Add query parameters to request URL
			// q := request.URL.Query()
			// q.Add("page_id", fmt.Sprintf("%d", tc.query.pageID))
			// q.Add("page_size", fmt.Sprintf("%d", tc.query.pageSize))
			// request.URL.RawQuery = q.Encode()

			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
		})
	}
}

func TestUpdateProductBrandAPI(t *testing.T) {
	admin, _ := randomPBrandSuperAdmin(t)
	productBrand := randomProductBrand()

	testCases := []struct {
		name           string
		body           fiber.Map
		productBrandID int64
		AdminID        int64
		setupAuth      func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs     func(store *mockdb.MockStore)
		checkResponse  func(t *testing.T, rsp *http.Response)
	}{
		{
			name:           "OK",
			AdminID:        admin.ID,
			productBrandID: productBrand.ID,
			body: fiber.Map{
				"brand_name": "new name",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateProductBrandParams{
					BrandName: "new name",
					ID:        productBrand.ID,
				}

				store.EXPECT().
					UpdateProductBrand(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(productBrand, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:           "Unauthorized",
			AdminID:        admin.ID,
			productBrandID: productBrand.ID,
			body: fiber.Map{
				"brand_name": "new name",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, false, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateProductBrandParams{
					BrandName: "new name",
					ID:        productBrand.ID,
				}

				store.EXPECT().
					UpdateProductBrand(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:           "NoAuthorization",
			productBrandID: productBrand.ID,
			AdminID:        admin.ID,
			body: fiber.Map{
				"brand_name": "new name",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateProductBrandParams{
					BrandName: "new name",
					ID:        productBrand.ID,
				}

				store.EXPECT().
					UpdateProductBrand(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:           "InternalError",
			productBrandID: productBrand.ID,
			AdminID:        admin.ID,
			body: fiber.Map{
				"brand_name": "new name",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateProductBrandParams{
					BrandName: "new name",
					ID:        productBrand.ID,
				}
				store.EXPECT().
					UpdateProductBrand(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.ProductBrand{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:           "InvalidID",
			AdminID:        admin.ID,
			productBrandID: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateProductBrand(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/admin/%d/v1/brands/%d", tc.AdminID, tc.productBrandID)
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

func TestDeleteProductBrandAPI(t *testing.T) {
	admin, _ := randomPBrandSuperAdmin(t)
	productBrand := randomProductBrand()

	testCases := []struct {
		name           string
		body           fiber.Map
		productBrandID int64
		AdminID        int64
		setupAuth      func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub      func(store *mockdb.MockStore)
		checkResponse  func(t *testing.T, rsp *http.Response)
	}{
		{
			name:           "OK",
			productBrandID: productBrand.ID,
			AdminID:        admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {

				store.EXPECT().
					DeleteProductBrand(gomock.Any(), gomock.Eq(productBrand.ID)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:           "Unauthorized",
			productBrandID: productBrand.ID,
			AdminID:        admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, 2, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {

				store.EXPECT().
					DeleteProductBrand(gomock.Any(), gomock.Eq(productBrand.ID)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:           "NoAuthorization",
			productBrandID: productBrand.ID,
			AdminID:        admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStub: func(store *mockdb.MockStore) {

				store.EXPECT().
					DeleteProductBrand(gomock.Any(), gomock.Eq(productBrand.ID)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:           "NotFound",
			productBrandID: productBrand.ID,
			AdminID:        admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {

				store.EXPECT().
					DeleteProductBrand(gomock.Any(), gomock.Eq(productBrand.ID)).
					Times(1).
					Return(pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
			},
		},
		{
			name:           "InternalError",
			productBrandID: productBrand.ID,
			AdminID:        admin.ID,
			body:           fiber.Map{},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {

				store.EXPECT().
					DeleteProductBrand(gomock.Any(), gomock.Eq(productBrand.ID)).
					Times(1).
					Return(pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:           "InvalidID",
			productBrandID: 0,
			AdminID:        admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteProductBrand(gomock.Any(), gomock.Any()).
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

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/admin/%d/v1/brands/%d", tc.AdminID, tc.productBrandID)
			request, err := http.NewRequest(fiber.MethodDelete, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(t, rsp)
		})

	}

}

func randomPBrandSuperAdmin(t *testing.T) (admin db.Admin, password string) {
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

func randomProductBrand() db.ProductBrand {
	return db.ProductBrand{
		ID:         util.RandomInt(1, 1000),
		BrandName:  util.RandomUser(),
		BrandImage: util.RandomURL(),
	}
}

func requireBodyMatchProductBrand(t *testing.T, body io.ReadCloser, productBrand db.ProductBrand) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotProductBrand db.ProductBrand
	err = json.Unmarshal(data, &gotProductBrand)
	require.NoError(t, err)
	require.Equal(t, productBrand, gotProductBrand)
}

func requireBodyMatchProductBrands(t *testing.T, body io.ReadCloser, productBrands []db.ProductBrand) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotProductBrands []db.ProductBrand
	err = json.Unmarshal(data, &gotProductBrands)
	require.NoError(t, err)
	require.Equal(t, productBrands, gotProductBrands)
}
