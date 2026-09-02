import { zodResolver } from '@hookform/resolvers/zod'
import { useQueryClient } from '@tanstack/react-query'
import { toast } from 'burnt'
import { useEffect, useMemo, useRef, useState } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'

import { getStageTranslations } from '@/features/stages/api/stages.api'
import { stageKeys } from '@/features/stages/api/stages.keys'
import { StagesScreen } from '@/features/stages/components/presentational/stages-screen'
import { useStages } from '@/features/stages/hooks/use-stages'
import {
  useCreateStage,
  useDeleteStage,
  useUpdateStage,
  useUploadStagePicture,
} from '@/features/stages/hooks/use-stage-mutations'
import {
  stageFormSchema,
  type StageFormValues,
  type StageInput,
  type StageResponse,
} from '@/features/stages/types/stages.types'
import { useDebouncedValue } from '@/shared/hooks/use-debounced-value'
import type { PickedPicture } from '@/shared/hooks/use-picture-picker'
import { usePicturePicker } from '@/shared/hooks/use-picture-picker'
import { DEFAULT_LOCALE, SUPPORTED_LOCALES } from '@/shared/i18n'
import {
  createEmptyStageTranslationsForm,
  fromStageTranslationsResponse,
  toStageTranslationsPayload,
} from '@/shared/lib/stage-translations'
import { mangaSchema, type Locale } from '@/shared/contracts/enums'

// A function, not a constant object - same reasoning as StandsContainer's
// createDefaultValues: reset()/useForm's defaultValues become this form's
// live nested state, so every "create" open needs its own fresh
// translations tree instead of sharing one mutable object across opens.
function createDefaultValues(): StageFormValues {
  return {
    manga: 'JOJO',
    order: 0,
    name: '',
    translations: createEmptyStageTranslationsForm(),
  }
}

function toInput(values: StageFormValues): StageInput {
  const { translations, ...rest } = values
  return { ...rest, translations: toStageTranslationsPayload(translations) }
}

