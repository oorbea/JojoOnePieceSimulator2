---
title: Victory/defeat cinematic
tags: [project, jojo-onepiece-simulator, frontend, feature, gameplay]
---

# Victory/defeat cinematic (2026-09-14)

Frontend-only feature: end-of-game and end-of-round cinematics so winning
and losing actually *feel* like something, in both [[gameplay-game-modes]].
Design pacted with the owner via brainstorming before any code (script,
durations, personalization level, audio strategy) — see the plan file this
was built from if it still exists locally; this note is the durable record.

## Design decisions (pacted, do not relitigate without asking)

- **Tone**: defeat is crude anime-cinematic (screen fracture, ink,
  desaturation, manga typography); victory is solemn/mythic (light, choir,
  names "recordados para siempre").
- **Duration**: big cinematics run 6-8s, skip enabled from 2s in, **per
  viewer** — one player skipping never advances anyone else, and never
  advances the server (same "skip only hides locally" precedent as
  [[game-round-result-2026-08-28]]'s `resultDismissed`).
- **Perspective**: GAUNTLET is collective (everyone sees the same
  victory/defeat); VERSUS is personal (winners see victory, losers see
  defeat, simultaneously, independently skippable).
- **Round-level**: a 1.5s full-screen flash at the start of RESOLVING.
  VERSUS: win/lose by your own team. GAUNTLET: a mini-victory on SURVIVE
  only — a FALL round never gets a mini-defeat, because it's also always the
  final round and the big defeat cinematic owns that moment (avoiding a
  false-alarm double beat).
- **Personalization**: "medium" — names, team/verdict text. **No avatar
  thumbnails and no power/stage art** inside the cinematic itself (scope cut
  made during implementation for time — `MatchOutcome`/`ParticipantOutcomeResponse`
  don't carry `avatarThumb` today either, so adding avatars would need a
  join against `snapshot.participants` that isn't guaranteed complete for a
  player who left; text-only sidesteps that entirely).
- **Audio**: added the app's first global mute/volume control
  (`shared/stores/audio-settings.store.ts`), since none existed and these
  cinematics are loud on purpose.
- **Replayable**: a "Watch again" button on `MatchResultScreen`, host and
  non-host both get it.

## What shipped

No backend changes, no contract regeneration — everything needed
(`matchRecap`'s per-seat `won: boolean | null`, `ROUND_RESOLVED`'s winner,
the fixed `ResultDuration`/RESOLVING pause) already existed. Pure additions:

- `features/game/lib/outcome-cinematic.ts` — the phase/timing tables
  (`cinematicTimeline`), same split-for-testability pattern as
  `lib/reel-geometry.ts`/`lib/loadout-reveal.ts`. Audio is the master clock:
  every phase's `startMs` is the beat map of the matching wav.
- `features/game/lib/game-result.ts` — added `roundOutcome(snapshot, you,
  round)`, sibling to `matchRecap`.
- `features/game/hooks/use-outcome-cinematic.ts` — decides eligibility
  (FINISHED + not aborted) and drives the one-shot-then-replayable
  lifecycle. Session state is keyed by `snapshot.id` so a rematch's new game
  is eligible again without an effect-based reset.
- `components/presentational/match/outcome-cinematic.tsx` — the big
  cinematic. **No `visible` prop** — the caller only mounts it while a
  play-through should show, so every mount starts fresh with no reset
  plumbing needed. Reanimated directly (not the CSS/Reanimated
  `.web.tsx`/`.native.tsx` split `bubble-field.tsx` uses for infinite
  loops) — same single-file precedent as `power-roulette.tsx` for
  finite-duration animation. Transform/opacity only, per project norm.
- `components/presentational/match/round-flash.tsx` — the 1.5s round beat,
  mounted inline in `match-screen.tsx` right before `RoundResultPanel`,
  tracked by round index so a re-render never replays it mid-window.
- `shared/hooks/use-sound.ts` — the one place that applies the global
  mute/volume store to an `expo-audio` player; `use-reveal-spin-sound.ts`
  (the sorteo reel) was refactored onto it too, so the mute toggle now
  covers it as well.
- i18n: `game.result.cinematic.*` (the big one) and
  `game.match.cinematic.*` (the round flash), all three locales, no em/en
  dashes (copy.test.ts).

## Gotchas hit

- **This repo's hooks lint is stricter than vanilla React**:
  `react-hooks/set-state-in-effect` rejects `setState` called synchronously
  in an effect body (only from inside an async callback/timeout is fine),
  and `react-hooks/refs` rejects reading OR writing a ref during render —
  so the usual "reset state via a ref comparison during render" trick
  doesn't fly here either. The pattern that satisfies both: fold the value
  you're resetting into the SAME state object as the id/prop it depends on
  (`{ gameId, seen, replaying }`), and derive a "fresh" default inline when
  they don't match, with no mutation of anything until an event actually
  fires. See `use-outcome-cinematic.ts`. For "this must restart every time
  a prop flips true", the simpler fix is often "just don't keep the
  component mounted when it's not showing" (`outcome-cinematic.tsx` has no
  `visible` prop at all for this reason) — React's own `key` remount does
  the reset for free.
