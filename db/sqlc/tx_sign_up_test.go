package db

import (
	"context"
	"testing"

	"github.com/cshop/v3/util"
	"github.com/stretchr/testify/require"
)

func TestSignUpTx(t *testing.T) {
	t.Parallel()

	user := SignUpTxParams{
		Username: util.RandomUser(),
		Email:    util.RandomUser(),
		Password: util.RandomUser(),
		// Telephone: int32(util.RandomInt(0, 1000000)),
	}

	result, err := testStore.SignUpTx(context.Background(), SignUpTxParams{
		Username: user.Username,
		Email:    user.Email,
		Password: user.Password,
		// Telephone: user.Telephone,
	})

	require.NoError(t, err)
	require.NotEmpty(t, result)
	require.Equal(t, user.Username, result.Username)
	require.Equal(t, user.Email, result.Email)
	// require.Equal(t, user.Telephone, result.Telephone)
	// require.Equal(t, user.Password, result.Password)
	require.False(t, result.IsBlocked)
	require.False(t, result.IsEmailVerified)
	// require.Empty(t, result.DefaultPayment)
	require.NotEmpty(t, result.SecretCode)

}
