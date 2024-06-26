package db

import (
	"context"
	"testing"

	"github.com/cshop/v3/util"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func createRandomPaymentMethod(t *testing.T) PaymentMethod {
	user1 := createRandomUser(t)
	paymentType := createRandomPaymentType(t)
	arg := CreatePaymentMethodParams{
		UserID:        user1.ID,
		PaymentTypeID: paymentType.ID,
		Provider:      util.RandomString(5),
	}
	paymentMethod, err := testStore.CreatePaymentMethod(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, paymentMethod)

	require.Equal(t, arg.UserID, paymentMethod.UserID)
	require.Equal(t, arg.PaymentTypeID, paymentMethod.PaymentTypeID)
	require.Equal(t, arg.Provider, paymentMethod.Provider)

	return paymentMethod
}

func TestCreatePaymentMethod(t *testing.T) {
	go createRandomPaymentMethod(t)
}

func TestGetPaymentMethod(t *testing.T) {
	paymentMethod1 := createRandomPaymentMethod(t)

	arg := GetPaymentMethodParams{
		// ID:            paymentMethod1.ID,
		UserID:        paymentMethod1.UserID,
		PaymentTypeID: paymentMethod1.PaymentTypeID,
	}

	paymentMethod2, err := testStore.GetPaymentMethod(context.Background(), arg)
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

	paymentMethod2, err := testStore.UpdatePaymentMethod(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, paymentMethod2)

	require.Equal(t, paymentMethod1.ID, paymentMethod2.ID)
	require.Equal(t, paymentMethod1.UserID, paymentMethod2.UserID)
	require.Equal(t, paymentMethod1.PaymentTypeID, paymentMethod2.PaymentTypeID)
	require.NotEqual(t, paymentMethod1.Provider, paymentMethod2.Provider)
}

func TestDeletePaymentMethod(t *testing.T) {
	paymentMethod1 := createRandomPaymentMethod(t)

	arg := DeletePaymentMethodParams{
		ID:     paymentMethod1.ID,
		UserID: paymentMethod1.UserID,
	}

	paymentMethod2, err := testStore.DeletePaymentMethod(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, paymentMethod2)

	paymentMethod3, err := testStore.DeletePaymentMethod(context.Background(), arg)

	require.Error(t, err)
	require.Empty(t, paymentMethod3)
	require.EqualError(t, err, pgx.ErrNoRows.Error())

}

func TestListPaymentMethods(t *testing.T) {
	user := createRandomUser(t)
	paymentTypes := make(map[int]PaymentType)

	for i := 0; i < 10; i++ {
		paymentType := createRandomPaymentType(t)
		paymentTypes[int(paymentType.ID)] = paymentType
		arg := CreatePaymentMethodParams{
			UserID:        user.ID,
			PaymentTypeID: paymentType.ID,
			Provider:      util.RandomString(5),
		}
		testStore.CreatePaymentMethod(context.Background(), arg)
	}
	arg := ListPaymentMethodsParams{
		Limit:  5,
		Offset: 0,
		UserID: user.ID,
	}

	PaymentMethods, err := testStore.ListPaymentMethods(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, PaymentMethods, len(paymentTypes))

	for _, PaymentMethod := range PaymentMethods {
		require.NotEmpty(t, PaymentMethod)

	}
}
