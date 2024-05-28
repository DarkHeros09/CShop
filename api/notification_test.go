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
	mockwk "github.com/cshop/v3/worker/mock"
	"github.com/gofiber/fiber/v2"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCreateNotificationAPI(t *testing.T) {
	user, _ := randomNotificationUser(t)
	notification := createRandomNotification(t, user)

	testCases := []struct {
		name          string
		body          fiber.Map
		UserID        int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(rsp *http.Response)
	}{
		{
			name:   "OK",
			UserID: user.ID,
			body: fiber.Map{
				"device_id": notification.DeviceID.String,
				"fcm_token": notification.FcmToken.String,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.CreateNotificationParams{
					UserID:   user.ID,
					DeviceID: notification.DeviceID,
					FcmToken: notification.FcmToken,
				}

				store.EXPECT().
					CreateNotification(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(notification, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchNotification(t, rsp.Body, notification)
			},
		},
		{
			name:   "NoAuthorization",
			UserID: user.ID,
			body: fiber.Map{
				"device_id": notification.DeviceID.String,
				"fcm_token": notification.FcmToken.String,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					CreateNotification(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:   "InternalError",
			UserID: user.ID,
			body: fiber.Map{
				"device_id": notification.DeviceID.String,
				"fcm_token": notification.FcmToken.String,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.CreateNotificationParams{
					UserID:   user.ID,
					DeviceID: notification.DeviceID,
					FcmToken: notification.FcmToken,
				}

				store.EXPECT().
					CreateNotification(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.Notification{}, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:   "InvalidUserID",
			UserID: 0,
			body: fiber.Map{
				"device_id": notification.DeviceID.String,
				"fcm_token": notification.FcmToken.String,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					CreateNotification(gomock.Any(), gomock.Any()).
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
			tc.buildStubs(store)

			server := newTestServer(t, store, worker)
			//recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/usr/v1/users/%d/notification", tc.UserID)
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

func TestGetNotificationAPI(t *testing.T) {
	user, _ := randomNotificationUser(t)
	notification := createRandomNotification(t, user)

	testCases := []struct {
		name          string
		UserID        int64
		DeviceID      string
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(rsp *http.Response)
	}{
		{
			name:     "OK",
			DeviceID: notification.DeviceID.String,
			UserID:   user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.GetNotificationParams{
					UserID:   user.ID,
					DeviceID: notification.DeviceID,
				}

				store.EXPECT().
					GetNotification(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(notification, nil)

			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchNotification(t, rsp.Body, notification)
			},
		},
		{
			name:     "NoAuthorization",
			DeviceID: notification.DeviceID.String,
			UserID:   user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetNotification(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:     "InternalError",
			DeviceID: notification.DeviceID.String,
			UserID:   user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.GetNotificationParams{
					UserID:   user.ID,
					DeviceID: notification.DeviceID,
				}

				store.EXPECT().
					GetNotification(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.Notification{}, pgx.ErrTxClosed)

			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:     "InvalidUserID",
			DeviceID: notification.DeviceID.String,
			UserID:   0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetNotification(gomock.Any(), gomock.Any()).
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
			tc.buildStubs(store)

			server := newTestServer(t, store, worker)
			//recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/usr/v1/users/%d/notification/%s", tc.UserID, tc.DeviceID)
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

func TestUpdateNotificationAPI(t *testing.T) {
	user, _ := randomNotificationUser(t)
	notification := createRandomNotification(t, user)
	fcmToken := util.RandomString(10)

	testCases := []struct {
		name          string
		DeviceID      string
		UserID        int64
		body          fiber.Map
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rsp *http.Response)
	}{
		{
			name:     "OK",
			DeviceID: notification.DeviceID.String,
			UserID:   user.ID,
			body: fiber.Map{
				"fcm_token": fcmToken,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.UpdateNotificationParams{
					DeviceID: notification.DeviceID,
					UserID:   user.ID,
					FcmToken: null.StringFrom(fcmToken),
				}

				store.EXPECT().
					UpdateNotification(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(notification, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:     "NoAuthorization",
			DeviceID: notification.DeviceID.String,
			UserID:   user.ID,
			body: fiber.Map{
				"fcm_token": fcmToken,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateNotification(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:     "InternalError",
			DeviceID: notification.DeviceID.String,
			UserID:   user.ID,
			body: fiber.Map{
				"fcm_token": fcmToken,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateNotificationParams{
					DeviceID: notification.DeviceID,
					UserID:   user.ID,
					FcmToken: null.StringFrom(fcmToken),
				}

				store.EXPECT().
					UpdateNotification(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.Notification{}, pgx.ErrTxClosed)

			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:     "InvalidDeviceID",
			DeviceID: notification.DeviceID.String,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateNotification(gomock.Any(), gomock.Any()).
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
			worker := mockwk.NewMockTaskDistributor(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store, worker)
			//recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/usr/v1/users/%d/notification/%s", tc.UserID, tc.DeviceID)
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

func TestDeleteNotificationAPI(t *testing.T) {
	user, _ := randomNotificationUser(t)
	notification := createRandomNotification(t, user)

	testCases := []struct {
		name          string
		DeviceID      string
		UserID        int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rsp *http.Response)
	}{
		{
			name:     "OK",
			UserID:   user.ID,
			DeviceID: notification.DeviceID.String,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteNotificationParams{
					UserID:   notification.UserID,
					DeviceID: notification.DeviceID,
				}

				store.EXPECT().
					DeleteNotification(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(notification, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:     "NotFound",
			UserID:   user.ID,
			DeviceID: notification.DeviceID.String,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteNotificationParams{
					UserID:   notification.UserID,
					DeviceID: notification.DeviceID,
				}

				store.EXPECT().
					DeleteNotification(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.Notification{}, pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
			},
		},
		{
			name:     "InternalError",
			UserID:   user.ID,
			DeviceID: notification.DeviceID.String,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.DeleteNotificationParams{
					UserID:   notification.UserID,
					DeviceID: notification.DeviceID,
				}

				store.EXPECT().
					DeleteNotification(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.Notification{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:     "InvalidID",
			UserID:   0,
			DeviceID: notification.DeviceID.String,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteNotification(gomock.Any(), gomock.Any()).
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
			worker := mockwk.NewMockTaskDistributor(ctrl)

			// build stubs
			tc.buildStub(store)

			// start test server and send request
			server := newTestServer(t, store, worker)
			//recorder := httptest.NewRecorder()

			// // Marshal body data to JSON
			// data, err := json.Marshal(tc.body)
			// require.NoError(t, err)

			url := fmt.Sprintf("/usr/v1/users/%d/notification/%s", tc.UserID, tc.DeviceID)
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

// func TestDeleteNotificationAllByUserAPI(t *testing.T) {
// 	user, _ := randomNotificationUser(t)
// 	shoppingCart := createRandomShoppingCart(t, user)
// 	notification := createRandomNotification(t, shoppingCart)

// 	var NotificationList []db.Notification
// 	NotificationList = append(NotificationList, notification...)

// 	testCases := []struct {
// 		name           string
// 		UserID         int64
// 		DeviceID int64
// 		setupAuth      func(t *testing.T, request *http.Request, tokenMaker token.Maker)
// 		buildStub      func(store *mockdb.MockStore)
// 		checkResponse  func(t *testing.T, rsp *http.Response)
// 	}{
// 		{
// 			name:           "OK",
// 			UserID:         user.ID,
// 			DeviceID: notification.DeviceID.String,
// 			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
// 				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
// 			},
// 			buildStub: func(store *mockdb.MockStore) {

// 				arg := db.DeleteNotificationAllByUserParams{
// 					UserID:         user.ID,
// 					DeviceID: notification.DeviceID.String,
// 				}
// 				store.EXPECT().
// 					DeleteNotificationAllByUser(gomock.Any(), gomock.Eq(arg)).
// 					Times(1).
// 					Return(notificationList, nil)
// 			},
// 			checkResponse: func(t *testing.T, rsp *http.Response) {
// 				require.Equal(t, http.StatusOK, rsp.StatusCode)
// 			},
// 		},
// 		{
// 			name:           "NotFound",
// 			UserID:         user.ID,
// 			DeviceID: notification.DeviceID.String,
// 			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
// 				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
// 			},
// 			buildStub: func(store *mockdb.MockStore) {
// 				arg := db.DeleteNotificationAllByUserParams{
// 					UserID:         user.ID,
// 					DeviceID: notification.DeviceID.String,
// 				}
// 				store.EXPECT().
// 					DeleteNotificationAllByUser(gomock.Any(), gomock.Eq(arg)).
// 					Times(1).
// 					Return([]db.Notification{}, pgx.ErrNoRows)
// 			},
// 			checkResponse: func(t *testing.T, rsp *http.Response) {
// 				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
// 			},
// 		},
// 		{
// 			name:           "InternalError",
// 			UserID:         user.ID,
// 			DeviceID: notification.DeviceID.String,
// 			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
// 				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, user.Username, time.Minute)
// 			},
// 			buildStub: func(store *mockdb.MockStore) {
// 				arg := db.DeleteNotificationAllByUserParams{
// 					UserID:         user.ID,
// 					DeviceID: notification.DeviceID.String,
// 				}
// 				store.EXPECT().
// 					DeleteNotificationAllByUser(gomock.Any(), gomock.Eq(arg)).
// 					Times(1).
// 					Return([]db.Notification{}, pgx.ErrTxClosed)
// 			},
// 			checkResponse: func(t *testing.T, rsp *http.Response) {
// 				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
// 			},
// 		},
// 		{
// 			name:           "InvalidID",
// 			UserID:         0,
// 			DeviceID: notification.DeviceID.String,
// 			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
// 				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, 0, user.Username, time.Minute)
// 			},
// 			buildStub: func(store *mockdb.MockStore) {
// 				store.EXPECT().
// 					DeleteNotification(gomock.Any(), gomock.Any()).
// 					Times(0)

// 			},
// 			checkResponse: func(t *testing.T, rsp *http.Response) {
// 				require.Equal(t, http.StatusBadRequest, rsp.StatusCode)
// 			},
// 		},
// 	}

// 	for i := range testCases {
// 		tc := testCases[i]

// 		t.Run(tc.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t) // no need to call defer ctrl.finish() in 1.6V

// 			store := mockdb.NewMockStore(ctrl)
// worker := mockwk.NewMockTaskDistributor(ctrl)

// 			// build stubs
// 			tc.buildStub(store)

// 			// start test server and send request
// 			server := newTestServer(t, store, worker)
// 			//recorder := httptest.NewRecorder()

// 			url := fmt.Sprintf("/usr/v1/users/%d/notification/%d", tc.UserID, tc.DeviceID)
// 			request, err := http.NewRequest(fiber.MethodDelete, url, nil)
// 			require.NoError(t, err)

// 			tc.setupAuth(t, request, server.tokenMaker)
// 			request.Header.Set("Content-Type", "application/json")

// 			rsp, err := server.router.Test(request)
// 			require.NoError(t, err)
// 			tc.checkResponse(t, rsp)
// 		})

// 	}

// }

func randomNotificationUser(t *testing.T) (user db.User, password string) {
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

func createRandomNotification(t *testing.T, user db.User) (Notification db.Notification) {
	Notification = db.Notification{
		UserID:   user.ID,
		DeviceID: null.StringFrom(util.RandomString(10)),
		FcmToken: null.StringFrom(util.RandomString(10)),
	}
	return
}

func requireBodyMatchNotification(t *testing.T, body io.ReadCloser, notification db.Notification) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotNotification db.Notification
	err = json.Unmarshal(data, &gotNotification)

	require.NoError(t, err)
	require.Equal(t, notification.UserID, gotNotification.UserID)
	require.Equal(t, notification.DeviceID, gotNotification.DeviceID)
	require.Equal(t, notification.FcmToken, gotNotification.FcmToken)
	require.Equal(t, notification.CreatedAt, gotNotification.CreatedAt)
	require.Equal(t, notification.CreatedAt, gotNotification.CreatedAt)
}
