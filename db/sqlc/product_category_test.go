package db

import (
	"context"
	"testing"

	"github.com/cshop/v3/util"
	"github.com/guregu/null"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/require"
)

func createRandomProductCategory(t *testing.T) ProductCategory {
	arg := CreateProductCategoryParams{
		ParentCategoryID: null.Int{},
		CategoryName:     util.RandomString(5),
	}

	productCategory, err := testQueires.CreateProductCategory(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, productCategory)

	require.Equal(t, arg.CategoryName, productCategory.CategoryName)
	require.NotEmpty(t, productCategory.ID)

	return productCategory
}

func createRandomProductCategoryParent(t *testing.T) ProductCategory {
	randomCategoryChan := make(chan ProductCategory)
	go func() {
		randomCategory := createRandomProductCategory(t)
		randomCategoryChan <- randomCategory
	}()
	randomCategory := <-randomCategoryChan
	arg := CreateProductCategoryParams{
		ParentCategoryID: null.IntFromPtr(&randomCategory.ID),
		CategoryName:     util.RandomString(5),
	}

	productCategory, err := testQueires.CreateProductCategory(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, productCategory)

	require.Equal(t, arg.CategoryName, productCategory.CategoryName)
	require.NotEmpty(t, productCategory.ID)

	return productCategory
}

func TestCreateProductCategory(t *testing.T) {
	go createRandomProductCategory(t)
}

func TestCreateProductCategoryParent(t *testing.T) {
	createRandomProductCategoryParent(t)
}

func TestGetProductCategory(t *testing.T) {
	productCategory1 := createRandomProductCategoryParent(t)
	productCategory2, err := testQueires.GetProductCategory(context.Background(), productCategory1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, productCategory2)

	require.Equal(t, productCategory1.ID, productCategory2.ID)
	require.Equal(t, productCategory1.ParentCategoryID, productCategory2.ParentCategoryID)
	require.Equal(t, productCategory1.CategoryName, productCategory2.CategoryName)
}

func TestGetProductCategoryByParent(t *testing.T) {
	productCategory1 := createRandomProductCategoryParent(t)

	arg := GetProductCategoryByParentParams{
		ID:               productCategory1.ID,
		ParentCategoryID: null.IntFromPtr(&productCategory1.ParentCategoryID.Int64),
	}
	productCategory2, err := testQueires.GetProductCategoryByParent(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, productCategory2)

	require.Equal(t, productCategory1.ID, productCategory2.ID)
	require.Equal(t, productCategory1.ParentCategoryID, productCategory2.ParentCategoryID)
	require.Equal(t, productCategory1.CategoryName, productCategory2.CategoryName)
}

func TestUpdateProductCategory(t *testing.T) {
	productCategory1 := createRandomProductCategory(t)
	arg := UpdateProductCategoryParams{
		ID:           productCategory1.ID,
		CategoryName: util.RandomString(5),
	}

	productCategory2, err := testQueires.UpdateProductCategory(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, productCategory2)

	require.Equal(t, productCategory1.ID, productCategory2.ID)
	require.Empty(t, productCategory1.ParentCategoryID)
	require.Empty(t, productCategory2.ParentCategoryID)
	require.NotEqual(t, productCategory1.CategoryName, productCategory2.CategoryName)
}

func TestUpdateProductCategoryParent(t *testing.T) {
	productCategory1 := createRandomProductCategoryParent(t)
	arg := UpdateProductCategoryParams{
		ID:               productCategory1.ID,
		ParentCategoryID: null.IntFromPtr(&productCategory1.ParentCategoryID.Int64),
		CategoryName:     util.RandomString(5),
	}

	productCategory2, err := testQueires.UpdateProductCategory(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, productCategory2)

	require.Equal(t, productCategory1.ID, productCategory2.ID)
	require.Equal(t, productCategory1.ParentCategoryID.Int64, productCategory2.ParentCategoryID.Int64)
	require.NotEqual(t, productCategory1.CategoryName, productCategory2.CategoryName)
}

func TestDeleteProductCategory(t *testing.T) {
	productCategory1 := createRandomProductCategory(t)

	arg := DeleteProductCategoryParams{
		ID: productCategory1.ID,
	}
	err := testQueires.DeleteProductCategory(context.Background(), arg)

	require.NoError(t, err)

	productCategory2, err := testQueires.GetProductCategory(context.Background(), productCategory1.ID)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, productCategory2)

}

func TestDeleteProductCategoryParent(t *testing.T) {
	productCategory1 := createRandomProductCategory(t)

	arg1 := DeleteProductCategoryParams{
		ID:               productCategory1.ID,
		ParentCategoryID: null.IntFromPtr(&productCategory1.ParentCategoryID.Int64),
	}
	err := testQueires.DeleteProductCategory(context.Background(), arg1)

	require.NoError(t, err)

	arg2 := GetProductCategoryByParentParams{
		ID:               productCategory1.ID,
		ParentCategoryID: null.IntFromPtr(&productCategory1.ParentCategoryID.Int64),
	}

	productCategory2, err := testQueires.GetProductCategoryByParent(context.Background(), arg2)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, productCategory2)

}

func TestListProductCategories(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomProductCategory(t)
	}
	arg := ListProductCategoriesParams{
		Limit:  5,
		Offset: 5,
	}

	userCategories, err := testQueires.ListProductCategories(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, userCategories, 5)

	for _, userCategory := range userCategories {
		require.NotEmpty(t, userCategory)

	}
}
