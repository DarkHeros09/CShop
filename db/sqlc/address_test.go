package db

import (
	"context"
	"database/sql"
	"testing"

	"cshop.com/v2/util"
	"github.com/stretchr/testify/require"
)

func createRandomAddress(t *testing.T) Address {

	arg := CreateAddressParams{
		AddressLine: util.RandomString(5),
		Region:      util.RandomString(5),
		City:        util.RandomString(5),
	}

	address, err := testQueires.CreateAddress(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, address)

	require.Equal(t, arg.AddressLine, address.AddressLine)
	require.Equal(t, arg.Region, address.Region)
	require.Equal(t, arg.City, address.City)

	return address
}

func TestCreateAddress(t *testing.T) {
	createRandomAddress(t)
}

func TestGetAddress(t *testing.T) {
	address1 := createRandomAddress(t)
	address2, err := testQueires.GetAddress(context.Background(), address1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, address2)

	require.Equal(t, address1.ID, address2.ID)
	require.Equal(t, address1.AddressLine, address2.AddressLine)
	require.Equal(t, address1.Region, address2.Region)
	require.Equal(t, address1.City, address2.City)

}

func TestUpdateAddressLine(t *testing.T) {
	address1 := createRandomAddress(t)
	arg := UpdateAddressParams{
		ID: address1.ID,
		AddressLine: sql.NullString{
			String: "New Address Line",
			Valid:  true,
		},
		Region: sql.NullString{
			String: "",
			Valid:  false,
		},
		City: sql.NullString{
			String: "",
			Valid:  false,
		},
	}
	address2, err := testQueires.UpdateAddress(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, address1)

	require.Equal(t, address1.ID, address2.ID)
	require.NotEqual(t, address1.AddressLine, address2.AddressLine)
	require.Equal(t, address1.Region, address2.Region)
	require.Equal(t, address1.City, address2.City)

}

func TestUpdateAddressCity(t *testing.T) {
	address1 := createRandomAddress(t)
	arg := UpdateAddressParams{
		ID: address1.ID,
		AddressLine: sql.NullString{
			String: "",
			Valid:  false,
		},
		Region: sql.NullString{
			String: "",
			Valid:  false,
		},
		City: sql.NullString{
			String: "New Benghazi",
			Valid:  true,
		},
	}

	address2, err := testQueires.UpdateAddress(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, address1)

	require.Equal(t, address1.ID, address2.ID)
	require.Equal(t, address1.AddressLine, address2.AddressLine)
	require.Equal(t, address1.Region, address2.Region)
	require.NotEqual(t, address1.City, address2.City)

}

func TestUpdateAddressLineAndCity(t *testing.T) {
	address1 := createRandomAddress(t)
	arg := UpdateAddressParams{
		ID: address1.ID,
		AddressLine: sql.NullString{
			String: "New Address Line",
			Valid:  true,
		},
		Region: sql.NullString{
			String: "",
			Valid:  false,
		},
		City: sql.NullString{
			String: "New Tubroq",
			Valid:  true,
		},
	}

	address2, err := testQueires.UpdateAddress(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, address1)

	require.Equal(t, address1.ID, address2.ID)
	require.NotEqual(t, address1.AddressLine, address2.AddressLine)
	require.Equal(t, address1.Region, address2.Region)
	require.NotEqual(t, address1.City, address2.City)

}

func TestDeleteAddress(t *testing.T) {
	address1 := createRandomAddress(t)
	err := testQueires.DeleteAddress(context.Background(), address1.ID)

	require.NoError(t, err)

	arg := address1.ID
	useraddress2, err := testQueires.GetAddress(context.Background(), arg)

	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, useraddress2)

}

func TestListAddresses(t *testing.T) {
	var lastAddress Address
	for i := 0; i < 10; i++ {
		lastAddress = createRandomAddress(t)
	}
	arg := ListAddressesByCityParams{
		City:   lastAddress.City,
		Limit:  5,
		Offset: 0,
	}

	addresses, err := testQueires.ListAddressesByCity(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, addresses, 1)

	for _, address := range addresses {
		require.NotEmpty(t, address)
		require.Equal(t, lastAddress.ID, address.ID)
		require.Equal(t, lastAddress.AddressLine, address.AddressLine)
		require.Equal(t, lastAddress.City, address.City)

	}
}
