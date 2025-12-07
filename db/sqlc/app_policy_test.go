package db

import (
	"context"
	"testing"

	"github.com/guregu/null/v6"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func createRandomAppPolicy(t *testing.T) AppPolicy {
	t.Helper()
	admin := createRandomAdmin(t)
	data := `
## المقدمة
نحن نلتزم بحماية خصوصيتك. تهدف هذه السياسة إلى شرح كيفية جمع البيانات واستخدامها ومشاركتها عند استخدام تطبيقنا.

## جمع المعلومات
نقوم بجمع أنواع مختلفة من المعلومات لضمان تحسين تجربة المستخدم، مثل:

- **المعلومات الشخصية**: مثل الاسم والبريد الإلكتروني.
- **المعلومات التقنية**: مثل نوع الجهاز ونظام التشغيل.

## استخدام المعلومات
المعلومات التي نجمعها تُستخدم للأغراض التالية:

- لتحسين التطبيق وتقديم الدعم الفني.
- لإرسال إشعارات أو تحديثات متعلقة بالخدمات.

## مشاركة المعلومات
نحن لا نبيع أو نشارك معلوماتك الشخصية مع أطراف ثالثة، إلا في الحالات الضرورية لتحسين خدماتنا أو الامتثال للقوانين.

## التغييرات على السياسة
قد نقوم بتحديث هذه السياسة من وقت لآخر. سيتم إعلامك بأي تغييرات رئيسية من خلال التطبيق.
`
	arg := CreateAppPolicyParams{
		AdminID: admin.ID,
		Policy:  null.StringFrom(data),
	}

	appPolicy, err := testStore.CreateAppPolicy(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, appPolicy)

	require.Equal(t, arg.Policy.String, appPolicy.Policy.String)

	require.NotEmpty(t, appPolicy.CreatedAt)

	return *appPolicy
}

func TestCreateAppPolicy(t *testing.T) {

	createRandomAppPolicy(t)
}

func TestGetAppPolicy(t *testing.T) {

	createRandomAppPolicy(t)

	appPolicy2, err := testStore.GetAppPolicy(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, appPolicy2)

	require.NotEmpty(t, appPolicy2.Policy.String)
	require.NotEmpty(t, appPolicy2.CreatedAt)
	require.NotEmpty(t, appPolicy2.UpdatedAt)

}

func TestUpdateAppPolicy(t *testing.T) {
	admin := createRandomAdmin(t)
	appPolicy1 := createRandomAppPolicy(t)

	arg := UpdateAppPolicyParams{
		ID:      appPolicy1.ID,
		Policy:  appPolicy1.Policy,
		AdminID: admin.ID,
	}

	appPolicy2, err := testStore.UpdateAppPolicy(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, appPolicy2)

	require.Equal(t, appPolicy1.Policy.String, appPolicy2.Policy.String)
	require.Equal(t, appPolicy1.CreatedAt, appPolicy2.CreatedAt)
	require.NotEqual(t, appPolicy1.UpdatedAt, appPolicy2.UpdatedAt)
}

func TestDeleteAppPolicy(t *testing.T) {
	admin := createRandomAdmin(t)
	appPolicy1 := createRandomAppPolicy(t)

	arg := DeleteAppPolicyParams{
		ID:      appPolicy1.ID,
		AdminID: admin.ID,
	}

	appPolicy2, err := testStore.DeleteAppPolicy(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, appPolicy2)

	appPolicy3, err := testStore.DeleteAppPolicy(context.Background(), arg)

	require.Error(t, err)
	require.Empty(t, appPolicy3)
	require.EqualError(t, err, pgx.ErrNoRows.Error())

}
