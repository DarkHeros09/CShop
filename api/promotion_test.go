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

func TestGetPromotionAPI(t *testing.T) {
	promotion := randomPromotion()

	testCases := []struct {
		name          string
		PromotionID   int64
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchPromotion(t, recorder.Body, promotion)
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
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

			url := fmt.Sprintf("/promotions/%d", tc.PromotionID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			//check response
			tc.checkResponse(t, recorder)
		})

	}

}

func TestCreatePromotionAPI(t *testing.T) {
	admin, _ := randomPSuperAdmin(t)
	promotion := randomPromotion()

	testCases := []struct {
		name          string
		body          gin.H
		PromotionID   int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
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
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchPromotion(t, recorder.Body, promotion)
			},
		},
		{
			name:        "NoAuthorization",
			PromotionID: promotion.ID,
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
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:        "Unauthorized",
			PromotionID: promotion.ID,
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
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
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
					CreatePromotion(gomock.Any(), gomock.Any()).
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

			url := "/promotions"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
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
		checkResponse func(recoder *httptest.ResponseRecorder)
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
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchPromotions(t, recorder.Body, Promotions)
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
					ListPromotions(gomock.Any(), gomock.Any()).
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
					ListPromotions(gomock.Any(), gomock.Any()).
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

			url := "/promotions"
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

func TestUpdatePromotionAPI(t *testing.T) {
	admin, _ := randomPSuperAdmin(t)
	promotion := randomPromotion()

	testCases := []struct {
		name          string
		body          gin.H
		PromotionID   int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:        "OK",
			PromotionID: promotion.ID,
			body: gin.H{
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:        "Unauthorized",
			PromotionID: promotion.ID,
			body: gin.H{
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:        "NoAuthorization",
			PromotionID: promotion.ID,
			body: gin.H{
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:        "InternalError",
			PromotionID: promotion.ID,
			body: gin.H{
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:        "InvalidID",
			PromotionID: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdatePromotion(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/promotions/%d", tc.PromotionID)
			request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestDeletePromotionAPI(t *testing.T) {
	admin, _ := randomPSuperAdmin(t)
	promotion := randomPromotion()

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
			PromotionID: promotion.ID,
			body: gin.H{
				"id": promotion.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeletePromotion(gomock.Any(), gomock.Eq(promotion.ID)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:        "Unauthorized",
			PromotionID: promotion.ID,
			body: gin.H{
				"id": promotion.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, 2, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeletePromotion(gomock.Any(), gomock.Eq(promotion.ID)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:        "NoAuthorization",
			PromotionID: promotion.ID,
			body: gin.H{
				"id": promotion.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeletePromotion(gomock.Any(), gomock.Eq(promotion.ID)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:        "NotFound",
			PromotionID: promotion.ID,
			body: gin.H{
				"id": promotion.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeletePromotion(gomock.Any(), gomock.Eq(promotion.ID)).
					Times(1).
					Return(pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:        "InternalError",
			PromotionID: promotion.ID,
			body: gin.H{
				"id": promotion.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeletePromotion(gomock.Any(), gomock.Eq(promotion.ID)).
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
				"id": 0,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeletePromotion(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/promotions/%d", tc.PromotionID)
			request, err := http.NewRequest(http.MethodDelete, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			//check response
			tc.checkResponse(t, recorder)
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
		StartDate:    time.Now().Truncate(0),
		EndDate:      time.Now().Truncate(0),
	}
}

func requireBodyMatchPromotion(t *testing.T, body *bytes.Buffer, Promotion db.Promotion) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotPromotion db.Promotion
	err = json.Unmarshal(data, &gotPromotion)
	require.NoError(t, err)
	require.Equal(t, Promotion, gotPromotion)
}

func requireBodyMatchPromotions(t *testing.T, body *bytes.Buffer, Promotions []db.Promotion) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotPromotions []db.Promotion
	err = json.Unmarshal(data, &gotPromotions)
	require.NoError(t, err)
	require.Equal(t, Promotions, gotPromotions)
}
