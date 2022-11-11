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

func TestGetProductCategoryAPI(t *testing.T) {
	productCategory := randomProductCategory()

	testCases := []struct {
		name              string
		productCategoryID int64
		buildStub         func(store *mockdb.MockStore)
		checkResponse     func(t *testing.T, recorder *httptest.ResponseRecorder)
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchProductCategory(t, recorder.Body, productCategory)
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
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

			url := fmt.Sprintf("/product-categories/%d", tc.productCategoryID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			//check response
			tc.checkResponse(t, recorder)
		})

	}

}

func TestCreateProductCategoryAPI(t *testing.T) {
	admin, _ := randomPCategorieSuperAdmin(t)
	productCategory := randomProductCategory()

	testCases := []struct {
		name              string
		body              gin.H
		productCategoryID int64
		setupAuth         func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs        func(store *mockdb.MockStore)
		checkResponse     func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
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
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchProductCategory(t, recorder.Body, productCategory)
			},
		},
		{
			name:              "NoAuthorization",
			productCategoryID: productCategory.ID,
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
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:              "Unauthorized",
			productCategoryID: productCategory.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, false, time.Minute)
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
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
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
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidID",
			body: gin.H{
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

			url := "/product-categories"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
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
		checkResponse func(recoder *httptest.ResponseRecorder)
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
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchProductCategories(t, recorder.Body, productCategories)
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
					ListProductCategories(gomock.Any(), gomock.Any()).
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
					ListProductCategories(gomock.Any(), gomock.Any()).
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

			url := "/product-categories"
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

func TestUpdateProductCategoryAPI(t *testing.T) {
	admin, _ := randomPCategorieSuperAdmin(t)
	productCategory := randomProductCategory()

	testCases := []struct {
		name              string
		body              gin.H
		productCategoryID int64
		setupAuth         func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs        func(store *mockdb.MockStore)
		checkResponse     func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:              "OK",
			productCategoryID: productCategory.ID,
			body: gin.H{
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:              "Unauthorized",
			productCategoryID: productCategory.ID,
			body: gin.H{
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:              "NoAuthorization",
			productCategoryID: productCategory.ID,
			body: gin.H{
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:              "InternalError",
			productCategoryID: productCategory.ID,
			body: gin.H{
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:              "InvalidID",
			productCategoryID: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateProductCategory(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/product-categories/%d", tc.productCategoryID)
			request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestDeleteProductCategoryAPI(t *testing.T) {
	admin, _ := randomPCategorieSuperAdmin(t)
	productCategory := randomProductCategory()

	testCases := []struct {
		name              string
		body              gin.H
		productCategoryID int64
		setupAuth         func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub         func(store *mockdb.MockStore)
		checkResponse     func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:              "OK",
			productCategoryID: productCategory.ID,
			body: gin.H{
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:              "Unauthorized",
			productCategoryID: productCategory.ID,
			body: gin.H{
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:              "NoAuthorization",
			productCategoryID: productCategory.ID,
			body: gin.H{
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:              "NotFound",
			productCategoryID: productCategory.ID,
			body: gin.H{
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:              "InternalError",
			productCategoryID: productCategory.ID,
			body: gin.H{
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:              "InvalidID",
			productCategoryID: 0,
			body: gin.H{
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

			url := fmt.Sprintf("/product-categories/%d", tc.productCategoryID)
			request, err := http.NewRequest(http.MethodDelete, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			//check response
			tc.checkResponse(t, recorder)
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

func requireBodyMatchProductCategory(t *testing.T, body *bytes.Buffer, productCategory db.ProductCategory) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotProductCategory db.ProductCategory
	err = json.Unmarshal(data, &gotProductCategory)
	require.NoError(t, err)
	require.Equal(t, productCategory, gotProductCategory)
}

func requireBodyMatchProductCategories(t *testing.T, body *bytes.Buffer, productCategories []db.ProductCategory) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotProductCategories []db.ProductCategory
	err = json.Unmarshal(data, &gotProductCategories)
	require.NoError(t, err)
	require.Equal(t, productCategories, gotProductCategories)
}
