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
	mockwk "github.com/cshop/v3/worker/mock"
	"github.com/gofiber/fiber/v2"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetProductConfigurationAPI(t *testing.T) {
	productConfiguration := randomProductConfiguration()

	testCases := []struct {
		name              string
		ProductItemID     int64
		VariationOptionID int64
		buildStub         func(store *mockdb.MockStore)
		checkResponse     func(t *testing.T, rsp *http.Response)
	}{
		{
			name:              "OK",
			ProductItemID:     productConfiguration.ProductItemID,
			VariationOptionID: productConfiguration.VariationOptionID,
			buildStub: func(store *mockdb.MockStore) {
				arg := db.GetProductConfigurationParams{
					ProductItemID:     productConfiguration.ProductItemID,
					VariationOptionID: productConfiguration.VariationOptionID,
				}
				store.EXPECT().
					GetProductConfiguration(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(productConfiguration, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchProductConfiguration(t, rsp.Body, productConfiguration)
			},
		},

		{
			name:              "NotFound",
			ProductItemID:     productConfiguration.ProductItemID,
			VariationOptionID: productConfiguration.VariationOptionID,
			buildStub: func(store *mockdb.MockStore) {
				arg := db.GetProductConfigurationParams{
					ProductItemID:     productConfiguration.ProductItemID,
					VariationOptionID: productConfiguration.VariationOptionID,
				}
				store.EXPECT().
					GetProductConfiguration(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.ProductConfiguration{}, pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
			},
		},
		{
			name:              "InternalError",
			ProductItemID:     productConfiguration.ProductItemID,
			VariationOptionID: productConfiguration.VariationOptionID,
			buildStub: func(store *mockdb.MockStore) {
				arg := db.GetProductConfigurationParams{
					ProductItemID:     productConfiguration.ProductItemID,
					VariationOptionID: productConfiguration.VariationOptionID,
				}
				store.EXPECT().
					GetProductConfiguration(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.ProductConfiguration{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:              "InvalidID",
			ProductItemID:     0,
			VariationOptionID: productConfiguration.VariationOptionID,
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetProductConfiguration(gomock.Any(), gomock.Any()).
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

			// build stubs
			tc.buildStub(store)

			// start test server and send request
			server := newTestServer(t, store, worker)
			//recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/api/v1/product-configurations/%d/variation-options/%d", tc.ProductItemID, tc.VariationOptionID)
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(t, rsp)
		})

	}

}

func TestCreateProductConfigurationAPI(t *testing.T) {
	admin, _ := randomProductConfigurationSuperAdmin(t)
	productConfiguration := randomProductConfiguration()

	testCases := []struct {
		name          string
		body          fiber.Map
		AdminID       int64
		ProductItemID int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(rsp *http.Response)
	}{
		{
			name:          "OK",
			AdminID:       admin.ID,
			ProductItemID: productConfiguration.ProductItemID,
			body: fiber.Map{
				"variation_id": productConfiguration.VariationOptionID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateProductConfigurationParams{
					ProductItemID:     productConfiguration.ProductItemID,
					VariationOptionID: productConfiguration.VariationOptionID,
				}

				store.EXPECT().
					CreateProductConfiguration(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(productConfiguration, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchProductConfiguration(t, rsp.Body, productConfiguration)
			},
		},
		{
			name:          "NoAuthorization",
			AdminID:       admin.ID,
			ProductItemID: productConfiguration.ProductItemID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateProductConfigurationParams{
					ProductItemID:     productConfiguration.ProductItemID,
					VariationOptionID: productConfiguration.VariationOptionID,
				}

				store.EXPECT().
					CreateProductConfiguration(gomock.Any(), gomock.Eq(arg)).
					Times(0)

			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:          "Unauthorized",
			AdminID:       admin.ID,
			ProductItemID: productConfiguration.ProductItemID,
			body: fiber.Map{
				"variation_id": productConfiguration.VariationOptionID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, false, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateProductConfigurationParams{
					ProductItemID:     productConfiguration.ProductItemID,
					VariationOptionID: productConfiguration.VariationOptionID,
				}

				store.EXPECT().
					CreateProductConfiguration(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:          "InternalError",
			AdminID:       admin.ID,
			ProductItemID: productConfiguration.ProductItemID,
			body: fiber.Map{
				"variation_id": productConfiguration.VariationOptionID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateProductConfiguration(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.ProductConfiguration{}, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:          "InvalidID",
			AdminID:       admin.ID,
			ProductItemID: 0,
			body: fiber.Map{
				"variation_id": productConfiguration.VariationOptionID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateProductConfiguration(gomock.Any(), gomock.Any()).
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
			tc.buildStubs(store)

			server := newTestServer(t, store, worker)
			//recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/admin/v1/admins/%d/product-configurations/%d", tc.AdminID, tc.ProductItemID)
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

func TestListProductConfigurationsAPI(t *testing.T) {
	n := 5
	type Query struct {
		pageID   int
		pageSize int
	}
	productConfigurations := make([]db.ProductConfiguration, n)
	for i := 0; i < n; i++ {
		productConfigurations[i] = randomProductConfiguration()
		testCases := []struct {
			name          string
			query         Query
			ProductItemID int64
			buildStubs    func(store *mockdb.MockStore)
			checkResponse func(rsp *http.Response)
		}{
			{
				name:          "OK",
				ProductItemID: productConfigurations[i].ProductItemID,
				query: Query{
					pageID:   1,
					pageSize: n,
				},
				buildStubs: func(store *mockdb.MockStore) {
					arg := db.ListProductConfigurationsParams{
						Limit:  int32(n),
						Offset: 0,
					}

					store.EXPECT().
						ListProductConfigurations(gomock.Any(), gomock.Eq(arg)).
						Times(1).
						Return(productConfigurations, nil)
				},
				checkResponse: func(rsp *http.Response) {
					require.Equal(t, http.StatusOK, rsp.StatusCode)
					requireBodyMatchProductConfigurations(t, rsp.Body, productConfigurations)
				},
			},
			{
				name:          "InternalError",
				ProductItemID: productConfigurations[i].ProductItemID,
				query: Query{
					pageID:   1,
					pageSize: n,
				},
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						ListProductConfigurations(gomock.Any(), gomock.Any()).
						Times(1).
						Return([]db.ProductConfiguration{}, pgx.ErrTxClosed)
				},
				checkResponse: func(rsp *http.Response) {
					require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
				},
			},
			{
				name:          "InvalidPageID",
				ProductItemID: productConfigurations[i].ProductItemID,
				query: Query{
					pageID:   -1,
					pageSize: n,
				},
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						ListProductConfigurations(gomock.Any(), gomock.Any()).
						Times(0)
				},
				checkResponse: func(rsp *http.Response) {
					require.Equal(t, http.StatusBadRequest, rsp.StatusCode)
				},
			},
			{
				name:          "InvalidPageSize",
				ProductItemID: productConfigurations[i].ProductItemID,
				query: Query{
					pageID:   1,
					pageSize: 100000,
				},
				buildStubs: func(store *mockdb.MockStore) {
					store.EXPECT().
						ListProductConfigurations(gomock.Any(), gomock.Any()).
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
				tc.buildStubs(store)

				server := newTestServer(t, store, worker)
				//recorder := httptest.NewRecorder()

				url := fmt.Sprintf("/api/v1/product-configurations/%d", tc.ProductItemID)
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

}

func TestUpdateProductConfigurationAPI(t *testing.T) {
	admin, _ := randomProductConfigurationSuperAdmin(t)
	productConfiguration := randomProductConfiguration()
	updatedOption := util.RandomMoney()

	testCases := []struct {
		name          string
		body          fiber.Map
		ProductItemID int64
		AdminID       int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rsp *http.Response)
	}{
		{
			name:          "OK",
			ProductItemID: productConfiguration.ProductItemID,
			AdminID:       admin.ID,
			body: fiber.Map{
				"variation_id": updatedOption,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateProductConfigurationParams{
					ProductItemID:     productConfiguration.ProductItemID,
					VariationOptionID: null.IntFrom(updatedOption),
				}

				store.EXPECT().
					UpdateProductConfiguration(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(productConfiguration, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:          "Unauthorized",
			ProductItemID: productConfiguration.ProductItemID,
			AdminID:       admin.ID,
			body: fiber.Map{
				"variation_id": updatedOption,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, false, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateProductConfigurationParams{
					ProductItemID:     productConfiguration.ProductItemID,
					VariationOptionID: null.IntFrom(updatedOption),
				}

				store.EXPECT().
					UpdateProductConfiguration(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:          "NoAuthorization",
			ProductItemID: productConfiguration.ProductItemID,
			AdminID:       admin.ID,
			body: fiber.Map{
				"variation_id": updatedOption,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateProductConfigurationParams{
					ProductItemID:     productConfiguration.ProductItemID,
					VariationOptionID: null.IntFrom(updatedOption),
				}

				store.EXPECT().
					UpdateProductConfiguration(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:          "InternalError",
			ProductItemID: productConfiguration.ProductItemID,
			AdminID:       admin.ID,
			body: fiber.Map{
				"variation_id": updatedOption,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateProductConfigurationParams{
					ProductItemID:     productConfiguration.ProductItemID,
					VariationOptionID: null.IntFrom(updatedOption),
				}
				store.EXPECT().
					UpdateProductConfiguration(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.ProductConfiguration{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:          "InvalidID",
			ProductItemID: 0,
			AdminID:       admin.ID,
			body: fiber.Map{
				"variation_id": updatedOption,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateProductConfiguration(gomock.Any(), gomock.Any()).
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
			tc.buildStubs(store)

			server := newTestServer(t, store, worker)
			//recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/admin/v1/admins/%d/product-configurations/%d", tc.AdminID, tc.ProductItemID)
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

func TestDeleteProductConfigurationAPI(t *testing.T) {
	admin, _ := randomProductConfigurationSuperAdmin(t)
	productConfiguration := randomProductConfiguration()

	testCases := []struct {
		name              string
		AdminID           int64
		ProductItemID     int64
		VariationOptionID int64
		setupAuth         func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub         func(store *mockdb.MockStore)
		checkResponse     func(t *testing.T, rsp *http.Response)
	}{
		{
			name:              "OK",
			AdminID:           admin.ID,
			ProductItemID:     productConfiguration.ProductItemID,
			VariationOptionID: productConfiguration.VariationOptionID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteProductConfigurationParams{
					ProductItemID:     productConfiguration.ProductItemID,
					VariationOptionID: productConfiguration.VariationOptionID,
				}

				store.EXPECT().
					DeleteProductConfiguration(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:              "Unauthorized",
			AdminID:           admin.ID,
			ProductItemID:     productConfiguration.ProductItemID,
			VariationOptionID: productConfiguration.VariationOptionID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, 2, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteProductConfigurationParams{
					ProductItemID:     productConfiguration.ProductItemID,
					VariationOptionID: productConfiguration.VariationOptionID,
				}

				store.EXPECT().
					DeleteProductConfiguration(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:              "NoAuthorization",
			AdminID:           admin.ID,
			ProductItemID:     productConfiguration.ProductItemID,
			VariationOptionID: productConfiguration.VariationOptionID,

			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteProductConfigurationParams{
					ProductItemID:     productConfiguration.ProductItemID,
					VariationOptionID: productConfiguration.VariationOptionID,
				}
				store.EXPECT().
					DeleteProductConfiguration(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:              "NotFound",
			AdminID:           admin.ID,
			ProductItemID:     productConfiguration.ProductItemID,
			VariationOptionID: productConfiguration.VariationOptionID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteProductConfigurationParams{
					ProductItemID:     productConfiguration.ProductItemID,
					VariationOptionID: productConfiguration.VariationOptionID,
				}

				store.EXPECT().
					DeleteProductConfiguration(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
			},
		},
		{
			name:              "InternalError",
			AdminID:           admin.ID,
			ProductItemID:     productConfiguration.ProductItemID,
			VariationOptionID: productConfiguration.VariationOptionID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteProductConfigurationParams{
					ProductItemID:     productConfiguration.ProductItemID,
					VariationOptionID: productConfiguration.VariationOptionID,
				}

				store.EXPECT().
					DeleteProductConfiguration(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:              "InvalidID",
			AdminID:           admin.ID,
			ProductItemID:     0,
			VariationOptionID: productConfiguration.VariationOptionID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteProductConfiguration(gomock.Any(), gomock.Any()).
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

			// build stubs
			tc.buildStub(store)

			// start test server and send request
			server := newTestServer(t, store, worker)
			//recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/admin/v1/admins/%d/product-configurations/%d/variation-options/%d", tc.AdminID, tc.ProductItemID, tc.VariationOptionID)
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

func randomProductConfigurationSuperAdmin(t *testing.T) (admin db.Admin, password string) {
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

func randomProductConfiguration() db.ProductConfiguration {
	return db.ProductConfiguration{
		ProductItemID:     util.RandomMoney(),
		VariationOptionID: util.RandomMoney(),
	}
}

func requireBodyMatchProductConfiguration(t *testing.T, body io.ReadCloser, productConfiguration db.ProductConfiguration) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotProductConfiguration db.ProductConfiguration
	err = json.Unmarshal(data, &gotProductConfiguration)
	require.NoError(t, err)
	require.Equal(t, productConfiguration, gotProductConfiguration)
}

func requireBodyMatchProductConfigurations(t *testing.T, body io.ReadCloser, ProductConfigurations []db.ProductConfiguration) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotProductConfigurations []db.ProductConfiguration
	err = json.Unmarshal(data, &gotProductConfigurations)
	require.NoError(t, err)
	require.Equal(t, ProductConfigurations, gotProductConfigurations)
}
