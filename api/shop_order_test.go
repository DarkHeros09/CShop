package api

import (
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
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func TestListShopOrderAPI(t *testing.T) {
	user, _ := randomSOUser(t)
	orderStatus := createRandomOrderStatus(t)
	// shopOrder := createRandomShopOrderForList(t, user, orderStatus)
	var shopOrders []db.ListShopOrdersByUserIDRow
	n := 5
	shopOrders = make([]db.ListShopOrdersByUserIDRow, n)
	for i := 0; i < n; i++ {
		shopOrders[i] = createRandomShopOrderForList(t, user, orderStatus)
	}

	type Query struct {
		pageID   int
		pageSize int
	}

	testCases := []struct {
		name          string
		query         Query
		UserID        int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(rsp *http.Response)
	}{
		{
			name: "OK",
			query: Query{
				pageID:   1,
				pageSize: n,
			},
			UserID: user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.ListShopOrdersByUserIDParams{
					Limit:  int32(n),
					Offset: 0,
					UserID: user.ID,
				}

				store.EXPECT().
					ListShopOrdersByUserID(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(shopOrders, nil)

			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchListShopOrder(t, rsp.Body, shopOrders)
			},
		},
		{
			name: "NoAuthorization",
			query: Query{
				pageID:   1,
				pageSize: n,
			},
			UserID: user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListShopOrdersByUserID(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name: "InternalError",
			query: Query{
				pageID:   1,
				pageSize: n,
			},
			UserID: user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.ListShopOrdersByUserIDParams{
					Limit:  int32(n),
					Offset: 0,
					UserID: user.ID,
				}

				store.EXPECT().
					ListShopOrdersByUserID(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return([]db.ListShopOrdersByUserIDRow{}, pgx.ErrTxClosed)

			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name: "InvalidUserID",
			query: Query{
				pageID:   1,
				pageSize: n,
			},
			UserID: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListShopOrdersByUserID(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/usr/v1/users/%d/shop-orders", tc.UserID)
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			// Add query parameters to request URL
			q := request.URL.Query()
			q.Add("page_id", fmt.Sprintf("%d", tc.query.pageID))
			q.Add("page_size", fmt.Sprintf("%d", tc.query.pageSize))
			request.URL.RawQuery = q.Encode()

			tc.setupAuth(t, request, server.tokenMaker)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
		})
	}
}

func randomSOUser(t *testing.T) (user db.User, password string) {
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

// func createRandomShopOrder(t *testing.T, user db.User) (shopOrder db.ShopOrder) {
// 	shopOrder = db.ShopOrder{
// 		ID:                util.RandomMoney(),
// 		UserID:            user.ID,
// 		PaymentMethodID:   util.RandomMoney(),
// 		ShippingAddressID: util.RandomMoney(),
// 		OrderTotal:        util.RandomDecimalString(1, 1000),
// 		ShippingMethodID:  util.RandomMoney(),
// 		OrderStatusID:     null.IntFrom(util.RandomMoney()),
// 	}
// 	return
// }

// func createRandomShopOrderForGet(t *testing.T, shopOrder db.ShopOrder) (ShopOrder db.GetShopOrderByUserIDOrderIDRow) {
// 	ShopOrder = db.GetShopOrderByUserIDOrderIDRow{
// 		ID:            util.RandomMoney(),
// 		ProductItemID: util.RandomMoney(),
// 		OrderID:       shopOrder.ID,
// 		Quantity:      int32(util.RandomMoney()),
// 		Price:         util.RandomDecimalString(1, 1000),
// 		UserID:        null.IntFrom(shopOrder.UserID),
// 	}
// 	return
// }

func requireBodyMatchListShopOrder(t *testing.T, body io.ReadCloser, shopOrder []db.ListShopOrdersByUserIDRow) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotShopOrder []db.ListShopOrdersByUserIDRow
	err = json.Unmarshal(data, &gotShopOrder)

	require.NoError(t, err)
	for i, gotOrderItem := range gotShopOrder {
		require.Equal(t, shopOrder[i].ID, gotOrderItem.ID)
		require.Equal(t, shopOrder[i].Status, gotOrderItem.Status)
		require.Equal(t, shopOrder[i].OrderNumber, gotOrderItem.OrderNumber)
		require.Equal(t, shopOrder[i].UserID, gotOrderItem.UserID)
		require.Equal(t, shopOrder[i].PaymentMethodID, gotOrderItem.PaymentMethodID)
		require.Equal(t, shopOrder[i].ShippingAddressID, gotOrderItem.ShippingAddressID)
		require.Equal(t, shopOrder[i].OrderTotal, gotOrderItem.OrderTotal)
		require.Equal(t, shopOrder[i].ShippingMethodID, gotOrderItem.ShippingMethodID)
		require.Equal(t, shopOrder[i].OrderStatusID, gotOrderItem.OrderStatusID)
	}
}

func createRandomShopOrderForList(t *testing.T, user db.User, orderStatus db.OrderStatus) (shopOrder db.ListShopOrdersByUserIDRow) {
	shopOrder = db.ListShopOrdersByUserIDRow{
		Status:            null.StringFrom(orderStatus.Status),
		ID:                util.RandomMoney(),
		TrackNumber:       util.RandomString(5),
		UserID:            user.ID,
		PaymentMethodID:   util.RandomMoney(),
		ShippingAddressID: util.RandomMoney(),
		OrderTotal:        fmt.Sprint(util.RandomMoney()),
		ShippingMethodID:  util.RandomMoney(),
		OrderStatusID:     null.IntFrom(orderStatus.ID),
	}
	return
}
