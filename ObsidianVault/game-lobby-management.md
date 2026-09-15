---
title: "Feature: lobby management, host powers, public browser (backend, 2026-08-12)"
tags:
  - project
  - jojo-onepiece-simulator
  - backend
  - feature
  - gameplay
---

# Lobby management, host powers, public browser (2026-08-12)

## Status

**See [[game-lobby-todo]] for the actionable remaining-work checklist** (backend endpoint tests,
frontend config-edit UI, pool-filter UI). This note is background/rationale only.

Backend done and tested (unit + real-Redis for the new public index, via a throwaway
`redis:8-alpine` container). Frontend UI (create/join/browse/lobby-room screens, WebSocket client)
is **not built yet** - next tanda. See [[gameplay-game-modes]] and [[game-realtime-transport]] for
what already existed before this pass.

## What this tanda added

- **Config is now editable while `LOBBY`**, not just at creation: `game.Config` gained
  `Visibility` (`PUBLIC|PRIVATE`), `VotingWindowSeconds` (5-180, per-lobby, replaces the global
  default as the per-round timer source), and `PoolFilter` (rarity/fruit-type allowlists + an
  explicit `PowerID` banlist, both optional and empty-means-unrestricted).
  `Game.Reconfigure(caller, cfg, newTeams, stages)` replaces the whole `Config` atomically
  (all-or-nothing - checked before any mutation), re-seating players when the mode itself changes
  (Gauntlet↔Versus): Gauntlet→Versus alternates join order across the two new teams, Versus→Gauntlet
  merges everyone onto the one team. Bots are auto-dropped if the new mode is Gauntlet or no longer
  allows them. Shrinking `TeamSize` below the seated human count is rejected
  (`ErrConfigWouldEvictPlayers`) rather than evicting anyone.
- **Versus Start gate relaxed**: teams must be **equal and non-empty**, not full to
  `Config.TeamSize()` (that stays the *capacity* ceiling enforced on join/switch). Empty team → 0
  players is `ErrNotEnoughPlayers`; unequal non-empty teams is `ErrTeamSizeMismatch`.
- **Real bug fixed while relaxing that gate**: `checkAbortConditions`'s "empty Versus team aborts"
  branch used to fire unconditionally, including in `LOBBY` - so leaving a team empty *before*
  starting (which `SwitchTeam` makes trivially reachable) would have killed the lobby. Now gated on
  `g.state != enums.Lobby`; only a mid-match empty team aborts.
- **Team switching**: `Game.SwitchTeam(caller, target, teamID)` - self-service for anyone, host-only
  to move someone else. `Participant.setTeam` is a new unexported mutator (teamID was previously
  private with no setter at all).
- **Host powers**: `Kick` (host-only, LOBBY-only, can't kick self, emits `PlayerKicked` *before* the
  underlying `PlayerLeft` so the transport can close the victim's socket deterministically instead of
  letting it dangle), `TransferHost` (host-only, target must be a connected human - reuses the
  existing `HostReassigned` event, no new wire type), `SetLocked` (host-only, LOBBY-only; `Join`
  checks it, `AddBot` does not - the host adding bots to a locked lobby is intentional).
