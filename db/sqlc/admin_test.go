package db

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/cshop/v3/util"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/require"
)

func createRandomAdmin(t *testing.T) Admin {
	hashedPassword, err := util.HashPassword(util.RandomString(6))
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword)

	adminType := createRandomAdminType(t)

	arg := CreateAdminParams{
		Username: util.RandomUser(),
		Email:    util.RandomEmail(),
		Password: hashedPassword,
		TypeID:   adminType.ID,
	}
	admin, err := testQueires.CreateAdmin(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, admin)

	require.Equal(t, arg.Username, admin.Username)
	require.Equal(t, arg.Email, admin.Email)
	require.Equal(t, arg.Password, admin.Password)
	require.Equal(t, arg.TypeID, admin.TypeID)

	require.NotZero(t, admin.ID)
	require.NotZero(t, admin.CreatedAt)
	require.True(t, admin.UpdatedAt.IsZero())

	return admin

}
func TestCreateAdmin(t *testing.T) {
	createRandomAdmin(t)
}

func TestGetAdmin(t *testing.T) {
	admin1 := createRandomAdmin(t)
	admin2, err := testQueires.GetAdmin(context.Background(), admin1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, admin2)

	require.Equal(t, admin1.ID, admin2.ID)
	require.Equal(t, admin1.Username, admin2.Username)
	require.Equal(t, admin1.Email, admin2.Email)
	require.Equal(t, admin1.Password, admin2.Password)
	require.Equal(t, admin1.TypeID, admin2.TypeID)
	require.True(t, admin2.Active)
	require.WithinDuration(t, admin1.CreatedAt, admin2.CreatedAt, time.Second)
	require.WithinDuration(t, admin1.UpdatedAt, admin2.UpdatedAt, time.Second)
}

func TestUpdateAdmin(t *testing.T) {
	admin1 := createRandomAdmin(t)

	arg := UpdateAdminParams{
		ID:     admin1.ID,
		Active: false,
	}

	admin2, err := testQueires.UpdateAdmin(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, admin2)

	require.Equal(t, admin1.ID, admin2.ID)
	require.Equal(t, admin1.Username, admin2.Username)
	require.Equal(t, admin1.Email, admin2.Email)
	require.Equal(t, admin1.Password, admin2.Password)
	require.Equal(t, arg.Active, admin2.Active)
	require.Equal(t, admin1.CreatedAt, admin2.CreatedAt, time.Second)
	require.NotEqual(t, admin1.UpdatedAt, admin2.UpdatedAt, time.Second)
}
func TestDeleteAdmin(t *testing.T) {
	admin1 := createRandomAdmin(t)

	err := testQueires.DeleteAdmin(context.Background(), admin1.ID)

	require.NoError(t, err)

	admin2, err := testQueires.GetAdmin(context.Background(), admin1.ID)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, admin2)
}

func TestListAdmins(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			createRandomAdmin(t)
			wg.Done()
		}()
	}
	wg.Wait()
	arg := ListAdminsParams{
		Limit:  5,
		Offset: 0,
	}

	admins, err := testQueires.ListAdmins(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, admins, 5)

	for _, Admin := range admins {
		require.NotEmpty(t, Admin)

	}
}
