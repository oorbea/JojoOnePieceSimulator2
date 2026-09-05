import { create } from 'zustand'
import { Platform } from 'react-native'

import { secureStorage } from '@/shared/lib/secure-storage'
import { refreshSession } from '@/shared/api/refresh'
import { postLogout } from '@/features/auth/api/auth.api'
import { REFRESH_TOKEN_KEY } from '@/shared/api/refresh-token-key'
import { fromUserResponse, type SessionUser } from '@/shared/stores/session-user'
import { clearPersistedQueryCache } from '@/shared/stores/query-cache-purge'

const LEGACY_SESSION_STORAGE_KEY = 'jops.session'

// Re-exported so existing importers of `SessionUser`/`fromUserResponse` from
// this module keep working unchanged - the type/helper itself now lives in
// session-user.ts to break a circular import with shared/api/refresh.ts.
export type { SessionUser }
export { fromUserResponse }

type Session = {
  accessToken: string
  user: SessionUser
}

type SessionState = {
  session: Session | null
  isHydrated: boolean
  hydrate: () => Promise<void>
  setSession: (session: Session) => void
  clearSession: () => Promise<void>
}

// The backend now issues a short-lived (15min) access JWT plus a rotating
// opaque refresh token: an HttpOnly/Secure cookie on web, SecureStore on
// native (see shared/api/refresh.ts). This store only ever holds the
// in-memory access token + user - it never itself persists anything; on
// boot it asks refresh.ts to trade whatever refresh token exists for a fresh
// access token, and a 401 on any request triggers the same silent refresh
// (shared/api/interceptors.ts) before falling back to a full sign-out.
export const useSessionStore = create<SessionState>((set) => ({
  session: null,
  isHydrated: false,

  hydrate: async () => {
    // One-shot legacy-key purge: this key held the whole {accessToken,
    // expiresAt, user} blob before the refresh-token rework. Harmless once
    // gone - cheap insurance against a stale value ever being read again.
    if (typeof window !== 'undefined' && window.localStorage) {
      window.localStorage.removeItem(LEGACY_SESSION_STORAGE_KEY)
    }

    try {
      const result = await refreshSession()
      if (result) {
        set({ session: { accessToken: result.accessToken, user: result.user }, isHydrated: true })
      } else {
        set({ session: null, isHydrated: true })
      }
    } catch {
      // A stuck spinner is worse than a false "logged out" - always land on
      // isHydrated: true even if refreshSession throws unexpectedly.
      set({ session: null, isHydrated: true })
    }
  },

  setSession: (session) => {
    set({ session })
  },

  clearSession: async () => {
    try {
      await postLogout()
    } catch {
      // Logout is best-effort server-side (it always responds 204 anyway) -
      // local cleanup below must happen regardless of a network failure.
    } finally {
      if (Platform.OS !== 'web') {
        await secureStorage.removeItem(REFRESH_TOKEN_KEY).catch(() => undefined)
      }
      // Another user's profile/lobby data must not linger in the TanStack
      // Query cache (in-memory or its AsyncStorage/localStorage-persisted
      // snapshot, see providers/query-client.ts) past logout - moving the
      // auth token out of storage while leaving that behind would be an
      // incomplete sign-out.
      await clearPersistedQueryCache().catch(() => undefined)
      set({ session: null })
    }
  },
}))

// Non-hook accessor for use outside React (axios interceptors).
export const getSessionToken = (): string | null =>
  useSessionStore.getState().session?.accessToken ?? null
