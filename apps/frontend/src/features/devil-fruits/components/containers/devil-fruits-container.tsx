import { zodResolver } from '@hookform/resolvers/zod'
import { useState } from 'react'
import { useForm } from 'react-hook-form'

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

const DEFAULT_VALUES: DevilFruitFormValues = {
  name: '',
  description: '',
  rarity: 'COMMON',
  skills: [],
  fruitType: 'PARAMECIA',
}

function toFormValues(devilFruit: DevilFruitResponse): DevilFruitFormValues {
  return {
    name: devilFruit.name,
    description: devilFruit.description,
    rarity: devilFruit.rarity,
    skills: devilFruit.skills,
    fruitType: devilFruit.fruitType,
  }
}

function toInput(values: DevilFruitFormValues): DevilFruitInput {
  return { ...values }
}

export function DevilFruitsContainer() {
  const { data: devilFruits, isLoading } = useDevilFruits()
  const createMutation = useCreateDevilFruit()
  const updateMutation = useUpdateDevilFruit()
  const deleteMutation = useDeleteDevilFruit()
  const uploadPictureMutation = useUploadDevilFruitPicture()
  const { pickPicture } = usePicturePicker()

  const [modalState, setModalState] = useState<{
    visible: boolean
    mode: 'create' | 'edit'
    editingFruit: DevilFruitResponse | null
  }>({ visible: false, mode: 'create', editingFruit: null })

  const [pendingPicture, setPendingPicture] = useState<PickedPicture | null>(null)
  const [fruitToDelete, setFruitToDelete] = useState<DevilFruitResponse | null>(null)

  const {
    control,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<DevilFruitFormValues>({
    resolver: zodResolver(devilFruitFormSchema),
    defaultValues: DEFAULT_VALUES,
  })

  const openCreate = () => {
    reset(DEFAULT_VALUES)
    setPendingPicture(null)
    setModalState({ visible: true, mode: 'create', editingFruit: null })
  }

  const openEdit = (devilFruit: DevilFruitResponse) => {
    reset(toFormValues(devilFruit))
    setPendingPicture(null)
    setModalState({ visible: true, mode: 'edit', editingFruit: devilFruit })
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
  })

  const onConfirmDelete = () => {
    if (!fruitToDelete) return
    deleteMutation.mutate(fruitToDelete.id, { onSuccess: () => setFruitToDelete(null) })
  }

  const pictureUri = pendingPicture?.uri ?? modalState.editingFruit?.pictureThumb ?? null

  return (
    <DevilFruitsScreen
      devilFruits={devilFruits ?? []}
      isLoading={isLoading}
      onCreateNew={openCreate}
      onEdit={openEdit}
      onDelete={setFruitToDelete}
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
