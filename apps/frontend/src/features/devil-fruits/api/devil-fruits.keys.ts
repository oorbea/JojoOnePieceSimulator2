import { queryKeys } from '@/shared/api/query-keys'
import { useLanguageStore } from '@/shared/stores/language.store'
import type { DevilFruitFilters } from '@/features/devil-fruits/types/devil-fruits.types'

// Same locale-branching shape as standKeys - see stands.keys.ts.
export const devilFruitKeys = {
  allLocales: [...queryKeys.root, 'devil-fruits'] as const,
  all: () => [...devilFruitKeys.allLocales, useLanguageStore.getState().locale] as const,
  list: (filters?: DevilFruitFilters) =>
    [...devilFruitKeys.all(), 'list', filters ?? {}] as const,
  detail: (id: string) => [...devilFruitKeys.all(), 'detail', id] as const,
}
