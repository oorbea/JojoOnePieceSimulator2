import { create } from 'zustand'

import { AsyncStorage } from '@/shared/lib/async-storage'

const MUTED_KEY = 'jops.audio.muted'
const VOLUME_KEY = 'jops.audio.volume'

type AudioSettingsState = {
  muted: boolean
  volume: number
  isHydrated: boolean
  hydrate: () => Promise<void>
  toggleMute: () => Promise<void>
  setVolume: (volume: number) => Promise<void>
}

function clampVolume(volume: number): number {
  return Math.min(1, Math.max(0, volume))
}

// The single global audio control the app has never had (see
// ObsidianVault/game-victory-defeat-cinematic-2026-09-14.md) - the victory/
// defeat cinematics are loud on purpose, and nobody should be forced to hear
// them at full volume in an office. Every sound in the app (the sorteo's
// reel ticks included, see use-reveal-spin-sound.ts) now reads through
// use-sound.ts, which is the one place that applies muted/volume.
export const useAudioSettingsStore = create<AudioSettingsState>((set) => ({
  muted: false,
  volume: 1,
  isHydrated: false,

  hydrate: async () => {
    const [rawMuted, rawVolume] = await Promise.all([
      AsyncStorage.getItem(MUTED_KEY),
      AsyncStorage.getItem(VOLUME_KEY),
    ])
    const muted = rawMuted === 'true'
    const parsedVolume = rawVolume === null ? NaN : Number(rawVolume)
    const volume = Number.isFinite(parsedVolume) ? clampVolume(parsedVolume) : 1
    set({ muted, volume, isHydrated: true })
  },

  toggleMute: async () => {
    const muted = !useAudioSettingsStore.getState().muted
    await AsyncStorage.setItem(MUTED_KEY, String(muted))
    set({ muted })
  },

  setVolume: async (volume) => {
    const clamped = clampVolume(volume)
    await AsyncStorage.setItem(VOLUME_KEY, String(clamped))
    set({ volume: clamped })
  },
}))
