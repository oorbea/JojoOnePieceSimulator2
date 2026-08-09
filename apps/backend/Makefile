# Local dev is Docker-only (see deployments/.env.example for the full env var
# set) - migrate-*/db-* exec into the running backend container rather than
# needing DATABASE_URL/goose on the host. test/test-vips still run on the
# host: the runtime image is a minimal multi-stage build with only the
# compiled binary, no Go toolchain or source, so tests can't run inside it.
COMPOSE := docker compose -f ../../deployments/docker-compose.yml -f ../../deployments/docker-compose.dev.yml
COMPOSE_TEST := docker compose -f ../../deployments/docker-compose.yml -f ../../deployments/docker-compose.test.yml

.PHONY: sqlc swagger migrate-up migrate-down migrate-status test test-vips db-up db-down test-docker test-vips-docker test-integration-docker

sqlc:
	sqlc generate

swagger:
	go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/app/main.go -o docs --parseInternal

# `up` is already applied automatically on backend startup (see
# internal/infrastructure/postgres/migrate.go) - kept here for parity/CI.
migrate-up:
	$(COMPOSE) exec backend sh -c 'goose -dir /migrations postgres "$$DATABASE_URL" up'

migrate-down:
	$(COMPOSE) exec backend sh -c 'goose -dir /migrations postgres "$$DATABASE_URL" down'

migrate-status:
	$(COMPOSE) exec backend sh -c 'goose -dir /migrations postgres "$$DATABASE_URL" status'

test:
	go test ./...

# Compiles and exercises the real libvips-backed image processor. Requires a
# system libvips (vips-dev); on Windows, run it inside the backend container
# instead (`docker build -f ../../deployments/docker/Dockerfile.backend .`
# from the repo root, or run this target in WSL2).
test-vips:
	go test -tags vips ./...

db-up:
	$(COMPOSE) up -d

db-down:
	$(COMPOSE) down

# Docker-based test runners - see deployments/docker-compose.test.yml for why
# these exist (some host machines' Windows Application Control policy
# randomly blocks host-run `go test` binaries). Prefer these over the
# host-based test/test-vips targets when in doubt.
test-docker:
	$(COMPOSE_TEST) run --rm backend-test go test ./...

test-vips-docker:
	$(COMPOSE_TEST) run --rm -e CGO_ENABLED=1 backend-test sh -c "apk add --no-cache build-base pkgconfig vips-dev >/dev/null && go test -tags vips ./..."

# Requires `make db-up` first - integration tests dial real Postgres/Redis by
# service name on the compose network.
test-integration-docker:
	$(COMPOSE_TEST) run --rm backend-test go test -tags integration ./...
