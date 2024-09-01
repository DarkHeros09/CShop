package db

import (
	"context"
	"testing"
	"time"

	"github.com/cshop/v3/util"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func createRandomProduct(t *testing.T) Product {
	category := createRandomProductCategory(t)
	brand := createRandomProductBrand(t)
	arg := CreateProductParams{
		CategoryID:  category.ID,
		BrandID:     brand.ID,
		Name:        util.RandomUser(),
		Description: util.RandomUser(),
		Active:      true,
	}

	product, err := testStore.CreateProduct(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, product)

	require.Equal(t, arg.CategoryID, product.CategoryID)
	require.Equal(t, arg.Name, product.Name)
	require.Equal(t, arg.Description, product.Description)
	// require.Equal(t, arg.ProductImage, product.ProductImage)
	require.Equal(t, arg.Active, product.Active)
	require.NotEmpty(t, product.CreatedAt)
	require.True(t, product.UpdatedAt.IsZero())
	require.True(t, product.Active)

	return product
}

func adminCreateRandomProduct(t *testing.T) Product {
	admin := createRandomAdmin(t)
	category := createRandomProductCategory(t)
	brand := createRandomProductBrand(t)
	arg := AdminCreateProductParams{
		AdminID:     admin.ID,
		CategoryID:  category.ID,
		BrandID:     brand.ID,
		Name:        util.RandomUser(),
		Description: util.RandomUser(),
		Active:      true,
	}

	product, err := testStore.AdminCreateProduct(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, product)

	require.Equal(t, arg.CategoryID, product.CategoryID)
	require.Equal(t, arg.Name, product.Name)
	require.Equal(t, arg.Description, product.Description)
	// require.Equal(t, arg.ProductImage, product.ProductImage)
	require.Equal(t, arg.Active, product.Active)
	require.NotEmpty(t, product.CreatedAt)
	require.True(t, product.UpdatedAt.IsZero())
	require.True(t, product.Active)

	return product
}
func TestCreateProduct(t *testing.T) {
	createRandomProduct(t)
}

func TestAdminCreateProduct(t *testing.T) {
	adminCreateRandomProduct(t)
}

func TestGetProduct(t *testing.T) {
	product1 := createRandomProduct(t)
	product2, err := testStore.GetProduct(context.Background(), product1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, product2)

	require.Equal(t, product1.ID, product2.ID)
	require.Equal(t, product1.CategoryID, product2.CategoryID)
	require.Equal(t, product1.Name, product2.Name)
	require.Equal(t, product1.Description, product2.Description)
	// require.Equal(t, product1.ProductImage, product2.ProductImage)
	require.Equal(t, product1.Active, product2.Active)
	require.Equal(t, product1.CreatedAt, product2.CreatedAt)
	require.Equal(t, product1.UpdatedAt, product2.UpdatedAt)
	require.True(t, product2.Active)

}

func TestUpdateProductName(t *testing.T) {
	product1 := createRandomProduct(t)
	arg := UpdateProductParams{
		ID:          product1.ID,
		CategoryID:  null.Int{},
		Name:        null.StringFrom(util.RandomString(5)),
		Description: null.String{},
		// ProductImage: null.String{},
		Active: null.Bool{},
	}

	product2, err := testStore.UpdateProduct(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, product2)

	require.Equal(t, product1.ID, product2.ID)
	require.Equal(t, product1.CategoryID, product2.CategoryID)
	require.NotEqual(t, product1.Name, product2.Name)
	require.Equal(t, product1.Description, product2.Description)
	require.Equal(t, product1.Active, product2.Active)
	require.True(t, product2.Active)
	require.WithinDuration(t, product1.CreatedAt, product2.CreatedAt, time.Second)
	require.NotEqual(t, product1.UpdatedAt, product2.UpdatedAt)

}

func TestUpdateProductCategoryAndActive(t *testing.T) {
	product1 := createRandomProduct(t)
	// category := createRandomProductCategory(t)
	arg := UpdateProductParams{
		ID:         product1.ID,
		CategoryID: null.IntFromPtr(&product1.CategoryID),
		Active:     null.BoolFrom(!product1.Active),
	}

	product2, err := testStore.UpdateProduct(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, product2)

	require.Equal(t, product1.ID, product2.ID)
	require.Equal(t, product1.CategoryID, product2.CategoryID)
	require.Equal(t, product1.Name, product2.Name)
	require.Equal(t, product1.Description, product2.Description)
	require.NotEqual(t, product1.Active, product2.Active)
	require.False(t, product2.Active)
	require.WithinDuration(t, product1.CreatedAt, product2.CreatedAt, time.Second)
	require.NotEqual(t, product1.UpdatedAt, product2.UpdatedAt)

}

func TestAdminUpdateProductCategoryAndActive(t *testing.T) {
	admin := createRandomAdmin(t)
	product1 := createRandomProduct(t)
	// category := createRandomProductCategory(t)
	arg := AdminUpdateProductParams{
		AdminID:    admin.ID,
		ID:         product1.ID,
		CategoryID: null.IntFromPtr(&product1.CategoryID),
		Active:     null.BoolFrom(!product1.Active),
	}

	product2, err := testStore.AdminUpdateProduct(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, product2)

	require.Equal(t, product1.ID, product2.ID)
	require.Equal(t, product1.CategoryID, product2.CategoryID)
	require.Equal(t, product1.Name, product2.Name)
	require.Equal(t, product1.Description, product2.Description)
	require.NotEqual(t, product1.Active, product2.Active)
	require.False(t, product2.Active)
	require.WithinDuration(t, product1.CreatedAt, product2.CreatedAt, time.Second)
	require.NotEqual(t, product1.UpdatedAt, product2.UpdatedAt)

}

func TestDeleteProduct(t *testing.T) {
	product1 := createRandomProduct(t)
	err := testStore.DeleteProduct(context.Background(), product1.ID)

	require.NoError(t, err)

	product2, err := testStore.GetProduct(context.Background(), product1.ID)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, product2)

}

