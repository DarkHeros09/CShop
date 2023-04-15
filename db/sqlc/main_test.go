package db

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/cshop/v3/util"
	"github.com/jackc/pgx/v4/pgxpool"
)

var testQueires *Queries
var testDB *pgxpool.Pool

func TestMain(m *testing.M) {
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
	testDB, err = pgxpool.Connect(context.Background(), config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db", err)
		os.Exit(m.Run())
	}
	// defer testDB.Close()

	testQueires = New(testDB)

}
