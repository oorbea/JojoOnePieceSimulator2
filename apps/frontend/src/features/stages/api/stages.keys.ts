import { queryKeys } from '@/shared/api/query-keys'
import { useLanguageStore } from '@/shared/stores/language.store'
import type { StageFilters } from '@/features/stages/types/stages.types'

// Same rationale as stands.keys.ts: `all` is a function since description is
// resolved server-side per Accept-Language, so each locale gets its own
// branch of the cache. `allLocales` is the unlocalized prefix mutations
// invalidate by, dropping every locale's cached copy at once.
export const stageKeys = {
  allLocales: [...queryKeys.root, 'stages'] as const,
  all: () => [...stageKeys.allLocales, useLanguageStore.getState().locale] as const,
  list: (filters?: StageFilters) => [...stageKeys.all(), 'list', filters ?? {}] as const,
  detail: (id: string) => [...stageKeys.all(), 'detail', id] as const,
  // Admin edit form only - carries every locale at once, so it hangs off
  // allLocales (not all()) and must not be branched by the active UI
  // locale. Mutations already invalidate allLocales, which drops this too.
  translations: (id: string) => [...stageKeys.allLocales, 'translations', id] as const,
}
