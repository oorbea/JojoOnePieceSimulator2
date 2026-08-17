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
