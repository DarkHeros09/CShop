package db

import (
	"context"
	"testing"

	"github.com/cshop/v3/util"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func createRandomProductCategory(t *testing.T) ProductCategory {
	// arg := util.RandomString(5)
	var productCategory ProductCategory
	var err error
	productCategories := []string{"أحذية", "حقائب", "إكسسوارات", "حجاب", "عبايات", "قمصان", "تنانير", "بناطيل", "الأطقم", "الفساتين"}
	for i := 0; i < len(productCategories); i++ {
		randomInt := util.RandomInt(0, int64(len(productCategories)-1))
		arg := CreateProductCategoryParams{
			ParentCategoryID: null.Int{},
			CategoryName:     productCategories[randomInt],
			CategoryImage:    util.RandomURL(),
		}
		productCategory, err = testStore.CreateProductCategory(context.Background(), arg)
		require.NoError(t, err)
		require.NotEmpty(t, productCategory)

		require.Equal(t, productCategories[randomInt], productCategory.CategoryName)

	}
	return productCategory
}

func createRandomProductCategoryForUpdateOrDelete(t *testing.T) ProductCategory {
	categoryName := util.RandomString(5)
	arg := CreateProductCategoryParams{
		ParentCategoryID: null.Int{},
		CategoryName:     categoryName,
		CategoryImage:    util.RandomURL(),
	}
	// productCategoryChan := make(chan ProductCategory)
	// errChan := make(chan error)

	// go func() {
	// 	productCategory, err := testStore.CreateProductCategory(context.Background(), arg)
	// 	productCategoryChan <- productCategory
	// 	errChan <- err
	// }()

	// err := <-errChan
	// productCategory := <-productCategoryChan
	productCategory, err := testStore.CreateProductCategory(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, productCategory)

	require.Equal(t, arg.CategoryName, productCategory.CategoryName)
	require.NotEmpty(t, productCategory.ID)

	return productCategory
}

func createRandomProductCategoryParent(t *testing.T) ProductCategory {
	// randomCategoryChan := make(chan ProductCategory)
	// go func() {
	randomCategory := createRandomProductCategory(t)
	// randomCategoryChan <- randomCategory
	// }()
	// randomCategory := <-randomCategoryChan
	arg := CreateProductCategoryParams{
		ParentCategoryID: null.IntFromPtr(&randomCategory.ID),
		CategoryName:     util.RandomString(5),
	}

	productCategory, err := testStore.CreateProductCategory(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, productCategory)

	require.Equal(t, arg.CategoryName, productCategory.CategoryName)
	require.NotEmpty(t, productCategory.ID)

	return productCategory
}

func adminCreateRandomProductCategoryForUpdateOrDelete(t *testing.T) ProductCategory {
	admin := createRandomAdmin(t)
	categoryName := util.RandomString(5)
	arg := AdminCreateProductCategoryParams{
		AdminID:          admin.ID,
		ParentCategoryID: null.Int{},
		CategoryName:     categoryName,
		CategoryImage:    util.RandomURL(),
	}
	// productCategoryChan := make(chan ProductCategory)
	// errChan := make(chan error)

	// go func() {
	// 	productCategory, err := testStore.CreateProductCategory(context.Background(), arg)
	// 	productCategoryChan <- productCategory
	// 	errChan <- err
	// }()

	// err := <-errChan
	// productCategory := <-productCategoryChan
	productCategory, err := testStore.AdminCreateProductCategory(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, productCategory)

	require.Equal(t, arg.CategoryName, productCategory.CategoryName)
	require.NotEmpty(t, productCategory.ID)

	return productCategory
}

func TestCreateProductCategory(t *testing.T) {
	go createRandomProductCategory(t)
}

func TestAdminCreateProductCategory(t *testing.T) {
	go adminCreateRandomProductCategoryForUpdateOrDelete(t)
}

func TestCreateProductCategoryParent(t *testing.T) {
	productCategory := createRandomProductCategoryParent(t)
	arg1 := DeleteProductCategoryParams{
		ID:               productCategory.ID,
		ParentCategoryID: productCategory.ParentCategoryID,
	}

	err := testStore.DeleteProductCategory(context.Background(), arg1)

	require.NoError(t, err)
}

func TestGetProductCategory(t *testing.T) {
	productCategory1 := createRandomProductCategoryParent(t)
	productCategory2, err := testStore.GetProductCategory(context.Background(), productCategory1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, productCategory2)

	require.Equal(t, productCategory1.ID, productCategory2.ID)
	require.Equal(t, productCategory1.ParentCategoryID, productCategory2.ParentCategoryID)
	require.Equal(t, productCategory1.CategoryName, productCategory2.CategoryName)

	arg1 := DeleteProductCategoryParams{
		ID:               productCategory1.ID,
		ParentCategoryID: productCategory1.ParentCategoryID,
	}

	err = testStore.DeleteProductCategory(context.Background(), arg1)

	require.NoError(t, err)
}

func TestGetProductCategoryByParent(t *testing.T) {
	productCategory1 := createRandomProductCategoryParent(t)

	arg := GetProductCategoryByParentParams{
		ID:               productCategory1.ID,
		ParentCategoryID: null.IntFromPtr(&productCategory1.ParentCategoryID.Int64),
	}
	productCategory2, err := testStore.GetProductCategoryByParent(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, productCategory2)

	require.Equal(t, productCategory1.ID, productCategory2.ID)
	require.Equal(t, productCategory1.ParentCategoryID, productCategory2.ParentCategoryID)
	require.Equal(t, productCategory1.CategoryName, productCategory2.CategoryName)

	arg1 := DeleteProductCategoryParams{
		ID:               productCategory1.ID,
		ParentCategoryID: productCategory1.ParentCategoryID,
	}

	err = testStore.DeleteProductCategory(context.Background(), arg1)

	require.NoError(t, err)
}

func TestUpdateProductCategory(t *testing.T) {
	productCategory1 := createRandomProductCategoryForUpdateOrDelete(t)
	arg := UpdateProductCategoryParams{
		ID:           productCategory1.ID,
		CategoryName: util.RandomString(5),
	}

	productCategory2, err := testStore.UpdateProductCategory(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, productCategory2)

	require.Equal(t, productCategory1.ID, productCategory2.ID)
	require.Empty(t, productCategory1.ParentCategoryID)
	require.Empty(t, productCategory2.ParentCategoryID)
	require.NotEqual(t, productCategory1.CategoryName, productCategory2.CategoryName)

	arg1 := DeleteProductCategoryParams{
		ID: productCategory2.ID,
	}

	err = testStore.DeleteProductCategory(context.Background(), arg1)

	require.NoError(t, err)
}

func TestUpdateProductCategoryParent(t *testing.T) {
	productCategory1 := createRandomProductCategoryParent(t)
	arg := UpdateProductCategoryParams{
		ID:               productCategory1.ID,
		ParentCategoryID: null.IntFromPtr(&productCategory1.ParentCategoryID.Int64),
		CategoryName:     util.RandomString(5),
	}

	productCategory2, err := testStore.UpdateProductCategory(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, productCategory2)

	require.Equal(t, productCategory1.ID, productCategory2.ID)
	require.Equal(t, productCategory1.ParentCategoryID.Int64, productCategory2.ParentCategoryID.Int64)
	require.NotEqual(t, productCategory1.CategoryName, productCategory2.CategoryName)

	arg1 := DeleteProductCategoryParams{
		ID:               productCategory2.ID,
		ParentCategoryID: productCategory2.ParentCategoryID,
	}

	err = testStore.DeleteProductCategory(context.Background(), arg1)

	require.NoError(t, err)
}

func TestDeleteProductCategory(t *testing.T) {
	productCategory1 := createRandomProductCategoryForUpdateOrDelete(t)

	arg := DeleteProductCategoryParams{
		ID: productCategory1.ID,
	}
	err := testStore.DeleteProductCategory(context.Background(), arg)

	require.NoError(t, err)

	productCategory2, err := testStore.GetProductCategory(context.Background(), productCategory1.ID)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, productCategory2)

}

func TestDeleteProductCategoryParent(t *testing.T) {
	productCategory1 := createRandomProductCategoryForUpdateOrDelete(t)

	arg1 := DeleteProductCategoryParams{
		ID:               productCategory1.ID,
		ParentCategoryID: null.IntFromPtr(&productCategory1.ParentCategoryID.Int64),
	}
	err := testStore.DeleteProductCategory(context.Background(), arg1)

	require.NoError(t, err)

	arg2 := GetProductCategoryByParentParams{
		ID:               productCategory1.ID,
		ParentCategoryID: null.IntFromPtr(&productCategory1.ParentCategoryID.Int64),
	}

	productCategory2, err := testStore.GetProductCategoryByParent(context.Background(), arg2)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, productCategory2)

}

func TestListProductCategories(t *testing.T) {
	for i := 0; i < 5; i++ {
		createRandomProductCategory(t)
	}
	// arg := ListProductCategoriesParams{
	// 	Limit:  5,
	// 	Offset: 0,
	// }

	productCategories, err := testStore.ListProductCategories(context.Background())
	require.NoError(t, err)
	// require.Len(t, productCategories, 5)

	for _, productCategory := range productCategories {
		require.NotEmpty(t, productCategory)

	}
}
