DB_HOST ?= localhost
DB_PORT ?= 6666
DB_VERSION ?= 16.0

DB_URL=postgresql://postgres:secret@$(DB_HOST):$(DB_PORT)/cshop?sslmode=disable

postgres:
	docker run --name psql_$(DB_VERSION)-cshop -p $(DB_PORT):5432 -e POSTGRES_USER=postgres \
	-e POSTGRES_PASSWORD=secret -d -e "TZ=Africa/Tripoli" -e "PGTZ=Africa/Tripoli" postgres:alpine

create_db:
	docker exec -it psql_$(DB_VERSION)-cshop createdb --username=postgres --owner=postgres cshop

drop_db:
	docker exec -it psql_$(DB_VERSION)-cshop dropdb cshop --username=postgres

init_migrate:
	migrate create -ext sql -dir db/migration -seq init_schema

new_migrate:
	migrate create -ext sql -dir db/migration -seq $(name)

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

migrate_up_docker:
	docker run --rm -v "${CURDIR}/db/migration":/migrations --network host \
	migrate/migrate -path /migrations -database "$(DB_URL)" -verbose up

migrate_up1_docker:
	docker run --rm -v "${CURDIR}/db/migration":/migrations --network host \
	migrate/migrate -path /migrations -database "$(DB_URL)" -verbose up 1

migrate_down_docker:
	docker run --rm -v "${CURDIR}/db/migration":/migrations --network host \
	migrate/migrate -path /migrations -database "$(DB_URL)" -verbose down -all

migrate_down1_docker:
	docker run --rm -v "${CURDIR}/db/migration":/migrations --network host \
	migrate/migrate -path /migrations -database "$(DB_URL)" -verbose down 1

migrate_fix_docker:
	docker run --rm -v "${CURDIR}/db/migration":/migrations --network host \
	migrate/migrate -path /migrations -database "$(DB_URL)" -verbose force 1

sqlc:
	sqlc generate

sqlcversion:
	docker run --rm -v "${CURDIR}":/src -w /src sqlc/sqlc version

sqlcwin:
	docker run --rm -v ${pwd}:/src -w /src sqlc/sqlc generate

sqlcfix:
	docker run --rm -v "${CURDIR}":/src -w /src sqlc/sqlc generate

init_docker:
	@pwsh -noprofile -command 'if ([bool]([System.Environment]::OSVersion))\
	{\
		if (![bool](docker ps 2>NUL))\
		{\
			Start-Process "C:\Program Files\Docker\Docker\Docker Desktop.exe" -WindowStyle Hidden\
		}\
		while (![bool](docker ps 2>NUL)) {}\
	}'

	@docker start redis
	@docker start psql_$(DB_VERSION)-cshop

test:
	go test -v -cover -timeout 1m -shuffle on -count=1 ./...

testwin:
	powershell -command "go test -v -cover -timeout 1m -shuffle on -count=1 ./...  | tee test_output.txt"
	powershell -command "Select-String -Path test_output.txt -Pattern 'FAIL'"
	powershell -command "del test_output.txt"

docker_login:
	powershell -command "$$DOCKER_ACCESS_TOKEN | docker login -u mohammednajib --password-stdin"

dagger_test:
	docker start redis
	docker start psql_$(DB_VERSION)-cshop
	docker login
	go run ./dagger/dagger_test_workflow.go

dagger_test2:
	go run ./dagger2/dagger_test_workflow.go

server:
	go run main.go

stop:
	@-docker stop $(shell docker ps -q >nul 2>nul) >nul 2>nul
	@echo "stopped all containers"
	@make -s close_docker
	@echo "closed docker"

close_docker:
	@-powershell -command 'Get-Process | Where-Object { $$_.Name -like "*Docker Desktop*" } | Stop-Process'
	@-powershell -command 'Get-Process | Where-Object { $$_.Name -like "*docker*" } | Stop-Process'

mock:
	mockgen --build_flags=--mod=mod -package mockdb -destination db/mock/store.go github.com/cshop/v3/db/sqlc Store
	mockgen --build_flags=--mod=mod -package mockwk -destination worker/mock/distributor.go github.com/cshop/v3/worker TaskDistributor

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

redis:
	docker run --name redis -p 6379:6379 -d redis:7-alpine

db_docs:
	dbdocs build .\doc\db.dbml

db_schema:
	dbml2sql doc/db.dbml --postgres -o doc/schema.sql

unocss:
	unocss "./web/views/*.html" -c "./web/uno.config.ts" -o "./web/styles/output.css" -m --watch

.PHONY: postgres create_db drop_db init_migrate new_migrate migrate_up migrate_down \
		migrate_up1 migrate_down1 sqlc sqlcwin sqlcfix triggers_up triggers_down \
		mock server proto protofix evans db_docs db_schema dagger_test \
		redis unocss init_docker stop close_docker