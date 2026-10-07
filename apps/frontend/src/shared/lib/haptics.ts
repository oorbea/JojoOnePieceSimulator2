// Short vibration cues for the moments a player must notice with the phone in
// hand (Vibration API: Android Chromium; iOS Safari doesn't implement it, so
// every call is a silent no-op there). Patterns are [vibrate, pause, vibrate,
// ...] in milliseconds, kept brief - the cue says "look at the screen", it
// isn't a ringtone.
export type HapticCue = 'powerLand' | 'votingOpened' | 'roundResolved' | 'gameFinished'

export const HAPTIC_PATTERNS: Record<HapticCue, number[]> = {
  // Your own Stand / Devil Fruit lands in the sorteo.
  powerLand: [40, 50, 110],
  // A vote (or tiebreak vote) just opened.
  votingOpened: [70],
  // The round's tally is in.
  roundResolved: [50, 40, 50],
  // Game over.
  gameFinished: [80, 60, 80, 60, 220],
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
