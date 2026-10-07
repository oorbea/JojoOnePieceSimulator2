import { AsyncStorage } from '@/shared/lib/async-storage'

// The store pulls in the real API client, whose module reads validated env at
// import time - stubbed like the game socket store test does.
jest.mock('@/shared/config/env', () => ({
  env: {
    EXPO_PUBLIC_API_URL: 'http://localhost/api/v1',
    EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID: 'test-client-id',
    EXPO_PUBLIC_BUILD_ID: 'test',
  },
}))

import type { PushApi, PushBrowser, PushSubscriptionLike } from '../../lib/push-controller'
import { canTogglePush, createPushStore, shouldShowPushPrompt } from '../push.store'

function makeSub(): PushSubscriptionLike {
  return {
    endpoint: 'https://fcm.googleapis.com/fcm/send/abc',
    toJSON: () => ({ keys: { p256dh: 'k', auth: 'a' } }),
    unsubscribe: jest.fn().mockResolvedValue(true),
  }
}

function setup(
  options: {
    permission?: NotificationPermission
    requestResult?: NotificationPermission
    existing?: PushSubscriptionLike | null
    enabled?: boolean
  } = {}
) {
  let permission = options.permission ?? 'default'
  const sub = makeSub()
  let stored: PushSubscriptionLike | null = options.existing ?? null
  const browser: PushBrowser = {
    supported: () => true,
    permission: () => permission,
    requestPermission: jest.fn(async () => {
      permission = options.requestResult ?? 'granted'
      return permission
    }),
    getSubscription: async () => stored,
    subscribe: jest.fn(async () => {
      stored = sub
      return sub
    }),
  }
  const api: PushApi = {
    getConfig: jest.fn().mockResolvedValue({ enabled: options.enabled ?? true, publicKey: 'AQAB' }),
    register: jest.fn().mockResolvedValue(undefined),
    remove: jest.fn().mockResolvedValue(undefined),
  }
  return { store: createPushStore({ browser, api }), browser, api, sub }
}

describe('push store', () => {
  beforeEach(async () => {
    await AsyncStorage.clear()
  })

  it('init loads the snapshot and the persisted dismissal', async () => {
    await AsyncStorage.setItem('jops.push-prompt-dismissed', 'true')
    const { store } = setup()

    await store.getState().init()

    expect(store.getState()).toMatchObject({
      isLoaded: true,
      supported: true,
      serverEnabled: true,
      permission: 'default',
      subscribed: false,
      promptDismissed: true,
    })
  })

  it('enable subscribes and flips the state', async () => {
    const { store, api } = setup()
    await store.getState().init()

    const failure = await store.getState().enable()

    expect(failure).toBeNull()
    expect(store.getState()).toMatchObject({ subscribed: true, permission: 'granted', busy: false })
    expect(api.register).toHaveBeenCalledTimes(1)
  })

  it('enable reports a blocked permission and remembers it', async () => {
    const { store } = setup({ permission: 'default', requestResult: 'denied' })
    await store.getState().init()

    const failure = await store.getState().enable()

    expect(failure).toBe('denied')
    expect(store.getState()).toMatchObject({ subscribed: false, permission: 'denied', busy: false })
  })

  it('enable reports a failed registration without claiming success', async () => {
    const { store, api } = setup({ permission: 'granted' })
    ;(api.register as jest.Mock).mockRejectedValue(new Error('500'))
    await store.getState().init()

    expect(await store.getState().enable()).toBe('failed')
    expect(store.getState().subscribed).toBe(false)
  })

  it('disable unsubscribes the device', async () => {
    const existing = makeSub()
    const { store, api } = setup({ permission: 'granted', existing })
    await store.getState().init()
    expect(store.getState().subscribed).toBe(true)

    await store.getState().disable()

    expect(store.getState().subscribed).toBe(false)
    expect(api.remove).toHaveBeenCalledWith(existing.endpoint)
    expect(existing.unsubscribe).toHaveBeenCalled()
  })

  it('dismissPrompt persists across a reload', async () => {
    const first = setup()
    await first.store.getState().init()
    await first.store.getState().dismissPrompt()

    const second = setup()
    await second.store.getState().init()
    expect(second.store.getState().promptDismissed).toBe(true)
  })

  it('reset forgets the account-specific state', async () => {
    const { store } = setup({ permission: 'granted', existing: makeSub() })
    await store.getState().init()

    store.getState().reset()

    expect(store.getState()).toMatchObject({ isLoaded: false, subscribed: false, supported: false })
  })
})

describe('shouldShowPushPrompt / canTogglePush', () => {
  const ready = {
    isLoaded: true,
    supported: true,
    serverEnabled: true,
    subscribed: false,
    permission: 'default' as const,
    promptDismissed: false,
  }

  it('prompts only when push can work and nothing has been decided', () => {
    expect(shouldShowPushPrompt(ready)).toBe(true)
    expect(shouldShowPushPrompt({ ...ready, isLoaded: false })).toBe(false)
    expect(shouldShowPushPrompt({ ...ready, supported: false })).toBe(false)
    expect(shouldShowPushPrompt({ ...ready, serverEnabled: false })).toBe(false)
    expect(shouldShowPushPrompt({ ...ready, subscribed: true })).toBe(false)
    expect(shouldShowPushPrompt({ ...ready, permission: 'denied' })).toBe(false)
    expect(shouldShowPushPrompt({ ...ready, promptDismissed: true })).toBe(false)
  })

  it('still prompts when permission was granted but this device is not subscribed', () => {
    expect(shouldShowPushPrompt({ ...ready, permission: 'granted' })).toBe(true)
  })

  it('offers the profile toggle whenever push can work here, even if blocked', () => {
    expect(canTogglePush(ready)).toBe(true)
    expect(canTogglePush({ ...ready, serverEnabled: false })).toBe(false)
    expect(canTogglePush({ ...ready, supported: false })).toBe(false)
    expect(canTogglePush({ ...ready, isLoaded: false })).toBe(false)
  })
})
