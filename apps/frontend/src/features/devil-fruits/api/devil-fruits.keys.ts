import { queryKeys } from '@/shared/api/query-keys'
import type { DevilFruitFilters } from '@/features/devil-fruits/types/devil-fruits.types'

export const devilFruitKeys = {
  all: [...queryKeys.root, 'devil-fruits'] as const,
  list: (filters?: DevilFruitFilters) => [...devilFruitKeys.all, 'list', filters ?? {}] as const,
  detail: (id: string) => [...devilFruitKeys.all, 'detail', id] as const,
}
