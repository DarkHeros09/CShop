package main

import (
	"context"
	"fmt"
	"os"

	"dagger.io/dagger"
)

func main() {
	ctx := context.Background()

	// create a Dagger client
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		panic(err)
	}
	defer client.Close()

	//OS
	platform := dagger.Platform("linux/amd64")

	// Database service used for application tests
	database := client.Container(dagger.ContainerOpts{Platform: platform}).From("postgres:15.4-alpine").
		// WithEnvVariable("BUST", time.Now().String()).
		WithEnvVariable("POSTGRES_USER", "postgres").
		WithEnvVariable("POSTGRES_PASSWORD", "secret").
		WithEnvVariable("POSTGRES_DB", "cshop").
		WithEnvVariable("PGPORT", "6666").
		WithEnvVariable("TZ", "Africa/Tripoli").
		WithEnvVariable("PGTZ", "Africa/Tripoli").
		WithExec([]string{"postgres"})
		// WithExec(nil).
		// WithExposedPort(5432)

	// Project to test
	src := client.Host().Directory(".")

	// client.Container(dagger.ContainerOpts{Platform: platform}).
	// 	From("migrate/migrate:latest").
	// 	WithDirectory("/src", src).
	// 	WithWorkdir("/src").
	// 	WithEntrypoint([]string{}).
	// 	WithExec([]string{"which", "migrate"}).
	// 	WithExec([]string{"migrate", "-path", "db/migration", "-database", "postgresql://postgres:secret@localhost:6666/cshop?sslmode=disable", "-verbose", "up"}).Stdout(ctx)

	// Run Service with tests
	container := client.Container(dagger.ContainerOpts{Platform: platform}).From("golang:1.21").
		// WithEnvVariable("BUST", time.Now().String()).
		WithEnvVariable("TZ", "Africa/Tripoli").
		WithServiceBinding("localhost", database). // bind database with the name db
		WithEnvVariable("DB_HOST", "localhost").   // db refers to the service binding
		WithEnvVariable("DB_PASSWORD", "secret").  // password set in db container
		WithEnvVariable("DB_USER", "postgres").    // default user in postgres image
		WithEnvVariable("DB_NAME", "cshop")        // default db name in postgres image

	build := container.
		WithDirectory("/src", src).
		WithWorkdir("/src").
		WithFile("migrate.tar.gz", client.HTTP("https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz")).
		WithExec([]string{"tar", "fxvz", "migrate.tar.gz", "-C", "/usr/bin/"}). // move to executables
		WithExec([]string{"rm", "-rf", "migrate.tar.gz"}).                      // delete unused files
		WithExec([]string{"which", "migrate"})                                  // test go migrate

	out, err := build.
		WithExec([]string{"make", "migrate_up"}). // run migrations
		WithExec([]string{"make", "test"}).       // execute go test
		Stdout(ctx)

	if err != nil {
		panic(err)
	}
	fmt.Print(out)
}
