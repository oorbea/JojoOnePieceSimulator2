import '../tamagui-web.css'

import { Slot } from 'expo-router'
import { StatusBar } from 'expo-status-bar'
import { useEffect } from 'react'
import { Platform, useColorScheme } from 'react-native'

import { AppProviders } from '@/providers/app-providers'
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

  useEffect(() => {
    void hydrate()
    void hydrateTheme()
    registerServiceWorker()
  }, [hydrate, hydrateTheme])

  return (
    <AppProviders>
      <ThemedStatusBar />
      <Slot />
    </AppProviders>
  )
}
