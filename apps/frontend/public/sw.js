// Network-first for navigations/HTML, cache-first only for hashed immutable
// static assets. CACHE_NAME is derived from the per-deploy BUILD_ID below, so
// stale shell caches are evicted on every deploy without hand-bumping
// anything - and the network-first navigation strategy is what keeps a
// not-yet-updated service worker self-healing: it still fetches the new
// index.html from the network on every load, it just never learns about the
// new deploy from its own cache the way the old cache-first-everything
// strategy did.
//
// IMG_CACHE_NAME is NOT build-scoped (and is only bumped by hand if its
// contents ever become invalid): excluding the `/p/...` private-media scope
// from isImmutableImage below (2026-09-09) did not bump it, because that
// change can never serve stale content - it only *stops* caching a scope it
// shouldn't have. Any already-cached `/p/...` entries in IMG_CACHE_NAME are
// simply never matched again and get evicted by trimImageCache's normal
// insertion-order cap - no active purge needed.
//
// History: the previous version cached every same-origin GET (including
// index.html and the JS bundle) cache-first with no revalidation, so once a
// browser had visited the site it kept being served whatever index.html/
// bundle existed at that visit, forever - admin CRUD writes worked (backend
// cache invalidation was fine) but the browser was still running pre-fix JS
// until the user manually cleared site data. See ObsidianVault's
// admin-crud-cache-stale-sw.md for the investigation.
//
// BUILD_ID is replaced at image build time (deployments/docker/
// Dockerfile.frontend) with EXPO_PUBLIC_BUILD_ID - the commit SHA in CD. That
// makes sw.js byte-different on every deploy, which is the only signal a
// browser has that a new service worker (and so a new app version) exists,
// and ties the shell cache to the build so it is evicted on every deploy
// instead of by hand-bumping a version string. Unstamped (expo start dev
// server) it stays the literal placeholder, which is still a valid cache name.
const BUILD_ID = '__BUILD_ID__'
const CACHE_NAME = `jops-shell-${BUILD_ID}`
const OFFLINE_URL = '/offline.html'
// '/' is the SPA shell; the icon is what offline.html shows.
const SHELL_URLS = ['/', '/manifest.json', OFFLINE_URL, '/icons/icon-192.png']

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

// No skipWaiting() here, deliberately. A new deploy's worker installs in the
// background and then WAITS: activating it swaps CACHE_NAME's cache (the
// activate handler below deletes the previous build's), which must not happen
// under a page still running the previous bundle - and must never interrupt
// a game in progress. The page decides when it is safe (features/pwa/lib/
// sw-update.ts) and asks for activation with a SKIP_WAITING message, then
// reloads on controllerchange. The very first install has no previous worker
// to wait behind, so it still activates immediately.
self.addEventListener('install', (event) => {
  event.waitUntil(caches.open(CACHE_NAME).then((cache) => cache.addAll(SHELL_URLS)))
})

self.addEventListener('message', (event) => {
  if (event.data && event.data.type === 'SKIP_WAITING') {
    self.skipWaiting()
  }
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
      .then(() => self.clients.claim())
  )
})

// Hashed build output (apps/frontend/dist/_expo/static/...) - the filename
// itself changes on any content change, so serving a cached copy forever is
// safe. Mirrors nginx.frontend.conf's `expires 1y; Cache-Control: public,
// immutable` rule for the same path.
function isImmutableStaticAsset(url) {
  return url.origin === self.location.origin && url.pathname.startsWith('/_expo/static/')
}

