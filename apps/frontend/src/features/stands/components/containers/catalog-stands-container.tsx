import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { getStandsPage } from '@/features/stands/api/stands.api'
import { standKeys } from '@/features/stands/api/stands.keys'
import { StandsScreen, type StandStatFilterKey } from '@/features/stands/components/presentational/stands-screen'
import { useStandOptions } from '@/features/stands/hooks/use-stand-options'
import type { StandFilters, StandResponse } from '@/features/stands/types/stands.types'
import { raritySchema, standStatSchema } from '@/shared/contracts/enums'
import { useDebouncedValue } from '@/shared/hooks/use-debounced-value'
import { usePaginatedCatalogue } from '@/shared/hooks/use-paginated-catalogue'

const STAT_FILTER_KEYS: StandStatFilterKey[] = [
  'attackPower',
  'speed',
  'attackRange',
  'endurance',
  'precision',
  'potential',
]

// Read-only counterpart to StandsContainer, for any logged-in user browsing
// the catalogue (see ObsidianVault/catalogo-publico-stands-devil-fruits-
// stages.md). Same search/filter wiring, minus everything write-only:
// mutations, clearEtags, the FAILED-picture toast watcher, and the
// translations fetch (that endpoint is admin-only and would 403 here).
export function CatalogStandsContainer() {
  const { t } = useTranslation()

  const [search, setSearch] = useState('')
  const debouncedSearch = useDebouncedValue(search)
  const [rarityFilter, setRarityFilter] = useState<string | null>(null)
  const [statFilters, setStatFilters] = useState<Record<StandStatFilterKey, string | null>>({
    attackPower: null,
    speed: null,
    attackRange: null,
    endurance: null,
    precision: null,
    potential: null,
  })
  const [evolvesFromFilter, setEvolvesFromFilter] = useState<string | null>(null)
  const [filtersExpanded, setFiltersExpanded] = useState(false)
  const [detailStand, setDetailStand] = useState<StandResponse | null>(null)

  const filters = useMemo(() => {
    const f: StandFilters = {}
    if (rarityFilter) f.rarity = rarityFilter as StandFilters['rarity']
    for (const key of STAT_FILTER_KEYS) {
      if (statFilters[key]) f[key] = statFilters[key] as StandFilters[typeof key]
    }
    if (debouncedSearch.trim()) f.q = debouncedSearch.trim()
    return f
  }, [rarityFilter, statFilters, debouncedSearch])

  const moreFiltersCount =
    STAT_FILTER_KEYS.filter((key) => statFilters[key]).length + (evolvesFromFilter ? 1 : 0)
  const hasActiveFilters = Boolean(rarityFilter) || moreFiltersCount > 0 || Boolean(filters.q)

  // The id/name-only /stands/options endpoint feeds the "Evolves From"
  // filter's own options - a full catalogue fetch here would duplicate the
  // grid's own payload and, under pagination, wouldn't even have the full
  // set. Same trap as the admin container's evolvesFromOptions: deriving
  // them from the filtered grid would make applying any filter narrow this
  // picker too.
  const { data: standOptions } = useStandOptions()

  const evolvesFromNameFilter = useMemo(() => {
    if (!evolvesFromFilter || !standOptions) return undefined
    return standOptions.find((s) => s.id === evolvesFromFilter)?.name
  }, [evolvesFromFilter, standOptions])

  const gridFilters = useMemo(
    () => (evolvesFromNameFilter ? { ...filters, evolvesFrom: evolvesFromNameFilter } : filters),
    [filters, evolvesFromNameFilter]
  )

  const appliedFilters = hasActiveFilters ? gridFilters : undefined
  const {
    items: stands,
    isLoading,
    isError,
    isFetchNextPageError,
    hasNextPage,
    isFetchingNextPage,
    fetchNextPage,
    total,
    refetch,
  } = usePaginatedCatalogue(
    standKeys.page(appliedFilters),
    (cursor, limit) => getStandsPage(appliedFilters, cursor, limit),
    { hasPendingPicture: (s) => s.pictureStatus === 'PENDING' }
  )

  const evolvesFromOptions = useMemo(
    () => (standOptions ?? []).map((s) => ({ value: s.id, label: s.name })),
    [standOptions]
  )
  const rarityFilterOptions = useMemo(
    () => raritySchema.options.map((v) => ({ value: v, label: t(`enums.rarity.${v}`) })),
    [t]
  )
  const statFilterOptions = useMemo(
    () => standStatSchema.options.map((v) => ({ value: v, label: t(`enums.standStat.${v}`) })),
    [t]
  )

  const onClearFilters = () => {
    setRarityFilter(null)
    setStatFilters({
      attackPower: null,
      speed: null,
      attackRange: null,
      endurance: null,
      precision: null,
      potential: null,
    })
    setEvolvesFromFilter(null)
  }

  return (
    <StandsScreen
      readOnly
      stands={stands ?? []}
      isLoading={isLoading}
      isError={isError}
      onRetry={() => void refetch()}
      search={search}
      onSearchChange={setSearch}
      rarityFilter={rarityFilter}
      rarityFilterOptions={rarityFilterOptions}
      onRarityFilterChange={setRarityFilter}
      statFilters={statFilters}
      statFilterOptions={statFilterOptions}
      onStatFilterChange={(key, value) => setStatFilters((prev) => ({ ...prev, [key]: value }))}
      evolvesFromFilter={evolvesFromFilter}
      evolvesFromFilterOptions={evolvesFromOptions}
      onEvolvesFromFilterChange={setEvolvesFromFilter}
      filtersExpanded={filtersExpanded}
      onToggleFilters={() => setFiltersExpanded((prev) => !prev)}
      moreFiltersCount={moreFiltersCount}
      onClearFilters={onClearFilters}
      hasActiveFilters={hasActiveFilters}
      detailStand={detailStand}
      onOpenDetail={setDetailStand}
      onCloseDetail={() => setDetailStand(null)}
      hasNextPage={hasNextPage}
      isFetchingNextPage={isFetchingNextPage}
      isLoadMoreError={isFetchNextPageError}
      onLoadMore={() => void fetchNextPage()}
      total={total}
    />
  )
}
