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

func TestCreateWishListItemAPI(t *testing.T) {
	user, _ := randomWLIUser(t)
	wishList := createRandomWishList(t, user)
	wishListItem := createRandomWishListItem(t, wishList)

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
				"product_item_id": wishListItem.ProductItemID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					GetWishListByUserID(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(wishList, nil)

				arg := db.CreateWishListItemParams{
					WishListID:    wishListItem.WishListID,
					ProductItemID: wishListItem.ProductItemID,
				}

				store.EXPECT().
					CreateWishListItem(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(wishListItem, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchWishListItem(t, recorder.Body, wishListItem)
			},
		},
		{
			name: "NoAuthorization",
			body: gin.H{
				"user_id":         user.ID,
				"product_item_id": wishListItem.ProductItemID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetWishListByUserID(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					CreateWishListItem(gomock.Any(), gomock.Any()).
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
				"product_item_id": wishListItem.ProductItemID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					GetWishListByUserID(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(db.WishList{}, pgx.ErrTxClosed)

				arg := db.CreateWishListItemParams{
					WishListID:    wishListItem.WishListID,
					ProductItemID: wishListItem.ProductItemID,
				}

				store.EXPECT().
					CreateWishListItem(gomock.Any(), gomock.Eq(arg)).
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
				"product_item_id": wishListItem.ProductItemID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetWishListByUserID(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					CreateWishListItem(gomock.Any(), gomock.Any()).
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

			url := "/users/wish-list"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestGetWishListItemAPI(t *testing.T) {
	user, _ := randomWLIUser(t)
	wishList := createRandomWishList(t, user)
	wishListItem := createRandomWishListItemForGet(t, wishList)

	testCases := []struct {
		name          string
		WishListID    int64
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name:       "OK",
			WishListID: wishList.ID,
			body: gin.H{
				"user_id": user.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.GetWishListItemByUserIDCartIDParams{
					UserID:     user.ID,
					WishListID: wishList.ID,
				}

				store.EXPECT().
					GetWishListItemByUserIDCartID(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(wishListItem, nil)

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchWishListItemForGet(t, recorder.Body, wishListItem)
			},
		},
		{
			name:       "NoAuthorization",
			WishListID: wishList.ID,
			body: gin.H{
				"user_id": user.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetWishListItemByUserIDCartID(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:       "InternalError",
			WishListID: wishList.ID,
			body: gin.H{
				"user_id": user.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.GetWishListItemByUserIDCartIDParams{
					UserID:     user.ID,
					WishListID: wishList.ID,
				}

				store.EXPECT().
					GetWishListItemByUserIDCartID(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.GetWishListItemByUserIDCartIDRow{}, pgx.ErrTxClosed)

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:       "InvalidUserID",
			WishListID: wishList.ID,
			body: gin.H{
				"user_id": 0,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetWishListItemByUserIDCartID(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/users/wish-list/%d", tc.WishListID)
			request, err := http.NewRequest(http.MethodGet, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestListWishListItemAPI(t *testing.T) {
	user, _ := randomWLIUser(t)
	wishList := createRandomWishList(t, user)

	n := 5
	wishListItems := make([]db.ListWishListItemsByUserIDRow, n)
	for i := 0; i < n; i++ {
		wishListItems[i] = createRandomListWishListItem(t, wishList)
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
					ListWishListItemsByUserID(gomock.Any(), gomock.Eq(wishList.UserID)).
					Times(1).
					Return(wishListItems, nil)

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchListWishListItem(t, recorder.Body, wishListItems)
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
					ListWishListItemsByUserID(gomock.Any(), gomock.Any()).
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
					ListWishListItemsByUserID(gomock.Any(), gomock.Eq(wishList.UserID)).
					Times(1).
					Return([]db.ListWishListItemsByUserIDRow{}, pgx.ErrTxClosed)

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
					ListWishListItemsByUserID(gomock.Any(), gomock.Any()).
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

			url := "/users/wish-list"
			request, err := http.NewRequest(http.MethodGet, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestUpdateWishListItemAPI(t *testing.T) {
	user, _ := randomWLIUser(t)
	wishList := createRandomWishList(t, user)
	wishListItem := createRandomWishListItemForUpdate(t, wishList)

	testCases := []struct {
		name          string
		WishListID    int64
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:       "OK",
			WishListID: wishList.ID,
			body: gin.H{
				"id":              wishListItem.ID,
				"user_id":         wishList.UserID,
				"product_item_id": wishListItem.ProductItemID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.UpdateWishListItemParams{
					ProductItemID: null.IntFrom(wishListItem.ProductItemID),
					ID:            wishListItem.ID,
					WishListID:    wishListItem.WishListID,
				}

				store.EXPECT().
					UpdateWishListItem(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(wishListItem, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:       "NoAuthorization",
			WishListID: wishList.ID,
			body: gin.H{
				"id":              wishListItem.ID,
				"user_id":         wishList.UserID,
				"product_item_id": wishListItem.ProductItemID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateWishListItem(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:       "InternalError",
			WishListID: wishList.ID,
			body: gin.H{
				"id":              wishListItem.ID,
				"user_id":         wishList.UserID,
				"product_item_id": wishListItem.ProductItemID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateWishListItemParams{
					ProductItemID: null.IntFrom(wishListItem.ProductItemID),
					ID:            wishListItem.ID,
					WishListID:    wishListItem.WishListID,
				}

				store.EXPECT().
					UpdateWishListItem(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.UpdateWishListItemRow{}, pgx.ErrTxClosed)

			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:       "InvalidCartID",
			WishListID: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateWishListItem(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/users/wish-list/%d", tc.WishListID)
			request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestDeleteWishListItemAPI(t *testing.T) {
	user, _ := randomWLIUser(t)
	wishList := createRandomWishList(t, user)
	wishListItem := createRandomWishListItemForUpdate(t, wishList)

	testCases := []struct {
		name           string
		WishListItemID int64
		body           gin.H
		setupAuth      func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub      func(store *mockdb.MockStore)
		checkResponse  func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:           "OK",
			WishListItemID: wishListItem.ID,
			body: gin.H{
				"user_id": wishList.UserID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteWishListItemParams{
					ID:     wishListItem.ID,
					UserID: wishListItem.UserID,
				}

				store.EXPECT().
					DeleteWishListItem(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:           "NotFound",
			WishListItemID: wishListItem.ID,
			body: gin.H{
				"user_id": wishList.UserID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteWishListItemParams{
					ID:     wishListItem.ID,
					UserID: wishListItem.UserID,
				}

				store.EXPECT().
					DeleteWishListItem(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:           "InternalError",
			WishListItemID: wishListItem.ID,
			body: gin.H{
				"user_id": wishList.UserID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteWishListItemParams{
					ID:     wishListItem.ID,
					UserID: wishListItem.UserID,
				}

				store.EXPECT().
					DeleteWishListItem(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:           "InvalidID",
			WishListItemID: 0,
			body: gin.H{
				"user_id": wishList.UserID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteWishListItem(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/users/wish-list/%d", tc.WishListItemID)
			request, err := http.NewRequest(http.MethodDelete, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			//check response
			tc.checkResponse(t, recorder)
		})

	}

}

func TestDeleteWishListItemAllByUserAPI(t *testing.T) {
	user, _ := randomWLIUser(t)
	wishList := createRandomWishList(t, user)
	wishListItem := createRandomWishListItem(t, wishList)

	var wishListItemList []db.WishListItem
	wishListItemList = append(wishListItemList, wishListItem)

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
				"user_id": wishList.UserID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {

				store.EXPECT().
					DeleteWishListItemAllByUser(gomock.Any(), gomock.Eq(wishList.UserID)).
					Times(1).
					Return(wishListItemList, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "NotFound",
			body: gin.H{
				"user_id": wishList.UserID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {

				store.EXPECT().
					DeleteWishListItemAllByUser(gomock.Any(), gomock.Eq(wishList.UserID)).
					Times(1).
					Return([]db.WishListItem{}, pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"user_id": wishList.UserID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {

				store.EXPECT().
					DeleteWishListItemAllByUser(gomock.Any(), gomock.Eq(wishList.UserID)).
					Times(1).
					Return([]db.WishListItem{}, pgx.ErrTxClosed)
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
					DeleteWishListItem(gomock.Any(), gomock.Any()).
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

			url := "/users/wish-list/delete-all"
			request, err := http.NewRequest(http.MethodDelete, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			//check response
			tc.checkResponse(t, recorder)
		})

	}

}

func randomWLIUser(t *testing.T) (user db.User, password string) {
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

func createRandomWishList(t *testing.T, user db.User) (shoppingSession db.WishList) {
	shoppingSession = db.WishList{
		ID:     util.RandomMoney(),
		UserID: user.ID,
	}
	return
}

func createRandomWishListItem(t *testing.T, wishList db.WishList) (wishListItem db.WishListItem) {
	wishListItem = db.WishListItem{
		ID:            util.RandomMoney(),
		WishListID:    wishList.ID,
		ProductItemID: util.RandomMoney(),
	}
	return
}

func createRandomWishListItemForGet(t *testing.T, wishList db.WishList) (wishListItem db.GetWishListItemByUserIDCartIDRow) {
	wishListItem = db.GetWishListItemByUserIDCartIDRow{
		ID:            util.RandomMoney(),
		WishListID:    wishList.ID,
		ProductItemID: util.RandomMoney(),
		UserID:        null.IntFrom(wishList.UserID),
	}
	return
}

func createRandomWishListItemForUpdate(t *testing.T, wishList db.WishList) (wishListItem db.UpdateWishListItemRow) {
	wishListItem = db.UpdateWishListItemRow{
		ID:            util.RandomMoney(),
		WishListID:    wishList.ID,
		ProductItemID: util.RandomMoney(),
		UserID:        wishList.UserID,
	}
	return
}

func createRandomListWishListItem(t *testing.T, wishList db.WishList) (wishListItem db.ListWishListItemsByUserIDRow) {
	wishListItem = db.ListWishListItemsByUserIDRow{
		UserID:        wishList.UserID,
		ID:            null.IntFrom(util.RandomMoney()),
		WishListID:    null.IntFrom(wishList.ID),
		ProductItemID: null.IntFrom(util.RandomMoney()),
		CreatedAt:     null.TimeFrom(time.Now()),
	}
	return
}

func requireBodyMatchWishListItem(t *testing.T, body *bytes.Buffer, wishListItem db.WishListItem) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotWishListItem db.WishListItem
	err = json.Unmarshal(data, &gotWishListItem)

	require.NoError(t, err)
	require.Equal(t, wishListItem.ID, gotWishListItem.ID)
	require.Equal(t, wishListItem.WishListID, gotWishListItem.WishListID)
	require.Equal(t, wishListItem.ProductItemID, gotWishListItem.ProductItemID)
	require.Equal(t, wishListItem.CreatedAt, gotWishListItem.CreatedAt)
}

func requireBodyMatchWishListItemForGet(t *testing.T, body *bytes.Buffer, wishListItem db.GetWishListItemByUserIDCartIDRow) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotWishListItem db.GetWishListItemByUserIDCartIDRow
	err = json.Unmarshal(data, &gotWishListItem)

	require.NoError(t, err)
	require.Equal(t, wishListItem.ID, gotWishListItem.ID)
	require.Equal(t, wishListItem.WishListID, gotWishListItem.WishListID)
	require.Equal(t, wishListItem.ProductItemID, gotWishListItem.ProductItemID)
	require.Equal(t, wishListItem.CreatedAt, gotWishListItem.CreatedAt)
	require.Equal(t, wishListItem.UserID, gotWishListItem.UserID)
}

func requireBodyMatchListWishListItem(t *testing.T, body *bytes.Buffer, WishListItem []db.ListWishListItemsByUserIDRow) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotWishListItem []db.ListWishListItemsByUserIDRow
	err = json.Unmarshal(data, &gotWishListItem)

	require.NoError(t, err)
	for i, gotCartItem := range gotWishListItem {
		require.Equal(t, WishListItem[i].ID, gotCartItem.ID)
		require.Equal(t, WishListItem[i].WishListID, gotCartItem.WishListID)
		require.Equal(t, WishListItem[i].ProductItemID, gotCartItem.ProductItemID)
		require.Equal(t, WishListItem[i].UserID, gotCartItem.UserID)
	}
}
