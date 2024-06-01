package db

import (
	"context"
	"testing"

	"github.com/cshop/v3/util"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func createRandomProductBrand(t *testing.T) ProductBrand {
	t.Helper()
	// arg := util.RandomString(5)
	var productBrand ProductBrand
	var err error
	productBrands := []string{"mango", "ted_baker", "zara", "shein", "LC"}
	brandsLogo := []string{
		"https://d1yjjnpx0p53s8.cloudfront.net/styles/logo-original-577x577/s3/102017/untitled-6_8.png?4kLt0HUeKlnH3zJVWgPKOf1FHrnY_5gH&itok=osnLnTRt",
		"https://d1yjjnpx0p53s8.cloudfront.net/styles/logo-original-577x577/s3/0004/1523/brand.gif?itok=gnchlsYq",
		"https://d1yjjnpx0p53s8.cloudfront.net/styles/logo-original-577x577/s3/102018/untitled-1_126.png?pxLdNigolZ6KO6cT8iVRAo_Zvg_quh0j&itok=8Zzi6Jbp",
		"https://upload.wikimedia.org/wikipedia/commons/2/25/Shein-logo.png",
		"https://d1yjjnpx0p53s8.cloudfront.net/styles/logo-original-577x577/s3/0023/5466/brand.gif?itok=oP8FEAL5",
		"https://d1yjjnpx0p53s8.cloudfront.net/styles/logo-original-577x577/s3/0008/5781/brand.gif?itok=pX0j4lkS",
	}
	for i := 0; i < len(productBrands); i++ {
		randomInt := util.RandomInt(0, int64(len(productBrands)-1))
		arg := CreateProductBrandParams{
			BrandName:  productBrands[randomInt],
			BrandImage: brandsLogo[i],
		}
		productBrand, err = testStore.CreateProductBrand(context.Background(), arg)
		require.NoError(t, err)
		require.NotEmpty(t, productBrand)

		require.Equal(t, productBrands[randomInt], productBrand.BrandName)

	}
	return productBrand
}

func createRandomProductBrandForUpdateOrDelete(t *testing.T) ProductBrand {
	t.Helper()
	brandName := util.RandomString(5)
	// arg := CreateProductBrandParams{
	// 	BrandName:     brandName,
	// }
	// productBrandChan := make(chan ProductBrand)
	// errChan := make(chan error)

	// go func() {
	// 	productBrand, err := testStore.CreateProductBrand(context.Background(), arg)
	// 	productBrandChan <- productBrand
	// 	errChan <- err
	// }()

	// err := <-errChan
	// productBrand := <-productBrandChan
	arg := CreateProductBrandParams{
		BrandName:  brandName,
		BrandImage: util.RandomURL(),
	}
	productBrand, err := testStore.CreateProductBrand(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, productBrand)

	require.Equal(t, brandName, productBrand.BrandName)
	require.NotEmpty(t, productBrand.ID)

	return productBrand
}

func adminCreateRandomProductBrandForUpdateOrDelete(t *testing.T) ProductBrand {
	t.Helper()
	admin := createRandomAdmin(t)
	brandName := util.RandomString(5)
	// arg := CreateProductBrandParams{
	// 	BrandName:     brandName,
	// }
	// productBrandChan := make(chan ProductBrand)
	// errChan := make(chan error)

	// go func() {
	// 	productBrand, err := testStore.CreateProductBrand(context.Background(), arg)
	// 	productBrandChan <- productBrand
	// 	errChan <- err
	// }()

	// err := <-errChan
	// productBrand := <-productBrandChan
	arg := AdminCreateProductBrandParams{
		AdminID:    admin.ID,
		BrandName:  brandName,
		BrandImage: util.RandomURL(),
	}
	productBrand, err := testStore.AdminCreateProductBrand(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, productBrand)

	require.Equal(t, brandName, productBrand.BrandName)
	require.NotEmpty(t, productBrand.ID)

	return productBrand
}

func TestCreateProductBrand(t *testing.T) {
	go createRandomProductBrand(t)
}
func TestCreateProductBrandForUpdateOrDelete(t *testing.T) {
	go createRandomProductBrandForUpdateOrDelete(t)
}

func TestAdminCreateProductBrand(t *testing.T) {
	go adminCreateRandomProductBrandForUpdateOrDelete(t)
}

func TestGetProductBrand(t *testing.T) {
	productBrand1 := createRandomProductBrand(t)
	productBrand2, err := testStore.GetProductBrand(context.Background(), productBrand1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, productBrand2)

	require.Equal(t, productBrand1.ID, productBrand2.ID)
	require.Equal(t, productBrand1.BrandName, productBrand2.BrandName)
}

func TestUpdateProductBrand(t *testing.T) {
	productBrand1 := createRandomProductBrandForUpdateOrDelete(t)
	arg := UpdateProductBrandParams{
		ID:        productBrand1.ID,
		BrandName: util.RandomString(5),
	}

	productBrand2, err := testStore.UpdateProductBrand(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, productBrand2)

	require.Equal(t, productBrand1.ID, productBrand2.ID)
	require.NotEqual(t, productBrand1.BrandName, productBrand2.BrandName)

	err = testStore.DeleteProductBrand(context.Background(), productBrand1.ID)

	require.NoError(t, err)
}

func TestDeleteProductBrand(t *testing.T) {
	productBrand1 := createRandomProductBrandForUpdateOrDelete(t)

	err := testStore.DeleteProductBrand(context.Background(), productBrand1.ID)

	require.NoError(t, err)

	productBrand2, err := testStore.GetProductBrand(context.Background(), productBrand1.ID)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, productBrand2)

}

func TestListProductBrands(t *testing.T) {
	for i := 0; i < 5; i++ {
		createRandomProductBrand(t)
	}
	// arg := ListProductBrandsParams{
	// 	Limit:  5,
	// 	Offset: 0,
	// }

	userBrands, err := testStore.ListProductBrands(context.Background())
	require.NoError(t, err)
	// require.Len(t, userBrands, 5)

	for _, userBrand := range userBrands {
		require.NotEmpty(t, userBrand)

	}
}
