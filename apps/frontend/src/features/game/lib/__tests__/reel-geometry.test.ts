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

  it('degenerates to a 3-row reel with the answer centred when there are no candidates', () => {
    const reel = buildReel([], 'SOLO')
    expect(reel).toEqual(['', 'SOLO', ''])
    expect(finalLabelIndex(reel.length)).toBe(1)
    expect(restRows(reel.length)).toBe(-0) // reel.length === WINDOW_ROWS, so no scroll at all
  })

  it('tickCount scales with pool size but stays within [10, 24]', () => {
    expect(tickCount(1)).toBe(10)
    expect(tickCount(4)).toBe(12)
    expect(tickCount(100)).toBe(24)
  })

  it('always appends exactly one trailing filler after the final label', () => {
    const candidates = ['a', 'b', 'c']
    const reel = buildReel(candidates, 'FINAL')
    expect(reel).toHaveLength(finalLabelIndex(reel.length) + 2)
    expect(reel[reel.length - 1]).not.toBe('FINAL')
  })
})

// This is the invariant that broke with a physics withSpring catch: at
// REVEAL_SPIN_MS=1650 and a realistic CATCH_MS, a spring needs far longer
// than the budget left after a long stagger delay to settle, so the phase
// machine's fixed timers cut to 'land' and hard-reset translateY mid-bounce
// - read as "still spinning toward another power." landingTiming's two
// bounded withTiming legs must never exceed the reel's own spin budget, for
// every lane a reveal can realistically have.
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
