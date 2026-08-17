import { useEffect } from 'react'
import { useAudioPlayer } from 'expo-audio'

// Synthesized locally (no external/copyrighted audio) - see the generator
// script this asset was produced from, kept alongside it for reference:
// apps/frontend/assets/audio/reel-spin.wav. A soft looping "whirring" hum
// with evenly-spaced ratchet clicks, ~2.7s and seamlessly loopable.
const REEL_SPIN_SOUND = require('../../../../assets/audio/reel-spin.wav')

// Plays a single continuous loop for as long as the whole sorteo overlay is
// mounted (RevealStage renders exactly while the reveal is in progress -
// see match-screen.tsx - so ITS OWN mount/unmount lifecycle already is
// "the animation is playing", no separate isRevealing prop needed here).
// Gated on `enabled` (callers pass `!reducedMotion`) - reduced motion
// already skips every visual spin/overshoot straight to the resting frame,
// so a spin sound playing over a static reel would be actively misleading,
// not just superfluous. No separate mute/settings toggle - not requested.
export function useRevealSpinSound(enabled: boolean): void {
  const player = useAudioPlayer(REEL_SPIN_SOUND)

  useEffect(() => {
    if (!enabled) return
    // player is expo-audio's SharedObject: a real native handle, not a
    // plain JS value - `.loop`/`.play()`/`.pause()` mutating it in place is
    // exactly its documented API (there's no `loop` AudioPlayerOptions
    // field to set this via useAudioPlayer's constructor instead).
    // eslint-disable-next-line react-hooks/immutability -- see above
    player.loop = true
    player.play()
    return () => player.pause()
    // eslint-disable-next-line react-hooks/exhaustive-deps -- player is a stable SharedObject from useAudioPlayer, not a reactive dep; only `enabled` should restart this effect
  }, [enabled])
}
