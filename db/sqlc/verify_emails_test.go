package db

import (
	"context"
	"testing"

	"github.com/cshop/v3/util"
	"github.com/guregu/null/v5"
	"github.com/stretchr/testify/require"
)

func createRandomVerifyEmail(t *testing.T, user User) VerifyEmail {
	t.Helper()
	arg := CreateVerifyEmailParams{
		UserID: null.IntFrom(user.ID),
		// Email:      user.Email,
		SecretCode: util.RandomString(6),
	}

	verifyEmail, err := testStore.CreateVerifyEmail(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, verifyEmail)

	require.Equal(t, arg.UserID, verifyEmail.UserID)
	// require.Equal(t, arg.Email, verifyEmail.Email)
	require.Equal(t, arg.SecretCode, verifyEmail.SecretCode)
	require.Equal(t, verifyEmail.IsUsed, false)

	require.NotEmpty(t, verifyEmail.CreatedAt)

	return verifyEmail
}

func TestCreateVerifyEmail(t *testing.T) {
	t.Parallel()
	user1 := createRandomUser(t)
	createRandomVerifyEmail(t, user1)
}

func TestGetVerifyEmail(t *testing.T) {
	t.Parallel()
	user1 := createRandomUser(t)
	verifyEmail := createRandomVerifyEmail(t, user1)

	verifyEmail2, err := testStore.GetVerifyEmail(context.Background(), verifyEmail.ID)
	require.NoError(t, err)
	require.NotEmpty(t, verifyEmail2)

	require.Equal(t, verifyEmail.ID, verifyEmail2.ID)
	require.Equal(t, verifyEmail.UserID, verifyEmail2.UserID)
	// require.Equal(t, verifyEmail.Email, verifyEmail2.Email)
	require.Equal(t, verifyEmail.SecretCode, verifyEmail2.SecretCode)
	require.Equal(t, verifyEmail.IsUsed, verifyEmail2.IsUsed)
}

func TestGetVerifyEmailByEmail(t *testing.T) {
	t.Parallel()
	user1 := createRandomUser(t)
	verifyEmail := createRandomVerifyEmail(t, user1)

	verifyEmail2, err := testStore.GetVerifyEmailByEmail(context.Background(), user1.Email)
	require.NoError(t, err)
	require.NotEmpty(t, verifyEmail2)

	require.Equal(t, verifyEmail.ID, verifyEmail2.ID)
	require.Equal(t, verifyEmail.UserID, verifyEmail2.UserID)
	// require.Equal(t, verifyEmail.Email, verifyEmail2.Email)
	require.Equal(t, verifyEmail.SecretCode, verifyEmail2.SecretCode)
	require.Equal(t, verifyEmail.IsUsed, verifyEmail2.IsUsed)
}

func TestUpdateVerifyEmail(t *testing.T) {
	user1 := createRandomUser(t)
	verifyEmail := createRandomVerifyEmail(t, user1)

	arg := UpdateVerifyEmailParams{
		Email:      user1.Email,
		SecretCode: verifyEmail.SecretCode,
	}

	verifyEmail2, err := testStore.UpdateVerifyEmail(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, verifyEmail2)

	verifyEmail3, err := testStore.GetVerifyEmail(context.Background(), verifyEmail.ID)
	require.NoError(t, err)
	require.NotEmpty(t, verifyEmail3)

	require.Equal(t, verifyEmail.ID, verifyEmail3.ID)
	require.Equal(t, verifyEmail.UserID, verifyEmail3.UserID)
	// require.Equal(t, verifyEmail.Email, verifyEmail2.Email)
	require.Equal(t, verifyEmail.SecretCode, verifyEmail3.SecretCode)
	require.NotEqual(t, verifyEmail.IsUsed, verifyEmail3.IsUsed)
	require.NotEqual(t, user1.IsEmailVerified, verifyEmail2.IsEmailVerified)
}
