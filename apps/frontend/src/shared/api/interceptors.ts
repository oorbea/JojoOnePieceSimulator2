import type { AxiosInstance, InternalAxiosRequestConfig } from 'axios'

import { etagCacheKey, getCachedResponse, rememberResponse } from '@/shared/api/etag'
import { toAppError } from '@/shared/api/errors'
import { useLanguageStore } from '@/shared/stores/language.store'
import { getSessionToken, useSessionStore } from '@/shared/stores/session.store'

// Stashed on the request config so the response interceptor doesn't have to
// recompute (and risk drifting from) the same cache key - the locale could
// theoretically change between request and response otherwise.
type ConfigWithEtagKey = InternalAxiosRequestConfig & { __etagCacheKey?: string }

export function registerInterceptors(client: AxiosInstance): void {
  client.interceptors.request.use((config: ConfigWithEtagKey) => {
    const token = getSessionToken()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }

    // Tells the backend which locale to resolve Stand/DevilFruit
    // description/skills into (see the resolveLocale middleware, apps/backend
    // .../api/endpoints/locale.go) - every request carries it, not just reads,
    // so admin writes and the cache layer see a consistent locale too.
    config.headers['Accept-Language'] = useLanguageStore.getState().locale

    // Conditional GET: attach the last-seen ETag for this exact URL/params/
    // locale so an unchanged resource comes back as a cheap 304 - but only
    // when we also kept the body it belongs to, since a 304 has none of its
    // own and axios would otherwise hand the caller an empty string.
    if (config.method === 'get' && config.url) {
      const key = etagCacheKey(config.url, config.params)
      config.__etagCacheKey = key
      const cached = getCachedResponse(key)
      if (cached) config.headers['If-None-Match'] = cached.etag
    }

    return config
  })

  client.interceptors.response.use(
    (response) => {
      const config = response.config as ConfigWithEtagKey
      if (config.method === 'get' && config.__etagCacheKey) {
        if (response.status === 304) {
          // No body on a 304 - serve the body we cached alongside the ETag
          // that just got revalidated, and normalize status so callers never
          // have to special-case 304 themselves.
          const cached = getCachedResponse(config.__etagCacheKey)
          if (cached) {
            response.status = 200
            response.data = cached.data
          }
        } else {
          rememberResponse(config.__etagCacheKey, response.headers.etag, response.data)
        }
      }
      return response
    },
    (error) => {
      if (error.response?.status === 401) {
        // Token is expired/invalid and there is no refresh token — the only
        // recovery path is a fresh Google sign-in, so drop the session.
        void useSessionStore.getState().clearSession()
      }
      return Promise.reject(toAppError(error))
    }
  )
}
