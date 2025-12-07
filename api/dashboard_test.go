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
	mockemail "github.com/cshop/v3/mail/mock"
	"github.com/cshop/v3/token"
	"github.com/cshop/v3/util"
	mockwk "github.com/cshop/v3/worker/mock"
	"github.com/gofiber/fiber/v2"
	"github.com/guregu/null/v6"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetDashboardInfoAPI(t *testing.T) {
	dashboardInfo := randomDashbosrdInfo()
	admin, _ := randomSuperAdmin(t)
	user, _ := randomUser(t)

	testCases := []struct {
		name          string
		AdminID       int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rsp *http.Response)
	}{
		{
			name:    "OK",
			AdminID: admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetActiveProductItems(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(int64(dashboardInfo.ActiveProducts), nil)
				store.EXPECT().
					GetTotalProductItems(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(int64(dashboardInfo.TotalProducts), nil)
				store.EXPECT().
					GetActiveUsersCount(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(int64(dashboardInfo.ActiveUsers), nil)
				store.EXPECT().
					GetTotalUsersCount(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(int64(dashboardInfo.TotalUsers), nil)
				arg := db.GetShopOrdersCountByStatusIdParams{
					OrderStatusID: null.IntFrom(2),
					AdminID:       admin.ID,
				}
				store.EXPECT().
					GetShopOrdersCountByStatusId(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(int64(dashboardInfo.ActiveOrders), nil)
				store.EXPECT().
					GetTotalShopOrder(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(int64(dashboardInfo.TotalOrders), nil)
				store.EXPECT().
					GetCompletedDailyOrderTotal(gomock.Any(), gomock.Eq(admin.ID)).
					Times(1).
					Return(dashboardInfo.TodayRevenue, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchDashboard(t, rsp.Body, dashboardInfo)
			},
		},
		{
			name:    "Unauthorized",
			AdminID: user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetActiveProductItems(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:    "InternalError",
			AdminID: admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetActiveProductItems(gomock.Any(), gomock.Any()).
					Times(1).
					Return(int64(0), pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:    "InvalidID",
			AdminID: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetActiveProductItems(gomock.Any(), gomock.Any()).
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
			ik := mockik.NewMockImageKitManagement(ctrl)
			mailSender := mockemail.NewMockEmailSender(ctrl)

			// build stubs
			tc.buildStub(store)

			// start test server and send request
			server := newTestServer(t, store, worker, ik, mailSender)
			//recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/admin/v1/admins/%d/dashboard", tc.AdminID)
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.adminTokenMaker)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			//check response
			tc.checkResponse(t, rsp)
		})

	}

}

func randomDashbosrdInfo() (dashboard dashboardJsonResponse) {
	dashboard = dashboardJsonResponse{
		ActiveProducts: int32(util.RandomInt(1, 100)),
		TotalProducts:  int32(util.RandomInt(1, 100)),
		ActiveUsers:    int32(util.RandomInt(1, 100)),
		TotalUsers:     int32(util.RandomInt(1, 100)),
		ActiveOrders:   int32(util.RandomInt(1, 100)),
		TotalOrders:    int32(util.RandomInt(1, 100)),
		TodayRevenue:   util.RandomDecimalString(1, 1000),
		MonthRevenue:   int32(util.RandomInt(1, 100)),
	}
	return
}

func requireBodyMatchDashboard(t *testing.T, body io.Reader, rsp dashboardJsonResponse) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotDashboard dashboardJsonResponse
	err = json.Unmarshal(data, &gotDashboard)

	require.NoError(t, err)
	require.Equal(t, rsp.ActiveOrders, gotDashboard.ActiveOrders)
	require.Equal(t, rsp.TotalOrders, gotDashboard.TotalOrders)
	require.Equal(t, rsp.ActiveProducts, gotDashboard.ActiveProducts)
	require.Equal(t, rsp.TotalProducts, gotDashboard.TotalProducts)
	require.Equal(t, rsp.ActiveUsers, gotDashboard.ActiveUsers)
	require.Equal(t, rsp.TotalUsers, gotDashboard.TotalUsers)
	require.Equal(t, rsp.ActiveOrders, gotDashboard.ActiveOrders)
	require.Equal(t, rsp.TodayRevenue, gotDashboard.TodayRevenue)

}
