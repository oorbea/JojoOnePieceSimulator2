import {
  PushEnableError,
  UNSUPPORTED_STATUS,
  disablePush,
  enablePush,
  loadPushStatus,
  toSubscribeRequest,
  urlBase64ToUint8Array,
  type PushApi,
  type PushBrowser,
  type PushSubscriptionLike,
} from '../push-controller'

function makeSub(overrides: Partial<ReturnType<PushSubscriptionLike['toJSON']>> = {}) {
  const sub: PushSubscriptionLike & { unsubscribe: jest.Mock } = {
    endpoint: 'https://fcm.googleapis.com/fcm/send/abc',
    toJSON: () => ({
      endpoint: 'https://fcm.googleapis.com/fcm/send/abc',
      keys: { p256dh: 'p256dh-key', auth: 'auth-secret' },
      ...overrides,
    }),
    unsubscribe: jest.fn().mockResolvedValue(true),
  }
  return sub
}

function makeBrowser(
  options: {
    supported?: boolean
    permission?: NotificationPermission
    requestResult?: NotificationPermission
    existing?: PushSubscriptionLike | null
  } = {}
) {
  let permission = options.permission ?? 'default'
  const created = makeSub()
  const browser = {
    supported: jest.fn(() => options.supported ?? true),
    permission: jest.fn(() => permission),
    requestPermission: jest.fn(async () => {
      permission = options.requestResult ?? 'granted'
      return permission
    }),
    getSubscription: jest.fn(async () => options.existing ?? null),
    subscribe: jest.fn(async () => created),
  } satisfies PushBrowser
  return { browser, created }
}

function makeApi(config = { enabled: true, publicKey: 'BPublicKey' }) {
  return {
    getConfig: jest.fn().mockResolvedValue(config),
    register: jest.fn().mockResolvedValue(undefined),
    remove: jest.fn().mockResolvedValue(undefined),
  } satisfies PushApi
}

describe('urlBase64ToUint8Array', () => {
  it('decodes unpadded base64url into bytes', () => {
    // "hello?>" -> aGVsbG8_Pg in base64url (no padding, uses _ for /)
    expect(Array.from(urlBase64ToUint8Array('aGVsbG8_Pg'))).toEqual([
      104, 101, 108, 108, 111, 63, 62,
    ])
  })

  it('handles the - and _ alphabet', () => {
    expect(Array.from(urlBase64ToUint8Array('-_8'))).toEqual([251, 255])
  })
})

describe('toSubscribeRequest', () => {
  it('maps a complete subscription', () => {
    expect(toSubscribeRequest(makeSub())).toEqual({
      endpoint: 'https://fcm.googleapis.com/fcm/send/abc',
      keys: { p256dh: 'p256dh-key', auth: 'auth-secret' },
    })
  })

  it('rejects a subscription missing its keys', () => {
    expect(toSubscribeRequest(makeSub({ keys: { p256dh: 'only' } }))).toBeNull()
    expect(toSubscribeRequest(makeSub({ keys: undefined }))).toBeNull()
  })
})

describe('loadPushStatus', () => {
  it('reports an unsupported browser without calling the server', async () => {
    const { browser } = makeBrowser({ supported: false })
    const api = makeApi()

    expect(await loadPushStatus(browser, api)).toEqual(UNSUPPORTED_STATUS)
    expect(api.getConfig).not.toHaveBeenCalled()
  })

  it('reports push disabled when the server has no VAPID keys', async () => {
    const { browser } = makeBrowser()
    const api = makeApi({ enabled: false, publicKey: '' })

    const snapshot = await loadPushStatus(browser, api)

    expect(snapshot).toMatchObject({ supported: true, serverEnabled: false, subscribed: false })
  })

  it('treats an unreachable config endpoint as push unavailable', async () => {
    const { browser } = makeBrowser()
    const api = makeApi()
    api.getConfig.mockRejectedValue(new Error('offline'))

    expect(await loadPushStatus(browser, api)).toMatchObject({ serverEnabled: false })
  })

  it('is not subscribed before permission is granted', async () => {
    const { browser } = makeBrowser({ permission: 'default' })

    expect(await loadPushStatus(browser, makeApi())).toMatchObject({
      serverEnabled: true,
      permission: 'default',
      subscribed: false,
    })
  })

  it('re-registers an existing subscription so the server keeps it fresh and owned', async () => {
    const existing = makeSub()
    const { browser } = makeBrowser({ permission: 'granted', existing })
    const api = makeApi()

    const snapshot = await loadPushStatus(browser, api)

    expect(snapshot).toMatchObject({ subscribed: true, publicKey: 'BPublicKey' })
    expect(api.register).toHaveBeenCalledWith(toSubscribeRequest(existing))
  })

  it('stays subscribed locally when the re-registration fails', async () => {
    const { browser } = makeBrowser({ permission: 'granted', existing: makeSub() })
    const api = makeApi()
    api.register.mockRejectedValue(new Error('offline'))

    expect(await loadPushStatus(browser, api)).toMatchObject({ subscribed: true })
  })

  it('is not subscribed when permission was granted but no subscription exists', async () => {
    const { browser } = makeBrowser({ permission: 'granted', existing: null })

    expect(await loadPushStatus(browser, makeApi())).toMatchObject({ subscribed: false })
  })
})

