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

func TestGetPromotionAPI(t *testing.T) {
	promotion := randomPromotion()

	testCases := []struct {
		name          string
		PromotionID   int64
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rsp *http.Response)
	}{
		{
			name:        "OK",
			PromotionID: promotion.ID,
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetPromotion(gomock.Any(), gomock.Eq(promotion.ID)).
					Times(1).
					Return(promotion, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchPromotion(t, rsp.Body, promotion)
			},
		},

		{
			name:        "NotFound",
			PromotionID: promotion.ID,
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetPromotion(gomock.Any(), gomock.Eq(promotion.ID)).
					Times(1).
					Return(db.Promotion{}, pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
			},
		},
		{
			name:        "InternalError",
			PromotionID: promotion.ID,
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetPromotion(gomock.Any(), gomock.Eq(promotion.ID)).
					Times(1).
					Return(db.Promotion{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:        "InvalidID",
			PromotionID: 0,
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetPromotion(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/api/v1/promotions/%d", tc.PromotionID)
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(t, rsp)
		})

	}

}

func TestCreatePromotionAPI(t *testing.T) {
	admin, _ := randomPSuperAdmin(t)
	promotion := randomPromotion()

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
				"name":          promotion.Name,
				"description":   promotion.Description,
				"discount_rate": promotion.DiscountRate,
				"active":        promotion.Active,
				"start_date":    promotion.StartDate,
				"end_date":      promotion.EndDate,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreatePromotionParams{
					Name:         promotion.Name,
					Description:  promotion.Description,
					DiscountRate: promotion.DiscountRate,
					Active:       promotion.Active,
					StartDate:    promotion.StartDate,
					EndDate:      promotion.EndDate,
				}

				store.EXPECT().
					CreatePromotion(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(promotion, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchPromotion(t, rsp.Body, promotion)
			},
		},
		{
			name:    "NoAuthorization",
			AdminID: admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreatePromotionParams{
					Name:         promotion.Name,
					Description:  promotion.Description,
					DiscountRate: promotion.DiscountRate,
					Active:       promotion.Active,
					StartDate:    promotion.StartDate,
					EndDate:      promotion.EndDate,
				}

				store.EXPECT().
					CreatePromotion(gomock.Any(), gomock.Eq(arg)).
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
				"name":          promotion.Name,
				"description":   promotion.Description,
				"discount_rate": promotion.DiscountRate,
				"active":        promotion.Active,
				"start_date":    promotion.StartDate,
				"end_date":      promotion.EndDate,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, false, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreatePromotionParams{
					Name:         promotion.Name,
					Description:  promotion.Description,
					DiscountRate: promotion.DiscountRate,
					Active:       promotion.Active,
					StartDate:    promotion.StartDate,
					EndDate:      promotion.EndDate,
				}

				store.EXPECT().
					CreatePromotion(gomock.Any(), gomock.Eq(arg)).
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
				"name":          promotion.Name,
				"description":   promotion.Description,
				"discount_rate": promotion.DiscountRate,
				"active":        promotion.Active,
				"start_date":    promotion.StartDate,
				"end_date":      promotion.EndDate,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreatePromotion(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Promotion{}, pgx.ErrTxClosed)
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
					CreatePromotion(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/admin/v1/admins/%d/promotions", tc.AdminID)
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

func TestListPromotionsAPI(t *testing.T) {
	n := 5
	Promotions := make([]db.Promotion, n)
	for i := 0; i < n; i++ {
		Promotions[i] = randomPromotion()
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
				arg := db.ListPromotionsParams{
					Limit:  int32(n),
					Offset: 0,
				}

				store.EXPECT().
					ListPromotions(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(Promotions, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchPromotions(t, rsp.Body, Promotions)
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
					ListPromotions(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.Promotion{}, pgx.ErrTxClosed)
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
					ListPromotions(gomock.Any(), gomock.Any()).
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
					ListPromotions(gomock.Any(), gomock.Any()).
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

			url := "/api/v1/promotions"
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

func TestUpdatePromotionAPI(t *testing.T) {
	admin, _ := randomPSuperAdmin(t)
	promotion := randomPromotion()

	testCases := []struct {
		name          string
		body          fiber.Map
		PromotionID   int64
		AdminID       int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rsp *http.Response)
	}{
		{
			name:        "OK",
			PromotionID: promotion.ID,
			AdminID:     admin.ID,
			body: fiber.Map{
				"name":          promotion.Name,
				"description":   "new discription",
				"discount_rate": promotion.DiscountRate,
				"active":        promotion.Active,
				"start_date":    promotion.StartDate,
				"end_date":      promotion.EndDate,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdatePromotionParams{
					Name:         null.StringFromPtr(&promotion.Name),
					Description:  null.StringFrom("new discription"),
					DiscountRate: null.IntFromPtr(&promotion.DiscountRate),
					Active:       null.BoolFromPtr(&promotion.Active),
					StartDate:    null.TimeFromPtr(&promotion.StartDate),
					EndDate:      null.TimeFromPtr(&promotion.EndDate),
					ID:           promotion.ID,
				}

				store.EXPECT().
					UpdatePromotion(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(promotion, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:        "Unauthorized",
			PromotionID: promotion.ID,
			AdminID:     admin.ID,
			body: fiber.Map{
				"name":          promotion.Name,
				"description":   "new discription",
				"discount_rate": promotion.DiscountRate,
				"active":        promotion.Active,
				"start_date":    promotion.StartDate,
				"end_date":      promotion.EndDate,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, false, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdatePromotionParams{
					Name:         null.StringFromPtr(&promotion.Name),
					Description:  null.StringFrom("new discription"),
					DiscountRate: null.IntFromPtr(&promotion.DiscountRate),
					Active:       null.BoolFromPtr(&promotion.Active),
					StartDate:    null.TimeFromPtr(&promotion.StartDate),
					EndDate:      null.TimeFromPtr(&promotion.EndDate),
					ID:           promotion.ID,
				}

				store.EXPECT().
					UpdatePromotion(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:        "NoAuthorization",
			PromotionID: promotion.ID,
			AdminID:     admin.ID,
			body: fiber.Map{
				"name":          promotion.Name,
				"description":   "new discription",
				"discount_rate": promotion.DiscountRate,
				"active":        promotion.Active,
				"start_date":    promotion.StartDate,
				"end_date":      promotion.EndDate,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdatePromotionParams{
					Name:         null.StringFromPtr(&promotion.Name),
					Description:  null.StringFrom("new discription"),
					DiscountRate: null.IntFromPtr(&promotion.DiscountRate),
					Active:       null.BoolFromPtr(&promotion.Active),
					StartDate:    null.TimeFromPtr(&promotion.StartDate),
					EndDate:      null.TimeFromPtr(&promotion.EndDate),
					ID:           promotion.ID,
				}

				store.EXPECT().
					UpdatePromotion(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:        "InternalError",
			PromotionID: promotion.ID,
			AdminID:     admin.ID,
			body: fiber.Map{
				"name":          promotion.Name,
				"description":   "new discription",
				"discount_rate": promotion.DiscountRate,
				"active":        promotion.Active,
				"start_date":    promotion.StartDate,
				"end_date":      promotion.EndDate,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdatePromotionParams{
					Name:         null.StringFromPtr(&promotion.Name),
					Description:  null.StringFrom("new discription"),
					DiscountRate: null.IntFromPtr(&promotion.DiscountRate),
					Active:       null.BoolFromPtr(&promotion.Active),
					StartDate:    null.TimeFromPtr(&promotion.StartDate),
					EndDate:      null.TimeFromPtr(&promotion.EndDate),
					ID:           promotion.ID,
				}
				store.EXPECT().
					UpdatePromotion(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.Promotion{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:        "InvalidID",
			PromotionID: 0,
			AdminID:     admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdatePromotion(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/admin/v1/admins/%d/promotions/%d", tc.AdminID, tc.PromotionID)
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

func TestDeletePromotionAPI(t *testing.T) {
	admin, _ := randomPSuperAdmin(t)
	promotion := randomPromotion()

	testCases := []struct {
		name          string
		PromotionID   int64
		AdminID       int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rsp *http.Response)
	}{
		{
			name:        "OK",
			PromotionID: promotion.ID,
			AdminID:     admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeletePromotion(gomock.Any(), gomock.Eq(promotion.ID)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:        "Unauthorized",
			PromotionID: promotion.ID,
			AdminID:     admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, 2, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeletePromotion(gomock.Any(), gomock.Eq(promotion.ID)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:        "NoAuthorization",
			PromotionID: promotion.ID,
			AdminID:     admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeletePromotion(gomock.Any(), gomock.Eq(promotion.ID)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:        "NotFound",
			PromotionID: promotion.ID,
			AdminID:     admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeletePromotion(gomock.Any(), gomock.Eq(promotion.ID)).
					Times(1).
					Return(pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
			},
		},
		{
			name:        "InternalError",
			PromotionID: promotion.ID,
			AdminID:     admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeletePromotion(gomock.Any(), gomock.Eq(promotion.ID)).
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
			AdminID:     admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeletePromotion(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/admin/v1/admins/%d/promotions/%d", tc.AdminID, tc.PromotionID)
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

func randomPSuperAdmin(t *testing.T) (admin db.Admin, password string) {
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

func randomPromotion() db.Promotion {
	return db.Promotion{
		ID:           util.RandomInt(1, 1000),
		Name:         util.RandomUser(),
		Description:  util.RandomUser(),
		DiscountRate: util.RandomInt(1, 1000),
		Active:       util.RandomBool(),
		StartDate:    time.Now().Truncate(0).Local().UTC(),
		EndDate:      time.Now().Truncate(0).Local().UTC(),
	}
}

func requireBodyMatchPromotion(t *testing.T, body io.ReadCloser, Promotion db.Promotion) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotPromotion db.Promotion
	err = json.Unmarshal(data, &gotPromotion)
	require.NoError(t, err)
	require.Equal(t, Promotion, gotPromotion)
}

func requireBodyMatchPromotions(t *testing.T, body io.ReadCloser, Promotions []db.Promotion) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotPromotions []db.Promotion
	err = json.Unmarshal(data, &gotPromotions)
	require.NoError(t, err)
	require.Equal(t, Promotions, gotPromotions)
}
