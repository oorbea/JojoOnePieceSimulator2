import { useLanguageStore } from '@/shared/stores/language.store'

// The backend sends ETag + Cache-Control: private on every read and honors
// If-None-Match with a 304 (apps/backend .../cache_headers.go). A 304 has no
// body, so the conditional-GET flow only works if the *body* that earned
// that ETag is cached alongside it here - remembering the ETag alone would
// leave axios handing back an empty response.data on every 304 (this broke
// the Stands/Devil Fruits lists: a fresh POST invalidated React Query, the
// refetch got a 304, and the "no data" empty state rendered because the
// container ended up with '' instead of the array).
//
// Keyed by URL + params + the active locale, mirroring the backend's own
// Vary: Authorization, Accept-Language - a filtered request or a different
// locale must never reuse another one's cached body.
type CachedResponse = { etag: string; data: unknown }

const cacheByKey = new Map<string, CachedResponse>()

// Stable across param-object identity/ordering, same reasoning as the
// backend's canonical filter cache key (stand_repository.go's
// standFilterKey) - `{a:1,b:2}` and `{b:2,a:1}` must hit the same entry.
export function etagCacheKey(url: string, params?: Record<string, unknown>): string {
  const locale = useLanguageStore.getState().locale
  const sortedParams = params
    ? Object.keys(params)
        .sort()
        .map((k) => `${k}=${String(params[k])}`)
        .join('&')
    : ''
  return `${locale}|${url}|${sortedParams}`
}

export function getCachedResponse(key: string): CachedResponse | undefined {
  return cacheByKey.get(key)
}

export function rememberResponse(key: string, etag: string | undefined, data: unknown): void {
  if (etag) cacheByKey.set(key, { etag, data })
}

export function clearEtags(): void {
  cacheByKey.clear()
}
