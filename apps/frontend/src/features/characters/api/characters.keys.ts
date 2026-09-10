import { queryKeys } from '@/shared/api/query-keys'
import { useLanguageStore } from '@/shared/stores/language.store'
import type {
  JojoCharacterFilters,
  OnePieceCharacterFilters,
} from '@/features/characters/types/characters.types'

// Two independent branches, one per kind - a JojoCharacter and an
// OnePieceCharacter are never the same cached list (two separate backend
// endpoints, two separate keyset paginations), so unlike stageKeys there is
// no single unlocalized root shared between them. Same allLocales/all()
// split as stageKeys within each branch - see its doc for why `all` is a
// function.
export const characterKeys = {
  jojo: {
    allLocales: [...queryKeys.root, 'jojo-characters'] as const,
    all: () => [...characterKeys.jojo.allLocales, useLanguageStore.getState().locale] as const,
    list: (filters?: JojoCharacterFilters) =>
      [...characterKeys.jojo.all(), 'list', filters ?? {}] as const,
    page: (filters?: JojoCharacterFilters) =>
      [...characterKeys.jojo.all(), 'page', filters ?? {}] as const,
    detail: (id: string) => [...characterKeys.jojo.all(), 'detail', id] as const,
    translations: (id: string) => [...characterKeys.jojo.allLocales, 'translations', id] as const,
  },
  onePiece: {
    allLocales: [...queryKeys.root, 'one-piece-characters'] as const,
    all: () => [...characterKeys.onePiece.allLocales, useLanguageStore.getState().locale] as const,
    list: (filters?: OnePieceCharacterFilters) =>
      [...characterKeys.onePiece.all(), 'list', filters ?? {}] as const,
    page: (filters?: OnePieceCharacterFilters) =>
      [...characterKeys.onePiece.all(), 'page', filters ?? {}] as const,
    detail: (id: string) => [...characterKeys.onePiece.all(), 'detail', id] as const,
    translations: (id: string) =>
      [...characterKeys.onePiece.allLocales, 'translations', id] as const,
  },
}
