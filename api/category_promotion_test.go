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

func TestGetCategoryPromotionAPI(t *testing.T) {
	categoryPromotion := randomCategoryPromotion()

	testCases := []struct {
		name          string
		CategoryID    int64
		PromotionID   int64
		AdminID       int64
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rsp *http.Response)
	}{
		{
			name:        "OK",
			PromotionID: categoryPromotion.PromotionID,
			CategoryID:  categoryPromotion.CategoryID,
			buildStub: func(store *mockdb.MockStore) {
				arg := db.GetCategoryPromotionParams{
					CategoryID:  categoryPromotion.CategoryID,
					PromotionID: categoryPromotion.PromotionID,
				}
				store.EXPECT().
					GetCategoryPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(categoryPromotion, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchCategoryPromotion(t, rsp.Body, categoryPromotion)
			},
		},

		{
			name:        "NotFound",
			PromotionID: categoryPromotion.PromotionID,
			CategoryID:  categoryPromotion.CategoryID,
			buildStub: func(store *mockdb.MockStore) {
				arg := db.GetCategoryPromotionParams{
					CategoryID:  categoryPromotion.CategoryID,
					PromotionID: categoryPromotion.PromotionID,
				}
				store.EXPECT().
					GetCategoryPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.CategoryPromotion{}, pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
			},
		},
		{
			name:        "InternalError",
			PromotionID: categoryPromotion.PromotionID,
			CategoryID:  categoryPromotion.CategoryID,
			buildStub: func(store *mockdb.MockStore) {
				arg := db.GetCategoryPromotionParams{
					CategoryID:  categoryPromotion.CategoryID,
					PromotionID: categoryPromotion.PromotionID,
				}
				store.EXPECT().
					GetCategoryPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.CategoryPromotion{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:        "InvalidID",
			PromotionID: 0,
			CategoryID:  categoryPromotion.CategoryID,
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetCategoryPromotion(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/api/v1/category-promotions/%d/categories/%d", tc.PromotionID, tc.CategoryID)
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(t, rsp)
		})

	}

}

func TestCreateCategoryPromotionAPI(t *testing.T) {
	admin, _ := randomCategoryPromotionSuperAdmin(t)
	categoryPromotion := randomCategoryPromotion()

	testCases := []struct {
		name          string
		body          fiber.Map
		CategoryID    int64
		PromotionID   int64
		AdminID       int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(rsp *http.Response)
	}{
		{
			name:        "OK",
			CategoryID:  categoryPromotion.CategoryID,
			PromotionID: categoryPromotion.PromotionID,
			AdminID:     admin.ID,
			body: fiber.Map{
				"active": categoryPromotion.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateCategoryPromotionParams{
					CategoryID:  categoryPromotion.CategoryID,
					PromotionID: categoryPromotion.PromotionID,
					Active:      categoryPromotion.Active,
				}

				store.EXPECT().
					CreateCategoryPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(categoryPromotion, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchCategoryPromotion(t, rsp.Body, categoryPromotion)
			},
		},
		{
			name:        "NoAuthorization",
			CategoryID:  categoryPromotion.CategoryID,
			PromotionID: categoryPromotion.PromotionID,
			AdminID:     admin.ID,
			body: fiber.Map{
				"active": categoryPromotion.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateCategoryPromotionParams{
					CategoryID:  categoryPromotion.CategoryID,
					PromotionID: categoryPromotion.PromotionID,
					Active:      categoryPromotion.Active,
				}

				store.EXPECT().
					CreateCategoryPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(0)

			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:        "Unauthorized",
			CategoryID:  categoryPromotion.CategoryID,
			PromotionID: categoryPromotion.PromotionID,
			AdminID:     admin.ID,
			body: fiber.Map{
				"active": categoryPromotion.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, false, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateCategoryPromotionParams{
					CategoryID:  categoryPromotion.CategoryID,
					PromotionID: categoryPromotion.PromotionID,
					Active:      categoryPromotion.Active,
				}

				store.EXPECT().
					CreateCategoryPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:        "InternalError",
			CategoryID:  categoryPromotion.CategoryID,
			PromotionID: categoryPromotion.PromotionID,
			AdminID:     admin.ID,
			body: fiber.Map{
				"active": categoryPromotion.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateCategoryPromotion(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.CategoryPromotion{}, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:        "InvalidID",
			CategoryID:  0,
			PromotionID: categoryPromotion.PromotionID,
			AdminID:     admin.ID,
			body: fiber.Map{
				"active": categoryPromotion.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateCategoryPromotion(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/admin/v1/admins/%d/category-promotions/%d/categories/%d", tc.AdminID, tc.PromotionID, tc.CategoryID)
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

func TestListCategoryPromotionsAPI(t *testing.T) {
	n := 5
	categoryPromotions := make([]db.CategoryPromotion, n)
	for i := 0; i < n; i++ {
		categoryPromotions[i] = randomCategoryPromotion()
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
				arg := db.ListCategoryPromotionsParams{
					Limit:  int32(n),
					Offset: 0,
				}

				store.EXPECT().
					ListCategoryPromotions(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(categoryPromotions, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchCategoryPromotions(t, rsp.Body, categoryPromotions)
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
					ListCategoryPromotions(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.CategoryPromotion{}, pgx.ErrTxClosed)
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
					ListCategoryPromotions(gomock.Any(), gomock.Any()).
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
					ListCategoryPromotions(gomock.Any(), gomock.Any()).
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

			url := "/api/v1/category-promotions"
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

func TestUpdateCategoryPromotionAPI(t *testing.T) {
	admin, _ := randomCategoryPromotionSuperAdmin(t)
	categoryPromotion := randomCategoryPromotion()

	testCases := []struct {
		name          string
		body          fiber.Map
		CategoryID    int64
		PromotionID   int64
		AdminID       int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rsp *http.Response)
	}{
		{
			name:        "OK",
			CategoryID:  categoryPromotion.CategoryID,
			PromotionID: categoryPromotion.PromotionID,
			AdminID:     admin.ID,
			body: fiber.Map{
				"active": categoryPromotion.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateCategoryPromotionParams{
					PromotionID: categoryPromotion.PromotionID,
					Active:      null.BoolFromPtr(&categoryPromotion.Active),
					CategoryID:  categoryPromotion.CategoryID,
				}

				store.EXPECT().
					UpdateCategoryPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(categoryPromotion, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:        "Unauthorized",
			CategoryID:  categoryPromotion.CategoryID,
			PromotionID: categoryPromotion.PromotionID,
			AdminID:     admin.ID,
			body: fiber.Map{
				"active": categoryPromotion.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, false, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateCategoryPromotionParams{
					PromotionID: categoryPromotion.PromotionID,
					Active:      null.BoolFromPtr(&categoryPromotion.Active),
					CategoryID:  categoryPromotion.CategoryID,
				}

				store.EXPECT().
					UpdateCategoryPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:        "NoAuthorization",
			CategoryID:  categoryPromotion.CategoryID,
			PromotionID: categoryPromotion.PromotionID,
			AdminID:     admin.ID,
			body: fiber.Map{
				"active": categoryPromotion.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateCategoryPromotionParams{
					PromotionID: categoryPromotion.PromotionID,
					Active:      null.BoolFromPtr(&categoryPromotion.Active),
					CategoryID:  categoryPromotion.CategoryID,
				}

				store.EXPECT().
					UpdateCategoryPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:        "InternalError",
			CategoryID:  categoryPromotion.CategoryID,
			PromotionID: categoryPromotion.PromotionID,
			AdminID:     admin.ID,
			body: fiber.Map{
				"active": categoryPromotion.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateCategoryPromotionParams{
					PromotionID: categoryPromotion.PromotionID,
					Active:      null.BoolFromPtr(&categoryPromotion.Active),
					CategoryID:  categoryPromotion.CategoryID,
				}
				store.EXPECT().
					UpdateCategoryPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.CategoryPromotion{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:        "InvalidID",
			CategoryID:  0,
			PromotionID: categoryPromotion.PromotionID,
			AdminID:     admin.ID,
			body: fiber.Map{
				"active": categoryPromotion.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateCategoryPromotion(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/admin/v1/admins/%d/category-promotions/%d/categories/%d", tc.AdminID, tc.PromotionID, tc.CategoryID)
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

func TestDeleteCategoryPromotionAPI(t *testing.T) {
	admin, _ := randomCategoryPromotionSuperAdmin(t)
	categoryPromotion := randomCategoryPromotion()

	testCases := []struct {
		name          string
		PromotionID   int64
		CategoryID    int64
		AdminID       int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rsp *http.Response)
	}{
		{
			name:        "OK",
			PromotionID: categoryPromotion.PromotionID,
			CategoryID:  categoryPromotion.CategoryID,
			AdminID:     admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteCategoryPromotionParams{
					CategoryID:  categoryPromotion.CategoryID,
					PromotionID: categoryPromotion.PromotionID,
				}
				store.EXPECT().
					DeleteCategoryPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:        "Unauthorized",
			PromotionID: categoryPromotion.PromotionID,
			CategoryID:  categoryPromotion.CategoryID,
			AdminID:     admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, 2, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteCategoryPromotionParams{
					CategoryID:  categoryPromotion.CategoryID,
					PromotionID: categoryPromotion.PromotionID,
				}
				store.EXPECT().
					DeleteCategoryPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:        "NoAuthorization",
			PromotionID: categoryPromotion.PromotionID,
			CategoryID:  categoryPromotion.CategoryID,
			AdminID:     admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteCategoryPromotionParams{
					CategoryID:  categoryPromotion.CategoryID,
					PromotionID: categoryPromotion.PromotionID,
				}
				store.EXPECT().
					DeleteCategoryPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:        "NotFound",
			PromotionID: categoryPromotion.PromotionID,
			CategoryID:  categoryPromotion.CategoryID,
			AdminID:     admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteCategoryPromotionParams{
					CategoryID:  categoryPromotion.CategoryID,
					PromotionID: categoryPromotion.PromotionID,
				}
				store.EXPECT().
					DeleteCategoryPromotion(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
			},
		},
		{
			name:        "InternalError",
			PromotionID: categoryPromotion.PromotionID,
			CategoryID:  categoryPromotion.CategoryID,
			AdminID:     admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteCategoryPromotionParams{
					CategoryID:  categoryPromotion.CategoryID,
					PromotionID: categoryPromotion.PromotionID,
				}
				store.EXPECT().
					DeleteCategoryPromotion(gomock.Any(), gomock.Eq(arg)).
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
			CategoryID:  categoryPromotion.CategoryID,
			AdminID:     admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteCategoryPromotion(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/admin/v1/admins/%d/category-promotions/%d/categories/%d", tc.AdminID, tc.PromotionID, tc.CategoryID)
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

func randomCategoryPromotionSuperAdmin(t *testing.T) (admin db.Admin, password string) {
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

func randomCategoryPromotion() db.CategoryPromotion {
	return db.CategoryPromotion{
		CategoryID:  util.RandomMoney(),
		PromotionID: util.RandomMoney(),
		Active:      util.RandomBool(),
	}
}

func requireBodyMatchCategoryPromotion(t *testing.T, body io.ReadCloser, categoryPromotion db.CategoryPromotion) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotCategoryPromotion db.CategoryPromotion
	err = json.Unmarshal(data, &gotCategoryPromotion)
	require.NoError(t, err)
	require.Equal(t, categoryPromotion, gotCategoryPromotion)
}

func requireBodyMatchCategoryPromotions(t *testing.T, body io.ReadCloser, CategoryPromotions []db.CategoryPromotion) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotCategoryPromotions []db.CategoryPromotion
	err = json.Unmarshal(data, &gotCategoryPromotions)
	require.NoError(t, err)
	require.Equal(t, CategoryPromotions, gotCategoryPromotions)
}
