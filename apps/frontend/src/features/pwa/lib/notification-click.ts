// What a notification tap sends to the open app (public/sw.js's
// notificationclick): {type: 'NOTIFICATION_CLICK', url}. The path is routed
// with the app's own router so the running session and game socket survive.
// Returns the in-app path to open, or null for anything else - including a
// url that is not a same-app path, which must never be navigated to.
export function parseNotificationClick(data: unknown): string | null {
  if (typeof data !== 'object' || data === null) return null
  const { type, url } = data as { type?: unknown; url?: unknown }
  if (type !== 'NOTIFICATION_CLICK') return null
  if (typeof url !== 'string' || !url.startsWith('/') || url.startsWith('//')) return null
  return url
}
