---
title: "Feature: match start + power assignment (frontend, 2026-08-14)"
tags:
  - project
  - jojo-onepiece-simulator
  - frontend
  - feature
  - gameplay
---

# Match start + power assignment (2026-08-14)

Closes the first half of [[game-lobby-todo]]'s §6 (in-match UI). Backend needed **no changes** -
start/assignment/voting-open have been complete and tested since
[[gameplay-application-layer]]/[[game-realtime-transport]]; this tanda was entirely frontend.

## What shipped

- **`EXPO_PUBLIC_SOCKET_URL` finally set** (was blank in every env file since the WS transport
  shipped, silently falling back to REST polling) - see [[docker-setup]] for the full writeup and
  the production repo-variable requirement.
- Real types for `GameLoadoutResponse`/`GameStageResponse` replacing the `Record<string, unknown>`
  placeholder (`types/game.types.ts`), plus the five missing level enums
  (`spinLevel`/`hamonLevel`/`fruitMastery`/`hakiLevel`/`physicalForm`) in `shared/lib/zod.ts`.
- Socket store gained a `live` slice (`assignmentSeq`, `votingRoundIndex`, `votingClosesAt`,
  `tiebreak`) populated from `LOADOUTS_ASSIGNED`/`VOTING_OPENED`/`TIEBREAK_OPENED`/`ROUND_RESOLVED` -
  previously these all fell into the generic bounded feed. `STATE`'s wholesale-replace rule for
  `snapshot` is untouched; nothing else writes to it.
- `lobby-room-screen.tsx` renders a real match view (`components/presentational/match/`:
  `stage-banner`, `loadout-card`, `trait-chips`, `match-roster`, `voting-status-bar`,
  `match-screen`) once `snapshot.state !== 'LOBBY'`, replacing the old bare
  `enums.gameState.*` label. **Same route** (`/play/[id]`) - the container still owns the one
  socket connection, no double `attach()`.
- Sequential animated reveal: `hooks/use-loadout-reveal.ts` chains a 450ms Reanimated Y-axis flip
  per participant (`loadout-card.tsx`), self revealed last (`lib/loadout-reveal.ts`'s
  `revealOrder`), collapsing to an instant reveal under `useReducedMotion()`. A "Skip" button is
  always rendered, not hidden behind a gesture.
- `lobby-rules.ts` gained `teamToneColor` (extracted from `team-column.tsx`'s private `TONE_BG`
  map) so the match roster and the lobby roster share one tone-to-token definition.

## The reveal-timing bug this avoided

`GameService.StartGame` runs start → `beginRound` (assign + `OpenVoting`) inside **one** `withGame`
call, and events publish only after it returns. So a client observes, strictly in order:
`GAME_STARTED` → `STATE` (already `state=VOTING`, loadouts present) → `LOADOUTS_ASSIGNED` → `STATE`
→ `VOTING_OPENED` → `STATE`. Two consequences baked into `lib/loadout-reveal.ts`'s `shouldReveal`:

1. **`ASSIGNING` is never realistically observable by a client** - no UI branches on it, and the
   reveal is never gated on `snapshot.state`.
2. **`LOADOUTS_ASSIGNED` arrives before its own `STATE` resend.** At the instant the frame bumps
   `live.assignmentSeq`, `snapshot` may still be the pre-assignment one. `shouldReveal` gates on
   *both* `assignmentSeq > revealedAssignmentSeq` *and* `hasAllLoadouts(snapshot)` *and* the
   snapshot's current round index matching the assigned one - never on the frame alone. Getting
   this wrong would have shown a reveal animation over stale/empty loadout data on the very first
   assignment of every game.

## Known backend bug worked around, not fixed

`dto.VoteCastPayload.VotesCast` is **always 0** - `game_ws_endpoints.go`'s `buildEventFrame` never
sets it. Harmless for this tanda (no vote UI was in scope), but blocks a live vote-count indicator
later without a small backend fix first (set `VotesCast` from `Ballot.Count()` when building the
frame).

## Locale gap, worth knowing before touching loadout card copy

`RepoPowerPool` resolves stand/devil-fruit data at a **fixed `enums.EnGB`**, unlike a Stage's
`description` (re-resolved per viewer locale server-side via `StageTextResolver`). So
`loadout.stand`/`loadout.devilFruit`'s `description`/`skills` text is always English, not the
viewer's locale - deliberately **not rendered** on `loadout-card.tsx` (only name, picture, rarity,
stat grid, fruit type, all locale-independent or numeric). If localized power text is ever wanted
on this card, look the power up by `id` in `useStands()`/`useDevilFruits()`'s already-viewer-locale
cache instead of trusting the loadout payload - don't patch `RepoPowerPool` for this alone.

## Deliberately cut from this pass (§6's second half, not forgotten)

Vote buttons, live vote counts, tiebreak-specific UI, round-resolved feedback, and the final result
screen (`GAME_FINISHED` still just toasts and routes to `/play`, as before this tanda). See
[[game-lobby-todo]]'s §6 for the reasoning on why that's a separate tanda.

## Verification

`tsc --noEmit` clean, `pnpm run test:ci` green (38 suites / 353 tests - a `StageCard` test flake
around a leaked Tamagui spring animation showed up once in a full run and passed both in isolation
and on immediate re-run; unrelated to this tanda's files, not chased further). New pure-logic tests:
`lib/__tests__/{match-rules,loadout-reveal}.test.ts`, extended `game-socket.store.test.ts`. No new
native render tests (consistent with prior lobby tandas - manual `local-up` walkthrough is the
gate for the animated/visual half).

Related: [[game-lobby-todo]], [[game-lobby-frontend]], [[gameplay-application-layer]],
[[game-realtime-transport]], [[docker-setup]], [[frontend-stack]], [[i18n-multi-language]].
