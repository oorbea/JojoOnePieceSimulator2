import { create } from 'zustand'

// REUSE_THRESHOLD_MS: share re-mints a fresh token only once the cached one
// has under 2 minutes left - the owner's decision. Mints one per share
// press would otherwise be a) wasteful (a token per click, all but the last
// expiring unused) and b) a real UX bug (handing a friend a link that's
// seconds from dying). Old tokens already shared stay valid until their own
// TTL regardless of what this cache does next - see game_invite.go's doc.
const REUSE_THRESHOLD_MS = 2 * 60 * 1000

type CachedInvite = {
  gameId: string
  // The join code the token was minted against - included in the cache key
  // so a code rotation (which the backend makes ALL outstanding invites for
  // this game die against immediately, see IGameInviteStore's doc) drops
  // the cached token for free the next time the lobby state updates,
  // without this store needing to know rotation happened at all.
  code: string
  token: string
  expiresAt: number
}

type GameInviteState = {
  cached: CachedInvite | null
  // Returns the cached token for (gameId, code) if it still has more than
  // REUSE_THRESHOLD_MS left, else null (caller should mint a new one).
  get: (gameId: string, code: string) => string | null
  set: (gameId: string, code: string, token: string, expiresAt: string) => void
}

export const useGameInviteStore = create<GameInviteState>((set, get) => ({
  cached: null,
  get: (gameId, code) => {
    const c = get().cached
    if (!c || c.gameId !== gameId || c.code !== code) return null
    if (c.expiresAt - Date.now() <= REUSE_THRESHOLD_MS) return null
    return c.token
  },
  set: (gameId, code, token, expiresAt) => {
    set({ cached: { gameId, code, token, expiresAt: new Date(expiresAt).getTime() } })
  },
}))
