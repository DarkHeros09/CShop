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
	mockemail "github.com/cshop/v3/mail/mock"
	"github.com/cshop/v3/token"
	"github.com/cshop/v3/util"
	mockwk "github.com/cshop/v3/worker/mock"
	"github.com/gofiber/fiber/v2"
	"github.com/guregu/null/v5"
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
				"size_value":      productSize.SizeValue,
				"product_item_id": productSize.ProductItemID,
				"qty":             productSize.Qty,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminCreateProductSizeParams{
					AdminID:       admin.ID,
					SizeValue:     productSize.SizeValue,
					ProductItemID: productSize.ProductItemID,
					Qty:           productSize.Qty,
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
					AdminID:       admin.ID,
					SizeValue:     productSize.SizeValue,
					ProductItemID: productSize.ProductItemID,
					Qty:           productSize.Qty,
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
				"size_value":      productSize.SizeValue,
				"product_item_id": productSize.ProductItemID,
				"qty":             productSize.Qty,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminCreateProductSizeParams{
					AdminID:       admin.ID,
					SizeValue:     productSize.SizeValue,
					ProductItemID: productSize.ProductItemID,
					Qty:           productSize.Qty,
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
				"size_value":      productSize.SizeValue,
				"product_item_id": productSize.ProductItemID,
				"qty":             productSize.Qty,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					AdminCreateProductSize(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, pgx.ErrTxClosed)
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
			mailSender := mockemail.NewMockEmailSender(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store, worker, ik, mailSender)
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
	productItem := randomProductItem()
	n := 5
	productSizes := make([]*db.ProductSize, n)
	for i := 0; i < n; i++ {
		productSizes[i] = randomProductSizesForList(productItem.ID)
	}

	testCases := []struct {
		name          string
		productItemID int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(rsp *http.Response)
	}{
		{
			name:          "OK",
			productItemID: productItem.ID,
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					ListProductSizesByProductItemID(gomock.Any(), productItem.ID).
					Times(1).
					Return(productSizes, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchProductSizesList(t, rsp.Body, productSizes)
			},
		},
		{
			name:          "InternalError",
			productItemID: productItem.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListProductSizesByProductItemID(gomock.Any(), productItem.ID).
					Times(1).
					Return(nil, pgx.ErrTxClosed)
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
			mailSender := mockemail.NewMockEmailSender(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store, worker, ik, mailSender)
			//recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/api/v1/sizes/%d", tc.productItemID)
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
		})
	}
}

func TestUpdateProductSizeAPI(t *testing.T) {
	admin, _ := randomProductSizeSuperAdmin(t)
	productSize := randomProductSize()

	testCases := []struct {
		name          string
		body          fiber.Map
		productSizeID int64
		AdminID       int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, rsp *http.Response)
	}{
		{
			name:          "OK",
			productSizeID: productSize.ID,
			AdminID:       admin.ID,
			body: fiber.Map{
				"size":            productSize.SizeValue,
				"product_item_id": productSize.ProductItemID,
				"qty":             productSize.Qty,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminUpdateProductSizeParams{
					ID:            productSize.ID,
					AdminID:       admin.ID,
					SizeValue:     null.StringFrom(productSize.SizeValue),
					ProductItemID: productSize.ProductItemID,
					Qty:           null.IntFrom(int64(productSize.Qty)),
				}

				store.EXPECT().
					AdminUpdateProductSize(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(productSize, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:          "Unauthorized",
			productSizeID: productSize.ID,
			AdminID:       admin.ID,
			body: fiber.Map{
				"size":            productSize.SizeValue,
				"product_item_id": productSize.ProductItemID,
				"qty":             productSize.Qty,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, false, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminUpdateProductSizeParams{
					ID:            productSize.ID,
					AdminID:       admin.ID,
					SizeValue:     null.StringFrom(productSize.SizeValue),
					ProductItemID: productSize.ProductItemID,
					Qty:           null.IntFrom(int64(productSize.Qty)),
				}

				store.EXPECT().
					AdminUpdateProductSize(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:          "NoAuthorization",
			productSizeID: productSize.ID,
			AdminID:       admin.ID,
			body: fiber.Map{
				"size":            productSize.SizeValue,
				"product_item_id": productSize.ProductItemID,
				"qty":             productSize.Qty,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminUpdateProductSizeParams{
					ID:            productSize.ID,
					AdminID:       admin.ID,
					SizeValue:     null.StringFrom(productSize.SizeValue),
					ProductItemID: productSize.ProductItemID,
					Qty:           null.IntFrom(int64(productSize.Qty)),
				}

				store.EXPECT().
					AdminUpdateProductSize(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:          "InternalError",
			productSizeID: productSize.ID,
			AdminID:       admin.ID,
			body: fiber.Map{
				"size":            productSize.SizeValue,
				"product_item_id": productSize.ProductItemID,
				"qty":             productSize.Qty,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminUpdateProductSizeParams{
					ID:            productSize.ID,
					AdminID:       admin.ID,
					SizeValue:     null.StringFrom(productSize.SizeValue),
					ProductItemID: productSize.ProductItemID,
					Qty:           null.IntFrom(int64(productSize.Qty)),
				}
				store.EXPECT().
					AdminUpdateProductSize(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(nil, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:          "InvalidID",
			productSizeID: 0,
			AdminID:       admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					AdminUpdateProductSize(gomock.Any(), gomock.Any()).
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
			ik := mockik.NewMockImageKitManagement(ctrl)
			mailSender := mockemail.NewMockEmailSender(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store, worker, ik, mailSender)
			//recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/admin/v1/admins/%d/sizes/%d", tc.AdminID, tc.productSizeID)
			request, err := http.NewRequest(fiber.MethodPut, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.adminTokenMaker)
			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(t, rsp)
		})
	}
}

func randomProductSizeSuperAdmin(t *testing.T) (admin db.Admin, password string) {
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

func randomProductSizesForList(productItemId int64) *db.ProductSize {
	return &db.ProductSize{
		ID:            util.RandomInt(1, 1000),
		SizeValue:     util.RandomSize(),
		ProductItemID: productItemId,
		Qty:           int32(util.RandomInt(1, 1000)),
	}
}

func requireBodyMatchProductSizesList(t *testing.T, body io.ReadCloser, productSizes []*db.ProductSize) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotProducts []*db.ProductSize
	err = json.Unmarshal(data, &gotProducts)
	require.NoError(t, err)
	require.Equal(t, productSizes, gotProducts)
}

func randomProductSize() *db.ProductSize {
	return &db.ProductSize{
		ID:            util.RandomInt(0, 1000),
		SizeValue:     util.RandomSize(),
		ProductItemID: util.RandomInt(0, 1000),
		Qty:           int32(util.RandomInt(0, 1000)),
	}
}

func randomProductSizeWithProductItemID(productItemId int64) *db.ProductSize {
	return &db.ProductSize{
		ID:            util.RandomInt(0, 1000),
		SizeValue:     util.RandomSize(),
		ProductItemID: productItemId,
		Qty:           int32(util.RandomInt(0, 1000)),
	}
}

func requireBodyMatchProductSize(t *testing.T, body io.ReadCloser, productSize *db.ProductSize) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotProductSize *db.ProductSize
	err = json.Unmarshal(data, &gotProductSize)
	require.NoError(t, err)
	require.Equal(t, productSize, gotProductSize)
}
