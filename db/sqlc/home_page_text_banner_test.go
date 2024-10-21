package db

import (
	"context"
	"testing"

	"github.com/cshop/v3/util"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func createRandomHomePageTextBanner(t *testing.T) HomePageTextBanner {
	admin := createRandomAdmin(t)

	arg := CreateHomePageTextBannerParams{
		Name:        util.RandomUser(),
		Description: util.RandomUser(),
		AdminID:     admin.ID,
	}
	productImage, err := testStore.CreateHomePageTextBanner(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, productImage)

	require.Equal(t, arg.Name, productImage.Name)
	require.Equal(t, arg.Description, productImage.Description)

	return productImage
}
func TestCreateHomePageTextBanner(t *testing.T) {
	createRandomHomePageTextBanner(t)
}

func TestGetHomePageTextBanner(t *testing.T) {
	productImage1 := createRandomHomePageTextBanner(t)
	productImage2, err := testStore.GetHomePageTextBanner(context.Background(), productImage1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, productImage2)

	require.Equal(t, productImage1.ID, productImage2.ID)
	require.Equal(t, productImage1.Name, productImage2.Name)
	require.Equal(t, productImage1.Description, productImage2.Description)

}

func TestUpdateHomePageTextBanner(t *testing.T) {
	admin := createRandomAdmin(t)
	productImage1 := createRandomHomePageTextBanner(t)
	arg := UpdateHomePageTextBannerParams{
		ID:          productImage1.ID,
		Description: null.StringFrom(util.RandomUser()),
		AdminID:     admin.ID,
	}

	productImage2, err := testStore.UpdateHomePageTextBanner(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, productImage2)

	require.Equal(t, productImage1.ID, productImage2.ID)
	require.Equal(t, productImage1.Name, productImage2.Name)
	require.NotEqual(t, productImage1.Description, productImage2.Description)
}

func TestDeleteHomePageTextBanner(t *testing.T) {
	admin := createRandomAdmin(t)
	productImage1 := createRandomHomePageTextBanner(t)
	arg := DeleteHomePageTextBannerParams{
		ID:      productImage1.ID,
		AdminID: admin.ID,
	}
	err := testStore.DeleteHomePageTextBanner(context.Background(), arg)

	require.NoError(t, err)

	productImage2, err := testStore.GetHomePageTextBanner(context.Background(), productImage1.ID)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, productImage2)

}

func TestListHomePageTextBanners(t *testing.T) {
	for i := 0; i < 5; i++ {
		createRandomHomePageTextBanner(t)
	}

	textBanners, err := testStore.ListHomePageTextBanners(context.Background())

	require.NoError(t, err)
	require.NotEmpty(t, textBanners)

	for _, textBanner := range textBanners {
		require.NotEmpty(t, textBanner)
	}

}
