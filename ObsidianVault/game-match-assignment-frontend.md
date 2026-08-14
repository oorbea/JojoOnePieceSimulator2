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

## Reveal never actually played, and only revealed one card at a time (fixed 2026-08-14)

The owner tried a real Gauntlet match: no reveal animation ever played (the countdown just ran with
every card face-down), and pressing Skip dumped every loadout at once - matching exactly what
"broken but only escapable via Skip" looks like.

**Root cause**: `use-loadout-reveal.ts`'s scheduling effect depended on `[active, orderKey]` and
returned `clearTimers` as its cleanup. But the very first thing that effect does is call
`markRevealed()`, which bumps the socket store's `revealedAssignmentSeq` up to `assignmentSeq` -
which is exactly what `shouldReveal()` (`loadout-reveal.ts`) gates on. So on the **very next
render**, `active` flips `true → false`, React runs the previous effect's cleanup (`clearTimers`,
since the dependency changed), and every just-scheduled `setTimeout` is cancelled before the first
one could ever fire. `revealedCount` stayed at 0 forever; `revealing` stayed `true` forever (nothing
ever set it `false`), which is why Skip kept showing. Skip itself was never broken - it doesn't go
through this effect, so it always could (and did) reveal everything at once.

**Fix**: the scheduling effect no longer returns a cleanup tied to `[active, stepsKey]` changing.
Pending timers are only cleared (a) explicitly, right before a **new** sequence schedules its own
timers (guarded by the existing `startedRef` re-entry check), and (b) on unmount, via a second
effect with an empty dependency array. A new test
(`hooks/__tests__/use-loadout-reveal.test.tsx`) reproduces the exact failure shape with a harness
whose own `markRevealed()` flips `active` on the same tick, asserting the steps still fire - this is
the only way to catch a regression here, since the "correct" code looks more suspicious (an effect
with no `active`-keyed cleanup) than the buggy version did.

**Lesson for any future effect that calls a callback capable of invalidating its own trigger
condition**: don't return a cleanup tied to that same dependency. Clear stale side effects
explicitly at the point a genuinely new run starts, not via the automatic cleanup-on-dependency-
change machinery - that machinery fires on the transition the callback itself just caused.

## Poder-a-poder reveal + Physical Form drawn first (2026-08-14, same pass)

Two more owner reports from the same playtest, fixed together with the bug above:

- **Reveal was one participant-card at a time (550ms each), never poder a poder.** Now each card's
  loadout slots reveal individually after the card flips, in the exact order
  `LoadoutBuilder.Build` draws them (see [[gameplay-domain-design]]'s Template Method entry for the
  new draw order). `lib/loadout-reveal.ts`'s `revealSteps(snapshot, selfParticipantId)` flattens
  `revealOrder` + per-participant `match-rules.ts`'s `loadoutSlots` into one global
  `{ participantId, slot }` timeline (`slot: -1` = the card-flip step); `use-loadout-reveal.ts` walks
  it with cumulative `flipStepMs`/`slotStepMs` delays (450ms/220ms, zeroed under reduced motion) and
  returns `visibleSlotsById` alongside the existing `revealedIds`. `loadout-card.tsx` now builds its
  own render rows from `loadoutSlots`, grouping consecutive scalar-chip slots into one flex-wrap row
  while the Stand/DevilFruit art blocks stay their own full-width row - each block reserves its usual
  fixed height (`110`/`90`) even before its slot is revealed, so a later slot popping in doesn't jump
  the layout.
- **"Only got a stand, physical form should come first."** Two independent causes: (a) the backend
  drew Physical Form *last* - fixed, see [[gameplay-domain-design]]; (b) the card showed
  Physical Form/Haki chips pinned at their `PRIVATE` floor even in a JoJo-only lobby (they have no
  `NONE` member), which reads as "the game gave me nothing" instead of "this manga isn't in play".
  `match-rules.ts`'s `loadoutTraits` was replaced by `loadoutSlots(loadout, mangas)`, which gates
  every One Piece slot (physicalForm/devilFruit/fruitMastery/haki×3) on `mangas.includes('ONE_PIECE')`
  and every JoJo slot (stand/hamon/spin) on `mangas.includes('JOJO')` - a single-manga lobby's card
  now only ever shows slots that manga can actually produce.

## Lobby manga selector moved out of "Lobby settings" (2026-08-14, same playtest)

The owner also couldn't tell which manga(s) a lobby was playing without opening the collapsed
config panel. The manga toggle (`MangaRow`, `components/presentational/manga-row.tsx`) moved to the
main lobby screen, always visible, with a check-mark on the active chip (color alone wasn't a legible
enough affordance) - removed entirely from `lobby-config-panel.tsx` rather than duplicated. Since
every other field in that panel needs an explicit "Save", but this one has no separate save step, it
autosaves: `lobby-room-container.tsx`'s `handleToggleConfigManga` computes the next `mangas` array
and calls a new shared `submitConfigForm(next)` immediately, passing `next` explicitly rather than
reading `form` state right after `setConfigForm` (which would still read the pre-toggle value, since
React batches the state update). Same playtest also moved `NumberStepper` to
`SettingRow`'s new `stacked` prop (label always above the control, not flipping to a wide row at
`$md` inside the panel's narrow `flexBasis:320` columns, which is what put the "max players" label
and its stepper at opposite ends of the row).

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
