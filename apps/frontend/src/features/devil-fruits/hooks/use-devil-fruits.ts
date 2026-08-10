import { useQuery, type Query } from '@tanstack/react-query'
import { useRef } from 'react'
import { Platform } from 'react-native'

import { getDevilFruits } from '@/features/devil-fruits/api/devil-fruits.api'
import { devilFruitKeys } from '@/features/devil-fruits/api/devil-fruits.keys'
import type { DevilFruitFilters, DevilFruitResponse } from '@/features/devil-fruits/types/devil-fruits.types'

// See use-stands.ts's useStands for why this polls only on native (web uses
// PictureEventsBridge's SSE push instead), and why the attempt count lives
// in a ref instead of query.state.dataUpdateCount.
const MAX_POLL_ATTEMPTS = 8
const MAX_POLL_INTERVAL_MS = 30_000
const BASE_POLL_INTERVAL_MS = 2_000

export function useDevilFruits(filters?: DevilFruitFilters) {
  const pollAttempts = useRef(0)

  return useQuery({
    queryKey: devilFruitKeys.list(filters),
    queryFn: () => getDevilFruits(filters),
    refetchInterval:
      Platform.OS === 'web'
        ? undefined
        : (query: Query<DevilFruitResponse[]>) => {
            const hasPending = query.state.data?.some((f) => f.pictureStatus === 'PENDING')
            if (!hasPending) {
              pollAttempts.current = 0
              return false
            }
            if (pollAttempts.current >= MAX_POLL_ATTEMPTS) return false
            const interval = Math.min(BASE_POLL_INTERVAL_MS * 2 ** pollAttempts.current, MAX_POLL_INTERVAL_MS)
            pollAttempts.current += 1
            return interval
          },
    refetchIntervalInBackground: true,
  })
}
