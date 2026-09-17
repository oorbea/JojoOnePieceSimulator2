import { useMutation } from '@tanstack/react-query'

import { createGameInvite } from '@/features/game/api/game.api'
import { useGameInviteStore } from '@/features/game/stores/game-invite.store'

// useGameInvite returns a function that resolves to a live invite token for
// (gameId, code): the cached one if it still has more than 2 minutes left
// (see game-invite.store.ts), otherwise a freshly minted one. Callers pass
// the lobby's *current* join code so a rotation invalidates the cache for
// free - see game-invite.store.ts's doc.
export function useGameInvite() {
  const getCached = useGameInviteStore((state) => state.get)
  const setCached = useGameInviteStore((state) => state.set)
  const mint = useMutation({ mutationFn: (gameId: string) => createGameInvite(gameId) })

  const getOrMintToken = async (gameId: string, code: string): Promise<string> => {
    const cached = getCached(gameId, code)
    if (cached) return cached
    const invite = await mint.mutateAsync(gameId)
    setCached(gameId, code, invite.token, invite.expiresAt)
    return invite.token
  }

  return { getOrMintToken, isMinting: mint.isPending }
}
