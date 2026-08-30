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
  `Game.AssignLoadouts`, then... **changed 2026-08-14** (see [[game-match-assignment-frontend]]'s
  sorteo redesign): only when loadouts were actually (re)assigned this call, `beginRound` now defers
  `Game.OpenVoting` behind `scheduleRevealDelay` instead of calling it inline. That method computes
  `revealDurationFor(g)` (`GameService`'s own helper, wrapping `game.RevealDuration`) — records the
  deadline in `s.revealEnds[gameID]` (read by the new `RevealEndsAt` accessor, served to a
  (re)connecting client as `GameStateResponse.RevealEndsAt` so it can resume the countdown instead
  of restarting from zero), and schedules a `Clock.AfterFunc` timer that, on firing, calls
  `openVotingAfterReveal` — which re-acquires the lock via `withGame` and *only then* does what
  `beginRound` used to do inline: `Game.OpenVoting` + `scheduleVotingTimer`. Rounds that don't
  reassign (Gauntlet after round 1) skip the delay entirely and call `OpenVoting` immediately, same
  as before — there's nothing new to reveal. **Amended 2026-08-30 (owner request, see
  [[game-match-assignment-frontend]]'s dated section)**: `game.RevealDuration` stopped being a pure
  function of `mangas` alone — it now also takes each participant's own landed
  Stand/DevilFruit and the lobby's `RevealSpeed`, reproducing V1's own per-participant tempo
  instead of one flat per-slot hold for everyone. Both server and client still arrive at the
  identical number without exchanging anything beyond what each already has (the snapshot's own
  `mangas`/`participants[*].loadout`/`config.revealSpeed`) — the "no shared code, same number"
  property survives, only the inputs it's a function of grew.
  **Consequence worth knowing**: the Game now sits in `ASSIGNING` with **zero `Rounds()`** for the
  whole reveal delay, since a `Round` is only created inside `OpenVoting` itself. Any code that
  assumed "loadouts assigned ⇒ a Round already exists" (the frontend's old `shouldReveal` did) breaks
  under this design — see [[game-match-assignment-frontend]] for the bug that caused and its fix.
  The reveal-delay timer shares `s.timers`/`cancelTimer` with the voting-window timer (they never
  coexist for the same `GameID`: the reveal timer fires, opens voting, and only *then* does the
  voting timer take its place) — so `AbortGame`/`finalizeLocked` cancelling during the reveal already
  works with no changes. `s.revealEnds` is process memory only, same known gap as `s.timers` itself:
  a backend restart mid-reveal loses the timer and the game gets stuck in `ASSIGNING` — not fixed
  this tanda, same category as the pre-existing voting-timer gap.
- `CastVote` — records the vote; if `Game.VotingComplete()` goes true (every connected human has
  voted), closes the window immediately instead of waiting out the timer.
- `CloseVotingWindow` — what the timer calls on expiry; also directly invocable. Shared private
  helper `closeVoting` (assumes the lock is already held) implements the full tally/tie/tiebreak
  contract:
  - clear winner → the round is already resolved by `Game.CloseVoting` itself, leaving state
    `RESOLVING`; the mode's outcome is **not** applied yet (see below).
  - **first** tie → `Game.CloseVoting` already flips state to `TIEBREAK` — `closeVoting` just
    starts a fresh window for the revote and returns.
  - **second** tie (the revote also tied, including zero votes) → `closeVoting` distinguishes this
    from the first tie by checking whether the state was already `TIEBREAK` *before* calling
    `Game.CloseVoting` (both ties otherwise look identical: `tied=true`, state stays `TIEBREAK`).
    Only then does it call `ports.ITiebreaker.Break(ctx, []string)`, converting `game.OptionID` to
    `string` and back exactly as the port's doc comment prescribes, then
    `Game.ResolveTiebreak(winner)` — which also leaves state `RESOLVING`, same as the clear-winner
    path.
  - **2026-08-28** (see [[game-round-result-2026-08-28]]): whenever `closeVoting` ends up with state
    `RESOLVING`, it no longer calls `beginRound` in the same breath. Instead it calls
    `scheduleResultDelay` — the exact structural twin of `scheduleRevealDelay`: holds the Game in
    `RESOLVING` for `game.ResultDuration` (6s, fixed), records the deadline in a new `s.resultEnds`
    map (`ResultEndsAt(id)` accessor, mirroring `RevealEndsAt`/`VotingEndsAt`), and the timer's
    callback `completeRoundAfterResult` re-acquires the lock, calls the new `Game.CompleteRound()`
    (applies `mode.ApplyRoundResult`, moves to `FINISHED`/`ASSIGNING`), and only then runs
    `beginRound` if there's a next round. This is what finally gives clients an observable window to
    render the round's outcome — before this, `resolveRound` advanced clear past `RESOLVING` in the
    same call and no client ever saw it.
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

**2026-08-26**: `scheduleVotingTimer` now also records the window's absolute deadline in
`s.votingEnds[gameID]`, read by a new `VotingEndsAt(id) (time.Time, bool)` accessor — an exact
mirror of `RevealEndsAt`/`s.revealEnds` above, serving `GameSnapshotResponse.VotingEndsAt` to a
(re)connecting client so it resumes the real countdown instead of a dead bar. `cancelTimer` clears
both `revealEnds` and `votingEnds` in the same place, so every teardown path (`closeVoting`,
`finalizeLocked`) already covers it with no extra code. See [[game-vote-buttons-2026-08-26]].

## Events (`game_event_hub.go`)

`GameEventHub` is `services.PictureEventHub`'s pattern ([[picture-events-sse]]) generalized to be
keyed **per `GameID`** (`map[GameID]map[chan GameEvent]struct{}`) instead of one flat subscriber
set — a slow/stuck subscriber on one lobby can never affect another's. Non-blocking `Publish`,
drops for a full subscriber (buffer 8) exactly like the picture hub. `withGame` drains
`Game.PullEvents()` and republishes every one after each mutation. No transport (WS/SSE) reads from
it yet.

## `ports.IGameStore` (`internal/domain/ports/game_store.go`)

Port unchanged: `Create/Get/GetByCode/Code/Save/Delete/DeleteExpired`, indexed by both `GameID` and
join code. **2026-08-11**: `infrastructure/gamestore.MemoryGameStore` is now joined by a
Redis-backed `infrastructure/gamestore/redis.Store`, selected in `main.go` whenever `REDIS_URL` is
set. See [[game-lobby-persistence]] for the snapshot seam (`entities/game.Snapshot`/`Restore`) that
made that possible, the key scheme, and the fail-closed contract. `infrastructure/gamestore.Reaper`
still sweeps Games untouched for `config.GameLobbyTTL` (default 2h) every `GameLobbyReapInterval`
(default 10m) — against Redis it's a harmless no-op, since Redis's own `PX` TTL does the actual
expiry now.

## Adapters (`internal/infrastructure/game/`, `internal/infrastructure/repositories/`)

- `CoinFlipTiebreaker` — `ports.ITiebreaker` via a uniform random pick. Swappable for an LLM-backed
  adapter later without touching `GameService`.
- `DefaultWeights` — `ports.IAssignmentWeights` that always returns `game.DefaultAssignmentWeights()`.
- `RepoPowerPool` — `ports.IGamePowerPool` wrapping the existing `IStandRepository.GetAll`/
  `IDevilFruitRepository.GetAll` at a fixed `enums.EnGB` (names, not localized descriptions, are
  all Loadout assignment needs).
- **2026-08-11**: `repositories.StageRepository` replaces the old hardcoded `StaticStageCatalog`,
  satisfying both `ports.IStageCatalog` (read side, used by `GameService`) and the new
  `ports.IStageRepository` (admin CRUD, see [[game-realtime-transport]]#stages). Backed by the
  `stages` table (migration `00008_stages.sql`), seeded with the same 19 names the old stub had.
- **2026-08-11**: `repositories.GameHistory` implements `ports.IGameHistory` against
  `game_results`/`game_result_participants` (migration `00009_game_history.sql`). `GameResult`
  gained a `Participants []ParticipantOutcome` field (filled by both `IGameMode.Outcome`
  implementations) so history can answer "what did I play", not just "what happened" — see
  [[game-lobby-persistence]].

`ports.IInventory` still has no adapter.

## Error mapping / i18n

`endpoints/error_codes.go` and `endpoints/errors.go` stay in lockstep (per that file's own
comment). **2026-08-11** additions: `STAGE_NOT_FOUND`/`STAGE_ALREADY_EXISTS` for the new stage
catalog, plus three sentinels that existed since this tanda's first pass but had no status/code
branch until now — `EMPTY_TEAM_NAME`, `INVALID_PARTICIPANT_KIND`, `INVALID_SQUAD_VERDICT` (all
fell through to 500 `INTERNAL` before) — and `UNKNOWN_COMMAND` for an unrecognized WebSocket
command. All translated in `en-GB`/`es-ES`/`ca-ES` under `"errors"`, alongside a new
`"enums"."manga"` map for the `JOJO`/`ONE_PIECE` wire codes.

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

## Still not built

**2026-08-11 update**: everything below this line shipped this tanda — Redis-backed `IGameStore`,
the WebSocket + HTTP transport, the stage catalog schema/CRUD, and `IGameHistory`. See
[[game-lobby-persistence]] for the store/persistence half and [[game-realtime-transport]] for the
transport half. What's genuinely still missing:

- `ports.IInventory` — `AbilitySource=INVENTORY` still returns 501 `INVENTORY_NOT_SUPPORTED`.
- Frontend: no WebSocket client yet (`socket.io-client` is installed but unwired), no game UI, no
  admin screens for the new `/stages` CRUD.
- A `Seq` field on `services.GameEvent` so the transport can *detect* a hub drop instead of relying
  entirely on the event+snapshot rule to self-heal it (see [[game-realtime-transport]]).
- `GET /games/preview?code=` (a pre-join preview) and `DELETE /games/{id}/participants/me` (leave
  without a live socket) — both deliberately deferred, see [[game-realtime-transport]].

Related: [[gameplay-game-modes]], [[gameplay-domain-design]], [[ADR]], [[backend-contract]],
[[picture-events-sse]], [[game-lobby-persistence]], [[game-realtime-transport]].
