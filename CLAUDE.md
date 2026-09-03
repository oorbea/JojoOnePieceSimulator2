# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Online game about JoJo's Bizarre Adventure and One Piece — players get a
randomized loadout (Stand or Devil Fruit) and play through game modes
(Gauntlet, Versus). Not a real monorepo: `apps/backend` (Go) and
`apps/frontend` (Expo) are two standalone projects with their own
lockfiles/toolchains, tied together only by `deployments/`.

**Always check `ObsidianVault/` before implementing non-trivial changes** —
it holds architecture decisions, norms, and gotchas discovered the hard way.
Start at `ObsidianVault/overview.md`, which links out to the rest. Write
learnings/decisions back to the vault when you're done (see
`ObsidianVault/zettelkasten-workflow.md`). Established norms worth knowing
up front: all UI/UX work goes through the `frontend-design` +
`ui-ux-pro-max` skills (`norma-diseno-ui-ux.md`); every button gets a
tooltip (`norma-tooltips-y-ayuda-contextual.md`); a change isn't done until
`ObsidianVault/norma-verificacion-docker.md`'s Docker-based verification
passes.

## Commands

### Local stack (both apps + Postgres + Redis)

```
docker compose -f deployments/docker-compose.yml -f deployments/docker-compose.dev.yml up -d --build
```

Or `.\start.ps1` for an interactive menu (same compose stack + goose
migrations + common Go commands). Frontend on `:3000`, backend on `:8080`.

### Backend (`apps/backend/`, run via `make` from that directory)

Local dev is Docker-only — `migrate-*`/`db-*` targets exec into the running
container. `test`/`test-vips` run on the host by design (the runtime image
has no Go toolchain), but **on this machine Windows Application Control
randomly blocks host-run `go test` binaries** — prefer the `-docker`
variants below whenever in doubt:

```
make test                    # go test ./... (host)
make test-docker             # same, via docker compose backend-test (preferred on this machine)
make test-vips-docker        # -tags vips (real libvips), via docker
make test-integration-docker # -tags integration; needs `make db-up` first
make types-check             # regenerate frontend contracts and fail if stale (host, safe)
make types-check-docker      # same, via docker
make migrate-up / migrate-down / migrate-status
make db-up / db-down
```

A single test file/package: `go test ./internal/application/services/... -run TestName`
(or the `-docker` equivalent via `docker compose ... run --rm backend-test go test <args>`).

### Frontend (`apps/frontend/`, pnpm)

```
pnpm install
pnpm web            # Expo web dev server
pnpm lint           # ESLint flat config
pnpm typecheck      # tsc --noEmit
pnpm test           # jest (jest-expo, two-project split: logic/native)
pnpm format         # prettier --write
pnpm build:web      # expo export -p web -> dist/
```

`pnpm-workspace.yaml` here is not a monorepo root — it only holds the
build-script allowlist pnpm needs for native postinstall scripts.

### Generated contracts

`apps/backend/cmd/typegen` is the single producer of
`apps/frontend/src/shared/contracts/` (every wire DTO, enum, WS
command/frame shape, error code) — Go is the source of truth. Run
`make types` (backend dir) after changing a Go DTO/enum/WS payload/error
code, or use the `regenerate-contracts` skill. Never hand-edit generated
files under `contracts/`; `make types-check`/CI's `contracts` job fails
loudly on drift instead of leaving it stale.

## Architecture

### Backend — hexagonal (`apps/backend/internal/`)

- **`domain/`** — framework-free core, no dependency on
  `application`/`infrastructure`. `entities/` (one package per bounded
  concept: `powers` (Stand/DevilFruit), `user`, `game` — the largest
  package, an explicit state machine over `enums.GameState`
  (`LOBBY → ASSIGNING → VOTING → [TIEBREAK] → RESOLVING → ... → FINISHED/ABORTED`)
  delegating mode-specific behaviour to an `IGameMode` strategy
  (`gauntlet_mode.go`/`versus_mode.go`)), `valueobjects/` (generic id
  helpers), `enums/` (~20 typed-constant packages), `ports/` (outbound
  interfaces implemented in `infrastructure/`, including the sentinel
  domain errors every layer matches against).
