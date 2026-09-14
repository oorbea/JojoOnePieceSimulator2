import { useEffect } from 'react'

import type { RevealPhaseKind } from '@/features/game/lib/loadout-reveal'
import { useSound } from '@/shared/hooks/use-sound'

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
// would be misleading, not just superfluous. Volume/mute now come from the
// global audio-settings store via use-sound.ts (2026-09-14) - previously
// there was no way to silence this short of OS-level muting.
export function useRevealSpinSound(phase: RevealPhaseKind, enabled: boolean): void {
  const spin = useSound(REEL_SPIN_SOUND, enabled)
  const land = useSound(REEL_LAND_SOUND, enabled)

  useEffect(() => {
    if (!enabled) return
    // Only stop the player this same phase actually started - stopping the
    // OTHER one too (the original version did, unconditionally) means
    // pausing a player that was never told to play on every single phase
    // transition, including the silent 'intro'/'outro' ones. On web that
    // stray pause() can race a play() promise that hasn't settled yet and
    // logs a benign but noisy "play() request was interrupted" AbortError.
    if (phase === 'spin') {
      spin.play()
      return spin.stop
    }
    if (phase === 'land') {
      land.play()
      return land.stop
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps -- spin/land are freshly built each render from stable SharedObjects, not reactive deps; only phase/enabled should restart this effect
  }, [phase, enabled])
}
