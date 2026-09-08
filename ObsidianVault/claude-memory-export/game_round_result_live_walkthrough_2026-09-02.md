---
name: game-round-result-live-walkthrough-2026-09-02
description: Live two-browser walkthrough closed the tie/revote/coin-flip + reconnect gap; found and fixed a real Redis-wire bug that a first-pass subagent investigation missed
metadata: 
  node_type: memory
  type: project
  originSessionId: 7d1946d5-2c78-4cf8-aa3a-99ec19417412
  modified: 2026-09-02T09:25:34.350Z
---

Closed the "live two-browser walkthrough" item open since 2026-08-26 ([[game_lobby_feature_closed_2026-08-13]], [[game_round_result_2026-08-28]]) - two real Google accounts, `local-up` Docker stack, tie forced via keyboard hotkeys, reconnect-mid-vote, coin-flip resolve all confirmed live. Full writeup and root cause in the project's own `ObsidianVault/game-round-result-live-walkthrough-2026-09-02.md`.

**Why this matters**: a real bug shipped and was believed verified (per the 2026-08-28 note) but never actually round-tripped through the live Redis store in that session. A subagent's first investigation pass, driving the repro through `GameService.CastVote` directly and through `game.Snapshot()`/`game.Restore()`, found "everything checks out" and added test coverage that still didn't catch it - because none of it touched `apps/backend/internal/infrastructure/gamestore/redis/wire.go`'s separate wire-format structs. Root cause: `wireRound` never got a `TiedVotes` field when `RoundSnapshot.TiedVotes` was added 2026-08-28, so it survived in-memory and even a Redis `Save`, but the next `Get` silently decoded it back to `nil`.

**How to apply**: when this repo adds a field to a domain `*Snapshot` type, grep `infrastructure/gamestore/redis/wire.go` for a matching `wire*` struct and confirm the new field is mirrored there too - the domain snapshot and the Redis wire format are two separate hand-maintained structs, not one. A test that only exercises `Snapshot()`/`Restore()` or an in-memory store does NOT prove a field survives the real Redis-backed `local-up` stack; the diagnostic that actually caught this was a live `GET /api/v1/games/:id` mid-TIEBREAK showing the field entirely absent from the JSON.

Also confirmed as a standing blocker: [[auth_google_only_no_dev_bypass]] (new memory).
