package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
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

func TestGetBrandPromotionAPI(t *testing.T) {
	brandPromotion := randomBrandPromotion()

	testCases := []struct {
		name          string
		BrandID       int64
		PromotionID   int64
		AdminID       int64
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rsp *http.Response)
	}{
		{
			name:        "OK",
			PromotionID: brandPromotion.PromotionID,
			BrandID:     brandPromotion.BrandID,
			buildStub: func(store *mockdb.MockStore) {
				arg := db.GetBrandPromotionParams{
					BrandID:     brandPromotion.BrandID,
					PromotionID: brandPromotion.PromotionID,
				}
				store.EXPECT().
					GetBrandPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(brandPromotion, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchBrandPromotion(t, rsp.Body, brandPromotion)
			},
		},

		{
			name:        "NotFound",
			PromotionID: brandPromotion.PromotionID,
			BrandID:     brandPromotion.BrandID,
			buildStub: func(store *mockdb.MockStore) {
				arg := db.GetBrandPromotionParams{
					BrandID:     brandPromotion.BrandID,
					PromotionID: brandPromotion.PromotionID,
				}
				store.EXPECT().
					GetBrandPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(nil, pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
			},
		},
		{
			name:        "InternalError",
			PromotionID: brandPromotion.PromotionID,
			BrandID:     brandPromotion.BrandID,
			buildStub: func(store *mockdb.MockStore) {
				arg := db.GetBrandPromotionParams{
					BrandID:     brandPromotion.BrandID,
					PromotionID: brandPromotion.PromotionID,
				}
				store.EXPECT().
					GetBrandPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(nil, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:        "InvalidID",
			PromotionID: 0,
			BrandID:     brandPromotion.BrandID,
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetBrandPromotion(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/api/v1/brand-promotions/%d/brands/%d", tc.PromotionID, tc.BrandID)
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(t, rsp)
		})

	}

}

func TestCreateBrandPromotionAPI(t *testing.T) {
	admin, _ := randomBrandPromotionSuperAdmin(t)
	brandPromotion := randomBrandPromotion()

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
				"brand_id":              brandPromotion.BrandID,
				"promotion_id":          brandPromotion.PromotionID,
				"brand_promotion_image": brandPromotion.BrandPromotionImage,
				"active":                brandPromotion.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminCreateBrandPromotionParams{
					AdminID:             admin.ID,
					BrandID:             brandPromotion.BrandID,
					PromotionID:         brandPromotion.PromotionID,
					BrandPromotionImage: brandPromotion.BrandPromotionImage,
					Active:              brandPromotion.Active,
				}

				store.EXPECT().
					AdminCreateBrandPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(brandPromotion, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchBrandPromotion(t, rsp.Body, brandPromotion)
			},
		},
		{
			name:    "NoAuthorization",
			AdminID: admin.ID,
			body: fiber.Map{
				"brand_id":              brandPromotion.BrandID,
				"promotion_id":          brandPromotion.PromotionID,
				"brand_promotion_image": brandPromotion.BrandPromotionImage,
				"active":                brandPromotion.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminCreateBrandPromotionParams{
					AdminID:             admin.ID,
					BrandID:             brandPromotion.BrandID,
					PromotionID:         brandPromotion.PromotionID,
					BrandPromotionImage: brandPromotion.BrandPromotionImage,
					Active:              brandPromotion.Active,
				}

				store.EXPECT().
					AdminCreateBrandPromotion(gomock.Any(), gomock.Eq(arg)).
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
				"brand_id":              brandPromotion.BrandID,
				"promotion_id":          brandPromotion.PromotionID,
				"brand_promotion_image": brandPromotion.BrandPromotionImage,
				"active":                brandPromotion.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, false, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminCreateBrandPromotionParams{
					AdminID:             admin.ID,
					BrandID:             brandPromotion.BrandID,
					PromotionID:         brandPromotion.PromotionID,
					BrandPromotionImage: brandPromotion.BrandPromotionImage,
					Active:              brandPromotion.Active,
				}

				store.EXPECT().
					AdminCreateBrandPromotion(gomock.Any(), gomock.Eq(arg)).
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
				"brand_id":              brandPromotion.BrandID,
				"promotion_id":          brandPromotion.PromotionID,
				"brand_promotion_image": brandPromotion.BrandPromotionImage,
				"active":                brandPromotion.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					AdminCreateBrandPromotion(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:    "InvalidID",
			AdminID: 0,
			body: fiber.Map{
				"brand_id":              brandPromotion.BrandID,
				"promotion_id":          brandPromotion.PromotionID,
				"brand_promotion_image": brandPromotion.BrandPromotionImage,
				"active":                brandPromotion.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateBrandPromotion(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/admin/v1/admins/%d/brand-promotions", tc.AdminID)
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

func TestListBrandPromotionsAPI(t *testing.T) {
	n := 5
	brandPromotions := make([]*db.BrandPromotion, n)
	for i := 0; i < n; i++ {
		brandPromotions[i] = randomBrandPromotion()
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
				arg := db.ListBrandPromotionsParams{
					Limit:  int32(n),
					Offset: 0,
				}

				store.EXPECT().
					ListBrandPromotions(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(brandPromotions, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchBrandPromotions(t, rsp.Body, brandPromotions)
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
					ListBrandPromotions(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, pgx.ErrTxClosed)
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
					ListBrandPromotions(gomock.Any(), gomock.Any()).
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
					ListBrandPromotions(gomock.Any(), gomock.Any()).
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

			url := "/api/v1/brand-promotions"
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			// Add query parameters to request URL
			q := request.URL.Query()
			q.Add("page_id", strconv.Itoa(tc.query.pageID))
			q.Add("page_size", strconv.Itoa(tc.query.pageSize))
			request.URL.RawQuery = q.Encode()

			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
		})
	}
}

func TestAdminUpdateBrandPromotionAPI(t *testing.T) {
	admin, _ := randomBrandPromotionSuperAdmin(t)
	brandPromotion := randomBrandPromotion()

	testCases := []struct {
		name          string
		body          fiber.Map
		BrandID       int64
		PromotionID   int64
		AdminID       int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rsp *http.Response)
	}{
		{
			name:        "OK",
			BrandID:     brandPromotion.BrandID,
			PromotionID: brandPromotion.PromotionID,
			AdminID:     admin.ID,
			body: fiber.Map{
				"active": brandPromotion.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminUpdateBrandPromotionParams{
					AdminID:     admin.ID,
					PromotionID: brandPromotion.PromotionID,
					Active:      null.BoolFromPtr(&brandPromotion.Active),
					BrandID:     brandPromotion.BrandID,
				}

				store.EXPECT().
					AdminUpdateBrandPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(brandPromotion, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:        "Unauthorized",
			BrandID:     brandPromotion.BrandID,
			PromotionID: brandPromotion.PromotionID,
			AdminID:     admin.ID,
			body: fiber.Map{
				"active": brandPromotion.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, false, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminUpdateBrandPromotionParams{
					AdminID:     admin.ID,
					PromotionID: brandPromotion.PromotionID,
					Active:      null.BoolFromPtr(&brandPromotion.Active),
					BrandID:     brandPromotion.BrandID,
				}

				store.EXPECT().
					AdminUpdateBrandPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:        "NoAuthorization",
			BrandID:     brandPromotion.BrandID,
			PromotionID: brandPromotion.PromotionID,
			AdminID:     admin.ID,
			body: fiber.Map{
				"active": brandPromotion.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminUpdateBrandPromotionParams{
					AdminID:     admin.ID,
					PromotionID: brandPromotion.PromotionID,
					Active:      null.BoolFromPtr(&brandPromotion.Active),
					BrandID:     brandPromotion.BrandID,
				}

				store.EXPECT().
					AdminUpdateBrandPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:        "InternalError",
			BrandID:     brandPromotion.BrandID,
			PromotionID: brandPromotion.PromotionID,
			AdminID:     admin.ID,
			body: fiber.Map{
				"active": brandPromotion.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminUpdateBrandPromotionParams{
					AdminID:     admin.ID,
					PromotionID: brandPromotion.PromotionID,
					Active:      null.BoolFromPtr(&brandPromotion.Active),
					BrandID:     brandPromotion.BrandID,
				}
				store.EXPECT().
					AdminUpdateBrandPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(nil, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:        "InvalidID",
			BrandID:     0,
			PromotionID: brandPromotion.PromotionID,
			AdminID:     admin.ID,
			body: fiber.Map{
				"active": brandPromotion.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					AdminUpdateBrandPromotion(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/admin/v1/admins/%d/brand-promotions/%d/brands/%d", tc.AdminID, tc.PromotionID, tc.BrandID)
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

func TestAdminListBrandPromotionsAPI(t *testing.T) {
	admin, _ := randomBrandPromotionSuperAdmin(t)
	n := 5
	brandPromotions := make([]*db.AdminListBrandPromotionsRow, n)
	for i := 0; i < n; i++ {
		brandPromotions[i] = randomBrandPromotionForAdmin()
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
					AdminListBrandPromotions(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(brandPromotions, nil)
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
					AdminListBrandPromotions(gomock.Any(), gomock.Eq(admin.ID)).
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
					AdminListBrandPromotions(gomock.Any(), gomock.Eq(admin.ID)).
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
					AdminListBrandPromotions(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, pgx.ErrTxClosed)
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
					AdminListBrandPromotions(gomock.Any(), gomock.Any()).
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
			// data, err := json.Marshal(tc.body)
			// require.NoError(t, err)

			url := fmt.Sprintf("/admin/v1/admins/%d/brand-promotions", tc.AdminID)
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

func TestDeleteBrandPromotionAPI(t *testing.T) {
	admin, _ := randomBrandPromotionSuperAdmin(t)
	brandPromotion := randomBrandPromotion()

	testCases := []struct {
		name          string
		PromotionID   int64
		BrandID       int64
		AdminID       int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rsp *http.Response)
	}{
		{
			name:        "OK",
			PromotionID: brandPromotion.PromotionID,
			BrandID:     brandPromotion.BrandID,
			AdminID:     admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteBrandPromotionParams{
					BrandID:     brandPromotion.BrandID,
					PromotionID: brandPromotion.PromotionID,
				}
				store.EXPECT().
					DeleteBrandPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:        "Unauthorized",
			PromotionID: brandPromotion.PromotionID,
			BrandID:     brandPromotion.BrandID,
			AdminID:     admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, 2, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteBrandPromotionParams{
					BrandID:     brandPromotion.BrandID,
					PromotionID: brandPromotion.PromotionID,
				}
				store.EXPECT().
					DeleteBrandPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:        "NoAuthorization",
			PromotionID: brandPromotion.PromotionID,
			BrandID:     brandPromotion.BrandID,
			AdminID:     admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteBrandPromotionParams{
					BrandID:     brandPromotion.BrandID,
					PromotionID: brandPromotion.PromotionID,
				}
				store.EXPECT().
					DeleteBrandPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:        "NotFound",
			PromotionID: brandPromotion.PromotionID,
			BrandID:     brandPromotion.BrandID,
			AdminID:     admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteBrandPromotionParams{
					BrandID:     brandPromotion.BrandID,
					PromotionID: brandPromotion.PromotionID,
				}
				store.EXPECT().
					DeleteBrandPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
			},
		},
		{
			name:        "InternalError",
			PromotionID: brandPromotion.PromotionID,
			BrandID:     brandPromotion.BrandID,
			AdminID:     admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteBrandPromotionParams{
					BrandID:     brandPromotion.BrandID,
					PromotionID: brandPromotion.PromotionID,
				}
				store.EXPECT().
					DeleteBrandPromotion(gomock.Any(), gomock.Eq(arg)).
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
			BrandID:     brandPromotion.BrandID,
			AdminID:     admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteBrandPromotion(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/admin/v1/admins/%d/brand-promotions/%d/brands/%d", tc.AdminID, tc.PromotionID, tc.BrandID)
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

func randomBrandPromotionSuperAdmin(t *testing.T) (admin *db.Admin, password string) {
	password = util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	admin = &db.Admin{
		ID:       util.RandomMoney(),
		Username: util.RandomUser(),
		Email:    util.RandomEmail(),
		Password: hashedPassword,
		Active:   true,
		TypeID:   1,
	}
	return
}

func randomBrandPromotion() *db.BrandPromotion {
	return &db.BrandPromotion{
		BrandID:             util.RandomMoney(),
		PromotionID:         util.RandomMoney(),
		Active:              util.RandomBool(),
		BrandPromotionImage: null.StringFrom(util.RandomURL()),
	}
}

func randomBrandPromotionForAdmin() *db.AdminListBrandPromotionsRow {
	return &db.AdminListBrandPromotionsRow{
		BrandID:             util.RandomMoney(),
		BrandName:           null.StringFrom(util.RandomUser()),
		PromotionID:         util.RandomMoney(),
		PromotionName:       util.RandomUser(),
		Active:              util.RandomBool(),
		BrandPromotionImage: null.StringFrom(util.RandomURL()),
	}
}

func requireBodyMatchBrandPromotion(t *testing.T, body io.ReadCloser, brandPromotion *db.BrandPromotion) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotBrandPromotion *db.BrandPromotion
	err = json.Unmarshal(data, &gotBrandPromotion)
	require.NoError(t, err)
	require.Equal(t, brandPromotion, gotBrandPromotion)
}

func requireBodyMatchBrandPromotions(t *testing.T, body io.ReadCloser, BrandPromotions []*db.BrandPromotion) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotBrandPromotions []*db.BrandPromotion
	err = json.Unmarshal(data, &gotBrandPromotions)
	require.NoError(t, err)
	require.Equal(t, BrandPromotions, gotBrandPromotions)
}
