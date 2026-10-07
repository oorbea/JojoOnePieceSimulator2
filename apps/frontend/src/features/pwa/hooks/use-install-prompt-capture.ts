import { useEffect } from 'react'
import { Platform } from 'react-native'

import { isStandaloneDisplay } from '@/shared/lib/pwa-display'
import {
  usePwaInstallStore,
  type BeforeInstallPromptEvent,
} from '@/shared/stores/pwa-install.store'

// Wires the browser's install events into the shared pwa-install store the
// app shell reads: captures the one-shot install prompt (and suppresses
// Chrome's own mini-infobar so the app's button/banner is the single entry
// point), and tracks whether the app is already running installed. Web only.
export function useInstallPromptCapture() {
  useEffect(() => {
    if (Platform.OS !== 'web') return

    const { hydrate, setDeferred, setInstalled } = usePwaInstallStore.getState()
    void hydrate()
    setInstalled(isStandaloneDisplay())

    const onBeforeInstallPrompt = (event: Event) => {
      event.preventDefault()
      setDeferred(event as BeforeInstallPromptEvent)
    }
    const onInstalled = () => setInstalled(true)

    window.addEventListener('beforeinstallprompt', onBeforeInstallPrompt)
    window.addEventListener('appinstalled', onInstalled)
    return () => {
      window.removeEventListener('beforeinstallprompt', onBeforeInstallPrompt)
      window.removeEventListener('appinstalled', onInstalled)
    }
  }, [])
}
