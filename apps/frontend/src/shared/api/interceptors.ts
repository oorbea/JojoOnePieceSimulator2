import type { AxiosInstance } from 'axios'

import { getKnownEtag, rememberEtag } from '@/shared/api/etag'
import { toAppError } from '@/shared/api/errors'
import { getSessionToken, useSessionStore } from '@/shared/stores/session.store'

export function registerInterceptors(client: AxiosInstance): void {
  client.interceptors.request.use((config) => {
    const token = getSessionToken()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }

    // Conditional GET: attach the last-seen ETag for this exact URL so an
    // unchanged resource comes back as a cheap 304 instead of a full body.
    if (config.method === 'get' && config.url) {
      const etag = getKnownEtag(config.url)
      if (etag) config.headers['If-None-Match'] = etag
    }

    return config
  })

  client.interceptors.response.use(
    (response) => {
      if (response.config.method === 'get' && response.config.url) {
        rememberEtag(response.config.url, response.headers.etag)
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
