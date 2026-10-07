import { useRouter } from 'expo-router'
import { useEffect } from 'react'
import { Platform } from 'react-native'

import { parseNotificationClick } from '../lib/notification-click'

// Routes a tapped push notification to its game inside the already-open app
// (see public/sw.js's notificationclick, which posts the path here instead of
// reloading the window). No-op off web.
export function useNotificationNavigation() {
  const router = useRouter()

  useEffect(() => {
    if (Platform.OS !== 'web' || typeof navigator === 'undefined' || !('serviceWorker' in navigator)) {
      return
    }
    const onMessage = (event: MessageEvent) => {
      const path = parseNotificationClick(event.data)
      if (path) router.navigate(path as never)
    }
    navigator.serviceWorker.addEventListener('message', onMessage)
    return () => navigator.serviceWorker.removeEventListener('message', onMessage)
  }, [router])
}
