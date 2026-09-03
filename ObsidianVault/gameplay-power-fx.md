---
title: "Feature (planned): per-power special effects in the sorteo"
tags:
  - project
  - jojo-onepiece-simulator
  - frontend
  - feature
  - planned
  - design
---

# Per-power special effects in the sorteo (planned, 2026-08-30)

**Status: planning only. Nothing in this note is built.** The owner explicitly asked for this to
be documented as a TODO alongside the jugador-por-jugador sorteo rewrite ([[game-match-assignment-
frontend]]'s 2026-08-30 section), not implemented in the same pass. Read that note first for the
reveal structure this plugs into: one participant's turn at a time, a `PowerRevealCard` full-screen
overlay when a Stand/Devil Fruit lands, an avatar strip for the rest of the lobby.

## Why

Certain iconic powers deserve to *look* like themselves when they're drawn, not just print their
name and stats: Gomu Gomu no Mi's owner should visibly bounce, Holy's Stand should feel grey and
thorny, The World should freeze the scene for a beat before continuing. This is **purely
cosmetic** — explicitly no gameplay effect, no stat change, nothing `Loadout`/`LoadoutEvaluator`
ever sees. It only exists to make certain draws feel special in the moment they're revealed.

## Open questions (confirmed with the owner, 2026-08-30)

| Question | Answer |
|---|---|
| Which powers get an effect? | **Curated by hand + a fallback by rarity/type.** A hand-picked registry the owner grows over time (Gomu Gomu no Mi, Holy's Stand, The World, etc. to start); anything not in that registry falls back to a generic effect keyed off `Rarity` (e.g. every `LEGENDARY` power) or off `FruitType` (Logia/Zoan/Paramecia each get a family-level treatment) rather than nothing at all. |
| Where does the effect show? | **Two places**: (a) the sorteo's own big reveal card (`PowerRevealCard`) — the overlay itself is decorated with the effect (bounce, greyscale+brambles, freeze-frame, etc.); (b) the participant's own avatar/lane during the *rest of the sorteo* after that slot lands — e.g. once someone's Stand effect triggers, their avatar in the bottom strip carries it until the reveal ends. **Not** persisted into the roster during voting rounds — purely a sorteo-scoped flourish, confirmed explicitly out of scope for now. |
| Can an effect change timing? | **Yes, capped.** An effect may add up to `RevealFxMaxMs` (reserved in `reveal.go`, currently `3000` and unused — nothing adds to it yet) to its own slot's hold, so e.g. The World gets its extra "frozen" beat baked into the total reveal duration the server already computes and the client already paces to. The backend sums this into `game.RevealDuration` once any effect actually declares a nonzero `extraHoldMs`; until then it's a no-op ceiling. |

## Shape sketch (not yet built)

A frontend-only registry, most naturally living beside `power-block.tsx`/`reveal-stage.tsx`:

```ts
// features/game/lib/power-fx.ts (SKETCH — not implemented)
type PowerFx = {
  /** transform/opacity-only, matching frontend-responsive-frutiger-aero.md's
   * "only transform/opacity animate" rule - no new CSS/RN style categories. */
  cardEffect: 'bounce' | 'greyscale-brambles' | 'freeze-frame' | ...
  avatarEffect?: 'bounce' | 'greyscale-brambles' | ...
  /** ms this effect adds to its own slot's hold, capped at RevealFxMaxMs. */
  extraHoldMs?: number
}

// Keyed by Stand/DevilFruit id (curated) — falls back to rarity/fruitType
// when no entry exists, never to "nothing".
const POWER_FX: Record<string /* power id */, PowerFx> = { /* seed catalog below */ }
const RARITY_FX_FALLBACK: Partial<Record<Rarity, PowerFx>> = { LEGENDARY: { cardEffect: 'glow-legendary' } }
const FRUIT_TYPE_FX_FALLBACK: Partial<Record<FruitType, PowerFx>> = {
  LOGIA: { cardEffect: 'elemental-shimmer' },
  ZOAN: { cardEffect: 'fur-flicker' },
  PARAMECIA: { cardEffect: 'soft-glow' },
}
```

## Seed catalog

**Superseded by [[gameplay-power-fx-stand-catalog]] (2026-09-03)** — the owner's full curated list
of 25 Stand effects (Star Platinum, The World, Holy's Stand, Crazy Diamond, Killer Queen, Gold
Experience (Requiem), King Crimson, Chariot Requiem, Stone Free, Weather Report (Heavy Weather),
Whitesnake, C-MOON, Made in Heaven, all four Tusk Acts, Ball Breaker, Dirty Deeds Done Dirt Cheap,
D4C Love Train, Soft & Wet (Go Beyond), Paisley Park, Wonder of U) lives there now, not in this
note. Devil Fruit effects are still undecided — see that note's own "Devil Fruits — not decided
yet" section. The owner will keep extending both catalogs over time; grow the sibling note(s), not
this one.

**Sound (2026-09-03, not yet reflected in the shape sketch below)**: every curated Stand entry also
gets its own war-cry sound effect (Tusk Act 1 is sound-*only*, no visual) — audio assets TBD, the
owner provides them when this is actually built. `PowerFx` needs an optional `soundEffect` field
alongside `cardEffect`/`avatarEffect` once that happens.

## Constraints this must respect when eventually built

- **Animation**: only `transform`/`opacity` animate (`frontend-responsive-frutiger-aero.md`) — no
  new gradient/shadow/filter categories invented per-effect.
- **Platform split**: any effect complex enough to need a real animation driver follows the
  existing `.web.tsx` (CSS keyframes) / `.native.tsx` (Reanimated worklets) split, same as other
  animated backdrops in this codebase.
- **Reduced motion**: `useReducedMotion()` must collapse every effect to its instant/static end
  state — a power still reads as "special" (still greyscale, still shows the freeze-frame's static
  composition) without the animation itself, consistent with how the rest of the reveal degrades.
- **No `Animated.ValueXY`** — this repo's eslint rule rejects it (see `game-lobby-todo.md`'s
  drag-to-move precedent); use `useState` + `useMemo`'d primitives instead.
- **Timing invariants stay pure and tested**: if `extraHoldMs` starts being nonzero, the sum going
  into `game.RevealDuration`/`revealTimeline` must be provable to stay within `RevealFxMaxMs`
  per slot, the same way `lib/reel-geometry.ts`'s `landingTiming` is unit-tested rather than
  trusted by inspection.
- **Backend/frontend agreement**: exactly like `RevealSpinCycles`, if `extraHoldMs` is nonzero it
  must be a **deterministic** function of the same inputs both sides already have (power id +
  RevealSpeed, most likely) — never server-random, never client-only — so the reveal timeline both
  sides compute never drifts.

## Deliberately still out of scope (per the owner's own scoping in this same conversation)

- Persisting any effect into the roster during voting rounds — sorteo-scoped only.
- Any gameplay/stat consequence — purely a presentation-layer table.
- Building the registry itself, the effect components, or wiring `extraHoldMs` into the backend —
  this note is the spec to build from next, not a partial implementation.

Related: [[gameplay-power-fx-stand-catalog]], [[game-match-assignment-frontend]],
[[gameplay-game-modes]], [[frontend-responsive-frutiger-aero]], [[norma-diseno-ui-ux]].
