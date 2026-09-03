---
title: "Every timed WS frame now stamps an authoritative closesAt"
tags:
  - project
  - jojo-onepiece-simulator
  - backend
  - frontend
  - gameplay
---

# Every timed WS frame now stamps an authoritative closesAt (2026-09-03)

Closes the last piece of the `closesAt` drift follow-up from [[game-lobby-todo]] §6 /
[[game-vote-buttons-2026-08-26]]: `VOTING_OPENED`/`TIEBREAK_OPENED`/`SUMMARY_OPENED` already got a
server-stamped deadline on 2026-09-01 ([[game-final-result-2026-09-01]]). The same bug still lived
in the two remaining timed frames:

- **`LOADOUTS_ASSIGNED`** sent `revealMs` (a *duration*). The client derived
  `revealEndsAt = Date.now() + revealMs`, so the deadline drifted by however long hub delivery
  took - and in the bad direction: the sorteo overlay could finish *after* voting had actually
  opened.
- **`ROUND_RESOLVED`** carried no deadline at all. `resultEndsAt` only ever arrived on the STATE
  frame that followed, forcing the frame handler to null it out and wait.

## The fix

`GameService.publish` now stamps `GameEvent.ClosesAt` for these two events too, straight from
`s.revealEnds`/`s.resultEnds` - the exact same maps `RevealEndsAt`/`ResultEndsAt` already serve to
a reconnecting client via the STATE snapshot. All **five** timed frames
(`VOTING_OPENED`/`TIEBREAK_OPENED`/`SUMMARY_OPENED`/`LOADOUTS_ASSIGNED`/`ROUND_RESOLVED`) now go
through one pattern: arm the phase timer inside the same `withGame` closure, *before* `withGame`
calls `publish` - see `publish`'s own doc comment in `game_service.go` for the full precondition
and the current list of relied-on call sites. Any future timed frame must follow the same ordering
or its deadline silently falls back to `frameDeadline`'s `time.Now()+window` synthesis.

`LoadoutsAssignedPayload.RevealMs` was replaced outright by `ClosesAt` (no compat field - single
frontend, deployed with the backend). `RoundResolvedPayload` gained `ClosesAt`. The internal,
never-serialized `GameEvent.RevealMs` field was renamed to `RevealWindow` to match (it's now purely
the zero-`closesAt` fallback window, same role as `VotingWindow`/`SummaryWindow`).

### Why the reveal didn't need a transported duration

`useLoadoutReveal` never used `revealMs` to lay out the sorteo's own timing - it computes its whole
phase timeline locally via `revealTimeline(...)`, a deterministic mirror of `reveal.go` (same
inputs both sides already have: mangas, each participant's landed loadout, `RevealSpeed`, gameId,
roundIndex). `revealMs` fed exactly one thing: a scale factor,
`scale = serverRevealMs / localTotalMs`, so a constants drift between the two sides degrades pacing
rather than desyncing "reveal looks done" from "voting is actually open".

Scaling against `(closesAt - Date.now()) / localTotalMs` instead is strictly better with the same
mechanism: the animation lands exactly on the server's real deadline, and a reconnect mid-reveal
compresses whatever's left instead of overshooting it. The deterministic-timeline invariant itself
- `revealTimeline`/`revealSpinCycles`/`RevealFxMaxMs` (see [[gameplay-power-fx]]) - is untouched.

### Accepted trade-off

`frameDeadline` still renders with `time.RFC3339` (second precision), so a stamped deadline can
land up to 999ms early. That's the safe direction (client finishes slightly *before* the server's
deadline, not after) and matches how `votingEndsAt`/`summaryEndsAt`/`resultEndsAt` already quantize
- format left alone rather than moving to `RFC3339Nano`.

`live.resultEndsAt` in the frontend socket store has no UI consumer yet - the `ROUND_RESOLVED` half
of this fix is a contract/state-correctness change with no visible behavior delta today.

**Closed (2026-09-03, later same day)**: `RoundResultPanel` now renders a `game.match.result.nextIn`
countdown ("Next round in Xs") off `resultEndsAt`, mirroring `VotingStatusBar`'s existing
`revealEndsAt` → `votingIn` pattern - same `useNow`/`secondsUntil` machinery, purely informational
like the rest of this feature's countdowns (server alone decides when RESOLVING ends; "skip" still
only hides the panel locally). Only wired for `variant: 'result'` - `variant: 'tie'` has no result
timer and always passes `resultEndsAt={null}`. New key added to all three locales. `tsc --noEmit`
and the touched jest suites (`match-rules`, `match-screen`) green.

## Verification

Backend: new `TestPublish_LoadoutsAssigned_CarriesRevealDeadline`/
`TestPublish_RoundResolved_CarriesResultDeadline` (`game_service_frame_closes_at_test.go`) assert
the published `GameEvent.ClosesAt` equals what `RevealEndsAt`/`ResultEndsAt` serve a reconnecting
client - the property a transport-level test alone can't catch (it can't see whether the accessor
and the stamped frame ever disagree). `TestBuildEventFrame_TimedFrames_UseStampedClosesAt` extended
to all five frames. `go build`/`go vet`/`go test ./...` clean via Docker (the two Redis-backed
packages, `gamestore/redis` and `streamticket/redis`, failed for lack of a running Redis - expected,
unrelated to this diff, not part of it).

Frontend: `tsc --noEmit` clean, `pnpm lint` 0 errors. `pnpm test:ci` full-parallelism run showed the
already-documented Docker flake pattern
([[norma-verificacion-docker]]) in `use-loadout-reveal.test.tsx`/`tooltip.test.tsx`/
`use-drop-zones.test.tsx` plus four unrelated "Super expression must either be null or a function"
suite failures; re-running the same set with `--maxWorkers=2` (touched files plus every suspect
suite) came back 19/19 suites, 199/199 tests green, confirming flake rather than regression before
committing.

No live browser walkthrough for this tanda (owner decision) - `live.resultEndsAt`'s eventual UI
consumer, if the owner wants one, is the natural place to actually exercise this live.

Related: [[game-lobby-todo]], [[game-vote-buttons-2026-08-26]], [[game-final-result-2026-09-01]],
[[game-match-assignment-frontend]], [[game-realtime-transport]], [[gameplay-power-fx]].
