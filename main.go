package main

import (
	"context"
	"log"

	firebase "firebase.google.com/go"
	"github.com/cshop/v3/api"
	db "github.com/cshop/v3/db/sqlc"
	"github.com/cshop/v3/util"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/api/option"
)

func main() {
	config, err := util.LoadConfig(".") // we use . because app.env is on the same level with main.go
	if err != nil {
		log.Fatal("cannot load config:", err)
	}
	conn, err := pgxpool.New(context.Background(), config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db", err)
	}

	opt := option.WithCredentialsFile("serviceAccountKey.json")

	fb, err := firebase.NewApp(context.Background(), nil, opt)

	if err != nil {
		log.Fatal("error initializing firebase:", err)
	}

	store := db.NewStore(conn)
	runFiberServer(config, store, fb)

}

func runFiberServer(config util.Config, store db.Store, fb *firebase.App) {
	server, err := api.NewServer(config, store, fb)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
