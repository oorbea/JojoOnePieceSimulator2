import { useQuery } from '@tanstack/react-query'

import { getInviteStatus } from '@/features/game/api/game.api'
import { gameKeys } from '@/features/game/api/game.keys'

// Public - no session required, safe to run before/without login (the
// whole point of the /join/[token] route). Never retries: the backend
// answers VALID/EXPIRED for literally any input, so a network error here is
// a real error, not a miss worth retrying into.
export function useInviteStatus(token: string) {
  return useQuery({
    queryKey: gameKeys.inviteStatus(token),
    queryFn: () => getInviteStatus(token),
    enabled: !!token,
    retry: false,
    staleTime: 0,
    gcTime: 0,
  })
}
