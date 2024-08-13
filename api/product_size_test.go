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
	mockik "github.com/cshop/v3/image/mock"
	"github.com/cshop/v3/token"
	"github.com/cshop/v3/util"
	mockwk "github.com/cshop/v3/worker/mock"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCreateProductSizeAPI(t *testing.T) {
	admin, _ := randomPCategorieSuperAdmin(t)
	productSize := randomProductSize()

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
				"size_value": productSize.SizeValue,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminCreateProductSizeParams{
					AdminID:   admin.ID,
					SizeValue: productSize.SizeValue,
				}

				store.EXPECT().
					AdminCreateProductSize(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(productSize, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchProductSize(t, rsp.Body, productSize)
			},
		},
		{
			name:    "NoAuthorization",
			AdminID: admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminCreateProductSizeParams{
					AdminID:   admin.ID,
					SizeValue: productSize.SizeValue,
				}

				store.EXPECT().
					AdminCreateProductSize(gomock.Any(), gomock.Eq(arg)).
					Times(0)

			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:    "Unauthorized",
			AdminID: admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, false, time.Minute)
			},
			body: fiber.Map{
				"size_value": productSize.SizeValue,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminCreateProductSizeParams{
					AdminID:   admin.ID,
					SizeValue: productSize.SizeValue,
				}

				store.EXPECT().
					AdminCreateProductSize(gomock.Any(), gomock.Eq(arg)).
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
				"size_value": productSize.SizeValue,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					AdminCreateProductSize(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.ProductSize{}, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:    "InvalidID",
			AdminID: admin.ID,
			body: fiber.Map{
				"id": "invalid",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					AdminCreateProductSize(gomock.Any(), gomock.Any()).
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

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/admin/v1/admins/%d/sizes", tc.AdminID)
			request, err := http.NewRequest(fiber.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.adminTokenMaker)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
		})
	}
}

func TestListProductSizeSizesAPI(t *testing.T) {
	n := 5
	productSizes := make([]db.ProductSize, n)
	for i := 0; i < n; i++ {
		productSizes[i] = randomProductSizesForList()
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
					ListProductSizes(gomock.Any()).
					Times(1).
					Return(productSizes, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchProductSizesList(t, rsp.Body, productSizes)
			},
		},
		{
			name: "InternalError",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListProductSizes(gomock.Any()).
					Times(1).
					Return([]db.ProductSize{}, pgx.ErrTxClosed)
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
			ik := mockik.NewMockImageKitManagement(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store, worker, ik)
			//recorder := httptest.NewRecorder()

			url := "/api/v1/sizes"
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
		})
	}
}

func randomProductSizesForList() db.ProductSize {
	return db.ProductSize{
		ID:        util.RandomInt(1, 1000),
		SizeValue: util.RandomSize(),
	}
}

func requireBodyMatchProductSizesList(t *testing.T, body io.ReadCloser, productSizes []db.ProductSize) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotProducts []db.ProductSize
	err = json.Unmarshal(data, &gotProducts)
	require.NoError(t, err)
	require.Equal(t, productSizes, gotProducts)
}

func randomProductSize() db.ProductSize {
	return db.ProductSize{
		ID:        util.RandomInt(0, 1000),
		SizeValue: util.RandomSize(),
	}
}

func requireBodyMatchProductSize(t *testing.T, body io.ReadCloser, productSize db.ProductSize) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotProductSize db.ProductSize
	err = json.Unmarshal(data, &gotProductSize)
	require.NoError(t, err)
	require.Equal(t, productSize, gotProductSize)
}
