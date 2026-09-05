import { Platform } from 'react-native'

import { apiClient } from '@/shared/api/client'
import { assertContract } from '@/shared/api/assert-contract'
import { secureStorage } from '@/shared/lib/secure-storage'
import { REFRESH_TOKEN_KEY } from '@/shared/api/refresh-token-key'
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
    }
    await apiClient.post('/auth/logout', undefined, { withCredentials: true, headers })
  } catch {
    // Swallowed by design - see doc comment above.
  }
}
