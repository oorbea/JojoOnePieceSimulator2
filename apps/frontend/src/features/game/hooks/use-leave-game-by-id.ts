import { useMutation, useQueryClient } from '@tanstack/react-query'

import { leaveGameById } from '@/features/game/api/game.api'
import { gameKeys } from '@/features/game/api/game.keys'

// Backs the invite-link "you're already in another game - leave it and
// join this one?" confirm (join-invite-container.tsx). There is no open WS
// socket to the game being left at that point (see leaveGameById's own
// doc), so this goes over REST instead of the WS LEAVE command
// use-game-commands.ts sends from inside an open lobby.
export function useLeaveGameById() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (gameId: string) => leaveGameById(gameId),
    onSuccess: (_data, gameId) => {
      queryClient.removeQueries({ queryKey: gameKeys.detail(gameId) })
      queryClient.removeQueries({ queryKey: gameKeys.mine() })
    },
  })
}
