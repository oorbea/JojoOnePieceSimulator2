import { useQuery, type Query } from '@tanstack/react-query'

import { getStands } from '@/features/stands/api/stands.api'
import { standKeys } from '@/features/stands/api/stands.keys'
import type { StandFilters, StandResponse } from '@/features/stands/types/stands.types'

// Same reasoning as use-profile.ts's pollInterval - the picture transcode is
// a fire-and-forget background job with no websocket to notify completion,
// so a PENDING Stand is only ever observed by polling this list query.
const MAX_POLL_ATTEMPTS = 8
const MAX_POLL_INTERVAL_MS = 30_000
const BASE_POLL_INTERVAL_MS = 2_000

function pollInterval(query: Query<StandResponse[]>): number | false {
  const hasPending = query.state.data?.some((s) => s.pictureStatus === 'PENDING')
  if (!hasPending) return false
  if (query.state.dataUpdateCount >= MAX_POLL_ATTEMPTS) return false
  return Math.min(BASE_POLL_INTERVAL_MS * 2 ** query.state.dataUpdateCount, MAX_POLL_INTERVAL_MS)
}

export function useStands(filters?: StandFilters) {
  return useQuery({
    queryKey: standKeys.list(filters),
    queryFn: () => getStands(filters),
    refetchInterval: pollInterval,
  })
}
