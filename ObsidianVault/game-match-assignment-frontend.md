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

**Superseded 2026-08-14 - see the "Sorteo redesign" section below.** Originally,
`GameService.StartGame` ran start → `beginRound` (assign + `OpenVoting`) inside **one** `withGame`
call, and events published only after it returned, so a client observed, strictly in order:
`GAME_STARTED` → `STATE` (already `state=VOTING`, loadouts present) → `LOADOUTS_ASSIGNED` → `STATE`
→ `VOTING_OPENED` → `STATE`. That made `ASSIGNING` genuinely unobservable (`OpenVoting` always ran
before the client could ever see a `STATE` frame in between) - true then, **false now**:
`GameService.scheduleRevealDelay` deliberately holds the Game in `ASSIGNING`, with zero Rounds, for
the whole sorteo, and that window is exactly what the reveal overlay covers. The other consequence
below still holds and is still handled the same way:

**`LOADOUTS_ASSIGNED` arrives before its own `STATE` resend.** At the instant the frame bumps
`live.assignmentSeq`, `snapshot` may still be the pre-assignment one. `shouldReveal` gates on *both*
`assignmentSeq > revealedAssignmentSeq` *and* `hasAllLoadouts(snapshot)` - never on the frame alone.
Getting this wrong would have shown a reveal animation over stale/empty loadout data on the very
first assignment of every game. (It no longer also gates on the round index matching - see the
"Sorteo redesign" section's bug writeup for why that check itself became the problem.)

## Known backend bug worked around, not fixed

`dto.VoteCastPayload.VotesCast` is **always 0** - `game_ws_endpoints.go`'s `buildEventFrame` never
sets it. Harmless for this tanda (no vote UI was in scope), but blocks a live vote-count indicator
later without a small backend fix first (set `VotesCast` from `Ballot.Count()` when building the
frame).

## Locale gap - CLOSED 2026-08-17, see the new section below

~~`RepoPowerPool` resolves stand/devil-fruit data at a fixed `enums.EnGB`... deliberately not
rendered on `loadout-card.tsx`~~ - no longer true. `GameEndpoints` now re-resolves a loadout's
Stand/DevilFruit `description`/`skills` per viewer locale at serialization time
(`standTextResolver`/`devilFruitTextResolver`, the exact analogue of `StageTextResolver`), so both
are safe to render directly. See "Loadout power text goes per-locale + roster redesign" below.

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

## Sorteo redesign: Wii Party-style ruleta, V1 tempo, voting waits for it (2026-08-14)

A real playtest surfaced four more problems on top of the ones above (manga visibility, stepper
layout - already covered elsewhere), all fixed together in one pass:

1. **The reveal had zero suspense** - a 450ms flip + 220ms per slot, card by card, over almost
   before it started.
2. **Voting was already open while the reveal played.** `GameService.beginRound` called
   `OpenVoting`+`scheduleVotingTimer` in the same synchronous call as `AssignLoadouts` - the
   deadline started ticking before anyone had seen a single power.
3. The owner wanted the tempo and ceremony of the original terminal-based
   *JoJoOnePiece_Simulator* (`github.com/oorbea/JoJoOnePiece_Simulator`), which paced a `delay(2.5)`
   hold after each power behind a `loadingScreen()` suspense beat.
4. A Wii Party-style spinning ruleta per participant, landing on the real answer - not just an
   instant pop-in.

**What shipped**: `game.RevealDuration(mangas)` (`apps/backend/.../reveal.go`) is a pure function of
a lobby's mangas alone - never of the actual random draws, so both backend and frontend
(`lib/loadout-reveal.ts`'s `revealDurationMs`) compute the identical total without exchanging
anything beyond `mangas` itself. `GameService.scheduleRevealDelay` (see
[[gameplay-application-layer]]) holds the Game in `ASSIGNING` for exactly that long before calling
`OpenVoting` - voting genuinely cannot open before the sorteo's own deadline, closing the gap in
problem 2 for real (not just cosmetically, as the old "hide the countdown while revealing" UI did).
`LOADOUTS_ASSIGNED`'s payload gained `revealMs`, the frontend's authoritative pacing input (see
`useLoadoutReveal`) - a constants drift between the two sides degrades the ruleta's pacing rather
than desyncing "reveal looks done" from "voting is actually open". `GameStateResponse` gained
`revealEndsAt` so a (re)connecting client can resume the countdown.

The reveal is now **poder a poder for every participant at once** (not participant a participant):
slot N always means the same slot for the whole lobby, driven by one global phase index
(`intro → [spin, land] × N slots → outro`), shown in a dedicated overlay (`match/reveal-stage.tsx`)
instead of animating the roster cards directly - `RevealLane`/`PowerRoulette` render one Wii
Party-style vertical slot-reel per participant, decorated with real catalog names (Stand/DevilFruit)
or every enum member (scalar slots) as filler, always landing on that participant's own actual
loadout value. `LoadoutCard` lost its flip animation and `visibleSlots` entirely - it only ever
renders a finished loadout now, since the ceremony happens in the overlay before the card ever
mounts. `loadoutSlots` (`match-rules.ts`) no longer omits a `NONE` spin/hamon/fruitMastery - every
slot a manga can produce is now always included (with its NONE value if that's what it drew), which
is what makes the slot **count** a pure function of `mangas` alone (mirrored on the backend as
`RevealSlots`). Tempo, roughly modeled on V1's `loadingScreen`/`delay` pacing but with the 10s
Stand-description hold trimmed to 4s (this card never renders a power's description - see the
locale-gap note below): ~1.1s intro, ~1.65s spin per slot, 2.5s hold per scalar slot / 4s for the
Stand/DevilFruit art blocks, ~3.3s outro. Full lobby (both mangas): ~45s. JoJo-only: ~18s.

### The bug this uncovered: `shouldReveal` assumed a Round already existed

`shouldReveal` used to require `currentRound(snapshot).index === live.assignedRoundIndex` as proof
the snapshot had caught up to a specific assignment - a reasonable guard when `AssignLoadouts` and
`OpenVoting` ran in the same synchronous call, since a Round (created inside `OpenVoting`) reliably
existed the instant loadouts did. That's exactly what `scheduleRevealDelay` breaks: the Game now
sits in `ASSIGNING` with **zero Rounds** for the whole sorteo. Gating on the round index made
`shouldReveal` false for that entire window and only flip true the moment `OpenVoting` finally
created the Round - i.e. exactly when voting had already opened, the one moment the sorteo must NOT
still be starting. Symptom in a live playtest: the roster rendered fully-formed instantly on Start
(no overlay at all), and the ruleta only appeared *after* the state label already said "Votación" -
replaying the whole ~45s animation retroactively, well past its purpose. Fix: drop the round check
entirely and gate on `snapshot.state === 'ASSIGNING'` instead - state is authoritative and needs no
Round to exist. **Lesson**: a gating check built for one delivery order between two events silently
breaks the moment either event's timing changes, even though the check itself never looked wrong in
isolation - re-derive gates from primitives (`state`) rather than from a side effect (`a Round
exists`) of the very event you're trying to delay.

## Verification

`tsc --noEmit` clean, `pnpm run test:ci` green (38 suites / 353 tests - a `StageCard` test flake
around a leaked Tamagui spring animation showed up once in a full run and passed both in isolation
and on immediate re-run; unrelated to this tanda's files, not chased further). New pure-logic tests:
`lib/__tests__/{match-rules,loadout-reveal}.test.ts`, extended `game-socket.store.test.ts`. No new
native render tests (consistent with prior lobby tandas - manual `local-up` walkthrough is the
gate for the animated/visual half).

## Reel geometry bug (the ruleta landed blank) + redesign, loadout power text per-locale, voting roster redesign (2026-08-17)

Owner report: the sorteo ruleta looked like every candidate flew past and it "landed on an empty
row" - every single spin, not intermittently.

**Root cause**: `power-roulette.tsx`'s clipping window used `justify="center"` on a strip taller
than the window. Flex centred the strip *before* any transform ran, so `translateY=0` showed the
strip's **middle** row, not its top one, and the old `restY = -(N-1)*ITEM_HEIGHT` (which assumed
row 0 starts flush with the window top) landed the window past the strip's actual end for any pool
bigger than 3 candidates - which every real reel is (10-25 rows). The landing frame was *always*
blank. Confirmed live via `local-up` + `claude-in-chrome`: every slot across a full both-mangas
sorteo (physicalForm, fruitMastery, spin, Stand) now lands centred with real neighbours visible
above/below, never blank.

**Fix + redesign** (`power-roulette.tsx`, `reveal-lane.tsx`, `reveal-stage.tsx`): dropped
`justify="center"`; moved to a 3-row window (`lib/reel-geometry.ts`'s `restRows`/`buildReel`/
`finalLabelIndex`, unit-tested directly - see `lib/__tests__/reel-geometry.test.ts` - so this exact
invariant can't silently regress again) with the landed value centred and a highlighted band;
overshoot-then-spring landing instead of a dead stop; a `@tamagui/linear-gradient` top/bottom fade
instead of a hard clip; a small per-lane stagger (capped at 30% of `REVEAL_SPIN_MS`) so lanes don't
all stop on the same frame. None of `loadout-reveal.ts`'s phase-timing constants changed - the
backend's `scheduleRevealDelay` is keyed to them, only the visuals inside each phase moved.

**Haki labels reworded** (es-ES/ca-ES only, owner-specified): "Haki de Armadura"/"Haki de
Observación"/"Haki del Rey" (ca-ES: "Haki d'Armadura"/"Haki d'Observació"/"Haki del Rei") -
`game.match.trait.*` in the two locale JSONs.

**Loadout power text goes per-locale**: see the "Locale gap" note above, now closed - backend
change in [[gameplay-application-layer]]/[[backend-contract]]'s update. `game.types.ts`'s comment
warning against rendering `stand`/`devilFruit` description+skills is stale and was corrected.

**Voting-round roster redesigned**: `MatchRoster` no longer renders the full `LoadoutCard` per
participant (art blocks, stat grid, every chip - too much at a glance during actual voting). New
`ParticipantTile` (avatar or colour-circle initial or robot icon + username only,
`participant-avatar.tsx` shared with the modal header) is what renders now.
`useHoverTrigger`/`TooltipCard` (generalized from `tooltip.tsx`'s existing
`useTooltipTrigger`/`TooltipBubble` - see [[a11y-web-leak]]'s sibling doc for the original
anchoring rationale, unchanged here) show the old `LoadoutCard` as a 1.5s hover (web) / long-press
(native, dismissed on release instead of a timer - a card is read while still pressing, not after)
card; a tap/click opens `LoadoutModal`, a near-fullscreen breakdown with the description/skills the
old card never had room for, one column on phones and Stand | Devil Fruit side by side from `$md`
up.

## Follow-up polish pass, from actual owner usage (2026-08-17, same day)

Four more fixes, all from the owner actually playing with the redesign above:

1. **The landing bounce still looked like it was sliding to another power.** Two compounding
   causes, both root-caused by reading the reanimated code rather than guessing: (a) the land-beat
   `scale` pop was applied to the *whole scrolling strip* (`Animated.View` carrying `translateY`),
   whose transform origin is the strip's own centre - for a 25+-row Stand/DevilFruit reel the
   visible window sits far from that centre, so scaling by 1.16 displaced the landed answer
   sideways-in-time by up to ~1.8 rows before settling back. Fixed by moving `scale` to a new outer
   `Animated.View` wrapping the whole fixed-size window (whose own centre *is* the highlight
   band's centre), leaving `translateY` on the inner strip alone. (b) The `withSpring` catch leg
   needed ~900ms to settle at its damping/stiffness, but `delay + duration` always equalled
   `spinMs` exactly, leaving as little as ~250ms - the phase machine's own fixed timers cut to
   'land' on schedule regardless and hard-reset `translateY`, snapping a still-mid-bounce reel to
   rest. Replaced the spring with two bounded `withTiming` legs (`landingTiming`,
   `lib/reel-geometry.ts`) whose total duration is provably `<= spinMs` for every lane - unit
   tested directly (`reel-geometry.test.ts`'s `landingTiming` describe block) rather than trusted
   by inspection, since this is exactly the kind of timing invariant that silently breaks if
   `MAX_STAGGER_MS`/`CATCH_MS` are ever retuned. `OVERSHOOT_PX` also shrunk (0.35→0.2 of a row) -
   it alone was large enough to bleed a neighbour's label into the highlight band.
2. **Hover-card delay**: 1.5s → 0.5s (`roster-participant.tsx`'s `HOVER_CARD_DELAY_MS`).
3. **The hover card ran off-screen.** `tooltip.tsx`'s `usePositionedOverlay` only clamped
   horizontally, only on web (native's measuring ref callback always early-returned), and never
   checked whether the card's *height* fit the room in whichever direction the above/below flip
   picked - a ~450px `LoadoutCard` is routinely taller than the space above or below a roster tile.
   Rewritten to measure via `onLayout` (fires on both platforms, unlike the old web-only
   `.offsetWidth` ref-callback trick) and position with explicit `top`/`left` pixel values clamped
   against `Dimensions.get('window')` on both platforms, dropping the web-only
   `translateX/Y`-percentage transform hack entirely. The clamping math itself is a pure
   `clampOverlayPosition` (new `shared/lib/overlay-position.ts`) so "an oversized overlay always
   stays fully on-screen, even when it's too tall for either side" is unit-tested directly instead
   of only exercised by hand.
4. **Added a spin sound.** One continuous loop for the whole reveal (not restarted per slot),
   synthesized locally via a one-off Python script (no external/copyrighted audio) into
   `assets/audio/reel-spin.wav`, played through `expo-audio`'s `useAudioPlayer`. Owned by
   `RevealStage` itself via a small `useRevealSpinSound` hook - `RevealStage` is mounted for
   exactly the reveal's duration (`match-screen.tsx` only renders it while `isRevealing`), so its
   own mount/unmount lifecycle already is the correct start/stop signal, no extra prop threading
   needed. Gated on `!reducedMotion` only (no new mute/settings store - not requested): reduced
   motion already skips every visual spin straight to rest, so the sound playing over a static
   reel would be misleading rather than just superfluous.

All four verified live via `local-up` + `claude-in-chrome` after a docker rebuild (the new
`expo-audio` dependency and the bundled `.wav` needed the frontend image rebuilt, not just a code
reload) - recorded GIFs of a landing on a scalar slot and confirmed the hover card clamping to the
top edge instead of running off past it near the bottom of the roster.

Related: [[game-lobby-todo]], [[game-lobby-frontend]], [[gameplay-application-layer]],
[[game-realtime-transport]], [[docker-setup]], [[frontend-stack]], [[i18n-multi-language]].
