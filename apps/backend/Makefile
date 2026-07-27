-include .env
export

GOOSE_DRIVER := postgres
GOOSE_DBSTRING := $(DATABASE_URL)
GOOSE_DIR := db/migrations

.PHONY: sqlc migrate-up migrate-down migrate-status test run db-up db-down

sqlc:
	sqlc generate

migrate-up:
	goose -dir $(GOOSE_DIR) $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)" up

migrate-down:
	goose -dir $(GOOSE_DIR) $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)" down

migrate-status:
	goose -dir $(GOOSE_DIR) $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)" status

test:
	go test ./...

run:
	go run ./cmd/app

db-up:
	docker compose -f ../../deployments/docker-compose.yml up -d

db-down:
	docker compose -f ../../deployments/docker-compose.yml down
