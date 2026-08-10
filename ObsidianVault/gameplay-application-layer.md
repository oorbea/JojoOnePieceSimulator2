---
title: "Feature: game application layer (2026-08-10)"
tags:
  - project
  - jojo-onepiece-simulator
  - backend
  - feature
  - gameplay
---

# Gameplay application layer (2026-08-10)

For the gameplay rules, see [[gameplay-game-modes]]. For the domain layer this sits on top of, see
[[gameplay-domain-design]]. This note is the technical map for the tanda that made the
Gauntlet/Versus domain actually runnable: `internal/application/services/game_service.go` plus the
cheap adapters behind it. No HTTP/WS routes exist yet — that's still the next tanda.

## Why

`game.Game` (built 2026-08-10, same day, earlier commit) was a complete but inert state machine:
nothing instantiated it, stored it, closed a voting window, resolved a tie, or drained its events.
This pass wires all of that up, end to end, behind fakes/stubs where a real adapter doesn't exist
yet (stage catalog, weights, tiebreaker) or isn't in scope for this tanda (Redis, history,
websockets) — every choice below was made by the owner when asked.

## `GameService` (`internal/application/services/game_service.go`)

One method per use case, all serialized per-`GameID` via `withGame` (a private helper: locks that
Game's own `*sync.Mutex`, loads it from the store, runs the closure, publishes whatever
`DomainEvent`s it produced — even on error, since a rejected call can still have partially mutated
the aggregate — then either finalizes the Game (state is `FINISHED`/`ABORTED`) or saves it back).
One mutex per `GameID`, not a global lock, so unrelated lobbies never contend.

