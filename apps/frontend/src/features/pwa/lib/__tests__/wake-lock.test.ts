import { createWakeLock, type WakeLockSentinelLike } from '../wake-lock'

function makeSentinel() {
  const releaseListeners: (() => void)[] = []
  const sentinel: WakeLockSentinelLike = {
    release: jest.fn().mockResolvedValue(undefined),
    addEventListener: (_type, listener) => {
      releaseListeners.push(listener)
    },
  }
  return { sentinel, fireRelease: () => releaseListeners.forEach((l) => l()) }
}

function makeEnv() {
  const visibilityListeners = new Set<() => void>()
  const doc = {
    visibilityState: 'visible',
    addEventListener: (_type: 'visibilitychange', l: () => void) => visibilityListeners.add(l),
    removeEventListener: (_type: 'visibilitychange', l: () => void) =>
      visibilityListeners.delete(l),
  }
  const sentinels: ReturnType<typeof makeSentinel>[] = []
  const api = {
    request: jest.fn(async () => {
      const made = makeSentinel()
      sentinels.push(made)
      return made.sentinel
    }),
  }
  const lock = createWakeLock(api, doc)
  return {
    api,
    doc,
    lock,
    sentinels,
    changeVisibility: (state: 'visible' | 'hidden') => {
      doc.visibilityState = state
      visibilityListeners.forEach((l) => l())
    },
    listenerCount: () => visibilityListeners.size,
  }
}

const flush = () => new Promise((resolve) => setTimeout(resolve, 0))

describe('createWakeLock', () => {
  it('requests a screen lock on start', async () => {
    const env = makeEnv()
    env.lock.start()
    await flush()
    expect(env.api.request).toHaveBeenCalledWith('screen')
  })

  it('re-acquires after the browser drops the lock on hide', async () => {
    const env = makeEnv()
    env.lock.start()
    await flush()

    // Page hidden: browser releases the sentinel by itself.
    env.sentinels[0].fireRelease()
    env.changeVisibility('hidden')
    await flush()
    expect(env.api.request).toHaveBeenCalledTimes(1)

    env.changeVisibility('visible')
    await flush()
    expect(env.api.request).toHaveBeenCalledTimes(2)
  })

  it('does not request a second lock while one is held', async () => {
    const env = makeEnv()
    env.lock.start()
    await flush()
    env.changeVisibility('visible')
    await flush()
    expect(env.api.request).toHaveBeenCalledTimes(1)
  })

  it('stop releases the lock and stops listening', async () => {
    const env = makeEnv()
    env.lock.start()
    await flush()

    env.lock.stop()
    expect(env.sentinels[0].sentinel.release).toHaveBeenCalledTimes(1)
    expect(env.listenerCount()).toBe(0)

    env.changeVisibility('visible')
    await flush()
    expect(env.api.request).toHaveBeenCalledTimes(1)
  })

  it('releases a lock that arrives after stop()', async () => {
    const env = makeEnv()
    env.lock.start()
    env.lock.stop()
    await flush()
    expect(env.sentinels[0].sentinel.release).toHaveBeenCalledTimes(1)
  })

  it('survives a denied request and retries on the next visibility change', async () => {
    const env = makeEnv()
    env.api.request.mockRejectedValueOnce(new Error('NotAllowedError'))
    env.lock.start()
    await flush()
    expect(env.sentinels).toHaveLength(0)

    env.changeVisibility('visible')
    await flush()
    expect(env.sentinels).toHaveLength(1)
  })

  it('start is idempotent', async () => {
    const env = makeEnv()
    env.lock.start()
    env.lock.start()
    await flush()
    expect(env.listenerCount()).toBe(1)
    expect(env.api.request).toHaveBeenCalledTimes(1)
  })
})
