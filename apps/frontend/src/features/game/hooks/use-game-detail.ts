import { useQuery } from '@tanstack/react-query'

import { getGame } from '@/features/game/api/game.api'
import { gameKeys } from '@/features/game/api/game.keys'
import type { SocketStatus } from '@/features/game/stores/game-socket.store'

// REST fallback for when the socket isn't open yet or is unavailable
// (see game.keys.ts's cache rules) - never polls while the socket is open,
// since the socket is the sole writer of this same query key then.
export function useGameDetail(gameId: string | null, socketStatus: SocketStatus) {
  return useQuery({
    queryKey: gameKeys.detail(gameId ?? ''),
    queryFn: () => getGame(gameId as string),
    enabled: !!gameId && socketStatus !== 'open',
    staleTime: 0,
    refetchInterval: socketStatus === 'unavailable' ? 5_000 : false,
  })
}
