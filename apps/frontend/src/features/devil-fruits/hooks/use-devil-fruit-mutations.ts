import { useMutation, useQueryClient } from '@tanstack/react-query'

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
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (input: DevilFruitInput) => createDevilFruit(input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: devilFruitKeys.allLocales })
      showSuccessToast('Devil Fruit created')
    },
  })
}

export function useUpdateDevilFruit() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: DevilFruitInput }) => updateDevilFruit(id, input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: devilFruitKeys.allLocales })
      showSuccessToast('Devil Fruit updated')
    },
  })
}

export function useUploadDevilFruitPicture() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, asset }: { id: string; asset: PickedPicture }) => uploadDevilFruitPicture(id, asset),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: devilFruitKeys.allLocales })
      showSuccessToast('Uploading picture. This takes a moment.')
    },
  })
}

export function useDeleteDevilFruit() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => deleteDevilFruit(id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: devilFruitKeys.allLocales })
      showSuccessToast('Devil Fruit deleted')
    },
  })
}