// The content-addressed media proxy (only same-origin where NPM/nginx routes
// `/api` and `/` off one host - see media-proxy-content-addressed.md's
// "MEDIA_BASE_URL / orígenes" section). The `/p/...` private/signed scope
// (user avatars) is deliberately excluded: its `exp` is quantized to
// MEDIA_PRIVATE_URL_TTL/2 windows (12h by default), so the same avatar's
// bytes get a brand-new URL twice a day - every rotation orphans the
// previous entry, and trimImageCache's insertion-order eviction then
// preferentially evicts the *public* (truly-forever-valid) entries that
// were inserted earlier, not the churning private ones. Avatars still get
// browser HTTP caching (the backend already sends
// `private, max-age, immutable` for them), just not Cache Storage. See
// ObsidianVault/sw-cache-media-scope-privado-2026-09-09.md.
function isImmutableImage(url) {
  return (
    url.origin === self.location.origin &&
    url.pathname.startsWith('/api/v1/media/') &&
    !url.pathname.startsWith('/api/v1/media/p/')
  )
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

// What a navigation gets when the network can't serve it: this is a SPA, so
// the cached shell boots the app on any path (the in-app offline banner
// explains the missing data). Only when there is no shell at all - a first
// visit that lost its connection mid-install - does the standalone offline
// page take over. undefined = nothing to offer.
async function offlineFallback() {
  const shell = await caches.match('/')
  if (shell) return shell
  return caches.match(OFFLINE_URL)
}

// Network-first: always tries to get the latest deploy's HTML, only falling
// back to whatever shell is cached when the network is unreachable (true
// offline, or the dev-only first-load race documented below) or the server
// side is down mid-deploy (a 5xx from the proxy while the frontend container
// restarts).
async function networkFirst(request) {
  try {
    const response = await fetch(request)
    if (response.ok) {
      const cache = await caches.open(CACHE_NAME)
      void cache.put(request, response.clone())
    } else if (request.mode === 'navigate' && response.status >= 500) {
      const fallback = await offlineFallback()
      if (fallback) return fallback
    }
    return response
  } catch (error) {
    const cached = await caches.match(request)
    if (cached) return cached
    if (request.mode === 'navigate') {
      const fallback = await offlineFallback()
      if (fallback) return fallback
    }
    // A rejected fetch with nothing cached yet (the browser cancelling this
    // SW-side request in favor of the real navigation's own fetch - a known
    // race right after the SW installs) must not surface as an unhandled
    // rejection.
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

// --- Web Push ---------------------------------------------------------------
//
// The backend (PushNotificationService) sends {title, body, url, tag} for the
// moments of a game worth waking a phone for. Chrome requires every push to
// end in a visible notification unless a window of this site is visible - so
// a visible window is the one case where the notification is skipped (the
// player is already looking at the live game), everything else shows one.

// Only same-origin paths may be opened from a notification: `url` comes from
// a push payload, and an absolute or protocol-relative value must never turn a
// tap into a navigation off-site.
function safeAppPath(raw) {
  return typeof raw === 'string' && raw.startsWith('/') && !raw.startsWith('//') ? raw : '/'
}

async function showPushNotification(data) {
  const windows = await self.clients.matchAll({ type: 'window', includeUncontrolled: true })
  if (windows.some((client) => client.visibilityState === 'visible')) return

  await self.registration.showNotification(data.title || 'JOPS', {
    body: data.body || '',
    icon: '/icons/icon-192.png',
    badge: '/icons/badge-96.png',
    // One notification per game: a newer moment replaces the previous one
    // (and re-alerts, since the content changed).
    tag: data.tag || undefined,
    renotify: !!data.tag,
    data: { url: safeAppPath(data.url) },
  })
}

self.addEventListener('push', (event) => {
  let data = {}
  try {
    data = event.data ? event.data.json() : {}
  } catch (error) {
    // A malformed payload still has to produce a notification (see above).
    data = {}
  }
  event.waitUntil(showPushNotification(data))
})

async function focusOrOpenApp(rawUrl) {
  const path = safeAppPath(rawUrl)
  const windows = await self.clients.matchAll({ type: 'window', includeUncontrolled: true })
  const existing = windows.find((client) => new URL(client.url).origin === self.location.origin)
  if (existing) {
    // Hand the path to the running app so it routes without a full reload
    // (a reload would drop its in-memory session and live game socket).
    await existing.focus()
    existing.postMessage({ type: 'NOTIFICATION_CLICK', url: path })
    return
  }
  await self.clients.openWindow(path)
}

self.addEventListener('notificationclick', (event) => {
  event.notification.close()
  const url = event.notification.data && event.notification.data.url
  event.waitUntil(focusOrOpenApp(url))
})
