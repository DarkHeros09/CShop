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
	"github.com/cshop/v3/token"
	"github.com/cshop/v3/util"
	"github.com/gofiber/fiber/v2"
	"github.com/guregu/null"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCreateShoppingCartItemAPI(t *testing.T) {
	user, _ := randomSCIUser(t)
	shoppingCart := createRandomShoppingCart(t, user)
	shoppingCartItem := createRandomShoppingCartItem1(t, shoppingCart)

	testCases := []struct {
		name           string
		body           fiber.Map
		UserID         int64
		ShoppingCartID int64
		setupAuth      func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs     func(store *mockdb.MockStore)
		checkResponse  func(rsp *http.Response)
	}{
		{
			name:           "OK",
			UserID:         user.ID,
			ShoppingCartID: shoppingCart.ID,
			body: fiber.Map{
				"product_item_id": shoppingCartItem.ProductItemID,
				"qty":             shoppingCartItem.Qty,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.CreateShoppingCartItemParams{
					ShoppingCartID: shoppingCartItem.ShoppingCartID,
					ProductItemID:  shoppingCartItem.ProductItemID,
					Qty:            shoppingCartItem.Qty,
				}

				store.EXPECT().
					CreateShoppingCartItem(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(shoppingCartItem, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchShoppingCartItem(t, rsp.Body, shoppingCartItem)
			},
		},
		{
			name:           "NoAuthorization",
			UserID:         user.ID,
			ShoppingCartID: shoppingCart.ID,
			body: fiber.Map{

				"product_item_id": shoppingCartItem.ProductItemID,
				"qty":             shoppingCartItem.Qty,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					CreateShoppingCartItem(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:           "InternalError",
			UserID:         user.ID,
			ShoppingCartID: shoppingCart.ID,
			body: fiber.Map{

				"product_item_id": shoppingCartItem.ProductItemID,
				"qty":             shoppingCartItem.Qty,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.CreateShoppingCartItemParams{
					ShoppingCartID: shoppingCartItem.ShoppingCartID,
					ProductItemID:  shoppingCartItem.ProductItemID,
					Qty:            shoppingCartItem.Qty,
				}

				store.EXPECT().
					CreateShoppingCartItem(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.ShoppingCartItem{}, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:           "InvalidUserID",
			UserID:         0,
			ShoppingCartID: shoppingCart.ID,
			body: fiber.Map{

				"product_item_id": shoppingCartItem.ProductItemID,
				"qty":             shoppingCartItem.Qty,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					CreateShoppingCartItem(gomock.Any(), gomock.Any()).
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

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/usr/v1/users/%d/carts/%d/items", tc.UserID, tc.ShoppingCartID)
			request, err := http.NewRequest(fiber.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
		})
	}
}

func TestGetShoppingCartItemAPI(t *testing.T) {
	user, _ := randomSCIUser(t)
	shoppingCart := createRandomShoppingCart(t, user)
	shoppingCartItem := createRandomShoppingCartItemForGet(t, shoppingCart)

	testCases := []struct {
		name           string
		UserID         int64
		ShoppingCartID int64
		setupAuth      func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs     func(store *mockdb.MockStore)
		checkResponse  func(rsp *http.Response)
	}{
		{
			name:           "OK",
			ShoppingCartID: shoppingCart.ID,
			UserID:         user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.GetShoppingCartItemByUserIDCartIDParams{
					UserID: user.ID,
					ID:     shoppingCart.ID,
				}

				store.EXPECT().
					GetShoppingCartItemByUserIDCartID(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(shoppingCartItem, nil)

			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchShoppingCartItemForGet(t, rsp.Body, shoppingCartItem)
			},
		},
		{
			name:           "NoAuthorization",
			ShoppingCartID: shoppingCart.ID,
			UserID:         user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetShoppingCartItemByUserIDCartID(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:           "InternalError",
			ShoppingCartID: shoppingCart.ID,
			UserID:         user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.GetShoppingCartItemByUserIDCartIDParams{
					UserID: user.ID,
					ID:     shoppingCart.ID,
				}

				store.EXPECT().
					GetShoppingCartItemByUserIDCartID(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return([]db.GetShoppingCartItemByUserIDCartIDRow{}, pgx.ErrTxClosed)

			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:           "InvalidUserID",
			ShoppingCartID: shoppingCart.ID,
			UserID:         0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetShoppingCartItemByUserIDCartID(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/usr/v1/users/%d/carts/%d/items", tc.UserID, tc.ShoppingCartID)
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
		})
	}
}

func TestListShoppingCartItemAPI(t *testing.T) {
	user, _ := randomSCIUser(t)
	shoppingCart := createRandomShoppingCart(t, user)

	n := 5
	shoppingCartItems := make([]db.ListShoppingCartItemsByUserIDRow, n)
	// productItems := make([]db.ProductItem, n)
	var productItems []db.ListProductItemsV2Row
	for i := 0; i < n; i++ {
		randomProductItems := randomProductItemNew()
		productItems = append(productItems, randomProductItems)
		shoppingCartItems[i] = createRandomListShoppingCartItem(t, shoppingCart, randomProductItems)
	}

	// fmt.Println("ShoppingCarts: ", shoppingCartItems)
	// fmt.Println("ShoppingCarts LEN: ", len(shoppingCartItems))
	// fmt.Println("ProductItems: ", productItems)
	// fmt.Println("ProductItems LEN: ", len(productItems))

	productIds := createProductIdsForCart(t, shoppingCartItems)
	// fmt.Println("ProductIDs: ", productIds)
	listProductsByIds := createRandomListProductsByIds(t, shoppingCartItems, productItems)
	// fmt.Println("ProductsListById: ", listProductsByIds)

	finalRsp := createFinalRspForCart(t, shoppingCartItems, listProductsByIds)

	testCases := []struct {
		name          string
		UserID        int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(rsp *http.Response)
	}{
		{
			name:   "OK",
			UserID: user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					ListShoppingCartItemsByUserID(gomock.Any(), gomock.Eq(shoppingCart.UserID)).
					Times(1).
					Return(shoppingCartItems, nil)

				store.EXPECT().
					ListProductItemsByIDs(gomock.Any(), gomock.Eq(productIds)).
					Times(1).
					Return(listProductsByIds, nil)

			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchFinalListShoppingCartItem(t, rsp.Body, finalRsp)
			},
		},
		{
			name:   "NoAuthorization",
			UserID: user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListShoppingCartItemsByUserID(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:   "InternalError",
			UserID: user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					ListShoppingCartItemsByUserID(gomock.Any(), gomock.Eq(shoppingCart.UserID)).
					Times(1).
					Return([]db.ListShoppingCartItemsByUserIDRow{}, pgx.ErrTxClosed)

			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:   "InvalidUserID",
			UserID: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListShoppingCartItemsByUserID(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/usr/v1/users/%d/carts/items", tc.UserID)
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
		})
	}
}

func TestUpdateShoppingCartItemAPI(t *testing.T) {
	user, _ := randomSCIUser(t)
	shoppingCart := createRandomShoppingCart(t, user)
	shoppingCartItem := createRandomShoppingCartItemForUpdate(t, shoppingCart)
	qty := util.RandomMoney()

	testCases := []struct {
		name               string
		ShoppingCartID     int64
		UserID             int64
		ShoppingCartItemID int64
		body               fiber.Map
		setupAuth          func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs         func(store *mockdb.MockStore)
		checkResponse      func(t *testing.T, rsp *http.Response)
	}{
		{
			name:               "OK",
			ShoppingCartID:     shoppingCart.ID,
			UserID:             user.ID,
			ShoppingCartItemID: shoppingCartItem.ID,
			body: fiber.Map{
				"product_item_id": shoppingCartItem.ProductItemID,
				"qty":             qty,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.UpdateShoppingCartItemParams{
					ProductItemID:  null.IntFrom(shoppingCartItem.ProductItemID),
					Qty:            null.IntFrom(qty),
					ID:             shoppingCartItem.ID,
					ShoppingCartID: shoppingCartItem.ShoppingCartID,
				}

				store.EXPECT().
					UpdateShoppingCartItem(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(shoppingCartItem, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:               "NoAuthorization",
			ShoppingCartID:     shoppingCart.ID,
			UserID:             user.ID,
			ShoppingCartItemID: shoppingCartItem.ID,
			body: fiber.Map{
				"product_item_id": shoppingCartItem.ProductItemID,
				"qty":             qty,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateShoppingCartItem(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:               "InternalError",
			ShoppingCartID:     shoppingCart.ID,
			UserID:             user.ID,
			ShoppingCartItemID: shoppingCartItem.ID,
			body: fiber.Map{
				"product_item_id": shoppingCartItem.ProductItemID,
				"qty":             qty,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateShoppingCartItemParams{
					ProductItemID:  null.IntFrom(shoppingCartItem.ProductItemID),
					Qty:            null.IntFrom(qty),
					ID:             shoppingCartItem.ID,
					ShoppingCartID: shoppingCartItem.ShoppingCartID,
				}

				store.EXPECT().
					UpdateShoppingCartItem(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.UpdateShoppingCartItemRow{}, pgx.ErrTxClosed)

			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:           "InvalidCartID",
			ShoppingCartID: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateShoppingCartItem(gomock.Any(), gomock.Any()).
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
			tc.buildStubs(store)

			server := newTestServer(t, store)
			//recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/usr/v1/users/%d/carts/%d/items/%d", tc.UserID, tc.ShoppingCartID, tc.ShoppingCartItemID)
			request, err := http.NewRequest(fiber.MethodPut, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(t, rsp)
		})
	}
}

func TestDeleteShoppingCartItemAPI(t *testing.T) {
	user, _ := randomSCIUser(t)
	shoppingCart := createRandomShoppingCart(t, user)
	shoppingCartItem := createRandomShoppingCartItemForUpdate(t, shoppingCart)

	testCases := []struct {
		name               string
		ShoppingCartItemID int64
		ShoppingCartID     int64
		UserID             int64
		setupAuth          func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub          func(store *mockdb.MockStore)
		checkResponse      func(t *testing.T, rsp *http.Response)
	}{
		{
			name:               "OK",
			ShoppingCartItemID: shoppingCartItem.ID,
			UserID:             user.ID,
			ShoppingCartID:     shoppingCart.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteShoppingCartItemParams{
					ShoppingCartItemID: shoppingCartItem.ID,
					UserID:             shoppingCartItem.UserID,
					ShoppingCartID:     shoppingCart.ID,
				}

				store.EXPECT().
					DeleteShoppingCartItem(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:               "NotFound",
			ShoppingCartItemID: shoppingCartItem.ID,
			UserID:             user.ID,
			ShoppingCartID:     shoppingCart.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteShoppingCartItemParams{
					ShoppingCartItemID: shoppingCartItem.ID,
					UserID:             shoppingCartItem.UserID,
					ShoppingCartID:     shoppingCart.ID,
				}

				store.EXPECT().
					DeleteShoppingCartItem(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
			},
		},
		{
			name:               "InternalError",
			ShoppingCartItemID: shoppingCartItem.ID,
			UserID:             user.ID,
			ShoppingCartID:     shoppingCart.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteShoppingCartItemParams{
					ShoppingCartItemID: shoppingCartItem.ID,
					UserID:             shoppingCartItem.UserID,
					ShoppingCartID:     shoppingCart.ID,
				}

				store.EXPECT().
					DeleteShoppingCartItem(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:               "InvalidID",
			ShoppingCartItemID: 0,
			UserID:             user.ID,
			ShoppingCartID:     shoppingCart.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteShoppingCartItem(gomock.Any(), gomock.Any()).
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

			// build stubs
			tc.buildStub(store)

			// start test server and send request
			server := newTestServer(t, store)
			//recorder := httptest.NewRecorder()

			// // Marshal body data to JSON
			// data, err := json.Marshal(tc.body)
			// require.NoError(t, err)

			url := fmt.Sprintf("/usr/v1/users/%d/carts/%d/items/%d", tc.UserID, tc.ShoppingCartID, tc.ShoppingCartItemID)
			request, err := http.NewRequest(fiber.MethodDelete, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(t, rsp)
		})

	}

}

func TestDeleteShoppingCartItemAllByUserAPI(t *testing.T) {
	user, _ := randomSCIUser(t)
	shoppingCart := createRandomShoppingCart(t, user)
	shoppingCartItem := createRandomShoppingCartItem(t, shoppingCart)

	var shoppingCartItemList []db.ShoppingCartItem
	shoppingCartItemList = append(shoppingCartItemList, shoppingCartItem...)

	testCases := []struct {
		name           string
		UserID         int64
		ShoppingCartID int64
		setupAuth      func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub      func(store *mockdb.MockStore)
		checkResponse  func(t *testing.T, rsp *http.Response)
	}{
		{
			name:           "OK",
			UserID:         user.ID,
			ShoppingCartID: shoppingCart.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {

				arg := db.DeleteShoppingCartItemAllByUserParams{
					UserID:         user.ID,
					ShoppingCartID: shoppingCart.ID,
				}
				store.EXPECT().
					DeleteShoppingCartItemAllByUser(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(shoppingCartItemList, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:           "NotFound",
			UserID:         user.ID,
			ShoppingCartID: shoppingCart.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteShoppingCartItemAllByUserParams{
					UserID:         user.ID,
					ShoppingCartID: shoppingCart.ID,
				}
				store.EXPECT().
					DeleteShoppingCartItemAllByUser(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return([]db.ShoppingCartItem{}, pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
			},
		},
		{
			name:           "InternalError",
			UserID:         user.ID,
			ShoppingCartID: shoppingCart.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteShoppingCartItemAllByUserParams{
					UserID:         user.ID,
					ShoppingCartID: shoppingCart.ID,
				}
				store.EXPECT().
					DeleteShoppingCartItemAllByUser(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return([]db.ShoppingCartItem{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:           "InvalidID",
			UserID:         0,
			ShoppingCartID: shoppingCart.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteShoppingCartItem(gomock.Any(), gomock.Any()).
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

			// build stubs
			tc.buildStub(store)

			// start test server and send request
			server := newTestServer(t, store)
			//recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/usr/v1/users/%d/carts/%d", tc.UserID, tc.ShoppingCartID)
			request, err := http.NewRequest(fiber.MethodDelete, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(t, rsp)
		})

	}

}

func TestUpdateFinishPurchaseItemAPI(t *testing.T) {
	user, _ := randomSCIUser(t)
	address := createRandomAddress(t)
	userAddress := createRandomUserAddress(t, user, address)
	shoppingCart := createRandomShoppingCart(t, user)
	shippingMethod := createRandomShippingMethod(t)
	paymentMethod := createRandomPaymentMethodUA(t, user)
	orderStatus := createRandomOrderStatus(t)
	orderTotal := util.RandomDecimalString(1, 100)
	finishedPurchase := createRandomFinishedPurchase(t)

	testCases := []struct {
		name           string
		body           fiber.Map
		UserID         int64
		ShoppingCartID int64
		setupAuth      func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs     func(store *mockdb.MockStore)
		checkResponse  func(t *testing.T, rsp *http.Response)
	}{
		{
			name:           "OK",
			UserID:         user.ID,
			ShoppingCartID: shoppingCart.ID,
			body: fiber.Map{

				"user_address_id": userAddress.AddressID,
				"payment_type_id": paymentMethod.PaymentTypeID,

				"shipping_method_id": shippingMethod.ID,
				"order_status_id":    orderStatus.ID,
				"order_total":        orderTotal,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.FinishedPurchaseTxParams{
					UserID:           user.ID,
					UserAddressID:    userAddress.AddressID,
					PaymentTypeID:    paymentMethod.PaymentTypeID,
					ShoppingCartID:   shoppingCart.ID,
					ShippingMethodID: shippingMethod.ID,
					OrderStatusID:    orderStatus.ID,
					OrderTotal:       orderTotal,
				}

				store.EXPECT().
					FinishedPurchaseTx(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(finishedPurchase, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:           "NoAuthorization",
			UserID:         user.ID,
			ShoppingCartID: shoppingCart.ID,
			body: fiber.Map{

				"user_address_id": userAddress.AddressID,
				"payment_type_id": paymentMethod.ID,

				"shipping_method_id": shippingMethod.ID,
				"order_status_id":    orderStatus.ID,
				"order_total":        orderTotal,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					FinishedPurchaseTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:           "InternalError",
			UserID:         user.ID,
			ShoppingCartID: shoppingCart.ID,
			body: fiber.Map{

				"user_address_id": userAddress.AddressID,
				"payment_type_id": paymentMethod.PaymentTypeID,

				"shipping_method_id": shippingMethod.ID,
				"order_status_id":    orderStatus.ID,
				"order_total":        orderTotal,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.FinishedPurchaseTxParams{
					UserID:           user.ID,
					UserAddressID:    userAddress.AddressID,
					PaymentTypeID:    paymentMethod.PaymentTypeID,
					ShoppingCartID:   shoppingCart.ID,
					ShippingMethodID: shippingMethod.ID,
					OrderStatusID:    orderStatus.ID,
					OrderTotal:       orderTotal,
				}

				store.EXPECT().
					FinishedPurchaseTx(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.FinishedPurchaseTxResult{}, pgx.ErrTxClosed)

			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name: "InvalidCartID",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					FinishedPurchaseTx(gomock.Any(), gomock.Any()).
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
			tc.buildStubs(store)

			server := newTestServer(t, store)
			//recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/usr/v1/users/%d/carts/%d/purchase", tc.UserID, tc.ShoppingCartID)
			request, err := http.NewRequest(fiber.MethodPut, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(t, rsp)
		})
	}
}

func randomSCIUser(t *testing.T) (user db.User, password string) {
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

func createRandomShoppingCart(t *testing.T, user db.User) (shoppingSession db.ShoppingCart) {
	shoppingSession = db.ShoppingCart{
		ID:     util.RandomMoney(),
		UserID: user.ID,
	}
	return
}

func createRandomShoppingCartItem1(t *testing.T, shoppingCart db.ShoppingCart) (shoppingCartItem db.ShoppingCartItem) {
	shoppingCartItem = db.ShoppingCartItem{
		ID:             util.RandomMoney(),
		ShoppingCartID: shoppingCart.ID,
		ProductItemID:  util.RandomMoney(),
		Qty:            int32(util.RandomMoney()),
	}
	return
}

func createRandomShoppingCartItem(t *testing.T, shoppingCart db.ShoppingCart) (shoppingCartItem []db.ShoppingCartItem) {
	shoppingCartItem = []db.ShoppingCartItem{
		{ID: util.RandomMoney(),
			ShoppingCartID: shoppingCart.ID,
			ProductItemID:  util.RandomMoney(),
			Qty:            int32(util.RandomMoney())},
	}
	return
}

func createRandomListProductsByIds(t *testing.T, shoppingCartItem []db.ListShoppingCartItemsByUserIDRow, productItem []db.ListProductItemsV2Row) (listByIds []db.ListProductItemsByIDsRow) {
	productsIds := make([]db.ListProductItemsByIDsRow, len(shoppingCartItem))

	for i := 0; i < len(shoppingCartItem); i++ {
		productsIds = append(productsIds, db.ListProductItemsByIDsRow{
			ID:            productItem[i].ID,
			Name:          null.StringFrom(util.RandomString(10)),
			ProductID:     productItem[i].ProductID,
			ProductImage1: productItem[i].ProductImage1,
			ProductImage2: productItem[i].ProductImage2,
			ProductImage3: productItem[i].ProductImage3,
			SizeValue:     productItem[i].SizeValue,
			ColorValue:    productItem[i].ColorValue,
			Price:         productItem[i].Price,
			Active:        productItem[i].Active,
		})
	}

	return
}

func createProductIdsForCart(t *testing.T, shoppingCartItems []db.ListShoppingCartItemsByUserIDRow) (productId []int64) {
	var productsIds []int64

	for i := 0; i < len(shoppingCartItems); i++ {
		productsIds = append(productsIds, shoppingCartItems[i].ProductItemID.Int64)
	}

	return productsIds
}
func createFinalRspForCart(t *testing.T, shopCartItems []db.ListShoppingCartItemsByUserIDRow, productItems []db.ListProductItemsByIDsRow) (rsp []listShoppingCartItemsResponse) {
	for i := 0; i < len(productItems); i++ {
		rsp = append(rsp, listShoppingCartItemsResponse{
			ID:             shopCartItems[i].ID,
			ShoppingCartID: shopCartItems[i].ShoppingCartID,
			CreatedAt:      shopCartItems[i].CreatedAt,
			UpdatedAt:      shopCartItems[i].UpdatedAt,
			ProductItemID:  shopCartItems[i].ProductItemID,
			Name:           productItems[i].Name,
			Qty:            shopCartItems[i].Qty,
			ProductID:      productItems[i].ProductID,
			ProductImage:   productItems[i].ProductImage1.String,
			Price:          productItems[i].Price,
			Active:         productItems[i].Active,
		})
	}
	return
}
func createRandomPaymentMethodUA(t *testing.T, user db.User) (paymentMethod db.PaymentMethod) {
	paymentMethod = db.PaymentMethod{
		ID:            util.RandomMoney(),
		UserID:        user.ID,
		PaymentTypeID: util.RandomMoney(),
		Provider:      util.RandomUser(),
		IsDefault:     util.RandomBool(),
	}
	return
}

func createRandomShippingMethod(t *testing.T) (shippingMethod db.ShippingMethod) {
	shippingMethod = db.ShippingMethod{
		ID:    util.RandomMoney(),
		Name:  util.RandomUser(),
		Price: util.RandomDecimalString(1, 100),
	}
	return
}

func createRandomOrderStatus(t *testing.T) (orderStatus db.OrderStatus) {
	orderStatus = db.OrderStatus{
		ID:     util.RandomMoney(),
		Status: util.RandomUser(),
	}
	return
}

func createRandomFinishedPurchase(t *testing.T) (finishedPurchase db.FinishedPurchaseTxResult) {
	finishedPurchase = db.FinishedPurchaseTxResult{
		UpdatedProductItemID: util.RandomMoney(),
		ShopOrderID:          util.RandomMoney(),
		ShopOrderItemID:      util.RandomMoney(),
	}
	return
}

func createRandomShoppingCartItemForGet(t *testing.T, shoppingCart db.ShoppingCart) (shoppingCartItem []db.GetShoppingCartItemByUserIDCartIDRow) {
	shoppingCartItem = []db.GetShoppingCartItemByUserIDCartIDRow{
		{ID: util.RandomMoney(),
			ShoppingCartID: shoppingCart.ID,
			ProductItemID:  util.RandomMoney(),
			Qty:            int32(util.RandomMoney()),
			UserID:         null.IntFrom(shoppingCart.UserID)},
	}
	return
}

func createRandomShoppingCartItemForUpdate(t *testing.T, shoppingCart db.ShoppingCart) (shoppingCartItem db.UpdateShoppingCartItemRow) {
	shoppingCartItem = db.UpdateShoppingCartItemRow{
		ID:             util.RandomMoney(),
		ShoppingCartID: shoppingCart.ID,
		ProductItemID:  util.RandomMoney(),
		Qty:            int32(util.RandomMoney()),
		UserID:         shoppingCart.UserID,
	}
	return
}

func createRandomListShoppingCartItem(t *testing.T, shoppingCart db.ShoppingCart, productItem db.ListProductItemsV2Row) (shoppingCartItem db.ListShoppingCartItemsByUserIDRow) {
	shoppingCartItem = db.ListShoppingCartItemsByUserIDRow{
		UserID:         shoppingCart.UserID,
		ID:             null.IntFrom(util.RandomMoney()),
		ShoppingCartID: null.IntFrom(shoppingCart.ID),
		ProductItemID:  null.IntFrom(productItem.ID),
		Qty:            null.IntFrom(util.RandomMoney()),
		CreatedAt:      null.TimeFrom(time.Now()),
	}
	return
}

func requireBodyMatchShoppingCartItem(t *testing.T, body io.ReadCloser, shoppingCartItem db.ShoppingCartItem) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotShoppingCartItem db.ShoppingCartItem
	err = json.Unmarshal(data, &gotShoppingCartItem)

	require.NoError(t, err)
	require.Equal(t, shoppingCartItem.ID, gotShoppingCartItem.ID)
	require.Equal(t, shoppingCartItem.ShoppingCartID, gotShoppingCartItem.ShoppingCartID)
	require.Equal(t, shoppingCartItem.ProductItemID, gotShoppingCartItem.ProductItemID)
	require.Equal(t, shoppingCartItem.Qty, gotShoppingCartItem.Qty)
	require.Equal(t, shoppingCartItem.CreatedAt, gotShoppingCartItem.CreatedAt)
}
func requireBodyMatchShoppingCartItemForCreate(t *testing.T, body io.ReadCloser, shoppingCartItems []db.ShoppingCartItem) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotShoppingCartItem []db.ShoppingCartItem
	err = json.Unmarshal(data, &gotShoppingCartItem)

	require.NoError(t, err)
	for i, shoppingCartItem := range shoppingCartItems {
		require.Equal(t, shoppingCartItem.ID, gotShoppingCartItem[i].ID)
		require.Equal(t, shoppingCartItem.ShoppingCartID, gotShoppingCartItem[i].ShoppingCartID)
		require.Equal(t, shoppingCartItem.ProductItemID, gotShoppingCartItem[i].ProductItemID)
		require.Equal(t, shoppingCartItem.Qty, gotShoppingCartItem[i].Qty)
		require.Equal(t, shoppingCartItem.CreatedAt, gotShoppingCartItem[i].CreatedAt)
	}
}

func requireBodyMatchShoppingCartItemForGet(t *testing.T, body io.ReadCloser, shoppingCartItem []db.GetShoppingCartItemByUserIDCartIDRow) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotShoppingCartItem []db.GetShoppingCartItemByUserIDCartIDRow
	err = json.Unmarshal(data, &gotShoppingCartItem)

	require.NoError(t, err)
	for i := 0; i < len(shoppingCartItem)-1; i++ {
		require.Equal(t, shoppingCartItem[i].ID, gotShoppingCartItem[i].ID)
		require.Equal(t, shoppingCartItem[i].ShoppingCartID, gotShoppingCartItem[i].ShoppingCartID)
		require.Equal(t, shoppingCartItem[i].ProductItemID, gotShoppingCartItem[i].ProductItemID)
		require.Equal(t, shoppingCartItem[i].Qty, gotShoppingCartItem[i].Qty)
		require.Equal(t, shoppingCartItem[i].CreatedAt, gotShoppingCartItem[i].CreatedAt)
		require.Equal(t, shoppingCartItem[i].UserID, gotShoppingCartItem[i].UserID)

	}
}

func requireBodyMatchListShoppingCartItem(t *testing.T, body io.ReadCloser, shoppingCartItem []db.ListShoppingCartItemsByUserIDRow) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotShoppingCartItem []db.ListShoppingCartItemsByUserIDRow
	err = json.Unmarshal(data, &gotShoppingCartItem)

	require.NoError(t, err)
	for i, gotCartItem := range gotShoppingCartItem {
		require.Equal(t, shoppingCartItem[i].ID, gotCartItem.ID)
		require.Equal(t, shoppingCartItem[i].ShoppingCartID, gotCartItem.ShoppingCartID)
		require.Equal(t, shoppingCartItem[i].ProductItemID, gotCartItem.ProductItemID)
		require.Equal(t, shoppingCartItem[i].UserID, gotCartItem.UserID)
	}
}

func requireBodyMatchFinalListShoppingCartItem(t *testing.T, body io.ReadCloser, finalRsp []listShoppingCartItemsResponse) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotShoppingCartItem []listShoppingCartItemsResponse
	err = json.Unmarshal(data, &gotShoppingCartItem)

	require.NoError(t, err)
	for i, gotCartItem := range gotShoppingCartItem {
		require.Equal(t, finalRsp[i].ID, gotCartItem.ID)
		require.Equal(t, finalRsp[i].ShoppingCartID, gotCartItem.ShoppingCartID)
		require.Equal(t, finalRsp[i].ProductItemID, gotCartItem.ProductItemID)
	}
}
