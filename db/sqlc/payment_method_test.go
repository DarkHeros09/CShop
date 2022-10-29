package db

import (
	"context"
	"testing"

	"github.com/cshop/v3/util"
	"github.com/guregu/null"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/require"
)

func createRandomPaymentMethod(t *testing.T) PaymentMethod {
	user1 := createRandomUser(t)
	paymentType := createRandomPaymentType(t)
	arg := CreatePaymentMethodParams{
		UserID:        user1.ID,
		PaymentTypeID: int32(paymentType.ID),
		Provider:      util.RandomString(5),
	}
	paymentMethod, err := testQueires.CreatePaymentMethod(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, paymentMethod)

	require.Equal(t, arg.UserID, paymentMethod.UserID)
	require.Equal(t, arg.PaymentTypeID, paymentMethod.PaymentTypeID)
	require.Equal(t, arg.Provider, paymentMethod.Provider)

	return paymentMethod
}

func TestCreatePaymentMethod(t *testing.T) {
	createRandomPaymentMethod(t)
}

func TestGetPaymentMethod(t *testing.T) {
	paymentMethod1 := createRandomPaymentMethod(t)

	paymentMethod2, err := testQueires.GetPaymentMethod(context.Background(), paymentMethod1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, paymentMethod2)

	require.Equal(t, paymentMethod1.ID, paymentMethod2.ID)
	require.Equal(t, paymentMethod1.UserID, paymentMethod2.UserID)
	require.Equal(t, paymentMethod1.PaymentTypeID, paymentMethod2.PaymentTypeID)
	require.Equal(t, paymentMethod1.Provider, paymentMethod2.Provider)
	require.Equal(t, paymentMethod1.IsDefault, paymentMethod2.IsDefault)
}

func TestUpdatePaymentMethodIsDefaultAndProvider(t *testing.T) {
	paymentMethod1 := createRandomPaymentMethod(t)
	arg := UpdatePaymentMethodParams{
		UserID:        null.Int{},
		PaymentTypeID: null.Int{},
		Provider:      null.StringFrom(util.RandomString(5)),

		ID: paymentMethod1.ID,
	}

	paymentMethod2, err := testQueires.UpdatePaymentMethod(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, paymentMethod2)

	require.Equal(t, paymentMethod1.ID, paymentMethod2.ID)
	require.Equal(t, paymentMethod1.UserID, paymentMethod2.UserID)
	require.Equal(t, paymentMethod1.PaymentTypeID, paymentMethod2.PaymentTypeID)
	require.NotEqual(t, paymentMethod1.Provider, paymentMethod2.Provider)
}

func TestDeletePaymentMethod(t *testing.T) {
	paymentMethod1 := createRandomPaymentMethod(t)

	err := testQueires.DeletePaymentMethod(context.Background(), paymentMethod1.ID)

	require.NoError(t, err)

	paymentMethod2, err := testQueires.GetPaymentMethod(context.Background(), paymentMethod1.ID)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, paymentMethod2)

}

func TestListPaymentMethods(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomPaymentMethod(t)
	}
	arg := ListPaymentMethodsParams{
		Limit:  5,
		Offset: 0,
	}

	PaymentMethods, err := testQueires.ListPaymentMethods(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, PaymentMethods, 5)

	for _, PaymentMethod := range PaymentMethods {
		require.NotEmpty(t, PaymentMethod)

	}
}
