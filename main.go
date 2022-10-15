package main

import (
	"database/sql"
	"log"

	"github.com/cshop/v3/api"
	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/util"
	_ "github.com/lib/pq"
)

func main() {
	config, err := util.LoadConfig(".") // we use . because app.env is on the same level with main.go
	if err != nil {
		log.Fatal("cannot load config:", err)
	}
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db", err)
	}

	store := db.NewStore(conn)
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
