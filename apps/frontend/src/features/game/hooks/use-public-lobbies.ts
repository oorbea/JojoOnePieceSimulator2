import { useQuery } from '@tanstack/react-query'

import { getPublicLobbies } from '@/features/game/api/game.api'
import { gameKeys } from '@/features/game/api/game.keys'

export function usePublicLobbies() {
  return useQuery({
    queryKey: gameKeys.publicList(),
    queryFn: () => getPublicLobbies(),
    staleTime: 5_000,
    refetchInterval: 15_000,
    refetchOnWindowFocus: true,
  })
}
