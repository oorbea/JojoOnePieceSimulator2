import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'

import { createStand, deleteStand, updateStand, uploadStandPicture } from '@/features/stands/api/stands.api'
import { standKeys } from '@/features/stands/api/stands.keys'
import type { StandInput } from '@/features/stands/types/stands.types'
import { clearEtags } from '@/shared/api/etag'
import type { PickedPicture } from '@/shared/hooks/use-picture-picker'
import { showSuccessToast } from '@/shared/lib/toast'

// Errors are handled globally (MutationCache.onError in
// src/providers/query-provider.tsx) — these hooks only wire success
// feedback and cache invalidation. Called from within components, so
// useTranslation() works here same as anywhere else - see toast.ts's
// showErrorToast for the one case (MutationCache.onError) that isn't.

export function useCreateStand() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (input: StandInput) => createStand(input),
    onSuccess: () => {
      clearEtags()
      void queryClient.invalidateQueries({ queryKey: standKeys.allLocales })
      showSuccessToast(t('toasts.standCreated'))
    },
  })
}

export function useUpdateStand() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: StandInput }) => updateStand(id, input),
    onSuccess: () => {
      clearEtags()
      void queryClient.invalidateQueries({ queryKey: standKeys.allLocales })
      showSuccessToast(t('toasts.standUpdated'))
    },
  })
}

export function useUploadStandPicture() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, asset }: { id: string; asset: PickedPicture }) => uploadStandPicture(id, asset),
    onSuccess: () => {
      clearEtags()
      void queryClient.invalidateQueries({ queryKey: standKeys.allLocales })
      showSuccessToast(t('toasts.standPictureUploading'))
    },
  })
}

export function useDeleteStand() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => deleteStand(id),
    onSuccess: () => {
      clearEtags()
      void queryClient.invalidateQueries({ queryKey: standKeys.allLocales })
      showSuccessToast(t('toasts.standDeleted'))
    },
  })
}
