# Infrastructure

Implementations of `internal/domain/ports`, plus the HTTP/WS transport
layer. Everything that talks to Postgres, Redis, R2/S3, Google, or the
network lives here.

## `api/`

- **`endpoints/`** — Chi HTTP handlers, one file per resource
  (`stand_endpoints.go`, `devil_fruit_endpoints.go`, `stage_endpoints.go` via
  the vertical-slice pattern described in the backend root README,
  `auth_endpoints.go`, `game_endpoints.go`, `game_ws_endpoints.go`,
  `events_endpoints.go` for SSE). Cross-cutting: `middleware.go` (auth/role
  checks), `cache_headers.go` (ETag/`If-None-Match`/304), `ratelimit.go`,
  `locale.go` (`Accept-Language` resolution), `context.go`, `error_codes.go`
  / `errors.go`.
- **`dto/`** — request/response structs for every endpoint plus
  `game_ws.go` (WebSocket frame payloads) and `picture_event_payload.go`
  (SSE payloads). These are the Go source of truth the frontend's generated
  `contracts/` are typegen'd from.
- **`apierr/codes.go`** — the stable machine-readable error-code strings
  returned in `{ error, details? }` bodies.

## `auth/`

`google_verifier.go` — verifies a Google ID token's signature/`aud`.
`jwt_issuer.go` — issues/verifies this backend's own HS256 access tokens
(`JWT_SECRET`, no refresh tokens).

## `cache/`

Redis-backed decorators over the domain repository ports (see the backend
root README's Caching section): `stand_repository.go`,
`devil_fruit_repository.go`, `stage_repository.go` (each with a matching
`*_snapshot.go` cached shape), `picture_storage.go` (presigned-URL cache),
`keys.go` (namespacing + generation-bump invalidation). `cache/redis/` holds
the thin Redis client wrapper (`cache.go`) shared by all of them.

## `game/`

Small game-specific port implementations: `coinflip_tiebreaker.go`,
`default_weights.go` (assignment weights), `repo_power_pool.go` (power pool
sourced from the Stand/DevilFruit repositories).

## `gamestore/`

Live game-state storage, separate from the Postgres catalogue repositories
since games are ephemeral/high-churn: `memory.go` + `reaper.go` (in-process
store with idle-game eviction, used when Redis isn't configured) and
`gamestore/redis/` (`store.go`, `wire.go` — Redis-backed store with its own
wire/serialization format for `Game` snapshots).

## `idgen/`, `random/`, `imaging/`, `powersnap/`

- **`idgen/uuid_generator.go`** — the `IIdGenerator` implementation (UUIDv4).
- **`random/math_rand_v2.go`** — the `ports.Random` implementation.
- **`imaging/`** — `processor.go` implements `ports.ImageProcessor`
  (libvips-backed, build-tagged `vips`); `processor_stub.go` is the no-op
  build without that tag.
- **`powersnap/`** — `stand.go`/`devil_fruit.go`, snapshot helpers for
  denormalizing a Power into the shape the game domain embeds in a
  `Loadout`.

## `postgres/`

`pool.go` (pgx pool setup), `migrate.go` (goose migrations at startup).
`postgres/db/` is **sqlc-generated, do not edit** — `models.go` plus one
`*.sql.go` per table (`stands`, `devil_fruits`, `stages`, `users`,
`game_history`, `storage_objects`), generated from `apps/backend/db/`.

## `repositories/`

The Postgres implementations of the domain repository ports, wrapped by
`cache/` above: `stand_repository.go`, `devil_fruit_repository.go`,
`stage_repository.go`, `user_repository.go`, `game_history.go`,
`storage_ledger.go`, each paired with a `*_mapper.go` (sqlc row ↔ domain
entity) and `power_translations.go`. `pg_errors.go` maps Postgres errors
(unique violation, etc.) onto `domain/ports` sentinel errors.

## `storage/`

- **`s3store/backend.go`** — `ports.StorageBackend` implementation against
  R2/S3 (presigned URLs, upload/delete).
- **`fallback/picture_storage.go`** — wraps a `StorageBackend` to implement
  `ports.PictureStorage`, degrading gracefully when storage is unavailable.
