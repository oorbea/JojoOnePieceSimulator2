import { Platform } from 'react-native'

import type { PushBrowser } from './push-controller'

function supported(): boolean {
  return (
    Platform.OS === 'web' &&
    typeof window !== 'undefined' &&
    'serviceWorker' in navigator &&
    'PushManager' in window &&
    'Notification' in window
  )
}

// The real browser behind PushBrowser. `userVisibleOnly: true` is mandatory
// (Chrome rejects subscriptions without it): every push must end in a visible
// notification, which public/sw.js guarantees except while a window is open.
export const webPushBrowser: PushBrowser = {
  supported,
  permission: () => Notification.permission,
  requestPermission: () => Notification.requestPermission(),
  getSubscription: async () => {
    const registration = await navigator.serviceWorker.ready
    return registration.pushManager.getSubscription()
  },
  subscribe: async (applicationServerKey) => {
    const registration = await navigator.serviceWorker.ready
    return registration.pushManager.subscribe({
      userVisibleOnly: true,
      applicationServerKey: applicationServerKey as BufferSource,
    })
  },
}
