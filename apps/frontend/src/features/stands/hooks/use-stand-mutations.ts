import { useMutation, useQueryClient } from '@tanstack/react-query'

import { createStand, deleteStand, updateStand, uploadStandPicture } from '@/features/stands/api/stands.api'
import { standKeys } from '@/features/stands/api/stands.keys'
import type { StandInput } from '@/features/stands/types/stands.types'
import type { PickedPicture } from '@/shared/hooks/use-picture-picker'
import { showSuccessToast } from '@/shared/lib/toast'

// Errors are handled globally (MutationCache.onError in
// src/providers/query-provider.tsx) — these hooks only wire success
// feedback and cache invalidation.

export function useCreateStand() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (input: StandInput) => createStand(input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: standKeys.all })
      showSuccessToast('Stand created')
    },
  })
}

export function useUpdateStand() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: StandInput }) => updateStand(id, input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: standKeys.all })
      showSuccessToast('Stand updated')
    },
  })
}

export function useUploadStandPicture() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, asset }: { id: string; asset: PickedPicture }) => uploadStandPicture(id, asset),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: standKeys.all })
      showSuccessToast('Uploading picture. This takes a moment.')
    },
  })
}

export function useDeleteStand() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => deleteStand(id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: standKeys.all })
      showSuccessToast('Stand deleted')
    },
  })
}
