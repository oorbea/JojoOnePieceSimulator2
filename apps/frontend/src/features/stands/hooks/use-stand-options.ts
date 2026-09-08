import { useQuery } from '@tanstack/react-query'

import { getStandOptions } from '@/features/stands/api/stands.api'
import { standKeys } from '@/features/stands/api/stands.keys'

// Backs the evolvesFrom picker's id->name lookup without fetching the full
// catalogue - see stands.api.ts's getStandOptions.
export function useStandOptions() {
  return useQuery({
    queryKey: standKeys.options(),
    queryFn: getStandOptions,
  })
}
