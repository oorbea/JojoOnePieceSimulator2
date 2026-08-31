---
name: game_round_result_2026-08-28
description: "Round-resolved vote-tally feature shipped 2026-08-28 (per-option counts + voter avatars, inline)"
metadata: 
  node_type: memory
  type: project
  originSessionId: 7dcd30a1-92c6-4f78-9c0b-584d8c0501a5
  modified: 2026-08-28T09:39:02.105Z
---

Shipped "show vote counts when a round ends" (backend + frontend), closing the "Round-resolved
feedback" bullet open since [[game_lobby_feature_closed_2026-08-13]]/sorteo_roster_redesign era.
Full writeup in the repo's own `ObsidianVault/game-round-result-2026-08-28.md`.

**Product decisions (owner, delegated technically)**: per-option count + nominal breakdown
(avatars only, name on hover), inline replacing VoteBar in the same slot, auto + skip button
(skip is client-only — server alone decides when RESOLVING ends), tied vote breakdown shown before
revote. Final-game result screen explicitly out of scope.

**Turned out not to be frontend-only**: `RESOLVING` was a same-call pass-through (`resolveRound`
advanced state in the same method) — split into `resolveRound` + new `Game.CompleteRound()`, held
apart by `GameService.scheduleResultDelay` (6s fixed, `game.ResultDuration`), mirroring the sorteo's
`scheduleRevealDelay` pattern exactly. Also added `Round.TiedVotes` since a tie's votes used to be
destroyed by `Ballot.Reset()` with nothing kept — now captured and revealed before the revote.

**Verified**: backend `go build`/`vet`/`test` green in Docker (existing voting tests needed
`CompleteRound()`/`advanceResult(deps)` calls added since closing a vote no longer lands on the next
state in the same call); frontend `typecheck`/`lint`/`test:ci` green in Docker, 445/445 (one
`tooltip.test.tsx` failure under full worker parallelism was the known Docker flake, confirmed by
re-running at `--maxWorkers=2` — see [[feedback_backend_tests_via_docker]]).

**Not yet done**: live two-browser `claude-in-chrome` walkthrough and a real `local-up` playthrough —
flagged as follow-up, not skipped silently.