describe('enablePush', () => {
  it('asks for permission, subscribes with the VAPID key and registers with the server', async () => {
    const { browser, created } = makeBrowser({ permission: 'default', requestResult: 'granted' })
    const api = makeApi()

    const result = await enablePush(browser, api, 'AQAB')

    expect(browser.requestPermission).toHaveBeenCalledTimes(1)
    expect(browser.subscribe).toHaveBeenCalledWith(urlBase64ToUint8Array('AQAB'))
    expect(api.register).toHaveBeenCalledWith(toSubscribeRequest(created))
    expect(result).toEqual({ permission: 'granted', subscribed: true })
  })

  it('does not prompt again when permission is already granted, and reuses a subscription', async () => {
    const existing = makeSub()
    const { browser } = makeBrowser({ permission: 'granted', existing })
    const api = makeApi()

    await enablePush(browser, api, 'AQAB')

    expect(browser.requestPermission).not.toHaveBeenCalled()
    expect(browser.subscribe).not.toHaveBeenCalled()
    expect(api.register).toHaveBeenCalledWith(toSubscribeRequest(existing))
  })

  it('reports a blocked permission without subscribing', async () => {
    const { browser } = makeBrowser({ permission: 'denied' })
    const api = makeApi()

    await expect(enablePush(browser, api, 'AQAB')).rejects.toMatchObject({
      reason: 'denied',
    })
    expect(browser.subscribe).not.toHaveBeenCalled()
    expect(api.register).not.toHaveBeenCalled()
  })

  it('reports a prompt the user denied', async () => {
    const { browser } = makeBrowser({ permission: 'default', requestResult: 'denied' })

    await expect(enablePush(browser, makeApi(), 'AQAB')).rejects.toMatchObject({
      reason: 'denied',
    })
  })

  it('reports a prompt closed without an answer as dismissed', async () => {
    const { browser } = makeBrowser({ permission: 'default', requestResult: 'default' })

    await expect(enablePush(browser, makeApi(), 'AQAB')).rejects.toMatchObject({
      reason: 'dismissed',
    })
  })

  it('reports an unsupported browser', async () => {
    const { browser } = makeBrowser({ supported: false })

    await expect(enablePush(browser, makeApi(), 'AQAB')).rejects.toBeInstanceOf(PushEnableError)
  })

  it('reports a server failure as failed', async () => {
    const { browser } = makeBrowser({ permission: 'granted' })
    const api = makeApi()
    api.register.mockRejectedValue(new Error('500'))

    await expect(enablePush(browser, api, 'AQAB')).rejects.toMatchObject({ reason: 'failed' })
  })
})

describe('disablePush', () => {
  it('removes the subscription on the server and in the browser', async () => {
    const existing = makeSub()
    const { browser } = makeBrowser({ permission: 'granted', existing })
    const api = makeApi()

    await disablePush(browser, api)

    expect(api.remove).toHaveBeenCalledWith('https://fcm.googleapis.com/fcm/send/abc')
    expect(existing.unsubscribe).toHaveBeenCalledTimes(1)
  })

  it('still unsubscribes locally when the server call fails', async () => {
    const existing = makeSub()
    const { browser } = makeBrowser({ permission: 'granted', existing })
    const api = makeApi()
    api.remove.mockRejectedValue(new Error('offline'))

    await disablePush(browser, api)

    expect(existing.unsubscribe).toHaveBeenCalledTimes(1)
  })

  it('is a no-op without a subscription or without support', async () => {
    const api = makeApi()
    await disablePush(makeBrowser({ existing: null }).browser, api)
    await disablePush(makeBrowser({ supported: false }).browser, api)
    expect(api.remove).not.toHaveBeenCalled()
  })
})
