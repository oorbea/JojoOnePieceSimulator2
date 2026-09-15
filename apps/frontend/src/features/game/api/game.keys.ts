import { queryKeys } from '@/shared/api/query-keys'

// Query-cache rules for the game feature (see query-provider.tsx's
// dehydrateOptions, which keys off this same 'games' segment):
//
// 1. The socket store is the only *writer* of `detail(id)` while the socket
//    is open (see hooks/use-game-socket.ts) - one writer means the cache can
//    never drift behind the socket.
// 2. REST GET /games/{id} is a seed-and-fallback only (cold start, resume,
//    socket unavailable), never polled while the socket is open.
// 3. `detail` is never persisted and is removed on leave/kick/abort.
export const gameKeys = {
  allLocales: [...queryKeys.root, 'games'] as const,
  detail: (id: string) => [...gameKeys.allLocales, 'detail', id] as const,
  preview: (code: string) => [...gameKeys.allLocales, 'preview', code] as const,
  publicList: () => [...gameKeys.allLocales, 'public'] as const,
  // `mine` (GET /games/me, see game.api.ts) follows the same rule as
  // `detail`: never persisted (the 'games' segment as a whole is excluded
  // from dehydration), a plain seed/fallback query that home-container.tsx
  // reads once on mount to offer "resume your game" - not polled, not
  // written to by the socket store.
  mine: () => [...gameKeys.allLocales, 'mine'] as const,
}
