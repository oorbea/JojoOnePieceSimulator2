import { HAPTIC_PATTERNS, vibrate, vibrationSupported, type HapticCue } from '../haptics'

type VibrateFn = (pattern: number[]) => boolean

function setVibrate(fn: VibrateFn | undefined) {
  Object.defineProperty(navigator, 'vibrate', { value: fn, configurable: true, writable: true })
}

describe('vibrate', () => {
  afterEach(() => setVibrate(undefined))

  it('sends the cue pattern to the Vibration API', () => {
    const spy = jest.fn<boolean, [number[]]>().mockReturnValue(true)
    setVibrate(spy)

    expect(vibrate('votingOpened')).toBe(true)
    expect(spy).toHaveBeenCalledWith(HAPTIC_PATTERNS.votingOpened)
  })

  it('is a silent no-op where the API does not exist (iOS, desktop)', () => {
    setVibrate(undefined)
    expect(vibrate('gameFinished')).toBe(false)
  })

  it('swallows a browser that throws', () => {
    setVibrate(() => {
      throw new Error('blocked')
    })
    expect(vibrate('roundResolved')).toBe(false)
  })

  it('keeps every pattern brief but strong enough to feel', () => {
    for (const cue of Object.keys(HAPTIC_PATTERNS) as HapticCue[]) {
      const pattern = HAPTIC_PATTERNS[cue]
      expect(pattern.length).toBeGreaterThan(0)
      expect(pattern.every((ms) => Number.isInteger(ms) && ms > 0)).toBe(true)
      expect(pattern.reduce((sum, ms) => sum + ms, 0)).toBeLessThanOrEqual(1200)
      // Even-index entries are the actual buzzes; very short ones are not
      // rendered by many phone motors.
      const buzzes = pattern.filter((_, i) => i % 2 === 0)
      expect(Math.min(...buzzes)).toBeGreaterThanOrEqual(120)
    }
  })
})

describe('vibrationSupported', () => {
  const originalMatchMedia = window.matchMedia
  afterEach(() => {
    setVibrate(undefined)
    window.matchMedia = originalMatchMedia
  })

  function setPointerCoarse(coarse: boolean) {
    window.matchMedia = ((query: string) => ({
      matches: query === '(pointer: coarse)' ? coarse : false,
    })) as unknown as typeof window.matchMedia
  }

  it('is true on a touch device that exposes the API', () => {
    setVibrate(() => true)
    setPointerCoarse(true)
    expect(vibrationSupported()).toBe(true)
  })

  it('is false on desktop even though Chrome defines navigator.vibrate', () => {
    setVibrate(() => false)
    setPointerCoarse(false)
    expect(vibrationSupported()).toBe(false)
  })

  it('is false when the API is missing', () => {
    setVibrate(undefined)
    setPointerCoarse(true)
    expect(vibrationSupported()).toBe(false)
  })
})
