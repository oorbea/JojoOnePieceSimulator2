---
title: "Feature: game lobby persistence (2026-08-11)"
tags:
  - project
  - jojo-onepiece-simulator
  - backend
  - feature
  - gameplay
---

# Game lobby persistence (2026-08-11)

Closes the first half of [[gameplay-application-layer]]'s "still not built" list: a Redis-backed
`ports.IGameStore`, a Postgres-backed stage catalog, and a persistent `ports.IGameHistory`. See
[[game-realtime-transport]] for the WebSocket/HTTP half that shipped alongside this.

## Status

Done. `MemoryGameStore` stays as the fallback when `REDIS_URL` is unset (and is what every
`services` unit test still uses); `main.go` picks the Redis-backed store whenever it's set.

## Why

`game.Game` had 9 private fields and no exported (de)serialization — `ports.IGameStore`'s own doc
explicitly called this out as "not designed yet, on purpose" in the previous tanda. Without a
snapshot seam, a Redis adapter is impossible: you cannot round-trip a mid-round match (votes,
loadouts, tiebreak state) through `MemoryGameStore`'s in-process pointer semantics.

## The snapshot seam (`entities/game/snapshot.go`)

- `Snapshot`/`ConfigSnapshot`/`ParticipantSnapshot`/`TeamSnapshot`/`StageSnapshot`/
  `RoundSnapshot`/`BallotSnapshot`/`VoteSnapshot`/`RoundResultSnapshot`/`LoadoutSnapshot` — plain
  structs, **no JSON tags** (those live in infrastructure, see below). Enums travel as their
  `String()` form and come back through the existing `enums.Parse*` functions. IDs stay `[16]byte`.
  `BallotSnapshot.Votes` is a **slice**, not a map — `[16]byte` isn't a valid JSON object key and a
  slice serializes deterministically.
- `Ballot` gained a new getter, `Votes() map[ParticipantID]OptionID` (copy-returning) — the one
  domain-level addition this seam required, since `HasVoted`/`Count` can't reconstruct *which*
  option someone voted for.
- `(*Game).Snapshot()` — pure read, does **not** drain `PullEvents()`. Safe because `withGame`
  always calls `publish(g)` (which drains events) before every `Save`, so `g.events` is empty at
  snapshot time in practice.
- `Restore(Snapshot) (*Game, error)` — rebuilds a `*Game` bypassing `addParticipant`'s capacity
  checks (a full lobby must still be restorable) but re-validating every value object through its
  normal constructor, so a corrupted/cross-version payload fails loudly. Always installs
  `DefaultLoadoutEvaluator{}` — the evaluator is behaviour, not data, and isn't part of a Snapshot.
  A small refactor (`modeFor(enums.GameModeKind) (IGameMode, error)`) extracted out of `NewGame`'s
  switch is shared by both constructors, since the mode is derived, never serialized.
- `LoadoutSnapshot` embeds `*powers.Stand`/`*powers.DevilFruit` **in full**, not by ID — a
  `Loadout` is documented as an immutable snapshot of abilities for a game/round, so an admin
  editing or deleting a power later must not retroactively change or brick a live match. The
  serialization logic itself was promoted out of `infrastructure/cache`'s private
  `standSnapshot`/`devilFruitSnapshot` into a new shared package, `infrastructure/powersnap`
  (`StandSnapshot`/`DevilFruitSnapshot` + `OfStand`/`OfDevilFruit`/`Hydrate`); `cache`'s own
  `marshalStand`/etc are now 3-line wrappers over it, untouched behaviourally.
- `AvailablePowers` is **never** persisted — it's never a field of `Game`, and
  `GameService.beginRound` rebuilds it fresh from `IGamePowerPool` on every assignment.

## Redis store (`infrastructure/gamestore/redis`)

`wire.go` mirrors `Snapshot` field-for-field with JSON tags in a versioned envelope
(`{v, updatedAt, game}`, `snapshotVersion = 1`). An unknown version is a hard **error**, never a
silent zero value — a live lobby unreadable after a bad deploy is a visible `GAME_NOT_FOUND`, not
silent corruption; the 2h TTL means it self-heals.

`store.go` — three Redis keys per game, one shared TTL refreshed on every `Create`/`Save`:

```
jojo:game:id:<gameID>     -> envelope JSON (the aggregate)
jojo:game:code:<CODE>     -> <gameID>, for GetByCode
jojo:game:codeof:<gameID> -> <CODE>, for Code and for refreshing the code index cheaply
```

Deliberately shares the `jojo:` root with `infrastructure/cache/redis` but **not** its
`<ns>:<gen>:<key>` layout — generation-based invalidation is a cache concept, and an `INCR` there
would orphan live lobbies. Every write is a single `EVAL` (same `redis.NewScript` pattern the
cache uses). `DeleteExpired` is a **deliberate no-op** returning 0 — expiry is delegated entirely
to Redis's own `PX`, which already implements "remove anything last saved more than `olderThan`
ago"; `Reaper` stays wired unconditionally against it since it's harmless (0 removed, no log line).

