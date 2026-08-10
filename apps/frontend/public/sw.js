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
const CACHE_NAME = 'jops-shell-v3'
const SHELL_URLS = ['/', '/manifest.json']

self.addEventListener('install', (event) => {
  event.waitUntil(caches.open(CACHE_NAME).then((cache) => cache.addAll(SHELL_URLS)))
  self.skipWaiting()
})

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches
      .keys()
      .then((keys) =>
        Promise.all(keys.filter((key) => key !== CACHE_NAME).map((key) => caches.delete(key)))
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

  if (request.mode === 'navigate' || request.headers.get('accept')?.includes('text/html')) {
    event.respondWith(networkFirst(request))
    return
  }

  // Everything else (manifest, icons, etc.) passes straight through - no
  // opportunistic caching of arbitrary same-origin GETs, which is what let
  // an old index.html linger indefinitely under the previous strategy.
})
