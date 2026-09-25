import {
  buildReel,
  finalLabelIndex,
  landingTiming,
  restRows,
  tickCount,
  WINDOW_ROWS,
} from '@/features/game/lib/reel-geometry'

// This is exactly the invariant that was broken before the sorteo reel fix:
// a `justify="center"` window flex-centred the strip before any transform
// ran, so the landed frame always sat past the strip's real end for any
// realistic candidate pool. These tests pin "the final label is always
// inside the visible window at rest" so a future refactor can't reintroduce
// that regression silently.
describe('reel geometry', () => {
  it.each([4, 5, 10, 24, 50])(
    'for a %i-candidate pool, the final label lands centred in the window',
    (poolSize) => {
      const candidates = Array.from({ length: poolSize }, (_, i) => `candidate-${i}`)
      const reel = buildReel(candidates, 'THE ANSWER')
      const idx = finalLabelIndex(reel.length)
      expect(reel[idx]).toBe('THE ANSWER')

      const restRowOffset = restRows(reel.length)
      // The window's visible row range, in reel-index units, once resting:
      // [-restRowOffset, -restRowOffset + WINDOW_ROWS).
      const firstVisible = -restRowOffset
      const lastVisible = firstVisible + WINDOW_ROWS - 1
      expect(idx).toBeGreaterThanOrEqual(firstVisible)
      expect(idx).toBeLessThanOrEqual(lastVisible)
      // Specifically the CENTRE row, not just "somewhere visible".
      expect(idx).toBe(firstVisible + Math.floor(WINDOW_ROWS / 2))

      // Never past the strip's actual end (the original bug).
      expect(lastVisible).toBeLessThan(reel.length)
      expect(firstVisible).toBeGreaterThanOrEqual(0)
    }
  )

  it('degenerates to a WINDOW_ROWS-row reel with the answer centred when there are no candidates', () => {
    const reel = buildReel([], 'SOLO')
    expect(reel).toHaveLength(WINDOW_ROWS)
    expect(reel[finalLabelIndex(reel.length)]).toBe('SOLO')
    expect(restRows(reel.length)).toBe(-0) // reel.length === WINDOW_ROWS, so no scroll at all
  })

  it('tickCount scales with pool size but stays within [30, 50] (owner request, 2026-09-25: too few visible options before)', () => {
    expect(tickCount(1)).toBe(30)
    expect(tickCount(4)).toBe(30)
    expect(tickCount(100)).toBe(50)
  })

  it('appends exactly WINDOW_ROWS-many trailing items after the final label', () => {
    const candidates = ['a', 'b', 'c']
    const reel = buildReel(candidates, 'FINAL')
    const idx = finalLabelIndex(reel.length)
    expect(reel).toHaveLength(idx + 1 + Math.floor(WINDOW_ROWS / 2))
    for (let i = idx + 1; i < reel.length; i++) {
      expect(reel[i]).not.toBe('FINAL')
    }
  })

  it('is a seeded shuffle, not a fixed i%n cycle - the same seed always reproduces the same reel', () => {
    const candidates = Array.from({ length: 20 }, (_, i) => `c${i}`)
    const a = buildReel(candidates, 'ANSWER', 42)
    const b = buildReel(candidates, 'ANSWER', 42)
    expect(a).toEqual(b)
  })

  it('different seeds produce different reels', () => {
    const candidates = Array.from({ length: 20 }, (_, i) => `c${i}`)
    const a = buildReel(candidates, 'ANSWER', 1)
    const b = buildReel(candidates, 'ANSWER', 2)
    expect(a).not.toEqual(b)
  })

  it('never repeats the same candidate on two adjacent rows, across many seeds', () => {
    const candidates = Array.from({ length: 20 }, (_, i) => `c${i}`)
    for (let seed = 0; seed < 30; seed++) {
      const reel = buildReel(candidates, 'ANSWER', seed)
      for (let i = 1; i < reel.length; i++) {
        expect(reel[i]).not.toBe(reel[i - 1])
      }
    }
  })
})

describe('landingTiming', () => {
  const REVEAL_SPIN_MS = 1650
  const CATCH_MS = 180

  it.each([0, 1, 2, 3, 4, 8])('lane %i finishes delay+decel+catch within spinMs', (laneIndex) => {
    const { delayMs, decelMs, catchMs } = landingTiming(REVEAL_SPIN_MS, laneIndex, CATCH_MS)
    expect(delayMs + decelMs + catchMs).toBeLessThanOrEqual(REVEAL_SPIN_MS)
    expect(catchMs).toBe(CATCH_MS)
    expect(decelMs).toBeGreaterThan(0)
  })

  it('a longer spin budget scales the same way', () => {
    const spinMs = 3000
    for (let laneIndex = 0; laneIndex < 10; laneIndex++) {
      const { delayMs, decelMs, catchMs } = landingTiming(spinMs, laneIndex, CATCH_MS)
      expect(delayMs + decelMs + catchMs).toBeLessThanOrEqual(spinMs)
    }
  })
})
