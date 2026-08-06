import * as Localization from 'expo-localization'
import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'

import caES from '@/shared/i18n/locales/ca-ES.json'
import enGB from '@/shared/i18n/locales/en-GB.json'
import esES from '@/shared/i18n/locales/es-ES.json'
import type { Locale } from '@/shared/lib/zod'

// Every supported locale, in the same order as the backend's
// enums.Locales() (apps/backend .../domain/enums/locale.go).
export const SUPPORTED_LOCALES: Locale[] = ['en-GB', 'es-ES', 'ca-ES']
export const DEFAULT_LOCALE: Locale = 'en-GB'

const resources = {
  'en-GB': enGB,
  'es-ES': esES,
  'ca-ES': caES,
}

// detectDeviceLocale maps the device's language tags onto a supported
// Locale, matching language-only prefixes too (e.g. "es-MX" -> "es-ES",
// mirroring the backend's parseAcceptLanguage (apps/backend
// .../api/endpoints/locale.go) so device detection and server-side
// fallback agree on the same rule.
export function detectDeviceLocale(): Locale {
  for (const { languageTag, languageCode } of Localization.getLocales()) {
    if (SUPPORTED_LOCALES.includes(languageTag as Locale)) {
      return languageTag as Locale
    }
    if (languageCode === 'es') return 'es-ES'
    if (languageCode === 'ca') return 'ca-ES'
    if (languageCode === 'en') return 'en-GB'
  }
  return DEFAULT_LOCALE
}

let initialized = false

// initI18n is idempotent - safe to call more than once (e.g. once eagerly
// below with the default locale, then again from language.store.ts's
// hydrate() once AsyncStorage/the session is read) without re-registering
// resources or throwing.
export function initI18n(initialLocale: Locale) {
  if (initialized) return i18n
  initialized = true
  // eslint-disable-next-line import/no-named-as-default-member -- default import is correct; i18next's named `use` export is unrelated (same reasoning as shared/api/client.ts's axios import)
  void i18n.use(initReactI18next).init({
    resources,
    lng: initialLocale,
    fallbackLng: DEFAULT_LOCALE,
    interpolation: { escapeValue: false }, // React already escapes.
    returnNull: false,
  })
  return i18n
}

// Initialized synchronously at import time, with the default locale, so
// every component tree that imports anything from this module (directly or
// via language.store.ts) can call useTranslation() from its very first
// render - hydrate() reading the real preference is necessarily async
// (AsyncStorage/expo-localization), but the app must never render before
// i18next itself is ready. hydrate() then either confirms this default or
// calls i18n.changeLanguage() to the real one.
initI18n(DEFAULT_LOCALE)

export default i18n
