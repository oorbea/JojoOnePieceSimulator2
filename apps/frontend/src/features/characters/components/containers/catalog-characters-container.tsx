import { useMemo, useState } from 'react'

import {
  getJojoCharactersPage,
  getOnePieceCharactersPage,
} from '@/features/characters/api/characters.api'
import { characterKeys } from '@/features/characters/api/characters.keys'
import { CharactersScreen } from '@/features/characters/components/presentational/characters-screen'
import { rowsForCharacter } from '@/features/characters/lib/character-stats'
import type {
  CharacterMangaFilter,
  JojoCharacterResponse,
  OnePieceCharacterResponse,
  TaggedCharacter,
} from '@/features/characters/types/characters.types'
import { useDebouncedValue } from '@/shared/hooks/use-debounced-value'
import { usePaginatedCatalogue } from '@/shared/hooks/use-paginated-catalogue'

// Read-only counterpart to CharactersContainer - see CatalogStagesContainer's
// doc for what this deliberately drops. Both paginated queries are always
// mounted (only `enabled` toggles which one(s) actually fetch) so the 'ALL'
// filter can run them side by side and merge their loaded pages - each kind
// keeps its own cursor/hasNextPage, "load more" just advances whichever
// still has one. This is exactly the two-cursors-merged cost the original
// exclusive-filter design avoided; the owner asked for "see all of them"
// anyway, so the merge happens here instead of pushing kind-mixing into the
// backend's keyset pagination.
export function CatalogCharactersContainer() {
  const [mangaFilter, setMangaFilter] = useState<CharacterMangaFilter>('JOJO')
  const [search, setSearch] = useState('')
  const debouncedSearch = useDebouncedValue(search)
  const [detailCharacter, setDetailCharacter] = useState<TaggedCharacter | null>(null)

  const filters = useMemo(() => {
    const f: { q?: string } = {}
    if (debouncedSearch.trim()) f.q = debouncedSearch.trim()
    return f
  }, [debouncedSearch])
  const hasActiveFilters = Object.keys(filters).length > 0
  const appliedFilters = hasActiveFilters ? filters : undefined

  const wantJojo = mangaFilter !== 'ONE_PIECE'
  const wantOnePiece = mangaFilter !== 'JOJO'

  const jojoPage = usePaginatedCatalogue<JojoCharacterResponse>(
    characterKeys.jojo.page(appliedFilters),
    (cursor, limit) => getJojoCharactersPage(appliedFilters, cursor, limit),
    { hasPendingPicture: (c) => c.pictureStatus === 'PENDING', enabled: wantJojo }
  )
  const onePiecePage = usePaginatedCatalogue<OnePieceCharacterResponse>(
    characterKeys.onePiece.page(appliedFilters),
    (cursor, limit) => getOnePieceCharactersPage(appliedFilters, cursor, limit),
    { hasPendingPicture: (c) => c.pictureStatus === 'PENDING', enabled: wantOnePiece }
  )

  const characters = useMemo<TaggedCharacter[]>(() => {
    const jojoTagged = wantJojo ? jojoPage.items.map((c) => ({ ...c, kind: 'JOJO' as const })) : []
    const onePieceTagged = wantOnePiece
      ? onePiecePage.items.map((c) => ({ ...c, kind: 'ONE_PIECE' as const }))
      : []
    return [...jojoTagged, ...onePieceTagged]
  }, [wantJojo, wantOnePiece, jojoPage.items, onePiecePage.items])

  const total =
    wantJojo && wantOnePiece
      ? jojoPage.total !== undefined && onePiecePage.total !== undefined
        ? jojoPage.total + onePiecePage.total
        : undefined
      : wantJojo
        ? jojoPage.total
        : onePiecePage.total

  const onLoadMore = () => {
    if (wantJojo && jojoPage.hasNextPage) void jojoPage.fetchNextPage()
    if (wantOnePiece && onePiecePage.hasNextPage) void onePiecePage.fetchNextPage()
  }

  return (
    <CharactersScreen
      readOnly
      characters={characters}
      rows={rowsForCharacter}
      isLoading={(wantJojo && jojoPage.isLoading) || (wantOnePiece && onePiecePage.isLoading)}
      isError={(wantJojo && jojoPage.isError) || (wantOnePiece && onePiecePage.isError)}
      onRetry={() => {
        if (wantJojo) void jojoPage.refetch()
        if (wantOnePiece) void onePiecePage.refetch()
      }}
      search={search}
      onSearchChange={setSearch}
      mangaFilter={mangaFilter}
      onMangaFilterChange={setMangaFilter}
      hasActiveFilters={hasActiveFilters}
      detailCharacter={detailCharacter}
      onOpenDetail={setDetailCharacter}
      onCloseDetail={() => setDetailCharacter(null)}
      hasNextPage={(wantJojo && jojoPage.hasNextPage) || (wantOnePiece && onePiecePage.hasNextPage)}
      isFetchingNextPage={jojoPage.isFetchingNextPage || onePiecePage.isFetchingNextPage}
      isLoadMoreError={jojoPage.isFetchNextPageError || onePiecePage.isFetchNextPageError}
      onLoadMore={onLoadMore}
      total={total}
    />
  )
}