- **`application/services/`** — use cases; depend only on `ports`
  interfaces injected at construction, no transport/concrete infra.
  CRUD-style catalogue services (`stand_service.go`, `devil_fruit_service.go`,
  `stage_service.go`) share one shape (`...Input` struct,
  Create/Update/Get/List/Filter/Delete). `game_service.go` (~1.5k lines) is
  the largest, orchestrating lobby → game lifecycle, voting, phase timers
  on top of the `game.Game` state machine. Event hubs
  (`game_event_hub.go`, `picture_event_hub.go`) fan domain events out to
  infra SSE/WS handlers.
- **`infrastructure/`** — everything touching Postgres/Redis/R2/Google/
  network. `api/endpoints/` (Chi HTTP handlers, one vertical slice per
  resource), `api/dto/` (the Go source of truth typegen reads from),
  `auth/` (Google ID token verification, HS256 JWT issuing), `cache/`
  (Redis decorators over repository ports), `postgres/db/` (sqlc-generated,
  do not edit), `repositories/` (Postgres impls, paired `*_mapper.go`),
  `gamestore/` (ephemeral live-game state, separate from Postgres
  catalogues — in-memory or Redis-backed), `storage/` (R2/S3 + graceful
  fallback), `imaging/` (libvips processor, build-tagged `vips`, with a
  no-op stub build).

Stands and Devil Fruits are parallel vertical slices sharing one CRUD +
picture-pipeline shape end to end (service → endpoint → repository →
cache), including a single shared background picture-transcode worker
(`picture_worker.go`) routed by `enums.PowerKind`. Catalogue reads
(`GET /api/v1/stands`, `/devil-fruits`, and single-item routes) are cached
in three layers: Redis repository cache (namespace-wide invalidation via
generation bump), presigned-URL cache (makes responses byte-stable), and
ETag/304 middleware. All caching fails open — `REDIS_URL` unset means
straight to Postgres/R2, and any Redis error/timeout is a cache miss, never
a request failure.

### Frontend — Feature-Driven + Container/Presentational (`apps/frontend/src/`)

Full convention: `src/features/README.md` — read before adding a feature.

- `app/` — Expo Router routes only, thin shims rendering a container.
- `src/features/<feature>/` — everything one capability owns (`api/`,
  `components/{containers,presentational}`, `hooks/`, `stores/`, `types/`,
  `index.ts` barrel). Copy `src/features/_template/` to start a new one.
  **Cross-feature imports go only through the other feature's `index.ts`.**
  Containers wire data/navigation/forms (TanStack Query, zustand,
  react-hook-form); presentational components are pure Tamagui UI, props
  in, JSX out, no hooks into query/store/router.
- `src/shared/` — same container/presentational split, plus the axios
  client + auth/ETag interceptors + error normalization
  (`shared/api/`), platform-agnostic wrappers (`shared/lib/`), app-wide
  zustand state (`shared/stores/`), validated env access
  (`shared/config/env.ts`), and **`shared/contracts/`** — generated, do not
  hand-edit (see above).

Talking to the backend: base path `/api/v1`; `POST /api/v1/auth/google`
returns a plain bearer JWT (24h, no refresh — a 401 clears the session and
requires re-login); every other route needs `Authorization: Bearer`, writes
also need `ADMIN` role; reads are conditional-GET aware (ETag/
`If-None-Match` handled by the axios interceptor); errors normalize to
`AppError` from `{ error, details?[] }`. Realtime transport is a native
`WebSocket` (no socket.io) at `/api/v1/games/{id}/ws`, client in
`src/features/game/stores/game-socket.store.ts` (refcounted, exponential
backoff, RESYNC on reconnect).

### Deployment

CI (`.github/workflows/ci.yml`) gates merges to `main` behind one
`ci-success` required check; CD (`cd.yml`) deploys on every push to `main`.
Prod is the same compose stack with `docker-compose.prod.yml` layered on;
see `deployments/README.md` for the server layout, required GitHub
secrets/variables, and the DB-tunnel override for admin Postgres access.
Auth secrets (`GOOGLE_CLIENT_ID`, `JWT_SECRET`, `ADMIN_EMAILS`) are the
easiest to get wrong — a bad value fails at boot or silently locks out
admin routes.
