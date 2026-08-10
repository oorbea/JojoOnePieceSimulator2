import { useQuery, type Query } from '@tanstack/react-query'
import { useRef } from 'react'
import { Platform } from 'react-native'

import { getStands } from '@/features/stands/api/stands.api'
import { standKeys } from '@/features/stands/api/stands.keys'
import type { StandFilters, StandResponse } from '@/features/stands/types/stands.types'

// On web, PictureEventsBridge (src/providers/picture-events-bridge.tsx)
// pushes picture-pipeline completion via SSE instead - polling here would
// be redundant. React Native has no EventSource, so native keeps this
// polling as its fallback; same reasoning as use-profile.ts's pollInterval.
const MAX_POLL_ATTEMPTS = 8
const MAX_POLL_INTERVAL_MS = 30_000
const BASE_POLL_INTERVAL_MS = 2_000

export function useStands(filters?: StandFilters) {
  // Attempt count must live outside query.state: `dataUpdateCount` is
  // cumulative for this query key's whole lifetime, including fetches from
  // before this component mounted - and, with PersistQueryClientProvider
  // persisting the cache to storage, from *previous page loads* too. Basing
  // the give-up threshold on it meant polling silently went permanently
  // dead for a query key once 8 refetches had ever happened, with no way to
  // observe a new PENDING picture except refetchOnWindowFocus (blur/focus).
  // A ref scoped to this hook instance resets every time nothing's pending,
  // so it only ever counts consecutive polls within the current wait.
  const pollAttempts = useRef(0)

  return useQuery({
    queryKey: standKeys.list(filters),
    queryFn: () => getStands(filters),
    refetchInterval:
      Platform.OS === 'web'
        ? undefined
        : (query: Query<StandResponse[]>) => {
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
