package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"cshop.com/v2/util"
	_ "github.com/lib/pq"
)

var testQueires *Queries
var testDB *sql.DB

func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}
	testDB, err = sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db", err)
	}

	testQueires = New(testDB)

	os.Exit(m.Run())
}
