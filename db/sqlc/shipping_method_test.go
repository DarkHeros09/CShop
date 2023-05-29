package db

import (
	"context"
	"testing"

	"github.com/cshop/v3/util"
	"github.com/guregu/null"
	"github.com/stretchr/testify/require"
)

func createRandomShippingMethod(t *testing.T) ShippingMethod {
	shippingMethods := []string{"توصيل سريع"}
	var shippingMethod ShippingMethod
	var err error
	for i := 0; i < len(shippingMethods); i++ {
		arg := CreateShippingMethodParams{
			Name:  shippingMethods[i],
			Price: util.RandomDecimalString(1, 100),
		}

		shippingMethod, err = testQueires.CreateShippingMethod(context.Background(), arg)
		require.NoError(t, err)
		require.NotEmpty(t, shippingMethod)

		require.Equal(t, arg.Name, shippingMethod.Name)
		require.Equal(t, arg.Price, shippingMethod.Price)
	}

	return shippingMethod
}
func TestCreateShippingMethod(t *testing.T) {
	createRandomShippingMethod(t)
}

func TestGetShippingMethod(t *testing.T) {
	shippingMethod1 := createRandomShippingMethod(t)
	shippingMethod2, err := testQueires.GetShippingMethod(context.Background(), shippingMethod1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, shippingMethod2)

	require.Equal(t, shippingMethod1.ID, shippingMethod2.ID)
	require.Equal(t, shippingMethod1.Price, shippingMethod2.Price)
	require.Equal(t, shippingMethod1.Name, shippingMethod2.Name)
}

func TestUpdateShippingMethodNameAndPrice(t *testing.T) {
	shippingMethod1 := createRandomShippingMethod(t)
	arg := UpdateShippingMethodParams{
		ID:    shippingMethod1.ID,
		Name:  null.StringFrom(shippingMethod1.Name),
		Price: null.StringFrom(util.RandomDecimalString(1, 100)),
	}

	shippingMethod2, err := testQueires.UpdateShippingMethod(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, shippingMethod2)

	require.Equal(t, shippingMethod1.ID, shippingMethod2.ID)
	require.NotEqual(t, shippingMethod1.Price, shippingMethod2.Price)
	require.Equal(t, shippingMethod1.Name, shippingMethod2.Name)
}

// func TestDeleteShippingMethod(t *testing.T) {
// 	shippingMethod1 := createRandomShippingMethod(t)
// 	err := testQueires.DeleteShippingMethod(context.Background(), shippingMethod1.ID)

// 	require.NoError(t, err)

// 	shippingMethod2, err := testQueires.GetShippingMethod(context.Background(), shippingMethod1.ID)

// 	require.Error(t, err)
// 	require.EqualError(t, err, pgx.ErrNoRows.Error())
// 	require.Empty(t, shippingMethod2)

// 	_ = createRandomShippingMethod(t)

// }

func TestListShippingMethods(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomShippingMethod(t)
	}
	// arg := ListShippingMethodsParams{
	// 	Limit:  5,
	// 	Offset: 0,
	// }

	shippingMethods, err := testQueires.ListShippingMethods(context.Background())

	require.NoError(t, err)
	require.NotEmpty(t, shippingMethods)

	for _, shippingMethod := range shippingMethods {
		require.NotEmpty(t, shippingMethod)
	}

}
