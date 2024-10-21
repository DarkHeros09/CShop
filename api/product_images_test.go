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

func TestCreateProductImagesAPI(t *testing.T) {
	admin, _ := randomProductImageSuperAdmin(t)
	productImage := randomProductImage()

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
				"product_image_1": productImage.ProductImage1,
				"product_image_2": productImage.ProductImage2,
				"product_image_3": productImage.ProductImage3,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminCreateProductImagesParams{
					AdminID:       admin.ID,
					ProductImage1: productImage.ProductImage1,
					ProductImage2: productImage.ProductImage2,
					ProductImage3: productImage.ProductImage3,
				}

				store.EXPECT().
					AdminCreateProductImages(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(productImage, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchProductImage(t, rsp.Body, productImage)
			},
		},
		{
			name:    "NoAuthorization",
			AdminID: admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminCreateProductImagesParams{
					AdminID:       admin.ID,
					ProductImage1: productImage.ProductImage1,
					ProductImage2: productImage.ProductImage2,
					ProductImage3: productImage.ProductImage3,
				}

				store.EXPECT().
					AdminCreateProductImages(gomock.Any(), gomock.Eq(arg)).
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
				"product_image_1": productImage.ProductImage1,
				"product_image_2": productImage.ProductImage2,
				"product_image_3": productImage.ProductImage3,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, false, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminCreateProductImagesParams{
					AdminID:       admin.ID,
					ProductImage1: productImage.ProductImage1,
					ProductImage2: productImage.ProductImage2,
					ProductImage3: productImage.ProductImage3,
				}

				store.EXPECT().
					AdminCreateProductImages(gomock.Any(), gomock.Eq(arg)).
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
				"product_image_1": productImage.ProductImage1,
				"product_image_2": productImage.ProductImage2,
				"product_image_3": productImage.ProductImage3,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					AdminCreateProductImages(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.ProductImage{}, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:    "InvalidID",
			AdminID: admin.ID,
			body: fiber.Map{
				"product_image_1": "",
				"product_image_2": productImage.ProductImage2,
				"product_image_3": productImage.ProductImage3,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					AdminCreateProductImages(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/admin/v1/admins/%d/product-images", tc.AdminID)
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

func TestListProductImagesV2API(t *testing.T) {
	n := 10
	images := make([]db.ListProductImagesV2Row, n)
	for i := 0; i < n; i++ {
		images[i] = randomListProductImagesV2()
	}

	type Query struct {
		Limit int
	}

	testCases := []struct {
		name          string
		query         Query
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(rsp *http.Response)
	}{
		{
			name: "OK",
			query: Query{
				Limit: n,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListProductImagesV2(gomock.Any(), gomock.Eq(int32(10))).
					Times(1).
					Return(images, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchListProductImagesV2(t, rsp.Body, images)
			},
		},
		{
			name: "InternalError",
			query: Query{
				Limit: 10,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListProductImagesV2(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.ListProductImagesV2Row{}, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name: "InvalidLimit",
			query: Query{
				Limit: 11,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListProductImagesV2(gomock.Any(), gomock.Any()).
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

			url := "/api/v1/images-v2"
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			// Add query parameters to request URL
			q := request.URL.Query()
			q.Add("limit", fmt.Sprintf("%d", tc.query.Limit))
			request.URL.RawQuery = q.Encode()

			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
		})
	}
}

func TestListProductImagesNextPageAPI(t *testing.T) {
	n := 10
	images1 := make([]db.ListProductImagesNextPageRow, n)
	images2 := make([]db.ListProductImagesNextPageRow, n)
	for i := 0; i < n; i++ {
		images1[i] = randomRestProductImagesList()
	}
	for i := 0; i < n; i++ {
		images2[i] = randomRestProductImagesList()
	}

	type Query struct {
		ProductCursor int
		Limit         int
	}

	testCases := []struct {
		name          string
		query         Query
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(rsp *http.Response)
	}{
		{
			name: "OK",
			query: Query{
				Limit:         n,
				ProductCursor: int(images1[len(images1)-1].ID),
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.ListProductImagesNextPageParams{
					Limit: int32(n),
					ID:    images1[len(images1)-1].ID,
				}

				store.EXPECT().
					ListProductImagesNextPage(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(images2, nil)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
				requireBodyMatchListProductImagesNextPage(t, rsp.Body, images2)
			},
		},
		{
			name: "InternalError",
			query: Query{
				Limit:         10,
				ProductCursor: int(images1[len(images1)-1].ID),
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					ListProductImagesNextPage(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.ListProductImagesNextPageRow{}, pgx.ErrTxClosed)
			},
			checkResponse: func(rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name: "InvalidLimit",
			query: Query{
				Limit:         11,
				ProductCursor: int(images1[len(images1)-1].ID),
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListProductImagesNextPage(gomock.Any(), gomock.Any()).
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

			url := "/api/v1/images-next-page"
			request, err := http.NewRequest(fiber.MethodGet, url, nil)
			require.NoError(t, err)

			// Add query parameters to request URL
			q := request.URL.Query()
			q.Add("limit", fmt.Sprintf("%d", tc.query.Limit))
			q.Add("product_cursor", fmt.Sprintf("%d", tc.query.ProductCursor))
			request.URL.RawQuery = q.Encode()

			request.Header.Set("Content-Type", "application/json")

			rsp, err := server.router.Test(request)
			require.NoError(t, err)
			tc.checkResponse(rsp)
		})
	}
}

func TestUpdateProductImageAPI(t *testing.T) {
	admin, _ := randomProductImageSuperAdmin(t)
	productImage := randomProductImage()

	testCases := []struct {
		name           string
		body           fiber.Map
		productImageID int64
		AdminID        int64
		setupAuth      func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs     func(store *mockdb.MockStore)
		checkResponse  func(t *testing.T, rsp *http.Response)
	}{
		{
			name:           "OK",
			productImageID: productImage.ID,
			AdminID:        admin.ID,
			body: fiber.Map{
				"product_image_1": productImage.ProductImage1,
				"product_image_2": productImage.ProductImage2,
				"product_image_3": productImage.ProductImage3,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminUpdateProductImageParams{
					ID:            productImage.ID,
					AdminID:       admin.ID,
					ProductImage1: null.StringFrom(productImage.ProductImage1),
					ProductImage2: null.StringFrom(productImage.ProductImage2),
					ProductImage3: null.StringFrom(productImage.ProductImage3),
				}

				store.EXPECT().
					AdminUpdateProductImage(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(productImage, nil)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusOK, rsp.StatusCode)
			},
		},
		{
			name:           "Unauthorized",
			productImageID: productImage.ID,
			AdminID:        admin.ID,
			body: fiber.Map{
				"product_image_1": productImage.ProductImage1,
				"product_image_2": productImage.ProductImage2,
				"product_image_3": productImage.ProductImage3,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, false, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminUpdateProductImageParams{
					ID:            productImage.ID,
					AdminID:       admin.ID,
					ProductImage1: null.StringFrom(productImage.ProductImage1),
					ProductImage2: null.StringFrom(productImage.ProductImage2),
					ProductImage3: null.StringFrom(productImage.ProductImage3),
				}

				store.EXPECT().
					AdminUpdateProductImage(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:           "NoAuthorization",
			productImageID: productImage.ID,
			AdminID:        admin.ID,
			body: fiber.Map{
				"product_image_1": productImage.ProductImage1,
				"product_image_2": productImage.ProductImage2,
				"product_image_3": productImage.ProductImage3,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminUpdateProductImageParams{
					ID:            productImage.ID,
					AdminID:       admin.ID,
					ProductImage1: null.StringFrom(productImage.ProductImage1),
					ProductImage2: null.StringFrom(productImage.ProductImage2),
					ProductImage3: null.StringFrom(productImage.ProductImage3),
				}

				store.EXPECT().
					AdminUpdateProductImage(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusUnauthorized, rsp.StatusCode)
			},
		},
		{
			name:           "InternalError",
			productImageID: productImage.ID,
			AdminID:        admin.ID,
			body: fiber.Map{
				"product_image_1": productImage.ProductImage1,
				"product_image_2": productImage.ProductImage2,
				"product_image_3": productImage.ProductImage3,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.AdminUpdateProductImageParams{
					ID:            productImage.ID,
					AdminID:       admin.ID,
					ProductImage1: null.StringFrom(productImage.ProductImage1),
					ProductImage2: null.StringFrom(productImage.ProductImage2),
					ProductImage3: null.StringFrom(productImage.ProductImage3),
				}
				store.EXPECT().
					AdminUpdateProductImage(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.ProductImage{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, rsp *http.Response) {
				require.Equal(t, http.StatusInternalServerError, rsp.StatusCode)
			},
		},
		{
			name:           "InvalidID",
			productImageID: 0,
			AdminID:        admin.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationForAdmin(t, request, tokenMaker, authorizationTypeBearer, admin.ID, admin.Username, admin.TypeID, admin.Active, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					AdminUpdateProductImage(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/admin/v1/admins/%d/product-images/%d", tc.AdminID, tc.productImageID)
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

func randomProductImageSuperAdmin(t *testing.T) (admin db.Admin, password string) {
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

func randomProductImage() db.ProductImage {
	return db.ProductImage{
		ID:            util.RandomInt(1, 10000),
		ProductImage1: util.RandomURL(),
		ProductImage2: util.RandomURL(),
		ProductImage3: util.RandomURL(),
	}
}

func requireBodyMatchProductImage(t *testing.T, body io.ReadCloser, productImage db.ProductImage) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotProductImage db.ProductImage
	err = json.Unmarshal(data, &gotProductImage)
	require.NoError(t, err)
	require.Equal(t, productImage, gotProductImage)
}

func requireBodyMatchListProductImagesV2(t *testing.T, body io.ReadCloser, productImages []db.ListProductImagesV2Row) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotProductImages []db.ListProductImagesV2Row
	err = json.Unmarshal(data, &gotProductImages)
	require.NoError(t, err)
	require.Equal(t, productImages, gotProductImages)
}

func requireBodyMatchListProductImagesNextPage(t *testing.T, body io.ReadCloser, productImages []db.ListProductImagesNextPageRow) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotProductImages []db.ListProductImagesNextPageRow
	err = json.Unmarshal(data, &gotProductImages)
	require.NoError(t, err)
	require.Equal(t, productImages, gotProductImages)
}

func randomListProductImagesV2() db.ListProductImagesV2Row {
	return db.ListProductImagesV2Row{
		ID:            util.RandomInt(1, 1000),
		ProductImage1: util.RandomURL(),
		ProductImage2: util.RandomURL(),
		ProductImage3: util.RandomURL(),
	}
}

func randomRestProductImagesList() db.ListProductImagesNextPageRow {
	return db.ListProductImagesNextPageRow{
		ID:            util.RandomInt(1, 1000),
		ProductImage1: util.RandomURL(),
		ProductImage2: util.RandomURL(),
		ProductImage3: util.RandomURL(),
	}
}
