import { readFileSync } from 'fs'
import { join } from 'path'

// Exercises the REAL public/sw.js (loaded from disk and evaluated against
// fakes), not a copy of its logic under src/ - the file is served verbatim
// to the browser as a plain classic service worker (no imports/exports),
// so it can't be turned into a module and imported normally. `new
// Function('self', 'caches', 'fetch', src)` shadows the three globals it
// touches without needing a fresh vm realm (which would lack `URL`/shared
// `Promise` identity, breaking `instanceof`/`.rejects`) and without
// assigning real jsdom globals (jsdom's `window.self`/`location` aren't
// assignable). URL/Promise still come from the real jsdom realm this file
// runs in, which is all sw.js needs beyond the three injected globals.
const SW_PATH = join(__dirname, '..', '..', '..', 'public', 'sw.js')
const SW_SRC = readFileSync(SW_PATH, 'utf8')
const ORIGIN = 'https://app.test'

interface FakeSelf {
  addEventListener: (type: string, handler: (event: FakeEvent) => void) => void
  skipWaiting: () => void
  clients: { claim: () => void }
  location: { origin: string }
}

interface FakeRequest {
  url: string
  method: string
  mode?: string
  headers: { get: (name: string) => string | null }
}

interface FakeResponse {
  ok: boolean
  tag: string
  clone: () => FakeResponse
}

interface FakeEvent {
  request: FakeRequest
  respondWith: (p: Promise<FakeResponse>) => void
  waitUntil: (p: Promise<unknown>) => void
}

interface FakeCache {
  keys: () => Promise<{ url: string }[]>
  match: (request: FakeRequest) => Promise<FakeResponse | undefined>
  put: (request: FakeRequest, response: FakeResponse) => Promise<void>
  delete: (requestOrKey: FakeRequest | { url: string }) => Promise<boolean>
  addAll: (urls: string[]) => Promise<void>
}

interface FakeCacheStorage {
  open: (name: string) => Promise<FakeCache>
  keys: () => Promise<string[]>
  match: (request: FakeRequest) => Promise<FakeResponse | undefined>
  delete: (name: string) => Promise<boolean>
}

type FetchImpl = (request: FakeRequest) => Promise<FakeResponse>
type SwFactory = (self: FakeSelf, caches: FakeCacheStorage, fetch: FetchImpl) => void

function req(path: string, init: Partial<FakeRequest> = {}): FakeRequest {
  const headers = init.headers ?? { get: () => null }
  return {
    url: path.startsWith('http') ? path : ORIGIN + path,
    method: init.method ?? 'GET',
    mode: init.mode,
    headers,
  }
}

function res(tag: string, ok = true): FakeResponse {
  return { ok, tag, clone: () => res(tag, ok) }
}

// toEqual can't structurally compare the `clone` function, so assertions on
// a returned FakeResponse compare the identifying fields instead.
function expectResponse(actual: FakeResponse | undefined, tag: string, ok = true) {
  expect(actual).toEqual(expect.objectContaining({ ok, tag }))
}

// Map preserves insertion order, matching what trimImageCache relies on
// (caches.keys() returns entries in insertion order on the real Cache API).
function makeCache(): FakeCache {
  const store = new Map<string, FakeResponse>()
  return {
    keys: async () => [...store.keys()].map((url) => ({ url })),
    match: async (request) => store.get(request.url),
    put: async (request, response) => {
      store.set(request.url, response)
    },
    // trimImageCache passes back a key object from keys(), not the
    // original request - delete must key off `.url`, not identity.
    delete: async (requestOrKey) => store.delete(requestOrKey.url),
    addAll: async (urls) => {
      for (const url of urls) {
        store.set(url.startsWith('http') ? url : ORIGIN + url, res(url))
      }
    },
  }
}

function makeCaches(): FakeCacheStorage {
  const named = new Map<string, FakeCache>()
  return {
    open: async (name) => {
      let cache = named.get(name)
      if (!cache) {
        cache = makeCache()
        named.set(name, cache)
      }
      return cache
    },
    keys: async () => [...named.keys()],
    match: async (request) => {
      for (const cache of named.values()) {
        const hit = await cache.match(request)
        if (hit) return hit
      }
      return undefined
    },
    delete: async (name) => named.delete(name),
  }
}

async function flush() {
  await new Promise((resolve) => setTimeout(resolve, 0))
}

