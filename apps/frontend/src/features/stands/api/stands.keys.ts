import { queryKeys } from '@/shared/api/query-keys'
import type { StandFilters } from '@/features/stands/types/stands.types'

export const standKeys = {
  all: [...queryKeys.root, 'stands'] as const,
  list: (filters?: StandFilters) => [...standKeys.all, 'list', filters ?? {}] as const,
  detail: (id: string) => [...standKeys.all, 'detail', id] as const,
}
