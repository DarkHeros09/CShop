package db

import (
	"context"
	"testing"
	"time"

	"github.com/cshop/v3/util"
	"github.com/google/uuid"
	"github.com/guregu/null/v5"
	"github.com/stretchr/testify/require"
)

func createRandomAdminSession(t *testing.T) AdminSession {
	t.Helper()
	admin1 := createRandomAdmin(t)
	arg := CreateAdminSessionParams{
		ID:           uuid.New(),
		AdminID:      admin1.ID,
		RefreshToken: util.RandomString(100),
		AdminAgent:   util.RandomString(6),
		ClientIp:     util.RandomString(10),
		IsBlocked:    false,
		ExpiresAt:    time.Now().Local().UTC(),
	}

	adminSession, err := testStore.CreateAdminSession(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, adminSession)

	require.Equal(t, arg.ID, adminSession.ID)
	require.Equal(t, arg.AdminID, adminSession.AdminID)
	require.Equal(t, arg.RefreshToken, adminSession.RefreshToken)
	require.Equal(t, arg.AdminAgent, adminSession.AdminAgent)
	require.Equal(t, arg.ClientIp, adminSession.ClientIp)
	require.Equal(t, arg.IsBlocked, adminSession.IsBlocked)
	require.WithinDuration(t, arg.ExpiresAt, adminSession.ExpiresAt, time.Second)

	require.NotEmpty(t, adminSession.CreatedAt)

	return *adminSession
}

func TestCreateAdminSession(t *testing.T) {
	t.Parallel()
	createRandomAdminSession(t)
}

func TestGetAdminSession(t *testing.T) {
	t.Parallel()
	adminSession1 := createRandomAdminSession(t)

	adminSession2, err := testStore.GetAdminSession(context.Background(), adminSession1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, adminSession2)

	require.Equal(t, adminSession1.ID, adminSession2.ID)
	require.Equal(t, adminSession1.AdminID, adminSession2.AdminID)
	require.Equal(t, adminSession1.RefreshToken, adminSession2.RefreshToken)
	require.Equal(t, adminSession1.AdminAgent, adminSession2.AdminAgent)
	require.Equal(t, adminSession1.ClientIp, adminSession2.ClientIp)
	require.Equal(t, adminSession1.IsBlocked, adminSession2.IsBlocked)
	require.Equal(t, adminSession1.CreatedAt, adminSession2.CreatedAt)
	require.Equal(t, adminSession1.ExpiresAt, adminSession2.ExpiresAt)
}

func TestUpdateAdminSession(t *testing.T) {
	adminSession1 := createRandomAdminSession(t)

	arg := UpdateAdminSessionParams{
		IsBlocked:    null.BoolFrom(!adminSession1.IsBlocked),
		ID:           adminSession1.ID,
		AdminID:      adminSession1.AdminID,
		RefreshToken: adminSession1.RefreshToken,
	}

	adminSession2, err := testStore.UpdateAdminSession(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, adminSession2)

	require.Equal(t, adminSession1.ID, adminSession2.ID)
	require.Equal(t, adminSession1.AdminID, adminSession2.AdminID)
	require.Equal(t, adminSession1.RefreshToken, adminSession2.RefreshToken)
	require.Equal(t, adminSession1.AdminAgent, adminSession2.AdminAgent)
	require.Equal(t, adminSession1.ClientIp, adminSession2.ClientIp)
	require.NotEqual(t, adminSession1.IsBlocked, adminSession2.IsBlocked)
	require.Equal(t, adminSession1.CreatedAt, adminSession2.CreatedAt)
	require.NotEqual(t, adminSession1.UpdatedAt, adminSession2.UpdatedAt)
	require.Equal(t, adminSession1.ExpiresAt, adminSession2.ExpiresAt)
}
