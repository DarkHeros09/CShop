package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/cshop/v3/util"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func createRandomAddress(user User, t *testing.T) Address {
	// defer goleak.VerifyNone(t)
	arg := CreateAddressParams{
		Name:        util.RandomString(5),
		UserID:      user.ID,
		Telephone:   fmt.Sprintf("%d", util.RandomInt(910000000, 929999999)),
		AddressLine: util.RandomString(5),
		Region:      util.RandomString(5),
		City:        util.RandomString(5),
	}

	address, err := testStore.CreateAddress(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, address)

	require.Equal(t, arg.Name, address.Name)
	require.Equal(t, arg.AddressLine, address.AddressLine)
	require.Equal(t, arg.Region, address.Region)
	require.Equal(t, arg.City, address.City)

	return address
}

func TestCreateAddress(t *testing.T) {
	user := createRandomUser(t)
	createRandomAddress(user, t)
}

func createRandomAddressWithUser(t *testing.T) Address {
	// defer goleak.VerifyNone(t)
	user := createRandomUser(t)
	arg := CreateAddressParams{
		Name:        util.RandomString(5),
		UserID:      user.ID,
		Telephone:   fmt.Sprintf("%d", util.RandomInt(910000000, 929999999)),
		AddressLine: util.RandomString(5),
		Region:      util.RandomString(5),
		City:        util.RandomString(5),
	}

	address, err := testStore.CreateAddress(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, address)

	require.Equal(t, arg.Name, address.Name)
	require.Equal(t, arg.AddressLine, address.AddressLine)
	require.Equal(t, arg.Region, address.Region)
	require.Equal(t, arg.City, address.City)

	return address
}

func TestCreateAddressWithUser(t *testing.T) {
	createRandomAddressWithUser(t)
}

func TestGetAddress(t *testing.T) {
	address1 := createRandomAddressWithUser(t)
	address2, err := testStore.GetAddress(context.Background(), address1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, address2)

	require.Equal(t, address1.ID, address2.ID)
	require.Equal(t, address1.AddressLine, address2.AddressLine)
	require.Equal(t, address1.Region, address2.Region)
	require.Equal(t, address1.City, address2.City)

}

func TestUpdateAddressLine(t *testing.T) {
	address1 := createRandomAddressWithUser(t)
	arg := UpdateAddressParams{
		ID:          address1.ID,
		UserID:      address1.UserID,
		AddressLine: null.StringFrom("New Address Line"),
	}
	address2, err := testStore.UpdateAddress(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, address1)

	require.Equal(t, address1.ID, address2.ID)
	require.NotEqual(t, address1.AddressLine, address2.AddressLine)
	require.Equal(t, address1.Region, address2.Region)
	require.Equal(t, address1.City, address2.City)

}

func TestUpdateAddressCity(t *testing.T) {
	address1 := createRandomAddressWithUser(t)
	arg := UpdateAddressParams{
		ID:     address1.ID,
		UserID: address1.UserID,
		Name:   null.StringFrom("New Name"),
		City:   null.StringFrom("New Benghazi"),
	}

	address2, err := testStore.UpdateAddress(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, address1)

	require.Equal(t, address1.ID, address2.ID)
	require.Equal(t, address1.AddressLine, address2.AddressLine)
	require.Equal(t, address1.Region, address2.Region)
	require.NotEqual(t, address1.City, address2.City)
	require.NotEqual(t, address1.Name, address2.Name)

}

func TestUpdateAddressLineAndCity(t *testing.T) {
	address1 := createRandomAddressWithUser(t)
	arg := UpdateAddressParams{
		ID:          address1.ID,
		UserID:      address1.UserID,
		AddressLine: null.StringFrom("New Address Line"),
		Region:      null.String{},
		City:        null.StringFrom("New Tubroq"),
	}

	address2, err := testStore.UpdateAddress(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, address1)

	require.Equal(t, address1.ID, address2.ID)
	require.NotEqual(t, address1.AddressLine, address2.AddressLine)
	require.Equal(t, address1.Region, address2.Region)
	require.NotEqual(t, address1.City, address2.City)

}

func TestDeleteAddress(t *testing.T) {
	address1 := createRandomAddressWithUser(t)
	err := testStore.DeleteAddress(context.Background(), address1.ID)

	require.NoError(t, err)

	arg := address1.ID
	useraddress2, err := testStore.GetAddress(context.Background(), arg)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, useraddress2)

}

func TestListAddresses(t *testing.T) {
	var lastAddress Address
	for i := 0; i < 10; i++ {
		lastAddress = createRandomAddressWithUser(t)
	}
	arg := ListAddressesByCityParams{
		City:   lastAddress.City,
		Limit:  5,
		Offset: 0,
	}

	addresses, err := testStore.ListAddressesByCity(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, addresses, 1)

	for _, address := range addresses {
		require.NotEmpty(t, address)
		require.Equal(t, lastAddress.ID, address.ID)
		require.Equal(t, lastAddress.AddressLine, address.AddressLine)
		require.Equal(t, lastAddress.City, address.City)

	}
}

func TestListUserAddresses(t *testing.T) {
	user := createRandomUser(t)
	var lastAddress Address
	for range 10 {
		lastAddress = createRandomAddress(user, t)
	}

	addresses, err := testStore.ListAddressesByUserID(context.Background(), user.ID)
	require.NoError(t, err)
	require.NotEmpty(t, addresses)

	// for _, address := range addresses {
	require.NotEmpty(t, addresses[len(addresses)-1])
	require.Equal(t, lastAddress.ID, addresses[len(addresses)-1].ID)
	require.Equal(t, lastAddress.AddressLine, addresses[len(addresses)-1].AddressLine)
	require.Equal(t, lastAddress.City, addresses[len(addresses)-1].City)

	// }

}
