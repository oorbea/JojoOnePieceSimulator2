// Network-first for navigations/HTML, cache-first only for hashed immutable
// static assets. Bump CACHE_NAME on every deploy that changes this file's
// own logic so stale caches get evicted instead of serving an old shell
// forever - but the network-first navigation strategy below is what makes
// every *other* deploy (i.e. one that doesn't touch this file) self-healing
// too: a stale service worker still fetches the new index.html from the
// network on every load, it just never learns about the new deploy from its
// own cache the way the old cache-first-everything strategy did.
//
// History: the previous version cached every same-origin GET (including
// index.html and the JS bundle) cache-first with no revalidation, so once a
// browser had visited the site it kept being served whatever index.html/
// bundle existed at that visit, forever - admin CRUD writes worked (backend
// cache invalidation was fine) but the browser was still running pre-fix JS
// until the user manually cleared site data. See ObsidianVault's
// admin-crud-cache-stale-sw.md for the investigation.
const CACHE_NAME = 'jops-shell-v4'
const SHELL_URLS = ['/', '/manifest.json']

// Separate cache for the content-addressed media proxy (T2 -
// media-proxy-content-addressed.md) - a group id is immutable by
// construction (it's derived from the rendition's own bytes), so this is a
// real cache-first win, unlike the shell/bundle above. Kept in its own
// namespace, distinct from CACHE_NAME, specifically so bumping CACHE_NAME on
// a shell/logic change never evicts it - see the `activate` handler below,
// which must special-case both names or every deploy would silently wipe
// every visitor's already-downloaded catalogue images.
const IMG_CACHE_NAME = 'jops-img-v1'
// No per-entry TTL is needed (immutable content never goes stale), but an
// unbounded cache would grow forever for a long-lived tab/PWA install - a
// simple insertion-order cap bounds it without needing real LRU bookkeeping,
// same trade-off shared/api/etag.ts's rememberResponse cap makes.
const IMG_CACHE_MAX_ENTRIES = 200

self.addEventListener('install', (event) => {
  event.waitUntil(caches.open(CACHE_NAME).then((cache) => cache.addAll(SHELL_URLS)))
  self.skipWaiting()
})

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches
      .keys()
      .then((keys) =>
        Promise.all(
          keys
            .filter((key) => key !== CACHE_NAME && key !== IMG_CACHE_NAME)
            .map((key) => caches.delete(key))
        )
      )
  )
  self.clients.claim()
})

// Hashed build output (apps/frontend/dist/_expo/static/...) - the filename
// itself changes on any content change, so serving a cached copy forever is
// safe. Mirrors nginx.frontend.conf's `expires 1y; Cache-Control: public,
// immutable` rule for the same path.
function isImmutableStaticAsset(url) {
  return url.origin === self.location.origin && url.pathname.startsWith('/_expo/static/')
}

// The content-addressed media proxy (only same-origin in prod, where NPM
// serves both `/` and `/api` off one host - see
// media-proxy-content-addressed.md's MEDIA_BASE_URL note). Local dev's
// absolute cross-origin MEDIA_BASE_URL never reaches here: the fetch
// handler below already returns early for any cross-origin request.
function isImmutableImage(url) {
  return url.origin === self.location.origin && url.pathname.startsWith('/api/v1/media/')
}

// Cache-first: instant/offline, correct because the URL is content-hashed.
async function cacheFirst(request) {
  const cached = await caches.match(request)
  if (cached) return cached

  const response = await fetch(request)
  if (response.ok) {
    const cache = await caches.open(CACHE_NAME)
    void cache.put(request, response.clone())
  }
  return response
}

// Same cache-first shape as cacheFirst above, but against IMG_CACHE_NAME
// (never evicted by a shell-only CACHE_NAME bump) and with an insertion-order
// cap so a long-lived session's catalogue browsing doesn't grow the cache
// without bound.
async function cacheFirstImage(request) {
  const cache = await caches.open(IMG_CACHE_NAME)
  const cached = await cache.match(request)
  if (cached) return cached

  const response = await fetch(request)
  if (response.ok) {
    void cache.put(request, response.clone()).then(() => trimImageCache(cache))
  }
  return response
}

// caches.keys() returns entries in insertion order, so the oldest excess
// entries are simply the first ones - cheap enough for a cap in the low
// hundreds, and correct without tracking last-access time.
async function trimImageCache(cache) {
  const keys = await cache.keys()
  const excess = keys.length - IMG_CACHE_MAX_ENTRIES
  if (excess <= 0) return
  await Promise.all(keys.slice(0, excess).map((key) => cache.delete(key)))
}

// Network-first: always tries to get the latest deploy's HTML, only falling
// back to whatever shell is cached when the network is unreachable (true
// offline, or the dev-only first-load race documented below).
async function networkFirst(request) {
  try {
    const response = await fetch(request)
    if (response.ok) {
      const cache = await caches.open(CACHE_NAME)
      void cache.put(request, response.clone())
    }
    return response
  } catch (error) {
    const cached = await caches.match(request)
    if (cached) return cached
    // A rejected fetch with nothing cached yet (offline on a first-ever
    // visit, or the browser cancelling this SW-side request in favor of the
    // real navigation's own fetch - a known race right after the SW
    // installs) must not surface as an unhandled rejection.
    throw error
  }
}

self.addEventListener('fetch', (event) => {
  const { request } = event
  // Only ever intercept same-origin GETs - API calls to the backend (a
  // different origin) must always hit the network untouched so auth/ETag
  // handling stays correct, and only GET is idempotent/cacheable.
  const url = new URL(request.url)
  if (request.method !== 'GET' || url.origin !== self.location.origin) {
    return
  }

  if (isImmutableStaticAsset(url)) {
    event.respondWith(cacheFirst(request))
    return
  }

  if (isImmutableImage(url)) {
    event.respondWith(cacheFirstImage(request))
    return
  }

  if (request.mode === 'navigate' || request.headers.get('accept')?.includes('text/html')) {
    event.respondWith(networkFirst(request))
    return
  }

  // Everything else (manifest, icons, etc.) passes straight through - no
  // opportunistic caching of arbitrary same-origin GETs, which is what let
  // an old index.html linger indefinitely under the previous strategy.
})
