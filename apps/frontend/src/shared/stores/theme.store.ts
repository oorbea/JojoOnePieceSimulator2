import { create } from 'zustand'

import { AsyncStorage } from '@/shared/lib/async-storage'

const THEME_STORAGE_KEY = 'jops.theme'

export type ThemeMode = 'system' | 'light' | 'dark'

type ThemeState = {
  mode: ThemeMode
  isHydrated: boolean
  hydrate: () => Promise<void>
  cycle: () => Promise<void>
}

const ORDER: ThemeMode[] = ['system', 'light', 'dark']

// Small, dedicated store rather than folding into session.store.ts — theme
// preference has nothing to do with auth and should survive logout.
export const useThemeStore = create<ThemeState>((set, get) => ({
  mode: 'system',
  isHydrated: false,

  hydrate: async () => {
    const raw = await AsyncStorage.getItem(THEME_STORAGE_KEY)
    const mode = raw === 'light' || raw === 'dark' || raw === 'system' ? raw : 'system'
    set({ mode, isHydrated: true })
  },

  cycle: async () => {
    const next = ORDER[(ORDER.indexOf(get().mode) + 1) % ORDER.length]
    await AsyncStorage.setItem(THEME_STORAGE_KEY, next)
    set({ mode: next })
  },
}))
