package db

import (
	"context"
	"testing"

	"github.com/cshop/v3/util"
	"github.com/guregu/null"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/require"
)

func createRandomPaymentType(t *testing.T) PaymentType {
	arg := util.RandomString(5)
	paymentType, err := testQueires.CreatePaymentType(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, paymentType)

	require.Equal(t, arg, paymentType.Value)

	return paymentType
}
func TestCreatePaymentType(t *testing.T) {
	createRandomPaymentType(t)
}

func TestGetPaymentType(t *testing.T) {
	paymentType1 := createRandomPaymentType(t)
	paymentType2, err := testQueires.GetPaymentType(context.Background(), paymentType1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, paymentType2)

	require.Equal(t, paymentType1.ID, paymentType2.ID)
	require.Equal(t, paymentType1.Value, paymentType2.Value)
}

func TestUpdatePaymentTypeNameAndCategoryID(t *testing.T) {
	paymentType1 := createRandomPaymentType(t)
	arg := UpdatePaymentTypeParams{
		Value: null.StringFrom(util.RandomString(5)),
		ID:    paymentType1.ID,
	}

	paymentType2, err := testQueires.UpdatePaymentType(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, paymentType2)

	require.Equal(t, paymentType1.ID, paymentType2.ID)
	require.NotEqual(t, paymentType1.Value, paymentType2.Value)
}

func TestDeletePaymentType(t *testing.T) {
	paymentType1 := createRandomPaymentType(t)
	err := testQueires.DeletePaymentType(context.Background(), paymentType1.ID)

	require.NoError(t, err)

	paymentType2, err := testQueires.GetPaymentType(context.Background(), paymentType1.ID)

	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, paymentType2)

}

func TestListPaymentTypes(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomPaymentType(t)
	}
	arg := ListPaymentTypesParams{
		Limit:  5,
		Offset: 5,
	}

	PaymentTypes, err := testQueires.ListPaymentTypes(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, PaymentTypes)

	for _, PaymentType := range PaymentTypes {
		require.NotEmpty(t, PaymentType)
	}

}
