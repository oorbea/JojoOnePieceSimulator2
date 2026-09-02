# Endpoints

Two catalogues share the same CRUD + picture-pipeline shape, wired as
parallel vertical slices: `/api/v1/stands` (`internal/infrastructure/api/endpoints/stand_endpoints.go`)
and `/api/v1/devil-fruits` (`internal/infrastructure/api/endpoints/devil_fruit_endpoints.go`).
Both expose `GET /`, `GET /{id}`, admin-only `POST /`, `PUT /{id}`,
`PATCH /{id}/picture`, `DELETE /{id}`. Devil fruits have no `evolvesFrom`
chain (that's Stand-only) but add a `fruitType` field/filter instead.

Picture uploads on both catalogues are transcoded by a single shared
background worker (`internal/application/services/picture_worker.go`),
routed to the right repository and object-storage key prefix
(`stands/...` vs `devil-fruits/...`) by the job's `enums.PowerKind`.

# Caching

`GET /api/v1/stands`, `GET /api/v1/devil-fruits`, and their single-item
routes are cached in three layers, each cutting a distinct cost out of the
read path:

1. **Repository cache** (`internal/infrastructure/cache/stand_repository.go`,
   `internal/infrastructure/cache/devil_fruit_repository.go`) - wraps
   `ports.IStandRepository`/`ports.IDevilFruitRepository` with Redis, avoiding
   the Postgres round trip on a hit. Each catalogue is cached under its own
   namespace (`stands`, `devil_fruits`); any write (`Save`/`Delete`/
   `UpdatePicture`, including the background picture worker publishing a
   finished transcode) invalidates that catalogue's whole namespace at once
   via an O(1) generation bump - no key enumeration. A `FindByID`/`FindByName`
   miss is itself cached briefly (a "tombstone"), so repeatedly requesting a
   nonexistent id doesn't keep hitting Postgres.
2. **Picture presign cache** (`internal/infrastructure/cache/picture_storage.go`) -
   wraps `ports.IPictureStorage`, caching the presigned R2 GET URL per object
   key (shared by both catalogues) so repeat reads return the *same* URL
   instead of a freshly signed one. This is what makes layer 3 possible: a
   presigned URL embeds a timestamp + signature, so without this cache no two
   responses for the same Stand/DevilFruit would ever be byte-identical.
3. **ETag / Cache-Control** (`internal/infrastructure/api/endpoints/cache_headers.go`) -
   an HTTP middleware on every read route that hashes the (now byte-stable)
   response body into an `ETag` and answers a matching `If-None-Match` with an
   empty `304`. Responses are marked `Cache-Control: private` (never
   `Cache-Control: public`) with `Vary: Authorization`, since every route
   behind this middleware requires a Bearer token.

**Fail-open.** All caching is optional and degrades gracefully: with
`REDIS_URL` unset (the default), no connection is ever attempted and every
request goes straight to Postgres/R2 - `make test`/`go run` need no Redis.
With Redis configured, any error or timeout from it (bounded by
`REDIS_OP_TIMEOUT`) is treated as a cache miss, never as a request failure -
an outage degrades latency, not availability. See `deployments/.env.example`
(there is no `apps/backend/.env.example` - the whole stack shares one) for
every `CACHE_*`/`REDIS_*` variable and its default.
