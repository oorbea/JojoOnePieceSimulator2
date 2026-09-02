# Local dev is Docker-only (see deployments/.env.example for the full env var
# set) - migrate-*/db-* exec into the running backend container rather than
# needing DATABASE_URL/goose on the host. test/test-vips still run on the
# host: the runtime image is a minimal multi-stage build with only the
# compiled binary, no Go toolchain or source, so tests can't run inside it.
COMPOSE := docker compose -f ../../deployments/docker-compose.yml -f ../../deployments/docker-compose.dev.yml
COMPOSE_TEST := docker compose -f ../../deployments/docker-compose.yml -f ../../deployments/docker-compose.test.yml

.PHONY: sqlc swagger migrate-up migrate-down migrate-status test test-vips db-up db-down test-docker test-vips-docker test-integration-docker types types-check types-docker types-check-docker

sqlc:
	sqlc generate

swagger:
	go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/app/main.go -o docs --parseInternal

# cmd/typegen is the single producer of apps/frontend/src/shared/contracts/ -
# see ObsidianVault/contratos-tipos-generados.md. `types` runs on the host
# (the App Control issue documented above is specific to freshly-built
# `go test` binaries; cmd/typegen imports no networking and isn't affected).
types:
	go run ./cmd/typegen -out ../frontend/src/shared/contracts

# Fails, naming the stale file, instead of silently leaving a stale
# contracts/ on disk. This is what CI's `contracts` job runs.
types-check:
	go run ./cmd/typegen -out ../frontend/src/shared/contracts -check

# Docker fallback for `types`, using the same App-Control workaround as
# test-docker - see docker-compose.test.yml's `typegen` service.
types-docker:
	$(COMPOSE_TEST) run --rm typegen go run ./cmd/typegen

# Regenerates in the container, then diffs ON THE HOST: golang:alpine has no
# git, and .git isn't mounted into the typegen service - mounting it just to
# run `git diff` would risk corrupting the index from a Linux container
# against a Windows checkout for no benefit.
types-check-docker: types-docker
	cd ../.. && git diff --exit-code -- apps/frontend/src/shared/contracts \
	  && test -z "$$(git status --porcelain -- apps/frontend/src/shared/contracts)"

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
