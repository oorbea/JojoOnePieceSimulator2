import { HAPTIC_PATTERNS, vibrate, type HapticCue } from '../haptics'

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

  it('keeps every pattern short and well-formed', () => {
    for (const cue of Object.keys(HAPTIC_PATTERNS) as HapticCue[]) {
      const pattern = HAPTIC_PATTERNS[cue]
      expect(pattern.length).toBeGreaterThan(0)
      expect(pattern.every((ms) => Number.isInteger(ms) && ms > 0)).toBe(true)
      expect(pattern.reduce((sum, ms) => sum + ms, 0)).toBeLessThanOrEqual(600)
    }
  })
})