function loadSw(caches: FakeCacheStorage, fetchImpl: FetchImpl) {
  const handlers: Record<string, (event: FakeEvent) => void> = {}
  const self: FakeSelf = {
    addEventListener: (type, handler) => {
      handlers[type] = handler
    },
    skipWaiting: jest.fn(),
    clients: { claim: jest.fn() },
    location: { origin: ORIGIN },
  }
  const factory = new Function('self', 'caches', 'fetch', SW_SRC) as unknown as SwFactory
  factory(self, caches, fetchImpl)
  return { handlers }
}

async function dispatchFetch(
  handlers: Record<string, (event: FakeEvent) => void>,
  request: FakeRequest
) {
  let respondWithPromise: Promise<FakeResponse> | undefined
  const event: FakeEvent = {
    request,
    respondWith: (p) => {
      respondWithPromise = p
    },
    waitUntil: () => {},
  }
  handlers.fetch(event)
  return respondWithPromise
}

async function dispatchLifecycle(
  handlers: Record<string, (event: FakeEvent) => void>,
  type: 'install' | 'activate'
) {
  const waited: Promise<unknown>[] = []
  const event: FakeEvent = {
    request: req('/'),
    respondWith: () => {},
    waitUntil: (p) => {
      waited.push(p)
    },
  }
  handlers[type](event)
  await Promise.all(waited)
}

