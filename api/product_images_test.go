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
			tc.buildStubs(store)

			server := newTestServer(t, store, worker, ik)
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
