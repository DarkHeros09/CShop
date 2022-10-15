package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/cshop/v3/util"
	"github.com/stretchr/testify/require"
)

func createRandomUserAddress(t *testing.T) UserAddress {
	user1 := createRandomUser(t)
	address1 := createRandomAddress(t)

	arg := CreateUserAddressParams{
		UserID:    user1.ID,
		AddressID: address1.ID,
		IsDefault: util.RandomBool(),
	}

	userAddress, err := testQueires.CreateUserAddress(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, userAddress)

	require.Equal(t, arg.UserID, userAddress.UserID)
	require.Equal(t, arg.AddressID, userAddress.AddressID)
	require.Equal(t, arg.IsDefault, userAddress.IsDefault)

	return userAddress

}

func TestCreateUserAddress(t *testing.T) {
	createRandomUserAddress(t)
}

func TestGetUserAddress(t *testing.T) {
	userAddress1 := createRandomUserAddress(t)
	arg := GetUserAddressParams{
		UserID:    userAddress1.UserID,
		AddressID: userAddress1.AddressID,
	}
	userAddress2, err := testQueires.GetUserAddress(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, userAddress2)

	require.Equal(t, userAddress1.UserID, userAddress2.UserID)
	require.Equal(t, userAddress1.AddressID, userAddress2.AddressID)
	require.Equal(t, userAddress1.IsDefault, userAddress2.IsDefault)

}

func TestUpdateUserAddress(t *testing.T) {
	userAddress1 := createRandomUserAddress(t)
	arg := UpdateUserAddressParams{
		UserID:    userAddress1.UserID,
		AddressID: userAddress1.AddressID,
		IsDefault: !userAddress1.IsDefault,
	}

	userAddress2, err := testQueires.UpdateUserAddress(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, userAddress2)

	require.Equal(t, userAddress1.UserID, userAddress2.UserID)
	require.Equal(t, arg.AddressID, userAddress2.AddressID)
	require.Equal(t, arg.IsDefault, userAddress2.IsDefault)

}

func TestDeleteUserAddress(t *testing.T) {
	useraddress1 := createRandomUserAddress(t)
	arg := DeleteUserAddressParams{
		UserID:    useraddress1.UserID,
		AddressID: useraddress1.AddressID,
	}
	err := testQueires.DeleteUserAddress(context.Background(), arg)

	require.NoError(t, err)

	arg1 := GetUserAddressParams{
		UserID:    useraddress1.UserID,
		AddressID: useraddress1.AddressID,
	}
	useraddress2, err := testQueires.GetUserAddress(context.Background(), arg1)

	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, useraddress2)

}

func TestListUserAddresses(t *testing.T) {
	var lastUserAddress UserAddress
	for i := 0; i < 10; i++ {
		lastUserAddress = createRandomUserAddress(t)
	}
	arg := ListUserAddressesParams{
		UserID: lastUserAddress.UserID,
		Limit:  5,
		Offset: 0,
	}

	userAddresses, err := testQueires.ListUserAddresses(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, userAddresses, 1)

	for _, userAddress := range userAddresses {
		require.NotEmpty(t, userAddress)
		require.Equal(t, lastUserAddress.UserID, userAddress.UserID)
		require.Equal(t, lastUserAddress.AddressID, userAddress.AddressID)
		require.Equal(t, lastUserAddress.IsDefault, userAddress.IsDefault)

	}
}