it('stays a plain classic worker (no import/export/require)', () => {
  expect(SW_SRC).not.toMatch(/^\s*import /m)
  expect(SW_SRC).not.toMatch(/^\s*export /m)
  expect(SW_SRC).not.toMatch(/require\(/)
})

it('install seeds the shell cache with the shell URLs', async () => {
  const caches = makeCaches()
  const fetchImpl = jest.fn<Promise<FakeResponse>, [FakeRequest]>()
  const { handlers } = loadSw(caches, fetchImpl)

  await dispatchLifecycle(handlers, 'install')

  const names = await caches.keys()
  expect(names).toHaveLength(1)
  const shellCache = await caches.open(names[0])
  const keys = await shellCache.keys()
  expect(keys.map((k) => k.url).sort()).toEqual([ORIGIN + '/', ORIGIN + '/manifest.json'])
})

it('activate preserves only the two live cache names and deletes the rest', async () => {
  const caches = makeCaches()
  const fetchImpl: FetchImpl = async (request) => res(request.url)
  const { handlers } = loadSw(caches, fetchImpl)

  // Discover the real shell/image cache names from behavior, not hardcoded
  // literals - so a future CACHE_NAME bump doesn't need this test edited.
  await dispatchLifecycle(handlers, 'install')
  const [shellName] = await caches.keys()
  await dispatchFetch(handlers, req('/api/v1/media/abc123/thumb.webp'))
  await flush()
  const imageName = (await caches.keys()).find((n) => n !== shellName)
  expect(imageName).toBeDefined()

  await caches.open('jops-shell-v3')
  await caches.open('unrelated-cache')

  await dispatchLifecycle(handlers, 'activate')

  const survivors = await caches.keys()
  expect(new Set(survivors)).toEqual(new Set([shellName, imageName]))
})

describe('media proxy routing', () => {
  it('does not intercept the private/signed scope (/api/v1/media/p/...)', async () => {
    const caches = makeCaches()
    const fetchImpl = jest.fn<Promise<FakeResponse>, [FakeRequest]>()
    const { handlers } = loadSw(caches, fetchImpl)

    const withAccept = req('/api/v1/media/p/1893456000/sig123/group1/thumb.webp', {
      headers: { get: (n) => (n === 'accept' ? 'image/webp,*/*' : null) },
    })
    const respondedA = await dispatchFetch(handlers, withAccept)
    expect(respondedA).toBeUndefined()

    const noAccept = req('/api/v1/media/p/1893456000/sig123/group1/thumb.webp')
    const respondedB = await dispatchFetch(handlers, noAccept)
    expect(respondedB).toBeUndefined()

    expect(fetchImpl).not.toHaveBeenCalled()
  })

  it('cache-firsts the public scope: miss fetches and caches, hit skips fetch', async () => {
    const caches = makeCaches()
    const fetchImpl = jest.fn<Promise<FakeResponse>, [FakeRequest]>(async (request) =>
      res(request.url)
    )
    const { handlers } = loadSw(caches, fetchImpl)

    const request = req('/api/v1/media/group1/card.webp')
    const missResponse = await dispatchFetch(handlers, request)
    expectResponse(missResponse, request.url)
    expect(fetchImpl).toHaveBeenCalledTimes(1)
    await flush()

    fetchImpl.mockClear()
    const hitResponse = await dispatchFetch(handlers, request)
    expectResponse(hitResponse, request.url)
    expect(fetchImpl).not.toHaveBeenCalled()
  })
})

describe('fetch handler guards', () => {
  it('ignores cross-origin GETs', async () => {
    const caches = makeCaches()
    const fetchImpl = jest.fn<Promise<FakeResponse>, [FakeRequest]>()
    const { handlers } = loadSw(caches, fetchImpl)

    const responded = await dispatchFetch(
      handlers,
      req('https://other.test/api/v1/media/g/thumb.webp')
    )
    expect(responded).toBeUndefined()
    expect(fetchImpl).not.toHaveBeenCalled()
  })

  it('ignores non-GET requests', async () => {
    const caches = makeCaches()
    const fetchImpl = jest.fn<Promise<FakeResponse>, [FakeRequest]>()
    const { handlers } = loadSw(caches, fetchImpl)

    const responded = await dispatchFetch(
      handlers,
      req('/api/v1/media/g/thumb.webp', { method: 'POST' })
    )
    expect(responded).toBeUndefined()
    expect(fetchImpl).not.toHaveBeenCalled()
  })
})

it('caches hashed static assets separately from media, under the shell cache', async () => {
  const caches = makeCaches()
  const fetchImpl: FetchImpl = async (request) => res(request.url)
  const { handlers } = loadSw(caches, fetchImpl)

  await dispatchLifecycle(handlers, 'install')
  const [shellName] = await caches.keys()

  await dispatchFetch(handlers, req('/_expo/static/js/web-abc123.js'))
  await flush()

  const shellCache = await caches.open(shellName)
  const keys = await shellCache.keys()
  expect(keys.some((k) => k.url === ORIGIN + '/_expo/static/js/web-abc123.js')).toBe(true)
})

it('trims the image cache to its cap by insertion order', async () => {
  const capMatch = SW_SRC.match(/IMG_CACHE_MAX_ENTRIES = (\d+)/)
  expect(capMatch).not.toBeNull()
  const cap = Number(capMatch![1])

  const caches = makeCaches()
  const fetchImpl: FetchImpl = async (request) => res(request.url)
  const { handlers } = loadSw(caches, fetchImpl)

  for (let i = 0; i < cap + 5; i++) {
    await dispatchFetch(handlers, req(`/api/v1/media/group${i}/thumb.webp`))
    await flush()
  }

  // install() was never dispatched in this test, so cacheFirstImage's
  // caches.open is the only call that ever names a cache - whatever name
  // exists at this point is the image cache.
  const [imageName] = await caches.keys()
  expect(imageName).toBeDefined()
  const imageCache = await caches.open(imageName)
  const keys = await imageCache.keys()

  expect(keys).toHaveLength(cap)
  expect(keys.some((k) => k.url.includes('group0/'))).toBe(false)
  expect(keys.some((k) => k.url.includes(`group${cap + 4}/`))).toBe(true)
})

describe('navigation fallback (networkFirst)', () => {
  it('serves and caches a successful network response', async () => {
    const caches = makeCaches()
    const fetchImpl: FetchImpl = async (request) => res(request.url)
    const { handlers } = loadSw(caches, fetchImpl)

    const request = req('/', { mode: 'navigate' })
    const response = await dispatchFetch(handlers, request)
    expectResponse(response, ORIGIN + '/')
  })

  it('falls back to the cached shell when the network fails', async () => {
    const caches = makeCaches()
    let fail = false
    const fetchImpl: FetchImpl = async (request) => {
      if (fail) throw new Error('offline')
      return res(request.url)
    }
    const { handlers } = loadSw(caches, fetchImpl)

    await dispatchFetch(handlers, req('/', { mode: 'navigate' }))
    fail = true
    const response = await dispatchFetch(handlers, req('/', { mode: 'navigate' }))
    expectResponse(response, ORIGIN + '/')
  })

  it('rethrows when the network fails with nothing cached', async () => {
    const caches = makeCaches()
    const fetchImpl: FetchImpl = async () => {
      throw new Error('offline')
    }
    const { handlers } = loadSw(caches, fetchImpl)

    await expect(dispatchFetch(handlers, req('/', { mode: 'navigate' }))).rejects.toThrow('offline')
  })
})
