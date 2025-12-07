package db

import (
	"context"
	"testing"

	"github.com/cshop/v3/util"
	"github.com/guregu/null/v6"
	"github.com/stretchr/testify/require"
)

func createRandomPaymentType(t *testing.T) PaymentType {
	// arg := util.RandomString(5)
	var paymentType *PaymentType
	var err error
	paymentTypes := []string{"نقداً", "إدفع لي", "سداد", "موبي كاش"}
	for i := 0; i < len(paymentTypes); i++ {
		randomInt := util.RandomInt(0, int64(len(paymentTypes)-1))
		paymentType, err = testStore.CreatePaymentType(context.Background(), paymentTypes[randomInt])
		require.NoError(t, err)
		require.NotEmpty(t, paymentType)

		require.Equal(t, paymentTypes[randomInt], paymentType.Value)

	}
	return *paymentType
}

func adminCreateRandomPaymentType(t *testing.T) PaymentType {
	admin := createRandomAdmin(t)
	var paymentType *PaymentType
	var err error
	paymentTypes := []string{"نقداً", "إدفع لي", "سداد", "موبي كاش"}
	for i := 0; i < len(paymentTypes); i++ {
		randomInt := util.RandomInt(0, int64(len(paymentTypes)-1))
		arg := AdminCreatePaymentTypeParams{
			AdminID:  admin.ID,
			Value:    paymentTypes[randomInt],
			IsActive: true,
		}
		paymentType, err = testStore.AdminCreatePaymentType(context.Background(), arg)
		require.NoError(t, err)
		require.NotEmpty(t, paymentType)

		require.Equal(t, paymentTypes[randomInt], paymentType.Value)

	}
	return *paymentType
}

func createRandomPaymentTypeForUpdateOrDelete(t *testing.T) PaymentType {
	arg := util.RandomString(5)
	paymentType, err := testStore.CreatePaymentType(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, paymentType)

	require.Equal(t, arg, paymentType.Value)

	return *paymentType
}
func TestCreatePaymentType(t *testing.T) {
	createRandomPaymentType(t)
}
func TestAdminCreatePaymentType(t *testing.T) {
	adminCreateRandomPaymentType(t)
}

func TestGetPaymentType(t *testing.T) {
	paymentType1 := createRandomPaymentType(t)
	paymentType2, err := testStore.GetPaymentType(context.Background(), paymentType1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, paymentType2)

	require.Equal(t, paymentType1.ID, paymentType2.ID)
	require.Equal(t, paymentType1.Value, paymentType2.Value)
}

func TestUpdatePaymentType(t *testing.T) {
	paymentType1 := createRandomPaymentTypeForUpdateOrDelete(t)
	arg := UpdatePaymentTypeParams{IsActive: null.BoolFrom(false),
		//last value from paymentType1 otherwise will get duplicate keys becuase of unique constraint
		Value: null.StringFrom(paymentType1.Value),
		ID:    paymentType1.ID,
	}

	paymentType2, err := testStore.UpdatePaymentType(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, paymentType2)

	require.Equal(t, paymentType1.ID, paymentType2.ID)
	require.Equal(t, paymentType1.Value, paymentType2.Value)
	require.NotEqual(t, paymentType1.IsActive, paymentType2.IsActive)
}

func TestDisablePaymentType(t *testing.T) {
	paymentType1 := createRandomPaymentTypeForUpdateOrDelete(t)

	arg := UpdatePaymentTypeParams{
		Value: null.StringFrom(paymentType1.Value),
		ID:    paymentType1.ID, IsActive: null.BoolFrom(false),
	}
	updatedpaymentType, err := testStore.UpdatePaymentType(context.Background(), arg)

	require.NotEmpty(t, updatedpaymentType)
	require.NoError(t, err)

	require.False(t, updatedpaymentType.IsActive)

}

func TestAdminUpdatePaymentType(t *testing.T) {
	admin := createRandomAdmin(t)
	paymentType1 := createRandomPaymentTypeForUpdateOrDelete(t)
	arg := AdminUpdatePaymentTypeParams{IsActive: null.BoolFrom(false),
		AdminID: admin.ID,
		Value:   null.StringFrom(paymentType1.Value),
		ID:      paymentType1.ID,
	}

	paymentType2, err := testStore.AdminUpdatePaymentType(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, paymentType2)

	require.Equal(t, paymentType1.ID, paymentType2.ID)
	require.Equal(t, paymentType1.Value, paymentType2.Value)
	require.NotEqual(t, paymentType1.IsActive, paymentType2.IsActive)
}

// func TestAdminDeletePaymentType(t *testing.T) {
// 	admin := createRandomAdmin(t)
// 	paymentType1 := createRandomPaymentType(t)

// 	arg := AdminDeletePaymentTypeParams{
// 		AdminID: admin.ID,
// 		ID:      paymentType1.ID,
// 	}
// 	err := testStore.AdminDeletePaymentType(context.Background(), arg)

// 	require.NoError(t, err)

// 	paymentType2, err := testStore.GetPaymentType(context.Background(), paymentType1.ID)

// 	require.Error(t, err)
// 	require.EqualError(t, err, pgx.ErrNoRows.Error())
// 	require.Empty(t, paymentType2)

// }

func TestListPaymentTypes(t *testing.T) {
	paymentTypes1 := []string{"نقداً", "إدفع لي", "سداد", "موبي كاش"}
	for i := 0; i < len(paymentTypes1); i++ {
		createRandomPaymentType(t)
	}
	// arg := ListPaymentTypesParams{
	// 	Limit:  int32(len(paymentTypes1)),
	// 	Offset: int32(len(paymentTypes1)),
	// }

	PaymentTypes, err := testStore.ListPaymentTypes(context.Background() /*arg*/)

	require.NoError(t, err)
	require.NotEmpty(t, PaymentTypes)

	for _, PaymentType := range PaymentTypes {
		require.NotEmpty(t, PaymentType)
	}

}

func TestAdminListPaymentTypes(t *testing.T) {
	admin := createRandomAdmin(t)
	paymentTypes1 := []string{"نقداً", "إدفع لي", "سداد", "موبي كاش"}
	for i := 0; i < len(paymentTypes1); i++ {
		adminCreateRandomPaymentType(t)
	}
	// arg := ListPaymentTypesParams{
	// 	Limit:  int32(len(paymentTypes1)),
	// 	Offset: int32(len(paymentTypes1)),
	// }

	paymentTypes, err := testStore.AdminListPaymentTypes(context.Background(), admin.ID)

	require.NoError(t, err)
	require.NotEmpty(t, paymentTypes)

	for _, paymentType := range paymentTypes {
		require.NotEmpty(t, paymentType)
	}

}
