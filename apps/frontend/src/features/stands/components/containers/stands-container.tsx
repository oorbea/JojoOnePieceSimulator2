import { zodResolver } from '@hookform/resolvers/zod'
import { useQueryClient } from '@tanstack/react-query'
import { toast } from 'burnt'
import { useEffect, useMemo, useRef, useState } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'

import { getStandTranslations } from '@/features/stands/api/stands.api'
import { standKeys } from '@/features/stands/api/stands.keys'
import {
  StandsScreen,
  type StandStatFilterKey,
} from '@/features/stands/components/presentational/stands-screen'
import { useStands } from '@/features/stands/hooks/use-stands'
import {
  useCreateStand,
  useDeleteStand,
  useUpdateStand,
  useUploadStandPicture,
} from '@/features/stands/hooks/use-stand-mutations'
import {
  standFormSchema,
  type StandFilters,
  type StandFormValues,
  type StandInput,
  type StandResponse,
} from '@/features/stands/types/stands.types'
import { useDebouncedValue } from '@/shared/hooks/use-debounced-value'
import type { PickedPicture } from '@/shared/hooks/use-picture-picker'
import { usePicturePicker } from '@/shared/hooks/use-picture-picker'
import { DEFAULT_LOCALE } from '@/shared/i18n'
import {
  createEmptyTranslationsForm,
  fromTranslationsResponse,
  toTranslationsPayload,
} from '@/shared/lib/power-translations'
import { raritySchema, standStatSchema, type Locale } from '@/shared/contracts/enums'

// Every stat filter key StandFilters exposes, in the order the "more
// filters" panel renders them - one useState per key would be six near
// identical declarations, so they live in a single Record instead.
const STAT_FILTER_KEYS = [
  'attackPower',
  'speed',
  'attackRange',
  'endurance',
  'precision',
  'potential',
] as const
type StatFilterKey = StandStatFilterKey

// A function, not a constant object: reset()/useForm's defaultValues become
// this form's live nested state, so every "create" open needs its own
// fresh translations tree instead of sharing one mutable object across opens.
function createDefaultValues(): StandFormValues {
  return {
    name: '',
    translations: createEmptyTranslationsForm(),
    rarity: 'COMMON',
    attackPower: 'NULL',
    speed: 'NULL',
    attackRange: 'NULL',
    endurance: 'NULL',
    precision: 'NULL',
    potential: 'NULL',
    evolvesFromId: null,
  }
}

function toInput(values: StandFormValues): StandInput {
  const { translations, evolvesFromId, ...rest } = values
  // The generated StandInput's evolvesFromId is `?:` (omittable), not
  // `| null` - the backend's *string,omitempty can't distinguish an
  // explicit null from an absent key, so there's nothing lost in dropping
  // the key instead of sending null. The form itself keeps `| null` (a
  // controlled <select>'s "no evolution" needs a concrete falsy value).
  return {
    ...rest,
    translations: toTranslationsPayload(translations),
    evolvesFromId: evolvesFromId ?? undefined,
  }
}

