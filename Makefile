DB_HOST ?= localhost
DB_PORT ?= 6666

DB_URL=postgresql://postgres:secret@$(DB_HOST):$(DB_PORT)/cshop?sslmode=disable

postgres:
	docker run --name psql_14.5-cshop -p 6666:5432 -e POSTGRES_USER=postgres \
	-e POSTGRES_PASSWORD=secret -d postgres:14.5-alpine

create_db:
	docker exec -it psql_14.5-cshop createdb --username=postgres --owner=postgres cshop

drop_db:
	docker exec -it psql_14.5-cshop dropdb cshop --username=postgres

init_migrate:
	migrate create -ext sql -dir db/migration -seq init_schema

triggers_up:
	migrate -path db/migration/triggers -database \
	"$(DB_URL)" -verbose up

triggers_down:
	migrate -path db/migration/triggers -database \
	"$(DB_URL)" -verbose up

migrate_up:
	migrate -path db/migration -database \
	"$(DB_URL)" -verbose up

migrate_up1:
	migrate -path db/migration -database \
	"$(DB_URL)" -verbose up 1

migrate_down:
	migrate -path db/migration -database \
	"$(DB_URL)" -verbose down

migrate_down1:
	migrate -path db/migration -database \
	"$(DB_URL)" -verbose down 1

migrate_fix:
	migrate -path db/migration -database \
	"$(DB_URL)" -verbose force 1

sqlc:
	sqlc generate

sqlcversion:
	docker run --rm -v ${CURDIR}:/src -w /src kjconroy/sqlc version

sqlcwin:
	docker run --rm -v ${pwd}:/src -w /src kjconroy/sqlc generate

sqlcfix:
	docker run --rm -v ${CURDIR}:/src -w /src kjconroy/sqlc generate

test:
	go test -v -cover -timeout 1m -shuffle on -count=1 ./...

dagger_test:
	go run ./dagger/dagger_test_workflow.go

server:
	go run main.go

mock:
	mockgen --build_flags=--mod=mod -package mockdb -destination db/mock/store.go github.com/cshop/v3/db/sqlc Store

proto:
	rm -f pb/*.go
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
    --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=pb --grpc-gateway_opt=paths=source_relative \
    proto/*.proto

protofix:
	del pb\\*.go
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
    --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=pb --grpc-gateway_opt=paths=source_relative \
    proto/*.proto

evans:
	evans --host localhost --port 9090 -r repl

db_docs:
	dbdocs build .\doc\db.dbml

db_schema:
	dbml2sql --postgres -o doc/schema.sql doc/db.dbml


.PHONY: postgres create_db drop_db init_migrate migrate_up migrate_down \
		migrate_up1 migrate_down1 sqlc sqlcwin sqlcfix triggers_up triggers_down \
		mock server proto protofix evans db_docs db_schema dagger_test