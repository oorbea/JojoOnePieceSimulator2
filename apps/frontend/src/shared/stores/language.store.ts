import { create } from 'zustand'

import { AsyncStorage } from '@/shared/lib/async-storage'
import type { Locale } from '@/shared/lib/zod'
import i18n, { DEFAULT_LOCALE, detectDeviceLocale } from '@/shared/i18n'

const LANGUAGE_STORAGE_KEY = 'jops.language'

type LanguageState = {
  locale: Locale
  isHydrated: boolean
  hydrate: () => Promise<void>
  setLocale: (locale: Locale) => Promise<void>
}

// Dedicated store, same reasoning as theme.store.ts: the login screen is
// pre-session and still needs a language before any user is known, and a
// language preference should survive logout. Once a session loads,
// app/_layout.tsx overwrites this with the backend's users.language -
// which then wins over whatever was detected/stored locally.
export const useLanguageStore = create<LanguageState>((set) => ({
  locale: DEFAULT_LOCALE,
  isHydrated: false,

  hydrate: async () => {
    const raw = await AsyncStorage.getItem(LANGUAGE_STORAGE_KEY)
    const locale = isLocale(raw) ? raw : detectDeviceLocale()
    void i18n.changeLanguage(locale)
    set({ locale, isHydrated: true })
  },

  setLocale: async (locale) => {
    await AsyncStorage.setItem(LANGUAGE_STORAGE_KEY, locale)
    await i18n.changeLanguage(locale)
    set({ locale })
  },
}))

function isLocale(value: string | null): value is Locale {
  return value === 'en-GB' || value === 'es-ES' || value === 'ca-ES'
}