export function StandsContainer() {
  const { t } = useTranslation()

  const [search, setSearch] = useState('')
  const debouncedSearch = useDebouncedValue(search)
  const [rarityFilter, setRarityFilter] = useState<string | null>(null)
  const [statFilters, setStatFilters] = useState<Record<StatFilterKey, string | null>>({
    attackPower: null,
    speed: null,
    attackRange: null,
    endurance: null,
    precision: null,
    potential: null,
  })
  const [evolvesFromFilter, setEvolvesFromFilter] = useState<string | null>(null)
  const [filtersExpanded, setFiltersExpanded] = useState(false)

  // evolvesFromFilter is a Stand id (what the picker needs to preselect a
  // value), but the backend's ?evolvesFrom= is the parent's *name* - resolve
  // it against the unfiltered roster below, once that's fetched.
  const filters = useMemo(() => {
    const f: StandFilters = {}
    if (rarityFilter) f.rarity = rarityFilter as StandFilters['rarity']
    for (const key of STAT_FILTER_KEYS) {
      if (statFilters[key]) f[key] = statFilters[key] as StandFilters[typeof key]
    }
    if (debouncedSearch.trim()) f.q = debouncedSearch.trim()
    return f
  }, [rarityFilter, statFilters, debouncedSearch])

  // Only counts the filters tucked inside the "More filters" disclosure
  // (stats + evolvesFrom) - the badge is about what's hidden, not about
  // search/rarity which are always visible in the row above it.
  const moreFiltersCount =
    STAT_FILTER_KEYS.filter((key) => statFilters[key]).length + (evolvesFromFilter ? 1 : 0)
  const hasActiveFilters = Boolean(rarityFilter) || moreFiltersCount > 0 || Boolean(filters.q)

  // The unfiltered roster feeds both the "Evolves From" filter's own option
  // list and the create/edit form's "Evolves From" picker - neither should
  // ever be narrowed by whatever's currently filtering the grid below, or
  // picking a parent Stand while a filter is active would silently exclude
  // valid choices. TanStack caches this under its own no-filter key
  // (standKeys.list(undefined)), so this is one extra request per screen
  // open, not per keystroke.
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

  const createMutation = useCreateStand()
  const updateMutation = useUpdateStand()
  const deleteMutation = useDeleteStand()
  const uploadPictureMutation = useUploadStandPicture()
  const { pickPicture } = usePicturePicker()
  const queryClient = useQueryClient()

  const [modalState, setModalState] = useState<{
    visible: boolean
    mode: 'create' | 'edit'
    editingStand: StandResponse | null
  }>({ visible: false, mode: 'create', editingStand: null })

  const [activeLocale, setActiveLocale] = useState<Locale>(DEFAULT_LOCALE)
  const [pendingPicture, setPendingPicture] = useState<PickedPicture | null>(null)
  const [standToDelete, setStandToDelete] = useState<StandResponse | null>(null)
  const [openingEditId, setOpeningEditId] = useState<string | null>(null)
  const [detailStand, setDetailStand] = useState<StandResponse | null>(null)

  const {
    control,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<StandFormValues>({
    resolver: zodResolver(standFormSchema),
    defaultValues: createDefaultValues(),
  })

  const evolvesFromOptions = useMemo(() => {
    if (!allStands) return []
    const editingId = modalState.editingStand?.id
    // On create, editingId is undefined - excluding "evolvesFrom === editingId"
    // in that case would exclude every Stand with no evolvesFrom set (i.e.
    // almost all of them), leaving the picker empty. Only apply that
    // exclusion while editing, where it actually prevents a Stand from
    // evolving from something that evolves from itself.
    return allStands
      .filter(
        (s) => s.id !== editingId && (editingId === undefined || s.evolvesFrom?.id !== editingId)
      )
      .map((s) => ({ value: s.id, label: s.name }))
  }, [allStands, modalState.editingStand])

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

  const openCreate = () => {
    reset(createDefaultValues())
    setPendingPicture(null)
    setActiveLocale(DEFAULT_LOCALE)
    setModalState({ visible: true, mode: 'create', editingStand: null })
  }

  // Waits for GET .../translations before opening the modal - a half-filled
  // form (public fields ready, translations still empty) would flicker
  // every locale's tab as blank for a beat, then jump non-en-GB tabs to
  // whatever they actually contain. Fetching first avoids that entirely,
  // no useEffect needed.
  const openEdit = async (stand: StandResponse) => {
    setOpeningEditId(stand.id)
    try {
      const translations = await queryClient.fetchQuery({
        queryKey: standKeys.translations(stand.id),
        queryFn: () => getStandTranslations(stand.id),
      })
      reset({
        name: stand.name,
        translations: fromTranslationsResponse(translations),
        rarity: stand.rarity,
        attackPower: stand.attackPower,
        speed: stand.speed,
        attackRange: stand.attackRange,
        endurance: stand.endurance,
        precision: stand.precision,
        potential: stand.potential,
        evolvesFromId: stand.evolvesFrom?.id ?? null,
      })
      setPendingPicture(null)
      setActiveLocale(DEFAULT_LOCALE)
      setModalState({ visible: true, mode: 'edit', editingStand: stand })
    } finally {
      setOpeningEditId(null)
    }
  }

  const closeModal = () => setModalState((prev) => ({ ...prev, visible: false }))

  const onPickPicture = async () => {
    const asset = await pickPicture()
    if (asset) setPendingPicture(asset)
  }

  const onSubmit = handleSubmit((values) => {
    const input = toInput(values)

    if (modalState.mode === 'create') {
      createMutation.mutate(input, {
        onSuccess: (created) => {
          if (pendingPicture)
            uploadPictureMutation.mutate({ id: created.id, asset: pendingPicture })
          closeModal()
        },
      })
      return
    }

    const editingStand = modalState.editingStand
    if (!editingStand) return

    updateMutation.mutate(
      { id: editingStand.id, input },
      {
        onSuccess: () => {
          if (pendingPicture)
            uploadPictureMutation.mutate({ id: editingStand.id, asset: pendingPicture })
          closeModal()
        },
      }
    )
  }, jumpToFirstErroredLocale)

  const onConfirmDelete = () => {
    if (!standToDelete) return
    deleteMutation.mutate(standToDelete.id, { onSuccess: () => setStandToDelete(null) })
  }

  const pictureUri = pendingPicture?.uri ?? (modalState.editingStand?.picture || null)

  // The picture worker is fire-and-forget with no push notification (see
  // ObsidianVault/backend-contract.md) - useStands' polling is what surfaces
  // a PENDING -> FAILED transition, and this is the only place that sees it
  // land, so it owns telling the user their upload didn't make it. Watches
  // the unfiltered roster, not the (possibly narrowed) grid list - a Stand
  // whose picture fails while a filter hides it would otherwise never
  // surface the toast.
  const previouslyPendingIds = useRef<Set<string>>(new Set())
  useEffect(() => {
    if (!allStands) return
    const failed = allStands.filter(
      (s) => s.pictureStatus === 'FAILED' && previouslyPendingIds.current.has(s.id)
    )
    failed.forEach(() => toast({ title: t('toasts.standPictureFailed'), preset: 'error' }))
    previouslyPendingIds.current = new Set(
      allStands.filter((s) => s.pictureStatus === 'PENDING').map((s) => s.id)
    )
  }, [allStands, t])

  // A translations.<locale> error is invisible if that locale's tab isn't
  // the active one - jump to the first one with an error on a failed
  // submit instead of leaving the user staring at a form that looks fine.
  function jumpToFirstErroredLocale(formErrors: typeof errors) {
    const erroredLocale = (['en-GB', 'es-ES', 'ca-ES'] as const).find(
      (locale) => formErrors.translations?.[locale]
    )
    if (erroredLocale) setActiveLocale(erroredLocale)
  }

  const erroredLocales = (['en-GB', 'es-ES', 'ca-ES'] as const).filter(
    (locale) => errors.translations?.[locale]
  )

  return (
    <StandsScreen
      stands={stands ?? []}
      isLoading={isLoading}
      isError={isError}
      onRetry={() => void refetch()}
      onCreateNew={openCreate}
      onEdit={(stand) => void openEdit(stand)}
      onDelete={setStandToDelete}
      openingEditId={openingEditId}
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
      form={{
        visible: modalState.visible,
        mode: modalState.mode,
        control,
        errors,
        onSubmit: () => void onSubmit(),
        onCancel: closeModal,
        isSaving: createMutation.isPending || updateMutation.isPending,
        evolvesFromOptions,
        pictureUri,
        onPickPicture: () => void onPickPicture(),
        isPictureBusy: uploadPictureMutation.isPending,
        activeLocale,
        onLocaleChange: setActiveLocale,
        erroredLocales,
      }}
      deleteConfirm={{
        visible: standToDelete !== null,
        isConfirming: deleteMutation.isPending,
        onConfirm: onConfirmDelete,
        onCancel: () => setStandToDelete(null),
        standName: standToDelete?.name,
      }}
    />
  )
}
