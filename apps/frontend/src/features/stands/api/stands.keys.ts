import { queryKeys } from '@/shared/api/query-keys'
import { useLanguageStore } from '@/shared/stores/language.store'
import type { StandFilters } from '@/features/stands/types/stands.types'

// `all` is a function, not a static array: description/skills are resolved
// server-side per Accept-Language (see interceptors.ts), so a query cached
// under one locale must never answer a read after the user switches
// language - each locale gets its own branch of the cache instead.
// `allLocales` is the unlocalized prefix shared by every locale's branch -
// invalidateQueries matches by prefix, so mutations use this one to drop
// every locale's cached copy of a stand at once (a write's translations
// touch all locales together, see StandRequest.translations).
export const standKeys = {
  allLocales: [...queryKeys.root, 'stands'] as const,
  all: () => [...standKeys.allLocales, useLanguageStore.getState().locale] as const,
  list: (filters?: StandFilters) => [...standKeys.all(), 'list', filters ?? {}] as const,
  detail: (id: string) => [...standKeys.all(), 'detail', id] as const,
}
