import { zodResolver } from '@hookform/resolvers/zod'
import { useMemo, useState } from 'react'
import { useForm } from 'react-hook-form'

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

const DEFAULT_VALUES: StandFormValues = {
  name: '',
  description: '',
  rarity: 'COMMON',
  skills: [],
  attackPower: 'NULL',
  speed: 'NULL',
  attackRange: 'NULL',
  endurance: 'NULL',
  precision: 'NULL',
  potential: 'NULL',
  evolvesFromId: null,
}

function toFormValues(stand: StandResponse): StandFormValues {
  return {
    name: stand.name,
    description: stand.description,
    rarity: stand.rarity,
    skills: stand.skills,
    attackPower: stand.attackPower,
    speed: stand.speed,
    attackRange: stand.attackRange,
    endurance: stand.endurance,
    precision: stand.precision,
    potential: stand.potential,
    evolvesFromId: stand.evolvesFrom?.id ?? null,
  }
}

function toInput(values: StandFormValues): StandInput {
  return { ...values }
}

export function StandsContainer() {
  const { data: stands, isLoading } = useStands()
  const createMutation = useCreateStand()
  const updateMutation = useUpdateStand()
  const deleteMutation = useDeleteStand()
  const uploadPictureMutation = useUploadStandPicture()
  const { pickPicture } = usePicturePicker()

  const [modalState, setModalState] = useState<{
    visible: boolean
    mode: 'create' | 'edit'
    editingStand: StandResponse | null
  }>({ visible: false, mode: 'create', editingStand: null })

  const [pendingPicture, setPendingPicture] = useState<PickedPicture | null>(null)
  const [standToDelete, setStandToDelete] = useState<StandResponse | null>(null)

  const {
    control,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<StandFormValues>({
    resolver: zodResolver(standFormSchema),
    defaultValues: DEFAULT_VALUES,
  })

  const evolvesFromOptions = useMemo(() => {
    if (!stands) return []
    const editingId = modalState.editingStand?.id
    return stands
      .filter((s) => s.id !== editingId && s.evolvesFrom?.id !== editingId)
      .map((s) => ({ value: s.id, label: s.name }))
  }, [stands, modalState.editingStand])

  const openCreate = () => {
    reset(DEFAULT_VALUES)
    setPendingPicture(null)
    setModalState({ visible: true, mode: 'create', editingStand: null })
  }

  const openEdit = (stand: StandResponse) => {
    reset(toFormValues(stand))
    setPendingPicture(null)
    setModalState({ visible: true, mode: 'edit', editingStand: stand })
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
  })

  const onConfirmDelete = () => {
    if (!standToDelete) return
    deleteMutation.mutate(standToDelete.id, { onSuccess: () => setStandToDelete(null) })
  }

  const pictureUri = pendingPicture?.uri ?? modalState.editingStand?.pictureThumb ?? null

  return (
    <StandsScreen
      stands={stands ?? []}
      isLoading={isLoading}
      onCreateNew={openCreate}
      onEdit={openEdit}
      onDelete={setStandToDelete}
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
