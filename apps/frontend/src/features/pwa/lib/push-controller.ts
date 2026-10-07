import type { PushConfigResponse, PushSubscribeRequest } from '@/shared/contracts/dto'

// The Web Push flow, apart from React and the real browser: every browser and
// network capability arrives through `PushBrowser` / `PushApi`, so each branch
// (denied, dismissed, already subscribed, offline server...) is unit-testable.

export type PushPermission = NotificationPermission | 'unsupported'

export type PushSubscriptionLike = {
  endpoint: string
  toJSON: () => { endpoint?: string; keys?: Record<string, string> }
  unsubscribe: () => Promise<boolean>
}

export type PushBrowser = {
  supported: () => boolean
  permission: () => NotificationPermission
  requestPermission: () => Promise<NotificationPermission>
  getSubscription: () => Promise<PushSubscriptionLike | null>
  subscribe: (applicationServerKey: Uint8Array) => Promise<PushSubscriptionLike>
}

export type PushApi = {
  getConfig: () => Promise<PushConfigResponse>
  register: (body: PushSubscribeRequest) => Promise<void>
  remove: (endpoint: string) => Promise<void>
}

export type PushStatus = {
  /** The browser can do Web Push at all (service worker + PushManager). */
  supported: boolean
  /** The server has VAPID keys configured. */
  serverEnabled: boolean
  publicKey: string
  permission: PushPermission
  /** This device has a live subscription for the signed-in account. */
  subscribed: boolean
}

export const UNSUPPORTED_STATUS: PushStatus = {
  supported: false,
  serverEnabled: false,
  publicKey: '',
  permission: 'unsupported',
  subscribed: false,
}

// Web Push wants the VAPID key as bytes; the server hands it over base64url.
export function urlBase64ToUint8Array(value: string): Uint8Array {
  const padding = '='.repeat((4 - (value.length % 4)) % 4)
  const base64 = (value + padding).replace(/-/g, '+').replace(/_/g, '/')
  const raw = atob(base64)
  const bytes = new Uint8Array(raw.length)
  for (let i = 0; i < raw.length; i++) bytes[i] = raw.charCodeAt(i)
  return bytes
}

// PushSubscription.toJSON() marks every field optional; the backend needs all
// three.
export function toSubscribeRequest(sub: PushSubscriptionLike): PushSubscribeRequest | null {
  const json = sub.toJSON()
  const endpoint = json.endpoint ?? sub.endpoint
  const p256dh = json.keys?.p256dh
  const auth = json.keys?.auth
  if (!endpoint || !p256dh || !auth) return null
  return { endpoint, keys: { p256dh, auth } }
}

// Reads the current state and, when this device is already subscribed,
// re-registers it with the server (best effort): that refreshes the stored
// keys and re-claims the endpoint if another account used the device last.
export async function loadPushStatus(browser: PushBrowser, api: PushApi): Promise<PushStatus> {
  if (!browser.supported()) return UNSUPPORTED_STATUS

  let config: PushConfigResponse = { enabled: false, publicKey: '' }
  try {
    config = await api.getConfig()
  } catch {
    // Treated as "push unavailable" until the next load; never an error UI.
  }
  const permission = browser.permission()
  if (!config.enabled) {
    return { supported: true, serverEnabled: false, publicKey: '', permission, subscribed: false }
  }

  let subscribed = false
  if (permission === 'granted') {
    const sub = await browser.getSubscription().catch(() => null)
    if (sub) {
      subscribed = true
      const request = toSubscribeRequest(sub)
      if (request) await api.register(request).catch(() => undefined)
    }
  }
  return {
    supported: true,
    serverEnabled: true,
    publicKey: config.publicKey,
    permission,
    subscribed,
  }
}

export type PushEnableFailure = 'unsupported' | 'denied' | 'dismissed' | 'failed'

export class PushEnableError extends Error {
  constructor(
    readonly reason: PushEnableFailure,
    readonly permission: PushPermission
  ) {
    super(`push could not be enabled: ${reason}`)
    this.name = 'PushEnableError'
  }
}

// Asks for permission when it hasn't been decided yet, subscribes this device
// and registers it with the server. Must run from a user gesture (a button
// tap): browsers refuse the permission prompt otherwise.
export async function enablePush(
  browser: PushBrowser,
  api: PushApi,
  publicKey: string
): Promise<{ permission: PushPermission; subscribed: true }> {
  if (!browser.supported()) throw new PushEnableError('unsupported', 'unsupported')

  let permission = browser.permission()
  if (permission === 'default') permission = await browser.requestPermission()
  if (permission === 'denied') throw new PushEnableError('denied', permission)
  // Closing the prompt without choosing leaves the permission at 'default'.
  if (permission !== 'granted') throw new PushEnableError('dismissed', permission)

  try {
    const sub =
      (await browser.getSubscription()) ??
      (await browser.subscribe(urlBase64ToUint8Array(publicKey)))
    const request = toSubscribeRequest(sub)
    if (!request) throw new Error('subscription is missing its keys')
    await api.register(request)
  } catch {
    throw new PushEnableError('failed', permission)
  }
  return { permission, subscribed: true }
}

// Drops this device's subscription on the server and in the browser. Used by
// the profile toggle and on logout. The server call is best effort: even if
// it fails, unsubscribing locally makes the endpoint dead, and the server
// deletes it the next time the push service answers 404/410 for it.
export async function disablePush(browser: PushBrowser, api: PushApi): Promise<void> {
  if (!browser.supported()) return
  const sub = await browser.getSubscription().catch(() => null)
  if (!sub) return
  await api.remove(sub.endpoint).catch(() => undefined)
  await sub.unsubscribe().catch(() => false)
}
