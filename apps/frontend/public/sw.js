// Minimal cache-first service worker for the app shell. Bump CACHE_NAME on
// every deploy that changes shell assets so stale caches get evicted instead
// of serving an old bundle forever.
const CACHE_NAME = 'jops-shell-v1'
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

self.addEventListener('fetch', (event) => {
  // Only cache same-origin GETs — never intercept API calls to the backend,
  // those must always hit the network so auth/ETag handling stays correct.
  const url = new URL(event.request.url)
  if (event.request.method !== 'GET' || url.origin !== self.location.origin) {
    return
  }

  event.respondWith(
    caches.match(event.request).then((cached) => {
      if (cached) return cached
      return fetch(event.request).then((response) => {
        if (response.ok) {
          const clone = response.clone()
          caches.open(CACHE_NAME).then((cache) => cache.put(event.request, clone))
        }
        return response
      })
    })
  )
})
