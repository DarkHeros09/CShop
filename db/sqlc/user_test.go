package db

import (
	"context"
	"testing"
	"time"

	"github.com/cshop/v3/util"
	"github.com/guregu/null"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/require"
)

func createRandomUser(t *testing.T) User {
	hashedPassword, err := util.HashPassword(util.RandomString(6))
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword)

	arg := CreateUserParams{
		Username:  util.RandomUser(),
		Email:     util.RandomEmail(),
		Password:  hashedPassword,
		Telephone: int32(util.RandomInt(0, 1000000)),
	}

	user, err := testQueires.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.Email, user.Email)
	require.Equal(t, arg.Password, user.Password)
	require.Equal(t, arg.Telephone, user.Telephone)

	require.NotZero(t, user.ID)
	require.NotZero(t, user.CreatedAt)
	require.True(t, user.UpdatedAt.IsZero())

	return user

}
func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestCreateUserWithCart(t *testing.T) {
	hashedPassword, err := util.HashPassword(util.RandomString(6))
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword)

	arg := CreateUserWithCartParams{
		Username:  util.RandomUser(),
		Email:     util.RandomEmail(),
		Password:  hashedPassword,
		Telephone: int32(util.RandomInt(0, 1000000)),
	}

	user, err := testQueires.CreateUserWithCart(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.Email, user.Email)
	require.Equal(t, arg.Password, user.Password)
	require.Equal(t, arg.Telephone, user.Telephone)

	require.NotZero(t, user.ID)
	require.NotZero(t, user.CreatedAt)
	require.True(t, user.UpdatedAt.IsZero())

	shoppingCart, err := testQueires.GetShoppingCartByUserID(context.Background(), user.ID)
	require.NoError(t, err)
	require.NotEmpty(t, shoppingCart)

}

func TestGetUser(t *testing.T) {
	user1 := createRandomUser(t)
	user2, err := testQueires.GetUser(context.Background(), user1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, user1.ID, user2.ID)
	require.Equal(t, user1.Email, user2.Email)
	require.Equal(t, user1.Password, user2.Password)
	require.Equal(t, user1.Telephone, user2.Telephone)
	require.WithinDuration(t, user1.CreatedAt, user2.CreatedAt, time.Second)
	require.WithinDuration(t, user1.UpdatedAt, user2.UpdatedAt, time.Second)
}

func TestGetUserByEmail(t *testing.T) {
	user1 := createRandomUser(t)
	user2, err := testQueires.GetUserByEmail(context.Background(), user1.Email)

	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, user1.ID, user2.ID)
	require.Equal(t, user1.Email, user2.Email)
	require.Equal(t, user1.Password, user2.Password)
	require.Equal(t, user1.Telephone, user2.Telephone)
	require.WithinDuration(t, user1.CreatedAt, user2.CreatedAt, time.Second)
	require.WithinDuration(t, user1.UpdatedAt, user2.UpdatedAt, time.Second)
}

func TestUpdateUser(t *testing.T) {
	user1 := createRandomUser(t)

	arg := UpdateUserParams{
		ID:        user1.ID,
		Telephone: null.IntFrom(util.RandomInt(0, 1000000)),
	}

	user2, err := testQueires.UpdateUser(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, user1.ID, user2.ID)
	require.Equal(t, user1.Email, user2.Email)
	require.Equal(t, user1.Password, user2.Password)
	require.Equal(t, int32(arg.Telephone.Int64), user2.Telephone)
	require.Equal(t, user1.CreatedAt, user2.CreatedAt, time.Second)
	require.NotEqual(t, user1.UpdatedAt, user2.UpdatedAt, time.Second)
}
func TestDeleteUser(t *testing.T) {
	user1 := createRandomUser(t)

	user2, err := testQueires.DeleteUser(context.Background(), user1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, user2)
	require.Equal(t, user1, user2)

	user3, err := testQueires.DeleteUser(context.Background(), user1.ID)

	require.Error(t, err)
	require.Empty(t, user3)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
}

func TestListUsers(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomUser(t)
	}
	arg := ListUsersParams{
		Limit:  5,
		Offset: 5,
	}

	users, err := testQueires.ListUsers(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, users, 5)

	for _, user := range users {
		require.NotEmpty(t, user)

	}
}
