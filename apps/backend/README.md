# Caching

`GET /api/v1/stands` and `GET /api/v1/stands/{id}` are cached in three layers,
each cutting a distinct cost out of the read path:

1. **Stand repository cache** (`internal/infrastructure/cache/stand_repository.go`) -
   wraps `ports.IStandRepository` with Redis, avoiding the Postgres round trip
   on a hit. Cached under a single `stands` namespace; any write
   (`Save`/`Delete`/`UpdatePicture`, including the background picture worker
   publishing a finished transcode) invalidates the whole namespace at once
   via an O(1) generation bump - no key enumeration. A `FindByID`/`FindByName`
   miss is itself cached briefly (a "tombstone"), so repeatedly requesting a
   nonexistent id doesn't keep hitting Postgres.
2. **Picture presign cache** (`internal/infrastructure/cache/picture_storage.go`) -
   wraps `ports.IPictureStorage`, caching the presigned R2 GET URL per object
   key so repeat reads return the *same* URL instead of a freshly signed one.
   This is what makes layer 3 possible: a presigned URL embeds a
   timestamp + signature, so without this cache no two responses for the same
   Stand would ever be byte-identical.
3. **ETag / Cache-Control** (`internal/infrastructure/api/endpoints/cache_headers.go`) -
   an HTTP middleware on both read routes that hashes the (now byte-stable)
   response body into an `ETag` and answers a matching `If-None-Match` with an
   empty `304`. Responses are marked `Cache-Control: private` (never
   `Cache-Control: public`) with `Vary: Authorization`, since every route
   behind this middleware requires a Bearer token.

**Fail-open.** All caching is optional and degrades gracefully: with
`REDIS_URL` unset (the default), no connection is ever attempted and every
request goes straight to Postgres/R2 - `make test`/`go run` need no Redis.
With Redis configured, any error or timeout from it (bounded by
`REDIS_OP_TIMEOUT`) is treated as a cache miss, never as a request failure -
an outage degrades latency, not availability. See `.env.example` for every
`CACHE_*`/`REDIS_*` variable and its default.
