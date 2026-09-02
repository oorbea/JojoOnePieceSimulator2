---
title: "Redis read-through cache for the Stage catalogue (2026-09-02)"
tags:
  - project
  - jojo-onepiece-simulator
  - backend
  - cache
  - performance
---

# Stage Redis cache (2026-09-02)

Closes the last uncached catalogue. Stand and DevilFruit have had a
read-through Redis decorator since the cache tanda; `Stage` did not, so every
admin list/filter/detail request **and** every `GameService.CreateGame` (via
`ports.IStageCatalog.Stages`) was a live Postgres round trip. Gap was flagged
in [[stages_admin_crud_2026-08-11]], [[admin-search-and-filters]],
[[game-stage-content]] and [[game-lobby-persistence]]; all now point here.

Branch `feat/stage-redis-cache`, four commits.

## What was built

- **`internal/infrastructure/cache/stage_snapshot.go`** - a package-local
  `stageSnapshot` + `marshal/unmarshalStage(s)`. `game.Stage`'s fields are all
  unexported and its own snapshot helpers (`game.snapshotStage`/`restoreStage`)
  are package-private, so this package carries its own JSON shape.
- **`internal/infrastructure/cache/stage_repository.go`** - the decorator.
  Read-through on `Stages` (per manga), `List` (per locale), `Filter` (per
  canonical filter + locale) and `FindByID` (per id + locale, with a
  not-found tombstone). `Save`/`Delete`/`UpdatePicture` delegate, then flush
  the whole `stages` namespace. `Translations` deliberately bypasses the
  cache - admin edit forms need a fresh read of every locale, same call as
  `StandRepository.Translations`.
- **`keys.go`** - `stagesNamespace`, `stageFilterKey`, `stageCatalogKey`.
- **`config.go` + `deployments/.env.example`** - `CACHE_STAGE_TTL`, default
  `5m`, same as Stand/DevilFruit.
- **`cmd/app/main.go`** - reordered wiring (below).
- **`stage_repository_test.go`** - 12 cases over an in-memory
  `countingStageRepository`.

## Design decisions

**`StageStore` is the wrapped type, not `ports.IStageRepository`.** The
concrete adapter satisfies *both* the admin CRUD port and the gameplay
catalogue port over the same table, and so must the decorator:

```go
type StageStore interface {
    ports.IStageRepository
    ports.IStageCatalog
}
```

Without this the decorator could only be handed to one of the two consumer
groups, and the other would read straight through - so an admin write would
invalidate a cache the gameplay path was not using, or worse, two separate
decorator instances would each hold entries the other's writes never flushed.

**`Stages(manga)` gets its own key, separate from `Filter{Manga: manga}`.**
They run the same query today, but they are different contracts: `Filter` is
admin-facing and locale-aware, `Stages` is gameplay-facing and resolves
`Description` at a fixed `enums.EnGB` (see `ports.IStageCatalog`'s doc on
*why* it takes no locale). Sharing one slot would make a future divergence in
either an ordering or a resolution rule silently answer the wrong request.
Pinned by `TestStageRepository_StagesAndFilterDoNotShareASlot`, which would
fail the moment the two keys collapsed.

**`main.go` reordering.** `stageRepository := repositories.NewStageRepository(pool)`
used to be built *after* the `if cfg.CacheEnabled && cfg.RedisURL != ""` block,
and was consumed raw by four call sites. It now moves up next to
`standRepository`/`devilFruitRepository` (it only needs `pool`), and the
decorator is constructed **once** inside the `if`, exposed through two
interface-typed variables:

```go
var stageRepo    ports.IStageRepository = stageRepository
var stageCatalog ports.IStageCatalog    = stageRepository
...
cachedStages := cache.NewStageRepository(stageRepository, redisCache, cfg.CacheStageTTL, cfg.CacheNotFoundTTL)
stageRepo    = cachedStages
stageCatalog = cachedStages
```

`stageRepo` feeds the picture publisher, `NewStageService` and
`NewGameEndpoints`; `stageCatalog` feeds `NewGameService`. Note the decorator
wraps `stageRepository` (the concrete adapter, which satisfies `StageStore`),
**not** `stageRepo` - the narrowed interface variable would not compile there,
and that is the shape that stops anyone accidentally building two decorators.

## Traps hit / worth remembering

- `ports.IStageRepository.FindByID` returns `(game.Stage, error)` **by value**,
  unlike Stand/DevilFruit's pointer. A tombstone hit therefore has to return
  `game.Stage{}, ports.ErrStageNotFound` - returning the zero `Stage` with a
  nil error would look like a real empty stage to callers. Covered by a test
  that asserts both the error *and* the zero value.
- `game.NewStage` takes **six** args (`id, manga, order, name, description,
  picture`), not five - the plan for this tanda assumed five. `hydrate()`
  follows `restoreStage`'s exact order anyway: `ParseManga` → `NewStage` →
  `ParsePictureStatus` → `SetPictureRenditions`.
- `game.StageID` is `[16]byte`, so it's assignable to/from a plain `[16]byte`
  snapshot field, and it satisfies `fmt.Stringer` - `idKey` takes it as-is.
- `CacheStageTTL` had to be touched in **four** places in `config.go` (doc
  comment, struct field, loader, return literal). Forgetting the return
  literal leaves it silently `0` at runtime with no error - the exact failure
  mode this note exists to warn about next time.
- `gofmt -l` lists essentially the whole backend on this machine (CRLF working
  tree), so it is useless as a diff check here. `go build ./...`,
  `go vet ./...` and `go test ./...` are the real signal.

## Verification

`go build ./...`, `go vet ./...` and the full `go test ./...` all clean
natively on the host (Go ran fine this session - the App Control block noted
in [[norma-verificacion-docker]] did not bite). No Docker stack was raised:
other agents were working in parallel on the shared local stack.

## Out of scope

- No per-key invalidation - the namespace is flushed wholesale on any write,
  same reasoning as `standsNamespace` (a Stage appears in `all`, in an unknown
  set of `filter:*` entries, and in a catalogue entry).
- No HTTP-level `Cache-Control` change for the stage endpoints;
  `CacheHTTPMaxAge` behaviour is untouched.
- Frontend caching (TanStack Query `stageKeys`) untouched - that layer already
  worked, see [[admin-search-and-filters]].

Related: [[admin-search-and-filters]], [[game-stage-content]],
[[game-lobby-persistence]], [[gameplay-application-layer]],
[[admin-crud-cache-stale-sw]].
