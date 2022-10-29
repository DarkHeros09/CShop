package db

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/cshop/v3/util"
	"github.com/jackc/pgx/v4"
)

var testQueires *Queries
var testDB *pgx.Conn

func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}
	// testDB, err = sql.Open(config.DBDriver, config.DBSource)
	testDB, err = pgx.Connect(context.Background(), config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db", err)
	}
	defer testDB.Close(context.Background())

	testQueires = New(testDB)

	os.Exit(m.Run())
}
