package api

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	firebase "firebase.google.com/go/v4"
	db "github.com/cshop/v3/db/sqlc"
	image "github.com/cshop/v3/image"
	"github.com/cshop/v3/mail"
	"github.com/cshop/v3/util"
	"github.com/cshop/v3/worker"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
)

func newTestServer(
	t *testing.T,
	store db.Store,
	taskDistributor worker.TaskDistributor,
	ik image.ImageKitManagement,
	sender mail.EmailSender,
) *Server {
	config := util.Config{
		UserTokenSymmetricKey:  util.RandomString(32),
		AdminTokenSymmetricKey: util.RandomString(32),
		AccessTokenDuration:    time.Minute,
	}

	opt := option.WithCredentialsFile("serviceAccountKey_test.json")

	fb, err := firebase.NewApp(context.Background(), nil, opt)

	if err != nil {
		log.Fatal("error initializing firebase:", err)
	}

	server, err := NewServer(config, store, fb, taskDistributor, ik, sender)
	require.NoError(t, err)

	return server
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
