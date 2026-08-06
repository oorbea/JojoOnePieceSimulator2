import { create } from 'zustand'

import { secureStorage } from '@/shared/lib/secure-storage'
import type { Locale, Role } from '@/shared/lib/zod'

const SESSION_STORAGE_KEY = 'jops.session'

export type SessionUser = {
  id: string
  email: string
  username: string
  completeName: string
  picture: string | null
  role: Role
  language: Locale
}

type Session = {
  accessToken: string
  expiresAt: string
  user: SessionUser
}

type SessionState = {
  session: Session | null
  isHydrated: boolean
  hydrate: () => Promise<void>
  setSession: (session: Session) => Promise<void>
  clearSession: () => Promise<void>
}

// The backend issues a plain bearer JWT with no refresh token (see
// apps/backend .../auth_endpoints.go), so "session" here is just the last
// issued token + its owner; expiry is handled by the 401 interceptor forcing
// a fresh Google sign-in, not by silent refresh.
export const useSessionStore = create<SessionState>((set) => ({
  session: null,
  isHydrated: false,

  hydrate: async () => {
    const raw = await secureStorage.getItem(SESSION_STORAGE_KEY)
    set({ session: raw ? (JSON.parse(raw) as Session) : null, isHydrated: true })
  },

  setSession: async (session) => {
    await secureStorage.setItem(SESSION_STORAGE_KEY, JSON.stringify(session))
    set({ session })
  },

  clearSession: async () => {
    await secureStorage.removeItem(SESSION_STORAGE_KEY)
    set({ session: null })
  },
}))

// Non-hook accessor for use outside React (axios interceptors).
export const getSessionToken = (): string | null =>
  useSessionStore.getState().session?.accessToken ?? null
