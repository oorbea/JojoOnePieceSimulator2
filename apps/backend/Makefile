-include .env
export

GOOSE_DRIVER := postgres
GOOSE_DBSTRING := $(DATABASE_URL)
GOOSE_DIR := db/migrations

.PHONY: sqlc swagger migrate-up migrate-down migrate-status test test-vips run db-up db-down

sqlc:
	sqlc generate

swagger:
	go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/app/main.go -o docs --parseInternal

migrate-up:
	goose -dir $(GOOSE_DIR) $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)" up

migrate-down:
	goose -dir $(GOOSE_DIR) $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)" down

migrate-status:
	goose -dir $(GOOSE_DIR) $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)" status

test:
	go test ./...

# Compiles and exercises the real libvips-backed image processor. Requires a
# system libvips (vips-dev); on Windows, run it inside the backend container
# instead (`docker build -f ../../deployments/docker/Dockerfile.backend .`
# from the repo root, or run this target in WSL2).
test-vips:
	go test -tags vips ./...

run:
	go run ./cmd/app

db-up:
	docker compose -f ../../deployments/docker-compose.yml up -d

db-down:
	docker compose -f ../../deployments/docker-compose.yml down
