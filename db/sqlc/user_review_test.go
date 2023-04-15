package db

import (
	"context"
	"sync"
	"testing"

	"github.com/cshop/v3/util"
	"github.com/guregu/null"
	"github.com/jackc/pgx/v4"
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

	arg := GetUserReviewParams{
		ID:     userReview1.ID,
		UserID: userReview1.UserID,
	}
	userReview2, err := testQueires.GetUserReview(context.Background(), arg)

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
		UserID:           userReview1.UserID,
		OrderedProductID: null.Int{},
		RatingValue:      null.IntFrom(0),
		ID:               userReview1.ID,
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

	arg1 := DeleteUserReviewParams{
		ID:     userReview1.ID,
		UserID: userReview1.UserID,
	}
	_, err := testQueires.DeleteUserReview(context.Background(), arg1)

	require.NoError(t, err)

	arg := GetUserReviewParams{
		ID:     userReview1.ID,
		UserID: userReview1.UserID,
	}
	userReview2, err := testQueires.GetUserReview(context.Background(), arg)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, userReview2)

}

func TestListUserReviews(t *testing.T) {
	lastUserReviewChan := make(chan UserReview)
	var wg sync.WaitGroup
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go func(i int) {
			lastUserReview := createRandomUserReview(t)
			wg.Done()
			if i == 4 {

				lastUserReviewChan <- lastUserReview
			}
		}(i)
	}
	lastUserReview := <-lastUserReviewChan
	wg.Wait()
	arg := ListUserReviewsParams{
		Limit:  5,
		Offset: 0,
		UserID: lastUserReview.UserID,
	}

	userReviewsChan := make(chan []UserReview)
	errChan := make(chan error)
	go func() {
		userReviews, err := testQueires.ListUserReviews(context.Background(), arg)
		userReviewsChan <- userReviews
		errChan <- err
	}()
	userReviews := <-userReviewsChan
	err := <-errChan

	require.NoError(t, err)
	require.NotEmpty(t, userReviews)

	for _, userReview := range userReviews {
		require.Equal(t, userReview.ID, userReviews[len(userReviews)-1].ID)
		require.Equal(t, userReview.OrderedProductID, userReviews[len(userReviews)-1].OrderedProductID)
		require.Equal(t, userReview.RatingValue, userReviews[len(userReviews)-1].RatingValue)
		require.Equal(t, userReview.UserID, userReviews[len(userReviews)-1].UserID)
	}

}
