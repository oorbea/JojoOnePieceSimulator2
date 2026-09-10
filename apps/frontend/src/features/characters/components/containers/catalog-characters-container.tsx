import { useMemo, useState } from 'react'

import {
  getJojoCharactersPage,
  getOnePieceCharactersPage,
} from '@/features/characters/api/characters.api'
import { characterKeys } from '@/features/characters/api/characters.keys'
import { CharactersScreen } from '@/features/characters/components/presentational/characters-screen'
import { JOJO_STAT_ROWS, ONE_PIECE_STAT_ROWS } from '@/features/characters/lib/character-stats'
import type {
  CharacterKind,
  JojoCharacterResponse,
  OnePieceCharacterResponse,
} from '@/features/characters/types/characters.types'
import { useDebouncedValue } from '@/shared/hooks/use-debounced-value'
import { usePaginatedCatalogue, type CataloguePage } from '@/shared/hooks/use-paginated-catalogue'

type AnyCharacterResponse = JojoCharacterResponse | OnePieceCharacterResponse

// Read-only counterpart to CharactersContainer - see
// CatalogStagesContainer's doc for what this deliberately drops. Only one
// of the two paginated queries is ever active (the manga toggle is
// exclusive), so this picks which fetch function/cache key to hand
// usePaginatedCatalogue based on the current selection instead of running
// both.
export function CatalogCharactersContainer() {
  const [mangaFilter, setMangaFilter] = useState<CharacterKind>('JOJO')
  const [search, setSearch] = useState('')
  const debouncedSearch = useDebouncedValue(search)
  const [detailCharacter, setDetailCharacter] = useState<
    JojoCharacterResponse | OnePieceCharacterResponse | null
  >(null)

  const filters = useMemo(() => {
    const f: { q?: string } = {}
    if (debouncedSearch.trim()) f.q = debouncedSearch.trim()
    return f
  }, [debouncedSearch])
  const hasActiveFilters = Object.keys(filters).length > 0
  const appliedFilters = hasActiveFilters ? filters : undefined

  const isJojo = mangaFilter === 'JOJO'
  const {
    items: characters,
    isLoading,
    isError,
    isFetchNextPageError,
    hasNextPage,
    isFetchingNextPage,
    fetchNextPage,
    total,
    refetch,
  } = usePaginatedCatalogue<AnyCharacterResponse>(
    isJojo ? characterKeys.jojo.page(appliedFilters) : characterKeys.onePiece.page(appliedFilters),
    (cursor, limit): Promise<CataloguePage<AnyCharacterResponse>> =>
      isJojo
        ? getJojoCharactersPage(appliedFilters, cursor, limit)
        : getOnePieceCharactersPage(appliedFilters, cursor, limit),
    { hasPendingPicture: (c) => c.pictureStatus === 'PENDING' }
  )

  const rows = isJojo ? JOJO_STAT_ROWS : ONE_PIECE_STAT_ROWS

  return (
    <CharactersScreen
      readOnly
      characters={characters}
      rows={rows as never}
      isLoading={isLoading}
      isError={isError}
      onRetry={() => void refetch()}
      search={search}
      onSearchChange={setSearch}
      mangaFilter={mangaFilter}
      onMangaFilterChange={setMangaFilter}
      hasActiveFilters={hasActiveFilters}
      detailCharacter={detailCharacter}
      onOpenDetail={setDetailCharacter}
      onCloseDetail={() => setDetailCharacter(null)}
      hasNextPage={hasNextPage}
      isFetchingNextPage={isFetchingNextPage}
      isLoadMoreError={isFetchNextPageError}
      onLoadMore={() => void fetchNextPage()}
      total={total}
    />
  )
}
