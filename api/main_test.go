package api

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	firebase "firebase.google.com/go"
	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/util"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
)

func newTestServer(t *testing.T, store db.Store) *Server {
	config := util.Config{
		TokenSymmetricKey:   util.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	opt := option.WithCredentialsFile("serviceAccountKey_test.json")

	fb, err := firebase.NewApp(context.Background(), nil, opt)

	if err != nil {
		log.Fatal("error initializing firebase:", err)
	}

	server, err := NewServer(config, store, fb)
	require.NoError(t, err)

	return server
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
