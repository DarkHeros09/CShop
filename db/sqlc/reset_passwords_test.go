package db

import (
	"context"
	"testing"
	"time"

	"github.com/cshop/v3/util"
	"github.com/stretchr/testify/require"
)

func createRandomResetPassword(t *testing.T, user User) ResetPassword {
	t.Helper()
	arg := CreateResetPasswordParams{
		UserID: user.ID,
		// Email:      user.Email,
		SecretCode: util.RandomString(6),
	}

	resetPassword, err := testStore.CreateResetPassword(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, resetPassword)

	require.Equal(t, arg.UserID, resetPassword.UserID)
	// require.Equal(t, arg.Email, resetPassword.Email)
	require.Equal(t, arg.SecretCode, resetPassword.SecretCode)
	require.Equal(t, resetPassword.IsUsed, false)

	require.NotEmpty(t, resetPassword.CreatedAt)

	return resetPassword
}

func TestCreateResetPassword(t *testing.T) {
	t.Parallel()
	user1 := createRandomUser(t)
	createRandomResetPassword(t, user1)
}

func TestGetResetPassword(t *testing.T) {
	t.Parallel()
	user1 := createRandomUser(t)
	resetPassword := createRandomResetPassword(t, user1)

	resetPassword2, err := testStore.GetResetPassword(context.Background(), resetPassword.ID)
	require.NoError(t, err)
	require.NotEmpty(t, resetPassword2)

	require.Equal(t, resetPassword.ID, resetPassword2.ID)
	require.Equal(t, resetPassword.UserID, resetPassword2.UserID)
	// require.Equal(t, resetPassword.Email, resetPassword2.Email)
	require.Equal(t, resetPassword.SecretCode, resetPassword2.SecretCode)
	require.Equal(t, resetPassword.IsUsed, resetPassword2.IsUsed)
}

func TestGetResetPasswordByEmail(t *testing.T) {
	t.Parallel()
	user1 := createRandomUser(t)
	resetPassword := createRandomResetPassword(t, user1)

	resetPassword2, err := testStore.GetResetPasswordsByEmail(context.Background(), user1.Email)
	require.NoError(t, err)
	require.NotEmpty(t, resetPassword2)

	require.Equal(t, resetPassword.ID, resetPassword2.ID)
	require.Equal(t, resetPassword.UserID, resetPassword2.UserID)
	// require.Equal(t, resetPassword.Email, resetPassword2.Email)
	require.Equal(t, resetPassword.SecretCode, resetPassword2.SecretCode)
	require.Equal(t, resetPassword.IsUsed, resetPassword2.IsUsed)
}

func TestUpdateResetPassword(t *testing.T) {
	user1 := createRandomUser(t)
	resetPassword := createRandomResetPassword(t, user1)

	arg := UpdateResetPasswordParams{
		ID:         resetPassword.ID,
		SecretCode: resetPassword.SecretCode,
	}

	resetPassword2, err := testStore.UpdateResetPassword(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, resetPassword2)

	updatedPasswordReset1, err := testStore.GetResetPassword(context.Background(), resetPassword.ID)
	require.NoError(t, err)
	require.NotEmpty(t, updatedPasswordReset1)

	require.Equal(t, resetPassword.ID, updatedPasswordReset1.ID)
	require.Equal(t, resetPassword.UserID, updatedPasswordReset1.UserID)
	// require.Equal(t, resetPassword.Email, resetPassword2.Email)
	require.Equal(t, resetPassword.SecretCode, updatedPasswordReset1.SecretCode)
	require.NotEqual(t, resetPassword.IsUsed, updatedPasswordReset1.IsUsed)
}

func TestGetLastUsedResetPassword(t *testing.T) {
	user1 := createRandomUser(t)
	resetPassword1 := createRandomResetPassword(t, user1)
	resetPassword2 := createRandomResetPassword(t, user1)

	arg1 := UpdateResetPasswordParams{
		ID:         resetPassword1.ID,
		SecretCode: resetPassword1.SecretCode,
	}

	updatedPasswordReset1, err := testStore.UpdateResetPassword(context.Background(), arg1)
	require.NoError(t, err)
	require.NotEmpty(t, resetPassword2)

	arg2 := UpdateResetPasswordParams{
		ID:         resetPassword2.ID,
		SecretCode: resetPassword2.SecretCode,
	}

	time.Sleep(time.Second)

	updatedPasswordReset2, err := testStore.UpdateResetPassword(context.Background(), arg2)
	require.NoError(t, err)
	require.NotEmpty(t, resetPassword2)

	lastUsedCode, err := testStore.GetLastUsedResetPassword(context.Background(), user1.Email)
	require.NoError(t, err)
	require.NotEmpty(t, lastUsedCode)

	require.Equal(t, updatedPasswordReset2, lastUsedCode)
	require.NotEqual(t, updatedPasswordReset1, lastUsedCode)
}
