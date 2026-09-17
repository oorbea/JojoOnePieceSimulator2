import { useQuery } from '@tanstack/react-query'

import { getInvitePreview } from '@/features/game/api/game.api'
import { gameKeys } from '@/features/game/api/game.keys'

// Authenticated - `enabled` gates this behind "signed in and the public
// status check already came back VALID", so it never fires for a visitor
// who hasn't logged in yet (that request would just 401).
export function useInvitePreview(token: string, enabled: boolean) {
  return useQuery({
    queryKey: gameKeys.invitePreview(token),
    queryFn: () => getInvitePreview(token),
    enabled: enabled && !!token,
    retry: false,
    staleTime: 0,
    gcTime: 0,
  })
}
