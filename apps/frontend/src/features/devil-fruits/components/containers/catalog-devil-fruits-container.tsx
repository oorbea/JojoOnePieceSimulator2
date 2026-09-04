import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { DevilFruitsScreen } from '@/features/devil-fruits/components/presentational/devil-fruits-screen'
import { useDevilFruits } from '@/features/devil-fruits/hooks/use-devil-fruits'
import type { DevilFruitFilters, DevilFruitResponse } from '@/features/devil-fruits/types/devil-fruits.types'
import { useDebouncedValue } from '@/shared/hooks/use-debounced-value'
import { fruitTypeSchema, raritySchema } from '@/shared/contracts/enums'

// Read-only counterpart to DevilFruitsContainer - see
// CatalogStandsContainer's doc comment for what this deliberately drops.
export function CatalogDevilFruitsContainer() {
  const { t } = useTranslation()

  const [search, setSearch] = useState('')
  const debouncedSearch = useDebouncedValue(search)
  const [rarityFilter, setRarityFilter] = useState<string | null>(null)
  const [fruitTypeFilter, setFruitTypeFilter] = useState<string | null>(null)
  const [detailFruit, setDetailFruit] = useState<DevilFruitResponse | null>(null)

  const filters = useMemo(() => {
    const f: DevilFruitFilters = {}
    if (rarityFilter) f.rarity = rarityFilter as DevilFruitFilters['rarity']
    if (fruitTypeFilter) f.fruitType = fruitTypeFilter as DevilFruitFilters['fruitType']
    if (debouncedSearch.trim()) f.q = debouncedSearch.trim()
    return f
  }, [rarityFilter, fruitTypeFilter, debouncedSearch])
  const hasActiveFilters = Object.keys(filters).length > 0

  const {
    data: devilFruits,
    isLoading,
    isError,
    refetch,
  } = useDevilFruits(hasActiveFilters ? filters : undefined)

  const rarityFilterOptions = useMemo(
    () => raritySchema.options.map((v) => ({ value: v, label: t(`enums.rarity.${v}`) })),
    [t]
  )
  const fruitTypeFilterOptions = useMemo(
    () => fruitTypeSchema.options.map((v) => ({ value: v, label: t(`enums.fruitType.${v}`) })),
    [t]
  )

  return (
    <DevilFruitsScreen
      readOnly
      devilFruits={devilFruits ?? []}
      isLoading={isLoading}
      isError={isError}
      onRetry={() => void refetch()}
      search={search}
      onSearchChange={setSearch}
      rarityFilter={rarityFilter}
      rarityFilterOptions={rarityFilterOptions}
      onRarityFilterChange={setRarityFilter}
      fruitTypeFilter={fruitTypeFilter}
      fruitTypeFilterOptions={fruitTypeFilterOptions}
      onFruitTypeFilterChange={setFruitTypeFilter}
      hasActiveFilters={hasActiveFilters}
      detailFruit={detailFruit}
      onOpenDetail={setDetailFruit}
      onCloseDetail={() => setDetailFruit(null)}
    />
  )
}
