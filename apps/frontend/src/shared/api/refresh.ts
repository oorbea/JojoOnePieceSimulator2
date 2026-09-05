import axios from 'axios'
import { Platform } from 'react-native'

import { env } from '@/shared/config/env'
import { secureStorage } from '@/shared/lib/secure-storage'
import { REFRESH_TOKEN_KEY } from '@/shared/api/refresh-token-key'
import { fromUserResponse, type SessionUser } from '@/shared/stores/session-user'
import type { LoginResponse } from '@/shared/contracts/dto'

// Bare axios instance, deliberately without registerInterceptors() - going
// through the interceptor-bearing apiClient would route a 401 from
// /auth/refresh itself back into the response interceptor's own refresh
// handling and recurse.
// eslint-disable-next-line import/no-named-as-default-member -- default import is correct; axios's named `create` export is unrelated
const refreshClient = axios.create({
  baseURL: env.EXPO_PUBLIC_API_URL,
  timeout: 15_000,
  withCredentials: true,
})

export type RefreshResult = { accessToken: string; user: SessionUser }

let inFlight: Promise<RefreshResult | null> | null = null

// Single-flight: N callers while a refresh is already in progress all await
// the same underlying request instead of firing N POST /auth/refresh calls
// (which would also mean N refresh-token rotations racing each other).
export function refreshSession(): Promise<RefreshResult | null> {
  inFlight ??= doRefresh().finally(() => {
    inFlight = null
  })
  return inFlight
}

async function doRefresh(): Promise<RefreshResult | null> {
  try {
    const headers: Record<string, string> = { 'X-JOPS-Refresh': '1' }

    if (Platform.OS !== 'web') {
      const stored = await secureStorage.getItem(REFRESH_TOKEN_KEY)
      if (!stored) return null
      headers['X-Refresh-Token'] = stored
      headers['X-Refresh-Token-Transport'] = 'header'
    }

    const { data } = await refreshClient.post<LoginResponse>('/auth/refresh', undefined, {
      headers,
    })

    if (Platform.OS !== 'web' && data.refreshToken) {
      await secureStorage.setItem(REFRESH_TOKEN_KEY, data.refreshToken)
    }

    return { accessToken: data.accessToken, user: fromUserResponse(data.user) }
  } catch {
    return null
  }
}
