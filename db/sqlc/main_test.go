package db

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/cshop/v3/util"
	"github.com/jackc/pgx/v5/pgxpool"
)

// var testStore.*Queries
// var testDB *pgxpool.Pool

var testStore Store

func TestMain(m *testing.M) {
	// goleak.VerifyTestMain(m)
	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}
	// testDBChan := make(chan *pgxpool.Pool)
	// errChan := make(chan error)
	// go func() {
	// 	testDB, err = pgxpool.Connect(context.Background(), config.DBSource)
	// 	testDBChan <- testDB
	// 	errChan <- err
	// }()

	// testDB := <-testDBChan
	// err = <-errChan
	testDB, err := pgxpool.New(context.Background(), config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db", err)
	}
	// defer testDB.Close()

	// testStore.= New(testDB)
	testStore = NewStore(testDB)

	os.Exit(m.Run())
}