- `CreateGame` — builds `Config`/`Team`(s), resolves stages from `ports.IStageCatalog` per selected
  manga (`game.Interleave` for Gauntlet's fixed round order, a flat pool for Versus), seats the
  host, generates a 6-char join code (`ABCDEFGHJKLMNPQRSTUVWXYZ23456789` — no `0/O/1/I` — retried up
  to 5 times against a collision, mirroring `AuthService.uniqueUsername`), and creates it in the
  store.
- `JoinByCode` / `LeaveGame` / `AddBot` / `RemoveBot` — membership. Bots are host-only and explicit
  (`AddBot`/`RemoveBot`), no autofill. `JoinByCode` auto-balances a new human onto whichever Versus
  Team has fewer members (ties keep Team A).
- `StartGame` → `beginRound` (private, called with the Game's lock already held): resolves
  `ports.IAssignmentWeights`, builds one fresh `game.AvailablePowers` **per Team** from
  `ports.IGamePowerPool` (only when it's the first round, or the mode's
  `IGameMode.ReassignsEachRound()` — i.e. every round for Versus, once for Gauntlet), calls
  `Game.AssignLoadouts`, then `Game.OpenVoting`, then starts that round's voting-window timer.
- `CastVote` — records the vote; if `Game.VotingComplete()` goes true (every connected human has
  voted), closes the window immediately instead of waiting out the timer.
- `CloseVotingWindow` — what the timer calls on expiry; also directly invocable. Shared private
  helper `closeVoting` (assumes the lock is already held) implements the full tally/tie/tiebreak
  contract:
  - clear winner → the round is already resolved by `Game.CloseVoting` itself; if the mode isn't
    finished, `beginRound` runs again for the next round.
  - **first** tie → `Game.CloseVoting` already flips state to `TIEBREAK` — `closeVoting` just
    starts a fresh window for the revote and returns.
  - **second** tie (the revote also tied, including zero votes) → `closeVoting` distinguishes this
    from the first tie by checking whether the state was already `TIEBREAK` *before* calling
    `Game.CloseVoting` (both ties otherwise look identical: `tied=true`, state stays `TIEBREAK`).
    Only then does it call `ports.ITiebreaker.Break(ctx, []string)`, converting `game.OptionID` to
    `string` and back exactly as the port's doc comment prescribes, then
    `Game.ResolveTiebreak(winner)`.
- `Disconnect` / `Reconnect` — presence is explicit API, no heartbeats in this tanda. A disconnect
  that leaves `VotingComplete()` true triggers the same early-close path as a vote would (a
  disconnected participant counts as a null vote, per [[gameplay-game-modes]]).
- `GetGame` / `GetGameByCode` / `GameCode` — reads, also taking the per-Game lock briefly since the
  store hands back a live, mutable `*game.Game` pointer (same trade-off every other service in this
  codebase makes by returning domain entities directly).
- `finalizeLocked` (private) — called automatically by `withGame` whenever a mutation leaves the
  Game `FINISHED`/`ABORTED`: computes `Game.Result()`, calls `ports.IGameHistory.Record` **only if
  non-nil** (best-effort, logs on error, never fails the request — there's no adapter for this port
  yet), deletes the Game from the store, cancels its timer, and frees its per-Game mutex.

## Voting timer (`clock.go`)

`Clock`/`Timer` wrap `time.AfterFunc` behind an interface purely so tests can drive the 30s window
(and its single revote window) deterministically via a fake clock's `Advance(d)`, instead of racing
a real timer — the same reason `PictureWorker` exposes `RunOnce` alongside `Start`. `VotingPolicy{
Window time.Duration }` carries the configured duration in; `config.GameVotingWindow` (default 30s,
`GAME_VOTING_WINDOW` env var) feeds it from `main.go`.

## Events (`game_event_hub.go`)

`GameEventHub` is `services.PictureEventHub`'s pattern ([[picture-events-sse]]) generalized to be
keyed **per `GameID`** (`map[GameID]map[chan GameEvent]struct{}`) instead of one flat subscriber
set — a slow/stuck subscriber on one lobby can never affect another's. Non-blocking `Publish`,
drops for a full subscriber (buffer 8) exactly like the picture hub. `withGame` drains
`Game.PullEvents()` and republishes every one after each mutation. No transport (WS/SSE) reads from
it yet.

## `ports.IGameStore` (`internal/domain/ports/game_store.go`)

New port: `Create/Get/GetByCode/Code/Save/Delete/DeleteExpired`, indexed by both `GameID` and join
code. **Deliberately has no snapshot/rehydration API** — the only adapter today
(`infrastructure/gamestore.MemoryGameStore`) is a plain in-process map, so none is needed. A
Redis-backed adapter behind this *same port* is the explicit next step (per the owner) —
`Game`'s fields are all private with no exported (de)serialization, so that adapter will need to
either add one or reconstruct via re-running events; not designed yet, on purpose.
`infrastructure/gamestore.Reaper` sweeps Games untouched for `config.GameLobbyTTL` (default 2h)
every `GameLobbyReapInterval` (default 10m), mirroring `StorageReconciler`'s `Start`/`RunOnce`
split as `Start`/`ReapOnce`.

## Cheap adapters (`internal/infrastructure/game/`)

Only the ports that need **zero migration** got a real adapter this tanda:

- `CoinFlipTiebreaker` — `ports.ITiebreaker` via a uniform random pick. Swappable for an LLM-backed
  adapter later without touching `GameService`.
- `DefaultWeights` — `ports.IAssignmentWeights` that always returns `game.DefaultAssignmentWeights()`.
- `RepoPowerPool` — `ports.IGamePowerPool` wrapping the existing `IStandRepository.GetAll`/
  `IDevilFruitRepository.GetAll` at a fixed `enums.EnGB` (names, not localized descriptions, are
  all Loadout assignment needs).
- `StaticStageCatalog` — `ports.IStageCatalog` **hardcoded** with JoJo's 8 parts and 11 One Piece
  sagas (same names as [[gameplay-game-modes]]'s rulebook). Explicit `TODO`-in-code stopgap: there
  is still no schema/CRUD for stage content (deliberately out of scope, per the owner) — this exists
  purely so the application layer is runnable end-to-end.

`ports.IGameHistory` and `ports.IInventory` still have **no adapter** — `GameService`'s `history`
dependency is wired as `nil` in `main.go`, which `finalizeLocked` tolerates by design.

## Error mapping / i18n

`endpoints/error_codes.go` and `endpoints/errors.go` (still in lockstep, per that file's own
comment) gained every new domain/application sentinel (`GAME_NOT_FOUND`, `NOT_HOST`, `GAME_FULL`,
`VOTING_CLOSED`, `INVENTORY_NOT_SUPPORTED`, ...) even though no route exists to trigger them yet —
done now so the mapping isn't forgotten once routes land. Same codes, translated, added to all three
locales (`en-GB`/`es-ES`/`ca-ES`) under `"errors"`.

## Tests

`game_service_test.go` (`package services_test`), hand-written fakes per the repo convention
(`fakeGameStore`, `fakeStageCatalog`, `fakeGamePowerPool`, `fakeAssignmentWeights`,
`fakeTiebreaker`, `fakeGameHistory`, `fakeRandom`, `fakeClock`/`fakeTimer`; reuses
`auth_service_test.go`'s existing `fakeUserRepository` rather than redeclaring it — Go's
one-package-per-`_test.go`-set rule caught that collision immediately). Covers creation, membership,
the full Gauntlet defeat/victory paths, a full 3-round Versus match, the tie → revote → tiebreaker
sequence (including asserting the `RoundResolved.DecidedByCoinFlip` event actually publishes), timer
expiry, host reassignment, last-human abort + finalize (with and without a history adapter), and a
concurrent-`CastVote` test for the per-Game locking. `game_event_hub_test.go` covers per-Game
isolation and the non-blocking-drop behavior. `-race` couldn't be run in this environment (no `gcc`
for cgo on this Windows box) — plain `go test ./internal/...` is green; re-run with `-race` wherever
cgo is available before trusting the concurrency test fully.

## Still not built (the actual next tanda)

- Redis-backed `IGameStore` adapter (behind the existing port).
- Websocket transport + HTTP routes/DTOs for every `GameService` method — nothing in
  `infrastructure/api` calls into this layer yet.
- Stage catalog schema/migration/admin CRUD (retiring `StaticStageCatalog`).
- `IGameHistory` and `IInventory` adapters.

Related: [[gameplay-game-modes]], [[gameplay-domain-design]], [[ADR]], [[backend-contract]],
[[picture-events-sse]].
