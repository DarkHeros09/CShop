package db

import (
	"context"
	"testing"

	"github.com/cshop/v3/util"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func createRandomNotification(t *testing.T) Notification {
	t.Helper()
	user1 := createRandomUser(t)
	arg := CreateNotificationParams{
		UserID:   user1.ID,
		DeviceID: null.StringFrom(util.RandomString(100)),
		FcmToken: null.StringFrom(util.RandomString(50)),
	}

	notification, err := testStore.CreateNotification(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, notification)

	require.Equal(t, arg.UserID, notification.UserID)
	require.Equal(t, arg.DeviceID, notification.DeviceID)
	require.Equal(t, arg.FcmToken, notification.FcmToken)

	require.NotEmpty(t, notification.CreatedAt)

	return notification
}

func TestCreateNotification(t *testing.T) {
	t.Parallel()
	createRandomNotification(t)
}

func TestGetNotification(t *testing.T) {
	t.Parallel()
	notification1 := createRandomNotification(t)

	arg := GetNotificationParams{
		UserID:   notification1.UserID,
		DeviceID: notification1.DeviceID,
	}

	notification2, err := testStore.GetNotification(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, notification2)

	require.Equal(t, notification1.UserID, notification2.UserID)
	require.Equal(t, notification1.DeviceID, notification2.DeviceID)
	require.Equal(t, notification1.FcmToken, notification2.FcmToken)
	require.Equal(t, notification1.CreatedAt, notification2.CreatedAt)
	require.Equal(t, notification1.UpdatedAt, notification2.UpdatedAt)

}
func TestGetNotificationV2(t *testing.T) {
	t.Parallel()
	notification1 := createRandomNotification(t)

	notification2, err := testStore.GetNotificationV2(context.Background(), notification1.UserID)
	require.NoError(t, err)
	require.NotEmpty(t, notification2)

	require.Equal(t, notification1.UserID, notification2.UserID)
	require.Equal(t, notification1.DeviceID, notification2.DeviceID)
	require.Equal(t, notification1.FcmToken, notification2.FcmToken)
	require.Equal(t, notification1.CreatedAt, notification2.CreatedAt)
	require.Equal(t, notification1.UpdatedAt, notification2.UpdatedAt)

}

func TestUpdateNotification(t *testing.T) {
	notification1 := createRandomNotification(t)

	arg := UpdateNotificationParams{
		FcmToken: null.StringFrom(util.RandomString(50)),
		UserID:   notification1.UserID,
		DeviceID: notification1.DeviceID,
	}

	notification2, err := testStore.UpdateNotification(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, notification2)

	require.Equal(t, notification1.UserID, notification2.UserID)
	require.Equal(t, notification1.DeviceID, notification2.DeviceID)
	require.NotEqual(t, notification1.FcmToken, notification2.FcmToken)
	require.Equal(t, notification1.CreatedAt, notification2.CreatedAt)
	require.NotEqual(t, notification1.UpdatedAt, notification2.UpdatedAt)
}

func TestDeleteNotification(t *testing.T) {
	notification1 := createRandomNotification(t)

	arg := DeleteNotificationParams{
		UserID:   notification1.UserID,
		DeviceID: notification1.DeviceID,
	}

	notification2, err := testStore.DeleteNotification(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, notification2)

	notification3, err := testStore.DeleteNotification(context.Background(), arg)

	require.Error(t, err)
	require.Empty(t, notification3)
	require.EqualError(t, err, pgx.ErrNoRows.Error())

}