- `expo-audio`'s `AudioPlayer.volume`/`.muted` are plain mutable properties
  by design (native SharedObject), but this project's
  `react-hooks/immutability` rule flags assigning to anything returned from
  a hook regardless — Reanimated's `useSharedValue().value` is special-cased
  by that rule, `useAudioPlayer()`'s return isn't. Fixed by moving the
  assignment into a `useEffect` + a documented `eslint-disable-next-line`,
  see `use-sound.ts`.

## Not done / left for the owner

- **Defeat audio is real (2026-09-14, same day).** The owner supplied 7
  one-shots (`crack`, `deep_drone`, `horn_of_doom`, `queue`, `scream`,
  `severe_blow`, `thud` — sourced from
  `C:\Users\meckp\Desktop\jojo_one_piece_simulator_sounds\defeat`, not
  checked into the repo). `scripts/compose_defeat_audio.py` (pure stdlib -
  no ffmpeg/numpy on this machine) resamples each to 44.1kHz, applies a
  per-layer gain to balance wildly inconsistent source levels (-36dB RMS to
  0dBFS-clipping across the 7 files), mixes them at the exact ms offsets
  `lib/outcome-cinematic.ts`'s `cinematicTimeline('defeat')` phases start
  at (thud@0/impact, crack+scream@400-500/fracture, severe_blow@1200/ink,
  horn_of_doom@2200/verdict, deep_drone@4000/silence, queue@5700/epilogue),
  applies a 400ms fade-out synced to the visual epilogue cross-fade, and
  peak-normalizes to -1dBFS (result: -14.4dBFS RMS, close to the ~-14 LUFS
  target given to the owner). Output is `assets/audio/defeat-full.wav`,
  exactly 7200ms = `cinematicTimeline('defeat').totalMs`. Re-run the script
  if the source one-shots change; `scripts/analyze_wav.py <file>...` prints
  duration/peak/RMS for any wav, useful for re-balancing.
- **Victory audio is real too (2026-09-14, later same day).** 7 more
  one-shots from the owner (`shimmer`, `female_choral`, `bell_epic_choir`,
  `arpegio`, `final_chord_sustain`, `queue`, `tick` — same source-folder
  convention, `.../jojo_one_piece_simulator_sounds/victory`, not checked
  into the repo). `scripts/compose_victory_audio.py` mixes the first 6 into
  `victory-full.wav` (breath/dawn/coronation/memory/seal/epilogue, same
  gain-balancing + fade-out + peak-normalize approach as
  `compose_defeat_audio.py` - both now share `audio_compose_lib.py` rather
  than duplicating the mixing code). `tick.wav` is deliberately **not**
  baked into the composite: the real winner count varies per match, so it's
  normalized standalone into `name-tick.wav`
  (`normalize_one_shot()`) and played live, once per name, by
  `outcome-cinematic.tsx`'s own "memory" beat - a `setTimeout` cascade
  (`revealedCount` state, capped at 400ms between names so a big squad's
  names don't run past the beat) reveals each name and fires the tick in
  step, replacing what was originally a single group fade-in with no name-
  by-name pacing. `name-tick.wav` existed as an unused asset+silent-
  placeholder before this - this is also what finally wired it up.
- **Round-flash one-shots are real too (2026-09-14, same day again).**
  `glissando_up.wav`/`rock_crack.wav` (same source-folder convention,
  `.../jojo_one_piece_simulator_sounds/rounds`) → `round-win.wav`/
  `round-lose.wav` via `scripts/compose_round_audio.py`. These needed a
  different treatment than defeat/victory: both source clips run ~2.2-2.4s
  but `round-flash.tsx` hard-`stop()`s playback the instant its own 1500ms
  timer ends, so a straight normalize would have cut them off mid-ring.
  `audio_compose_lib.py` gained `produce_one_shot()` for this - trims to
  1450ms with a 350ms fade-out (timed to land inside round-flash.tsx's own
  350ms visual fade-out tail) before peak-normalizing to -1dBFS.
  `rock_crack.wav`'s source clipped at 0.0dBFS; now safely under.
- **All five cinematic sounds are now real** (`defeat-full`, `victory-full`,
  `name-tick`, `round-win`, `round-lose`) - `scripts/generate-silent-
  placeholder-audio.py` is deleted, it had nothing left to generate.
- Swapping any of these files in `assets/audio/` is a straight replacement,
  no code change — but **needs a docker image rebuild**, not just a
  reload, same as the sorteo's own audio precedent.
- **No avatar thumbnails in the cinematic** (see Personalization above) —
  revisit only if the owner asks; would need `ParticipantOutcomeResponse`
  to start carrying `avatarThumb` (a backend/contract change) to do
  properly for a player who left mid-game.
- **Not yet verified live** (two-tab/two-account Versus test, Gauntlet
  survive/fall, reduced-motion, mute, keyboard) — auth is Google-only with
  no dev bypass, so this needs two real logins, one per tab, logged in
  sequentially (logging into a second tab that's already authenticated
  re-reads localStorage and swaps the first account instead of adding a
  second one). Automated tests (unit + component) are green; `/verify`'s
  Docker suite has not been run for this change yet.
