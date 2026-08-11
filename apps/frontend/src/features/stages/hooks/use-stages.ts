import { useQuery, type Query } from '@tanstack/react-query'
import { useRef } from 'react'
import { Platform } from 'react-native'

import { getStages } from '@/features/stages/api/stages.api'
import { stageKeys } from '@/features/stages/api/stages.keys'
import type { StageFilters, StageResponse } from '@/features/stages/types/stages.types'

// On web, PictureEventsBridge (src/providers/picture-events-bridge.tsx)
// pushes picture-pipeline completion via SSE instead - polling here would
// be redundant. React Native has no EventSource, so native keeps this
// polling as its fallback; same reasoning as use-stands.ts's pollInterval.
const MAX_POLL_ATTEMPTS = 8
const MAX_POLL_INTERVAL_MS = 30_000
const BASE_POLL_INTERVAL_MS = 2_000

export function useStages(filters?: StageFilters) {
  // See use-stands.ts's comment on why this must be a ref scoped to this
  // hook instance rather than derived from query.state.dataUpdateCount.
  const pollAttempts = useRef(0)

  return useQuery({
    queryKey: stageKeys.list(filters),
    queryFn: () => getStages(filters),
    refetchInterval:
      Platform.OS === 'web'
        ? undefined
        : (query: Query<StageResponse[]>) => {
            const hasPending = query.state.data?.some((s) => s.pictureStatus === 'PENDING')
            if (!hasPending) {
              pollAttempts.current = 0
              return false
            }
            if (pollAttempts.current >= MAX_POLL_ATTEMPTS) return false
            const interval = Math.min(BASE_POLL_INTERVAL_MS * 2 ** pollAttempts.current, MAX_POLL_INTERVAL_MS)
            pollAttempts.current += 1
            return interval
          },
    // Without this, a backgrounded/inactive native tab has its poll
    // throttled by timer coalescing. Irrelevant on web since refetchInterval
    // is disabled there.
    refetchIntervalInBackground: true,
  })
}
