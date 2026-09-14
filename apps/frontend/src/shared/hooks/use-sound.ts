import { useEffect } from 'react'
import { useAudioPlayer } from 'expo-audio'
import type { AudioPlayer } from 'expo-audio'

import { useAudioSettingsStore } from '@/shared/stores/audio-settings.store'

export type SoundControls = {
  play: () => void
  stop: () => void
}

// The one place that applies the global mute/volume store
// (audio-settings.store.ts) to an expo-audio player - every sound in the app
// goes through this instead of calling useAudioPlayer directly, so the mute
// toggle actually mutes everything, not just whichever hook remembered to
// check it. Not unit-tested: like the sound hook it replaces
// (use-reveal-spin-sound.ts, its only prior consumer), it's a thin
// side-effecting wrapper around expo-audio's native player with nothing left
// to assert once you subtract the untestable playback call itself.
export function useSound(source: number, enabled: boolean): SoundControls {
  const player: AudioPlayer = useAudioPlayer(source)
  const muted = useAudioSettingsStore((state) => state.muted)
  const volume = useAudioSettingsStore((state) => state.volume)

  useEffect(() => {
    // expo-audio's AudioPlayer is a native SharedObject whose API IS direct
    // property assignment (see its own .d.ts) - eslint's react-hooks/
    // immutability rule doesn't special-case it the way it does Reanimated's
    // useSharedValue, but syncing it here (an effect, not render) is exactly
    // the "update an external system from React state" case that rule
    // exists to steer toward anyway.
    // eslint-disable-next-line react-hooks/immutability -- see comment above
    player.volume = volume

    player.muted = muted
  }, [player, volume, muted])

  return {
    play: () => {
      if (!enabled || muted) return
      player.seekTo(0).catch(() => {})
      player.play()
    },
    stop: () => {
      player.pause()
    },
  }
}
