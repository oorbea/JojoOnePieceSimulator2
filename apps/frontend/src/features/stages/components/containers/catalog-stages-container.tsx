import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { StagesScreen } from '@/features/stages/components/presentational/stages-screen'
import { useStages } from '@/features/stages/hooks/use-stages'
import type { StageInput, StageResponse } from '@/features/stages/types/stages.types'
import { useDebouncedValue } from '@/shared/hooks/use-debounced-value'
import { mangaSchema } from '@/shared/contracts/enums'

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

  const {
    data: stages,
    isLoading,
    isError,
    refetch,
  } = useStages(hasStageFilters ? stageFilters : undefined)

  // Same defensive client-side ordering as StagesContainer - the backend
  // already orders this way, this just guards against relying on it.
  const visibleStages = useMemo(() => {
    if (!stages) return []
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
    />
  )
}
