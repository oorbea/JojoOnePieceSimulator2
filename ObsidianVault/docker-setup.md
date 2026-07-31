---
title: Docker setup
tags:
  - project
  - jojo-onepiece-simulator
  - docker
---

# Docker

## Frontend images

- **Dev**: `apps/frontend/Dockerfile.dev` — `node:22-alpine`, corepack pnpm, `pnpm install --frozen-lockfile`, expects source volume-mounted, `CMD pnpm expo start --web --port 8081`.
- **Prod**: `deployments/docker/Dockerfile.frontend` — multi-stage. Builder = `node:22-alpine` → `pnpm expo export -p web` → `dist/`. Runtime = `nginx:alpine`, non-root user uid 10001 (matches backend convention), serves via `deployments/docker/nginx.frontend.conf` (long-cache `/_expo/static/`, no-cache `sw.js`/`manifest.json`, SPA fallback).
- Build context for prod image = **repo root** (matches backend's `Dockerfile.backend` convention), not `apps/frontend`.
- Both need `RUN touch tamagui-web.css` after source COPY — see [[frontend-stack]] gotchas.
- Prod build needs `ARG`/`ENV` for `EXPO_PUBLIC_API_URL` / `EXPO_PUBLIC_SOCKET_URL` since Expo inlines them at build time.

## docker-compose.yml

- Backend service: added `CORS_ALLOWED_ORIGINS: http://localhost:${FRONTEND_PORT:-3000}` — otherwise backend deny-all blocks the browser.
- `frontend` service: build context `..` (repo root), dockerfile `deployments/docker/Dockerfile.frontend`, port `${FRONTEND_PORT:-3000}:80`, `depends_on: backend`.
- All 4 services (`postgres`, `redis`, `backend`, `frontend`) join an **external** `public-net` network (`networks: { public-net: { external: true } }`). Compose does NOT create it — the target server already has it, shared across services on that host. Must exist before `docker compose up` (`docker network create public-net`), otherwise compose errors at `up` time (not at `config` time — `docker compose config` validates fine even if the network doesn't exist yet).
- `postgres` and `redis` no longer publish host ports — reachable only inside `public-net`, by service name (`postgres:5432`, `redis:6379`). `backend` keeps its published port since it's the actual API entrypoint other things (frontend on the same host, or a reverse proxy) need to reach.
- All 4 services now have `restart: unless-stopped` (none did before).

## Env vars — single source of truth (2026-07-31)

- `apps/backend/.env.example` **deleted**. Local dev is Docker-only now — no more `go run`/`make run` path on host, so no reason to keep a second, overlapping env template.
- `deployments/.env.example` is the ONLY env reference in the repo: superset of backend runtime vars + compose interpolation vars + frontend build-time vars, each line marked `[SECRET]` or `[CONFIG]`. Markers exist so wiring GitHub Actions later is mechanical — `[SECRET]` → repo/environment secrets, `[CONFIG]` → workflow env or repo vars. Compose reads `${VAR}` the same way whether it comes from a `.env` file or the CI runner's real environment, so no `.env` file is required in CI at all.
- `CORS_ALLOWED_ORIGINS` and `REDIS_URL` are deliberately **absent** from `.env.example` (with a comment explaining why) — both are computed/overridden directly in `docker-compose.yml`'s backend `environment:` block, so putting them in `.env` would be silently ignored.
- Backend `env_file` now points at `deployments/.env` (was `apps/backend/.env`) with `required: false`, so compose doesn't fail when no `.env` file exists (CI case — vars come from the runner's environment instead).

## Backend image additions (2026-07-31)

`deployments/docker/Dockerfile.backend` runtime stage is minimal alpine — only the compiled `server` binary, no Go toolchain, no source. The app only embeds+auto-runs `goose up` on startup (`internal/infrastructure/postgres/migrate.go`), so there was no way to run `migrate-down`/`migrate-status` against a Dockerized Postgres. Fixed by adding, in the same multi-stage build:

- A second binary: the `goose` CLI, built via `go build -o /out/goose github.com/pressly/goose/v3/cmd/goose` — already a `go.mod` dependency, no extra fetch needed.
- The raw migration `.sql` files copied to `/migrations` in the runtime image (the app's own copy is `//go:embed`-compiled into the binary, not usable by an external CLI).

`apps/backend/Makefile`'s `migrate-up/down/status` now run via `docker compose exec backend sh -c 'goose -dir /migrations postgres "$DATABASE_URL" ...'` instead of a local `goose` + host `DATABASE_URL`. `db-up`/`db-down` unchanged (already used compose).

**Gotcha**: `test`/`test-vips` Makefile targets must stay host-based (`go test ./...`) — I initially tried routing them through `docker compose exec backend` too, but the runtime image has no Go toolchain or source to test against. Only migrations moved into the container; tests didn't.

## Root .dockerignore gotcha

Root `.dockerignore` (`C:\code\JojoOnePieceSimulator2\.dockerignore`) needed `**/node_modules`, `**/dist`, `**/.expo`, `**/tamagui-web.css` added — without it local `node_modules` leaked into build context (repo root) and clashed with fresh `pnpm install` inside the image (`ERR_PNPM_ABORTED_REMOVE_MODULES_DIR_NO_TTY`). But excluding `tamagui-web.css` from context also means it must be created inside the Dockerfile via `touch`, not copied in.

## Verified

Prod image built + run standalone: `docker run` + curl confirmed `/`, `/manifest.json`, `/sw.js` all correct (200s, `sw.js` has `Cache-Control: no-cache`). NOT yet verified: `docker compose up` running backend+frontend together (deferred as a manual/optional check per plan).

Related: [[frontend-stack]], [[backend-contract]]