- **Public lobby browser**: `ports.IGameStore.ListPublic(ctx, limit)` returns `[]*game.Game`.
  `game.Game.IsPubliclyJoinable()` (`state==LOBBY && visibility==PUBLIC && !locked`) is the single
  predicate both adapters key off - no separate "mark public" call, `Create`/`Save` just check the
  `*game.Game` they're already given. Memory adapter: plain map scan (no SCAN ban there - that's a
  Redis-specific constraint). Redis adapter: a `jojo:game:public` ZSET, member = GameID, score = this
  write's absolute TTL-expiry (`now+ttl` in ms) - so pruning stale members is `ZREMRANGEBYSCORE key
  -inf now`, done on every read and on every `Create`/`Save`. `createScript`/`saveScript`/
  `deleteScript` all extended with the extra KEYS/ARGV to keep the index in lockstep with the payload
  write, still one EVAL round trip each. Verified against a real Redis container, not just the
  in-memory fake.
- New endpoints: `GET /games/public?limit=` and `GET /games/preview?code=` (both roster-free,
  loadout-free, and - unlike `GET /games/by-code/{code}` - reachable by a non-participant;
  `preview` also works for `PRIVATE` lobbies since the code itself is the credential),
  `POST /games/{id}/join` (public-browser join path, 403 `LOBBY_PRIVATE` on anything not `PUBLIC`),
  `PATCH /games/{id}/config` (host-only config edit).
- New WS commands: `SWITCH_TEAM`, `MOVE_PLAYER` (host moving someone else - same effect as
  `SWITCH_TEAM` with an explicit target, kept as a separate command so a client never omits a field
  to mean "myself"), `KICK`, `TRANSFER_HOST`, `SET_LOCK`, `UPDATE_CONFIG` (full-replacement body, same
  shape as `POST /games`). New frames: `TEAM_CHANGED`, `PLAYER_KICKED`, `LOBBY_LOCK_CHANGED`,
  `CONFIG_UPDATED`. `wsReadLimit` raised 4096→16384 - `UPDATE_CONFIG` can carry a sizeable banned-
  PowerID array. The kicked participant's own socket is now explicitly closed
  (`StatusNormalClosure`) right after their `PLAYER_KICKED` frame, instead of resending `STATE` and
  leaving the socket to fail on its next `RESYNC`.
- **Per-lobby voting window fully wired through**: `services.GameEvent` gained a `VotingWindow
  time.Duration` field (set in `publish` from `g.Config()`), so `VOTING_OPENED`/`TIEBREAK_OPENED`'s
  `closesAt` reflects each lobby's own configured window, not the one global default.
  `GameWSConfig.VotingWindow` is now only a fallback for a zero value.
- **Snapshot versioning decision**: deliberately did **not** bump `snapshotVersion` (still 1).
  Every new field is additive (`omitempty` on the wire, safe default in `game.Restore`: absent
  `Visibility`→`PRIVATE`, absent/zero `VotingWindowSeconds`→`game.DefaultVotingWindowSeconds` (30),
  absent `PoolFilter`→unrestricted, absent `Locked`→`false`), so a lobby saved by the previous build
  decodes cleanly instead of dying with a version-mismatch hard error mid-TTL. Bump to 2 only if a
  future field has no safe default.

## Deliberately deferred to the frontend tanda

No frontend WebSocket client, no lobby UI at all yet (create/join/browse/room screens, host config
panel, team-switch UI, kick/transfer/lock controls) - see [[frontend-stack]] and
[[frontend-responsive-frutiger-aero]] for the design system it must compose from, and
[[norma-diseno-ui-ux]] for the mandatory skills-before-JSX rule. `apps/frontend`'s
`EXPO_PUBLIC_SOCKET_URL` env var already exists (unset in `.env.example`) from an earlier pass.

Also not done: a dedicated HTTP-endpoint test file for the game feature (`game_endpoints_test.go`,
`game_ws_endpoints_test.go` don't exist at all yet, for *any* of the game routes, not just the new
ones) - building one needs a fairly large local fake set (`IUserRepository`, `IStageRepository`,
`IGamePowerPool`, etc., mirroring what `game_service_test.go` already has in the `services_test`
package but nothing reusable exists in `endpoints_test`). Coverage today is domain-level (new file
`lobby_management_test.go`, `pool_filter_test.go`) + application-service-level (extended
`game_service_test.go`) + a real-Redis-verified `ListPublic` (extended `store_test.go`) - all
passing, just not exercised through the actual HTTP/WS handlers end-to-end.

## 2026-09-15: kicked-frame delivery race, and join-code regeneration

`Kick`'s `PLAYER_KICKED` frame to the victim was silently lost almost every time: `forwardEvents`
enqueued it on the buffered `outbound` channel (drained asynchronously by `writePump`) and then
immediately called `conn.Close(StatusNormalClosure)` itself - the close almost always won the race,
so `writePump` never got to write the frame it had just been handed. The existing test only
asserted the frame was *enqueued*, never that it was actually written before the close, so this
shipped clean. Fixed by making "write this, then close" one atomic item (`outMsg{data, closeCode,
closeText}`) on the channel, so `writePump` itself performs the close right after the write -
**never call `conn.Close` from a goroutine other than the one draining `outbound`, and never
right after (not through) an enqueue you expect to be delivered first.** The regression test now
runs `writePump` for real and asserts the frame is actually read off the wire before the close.

Frontend-side, the victim missing that frame (or an expired lobby disconnect-grace's `Leave`
happening while they're away) used to reconnect straight into a 403/404 mint rejection that only
set `status: 'closed'` with no `terminal` - an orphaned screen with a stale roster and every action
rejected. Added a `REMOVED` terminal (distinct from `KICKED`, same toast-and-redirect-to-`/play`),
guarded so it never overwrites an already-set `FINISHED`/`ABORTED` terminal when that game's short
TTL later 404s the same mint call.

Added a host-only, LOBBY-only `REGENERATE_CODE` command (mirrors `SET_LOCK`'s shape exactly - see
that command's nine-step path in this file's own history as the template) so a kicked player who
remembers the old code can't just rejoin with it. Key design point: **the join code is not part of
the `Game` aggregate** - it lives only in the store's side index (`byCode` map / Redis
`jojo:game:code:*`+`codeof:*` keys), so `Game.Snapshot`/`Restore` and `redis/wire.go` are
untouched, and the three-place rule below doesn't apply to it. `ports.IGameStore.SetCode` claims
the new code and releases the old one atomically; the Redis Lua script reads the game's own
`PTTL` (not the store's configured lobby TTL) so rotating the code on a near-expiry/terminal game
never resurrects it for a full lobby lifetime. Rotating only ever locks out a **private** lobby -
a public one is still joinable via the browser (`JoinByID`), so the frontend confirm sheet says so.

## Gotchas hit

- `checkPoolSufficiency` (new, in `GameService.StartGame`, runs *before* `g.Start` so a rejected
  start leaves the Game in `LOBBY` instead of stranding it in `ASSIGNING` with no way back) must
  compare the filtered pool against **actual current team occupancy** (`max(t.Size())` over
  `g.Teams()`), not `Config.TeamSize()`'s capacity ceiling - the first version compared against
  capacity and broke every existing test whose lobby never filled up to its configured max.
- Redis Lua scripts pass the "is this game publicly joinable right now" flag as a plain `ARGV`
  string (`"1"`/`"0"`) computed in Go via `g.IsPubliclyJoinable()` right before the script runs -
  cheaper and simpler than a second round trip or a `SetPublic` port method, and it's the same
  `*game.Game` `Create`/`Save` already received.

Related: [[gameplay-game-modes]], [[gameplay-application-layer]], [[game-lobby-persistence]],
[[game-realtime-transport]], [[zettelkasten-workflow]].
