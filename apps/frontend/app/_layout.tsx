import '../tamagui-web.css'

import { Slot } from 'expo-router'
import { StatusBar } from 'expo-status-bar'
import { useEffect } from 'react'
import { Platform, useColorScheme } from 'react-native'

import { AppProviders } from '@/providers/app-providers'
import { useLanguageStore } from '@/shared/stores/language.store'
import { useSessionStore } from '@/shared/stores/session.store'
import { useThemeStore } from '@/shared/stores/theme.store'

function registerServiceWorker() {
  if (Platform.OS !== 'web') return
  if (typeof navigator === 'undefined' || !('serviceWorker' in navigator)) return

  navigator.serviceWorker.register('/sw.js').catch((error) => {
    console.warn('Service worker registration failed', error)
  })
}

function ThemedStatusBar() {
  const scheme = useColorScheme()
  const mode = useThemeStore((state) => state.mode)
  const resolved = mode === 'system' ? (scheme ?? 'light') : mode
  return <StatusBar style={resolved === 'dark' ? 'light' : 'dark'} />
}

export default function RootLayout() {
  const hydrate = useSessionStore((state) => state.hydrate)
  const hydrateTheme = useThemeStore((state) => state.hydrate)
  const hydrateLanguage = useLanguageStore((state) => state.hydrate)
  const sessionLanguage = useSessionStore((state) => state.session?.user.language)
  const storeLocale = useLanguageStore((state) => state.locale)
  const setLocale = useLanguageStore((state) => state.setLocale)

  useEffect(() => {
    void hydrate()
    void hydrateTheme()
    void hydrateLanguage()
    registerServiceWorker()
  }, [hydrate, hydrateTheme, hydrateLanguage])

  // Once a session is known, the backend's users.language is the source of
  // truth and overrides whatever was device-detected/stored pre-login - see
  // language.store.ts.
  useEffect(() => {
    if (sessionLanguage && sessionLanguage !== storeLocale) {
      void setLocale(sessionLanguage)
    }
  }, [sessionLanguage, storeLocale, setLocale])

  return (
    <AppProviders>
      <ThemedStatusBar />
      <Slot />
    </AppProviders>
  )
}
