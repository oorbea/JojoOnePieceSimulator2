import { useMutation, useQueryClient } from '@tanstack/react-query'

import { joinGameByInvite } from '@/features/game/api/game.api'
import { gameKeys } from '@/features/game/api/game.keys'

// Mirrors useJoinGameByCode (use-join-game.ts) - same seed-the-cache
// onSuccess so the lobby route this redirects to doesn't have to refetch.
export function useJoinGameByInvite() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (token: string) => joinGameByInvite(token),
    onSuccess: (data) => {
      queryClient.setQueryData(gameKeys.detail(data.game.id), data)
    },
  })
}
