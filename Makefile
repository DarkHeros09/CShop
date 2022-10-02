postgres:
	docker run --name psql_14.5-cshop -p 6666:5432 -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=secret -d postgres:14.5-alpine

createdb:
	docker exec -it psql_14.5-cshop createdb --username=postgres --owner=postgres cshop

dropdb:
	docker exec -it psql_14.5-cshop dropdb cshop --username=postgres

initmigrate:
	migrate create -ext sql -dir db/migration -seq init_schema

triggersup:
	migrate -path db/migration/triggers -database "postgresql://postgres:secret@localhost:6666/cshop?sslmode=disable" -verbose up

triggersdown:
	migrate -path db/migration/triggers -database "postgresql://postgres:secret@localhost:6666/cshop?sslmode=disable" -verbose up

migrateup:
	migrate -path db/migration -database "postgresql://postgres:secret@localhost:6666/cshop?sslmode=disable" -verbose up

migrateup1:
	migrate -path db/migration -database "postgresql://postgres:secret@localhost:6666/cshop?sslmode=disable" -verbose up 1

migratedown:
	migrate -path db/migration -database "postgresql://postgres:secret@localhost:6666/cshop?sslmode=disable" -verbose down

migratedown1:
	migrate -path db/migration -database "postgresql://postgres:secret@localhost:6666/cshop?sslmode=disable" -verbose down 1

sqlc:
	sqlc generate

sqlcversion:
	docker run --rm -v ${CURDIR}:/src -w /src kjconroy/sqlc version

sqlcwin:
	docker run --rm -v ${pwd}:/src -w /src kjconroy/sqlc generate

sqlcfix:
	docker run --rm -v ${CURDIR}:/src -w /src kjconroy/sqlc generate

test:
	go test -v -cover ./...

.PHONY: postgres createdb dropdb initmigrate migrateup migratedown sqlc sqlcwin sqlcfix triggersup triggersdown