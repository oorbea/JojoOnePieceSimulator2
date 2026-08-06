import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'

import {
  createDevilFruit,
  deleteDevilFruit,
  updateDevilFruit,
  uploadDevilFruitPicture,
} from '@/features/devil-fruits/api/devil-fruits.api'
import { devilFruitKeys } from '@/features/devil-fruits/api/devil-fruits.keys'
import type { DevilFruitInput } from '@/features/devil-fruits/types/devil-fruits.types'
import type { PickedPicture } from '@/shared/hooks/use-picture-picker'
import { showSuccessToast } from '@/shared/lib/toast'

export function useCreateDevilFruit() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (input: DevilFruitInput) => createDevilFruit(input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: devilFruitKeys.allLocales })
      showSuccessToast(t('toasts.devilFruitCreated'))
    },
  })
}

export function useUpdateDevilFruit() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: DevilFruitInput }) => updateDevilFruit(id, input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: devilFruitKeys.allLocales })
      showSuccessToast(t('toasts.devilFruitUpdated'))
    },
  })
}

export function useUploadDevilFruitPicture() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, asset }: { id: string; asset: PickedPicture }) => uploadDevilFruitPicture(id, asset),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: devilFruitKeys.allLocales })
      showSuccessToast(t('toasts.devilFruitPictureUploading'))
    },
  })
}

export function useDeleteDevilFruit() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => deleteDevilFruit(id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: devilFruitKeys.allLocales })
      showSuccessToast(t('toasts.devilFruitDeleted'))
    },
  })
}
