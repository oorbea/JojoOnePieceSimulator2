import { Platform } from 'react-native'

// Stashes the invite token a not-yet-authenticated visitor tried to open,
// so it survives the round trip through Google's full-page web redirect
// (use-google-auth.ts's web flow navigates away from /join/[token]
// entirely, and on return it wipes window.location's query/hash - see that
// file's web-redirect effect) and login.tsx can send them back into the
// lobby once signed in.
//
// sessionStorage, not localStorage: it survives a full-page navigation in
// the SAME tab (the whole requirement) and dies with the tab, so a second
// tab/device can never "resume" someone else's invite by sharing storage.
// Native has no full-page-redirect round trip to survive (expo-auth-session
// stays in-app), so a plain module-level variable is enough there.
const KEY = 'jojo.pendingInvite'
const MAX_AGE_MS = 15 * 60 * 1000 // matches the backend's own GameInviteTTL

type StashedInvite = { token: string; at: number }

let nativePending: StashedInvite | null = null

function hasSessionStorage(): boolean {
  return Platform.OS === 'web' && typeof window !== 'undefined' && !!window.sessionStorage
}

export function setPendingInvite(token: string): void {
  const entry: StashedInvite = { token, at: Date.now() }
  if (hasSessionStorage()) {
    try {
      window.sessionStorage.setItem(KEY, JSON.stringify(entry))
    } catch {
      // Private browsing / storage disabled - the invite just won't survive
      // the redirect; nothing else this function can do about it.
    }
    return
  }
  nativePending = entry
}

// readPendingInvite returns the stashed token, or null if there is none or
// it's gone stale (older than the invite's own TTL - no point resuming
// into a link that's already dead).
export function readPendingInvite(): string | null {
  let raw: string | null = null
  if (hasSessionStorage()) {
    try {
      raw = window.sessionStorage.getItem(KEY)
    } catch {
      return null
    }
  } else if (nativePending) {
    raw = JSON.stringify(nativePending)
  }
  if (!raw) return null

  try {
    const entry = JSON.parse(raw) as StashedInvite
    if (Date.now() - entry.at > MAX_AGE_MS) return null
    return entry.token
  } catch {
    return null
  }
}

export function clearPendingInvite(): void {
  if (hasSessionStorage()) {
    try {
      window.sessionStorage.removeItem(KEY)
    } catch {
      // Nothing to clear if storage itself is unavailable.
    }
    return
  }
  nativePending = null
}
