import { useEffect } from 'react'
import { useAudioPlayer } from 'expo-audio'

import type { RevealPhaseKind } from '@/features/game/lib/loadout-reveal'

// Synthesized locally (no external/copyrighted audio) - see the generator
// script these assets were produced from, kept for reference. reel-spin.wav
// is a Wii Party style wheel: a run of bright decelerating "tock" ticks
// timed to the SAME Easing.out(cubic) curve the reel itself decelerates on
// (power-roulette.tsx), lasting exactly REVEAL_SPIN_MS so it finishes right
// as the reel lands - not a continuous drone. reel-land.wav is a short
// ascending three-note chime for the landing beat.
const REEL_SPIN_SOUND = require('../../../../assets/audio/reel-spin.wav')
const REEL_LAND_SOUND = require('../../../../assets/audio/reel-land.wav')

// Plays the tick run once per 'spin' phase (there's one per slot - 9 for a
// both-mangas lobby) and the landing chime once per 'land' phase, driven
// directly by RevealStage's own phase prop instead of one background loop
// for the whole reveal - the previous continuous hum had no relationship
// to the actual reel motion and read as an unpleasant drone. Silent during
// 'intro'/'outro' (nothing is spinning then). Gated on `enabled` (callers
// pass `!reducedMotion`) - reduced motion already skips every visual
// spin/overshoot straight to rest, so ticking sound over a static reel
// would be misleading, not just superfluous. No mute/settings toggle - not
// requested.
export function useRevealSpinSound(phase: RevealPhaseKind, enabled: boolean): void {
  const spinPlayer = useAudioPlayer(REEL_SPIN_SOUND)
  const landPlayer = useAudioPlayer(REEL_LAND_SOUND)

  useEffect(() => {
    if (!enabled) return
    // Only pause the player this same phase actually started - pausing the
    // OTHER one too (the original version did, unconditionally) means
    // pausing a player that was never told to play on every single phase
    // transition, including the silent 'intro'/'outro' ones. On web that
    // stray pause() can race a play() promise that hasn't settled yet and
    // logs a benign but noisy "play() request was interrupted" AbortError.
    // Even with this guard, expo-audio's web shim can still log that same
    // AbortError on the ~1650ms 'spin' clip's own natural end-of-phase
    // pause() (browsers don't guarantee a play() promise settles before a
    // near-simultaneous pause()) - it's a console warning only, playback
    // itself is unaffected, and there's no promise this code is handed back
    // to attach a .catch() to (expo-audio's play()/pause() are typed void).
    if (phase === 'spin') {
      spinPlayer.seekTo(0).catch(() => {})
      spinPlayer.play()
      return () => spinPlayer.pause()
    }
    if (phase === 'land') {
      landPlayer.seekTo(0).catch(() => {})
      landPlayer.play()
      return () => landPlayer.pause()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps -- spinPlayer/landPlayer are stable SharedObjects from useAudioPlayer, not reactive deps; only phase/enabled should restart this effect
  }, [phase, enabled])
}
