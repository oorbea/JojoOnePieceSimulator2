// Storage for a dev-login (POST /auth/dev-login) refresh token - see
// features/auth's dev-login-container and refresh.ts's doRefresh. Backed by
// sessionStorage rather than the shared, browser-wide refresh cookie the
// Google flow uses (session.store.ts/secure-storage.ts): sessionStorage is
// scoped to one tab, so several dev accounts can be logged into at once in
// separate tabs of the same browser, which a shared cookie can't do (see
// ObsidianVault/dev-auth-bypass.md).
//
// Web-only, exactly like the dev-login flow itself; a no-op everywhere else
// (native has no sessionStorage, and dev-login never renders there - see
// env.EXPO_PUBLIC_DEV_AUTH). Wrapped in try/catch since sessionStorage can
// throw in a locked-down context (an iframe with storage disabled, a very
// old/strict browser setting) - a dev convenience must never crash the app.
const DEV_REFRESH_TOKEN_KEY = 'jops.dev_rt'

function hasSessionStorage(): boolean {
  return typeof window !== 'undefined' && !!window.sessionStorage
}

export function getDevRefreshToken(): string | null {
  if (!hasSessionStorage()) return null
  try {
    return window.sessionStorage.getItem(DEV_REFRESH_TOKEN_KEY)
  } catch {
    return null
  }
}

export function setDevRefreshToken(token: string): void {
  if (!hasSessionStorage()) return
  try {
    window.sessionStorage.setItem(DEV_REFRESH_TOKEN_KEY, token)
  } catch {
    // Best-effort - see doc comment above.
  }
}

export function clearDevRefreshToken(): void {
  if (!hasSessionStorage()) return
  try {
    window.sessionStorage.removeItem(DEV_REFRESH_TOKEN_KEY)
  } catch {
    // Best-effort - see doc comment above.
  }
}
