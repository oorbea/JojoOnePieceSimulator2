import { useSyncExternalStore } from 'react'
import { Platform } from 'react-native'

const isBrowser = Platform.OS === 'web' && typeof window !== 'undefined'

function subscribe(onChange: () => void) {
  if (!isBrowser) return () => {}
  window.addEventListener('online', onChange)
  window.addEventListener('offline', onChange)
  return () => {
    window.removeEventListener('online', onChange)
    window.removeEventListener('offline', onChange)
  }
}

// Reactive `navigator.onLine`. The browser only knows "no network at all"
// (airplane mode, no signal) - a captive portal or a dead upstream still
// reads online - so treat it as a hint for a banner, not a guarantee that
// requests will succeed. Always "online" off web and during static render.
export function useOnlineStatus(): boolean {
  return useSyncExternalStore(
    subscribe,
    () => (isBrowser ? navigator.onLine : true),
    () => true
  )
}
