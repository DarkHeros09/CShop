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

func TestGetShopOrderItemAPI(t *testing.T) {
	user, _ := randomSOIUser(t)
	shopOrder := createRandomShopOrder(t, user)
	var shopOrderItemsList []db.ListShopOrderItemsByUserIDOrderIDRow
	for i := 0; i < 5; i++ {
		shopOrderItems := createRandomShopOrderItem(t, shopOrder)
		shopOrderItemsList = append(shopOrderItemsList, shopOrderItems)
	}
	// shopOrderItem := createRandomShopOrderItemForGet(t, shopOrder)

	testCases := []struct {
		name          string
		ShopOrderID   int64
		UserID        int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(rsp *http.Response)
	}{
		{
			name:        "OK",
			ShopOrderID: shopOrder.ID,
			UserID:      user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.ListShopOrderItemsByUserIDOrderIDParams{
					UserID:  user.ID,
					OrderID: shopOrder.ID,
				}

				store.EXPECT().
					ListShopOrderItemsByUserIDOrderID(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(shopOrderItemsList, nil)

			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchShopOrderItemForList(t, rsp.Body, shopOrderItemsList)
			},
		},
		{
			name:        "NoAuthorization",
			ShopOrderID: shopOrder.ID,
			UserID:      user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListShopOrderItemsByUserIDOrderID(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:        "InternalError",
			ShopOrderID: shopOrder.ID,
			UserID:      user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListShopOrderItemsByUserIDOrderIDParams{
					UserID:  user.ID,
					OrderID: shopOrder.ID,
				}

				store.EXPECT().
					ListShopOrderItemsByUserIDOrderID(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return([]db.ListShopOrderItemsByUserIDOrderIDRow{}, pgx.ErrTxClosed)

			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:        "InvalidUserID",
			ShopOrderID: shopOrder.ID,
			UserID:      0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListShopOrderItemsByUserIDOrderID(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/usr/v1/users/%d/shop-order-items/%d", tc.UserID, tc.ShopOrderID)
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.userTokenMaker)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
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
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchListShopOrderItem(t, rsp.Body, shopOrderItems)
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
					ListShopOrderItemsByUserID(gomock.Any(), gomock.Any()).
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
					ListShopOrderItemsByUserID(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/usr/v1/users/%d/shop-order-items", tc.UserID)
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			// Add query parameters to request URL
			q := request.URL.Query()
			q.Add("page_id", fmt.Sprintf("%d", tc.query.pageID))
			q.Add("page_size", fmt.Sprintf("%d", tc.query.pageSize))
			request.URL.RawQuery = q.Encode()

			tc.setupAuth(t, request, server.userTokenMaker)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
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
		ID:     util.RandomMoney(),
		UserID: user.ID,
		// PaymentMethodID:   util.RandomMoney(),
		ShippingAddressID: util.RandomMoney(),
		OrderTotal:        util.RandomDecimalString(1, 1000),
		ShippingMethodID:  util.RandomMoney(),
		OrderStatusID:     null.IntFrom(util.RandomMoney()),
	}
	return
}

func createRandomShopOrderItem(t *testing.T, shopOrder db.ShopOrder) (shopOrderItems db.ListShopOrderItemsByUserIDOrderIDRow) {
	shopOrderItems = db.ListShopOrderItemsByUserIDOrderIDRow{
		Status:        null.StringFrom(util.RandomUser()),
		ID:            util.RandomMoney(),
		ProductItemID: util.RandomMoney(),
		OrderID:       util.RandomMoney(),
		Quantity:      int32(util.RandomMoney()),
		Price:         fmt.Sprint(int32(util.RandomMoney())),
		ProductName:   null.StringFrom(util.RandomUser()),
		ProductImage:  null.StringFrom(util.RandomURL()),
		ProductActive: null.BoolFrom(util.RandomBool()),
		AddressLine:   null.StringFrom(util.RandomUser()),
		Region:        null.StringFrom(util.RandomUser()),
		City:          null.StringFrom(util.RandomUser()),
		// PaymentType:   null.StringFrom(util.RandomUser()),
	}
	return
}

func createRandomShopOrderItemForGet(t *testing.T, shopOrder db.ShopOrder) (ShopOrderItem db.ShopOrderItem) {
	ShopOrderItem = db.ShopOrderItem{
		ID:            util.RandomMoney(),
		ProductItemID: util.RandomMoney(),
		OrderID:       shopOrder.ID,
		Quantity:      int32(util.RandomMoney()),
		Price:         util.RandomDecimalString(1, 1000),
		// UserID:        null.IntFrom(shopOrder.UserID),
	}
	return
}

func createRandomListShopOrderItem(t *testing.T, shopOrder db.ShopOrder) (ShopOrderItem db.ListShopOrderItemsByUserIDRow) {
	ShopOrderItem = db.ListShopOrderItemsByUserIDRow{
		UserID:        shopOrder.UserID,
		ID:            util.RandomMoney(),
		ProductItemID: null.IntFrom(util.RandomMoney()),
		OrderID:       null.IntFrom(shopOrder.ID),
		Quantity:      null.IntFrom(util.RandomMoney()),
		Price:         null.StringFrom(util.RandomDecimalString(1, 1000)),
	}
	return
}

func requireBodyMatchShopOrderItemForList(t *testing.T, body io.ReadCloser, shopOrderItem []db.ListShopOrderItemsByUserIDOrderIDRow) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotShopOrderItem []db.ListShopOrderItemsByUserIDOrderIDRow
	err = json.Unmarshal(data, &gotShopOrderItem)

	require.NoError(t, err)
	for i := range gotShopOrderItem {
		require.Equal(t, shopOrderItem[i].ID, gotShopOrderItem[i].ID)
		require.Equal(t, shopOrderItem[i].OrderID, gotShopOrderItem[i].OrderID)
		require.Equal(t, shopOrderItem[i].ProductItemID, gotShopOrderItem[i].ProductItemID)
		require.Equal(t, shopOrderItem[i].Quantity, gotShopOrderItem[i].Quantity)
		require.Equal(t, shopOrderItem[i].Price, gotShopOrderItem[i].Price)
		require.Equal(t, shopOrderItem[i].CreatedAt, gotShopOrderItem[i].CreatedAt)
	}
	// require.Equal(t, ShopOrderItem.UserID, gotShopOrderItem.UserID)
}

func requireBodyMatchListShopOrderItem(t *testing.T, body io.ReadCloser, shopOrderItem []db.ListShopOrderItemsByUserIDRow) {
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
