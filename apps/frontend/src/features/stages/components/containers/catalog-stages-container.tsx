import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { getStagesPage } from '@/features/stages/api/stages.api'
import { stageKeys } from '@/features/stages/api/stages.keys'
import { StagesScreen } from '@/features/stages/components/presentational/stages-screen'
import type { StageInput, StageResponse } from '@/features/stages/types/stages.types'
import { mangaSchema } from '@/shared/contracts/enums'
import { useDebouncedValue } from '@/shared/hooks/use-debounced-value'
import { usePaginatedCatalogue } from '@/shared/hooks/use-paginated-catalogue'

// Read-only counterpart to StagesContainer - see
// CatalogStandsContainer's doc comment for what this deliberately drops.
export function CatalogStagesContainer() {
  const { t } = useTranslation()
  const [mangaFilter, setMangaFilter] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const debouncedSearch = useDebouncedValue(search)
  const [detailStage, setDetailStage] = useState<StageResponse | null>(null)

  const stageFilters = useMemo(() => {
    const filters: { manga?: StageInput['manga']; q?: string } = {}
    if (mangaFilter) filters.manga = mangaFilter as StageInput['manga']
    if (debouncedSearch.trim()) filters.q = debouncedSearch.trim()
    return filters
  }, [mangaFilter, debouncedSearch])
  const hasStageFilters = Object.keys(stageFilters).length > 0

  const appliedFilters = hasStageFilters ? stageFilters : undefined
  const {
    items: stages,
    isLoading,
    isError,
    isFetchNextPageError,
    hasNextPage,
    isFetchingNextPage,
    fetchNextPage,
    total,
    refetch,
  } = usePaginatedCatalogue(
    stageKeys.page(appliedFilters),
    (cursor, limit) => getStagesPage(appliedFilters, cursor, limit),
    { hasPendingPicture: (s) => s.pictureStatus === 'PENDING' }
  )

  // Same defensive client-side ordering as StagesContainer - the backend
  // already orders this way (and the paginated cursor depends on it, see
  // ObsidianVault/catalogue-pagination.md's ::manga cast trap), this just
  // guards against relying on it.
  const visibleStages = useMemo(() => {
    return [...stages].sort((a, b) =>
      a.manga === b.manga ? a.order - b.order : a.manga.localeCompare(b.manga)
    )
  }, [stages])

  const mangaFilterOptions = useMemo(
    () => mangaSchema.options.map((v) => ({ value: v, label: t(`enums.manga.${v}`) })),
    [t]
  )

  return (
    <StagesScreen
      readOnly
      stages={visibleStages}
      isLoading={isLoading}
      isError={isError}
      onRetry={() => void refetch()}
      search={search}
      onSearchChange={setSearch}
      mangaFilter={mangaFilter}
      mangaFilterOptions={mangaFilterOptions}
      onMangaFilterChange={setMangaFilter}
      hasActiveFilters={hasStageFilters}
      detailStage={detailStage}
      onOpenDetail={setDetailStage}
      onCloseDetail={() => setDetailStage(null)}
      hasNextPage={hasNextPage}
      isFetchingNextPage={isFetchingNextPage}
      isLoadMoreError={isFetchNextPageError}
      onLoadMore={() => void fetchNextPage()}
      total={total}
    />
  )
}
