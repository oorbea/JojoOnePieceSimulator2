import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { StandsScreen, type StandStatFilterKey } from '@/features/stands/components/presentational/stands-screen'
import { useStands } from '@/features/stands/hooks/use-stands'
import type { StandFilters, StandResponse } from '@/features/stands/types/stands.types'
import { useDebouncedValue } from '@/shared/hooks/use-debounced-value'
import { raritySchema, standStatSchema } from '@/shared/contracts/enums'

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

  // Unfiltered roster feeds the "Evolves From" filter's own options - same
  // trap as the admin container's evolvesFromOptions: deriving them from the
  // filtered grid would make applying any filter narrow this picker too.
  const { data: allStands } = useStands()

  const evolvesFromNameFilter = useMemo(() => {
    if (!evolvesFromFilter || !allStands) return undefined
    return allStands.find((s) => s.id === evolvesFromFilter)?.name
  }, [evolvesFromFilter, allStands])

  const gridFilters = useMemo(
    () => (evolvesFromNameFilter ? { ...filters, evolvesFrom: evolvesFromNameFilter } : filters),
    [filters, evolvesFromNameFilter]
  )

  const {
    data: stands,
    isLoading,
    isError,
    refetch,
  } = useStands(hasActiveFilters ? gridFilters : undefined)

  const evolvesFromOptions = useMemo(
    () => (allStands ?? []).map((s) => ({ value: s.id, label: s.name })),
    [allStands]
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
    />
  )
}
