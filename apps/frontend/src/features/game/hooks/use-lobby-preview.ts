import { useQuery } from '@tanstack/react-query'

import { previewGameByCode } from '@/features/game/api/game.api'
import { gameKeys } from '@/features/game/api/game.keys'
import { isCompleteCode } from '@/features/game/lib/game-code'

export function useLobbyPreview(code: string) {
  return useQuery({
    queryKey: gameKeys.preview(code),
    queryFn: () => previewGameByCode(code),
    enabled: isCompleteCode(code),
    retry: false,
    staleTime: 0,
    gcTime: 0,
  })
}
