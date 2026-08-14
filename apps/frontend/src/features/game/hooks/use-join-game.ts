import { useMutation, useQueryClient } from '@tanstack/react-query'

import { joinGameByCode, joinGameById } from '@/features/game/api/game.api'
import { gameKeys } from '@/features/game/api/game.keys'

export function useJoinGameByCode() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (code: string) => joinGameByCode(code),
    onSuccess: (data) => {
      queryClient.setQueryData(gameKeys.detail(data.game.id), data)
    },
  })
}

export function useJoinGameById() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (gameId: string) => joinGameById(gameId),
    onSuccess: (data) => {
      queryClient.setQueryData(gameKeys.detail(data.game.id), data)
    },
  })
}
