// Pure geometry for PowerRoulette's slot-machine reel - split out from the
// component so the "does the landed row actually end up inside the visible
// window" invariant is unit-testable without rendering reanimated. This is
// exactly the invariant that broke the original reel (see the fixed
// `justify="center"` bug in power-roulette.tsx's history): the window always
// shows `WINDOW_ROWS` rows, and `restY` must place `finalLabel` in the centre
// one, never past the strip's end.

export const WINDOW_ROWS = 3

// How many extra candidate ticks scroll past before landing - varies by the
// candidate pool size so a short list (e.g. 4 haki levels) doesn't look like
// it's repeating the same handful of names twice as fast as a long one
// (dozens of Stands). The final tick is always the real answer.
export function tickCount(poolSize: number): number {
  return Math.max(10, Math.min(24, poolSize * 3))
}

// Reel layout: [ticks...] + finalLabel + one trailing filler. The trailing
// filler exists purely so the landed row always has a plausible neighbour
// below it instead of a blank window edge - it is never the answer.
export function buildReel(candidates: string[], finalLabel: string): string[] {
  if (candidates.length === 0) return ['', finalLabel, '']
  const ticks = tickCount(candidates.length)
  const items: string[] = []
  for (let i = 0; i < ticks; i++) items.push(candidates[i % candidates.length])
  items.push(finalLabel)
  items.push(candidates[ticks % candidates.length])
  return items
}

// Index of finalLabel within a reel built by buildReel - always the second
// to last element (ticks..., finalLabel, trailing filler).
export function finalLabelIndex(reelLength: number): number {
  return reelLength - 2
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
