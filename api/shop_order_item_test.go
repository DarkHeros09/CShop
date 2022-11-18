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

func TestGetShopOrderItemAPI(t *testing.T) {
	user, _ := randomSOIUser(t)
	shopOrder := createRandomShopOrder(t, user)
	shopOrderItem := createRandomShopOrderItemForGet(t, shopOrder)

	testCases := []struct {
		name          string
		ShopOrderID   int64
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name:        "OK",
			ShopOrderID: shopOrder.ID,
			body: gin.H{
				"user_id": user.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.GetShopOrderItemByUserIDOrderIDParams{
					UserID:  user.ID,
					OrderID: shopOrder.ID,
				}

				store.EXPECT().
					GetShopOrderItemByUserIDOrderID(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(shopOrderItem, nil)

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchShopOrderItemForGet(t, recorder.Body, shopOrderItem)
			},
		},
		{
			name:        "NoAuthorization",
			ShopOrderID: shopOrder.ID,
			body: gin.H{
				"user_id": user.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetShopOrderItemByUserIDOrderID(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:        "InternalError",
			ShopOrderID: shopOrder.ID,
			body: gin.H{
				"user_id": user.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.GetShopOrderItemByUserIDOrderIDParams{
					UserID:  user.ID,
					OrderID: shopOrder.ID,
				}

				store.EXPECT().
					GetShopOrderItemByUserIDOrderID(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.GetShopOrderItemByUserIDOrderIDRow{}, pgx.ErrTxClosed)

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:        "InvalidUserID",
			ShopOrderID: shopOrder.ID,
			body: gin.H{
				"user_id": 0,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetShopOrderItemByUserIDOrderID(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/users/shop-order/%d", tc.ShopOrderID)
			request, err := http.NewRequest(http.MethodGet, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestListShopOrderItemAPI(t *testing.T) {
	user, _ := randomSOIUser(t)
	shopOrder := createRandomShopOrder(t, user)

	n := 5
	shopOrderItems := make([]db.ListShopOrderItemsByUserIDRow, n)
	for i := 0; i < n; i++ {
		shopOrderItems[i] = createRandomListShopOrderItem(t, shopOrder)
	}

	type Query struct {
		pageID   int
		pageSize int
	}

	testCases := []struct {
		name          string
		query         Query
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			query: Query{
				pageID:   1,
				pageSize: n,
			},
			body: gin.H{
				"user_id": user.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.ListShopOrderItemsByUserIDParams{
					Limit:  int32(n),
					Offset: 0,
					UserID: user.ID,
				}

				store.EXPECT().
					ListShopOrderItemsByUserID(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(shopOrderItems, nil)

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchListShopOrderItem(t, recorder.Body, shopOrderItems)
			},
		},
		{
			name: "NoAuthorization",
			query: Query{
				pageID:   1,
				pageSize: n,
			},
			body: gin.H{
				"user_id": user.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListShopOrderItemsByUserID(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InternalError",
			query: Query{
				pageID:   1,
				pageSize: n,
			},
			body: gin.H{
				"user_id": user.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.ListShopOrderItemsByUserIDParams{
					Limit:  int32(n),
					Offset: 0,
					UserID: user.ID,
				}

				store.EXPECT().
					ListShopOrderItemsByUserID(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return([]db.ListShopOrderItemsByUserIDRow{}, pgx.ErrTxClosed)

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidUserID",
			query: Query{
				pageID:   1,
				pageSize: n,
			},
			body: gin.H{
				"user_id": 0,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListShopOrderItemsByUserID(gomock.Any(), gomock.Any()).
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

			url := "/users/shop-order"
			request, err := http.NewRequest(http.MethodGet, url, bytes.NewReader(data))
			require.NoError(t, err)

			// Add query parameters to request URL
			q := request.URL.Query()
			q.Add("page_id", fmt.Sprintf("%d", tc.query.pageID))
			q.Add("page_size", fmt.Sprintf("%d", tc.query.pageSize))
			request.URL.RawQuery = q.Encode()

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func randomSOIUser(t *testing.T) (user db.User, password string) {
	password = util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	user = db.User{
		ID:        util.RandomMoney(),
		Username:  util.RandomUser(),
		Password:  hashedPassword,
		Telephone: int32(util.RandomInt(7, 7)),
		Email:     util.RandomEmail(),
	}
	return
}

func createRandomShopOrder(t *testing.T, user db.User) (shopOrder db.ShopOrder) {
	shopOrder = db.ShopOrder{
		ID:                util.RandomMoney(),
		UserID:            user.ID,
		PaymentMethodID:   util.RandomMoney(),
		ShippingAddressID: util.RandomMoney(),
		OrderTotal:        util.RandomDecimalString(1, 1000),
		ShippingMethodID:  util.RandomMoney(),
		OrderStatusID:     util.RandomMoney(),
	}
	return
}

func createRandomShopOrderItemForGet(t *testing.T, shopOrder db.ShopOrder) (ShopOrderItem db.GetShopOrderItemByUserIDOrderIDRow) {
	ShopOrderItem = db.GetShopOrderItemByUserIDOrderIDRow{
		ID:            util.RandomMoney(),
		ProductItemID: util.RandomMoney(),
		OrderID:       shopOrder.ID,
		Quantity:      int32(util.RandomMoney()),
		Price:         util.RandomDecimalString(1, 1000),
		UserID:        null.IntFrom(shopOrder.UserID),
	}
	return
}

func createRandomListShopOrderItem(t *testing.T, shopOrder db.ShopOrder) (ShopOrderItem db.ListShopOrderItemsByUserIDRow) {
	ShopOrderItem = db.ListShopOrderItemsByUserIDRow{
		UserID:        shopOrder.UserID,
		ID:            null.IntFrom(util.RandomMoney()),
		ProductItemID: null.IntFrom(util.RandomMoney()),
		OrderID:       null.IntFrom(shopOrder.ID),
		Quantity:      null.IntFrom(util.RandomMoney()),
		Price:         null.StringFrom(util.RandomDecimalString(1, 1000)),
	}
	return
}

func requireBodyMatchShopOrderItemForGet(t *testing.T, body *bytes.Buffer, ShopOrderItem db.GetShopOrderItemByUserIDOrderIDRow) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotShopOrderItem db.GetShopOrderItemByUserIDOrderIDRow
	err = json.Unmarshal(data, &gotShopOrderItem)

	require.NoError(t, err)
	require.Equal(t, ShopOrderItem.ID, gotShopOrderItem.ID)
	require.Equal(t, ShopOrderItem.OrderID, gotShopOrderItem.OrderID)
	require.Equal(t, ShopOrderItem.ProductItemID, gotShopOrderItem.ProductItemID)
	require.Equal(t, ShopOrderItem.Quantity, gotShopOrderItem.Quantity)
	require.Equal(t, ShopOrderItem.Price, gotShopOrderItem.Price)
	require.Equal(t, ShopOrderItem.CreatedAt, gotShopOrderItem.CreatedAt)
	require.Equal(t, ShopOrderItem.UserID, gotShopOrderItem.UserID)
}

func requireBodyMatchListShopOrderItem(t *testing.T, body *bytes.Buffer, shopOrderItem []db.ListShopOrderItemsByUserIDRow) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotShopOrderItem []db.ListShopOrderItemsByUserIDRow
	err = json.Unmarshal(data, &gotShopOrderItem)

	require.NoError(t, err)
	for i, gotOrderItem := range gotShopOrderItem {
		require.Equal(t, shopOrderItem[i].ID, gotOrderItem.ID)
		require.Equal(t, shopOrderItem[i].OrderID, gotOrderItem.OrderID)
		require.Equal(t, shopOrderItem[i].ProductItemID, gotOrderItem.ProductItemID)
		require.Equal(t, shopOrderItem[i].UserID, gotOrderItem.UserID)
		require.Equal(t, shopOrderItem[i].Price, gotOrderItem.Price)
		require.Equal(t, shopOrderItem[i].Quantity, gotOrderItem.Quantity)
	}
}
