import { Platform } from 'react-native'

import { apiClient } from '@/shared/api/client'
import { assertContract } from '@/shared/api/assert-contract'
import { secureStorage } from '@/shared/lib/secure-storage'
import { REFRESH_TOKEN_KEY } from '@/shared/api/refresh-token-key'
import { getDevRefreshToken } from '@/shared/api/dev-refresh-token'
import { loginResponseSchema } from '@/shared/contracts/dto'
import type { AuthGoogleResponse } from '@/features/auth/types/auth.types'

export async function postGoogleAuth(idToken: string): Promise<AuthGoogleResponse> {
  const headers: Record<string, string> = {}
  // Native ignores cookies, so it opts into getting the refresh token back
  // in the response body instead (still also set as a cookie, which native
  // just never reads). Web gets it ONLY via the Set-Cookie response header,
  // which requires withCredentials so the browser actually stores it.
  if (Platform.OS !== 'web') headers['X-Refresh-Token-Transport'] = 'header'

  const response = await apiClient.post<AuthGoogleResponse>(
    '/auth/google',
    { idToken },
    { withCredentials: true, headers }
  )
  if (__DEV__) assertContract(loginResponseSchema, response.data, 'POST /auth/google')
  return response.data
}

// Local-only dev login (POST /auth/dev-login) - see dev-login-container.tsx.
// Always requests the header transport: dev-login never sets the shared
// refresh cookie server-side either (see AuthEndpoints.devLogin's doc), so
// the caller must keep the returned refreshToken itself (dev-refresh-token.ts).
export async function postDevLogin(name: string, admin: boolean): Promise<AuthGoogleResponse> {
  const response = await apiClient.post<AuthGoogleResponse>(
    '/auth/dev-login',
    { name, admin },
    { headers: { 'X-Refresh-Token-Transport': 'header' } }
  )
  if (__DEV__) assertContract(loginResponseSchema, response.data, 'POST /auth/dev-login')
  return response.data
}

// Always resolves, never throws - logout is best-effort from the UI's point
// of view (the backend itself always answers 204 regardless of whether the
// refresh token it was given was valid), and session.store.ts's clearSession
// must not be blocked by a network failure here.
export async function postLogout(): Promise<void> {
  try {
    const headers: Record<string, string> = { 'X-JOPS-Refresh': '1' }
    if (Platform.OS !== 'web') {
      const stored = await secureStorage.getItem(REFRESH_TOKEN_KEY)
      if (stored) headers['X-Refresh-Token'] = stored
    } else {
      // A dev-login session has no cookie to fall back on (see
      // postDevLogin's doc) - its refresh token must be sent explicitly or
      // /auth/logout has nothing to revoke.
      const devToken = getDevRefreshToken()
      if (devToken) headers['X-Refresh-Token'] = devToken
    }
    await apiClient.post('/auth/logout', undefined, { withCredentials: true, headers })
  } catch {
    // Swallowed by design - see doc comment above.
  }
}
