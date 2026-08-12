import { useMutation, useQueryClient } from '@tanstack/react-query'

import { createGame } from '@/features/game/api/game.api'
import { gameKeys } from '@/features/game/api/game.keys'
import type { CreateGameInput } from '@/features/game/types/game.types'

export function useCreateGame() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateGameInput) => createGame(input),
    onSuccess: (data) => {
      queryClient.setQueryData(gameKeys.detail(data.game.id), data)
    },
  })
}
