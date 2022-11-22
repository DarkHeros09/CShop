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

func TestCreateShoppingCartItemAPI(t *testing.T) {
	user, _ := randomSCIUser(t)
	shoppingCart := createRandomShoppingCart(t, user)
	shoppingCartItem := createRandomShoppingCartItem(t, shoppingCart)

	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"user_id":         user.ID,
				"product_item_id": shoppingCartItem.ProductItemID,
				"qty":             shoppingCartItem.Qty,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					GetShoppingCartByUserID(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(shoppingCart, nil)

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
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchShoppingCartItem(t, recorder.Body, shoppingCartItem)
			},
		},
		{
			name: "NoAuthorization",
			body: gin.H{
				"user_id":         user.ID,
				"product_item_id": shoppingCartItem.ProductItemID,
				"qty":             shoppingCartItem.Qty,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetShoppingCartByUserID(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					CreateShoppingCartItem(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"user_id":         user.ID,
				"product_item_id": shoppingCartItem.ProductItemID,
				"qty":             shoppingCartItem.Qty,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					GetShoppingCartByUserID(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(db.ShoppingCart{}, pgx.ErrTxClosed)

				arg := db.CreateShoppingCartItemParams{
					ShoppingCartID: shoppingCartItem.ShoppingCartID,
					ProductItemID:  shoppingCartItem.ProductItemID,
					Qty:            shoppingCartItem.Qty,
				}

				store.EXPECT().
					CreateShoppingCartItem(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidUserID",
			body: gin.H{
				"user_id":         0,
				"product_item_id": shoppingCartItem.ProductItemID,
				"qty":             shoppingCartItem.Qty,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetShoppingCartByUserID(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					CreateShoppingCartItem(gomock.Any(), gomock.Any()).
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

			url := "/users/cart"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestGetShoppingCartItemAPI(t *testing.T) {
	user, _ := randomSCIUser(t)
	shoppingCart := createRandomShoppingCart(t, user)
	shoppingCartItem := createRandomShoppingCartItemForGet(t, shoppingCart)

	testCases := []struct {
		name           string
		ShoppingCartID int64
		body           gin.H
		setupAuth      func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs     func(store *mockdb.MockStore)
		checkResponse  func(recoder *httptest.ResponseRecorder)
	}{
		{
			name:           "OK",
			ShoppingCartID: shoppingCart.ID,
			body: gin.H{
				"user_id": user.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.GetShoppingCartItemByUserIDCartIDParams{
					UserID:         user.ID,
					ShoppingCartID: shoppingCart.ID,
				}

				store.EXPECT().
					GetShoppingCartItemByUserIDCartID(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(shoppingCartItem, nil)

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchShoppingCartItemForGet(t, recorder.Body, shoppingCartItem)
			},
		},
		{
			name:           "NoAuthorization",
			ShoppingCartID: shoppingCart.ID,
			body: gin.H{
				"user_id": user.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetShoppingCartItemByUserIDCartID(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:           "InternalError",
			ShoppingCartID: shoppingCart.ID,
			body: gin.H{
				"user_id": user.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.GetShoppingCartItemByUserIDCartIDParams{
					UserID:         user.ID,
					ShoppingCartID: shoppingCart.ID,
				}

				store.EXPECT().
					GetShoppingCartItemByUserIDCartID(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.GetShoppingCartItemByUserIDCartIDRow{}, pgx.ErrTxClosed)

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:           "InvalidUserID",
			ShoppingCartID: shoppingCart.ID,
			body: gin.H{
				"user_id": 0,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetShoppingCartItemByUserIDCartID(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/users/cart/%d", tc.ShoppingCartID)
			request, err := http.NewRequest(http.MethodGet, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestListShoppingCartItemAPI(t *testing.T) {
	user, _ := randomSCIUser(t)
	shoppingCart := createRandomShoppingCart(t, user)

	n := 5
	shoppingCartItems := make([]db.ListShoppingCartItemsByUserIDRow, n)
	for i := 0; i < n; i++ {
		shoppingCartItems[i] = createRandomListShoppingCartItem(t, shoppingCart)
	}

	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"user_id": user.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					ListShoppingCartItemsByUserID(gomock.Any(), gomock.Eq(shoppingCart.UserID)).
					Times(1).
					Return(shoppingCartItems, nil)

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchListShoppingCartItem(t, recorder.Body, shoppingCartItems)
			},
		},
		{
			name: "NoAuthorization",
			body: gin.H{
				"user_id": user.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListShoppingCartItemsByUserID(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"user_id": user.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					ListShoppingCartItemsByUserID(gomock.Any(), gomock.Eq(shoppingCart.UserID)).
					Times(1).
					Return([]db.ListShoppingCartItemsByUserIDRow{}, pgx.ErrTxClosed)

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidUserID",
			body: gin.H{
				"user_id": 0,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListShoppingCartItemsByUserID(gomock.Any(), gomock.Any()).
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

			url := "/users/cart"
			request, err := http.NewRequest(http.MethodGet, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestUpdateShoppingCartItemAPI(t *testing.T) {
	user, _ := randomSCIUser(t)
	shoppingCart := createRandomShoppingCart(t, user)
	shoppingCartItem := createRandomShoppingCartItemForUpdate(t, shoppingCart)
	qty := util.RandomMoney()

	testCases := []struct {
		name           string
		ShoppingCartID int64
		body           gin.H
		setupAuth      func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs     func(store *mockdb.MockStore)
		checkResponse  func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:           "OK",
			ShoppingCartID: shoppingCart.ID,
			body: gin.H{
				"id":              shoppingCartItem.ID,
				"user_id":         shoppingCart.UserID,
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:           "NoAuthorization",
			ShoppingCartID: shoppingCart.ID,
			body: gin.H{
				"id":              shoppingCartItem.ID,
				"user_id":         shoppingCart.UserID,
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:           "InternalError",
			ShoppingCartID: shoppingCart.ID,
			body: gin.H{
				"id":              shoppingCartItem.ID,
				"user_id":         shoppingCart.UserID,
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
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

			url := fmt.Sprintf("/users/cart/%d", tc.ShoppingCartID)
			request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
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
		body               gin.H
		setupAuth          func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub          func(store *mockdb.MockStore)
		checkResponse      func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:               "OK",
			ShoppingCartItemID: shoppingCartItem.ID,
			body: gin.H{
				"user_id": shoppingCart.UserID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteShoppingCartItemParams{
					ID:     shoppingCartItem.ID,
					UserID: shoppingCartItem.UserID,
				}

				store.EXPECT().
					DeleteShoppingCartItem(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:               "NotFound",
			ShoppingCartItemID: shoppingCartItem.ID,
			body: gin.H{
				"user_id": shoppingCart.UserID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteShoppingCartItemParams{
					ID:     shoppingCartItem.ID,
					UserID: shoppingCartItem.UserID,
				}

				store.EXPECT().
					DeleteShoppingCartItem(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:               "InternalError",
			ShoppingCartItemID: shoppingCartItem.ID,
			body: gin.H{
				"user_id": shoppingCart.UserID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteShoppingCartItemParams{
					ID:     shoppingCartItem.ID,
					UserID: shoppingCartItem.UserID,
				}

				store.EXPECT().
					DeleteShoppingCartItem(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:               "InvalidID",
			ShoppingCartItemID: 0,
			body: gin.H{
				"user_id": shoppingCart.UserID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteShoppingCartItem(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/users/cart/%d", tc.ShoppingCartItemID)
			request, err := http.NewRequest(http.MethodDelete, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			//check response
			tc.checkResponse(t, recorder)
		})

	}

}

func TestDeleteShoppingCartItemAllByUserAPI(t *testing.T) {
	user, _ := randomSCIUser(t)
	shoppingCart := createRandomShoppingCart(t, user)
	shoppingCartItem := createRandomShoppingCartItem(t, shoppingCart)

	var shoppingCartItemList []db.ShoppingCartItem
	shoppingCartItemList = append(shoppingCartItemList, shoppingCartItem)

	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"user_id": shoppingCart.UserID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {

				store.EXPECT().
					DeleteShoppingCartItemAllByUser(gomock.Any(), gomock.Eq(shoppingCart.UserID)).
					Times(1).
					Return(shoppingCartItemList, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "NotFound",
			body: gin.H{
				"user_id": shoppingCart.UserID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {

				store.EXPECT().
					DeleteShoppingCartItemAllByUser(gomock.Any(), gomock.Eq(shoppingCart.UserID)).
					Times(1).
					Return([]db.ShoppingCartItem{}, pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"user_id": shoppingCart.UserID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {

				store.EXPECT().
					DeleteShoppingCartItemAllByUser(gomock.Any(), gomock.Eq(shoppingCart.UserID)).
					Times(1).
					Return([]db.ShoppingCartItem{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidID",
			body: gin.H{
				"user_id": 0,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteShoppingCartItem(gomock.Any(), gomock.Any()).
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

			url := "/users/cart/delete-all"
			request, err := http.NewRequest(http.MethodDelete, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			//check response
			tc.checkResponse(t, recorder)
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
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"user_id":            userAddress.UserID,
				"user_address_id":    userAddress.AddressID,
				"payment_method_id":  paymentMethod.ID,
				"shopping_cart_id":   shoppingCart.ID,
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
					PaymentMethodID:  paymentMethod.ID,
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "NoAuthorization",
			body: gin.H{
				"user_id":            userAddress.UserID,
				"user_address_id":    userAddress.AddressID,
				"payment_method_id":  paymentMethod.ID,
				"shopping_cart_id":   shoppingCart.ID,
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"user_id":            userAddress.UserID,
				"user_address_id":    userAddress.AddressID,
				"payment_method_id":  paymentMethod.ID,
				"shopping_cart_id":   shoppingCart.ID,
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
					PaymentMethodID:  paymentMethod.ID,
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
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
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

			url := "/users/cart/purchase"
			request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
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

func createRandomShoppingCartItem(t *testing.T, shoppingCart db.ShoppingCart) (shoppingCartItem db.ShoppingCartItem) {
	shoppingCartItem = db.ShoppingCartItem{
		ID:             util.RandomMoney(),
		ShoppingCartID: shoppingCart.ID,
		ProductItemID:  util.RandomMoney(),
		Qty:            int32(util.RandomMoney()),
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

func createRandomShoppingCartItemForGet(t *testing.T, shoppingCart db.ShoppingCart) (shoppingCartItem db.GetShoppingCartItemByUserIDCartIDRow) {
	shoppingCartItem = db.GetShoppingCartItemByUserIDCartIDRow{
		ID:             util.RandomMoney(),
		ShoppingCartID: shoppingCart.ID,
		ProductItemID:  util.RandomMoney(),
		Qty:            int32(util.RandomMoney()),
		UserID:         null.IntFrom(shoppingCart.UserID),
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

func createRandomListShoppingCartItem(t *testing.T, shoppingCart db.ShoppingCart) (shoppingCartItem db.ListShoppingCartItemsByUserIDRow) {
	shoppingCartItem = db.ListShoppingCartItemsByUserIDRow{
		UserID:         shoppingCart.UserID,
		ID:             null.IntFrom(util.RandomMoney()),
		ShoppingCartID: null.IntFrom(shoppingCart.ID),
		ProductItemID:  null.IntFrom(util.RandomMoney()),
		Qty:            null.IntFrom(util.RandomMoney()),
		CreatedAt:      null.TimeFrom(time.Now()),
	}
	return
}

func requireBodyMatchShoppingCartItem(t *testing.T, body *bytes.Buffer, shoppingCartItem db.ShoppingCartItem) {
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

func requireBodyMatchShoppingCartItemForGet(t *testing.T, body *bytes.Buffer, shoppingCartItem db.GetShoppingCartItemByUserIDCartIDRow) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotShoppingCartItem db.GetShoppingCartItemByUserIDCartIDRow
	err = json.Unmarshal(data, &gotShoppingCartItem)

	require.NoError(t, err)
	require.Equal(t, shoppingCartItem.ID, gotShoppingCartItem.ID)
	require.Equal(t, shoppingCartItem.ShoppingCartID, gotShoppingCartItem.ShoppingCartID)
	require.Equal(t, shoppingCartItem.ProductItemID, gotShoppingCartItem.ProductItemID)
	require.Equal(t, shoppingCartItem.Qty, gotShoppingCartItem.Qty)
	require.Equal(t, shoppingCartItem.CreatedAt, gotShoppingCartItem.CreatedAt)
	require.Equal(t, shoppingCartItem.UserID, gotShoppingCartItem.UserID)
}

func requireBodyMatchListShoppingCartItem(t *testing.T, body *bytes.Buffer, shoppingCartItem []db.ListShoppingCartItemsByUserIDRow) {
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
