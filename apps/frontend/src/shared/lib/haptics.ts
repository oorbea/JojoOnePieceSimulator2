// Vibration cues for the moments a player must notice with the phone in hand
// (Vibration API: Android Chromium; iOS Safari doesn't implement it, so every
// call is a silent no-op there). Patterns are [vibrate, pause, vibrate, ...] in
// milliseconds.
//
// Pulses are deliberately long (>= 120 ms): the first, very short patterns
// (40-70 ms) are below what many phone motors render at all, and a cue nobody
// can feel is the same as no cue.
export type HapticCue =
  | 'powerLand'
  | 'votingOpened'
  | 'roundResolved'
  | 'gameFinished'
  | 'test'

export const HAPTIC_PATTERNS: Record<HapticCue, number[]> = {
  // Your own Stand / Devil Fruit lands in the sorteo.
  powerLand: [120, 80, 220],
  // A vote (or tiebreak vote) just opened.
  votingOpened: [260],
  // The round's tally is in.
  roundResolved: [150, 100, 150],
  // Game over.
  gameFinished: [200, 100, 200, 100, 450],
  // The profile's "try vibration" button.
  test: [300],
}

// A phone-like device that exposes the API. Desktop Chrome defines
// navigator.vibrate too (it just does nothing), so the coarse-pointer check is
// what separates "can actually buzz" from "has the function".
export function vibrationSupported(): boolean {
  if (typeof navigator === 'undefined' || typeof navigator.vibrate !== 'function') return false
  return typeof window !== 'undefined' && window.matchMedia?.('(pointer: coarse)').matches === true
}

export function vibrate(cue: HapticCue): boolean {
  if (typeof navigator === 'undefined' || typeof navigator.vibrate !== 'function') return false
  try {
    return navigator.vibrate(HAPTIC_PATTERNS[cue])
  } catch {
    // Some browsers throw when the page has no user activation yet or the
    // call comes from a frame without permission; a missed buzz is harmless.
    return false
  }
}
