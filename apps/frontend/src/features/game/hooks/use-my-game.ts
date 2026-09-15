import { useQuery } from '@tanstack/react-query'

import { getMyGame } from '@/features/game/api/game.api'
import { gameKeys } from '@/features/game/api/game.keys'

// Fetched once per app open (see home-container.tsx) so a player who closed
// the tab/reloaded lands back on their game instead of nowhere - the other
// half of what the disconnect grace period fixes on the backend (see
// GameService's grace timers). `enabled` lets a caller gate this behind
// "user is signed in and looking at a screen where resuming makes sense"
// (the home screen), not fire from every mount app-wide.
export function useMyGame(enabled: boolean) {
  return useQuery({
    queryKey: gameKeys.mine(),
    queryFn: getMyGame,
    enabled,
    staleTime: 0,
  })
}
