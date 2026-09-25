import { mulberry32 } from '@/features/game/lib/prng'

// Pure geometry for PowerRoulette's slot-machine reel - split out from the
// component so the "does the landed row actually end up inside the visible
// window" invariant is unit-testable without rendering reanimated. This is
// exactly the invariant that broke the original reel (see the fixed
// `justify="center"` bug in power-roulette.tsx's history): the window always
// shows `WINDOW_ROWS` rows, and `restY` must place `finalLabel` in the centre
// one, never past the strip's end.
//
// WINDOW_ROWS went from 3 to 5 rows (owner request, 2026-09-25 playtest
// feedback: too few visible options, and the classic slot-reel look reads
// small). It must stay ODD - the geometry below assumes a single centre row
// with an equal count above and below it (see finalLabelIndex/buildReel's
// TRAILING_FILLERS).

export const WINDOW_ROWS = 5

// How many items trail the landed one - half the window (rounded down),
// mirroring however many lead it. For WINDOW_ROWS=5 that's 2 above, 2
// below, finalLabel dead centre.
const TRAILING_FILLERS = Math.floor(WINDOW_ROWS / 2)

// How many extra candidate ticks scroll past before landing - varies by the
// candidate pool size so a short list (e.g. 4 haki levels) doesn't look like
// it's repeating the same handful of names twice as fast as a long one
// (dozens of Stands/Devil Fruits). Floored at a minimum that reads as a real
// shuffled deck rather than a handful of names cycling past twice (owner
// request, 2026-09-25 playtest feedback: "en la ruleta se ven muy pocas
// opciones posibles y duplicadas").
export function tickCount(poolSize: number): number {
  return Math.max(30, Math.min(50, poolSize * 3))
}

// Reel layout: [ticks...] + finalLabel + TRAILING_FILLERS trailing items.
// The ticks are a SEEDED SHUFFLE of `candidates` sampled with replacement
// (never the same candidate on two adjacent rows), not `candidates`
// repeated in a fixed i%n cycle - that fixed cycling is exactly what made
// the old reel read as "so few options, and duplicated" (same playtest
// feedback tickCount's doc cites). `seed` defaults to 0 so an unseeded call
// stays fully deterministic (existing geometry tests rely on this), but a
// real reveal always passes the same per-(game,round,participant,slot) seed
// loadout-reveal.ts already derives for revealSpinCycles, so every device
// draws the identical "random" reel.
export function buildReel(candidates: string[], finalLabel: string, seed = 0): string[] {
  const leadingMin = Math.max(0, WINDOW_ROWS - 1 - TRAILING_FILLERS)
  if (candidates.length === 0) {
    return [...Array(leadingMin).fill(''), finalLabel, ...Array(TRAILING_FILLERS).fill('')]
  }

  const rand = mulberry32(seed)
  const ticks = Math.max(tickCount(candidates.length), leadingMin)
  const items: string[] = []
  let prev: string | null = null
  const draw = (avoid: (string | null)[]): string => {
    let pick = candidates[Math.floor(rand() * candidates.length)]
    let guard = 0
    while (avoid.includes(pick) && candidates.length > 1 && guard < 8) {
      pick = candidates[Math.floor(rand() * candidates.length)]
      guard += 1
    }
    return pick
  }
  for (let i = 0; i < ticks; i++) {
    // The tick right before finalLabel also avoids matching it, so the
    // reel never shows the same name twice in a row across the seam.
    const pick = draw(i === ticks - 1 ? [prev, finalLabel] : [prev])
    items.push(pick)
    prev = pick
  }
  items.push(finalLabel)
  prev = finalLabel
  for (let i = 0; i < TRAILING_FILLERS; i++) {
    const pick = draw([prev])
    items.push(pick)
    prev = pick
  }
  return items
}

// Index of finalLabel within a reel built by buildReel - always
// TRAILING_FILLERS items before the end (ticks..., finalLabel, trailing
// fillers).
export function finalLabelIndex(reelLength: number): number {
  return reelLength - 1 - TRAILING_FILLERS
}

// Resting translateY (in item-height units) that places finalLabel in the
// centre row of a WINDOW_ROWS-row window: the window shows items
// [reelLength-WINDOW_ROWS, ..., reelLength-1], and finalLabel
// (reelLength-2) must be the middle one of those.
export function restRows(reelLength: number): number {
  return -(reelLength - WINDOW_ROWS)
}

// Per-lane stagger so every carril doesn't stop on the exact same frame -
// capped at 30% of the spin budget so even the last lane's landing timeline
// still fits inside spinMs.
const MAX_STAGGER_MS = 70
const MAX_STAGGER_SHARE = 0.3

export type LandingTiming = { delayMs: number; decelMs: number; catchMs: number }

// The reel's landing timeline: decelerate-past-target then a short bounded
// catch, split out so "does the WHOLE animated sequence (delay + decel +
// catch) finish before spinMs elapses" is unit-testable without rendering
// reanimated. This is exactly the invariant a physics withSpring catch used
// to violate: at ~900ms to settle, it routinely outlived the ~250ms budget
// left after a long stagger delay, so the phase machine's fixed timers cut
// to 'land' and hard-reset translateY mid-bounce - read as "still moving."
// Both legs here are bounded withTiming calls, so the total is exact and
// deterministic; catchMs is a fixed constant, decelMs absorbs whatever
// budget remains after delay and catchMs are accounted for.
export function landingTiming(spinMs: number, laneIndex: number, catchMs: number): LandingTiming {
  const delayMs = Math.min(laneIndex * MAX_STAGGER_MS, spinMs * MAX_STAGGER_SHARE)
  const duration = Math.max(spinMs - delayMs, spinMs * 0.5)
  const decelMs = Math.max(duration - catchMs, duration * 0.5)
  return { delayMs, decelMs, catchMs }
}
