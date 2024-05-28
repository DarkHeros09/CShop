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

func TestGetHomePageTextBannerAPI(t *testing.T) {
	textBanner := randomHomePageTextBanner()

	testCases := []struct {
		name                 string
		AdminID              int64
		HomePageTextBannerID int64
		buildStub            func(store *mockdb.MockStore)
		checkResponse        func(t *testing.T, rsp *http.Response)
	}{
		{
			name: "OK",

			HomePageTextBannerID: textBanner.ID,
			buildStub: func(store *mockdb.MockStore) {
				// arg := db.GetHomePageTextBannerParams{
				// 	HomePageTextBannerID:     textBanner.ID,

				// }
				store.EXPECT().
					GetHomePageTextBanner(gomock.Any(), gomock.Eq(textBanner.ID)).
					Times(1).
					Return(textBanner, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchHomePageTextBanner(t, rsp.Body, textBanner)
			},
		},

		{
			name: "NotFound",

			HomePageTextBannerID: textBanner.ID,
			buildStub: func(store *mockdb.MockStore) {
				// arg := db.GetHomePageTextBannerParams{
				// 	HomePageTextBannerID:     textBanner.ID,

				// }
				store.EXPECT().
					GetHomePageTextBanner(gomock.Any(), gomock.Eq(textBanner.ID)).
					Times(1).
					Return(db.HomePageTextBanner{}, pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
			},
		},
		{
			name: "InternalError",

			HomePageTextBannerID: textBanner.ID,
			buildStub: func(store *mockdb.MockStore) {
				// arg := db.GetHomePageTextBannerParams{
				// 	HomePageTextBannerID:     textBanner.ID,

				// }
				store.EXPECT().
					GetHomePageTextBanner(gomock.Any(), gomock.Eq(textBanner.ID)).
					Times(1).
					Return(db.HomePageTextBanner{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:                 "InvalidID",
			HomePageTextBannerID: 0,
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetHomePageTextBanner(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/api/v1/text-banners/%d", tc.HomePageTextBannerID)
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(t, rsp)
		})

	}

}

func TestCreateHomePageTextBannerAPI(t *testing.T) {
	admin, _ := randomHomePageTextBannerSuperAdmin(t)
	textBanner := randomHomePageTextBanner()

	testCases := []struct {
		name          string
		body          fiber.Map
		AdminID       int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(rsp *http.Response)
	}{
		{
			name:    "OK",
			AdminID: admin.ID,
			body: fiber.Map{
				"name":        textBanner.Name,
				"description": textBanner.Description,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateHomePageTextBannerParams{
					Name:        textBanner.Name,
					Description: textBanner.Description,
				}

				store.EXPECT().
					CreateHomePageTextBanner(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(textBanner, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchHomePageTextBanner(t, rsp.Body, textBanner)
			},
		},
		{
			name:    "NoAuthorization",
			AdminID: admin.ID,
			body: fiber.Map{
				"name":        textBanner.Name,
				"description": textBanner.Description,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateHomePageTextBannerParams{
					Name:        textBanner.Name,
					Description: textBanner.Description,
				}

				store.EXPECT().
					CreateHomePageTextBanner(gomock.Any(), gomock.Eq(arg)).
					Times(0)

			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:    "Unauthorized",
			AdminID: admin.ID,
			body: fiber.Map{
				"name":        textBanner.Name,
				"description": textBanner.Description,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, false, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateHomePageTextBannerParams{
					Name:        textBanner.Name,
					Description: textBanner.Description,
				}

				store.EXPECT().
					CreateHomePageTextBanner(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:    "InternalError",
			AdminID: admin.ID,
			body: fiber.Map{
				"name":        textBanner.Name,
				"description": textBanner.Description,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateHomePageTextBanner(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.HomePageTextBanner{}, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:    "InvalidID",
			AdminID: 0,
			body: fiber.Map{
				"name":        nil,
				"description": textBanner.Description,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateHomePageTextBanner(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/admin/%d/v1/text-banners", tc.AdminID)
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

func TestListHomePageTextBannersAPI(t *testing.T) {
	n := 5
	textBanners := make([]db.HomePageTextBanner, n)
	for i := 0; i < n; i++ {
		textBanners[i] = randomHomePageTextBanner()
	}

	testCases := []struct {
		name          string
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(rsp *http.Response)
	}{
		{
			name: "OK",
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					ListHomePageTextBanners(gomock.Any()).
					Times(1).
					Return(textBanners, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchHomePageTextBanners(t, rsp.Body, textBanners)
			},
		},
		{
			name: "InternalError",

			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListHomePageTextBanners(gomock.Any()).
					Times(1).
					Return([]db.HomePageTextBanner{}, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
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

			url := "/api/v1/text-banners"
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			// // Add query parameters to request URL
			// q := request.URL.Query()
			// q.Add("page_id", fmt.Sprintf("%d", tc.query.pageID))
			// q.Add("page_size", fmt.Sprintf("%d", tc.query.pageSize))
			// request.URL.RawQuery = q.Encode()

			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
		})
	}
}

func TestUpdateHomePageTextBannerAPI(t *testing.T) {
	admin, _ := randomHomePageTextBannerSuperAdmin(t)
	textBanner := randomHomePageTextBanner()

	changedName := util.RandomUser()

	testCases := []struct {
		name                 string
		body                 fiber.Map
		HomePageTextBannerID int64
		AdminID              int64
		setupAuth            func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs           func(store *mockdb.MockStore)
		checkResponse        func(t *testing.T, rsp *http.Response)
	}{
		{
			name:                 "OK",
			HomePageTextBannerID: textBanner.ID,
			AdminID:              admin.ID,
			body: fiber.Map{
				"name":        changedName,
				"description": textBanner.Description,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateHomePageTextBannerParams{
					ID:          textBanner.ID,
					Name:        null.StringFrom(changedName),
					Description: null.StringFrom(textBanner.Description),
				}

				store.EXPECT().
					UpdateHomePageTextBanner(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(textBanner, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:                 "Unauthorized",
			HomePageTextBannerID: textBanner.ID,

			AdminID: admin.ID,
			body: fiber.Map{
				"name":        changedName,
				"description": textBanner.Description,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, false, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateHomePageTextBannerParams{
					ID:          textBanner.ID,
					Name:        null.StringFrom(changedName),
					Description: null.StringFrom(textBanner.Description),
				}

				store.EXPECT().
					UpdateHomePageTextBanner(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:                 "NoAuthorization",
			HomePageTextBannerID: textBanner.ID,

			AdminID: admin.ID,
			body: fiber.Map{
				"name":        changedName,
				"description": textBanner.Description,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateHomePageTextBannerParams{
					ID:          textBanner.ID,
					Name:        null.StringFrom(changedName),
					Description: null.StringFrom(textBanner.Description),
				}

				store.EXPECT().
					UpdateHomePageTextBanner(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:                 "InternalError",
			HomePageTextBannerID: textBanner.ID,
			AdminID:              admin.ID,
			body: fiber.Map{
				"name":        changedName,
				"description": textBanner.Description,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateHomePageTextBannerParams{
					ID:          textBanner.ID,
					Name:        null.StringFrom(changedName),
					Description: null.StringFrom(textBanner.Description),
				}
				store.EXPECT().
					UpdateHomePageTextBanner(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.HomePageTextBanner{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:                 "InvalidID",
			HomePageTextBannerID: 0,
			AdminID:              admin.ID,
			body: fiber.Map{
				"name":        changedName,
				"description": textBanner.Description,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateHomePageTextBanner(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/admin/%d/v1/text-banners/%d", tc.AdminID, tc.HomePageTextBannerID)
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

func TestDeleteHomePageTextBannerAPI(t *testing.T) {
	admin, _ := randomHomePageTextBannerSuperAdmin(t)
	textBanner := randomHomePageTextBanner()

	testCases := []struct {
		name                 string
		HomePageTextBannerID int64
		AdminID              int64
		setupAuth            func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub            func(store *mockdb.MockStore)
		checkResponse        func(t *testing.T, rsp *http.Response)
	}{
		{
			name:                 "OK",
			HomePageTextBannerID: textBanner.ID,
			AdminID:              admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				// arg := db.DeleteHomePageTextBannerParams{
				// 	HomePageTextBannerID:     textBanner.ID,

				// }
				store.EXPECT().
					DeleteHomePageTextBanner(gomock.Any(), gomock.Eq(textBanner.ID)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name: "Unauthorized",

			HomePageTextBannerID: textBanner.ID,
			AdminID:              admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, 2, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				// arg := db.DeleteHomePageTextBannerParams{
				// 	HomePageTextBannerID:     textBanner.ID,

				// }
				store.EXPECT().
					DeleteHomePageTextBanner(gomock.Any(), gomock.Eq(textBanner.ID)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name: "NoAuthorization",

			HomePageTextBannerID: textBanner.ID,
			AdminID:              admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStub: func(store *mockdb.MockStore) {
				// arg := db.DeleteHomePageTextBannerParams{
				// 	HomePageTextBannerID:     textBanner.ID,

				// }
				store.EXPECT().
					DeleteHomePageTextBanner(gomock.Any(), gomock.Eq(textBanner.ID)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name: "NotFound",

			HomePageTextBannerID: textBanner.ID,
			AdminID:              admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				// arg := db.DeleteHomePageTextBannerParams{
				// 	HomePageTextBannerID:     textBanner.ID,

				// }
				store.EXPECT().
					DeleteHomePageTextBanner(gomock.Any(), gomock.Eq(textBanner.ID)).
					Times(1).
					Return(pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusNotFound, rsp.StatusCode)
			},
		},
		{
			name: "InternalError",

			HomePageTextBannerID: textBanner.ID,
			AdminID:              admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				// arg := db.DeleteHomePageTextBannerParams{
				// 	HomePageTextBannerID:     textBanner.ID,

				// }
				store.EXPECT().
					DeleteHomePageTextBanner(gomock.Any(), gomock.Eq(textBanner.ID)).
					Times(1).
					Return(pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:                 "InvalidID",
			HomePageTextBannerID: 0,
			AdminID:              admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteHomePageTextBanner(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/admin/%d/v1/text-banners/%d", tc.AdminID, tc.HomePageTextBannerID)
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

func randomHomePageTextBannerSuperAdmin(t *testing.T) (admin db.Admin, password string) {
	password = util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	admin = db.Admin{
		ID:       util.RandomMoney(),
		Username: util.RandomUser(),
		Email:    util.RandomEmail(),
		Password: hashedPassword,
		Active:   true,
		TypeID:   1,
	}
	return
}

func randomHomePageTextBanner() db.HomePageTextBanner {
	return db.HomePageTextBanner{
		ID:          util.RandomMoney(),
		Name:        util.RandomUser(),
		Description: util.RandomUser(),
	}
}

func requireBodyMatchHomePageTextBanner(t *testing.T, body io.ReadCloser, textBanner db.HomePageTextBanner) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotHomePageTextBanner db.HomePageTextBanner
	err = json.Unmarshal(data, &gotHomePageTextBanner)
	require.NoError(t, err)
	require.Equal(t, textBanner, gotHomePageTextBanner)
}

func requireBodyMatchHomePageTextBanners(t *testing.T, body io.ReadCloser, HomePageTextBanners []db.HomePageTextBanner) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotHomePageTextBanners []db.HomePageTextBanner
	err = json.Unmarshal(data, &gotHomePageTextBanners)
	require.NoError(t, err)
	require.Equal(t, HomePageTextBanners, gotHomePageTextBanners)
}