export function StagesContainer() {
  const { t } = useTranslation()
  const [mangaFilter, setMangaFilter] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  // Search is server-side (?q=, ILIKE over name + the locale-resolved
  // description) - debounced so typing doesn't fire a request per
  // keystroke. The manga filter goes to the server too (?manga=); both
  // share one cache branch keyed by the combined filters object.
  const debouncedSearch = useDebouncedValue(search)
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
  const createMutation = useCreateStage()
  const updateMutation = useUpdateStage()
  const deleteMutation = useDeleteStage()
  const uploadPictureMutation = useUploadStagePicture()
  const { pickPicture } = usePicturePicker()
  const queryClient = useQueryClient()

  const [modalState, setModalState] = useState<{
    visible: boolean
    mode: 'create' | 'edit'
    editingStage: StageResponse | null
  }>({ visible: false, mode: 'create', editingStage: null })

  const [activeLocale, setActiveLocale] = useState<Locale>(DEFAULT_LOCALE)
  const [pendingPicture, setPendingPicture] = useState<PickedPicture | null>(null)
  const [stageToDelete, setStageToDelete] = useState<StageResponse | null>(null)
  const [openingEditId, setOpeningEditId] = useState<string | null>(null)

  const {
    control,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<StageFormValues>({
    resolver: zodResolver(stageFormSchema),
    defaultValues: createDefaultValues(),
  })

  // Search and manga are both server-side filters now (see stageFilters
  // above) - only the display order (grouped by manga, ordered by
  // position) is still client-side, since the backend's FilterStageRows
  // already orders that way but a defensive sort costs nothing here.
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

  const openCreate = () => {
    reset(createDefaultValues())
    setPendingPicture(null)
    setActiveLocale(DEFAULT_LOCALE)
    setModalState({ visible: true, mode: 'create', editingStage: null })
  }

  // Waits for GET .../translations before opening the modal - same reason
  // as StandsContainer.openEdit: avoids every locale tab flashing blank
  // then jumping to its real content.
  const openEdit = async (stage: StageResponse) => {
    setOpeningEditId(stage.id)
    try {
      const translations = await queryClient.fetchQuery({
        queryKey: stageKeys.translations(stage.id),
        queryFn: () => getStageTranslations(stage.id),
      })
      reset({
        manga: stage.manga,
        order: stage.order,
        name: stage.name,
        translations: fromStageTranslationsResponse(translations),
      })
      setPendingPicture(null)
      setActiveLocale(DEFAULT_LOCALE)
      setModalState({ visible: true, mode: 'edit', editingStage: stage })
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

    const editingStage = modalState.editingStage
    if (!editingStage) return

    updateMutation.mutate(
      { id: editingStage.id, input },
      {
        onSuccess: () => {
          if (pendingPicture)
            uploadPictureMutation.mutate({ id: editingStage.id, asset: pendingPicture })
          closeModal()
        },
      }
    )
  }, jumpToFirstErroredLocale)

  const onConfirmDelete = () => {
    if (!stageToDelete) return
    deleteMutation.mutate(stageToDelete.id, { onSuccess: () => setStageToDelete(null) })
  }

  const pictureUri = pendingPicture?.uri ?? (modalState.editingStage?.picture || null)

  // The picture worker is fire-and-forget with no push notification other
  // than SSE - useStages' polling (native) / the SSE bridge (web) is what
  // surfaces a PENDING -> FAILED transition, and this is the only place
  // that sees it land, so it owns telling the user their upload didn't make
  // it. Same pattern as StandsContainer.
  const previouslyPendingIds = useRef<Set<string>>(new Set())
  useEffect(() => {
    if (!stages) return
    const failed = stages.filter(
      (s) => s.pictureStatus === 'FAILED' && previouslyPendingIds.current.has(s.id)
    )
    failed.forEach(() => toast({ title: t('toasts.stagePictureFailed'), preset: 'error' }))
    previouslyPendingIds.current = new Set(
      stages.filter((s) => s.pictureStatus === 'PENDING').map((s) => s.id)
    )
  }, [stages, t])

  // A translations.<locale> error is invisible if that locale's tab isn't
  // the active one - jump to the first one with an error on a failed
  // submit instead of leaving the user staring at a form that looks fine.
  function jumpToFirstErroredLocale(formErrors: typeof errors) {
    const erroredLocale = SUPPORTED_LOCALES.find((locale) => formErrors.translations?.[locale])
    if (erroredLocale) setActiveLocale(erroredLocale)
  }

  const erroredLocales = SUPPORTED_LOCALES.filter((locale) => errors.translations?.[locale])

  return (
    <StagesScreen
      stages={visibleStages}
      isLoading={isLoading}
      isError={isError}
      onRetry={() => void refetch()}
      onCreateNew={openCreate}
      onEdit={(stage) => void openEdit(stage)}
      onDelete={setStageToDelete}
      openingEditId={openingEditId}
      search={search}
      onSearchChange={setSearch}
      mangaFilter={mangaFilter}
      mangaFilterOptions={mangaFilterOptions}
      onMangaFilterChange={setMangaFilter}
      hasActiveFilters={hasStageFilters}
      form={{
        visible: modalState.visible,
        mode: modalState.mode,
        control,
        errors,
        onSubmit: () => void onSubmit(),
        onCancel: closeModal,
        isSaving: createMutation.isPending || updateMutation.isPending,
        pictureUri,
        onPickPicture: () => void onPickPicture(),
        isPictureBusy: uploadPictureMutation.isPending,
        activeLocale,
        onLocaleChange: setActiveLocale,
        erroredLocales,
      }}
      deleteConfirm={{
        visible: stageToDelete !== null,
        isConfirming: deleteMutation.isPending,
        onConfirm: onConfirmDelete,
        onCancel: () => setStageToDelete(null),
        stageName: stageToDelete?.name,
      }}
    />
  )
}
