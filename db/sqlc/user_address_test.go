package db

import (
	"context"
	"testing"

	"github.com/guregu/null/v5"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func createRandomUserAddress(t *testing.T) UserAddress {
	user1 := createRandomUser(t)
	address1 := createRandomAddress(t)

	arg := CreateUserAddressParams{
		UserID:    user1.ID,
		AddressID: address1.ID,
		// DefaultAddress: null.IntFromPtr(&address1.ID),
	}

	userAddress, err := testStore.CreateUserAddress(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, userAddress)

	require.Equal(t, arg.UserID, userAddress.UserID)
	require.Equal(t, arg.AddressID, userAddress.AddressID)
	// require.Equal(t, arg.DefaultAddress, userAddress.DefaultAddress)

	return userAddress

}

func TestCreateUserAddress(t *testing.T) {
	createRandomUserAddress(t)
}

func createRandomUserAddressWithAddress(t *testing.T) CreateUserAddressWithAddressRow {
	t.Helper()
	user1 := createRandomUser(t)
	address1 := createRandomAddress(t)

	arg := CreateUserAddressWithAddressParams{
		AddressLine:    address1.AddressLine,
		Region:         address1.Region,
		City:           address1.City,
		UserID:         user1.ID,
		DefaultAddress: null.IntFromPtr(&address1.ID),
	}

	userAddress, err := testStore.CreateUserAddressWithAddress(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, userAddress)

	require.Equal(t, arg.UserID, userAddress.UserID)
	require.Equal(t, arg.AddressLine, userAddress.AddressLine)
	require.Equal(t, arg.Region, userAddress.Region)
	require.Equal(t, arg.City, userAddress.City)
	require.Equal(t, arg.DefaultAddress, userAddress.DefaultAddress)

	return userAddress

}

func TestCreateUserAddressWithAddress(t *testing.T) {
	t.Parallel()
	createRandomUserAddressWithAddress(t)
}

func TestGetUserAddress(t *testing.T) {
	t.Parallel()
	userAddress1 := createRandomUserAddress(t)
	arg := GetUserAddressParams{
		UserID:    userAddress1.UserID,
		AddressID: userAddress1.AddressID,
	}
	userAddress2, err := testStore.GetUserAddress(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, userAddress2)

	require.Equal(t, userAddress1.UserID, userAddress2.UserID)
	require.Equal(t, userAddress1.AddressID, userAddress2.AddressID)
	require.Equal(t, userAddress1.DefaultAddress, userAddress2.DefaultAddress)

}

func TestGetUserAddressWithAddress(t *testing.T) {
	userAddress1 := createRandomUserAddressWithAddress(t)

	arg := GetUserAddressWithAddressParams{
		UserID:    userAddress1.UserID,
		AddressID: userAddress1.AddressID,
	}

	userAddress2, err := testStore.GetUserAddressWithAddress(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, userAddress2)

	require.Equal(t, userAddress1.UserID, userAddress2.UserID)
	require.Equal(t, userAddress1.AddressID, userAddress2.AddressID)
	require.Equal(t, userAddress1.AddressLine, userAddress2.AddressLine)
	require.Equal(t, userAddress1.Region, userAddress2.Region)
	require.Equal(t, userAddress1.City, userAddress2.City)

}

func TestCheckUserAddressDefaultAddress(t *testing.T) {
	userAddress1 := createRandomUserAddressWithAddress(t)

	userAddressLength, err := testStore.CheckUserAddressDefaultAddress(context.Background(), userAddress1.UserID)
	require.NoError(t, err)
	require.NotEmpty(t, userAddressLength)

	require.Greater(t, userAddressLength, int64(0))

	arg := DeleteUserAddressParams{
		UserID:    userAddress1.UserID,
		AddressID: userAddress1.AddressID,
	}
	userAddress2, err := testStore.DeleteUserAddress(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, userAddress2)

	userAddressLength2, err := testStore.CheckUserAddressDefaultAddress(context.Background(), userAddress1.UserID)
	require.NoError(t, err)
	require.Empty(t, userAddressLength2)

	require.Equal(t, userAddressLength2, int64(0))
}

func TestUpdateUserAddress(t *testing.T) {
	userAddress1 := createRandomUserAddress(t)
	arg := UpdateUserAddressParams{
		UserID:         userAddress1.UserID,
		AddressID:      userAddress1.AddressID,
		DefaultAddress: null.IntFromPtr(&userAddress1.DefaultAddress.Int64),
	}

	userAddress2, err := testStore.UpdateUserAddress(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, userAddress2)

	require.Equal(t, userAddress1.UserID, userAddress2.UserID)
	require.Equal(t, arg.AddressID, userAddress2.AddressID)
	require.Equal(t, arg.DefaultAddress, userAddress2.DefaultAddress)

}

func TestDeleteUserAddress(t *testing.T) {
	userAddress1 := createRandomUserAddress(t)
	arg := DeleteUserAddressParams{
		UserID:    userAddress1.UserID,
		AddressID: userAddress1.AddressID,
	}
	userAddress2, err := testStore.DeleteUserAddress(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, userAddress2)

	userAddress3, err := testStore.DeleteUserAddress(context.Background(), arg)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, userAddress3)
}

func TestListUserAddresses(t *testing.T) {
	t.Parallel()
	lastUserAddressChan := make(chan UserAddress)
	for i := 0; i < 10; i++ {
		go func(i int) {
			if i == 9 {
				lastUserAddress := createRandomUserAddress(t)
				lastUserAddressChan <- lastUserAddress
			} else {

				createRandomUserAddress(t)
			}

		}(i)
	}
	lastUserAddress := <-lastUserAddressChan
	arg := ListUserAddressesParams{
		UserID: lastUserAddress.UserID,
		Limit:  5,
		Offset: 0,
	}

	userAddresses, err := testStore.ListUserAddresses(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, userAddresses, 1)

	for _, userAddress1 := range userAddresses {
		userAddress := userAddress1
		require.NotEmpty(t, userAddress)
		require.Equal(t, lastUserAddress.UserID, userAddress.UserID)
		require.Equal(t, lastUserAddress.AddressID, userAddress.AddressID)
		require.Equal(t, lastUserAddress.DefaultAddress, userAddress.DefaultAddress)

	}
}
