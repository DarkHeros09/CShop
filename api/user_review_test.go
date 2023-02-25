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

func TestCreateUserReviewAPI(t *testing.T) {
	user, _ := randomURUser(t)
	userReview := createRandomUserReview(t, user)

	testCases := []struct {
		name          string
		body          fiber.Map
		ID            int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(rsp *http.Response)
	}{
		{
			name: "OK",
			ID:   user.ID,
			body: fiber.Map{
				"ordered_product_id": userReview.OrderedProductID,
				"rating_value":       userReview.RatingValue,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.CreateUserReviewParams{
					UserID:           user.ID,
					OrderedProductID: userReview.OrderedProductID,
					RatingValue:      userReview.RatingValue,
				}

				store.EXPECT().
					CreateUserReview(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(userReview, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchUserReview(t, rsp.Body, userReview)
			},
		},
		{
			name: "NoAuthorization",
			ID:   user.ID,
			body: fiber.Map{
				"ordered_product_id": userReview.OrderedProductID,
				"rating_value":       userReview.RatingValue,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUserReview(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name: "InternalError",
			ID:   user.ID,
			body: fiber.Map{
				"ordered_product_id": userReview.OrderedProductID,
				"rating_value":       userReview.RatingValue,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateUserReviewParams{
					UserID:           user.ID,
					OrderedProductID: userReview.OrderedProductID,
					RatingValue:      userReview.RatingValue,
				}

				store.EXPECT().
					CreateUserReview(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.UserReview{}, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name: "InvalidUserID",
			ID:   0,
			body: fiber.Map{
				"ordered_product_id": userReview.OrderedProductID,
				"rating_value":       userReview.RatingValue,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUserReview(gomock.Any(), gomock.Any()).
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
			// //recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/usr/v1/users/%d/reviews", tc.ID)
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

func TestGetUserReviewAPI(t *testing.T) {
	user, _ := randomURUser(t)
	userReview := createRandomUserReview(t, user)

	testCases := []struct {
		name          string
		ID            int64
		UserID        int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rsp *http.Response)
	}{
		{
			name:   "OK",
			ID:     userReview.ID,
			UserID: user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.GetUserReviewParams{
					UserID: userReview.UserID,
					ID:     userReview.ID,
				}
				store.EXPECT().
					GetUserReview(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(userReview, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchUserReview(t, rsp.Body, userReview)
			},
		},
		{
			name:   "NoAuthorization",
			ID:     userReview.ID,
			UserID: user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserReview(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:   "NotFound",
			ID:     userReview.ID,
			UserID: user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.GetUserReviewParams{
					UserID: userReview.UserID,
					ID:     userReview.ID,
				}
				store.EXPECT().
					GetUserReview(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.UserReview{}, pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
			},
		},
		{
			name:   "InternalError",
			ID:     userReview.ID,
			UserID: user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.GetUserReviewParams{
					UserID: userReview.UserID,
					ID:     userReview.ID,
				}
				store.EXPECT().
					GetUserReview(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.UserReview{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:   "InvalidID",
			ID:     0,
			UserID: user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserReview(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/usr/v1/users/%d/reviews/%d", tc.UserID, tc.ID)
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(t, rsp)
		})

	}

}

func TestListUsersReviewAPI(t *testing.T) {
	n := 5
	userReviews := make([]db.UserReview, n)
	user, _ := randomURUser(t)
	userReview1 := createRandomUserReview(t, user)
	userReview2 := createRandomUserReview(t, user)
	userReview3 := createRandomUserReview(t, user)

	userReviews = append(userReviews, userReview1, userReview2, userReview3)

	type Query struct {
		pageID   int
		pageSize int
	}

	testCases := []struct {
		name          string
		query         Query
		ID            int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(rsp *http.Response)
	}{
		{
			name: "OK",
			ID:   user.ID,
			query: Query{
				pageID:   1,
				pageSize: n,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListUserReviewsParams{
					UserID: user.ID,
					Limit:  int32(n),
					Offset: 0,
				}

				store.EXPECT().
					ListUserReviews(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(userReviews, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchUserReviews(t, rsp.Body, userReviews)
			},
		},
		{
			name: "InternalError",
			ID:   user.ID,
			query: Query{
				pageID:   1,
				pageSize: n,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListUserReviews(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.UserReview{}, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name: "InvalidPageID",
			ID:   user.ID,
			query: Query{
				pageID:   -1,
				pageSize: n,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListUserReviews(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusBadRequest, rsp.StatusCode)
			},
		},
		{
			name: "InvalidPageSize",
			ID:   user.ID,
			query: Query{
				pageID:   1,
				pageSize: 100000,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListUserReviews(gomock.Any(), gomock.Any()).
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
			// //recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/usr/v1/users/%d/reviews", tc.ID)
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

func TestUpdateUserReviewAPI(t *testing.T) {
	user, _ := randomURUser(t)
	userReview := createRandomUserReview(t, user)

	testCases := []struct {
		name          string
		ID            int64
		UserID        int64
		body          fiber.Map
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rsp *http.Response)
	}{
		{
			name:   "OK",
			ID:     userReview.ID,
			UserID: userReview.UserID,
			body: fiber.Map{
				"ordered_product_id": userReview.OrderedProductID,
				"rating_value":       3,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.UpdateUserReviewParams{
					UserID:           userReview.UserID,
					OrderedProductID: null.IntFrom(userReview.OrderedProductID),
					RatingValue:      null.IntFrom(3),
					ID:               userReview.ID,
				}

				store.EXPECT().
					UpdateUserReview(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(userReview, nil)

			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:   "NoAuthorization",
			ID:     userReview.ID,
			UserID: userReview.UserID,
			body: fiber.Map{
				"ordered_product_id": userReview.OrderedProductID,
				"rating_value":       3,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateUserReview(gomock.Any(), gomock.Any()).
					Times(0)

			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:   "InternalError",
			ID:     userReview.ID,
			UserID: userReview.UserID,
			body: fiber.Map{
				"ordered_product_id": userReview.OrderedProductID,
				"rating_value":       3,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateUserReviewParams{
					UserID:           userReview.UserID,
					OrderedProductID: null.IntFrom(userReview.OrderedProductID),
					RatingValue:      null.IntFrom(3),
					ID:               userReview.ID,
				}

				store.EXPECT().
					UpdateUserReview(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.UserReview{}, pgx.ErrTxClosed)

			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:   "InvalidUserID",
			ID:     0,
			UserID: user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateUserReview(gomock.Any(), gomock.Any()).
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
			// //recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/usr/v1/users/%d/reviews/%d", tc.UserID, tc.ID)
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

func TestDeleteUserReviewAPI(t *testing.T) {
	user, _ := randomURUser(t)
	userReview := createRandomUserReview(t, user)

	testCases := []struct {
		name          string
		ID            int64
		UserID        int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rsp *http.Response)
	}{
		{
			name:   "OK",
			ID:     userReview.ID,
			UserID: user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteUserReviewParams{
					UserID: userReview.UserID,
					ID:     userReview.ID,
				}

				store.EXPECT().
					DeleteUserReview(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(userReview, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:   "NotFound",
			ID:     userReview.ID,
			UserID: user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				// arg := db.DeleteUserReviewParams{
				// 	UserID: userReview.UserID,
				// 	ID:     userReview.ID,
				// }
				store.EXPECT().
					DeleteUserReview(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.UserReview{}, pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
			},
		},
		{
			name:   "InternalError",
			ID:     userReview.ID,
			UserID: user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				// arg := db.DeleteUserReviewParams{
				// 	UserID: userReview.UserID,
				// 	ID:     userReview.ID,
				// }
				store.EXPECT().
					DeleteUserReview(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.UserReview{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:   "InvalidID",
			ID:     0,
			UserID: user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteUserReview(gomock.Any(), gomock.Any()).
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
			// //recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/usr/v1/users/%d/reviews/%d", tc.UserID, tc.ID)
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

func randomURUser(t *testing.T) (user db.User, password string) {
	password = util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	user = db.User{
		ID:        util.RandomMoney(),
		Username:  util.RandomUser(),
		Email:     util.RandomEmail(),
		Password:  hashedPassword,
		Telephone: int32(util.RandomInt(910000000, 929999999)),
	}
	return
}

func createRandomUserReview(t *testing.T, user db.User) (userReview db.UserReview) {
	userReview = db.UserReview{
		ID:               util.RandomMoney(),
		UserID:           user.ID,
		OrderedProductID: util.RandomMoney(),
		RatingValue:      2,
	}
	return
}

func requireBodyMatchUserReview(t *testing.T, body io.ReadCloser, userReview db.UserReview) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotUserReview db.UserReview
	err = json.Unmarshal(data, &gotUserReview)

	require.NoError(t, err)
	require.Equal(t, userReview.UserID, gotUserReview.UserID)
	require.Equal(t, userReview.ID, gotUserReview.ID)
	require.Equal(t, userReview.OrderedProductID, gotUserReview.OrderedProductID)
	require.Equal(t, userReview.RatingValue, gotUserReview.RatingValue)
}

func requireBodyMatchUserReviews(t *testing.T, body io.ReadCloser, userReviews []db.UserReview) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotUserReviews []db.UserReview
	err = json.Unmarshal(data, &gotUserReviews)
	require.NoError(t, err)
	require.Equal(t, userReviews, gotUserReviews)
}
