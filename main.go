package main

import (
	"context"
	"log"

	"github.com/cshop/v3/api"
	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/util"
	"github.com/jackc/pgx/v4"
)

func main() {
	config, err := util.LoadConfig(".") // we use . because app.env is on the same level with main.go
	if err != nil {
		log.Fatal("cannot load config:", err)
	}
	conn, err := pgx.Connect(context.Background(), config.DBSource)
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
