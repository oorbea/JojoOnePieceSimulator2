import { zodResolver } from '@hookform/resolvers/zod'
import { useQueryClient } from '@tanstack/react-query'
import { useMemo, useState } from 'react'
import { useForm } from 'react-hook-form'

import { getStandTranslations } from '@/features/stands/api/stands.api'
import { standKeys } from '@/features/stands/api/stands.keys'
import { StandsScreen } from '@/features/stands/components/presentational/stands-screen'
import { useStands } from '@/features/stands/hooks/use-stands'
import {
  useCreateStand,
  useDeleteStand,
  useUpdateStand,
  useUploadStandPicture,
} from '@/features/stands/hooks/use-stand-mutations'
import { standFormSchema, type StandFormValues, type StandInput, type StandResponse } from '@/features/stands/types/stands.types'
import type { PickedPicture } from '@/shared/hooks/use-picture-picker'
import { usePicturePicker } from '@/shared/hooks/use-picture-picker'
import { DEFAULT_LOCALE } from '@/shared/i18n'
import { createEmptyTranslationsForm, fromTranslationsResponse, toTranslationsPayload } from '@/shared/lib/power-translations'
import type { Locale } from '@/shared/lib/zod'

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
  const { translations, ...rest } = values
  return { ...rest, translations: toTranslationsPayload(translations) }
}

export function StandsContainer() {
  const { data: stands, isLoading, isError, refetch } = useStands()
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
    if (!stands) return []
    const editingId = modalState.editingStand?.id
    return stands
      .filter((s) => s.id !== editingId && s.evolvesFrom?.id !== editingId)
      .map((s) => ({ value: s.id, label: s.name }))
  }, [stands, modalState.editingStand])

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
          if (pendingPicture) uploadPictureMutation.mutate({ id: created.id, asset: pendingPicture })
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
          if (pendingPicture) uploadPictureMutation.mutate({ id: editingStand.id, asset: pendingPicture })
          closeModal()
        },
      }
    )
  }, jumpToFirstErroredLocale)

  const onConfirmDelete = () => {
    if (!standToDelete) return
    deleteMutation.mutate(standToDelete.id, { onSuccess: () => setStandToDelete(null) })
  }

  const pictureUri = pendingPicture?.uri ?? modalState.editingStand?.pictureThumb ?? null

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
