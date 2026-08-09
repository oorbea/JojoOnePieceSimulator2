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
- 2026-08-09: added the "Object Storage fallback chain" + per-provider (`R2_*`/`B2_*`/`SUPABASE_*`) sections for the storage tiers — see [[storage-fallback-chain]]. `B2_*`/`SUPABASE_*` are conditionally required: only when their name appears in `STORAGE_PROVIDERS` (default is `r2` only, so a fresh checkout needs none of them filled in).

## Backend image additions (2026-07-31)

`deployments/docker/Dockerfile.backend` runtime stage is minimal alpine — only the compiled `server` binary, no Go toolchain, no source. The app only embeds+auto-runs `goose up` on startup (`internal/infrastructure/postgres/migrate.go`), so there was no way to run `migrate-down`/`migrate-status` against a Dockerized Postgres. Fixed by adding, in the same multi-stage build:

- A second binary: the `goose` CLI, built via `go build -o /out/goose github.com/pressly/goose/v3/cmd/goose` — already a `go.mod` dependency, no extra fetch needed.
- The raw migration `.sql` files copied to `/migrations` in the runtime image (the app's own copy is `//go:embed`-compiled into the binary, not usable by an external CLI).

`apps/backend/Makefile`'s `migrate-up/down/status` now run via `docker compose exec backend sh -c 'goose -dir /migrations postgres "$DATABASE_URL" ...'` instead of a local `goose` + host `DATABASE_URL`. `db-up`/`db-down` unchanged (already used compose).

**Gotcha**: `test`/`test-vips` Makefile targets must stay host-based (`go test ./...`) — I initially tried routing them through `docker compose exec backend` too, but the runtime image has no Go toolchain or source to test against. Only migrations moved into the container; tests didn't.

## Root .dockerignore gotcha

Root `.dockerignore` (`C:\code\JojoOnePieceSimulator2\.dockerignore`) needed `**/node_modules`, `**/dist`, `**/.expo`, `**/tamagui-web.css` added — without it local `node_modules` leaked into build context (repo root) and clashed with fresh `pnpm install` inside the image (`ERR_PNPM_ABORTED_REMOVE_MODULES_DIR_NO_TTY`). But excluding `tamagui-web.css` from context also means it must be created inside the Dockerfile via `touch`, not copied in.

## Verified

Prod image built + run standalone: `docker run` + curl confirmed `/`, `/manifest.json`, `/sw.js` all correct (200s, `sw.js` has `Cache-Control: no-cache`). `docker compose up` running backend+frontend together, full Google login flow end-to-end: verified working 2026-08-01 (see [[frontend-stack]] for the 3 bugs found/fixed along the way).

## Rebuild gotcha (2026-08-01)

`docker compose up -d --build` alone is not always enough to prove a frontend fix landed — BuildKit layer caching can report the `pnpm expo export` step as `CACHED` even after a real source change, if an earlier layer's cache key didn't bust correctly. When in doubt (e.g. verifying a fix actually took effect), use `docker compose build --no-cache frontend` then `docker compose up -d --force-recreate frontend`, and independently confirm via `docker compose exec frontend grep ... /usr/share/nginx/html/...` that the expected string/behavior is actually in the served files — don't trust the build log alone.

## Compose split: base / dev / prod (2026-08-02)

`deployments/docker-compose.yml` no longer publishes `ports:` for `backend`/`frontend` — it's the shared base for both environments now. Two overrides layer on top:

- `deployments/docker-compose.dev.yml` — adds back `backend`'s `${PORT:-8080}:${PORT:-8080}` and `frontend`'s `${FRONTEND_PORT:-3000}:80`. `apps/backend/Makefile`'s `$(COMPOSE)` now passes both `-f` files.
- `deployments/docker-compose.prod.yml` — overrides `CORS_ALLOWED_ORIGINS` to `https://jojo-one-piece-simulator.duckdns.org`. No ports at all: in prod, Nginx Proxy Manager (already on the host's `public-net`) reaches both services by container name.

**Gotcha discovered**: Compose *merges* `ports:` lists across `-f` files instead of replacing them — an override cannot un-publish a port the base file already publishes. That's why publishing had to move entirely out of the base file into `docker-compose.dev.yml`, rather than trying to have `docker-compose.prod.yml` "remove" it.

Also added a `backend` healthcheck to the base file (`wget -qO- http://localhost:${PORT:-8080}/health`, busybox wget already in the alpine runtime image) — needed by the CD pipeline to know when the new container is actually ready, see [[cicd-deployment]].

Related: [[frontend-stack]], [[backend-contract]], [[cicd-deployment]]