**Fail-CLOSED**, the opposite of `infrastructure/cache/redis.Cache`'s fail-open contract: every
error surfaces to the caller, a corrupt payload is an error (not a miss, and the key is left in
place for debugging), `New` PINGs at startup and `main.go` fatals on failure, and it uses its own
`GameStoreOpTimeout` (default 2s — far above the cache's 200ms `REDIS_OP_TIMEOUT`, since this store
should wait out transient latency rather than fail a vote).

**Bug caught during integration testing**: the first draft of `createScript` stored the join code
itself under the code→id key (`SET codeKey <code>`) instead of the game's actual ID, so
`GetByCode` always failed with "invalid id". Caught by
`TestStore_CreateGetSave_RoundTrip` against a real Redis instance before it shipped — worth noting
because it silently passed `go vet`/`go build` and would have made every `GetByCode` call in
production 404.

**Semantic difference from `MemoryGameStore`** worth knowing: the in-memory store returns the same
`*game.Game` pointer it stored, so a partial mutation followed by a failing step in `withGame`'s
error path is still visible on the next `Get`. The Redis store returns a fresh object per `Get`, so
that same partial mutation is discarded (arguably more correct/transactional, but a real behaviour
difference between the two adapters).

## Stage catalog (`db/migrations/00008_stages.sql`, `repositories.StageRepository`)

Replaces the old `internal/infrastructure/game/static_stage_catalog.go` stub (now deleted). New
`stages` table: `manga` enum, `position integer`, `name text`, `UNIQUE (manga, name)` — no unique
on `(manga, position)` on purpose (an admin swapping two positions in one transaction would
otherwise deadlock on it; duplicate positions are harmless since `game.Interleave` sorts anyway).
Seeded with the same 19 names the old stub hardcoded. New port `ports.IStageRepository`
(`List/FindByID/Save/Delete`) alongside the existing `ports.IStageCatalog` — one adapter
(`StageRepository`) satisfies both. Admin CRUD is exposed at `/api/v1/stages` — see
[[game-realtime-transport]]. Stage names stay untranslated (same reasoning `i18n-multi-language.md`
gives for `powers.name`: they're proper nouns).

## Game history (`db/migrations/00009_game_history.sql`, `repositories.GameHistory`)

Two tables: `game_results` (PK on `game_id`, making `Record` idempotent under
`finalizeLocked`'s best-effort retry — `ON CONFLICT (game_id) DO NOTHING`) and
`game_result_participants` (`user_id` nullable with `ON DELETE SET NULL`, **never** `CASCADE` — a
user deleting their account must not erase the other players' match history; `display_name` is
snapshotted so the row stays readable once `user_id` goes null). Required extending the domain:
`GameResult` gained `Participants []ParticipantOutcome`, filled by both `GauntletMode.Outcome` and
`VersusMode.Outcome` via a new private `participantOutcomes(g *Game)` helper.

## Config / compose

New `GAME_STORE_OP_TIMEOUT` (default 2s). `deployments/docker-compose.yml`'s `redis` service
gained `--appendonly yes` + a named volume (`redis-data`) — it now holds live match state, not
just a cache, so a Redis restart must not silently wipe every in-progress game.

## Tests

`entities/game/snapshot_test.go` — full round trip of a deliberately gnarly mid-match Versus game
(2 teams, a bot, a disconnected/reassigned host, an evolved Stand + a DevilFruit loadout, one
resolved round, one round left genuinely mid-`TIEBREAK`), asserting both data fidelity and that the
restored `*Game` still *behaves* (`CloseVoting` on the restored ballot, no evaluator-nil panic).
Plus rejection tests for unknown enums, a malformed ballot, and a zero `Snapshot`.
`infrastructure/gamestore/redis/wire_test.go` — pure encode/decode round trip (no Redis), unknown
version, truncated JSON. `store_test.go` — env-gated on `TEST_REDIS_URL` (skips without it, same
convention as `infrastructure/cache/redis`), covers create/get/save/delete, the duplicate-code
conflict, TTL refresh, and the corrupt-payload-survives-and-errors case. `stage_repository_test.go`
and `game_history_test.go` are `//go:build integration`, run against real Postgres via
`make test-integration-docker` — added `TEST_REDIS_URL` to `docker-compose.test.yml` too, since it
was missing and every Redis-gated unit test was silently skipping even under that target.

## Known debt / deliberately not done here

- `Restore` re-validates loadout invariants (`ErrSpin4Required`, which depends on the by-name
  `requiresSpin4Names` lookup in `power_traits.go` — see [[gameplay-domain-design]]'s existing
  debt note). If that table changes while lobbies are live, an affected match becomes unrestorable
  under fail-closed semantics. TTL (2h) bounds the blast radius; not fixed here.
- `IStageCatalog.Stages` is now a Postgres round trip on every `CreateGame`. Trivial today; a cache
  decorator mirroring `infrastructure/cache/stand_repository.go` is the obvious follow-up if it
  ever matters.
- Single-instance assumption baked in further: the Redis store has no distributed lock (per-`GameID`
  in-process mutexes still suffice), and scaling the backend horizontally would additionally need a
  Redis pub/sub sibling for `GameEventHub` — see [[ADR]].

Related: [[gameplay-application-layer]], [[gameplay-domain-design]], [[game-realtime-transport]],
[[ADR]], [[backend-contract]].
