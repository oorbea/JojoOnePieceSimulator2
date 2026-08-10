import { useQuery, type Query } from '@tanstack/react-query'

import { getDevilFruits } from '@/features/devil-fruits/api/devil-fruits.api'
import { devilFruitKeys } from '@/features/devil-fruits/api/devil-fruits.keys'
import type { DevilFruitFilters, DevilFruitResponse } from '@/features/devil-fruits/types/devil-fruits.types'

// See use-stands.ts's pollInterval for why this polls rather than waiting
// for a push notification that doesn't exist.
const MAX_POLL_ATTEMPTS = 8
const MAX_POLL_INTERVAL_MS = 30_000
const BASE_POLL_INTERVAL_MS = 2_000

function pollInterval(query: Query<DevilFruitResponse[]>): number | false {
  const hasPending = query.state.data?.some((f) => f.pictureStatus === 'PENDING')
  if (!hasPending) return false
  if (query.state.dataUpdateCount >= MAX_POLL_ATTEMPTS) return false
  return Math.min(BASE_POLL_INTERVAL_MS * 2 ** query.state.dataUpdateCount, MAX_POLL_INTERVAL_MS)
}

export function useDevilFruits(filters?: DevilFruitFilters) {
  return useQuery({
    queryKey: devilFruitKeys.list(filters),
    queryFn: () => getDevilFruits(filters),
    refetchInterval: pollInterval,
  })
}
