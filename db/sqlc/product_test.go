package db

import (
	"context"
	"testing"
	"time"

	"github.com/cshop/v3/util"
	"github.com/guregu/null"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/require"
)

func createRandomProduct(t *testing.T) Product {
	category := createRandomProductCategoryParent(t)
	arg := CreateProductParams{
		CategoryID:   category.ID,
		Name:         util.RandomUser(),
		Description:  util.RandomUser(),
		ProductImage: util.RandomURL(),
		Active:       true,
	}

	product, err := testQueires.CreateProduct(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, product)

	require.Equal(t, arg.CategoryID, product.CategoryID)
	require.Equal(t, arg.Name, product.Name)
	require.Equal(t, arg.Description, product.Description)
	require.Equal(t, arg.ProductImage, product.ProductImage)
	require.Equal(t, arg.Active, product.Active)
	require.NotEmpty(t, product.CreatedAt)
	require.True(t, product.UpdatedAt.IsZero())
	require.True(t, product.Active)

	return product
}
func TestCreateProduct(t *testing.T) {
	createRandomProduct(t)
}

func TestGetProduct(t *testing.T) {
	product1 := createRandomProduct(t)
	product2, err := testQueires.GetProduct(context.Background(), product1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, product2)

	require.Equal(t, product1.ID, product2.ID)
	require.Equal(t, product1.CategoryID, product2.CategoryID)
	require.Equal(t, product1.Name, product2.Name)
	require.Equal(t, product1.Description, product2.Description)
	require.Equal(t, product1.ProductImage, product2.ProductImage)
	require.Equal(t, product1.Active, product2.Active)
	require.Equal(t, product1.CreatedAt, product2.CreatedAt)
	require.Equal(t, product1.UpdatedAt, product2.UpdatedAt)
	require.True(t, product2.Active)

}

func TestUpdateProductName(t *testing.T) {
	product1 := createRandomProduct(t)
	arg := UpdateProductParams{
		ID:           product1.ID,
		CategoryID:   null.Int{},
		Name:         null.StringFrom(util.RandomString(5)),
		Description:  null.String{},
		ProductImage: null.String{},
		Active:       null.Bool{},
	}

	product2, err := testQueires.UpdateProduct(context.Background(), arg)

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
	category := createRandomProductCategoryParent(t)
	arg := UpdateProductParams{
		ID:           product1.ID,
		CategoryID:   null.IntFromPtr(&category.ParentCategoryID.Int64),
		Name:         null.String{},
		Description:  null.String{},
		ProductImage: null.String{},
		Active:       null.BoolFrom(!product1.Active),
	}

	product2, err := testQueires.UpdateProduct(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, product2)

	require.Equal(t, product1.ID, product2.ID)
	require.NotEqual(t, product1.CategoryID, product2.CategoryID)
	require.Equal(t, product1.Name, product2.Name)
	require.Equal(t, product1.Description, product2.Description)
	require.NotEqual(t, product1.Active, product2.Active)
	require.False(t, product2.Active)
	require.WithinDuration(t, product1.CreatedAt, product2.CreatedAt, time.Second)
	require.NotEqual(t, product1.UpdatedAt, product2.UpdatedAt)

}

func TestDeleteProduct(t *testing.T) {
	product1 := createRandomProduct(t)
	err := testQueires.DeleteProduct(context.Background(), product1.ID)

	require.NoError(t, err)

	product2, err := testQueires.GetProduct(context.Background(), product1.ID)

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
		Offset: 5,
	}

	products, err := testQueires.ListProducts(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, products)

	for _, product := range products {
		require.NotEmpty(t, product)
	}

}
