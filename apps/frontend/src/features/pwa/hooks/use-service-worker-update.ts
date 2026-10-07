import { useEffect } from 'react'
import { Platform } from 'react-native'

import { useGameActivityStore } from '@/shared/stores/game-activity.store'

import {
  UPDATE_CHECK_INTERVAL_MS,
  createReloadGuard,
  createUpdateController,
  type ServiceWorkerContainerLike,
} from '../lib/sw-update'

function sessionStorageOrUndefined(): Storage | undefined {
  try {
    return window.sessionStorage
  } catch {
    return undefined
  }
}

// Registers the service worker and keeps the installed PWA on the latest
// deploy: checks for a new build when the app returns to the foreground and
// every UPDATE_CHECK_INTERVAL_MS, activates it and reloads - except while a
// game is in progress, where it waits until the game ends. See
// ../lib/sw-update.ts for the flow. No-op off web.
export function useServiceWorkerUpdate() {
  useEffect(() => {
    if (Platform.OS !== 'web') return
    if (typeof navigator === 'undefined' || !('serviceWorker' in navigator)) return

    const controller = createUpdateController({
      // The DOM's ServiceWorkerContainer satisfies the structural subset
      // the controller needs.
      container: navigator.serviceWorker as unknown as ServiceWorkerContainerLike,
      getActivity: () => useGameActivityStore.getState().activity,
      reload: () => window.location.reload(),
      guard: createReloadGuard(sessionStorageOrUndefined()),
    })

    controller.start().catch((error) => {
      console.warn('Service worker registration failed', error)
    })

    // A deferred update becomes applicable the moment the game ends.
    const unsubscribeActivity = useGameActivityStore.subscribe(controller.flush)

    const onVisibility = () => {
      if (document.visibilityState === 'visible') controller.checkForUpdate()
    }
    document.addEventListener('visibilitychange', onVisibility)
    const interval = setInterval(controller.checkForUpdate, UPDATE_CHECK_INTERVAL_MS)

    return () => {
      controller.dispose()
      unsubscribeActivity()
      document.removeEventListener('visibilitychange', onVisibility)
      clearInterval(interval)
    }
  }, [])
}