func TestListProducts(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomProduct(t)
	}
	arg := ListProductsParams{
		Limit:  5,
		Offset: 0,
	}

	products, err := testStore.ListProducts(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, products)

	for _, product := range products {
		require.NotEmpty(t, product)
	}

}

func TestGetProductsByIDs(t *testing.T) {
	var listOfIds []int64
	for i := 0; i < 10; i++ {
		product := createRandomProduct(t)
		listOfIds = append(listOfIds, product.ID)
	}

	products, err := testStore.GetProductsByIDs(context.Background(), listOfIds)

	require.NoError(t, err)
	require.NotEmpty(t, products)

	for _, product := range products {
		require.NotEmpty(t, product)
		require.Contains(t, listOfIds, product.ID)
	}

}

func TestListProductsV2(t *testing.T) {
	for i := 0; i < 30; i++ {
		createRandomProduct(t)
	}

	limit := 10

	initialSearchResult, err := testStore.ListProductsV2(context.Background(), int32(limit))
	// fmt.Println(initialSearchResult)
	require.NoError(t, err)
	require.NotEmpty(t, initialSearchResult)
	require.Equal(t, len(initialSearchResult), 10)

	arg1 := ListProductsNextPageParams{
		Limit: 10,
		ID:    initialSearchResult[len(initialSearchResult)-1].ID,
	}

	secondPage, err := testStore.ListProductsNextPage(context.Background(), arg1)
	// fmt.Println(secondPage)
	require.NoError(t, err)
	require.NotEmpty(t, secondPage)
	require.Equal(t, len(secondPage), 10)
	require.Greater(t, initialSearchResult[len(initialSearchResult)-1].ID, secondPage[len(secondPage)-1].ID)

	arg2 := ListProductsNextPageParams{
		Limit: 10,
		ID:    secondPage[len(secondPage)-1].ID,
	}

	thirdPage, err := testStore.ListProductsNextPage(context.Background(), arg2)
	require.NoError(t, err)
	require.NotEmpty(t, secondPage)
	require.Equal(t, len(initialSearchResult), 10)
	require.Greater(t, secondPage[len(secondPage)-1].ID, thirdPage[len(thirdPage)-1].ID)
	require.Greater(t, initialSearchResult[len(initialSearchResult)-1].ID, thirdPage[len(thirdPage)-1].ID)
}

func TestSearchProducts(t *testing.T) {

	product := createRandomProduct(t)

	product, err := testStore.GetProduct(context.Background(), product.ID)

	require.NoError(t, err)
	require.NotEmpty(t, product)

	arg1 := SearchProductsParams{
		Limit: 10,
		Query: product.Name,
	}

	searchedProduct, err := testStore.SearchProducts(context.Background(), arg1)

	require.NoError(t, err)
	require.NotEmpty(t, searchedProduct)
	require.Equal(t, product.ID, searchedProduct[len(searchedProduct)-1].ID)

	arg2 := SearchProductsNextPageParams{
		Limit:     10,
		ProductID: searchedProduct[len(searchedProduct)-1].ID,
		Query:     product.Name,
	}

	searchedRestProduct, err := testStore.SearchProductsNextPage(context.Background(), arg2)

	require.NoError(t, err)
	require.Empty(t, searchedRestProduct)

}
