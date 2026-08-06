import { useQuery } from '@tanstack/react-query'

import { getDevilFruits } from '@/features/devil-fruits/api/devil-fruits.api'
import { devilFruitKeys } from '@/features/devil-fruits/api/devil-fruits.keys'
import type { DevilFruitFilters } from '@/features/devil-fruits/types/devil-fruits.types'

export function useDevilFruits(filters?: DevilFruitFilters) {
  return useQuery({
    queryKey: devilFruitKeys.list(filters),
    queryFn: () => getDevilFruits(filters),
  })
}
