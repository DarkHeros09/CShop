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

func TestGetProductPromotionAPI(t *testing.T) {
	productPromotion := randomProductPromotion()

	testCases := []struct {
		name          string
		PromotionID   int64
		ProductID     int64
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(rsp *http.Response)
	}{
		{
			name:        "OK",
			PromotionID: productPromotion.PromotionID,
			ProductID:   productPromotion.ProductID,
			buildStub: func(store *mockdb.MockStore) {

				arg := db.GetProductPromotionParams{
					ProductID:   productPromotion.ProductID,
					PromotionID: productPromotion.PromotionID,
				}

				store.EXPECT().
					GetProductPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(productPromotion, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchProductPromotion(t, rsp.Body, productPromotion)
			},
		},

		{
			name:        "NotFound",
			PromotionID: productPromotion.PromotionID,
			ProductID:   productPromotion.ProductID,
			buildStub: func(store *mockdb.MockStore) {
				arg := db.GetProductPromotionParams{
					ProductID:   productPromotion.ProductID,
					PromotionID: productPromotion.PromotionID,
				}

				store.EXPECT().
					GetProductPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.ProductPromotion{}, pgx.ErrNoRows)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
			},
		},
		{
			name:        "InternalError",
			PromotionID: productPromotion.PromotionID,
			ProductID:   productPromotion.ProductID,
			buildStub: func(store *mockdb.MockStore) {
				arg := db.GetProductPromotionParams{
					ProductID:   productPromotion.ProductID,
					PromotionID: productPromotion.PromotionID,
				}

				store.EXPECT().
					GetProductPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.ProductPromotion{}, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:        "InvalidID",
			PromotionID: 0,
			ProductID:   productPromotion.ProductID,
			buildStub: func(store *mockdb.MockStore) {

				store.EXPECT().
					GetProductPromotion(gomock.Any(), gomock.Any()).
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
			ctrl := gomock.NewController(t) // no need to call defer ctrl.finish() in 1.6V

			store := mockdb.NewMockStore(ctrl)
			worker := mockwk.NewMockTaskDistributor(ctrl)
			ik := mockik.NewMockImageKitManagement(ctrl)

			// build stubs
			tc.buildStub(store)

			// start test server and send request
			server := newTestServer(t, store, worker, ik)
			//recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/api/v1/product-promotions/%d/products/%d", tc.PromotionID, tc.ProductID)
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			//check response
			tc.checkResponse(rsp)
		})

	}

}

func TestCreateProductPromotionAPI(t *testing.T) {
	admin, _ := randomProductPromotionSuperAdmin(t)
	productPromotion := randomProductPromotion()

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
				"product_id":              productPromotion.ProductID,
				"promotion_id":            productPromotion.PromotionID,
				"product_promotion_image": productPromotion.ProductPromotionImage,
				"active":                  productPromotion.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminCreateProductPromotionParams{
					AdminID:               admin.ID,
					ProductID:             productPromotion.ProductID,
					PromotionID:           productPromotion.PromotionID,
					ProductPromotionImage: productPromotion.ProductPromotionImage,
					Active:                productPromotion.Active,
				}

				store.EXPECT().
					AdminCreateProductPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(productPromotion, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchProductPromotion(t, rsp.Body, productPromotion)
			},
		},
		{
			name:    "NoAuthorization",
			AdminID: admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminCreateProductPromotionParams{
					AdminID:               admin.ID,
					ProductID:             productPromotion.ProductID,
					PromotionID:           productPromotion.PromotionID,
					ProductPromotionImage: productPromotion.ProductPromotionImage,
					Active:                productPromotion.Active,
				}

				store.EXPECT().
					AdminCreateProductPromotion(gomock.Any(), gomock.Eq(arg)).
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
				"product_id":              productPromotion.ProductID,
				"promotion_id":            productPromotion.PromotionID,
				"product_promotion_image": productPromotion.ProductPromotionImage,
				"active":                  productPromotion.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, false, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminCreateProductPromotionParams{
					AdminID:               admin.ID,
					ProductID:             productPromotion.ProductID,
					PromotionID:           productPromotion.PromotionID,
					ProductPromotionImage: productPromotion.ProductPromotionImage,
					Active:                productPromotion.Active,
				}

				store.EXPECT().
					AdminCreateProductPromotion(gomock.Any(), gomock.Eq(arg)).
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
				"product_id":              productPromotion.ProductID,
				"promotion_id":            productPromotion.PromotionID,
				"product_promotion_image": productPromotion.ProductPromotionImage,
				"active":                  productPromotion.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					AdminCreateProductPromotion(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.ProductPromotion{}, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:    "InvalidID",
			AdminID: 0,
			body: fiber.Map{
				"product_id":              productPromotion.ProductID,
				"promotion_id":            productPromotion.PromotionID,
				"product_promotion_image": productPromotion.ProductPromotionImage,
				"active":                  productPromotion.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					AdminCreateProductPromotion(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/admin/v1/admins/%d/product-promotions", tc.AdminID)
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
		checkResponse func(rsp *http.Response)
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
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchProductPromotions(t, rsp.Body, productPromotions)
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
					ListProductPromotions(gomock.Any(), gomock.Any()).
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
					ListProductPromotions(gomock.Any(), gomock.Any()).
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

			url := "/api/v1/product-promotions"
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

func TestAdminUpdateProductPromotionAPI(t *testing.T) {
	admin, _ := randomProductPromotionSuperAdmin(t)
	productPromotion := randomProductPromotion()

	testCases := []struct {
		name          string
		body          fiber.Map
		PromotionID   int64
		ProductID     int64
		AdminID       int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rsp *http.Response)
	}{
		{
			name:        "OK",
			PromotionID: productPromotion.PromotionID,
			ProductID:   productPromotion.ProductID,
			AdminID:     admin.ID,
			body: fiber.Map{
				"active": productPromotion.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminUpdateProductPromotionParams{
					AdminID:     admin.ID,
					PromotionID: productPromotion.PromotionID,
					Active:      null.BoolFromPtr(&productPromotion.Active),
					ProductID:   productPromotion.ProductID,
				}

				store.EXPECT().
					AdminUpdateProductPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(productPromotion, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:        "Unauthorized",
			PromotionID: productPromotion.PromotionID,
			ProductID:   productPromotion.ProductID,
			AdminID:     admin.ID,
			body: fiber.Map{
				"active": productPromotion.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, false, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminUpdateProductPromotionParams{
					AdminID:     admin.ID,
					PromotionID: productPromotion.PromotionID,
					Active:      null.BoolFromPtr(&productPromotion.Active),
					ProductID:   productPromotion.ProductID,
				}

				store.EXPECT().
					AdminUpdateProductPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:        "NoAuthorization",
			PromotionID: productPromotion.PromotionID,
			ProductID:   productPromotion.ProductID,
			AdminID:     admin.ID,
			body: fiber.Map{
				"active": productPromotion.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminUpdateProductPromotionParams{
					AdminID:     admin.ID,
					PromotionID: productPromotion.PromotionID,
					Active:      null.BoolFromPtr(&productPromotion.Active),
					ProductID:   productPromotion.ProductID,
				}

				store.EXPECT().
					AdminUpdateProductPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:        "InternalError",
			PromotionID: productPromotion.PromotionID,
			ProductID:   productPromotion.ProductID,
			AdminID:     admin.ID,
			body: fiber.Map{
				"active": productPromotion.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminUpdateProductPromotionParams{
					AdminID:     admin.ID,
					PromotionID: productPromotion.PromotionID,
					Active:      null.BoolFromPtr(&productPromotion.Active),
					ProductID:   productPromotion.ProductID,
				}
				store.EXPECT().
					AdminUpdateProductPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.ProductPromotion{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:        "InvalidID",
			PromotionID: 0,
			ProductID:   productPromotion.ProductID,
			AdminID:     admin.ID,
			body: fiber.Map{
				"active": productPromotion.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					AdminUpdateProductPromotion(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/admin/v1/admins/%d/product-promotions/%d/products/%d", tc.AdminID, tc.PromotionID, tc.ProductID)
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

func TestAdminListProductPromotionsAPI(t *testing.T) {
	admin, _ := randomProductPromotionSuperAdmin(t)
	n := 5
	productPromotions := make([]db.AdminListProductPromotionsRow, n)
	for i := 0; i < n; i++ {
		productPromotions[i] = randomProductPromotionForAdmin()
	}

	testCases := []struct {
		name          string
		AdminID       int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rsp *http.Response)
	}{
		{
			name:    "OK",
			AdminID: admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					AdminListProductPromotions(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(productPromotions, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:    "Unauthorized",
			AdminID: admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, false, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					AdminListProductPromotions(gomock.Any(), gomock.Eq(admin.ID)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:    "NoAuthorization",
			AdminID: admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					AdminListProductPromotions(gomock.Any(), gomock.Eq(admin.ID)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:    "InternalError",
			AdminID: admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					AdminListProductPromotions(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.AdminListProductPromotionsRow{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:    "InvalidID",
			AdminID: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					AdminListProductPromotions(gomock.Any(), gomock.Any()).
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
			// data, err := json.Marshal(tc.body)
			// require.NoError(t, err)

			url := fmt.Sprintf("/admin/v1/admins/%d/product-promotions", tc.AdminID)
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.adminTokenMaker)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(t, rsp)
		})
	}
}

func TestDeleteProductPromotionAPI(t *testing.T) {
	admin, _ := randomProductPromotionSuperAdmin(t)
	productPromotion := randomProductPromotion()

	testCases := []struct {
		name          string
		PromotionID   int64
		ProductID     int64
		AdminID       int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rsp *http.Response)
	}{
		{
			name:        "OK",
			PromotionID: productPromotion.PromotionID,
			ProductID:   productPromotion.ProductID,
			AdminID:     admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteProductPromotionParams{
					ProductID:   productPromotion.ProductID,
					PromotionID: productPromotion.PromotionID,
				}
				store.EXPECT().
					DeleteProductPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:        "Unauthorized",
			PromotionID: productPromotion.PromotionID,
			ProductID:   productPromotion.ProductID,
			AdminID:     admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, 2, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteProductPromotionParams{
					ProductID:   productPromotion.ProductID,
					PromotionID: productPromotion.PromotionID,
				}
				store.EXPECT().
					DeleteProductPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:        "NoAuthorization",
			PromotionID: productPromotion.PromotionID,
			ProductID:   productPromotion.ProductID,
			AdminID:     admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteProductPromotionParams{
					ProductID:   productPromotion.ProductID,
					PromotionID: productPromotion.PromotionID,
				}
				store.EXPECT().
					DeleteProductPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:        "NotFound",
			PromotionID: productPromotion.PromotionID,
			ProductID:   productPromotion.ProductID,
			AdminID:     admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteProductPromotionParams{
					ProductID:   productPromotion.ProductID,
					PromotionID: productPromotion.PromotionID,
				}
				store.EXPECT().
					DeleteProductPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
			},
		},
		{
			name:        "InternalError",
			PromotionID: productPromotion.PromotionID,
			ProductID:   productPromotion.ProductID,
			AdminID:     admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteProductPromotionParams{
					ProductID:   productPromotion.ProductID,
					PromotionID: productPromotion.PromotionID,
				}
				store.EXPECT().
					DeleteProductPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:        "InvalidID",
			PromotionID: 0,
			ProductID:   productPromotion.ProductID,
			AdminID:     admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteProductPromotion(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/admin/v1/admins/%d/product-promotions/%d/products/%d", tc.AdminID, tc.PromotionID, tc.ProductID)
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
		ProductID:             util.RandomMoney(),
		PromotionID:           util.RandomMoney(),
		Active:                util.RandomBool(),
		ProductPromotionImage: null.StringFrom(util.RandomURL()),
	}
}

func randomProductPromotionForAdmin() db.AdminListProductPromotionsRow {
	return db.AdminListProductPromotionsRow{
		ProductID:             util.RandomMoney(),
		ProductName:           null.StringFrom(util.RandomUser()),
		PromotionID:           util.RandomMoney(),
		PromotionName:         null.StringFrom(util.RandomUser()),
		Active:                util.RandomBool(),
		ProductPromotionImage: null.StringFrom(util.RandomURL()),
	}
}

func requireBodyMatchProductPromotion(t *testing.T, body io.ReadCloser, productPromotion db.ProductPromotion) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotProductPromotion db.ProductPromotion
	err = json.Unmarshal(data, &gotProductPromotion)
	require.NoError(t, err)
	require.Equal(t, productPromotion.ProductID, gotProductPromotion.ProductID)
	require.Equal(t, productPromotion.PromotionID, gotProductPromotion.PromotionID)
	require.Equal(t, productPromotion.Active, gotProductPromotion.Active)
}

func requireBodyMatchProductPromotions(t *testing.T, body io.ReadCloser, productPromotions []db.ProductPromotion) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotProductPromotions []db.ProductPromotion
	err = json.Unmarshal(data, &gotProductPromotions)
	require.NoError(t, err)
	require.Equal(t, productPromotions, gotProductPromotions)
}
