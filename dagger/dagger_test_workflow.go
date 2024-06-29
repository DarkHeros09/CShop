package main

import (
	"context"
	"fmt"
	"os"
	"time"

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

	// create a cache volume
	goCache := client.CacheVolume("go-mod")

	// OS
	platform := dagger.Platform("linux/amd64")

	// Database service used for application tests
	database, err := client.Container(dagger.ContainerOpts{Platform: platform}).From("postgres:alpine").
		// WithEnvVariable("BUST", time.Now().String()).
		WithEnvVariable("POSTGRES_USER", "postgres").
		WithEnvVariable("POSTGRES_PASSWORD", "secret").
		WithEnvVariable("POSTGRES_DB", "cshop").
		WithEnvVariable("PGPORT", "6666").
		WithEnvVariable("TZ", "Africa/Tripoli").
		WithEnvVariable("PGTZ", "Africa/Tripoli").
		WithExec([]string{"postgres"}).
		WithExposedPort(6666).
		AsService().
		Start(ctx)

	if err != nil {
		panic(err)
	}

	// Project to test
	src := client.Host().Directory(".").
		WithoutDirectory("./.github").
		WithoutDirectory("./tmp").
		WithoutDirectory("./web").
		WithoutDirectory("./doc").
		WithoutDirectory("./dagger").
		WithoutFile("./.air.toml").
		WithoutFile("./dbml-error.log").
		WithoutFile("./pre-cmd.txt").
		WithoutFile("./post_cmd.txt").
		WithoutFile("./.env").
		WithoutFile("./.env.me").
		WithoutFile("./.env.previous").
		WithoutFile("./.env.vault")

	// multi stage build - stage 1
	// golang-migrate image
	migrate := client.Container(dagger.ContainerOpts{Platform: platform}).
		From("migrate/migrate:latest").
		WithEnvVariable("BUST", time.Now().String()).
		WithDirectory("/src", src).
		WithWorkdir("/src").
		WithEntrypoint([]string{}).
		WithServiceBinding("localhost", database). // bind database with the name db
		WithEnvVariable("DB_HOST", "localhost").   // db refers to the service binding
		WithEnvVariable("DB_PASSWORD", "secret").  // password set in db container
		WithEnvVariable("DB_USER", "postgres").    // default user in postgres image
		WithEnvVariable("DB_NAME", "cshop").       // default db name in postgres image
		WithExec([]string{"which", "migrate"}).
		WithExec([]string{"migrate", "-path", "db/migration", "-database", "postgresql://postgres:secret@localhost:6666/cshop?sslmode=disable", "-verbose", "up"})

	// multi stage build - stage 2
	// golang image with cached dependencies
	container := client.Container(dagger.ContainerOpts{Platform: platform}).
		From("golang:1.22").
		// WithEnvVariable("BUST", time.Now().String()).
		WithEnvVariable("TZ", "Africa/Tripoli").
		WithServiceBinding("localhost", database). // bind database with the name db
		WithEnvVariable("DB_HOST", "localhost").   // db refers to the service binding
		WithEnvVariable("DB_PASSWORD", "secret").  // password set in db container
		WithEnvVariable("DB_USER", "postgres").    // default user in postgres image
		WithEnvVariable("DB_NAME", "cshop").       // default db name in postgres image
		WithDirectory("/src", migrate.Directory("/src")).
		WithWorkdir("/src").
		WithMountedCache("/go/pkg/mod", goCache).
		WithExec([]string{"go", "mod", "download"}) // download go modules

	// multi stage build - stage 3
	// Run Service with tests
	out, err := container.
		WithFile("install.sh", client.HTTP("https://dotenvx.sh/?version=0.45.0")).
		WithExec([]string{"chmod", "+x", "install.sh"}).
		WithExec([]string{"./install.sh"}).
		WithExec([]string{"dotenvx", "run", "-f", ".env.test", "--", "make", "test"}). // execute go test
		Stdout(ctx)

	if err != nil {
		panic(err)
	}
	fmt.Print(out)
}
