package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/cshop/v3/util"
	"github.com/stretchr/testify/require"
)

func createRandomUserReview(t *testing.T) UserReview {
	user := createRandomUser(t)
	shopOrderItem := createRandomShopOrderItem(t)
	arg := CreateUserReviewParams{
		UserID:           user.ID,
		OrderedProductID: shopOrderItem.ID,
		RatingValue:      int32(util.RandomInt(1, 5)),
	}
	userReview, err := testQueires.CreateUserReview(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, userReview)

	require.Equal(t, arg.UserID, userReview.UserID)
	require.Equal(t, arg.OrderedProductID, userReview.OrderedProductID)
	require.Equal(t, arg.RatingValue, userReview.RatingValue)

	return userReview
}
func TestCreateUserReview(t *testing.T) {
	createRandomUserReview(t)
}

func TestGetUserReview(t *testing.T) {
	userReview1 := createRandomUserReview(t)
	userReview2, err := testQueires.GetUserReview(context.Background(), userReview1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, userReview2)

	require.Equal(t, userReview1.ID, userReview2.ID)
	require.Equal(t, userReview1.UserID, userReview2.UserID)
	require.Equal(t, userReview1.OrderedProductID, userReview2.OrderedProductID)
	require.Equal(t, userReview1.RatingValue, userReview2.RatingValue)
}

func TestUpdateUserReviewRating(t *testing.T) {
	userReview1 := createRandomUserReview(t)
	arg := UpdateUserReviewParams{
		UserID:           sql.NullInt64{},
		OrderedProductID: sql.NullInt64{},
		RatingValue: sql.NullInt32{
			Int32: 0,
			Valid: true,
		},
		ID: userReview1.ID,
	}

	userReview2, err := testQueires.UpdateUserReview(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, userReview2)

	require.Equal(t, userReview1.ID, userReview2.ID)
	require.Equal(t, userReview1.UserID, userReview2.UserID)
	require.Equal(t, userReview1.OrderedProductID, userReview2.OrderedProductID)
	require.NotEqual(t, userReview1.RatingValue, userReview2.RatingValue)
}

func TestDeleteUserReview(t *testing.T) {
	userReview1 := createRandomUserReview(t)
	err := testQueires.DeleteUserReview(context.Background(), userReview1.ID)

	require.NoError(t, err)

	userReview2, err := testQueires.GetUserReview(context.Background(), userReview1.ID)

	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, userReview2)

}

func TestListUserReviews(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomUserReview(t)
	}
	arg := ListUserReviewsParams{
		Limit:  5,
		Offset: 5,
	}

	UserReviews, err := testQueires.ListUserReviews(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, UserReviews)

	for _, UserReview := range UserReviews {
		require.NotEmpty(t, UserReview)
	}

}
