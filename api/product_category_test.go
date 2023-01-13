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

func TestGetProductCategoryAPI(t *testing.T) {
	productCategory := randomProductCategory()

	testCases := []struct {
		name              string
		productCategoryID int64
		buildStub         func(store *mockdb.MockStore)
		checkResponse     func(t *testing.T, rsp *http.Response)
	}{
		{
			name:              "OK",
			productCategoryID: productCategory.ID,
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetProductCategory(gomock.Any(), gomock.Eq(productCategory.ID)).
					Times(1).
					Return(productCategory, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchProductCategory(t, rsp.Body, productCategory)
			},
		},

		{
			name:              "NotFound",
			productCategoryID: productCategory.ID,
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetProductCategory(gomock.Any(), gomock.Eq(productCategory.ID)).
					Times(1).
					Return(db.ProductCategory{}, pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
			},
		},
		{
			name:              "InternalError",
			productCategoryID: productCategory.ID,
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetProductCategory(gomock.Any(), gomock.Eq(productCategory.ID)).
					Times(1).
					Return(db.ProductCategory{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:              "InvalidID",
			productCategoryID: 0,
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetProductCategory(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/api/v1/categories/%d", tc.productCategoryID)
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(t, rsp)
		})

	}

}

func TestCreateProductCategoryAPI(t *testing.T) {
	admin, _ := randomPCategorieSuperAdmin(t)
	productCategory := randomProductCategory()

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
				"category_name":      productCategory.CategoryName,
				"parent_category_id": productCategory.ParentCategoryID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateProductCategoryParams{
					ParentCategoryID: productCategory.ParentCategoryID,
					CategoryName:     productCategory.CategoryName,
				}

				store.EXPECT().
					CreateProductCategory(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(productCategory, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchProductCategory(t, rsp.Body, productCategory)
			},
		},
		{
			name:    "NoAuthorization",
			AdminID: admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateProductCategoryParams{
					ParentCategoryID: productCategory.ParentCategoryID,
					CategoryName:     productCategory.CategoryName,
				}

				store.EXPECT().
					CreateProductCategory(gomock.Any(), gomock.Eq(arg)).
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
				"category_name":      productCategory.CategoryName,
				"parent_category_id": productCategory.ParentCategoryID,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateProductCategoryParams{
					ParentCategoryID: productCategory.ParentCategoryID,
					CategoryName:     productCategory.CategoryName,
				}

				store.EXPECT().
					CreateProductCategory(gomock.Any(), gomock.Eq(arg)).
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
				"category_name":      productCategory.CategoryName,
				"parent_category_id": productCategory.ParentCategoryID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateProductCategory(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.ProductCategory{}, pgx.ErrTxClosed)
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
					CreateProductCategory(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/api/admin/%d/v1/categories", tc.AdminID)
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

func TestListProductCategoriesAPI(t *testing.T) {
	n := 5
	productCategories := make([]db.ProductCategory, n)
	for i := 0; i < n; i++ {
		productCategories[i] = randomProductCategory()
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
				arg := db.ListProductCategoriesParams{
					Limit:  int32(n),
					Offset: 0,
				}

				store.EXPECT().
					ListProductCategories(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(productCategories, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchProductCategories(t, rsp.Body, productCategories)
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
					ListProductCategories(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.ProductCategory{}, pgx.ErrTxClosed)
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
					ListProductCategories(gomock.Any(), gomock.Any()).
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
					ListProductCategories(gomock.Any(), gomock.Any()).
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

			url := "/api/v1/categories"
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

func TestUpdateProductCategoryAPI(t *testing.T) {
	admin, _ := randomPCategorieSuperAdmin(t)
	productCategory := randomProductCategory()

	testCases := []struct {
		name              string
		body              fiber.Map
		productCategoryID int64
		AdminID           int64
		setupAuth         func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs        func(store *mockdb.MockStore)
		checkResponse     func(t *testing.T, rsp *http.Response)
	}{
		{
			name:              "OK",
			AdminID:           admin.ID,
			productCategoryID: productCategory.ID,
			body: fiber.Map{
				"category_name":      "new name",
				"parent_category_id": productCategory.ParentCategoryID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateProductCategoryParams{
					CategoryName:     "new name",
					ID:               productCategory.ID,
					ParentCategoryID: productCategory.ParentCategoryID,
				}

				store.EXPECT().
					UpdateProductCategory(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(productCategory, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:              "Unauthorized",
			AdminID:           admin.ID,
			productCategoryID: productCategory.ID,
			body: fiber.Map{
				"category_name":      "new name",
				"parent_category_id": productCategory.ParentCategoryID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, false, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateProductCategoryParams{
					CategoryName:     "new name",
					ID:               productCategory.ID,
					ParentCategoryID: productCategory.ParentCategoryID,
				}

				store.EXPECT().
					UpdateProductCategory(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:              "NoAuthorization",
			productCategoryID: productCategory.ID,
			AdminID:           admin.ID,
			body: fiber.Map{
				"category_name":      "new name",
				"parent_category_id": productCategory.ParentCategoryID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateProductCategoryParams{
					CategoryName:     "new name",
					ID:               productCategory.ID,
					ParentCategoryID: productCategory.ParentCategoryID,
				}

				store.EXPECT().
					UpdateProductCategory(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:              "InternalError",
			productCategoryID: productCategory.ID,
			AdminID:           admin.ID,
			body: fiber.Map{
				"category_name":      "new name",
				"parent_category_id": productCategory.ParentCategoryID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateProductCategoryParams{
					CategoryName:     "new name",
					ID:               productCategory.ID,
					ParentCategoryID: productCategory.ParentCategoryID,
				}
				store.EXPECT().
					UpdateProductCategory(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.ProductCategory{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:              "InvalidID",
			AdminID:           admin.ID,
			productCategoryID: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateProductCategory(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/api/admin/%d/v1/categories/%d", tc.AdminID, tc.productCategoryID)
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

func TestDeleteProductCategoryAPI(t *testing.T) {
	admin, _ := randomPCategorieSuperAdmin(t)
	productCategory := randomProductCategory()

	testCases := []struct {
		name              string
		body              fiber.Map
		productCategoryID int64
		AdminID           int64
		setupAuth         func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub         func(store *mockdb.MockStore)
		checkResponse     func(t *testing.T, rsp *http.Response)
	}{
		{
			name:              "OK",
			productCategoryID: productCategory.ID,
			AdminID:           admin.ID,
			body: fiber.Map{
				"parent_category_id": productCategory.ParentCategoryID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteProductCategoryParams{
					ID:               productCategory.ID,
					ParentCategoryID: null.IntFromPtr(&productCategory.ParentCategoryID.Int64),
				}
				store.EXPECT().
					DeleteProductCategory(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:              "Unauthorized",
			productCategoryID: productCategory.ID,
			AdminID:           admin.ID,
			body: fiber.Map{
				"parent_category_id": productCategory.ParentCategoryID, "id": productCategory.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, 2, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteProductCategoryParams{
					ID:               productCategory.ID,
					ParentCategoryID: null.IntFromPtr(&productCategory.ParentCategoryID.Int64),
				}
				store.EXPECT().
					DeleteProductCategory(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:              "NoAuthorization",
			productCategoryID: productCategory.ID,
			AdminID:           admin.ID,
			body: fiber.Map{
				"parent_category_id": productCategory.ParentCategoryID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteProductCategoryParams{
					ID:               productCategory.ID,
					ParentCategoryID: null.IntFromPtr(&productCategory.ParentCategoryID.Int64),
				}
				store.EXPECT().
					DeleteProductCategory(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:              "NotFound",
			productCategoryID: productCategory.ID,
			AdminID:           admin.ID,
			body: fiber.Map{
				"parent_category_id": productCategory.ParentCategoryID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteProductCategoryParams{
					ID:               productCategory.ID,
					ParentCategoryID: null.IntFromPtr(&productCategory.ParentCategoryID.Int64),
				}
				store.EXPECT().
					DeleteProductCategory(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
			},
		},
		{
			name:              "InternalError",
			productCategoryID: productCategory.ID,
			AdminID:           admin.ID,
			body: fiber.Map{
				"parent_category_id": productCategory.ParentCategoryID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteProductCategoryParams{
					ID:               productCategory.ID,
					ParentCategoryID: null.IntFromPtr(&productCategory.ParentCategoryID.Int64),
				}
				store.EXPECT().
					DeleteProductCategory(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:              "InvalidID",
			productCategoryID: 0,
			AdminID:           admin.ID,
			body: fiber.Map{
				"parent_category_id": 0,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteProductCategory(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/api/admin/%d/v1/categories/%d", tc.AdminID, tc.productCategoryID)
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

func randomPCategorieSuperAdmin(t *testing.T) (admin db.Admin, password string) {
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

func randomProductCategory() db.ProductCategory {
	return db.ProductCategory{
		ID:               util.RandomInt(1, 1000),
		ParentCategoryID: null.IntFrom(util.RandomMoney()),
		CategoryName:     util.RandomUser(),
	}
}

func requireBodyMatchProductCategory(t *testing.T, body io.ReadCloser, productCategory db.ProductCategory) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotProductCategory db.ProductCategory
	err = json.Unmarshal(data, &gotProductCategory)
	require.NoError(t, err)
	require.Equal(t, productCategory, gotProductCategory)
}

func requireBodyMatchProductCategories(t *testing.T, body io.ReadCloser, productCategories []db.ProductCategory) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotProductCategories []db.ProductCategory
	err = json.Unmarshal(data, &gotProductCategories)
	require.NoError(t, err)
	require.Equal(t, productCategories, gotProductCategories)
}
