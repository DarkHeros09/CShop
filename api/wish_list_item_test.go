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
		body          fiber.Map
		UserID        int64
		WishListID    int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(rsp *http.Response)
	}{
		{
			name:       "OK",
			UserID:     user.ID,
			WishListID: wishList.ID,
			body: fiber.Map{
				"product_item_id": wishListItem.ProductItemID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.CreateWishListItemParams{
					WishListID:    wishListItem.WishListID,
					ProductItemID: wishListItem.ProductItemID,
				}

				store.EXPECT().
					CreateWishListItem(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(wishListItem, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchWishListItem(t, rsp.Body, wishListItem)
			},
		},
		{
			name:       "NoAuthorization",
			UserID:     user.ID,
			WishListID: wishList.ID,
			body: fiber.Map{
				"product_item_id": wishListItem.ProductItemID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					CreateWishListItem(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:       "InternalError",
			UserID:     user.ID,
			WishListID: wishList.ID,
			body: fiber.Map{
				"product_item_id": wishListItem.ProductItemID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.CreateWishListItemParams{
					WishListID:    wishListItem.WishListID,
					ProductItemID: wishListItem.ProductItemID,
				}

				store.EXPECT().
					CreateWishListItem(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.WishListItem{}, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:       "InvalidUserID",
			UserID:     0,
			WishListID: wishList.ID,
			body: fiber.Map{
				"product_item_id": wishListItem.ProductItemID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					CreateWishListItem(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/usr/v1/users/%d/wish-lists/%d", tc.UserID, tc.WishListID)
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

func TestGetWishListItemAPI(t *testing.T) {
	user, _ := randomWLIUser(t)
	wishList := createRandomWishList(t, user)
	wishListItem := createRandomWishListItemForGet(t, wishList)

	testCases := []struct {
		name           string
		WishListID     int64
		WishListItemID int64
		UserID         int64
		setupAuth      func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs     func(store *mockdb.MockStore)
		checkResponse  func(rsp *http.Response)
	}{
		{
			name:           "OK",
			WishListID:     wishList.ID,
			WishListItemID: wishListItem.ID,
			UserID:         user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.GetWishListItemByUserIDCartIDParams{
					UserID:     user.ID,
					WishListID: wishList.ID,
					ID:         wishListItem.ID,
				}

				store.EXPECT().
					GetWishListItemByUserIDCartID(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(wishListItem, nil)

			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchWishListItemForGet(t, rsp.Body, wishListItem)
			},
		},
		{
			name:           "NoAuthorization",
			WishListID:     wishList.ID,
			UserID:         user.ID,
			WishListItemID: wishListItem.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetWishListItemByUserIDCartID(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:           "InternalError",
			WishListID:     wishList.ID,
			UserID:         user.ID,
			WishListItemID: wishListItem.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.GetWishListItemByUserIDCartIDParams{
					UserID:     user.ID,
					WishListID: wishList.ID,
					ID:         wishListItem.ID,
				}

				store.EXPECT().
					GetWishListItemByUserIDCartID(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.WishListItem{}, pgx.ErrTxClosed)

			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:           "InvalidUserID",
			WishListID:     wishList.ID,
			UserID:         0,
			WishListItemID: wishListItem.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetWishListItemByUserIDCartID(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/usr/v1/users/%d/wish-lists/%d/items/%d", tc.UserID, tc.WishListID, tc.WishListItemID)
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

// ! not query
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
		UserID        int64
		WishID        int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(rsp *http.Response)
	}{
		{
			name:   "OK",
			UserID: user.ID,
			WishID: wishList.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					ListWishListItemsByUserID(gomock.Any(), gomock.Eq(wishList.UserID)).
					Times(1).
					Return(wishListItems, nil)

			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchListWishListItem(t, rsp.Body, wishListItems)
			},
		},
		{
			name:   "NoAuthorization",
			UserID: user.ID,
			WishID: wishList.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListWishListItemsByUserID(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:   "InternalError",
			UserID: user.ID,
			WishID: wishList.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					ListWishListItemsByUserID(gomock.Any(), gomock.Eq(wishList.UserID)).
					Times(1).
					Return([]db.ListWishListItemsByUserIDRow{}, pgx.ErrTxClosed)

			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:   "InvalidUserID",
			UserID: 0,
			WishID: wishList.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListWishListItemsByUserID(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/usr/v1/users/%d/wish-lists/%d", tc.UserID, tc.WishID)
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

func TestUpdateWishListItemAPI(t *testing.T) {
	user, _ := randomWLIUser(t)
	wishList := createRandomWishList(t, user)
	wishListItem := createRandomWishListItemForUpdate(t, wishList)

	testCases := []struct {
		name           string
		UserID         int64
		WishListID     int64
		WishListItemID int64
		body           fiber.Map
		setupAuth      func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs     func(store *mockdb.MockStore)
		checkResponse  func(t *testing.T, rsp *http.Response)
	}{
		{
			name:           "OK",
			WishListID:     wishList.ID,
			WishListItemID: wishListItem.ID,
			UserID:         wishList.UserID,
			body: fiber.Map{
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
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:           "NoAuthorization",
			WishListID:     wishList.ID,
			WishListItemID: wishListItem.ID,
			UserID:         wishList.UserID,
			body: fiber.Map{
				"product_item_id": wishListItem.ProductItemID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateWishListItem(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:           "InternalError",
			WishListID:     wishList.ID,
			WishListItemID: wishListItem.ID,
			UserID:         wishList.UserID,
			body: fiber.Map{
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
					Return(db.WishListItem{}, pgx.ErrTxClosed)

			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:           "InvalidWishListID",
			WishListID:     0,
			WishListItemID: wishListItem.ID,
			UserID:         wishList.UserID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateWishListItem(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/usr/v1/users/%d/wish-lists/%d/items/%d", tc.UserID, tc.WishListID, tc.WishListItemID)
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

func TestDeleteWishListItemAPI(t *testing.T) {
	user, _ := randomWLIUser(t)
	wishList := createRandomWishList(t, user)
	wishListItem := createRandomWishListItemForUpdate(t, wishList)

	testCases := []struct {
		name           string
		WishListItemID int64
		UserID         int64
		WishListID     int64
		setupAuth      func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub      func(store *mockdb.MockStore)
		checkResponse  func(t *testing.T, rsp *http.Response)
	}{
		{
			name:           "OK",
			WishListItemID: wishListItem.ID,
			UserID:         wishList.UserID,
			WishListID:     wishList.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteWishListItemParams{
					ID:         wishListItem.ID,
					WishListID: wishList.ID,
				}

				store.EXPECT().
					DeleteWishListItem(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:           "NotFound",
			WishListItemID: wishListItem.ID,
			UserID:         wishList.UserID,
			WishListID:     wishList.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteWishListItemParams{
					ID:         wishListItem.ID,
					WishListID: wishList.ID,
				}

				store.EXPECT().
					DeleteWishListItem(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
			},
		},
		{
			name:           "InternalError",
			WishListItemID: wishListItem.ID,
			UserID:         wishList.UserID,
			WishListID:     wishList.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteWishListItemParams{
					ID:         wishListItem.ID,
					WishListID: wishList.ID,
				}

				store.EXPECT().
					DeleteWishListItem(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:           "InvalidID",
			WishListItemID: 0,
			UserID:         wishList.UserID,
			WishListID:     wishList.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteWishListItem(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/usr/v1/users/%d/wish-lists/%d/items/%d", tc.UserID, tc.WishListID, tc.WishListItemID)
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

func TestDeleteWishListItemAllAPI(t *testing.T) {
	user, _ := randomWLIUser(t)
	wishList := createRandomWishList(t, user)
	wishListItem := createRandomWishListItem(t, wishList)

	var wishListItemList []db.WishListItem
	wishListItemList = append(wishListItemList, wishListItem)

	testCases := []struct {
		name          string
		UserID        int64
		WishListID    int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rsp *http.Response)
	}{
		{
			name:       "OK",
			UserID:     wishList.UserID,
			WishListID: wishList.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {

				store.EXPECT().
					DeleteWishListItemAll(gomock.Any(), gomock.Eq(wishList.ID)).
					Times(1).
					Return(wishListItemList, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:       "NotFound",
			UserID:     wishList.UserID,
			WishListID: wishList.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {

				store.EXPECT().
					DeleteWishListItemAll(gomock.Any(), gomock.Eq(wishList.ID)).
					Times(1).
					Return([]db.WishListItem{}, pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
			},
		},
		{
			name:       "InternalError",
			UserID:     wishList.UserID,
			WishListID: wishList.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {

				store.EXPECT().
					DeleteWishListItemAll(gomock.Any(), gomock.Eq(wishList.ID)).
					Times(1).
					Return([]db.WishListItem{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:       "InvalidID",
			UserID:     0,
			WishListID: wishList.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteWishListItem(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/usr/v1/users/%d/wish-lists/%d", tc.UserID, tc.WishListID)
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

func createRandomWishListItemForGet(t *testing.T, wishList db.WishList) (wishListItem db.WishListItem) {
	wishListItem = db.WishListItem{
		ID:            util.RandomMoney(),
		WishListID:    wishList.ID,
		ProductItemID: util.RandomMoney(),
	}
	return
}

func createRandomWishListItemForUpdate(t *testing.T, wishList db.WishList) (wishListItem db.WishListItem) {
	wishListItem = db.WishListItem{
		ID:            util.RandomMoney(),
		WishListID:    wishList.ID,
		ProductItemID: util.RandomMoney(),
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

func requireBodyMatchWishListItem(t *testing.T, body io.ReadCloser, wishListItem db.WishListItem) {
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

func requireBodyMatchWishListItemForGet(t *testing.T, body io.ReadCloser, wishListItem db.WishListItem) {
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

func requireBodyMatchListWishListItem(t *testing.T, body io.ReadCloser, WishListItem []db.ListWishListItemsByUserIDRow) {
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
