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

func TestGetCategoryPromotionAPI(t *testing.T) {
	categoryPromotion := randomCategoryPromotion()

	testCases := []struct {
		name          string
		CategoryID    int64
		body          gin.H
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:       "OK",
			CategoryID: categoryPromotion.PromotionID,
			body: gin.H{
				"category_id": categoryPromotion.CategoryID,
			},
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchCategoryPromotion(t, recorder.Body, categoryPromotion)
			},
		},

		{
			name:       "NotFound",
			CategoryID: categoryPromotion.PromotionID,
			body: gin.H{
				"category_id": categoryPromotion.CategoryID,
			},
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:       "InternalError",
			CategoryID: categoryPromotion.PromotionID,
			body: gin.H{
				"category_id": categoryPromotion.CategoryID,
			},
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:       "InvalidID",
			CategoryID: 0,
			body: gin.H{
				"category_id": categoryPromotion.CategoryID,
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetCategoryPromotion(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/category-promotions/%d", tc.CategoryID)
			request, err := http.NewRequest(http.MethodGet, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			//check response
			tc.checkResponse(t, recorder)
		})

	}

}

func TestCreateCategoryPromotionAPI(t *testing.T) {
	admin, _ := randomCategoryPromotionSuperAdmin(t)
	categoryPromotion := randomCategoryPromotion()

	testCases := []struct {
		name          string
		body          gin.H
		CategoryID    int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"Category_id":  categoryPromotion.CategoryID,
				"promotion_id": categoryPromotion.PromotionID,
				"active":       categoryPromotion.Active,
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
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchCategoryPromotion(t, recorder.Body, categoryPromotion)
			},
		},
		{
			name:       "NoAuthorization",
			CategoryID: categoryPromotion.CategoryID,
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
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:       "Unauthorized",
			CategoryID: categoryPromotion.CategoryID,
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
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"Category_id":  categoryPromotion.CategoryID,
				"promotion_id": categoryPromotion.PromotionID,
				"active":       categoryPromotion.Active,
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
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidID",
			body: gin.H{
				"Category_id":  "invalid",
				"promotion_id": categoryPromotion.PromotionID,
				"active":       categoryPromotion.Active,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateCategoryPromotion(gomock.Any(), gomock.Any()).
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

			url := "/category-promotions"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
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
		checkResponse func(recoder *httptest.ResponseRecorder)
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
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchCategoryPromotions(t, recorder.Body, categoryPromotions)
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
					ListCategoryPromotions(gomock.Any(), gomock.Any()).
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
					ListCategoryPromotions(gomock.Any(), gomock.Any()).
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

			url := "/category-promotions"
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

func TestUpdateCategoryPromotionAPI(t *testing.T) {
	admin, _ := randomCategoryPromotionSuperAdmin(t)
	categoryPromotion := randomCategoryPromotion()

	testCases := []struct {
		name          string
		body          gin.H
		CategoryID    int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:       "OK",
			CategoryID: categoryPromotion.CategoryID,
			body: gin.H{
				"Category_id":  categoryPromotion.CategoryID,
				"promotion_id": categoryPromotion.PromotionID,
				"active":       categoryPromotion.Active,
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:       "Unauthorized",
			CategoryID: categoryPromotion.CategoryID,
			body: gin.H{
				"Category_id":  categoryPromotion.CategoryID,
				"promotion_id": categoryPromotion.PromotionID,
				"active":       categoryPromotion.Active,
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:       "NoAuthorization",
			CategoryID: categoryPromotion.CategoryID,
			body: gin.H{
				"Category_id":  categoryPromotion.CategoryID,
				"promotion_id": categoryPromotion.PromotionID,
				"active":       categoryPromotion.Active,
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:       "InternalError",
			CategoryID: categoryPromotion.CategoryID,
			body: gin.H{
				"Category_id":  categoryPromotion.CategoryID,
				"promotion_id": categoryPromotion.PromotionID,
				"active":       categoryPromotion.Active,
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:       "InvalidID",
			CategoryID: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateCategoryPromotion(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/category-promotions/%d", tc.CategoryID)
			request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestDeleteCategoryPromotionAPI(t *testing.T) {
	admin, _ := randomCategoryPromotionSuperAdmin(t)
	categoryPromotion := randomCategoryPromotion()

	testCases := []struct {
		name          string
		body          gin.H
		PromotionID   int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:        "OK",
			PromotionID: categoryPromotion.PromotionID,
			body: gin.H{
				"category_id": categoryPromotion.CategoryID,
			},
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:        "Unauthorized",
			PromotionID: categoryPromotion.PromotionID,
			body: gin.H{
				"category_id": categoryPromotion.CategoryID,
			},
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:        "NoAuthorization",
			PromotionID: categoryPromotion.PromotionID,
			body: gin.H{
				"category_id": categoryPromotion.CategoryID,
			},
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:        "NotFound",
			PromotionID: categoryPromotion.PromotionID,
			body: gin.H{
				"category_id": categoryPromotion.CategoryID,
			},
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:        "InternalError",
			PromotionID: categoryPromotion.PromotionID,
			body: gin.H{
				"category_id": categoryPromotion.CategoryID,
			},
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:        "InvalidID",
			PromotionID: 0,
			body: gin.H{
				"category_id": 0,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteCategoryPromotion(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/category-promotions/%d", tc.PromotionID)
			request, err := http.NewRequest(http.MethodDelete, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			//check response
			tc.checkResponse(t, recorder)
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

func requireBodyMatchCategoryPromotion(t *testing.T, body *bytes.Buffer, categoryPromotion db.CategoryPromotion) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotCategoryPromotion db.CategoryPromotion
	err = json.Unmarshal(data, &gotCategoryPromotion)
	require.NoError(t, err)
	require.Equal(t, categoryPromotion, gotCategoryPromotion)
}

func requireBodyMatchCategoryPromotions(t *testing.T, body *bytes.Buffer, CategoryPromotions []db.CategoryPromotion) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotCategoryPromotions []db.CategoryPromotion
	err = json.Unmarshal(data, &gotCategoryPromotions)
	require.NoError(t, err)
	require.Equal(t, CategoryPromotions, gotCategoryPromotions)
}
