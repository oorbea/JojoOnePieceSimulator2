import '../tamagui-web.css'

import { Slot } from 'expo-router'
import { StatusBar } from 'expo-status-bar'
import { Platform } from 'react-native'
import { useEffect } from 'react'

import { AppProviders } from '@/providers/app-providers'
import { useSessionStore } from '@/shared/stores/session.store'

function registerServiceWorker() {
  if (Platform.OS !== 'web') return
  if (typeof navigator === 'undefined' || !('serviceWorker' in navigator)) return

  navigator.serviceWorker.register('/sw.js').catch((error) => {
    console.warn('Service worker registration failed', error)
  })
}

export default function RootLayout() {
  const hydrate = useSessionStore((state) => state.hydrate)

  useEffect(() => {
    void hydrate()
    registerServiceWorker()
  }, [hydrate])

  return (
    <AppProviders>
      <StatusBar style="auto" />
      <Slot />
    </AppProviders>
  )
}
