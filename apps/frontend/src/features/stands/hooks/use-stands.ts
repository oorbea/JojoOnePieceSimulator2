import { useQuery } from '@tanstack/react-query'

import { getStands } from '@/features/stands/api/stands.api'
import { standKeys } from '@/features/stands/api/stands.keys'
import type { StandFilters } from '@/features/stands/types/stands.types'

export function useStands(filters?: StandFilters) {
  return useQuery({
    queryKey: standKeys.list(filters),
    queryFn: () => getStands(filters),
  })
}
