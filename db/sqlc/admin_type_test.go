package db

import (
	"context"
	"testing"
	"time"

	"github.com/cshop/v3/util"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func createRandomAdminType(t *testing.T) AdminType {
	arg := util.RandomString(6)

	adminType, err := testStore.CreateAdminType(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, adminType)

	require.Equal(t, arg, adminType.AdminType)

	require.NotEmpty(t, adminType.ID)
	require.NotEmpty(t, adminType.CreatedAt)
	require.True(t, adminType.UpdatedAt.IsZero())

	return adminType
}
func TestCreateAdminType(t *testing.T) {
	go createRandomAdminType(t)
}

func TestGetAdminType(t *testing.T) {
	adminType1 := createRandomAdminType(t)
	adminType2, err := testStore.GetAdminType(context.Background(), adminType1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, adminType2)

	require.Equal(t, adminType1.ID, adminType2.ID)
	require.Equal(t, adminType1.AdminType, adminType2.AdminType)
	require.Equal(t, adminType1.CreatedAt, adminType2.CreatedAt, time.Second)
	require.Equal(t, adminType1.UpdatedAt, adminType2.UpdatedAt, time.Second)
}

func TestUpdateAdminType(t *testing.T) {
	adminType1 := createRandomAdminType(t)
	arg := UpdateAdminTypeParams{
		ID:        adminType1.ID,
		AdminType: util.RandomString(6),
	}

	adminType2, err := testStore.UpdateAdminType(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, adminType2)

	require.Equal(t, adminType1.ID, adminType2.ID)
	require.NotEqual(t, adminType1.AdminType, adminType2.AdminType)
	require.Equal(t, adminType1.CreatedAt, adminType2.CreatedAt, time.Second)
	require.NotEqual(t, adminType1.UpdatedAt, adminType2.UpdatedAt, time.Second)
}

func TestDeleteAdminTypeByID(t *testing.T) {
	adminType1 := createRandomAdminType(t)
	err := testStore.DeleteAdminTypeByID(context.Background(), adminType1.ID)

	require.NoError(t, err)

	adminType2, err := testStore.GetAdminType(context.Background(), adminType1.ID)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, adminType2)

}

func TestDeleteAdminTypeByType(t *testing.T) {
	adminType1 := createRandomAdminType(t)
	err := testStore.DeleteAdminTypeByType(context.Background(), adminType1.AdminType)

	require.NoError(t, err)

	adminType2, err := testStore.GetAdminType(context.Background(), adminType1.ID)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, adminType2)

}

func TestListAdminTypes(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomAdminType(t)
	}
	arg := ListAdminTypesParams{
		Limit:  5,
		Offset: 5,
	}

	adminTypes, err := testStore.ListAdminTypes(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, adminTypes, 5)

	for _, adminType := range adminTypes {
		require.NotEmpty(t, adminType)

	}
}
