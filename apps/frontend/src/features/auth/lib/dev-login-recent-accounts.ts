// Remembers the dev-login accounts this browser has used, so a returning
// developer can tap back in instead of retyping a name - see
// dev-login-container.tsx. Deliberately localStorage (persists across tabs
// and restarts), unlike the per-tab session itself (dev-refresh-token.ts,
// sessionStorage) - this is just a convenience list, not a credential.
const RECENT_ACCOUNTS_KEY = 'jops.dev_accounts'
const MAX_RECENT_ACCOUNTS = 8

export type RecentDevAccount = { name: string; admin: boolean }

function hasLocalStorage(): boolean {
  return typeof window !== 'undefined' && !!window.localStorage
}

export function getRecentDevAccounts(): RecentDevAccount[] {
  if (!hasLocalStorage()) return []
  try {
    const raw = window.localStorage.getItem(RECENT_ACCOUNTS_KEY)
    if (!raw) return []
    const parsed: unknown = JSON.parse(raw)
    if (!Array.isArray(parsed)) return []
    return parsed.filter(
      (entry): entry is RecentDevAccount =>
        !!entry && typeof entry.name === 'string' && typeof entry.admin === 'boolean'
    )
  } catch {
    return []
  }
}

// Moves (or adds) account to the front of the list, capped at
// MAX_RECENT_ACCOUNTS - the most recently used accounts stay reachable
// without the list growing forever across a long-lived dev machine.
export function rememberDevAccount(account: RecentDevAccount): void {
  if (!hasLocalStorage()) return
  try {
    const rest = getRecentDevAccounts().filter((a) => a.name !== account.name)
    const next = [account, ...rest].slice(0, MAX_RECENT_ACCOUNTS)
    window.localStorage.setItem(RECENT_ACCOUNTS_KEY, JSON.stringify(next))
  } catch {
    // Best-effort - see dev-refresh-token.ts's same reasoning.
  }
}
