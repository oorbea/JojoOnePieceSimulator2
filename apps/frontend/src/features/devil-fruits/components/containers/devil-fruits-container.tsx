import { zodResolver } from '@hookform/resolvers/zod'
import { useQueryClient } from '@tanstack/react-query'
import { toast } from 'burnt'
import { useEffect, useRef, useState } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'

import { getDevilFruitTranslations } from '@/features/devil-fruits/api/devil-fruits.api'
import { devilFruitKeys } from '@/features/devil-fruits/api/devil-fruits.keys'
import { DevilFruitsScreen } from '@/features/devil-fruits/components/presentational/devil-fruits-screen'
import { useDevilFruits } from '@/features/devil-fruits/hooks/use-devil-fruits'
import {
  useCreateDevilFruit,
  useDeleteDevilFruit,
  useUpdateDevilFruit,
  useUploadDevilFruitPicture,
} from '@/features/devil-fruits/hooks/use-devil-fruit-mutations'
import {
  devilFruitFormSchema,
  type DevilFruitFormValues,
  type DevilFruitInput,
  type DevilFruitResponse,
} from '@/features/devil-fruits/types/devil-fruits.types'
import type { PickedPicture } from '@/shared/hooks/use-picture-picker'
import { usePicturePicker } from '@/shared/hooks/use-picture-picker'
import { DEFAULT_LOCALE } from '@/shared/i18n'
import { createEmptyTranslationsForm, fromTranslationsResponse, toTranslationsPayload } from '@/shared/lib/power-translations'
import type { Locale } from '@/shared/lib/zod'

// A function, not a constant object - see stands-container.tsx's
// createDefaultValues for why.
function createDefaultValues(): DevilFruitFormValues {
  return {
    name: '',
    translations: createEmptyTranslationsForm(),
    rarity: 'COMMON',
    fruitType: 'PARAMECIA',
  }
}

function toInput(values: DevilFruitFormValues): DevilFruitInput {
  const { translations, ...rest } = values
  return { ...rest, translations: toTranslationsPayload(translations) }
}

export function DevilFruitsContainer() {
  const { t } = useTranslation()
  const { data: devilFruits, isLoading, isError, refetch } = useDevilFruits()
  const createMutation = useCreateDevilFruit()
  const updateMutation = useUpdateDevilFruit()
  const deleteMutation = useDeleteDevilFruit()
  const uploadPictureMutation = useUploadDevilFruitPicture()
  const { pickPicture } = usePicturePicker()
  const queryClient = useQueryClient()

  const [modalState, setModalState] = useState<{
    visible: boolean
    mode: 'create' | 'edit'
    editingFruit: DevilFruitResponse | null
  }>({ visible: false, mode: 'create', editingFruit: null })

  const [activeLocale, setActiveLocale] = useState<Locale>(DEFAULT_LOCALE)
  const [pendingPicture, setPendingPicture] = useState<PickedPicture | null>(null)
  const [fruitToDelete, setFruitToDelete] = useState<DevilFruitResponse | null>(null)
  const [openingEditId, setOpeningEditId] = useState<string | null>(null)

  const {
    control,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<DevilFruitFormValues>({
    resolver: zodResolver(devilFruitFormSchema),
    defaultValues: createDefaultValues(),
  })

  const openCreate = () => {
    reset(createDefaultValues())
    setPendingPicture(null)
    setActiveLocale(DEFAULT_LOCALE)
    setModalState({ visible: true, mode: 'create', editingFruit: null })
  }

  // See stands-container.tsx's openEdit for why this awaits the
  // translations fetch before opening the modal.
  const openEdit = async (devilFruit: DevilFruitResponse) => {
    setOpeningEditId(devilFruit.id)
    try {
      const translations = await queryClient.fetchQuery({
        queryKey: devilFruitKeys.translations(devilFruit.id),
        queryFn: () => getDevilFruitTranslations(devilFruit.id),
      })
      reset({
        name: devilFruit.name,
        translations: fromTranslationsResponse(translations),
        rarity: devilFruit.rarity,
        fruitType: devilFruit.fruitType,
      })
      setPendingPicture(null)
      setActiveLocale(DEFAULT_LOCALE)
      setModalState({ visible: true, mode: 'edit', editingFruit: devilFruit })
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

    const editingFruit = modalState.editingFruit
    if (!editingFruit) return

    updateMutation.mutate(
      { id: editingFruit.id, input },
      {
        onSuccess: () => {
          if (pendingPicture) uploadPictureMutation.mutate({ id: editingFruit.id, asset: pendingPicture })
          closeModal()
        },
      }
    )
  }, jumpToFirstErroredLocale)

  const onConfirmDelete = () => {
    if (!fruitToDelete) return
    deleteMutation.mutate(fruitToDelete.id, { onSuccess: () => setFruitToDelete(null) })
  }

  const pictureUri = pendingPicture?.uri ?? (modalState.editingFruit?.picture || null)

  // See stands-container.tsx's failed-upload effect for why this owns
  // surfacing a PENDING -> FAILED transition rather than the mutation hook.
  const previouslyPendingIds = useRef<Set<string>>(new Set())
  useEffect(() => {
    if (!devilFruits) return
    const failed = devilFruits.filter(
      (f) => f.pictureStatus === 'FAILED' && previouslyPendingIds.current.has(f.id)
    )
    failed.forEach(() => toast({ title: t('toasts.devilFruitPictureFailed'), preset: 'error' }))
    previouslyPendingIds.current = new Set(
      devilFruits.filter((f) => f.pictureStatus === 'PENDING').map((f) => f.id)
    )
  }, [devilFruits, t])

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
    <DevilFruitsScreen
      devilFruits={devilFruits ?? []}
      isLoading={isLoading}
      isError={isError}
      onRetry={() => void refetch()}
      onCreateNew={openCreate}
      onEdit={(devilFruit) => void openEdit(devilFruit)}
      onDelete={setFruitToDelete}
      openingEditId={openingEditId}
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
        visible: fruitToDelete !== null,
        isConfirming: deleteMutation.isPending,
        onConfirm: onConfirmDelete,
        onCancel: () => setFruitToDelete(null),
        fruitName: fruitToDelete?.name,
      }}
    />
  )
}
