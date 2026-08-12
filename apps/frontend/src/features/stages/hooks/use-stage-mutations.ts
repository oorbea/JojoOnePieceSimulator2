import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'

import {
  createStage,
  deleteStage,
  updateStage,
  uploadStagePicture,
} from '@/features/stages/api/stages.api'
import { stageKeys } from '@/features/stages/api/stages.keys'
import type { StageInput } from '@/features/stages/types/stages.types'
import { clearEtags } from '@/shared/api/etag'
import type { PickedPicture } from '@/shared/hooks/use-picture-picker'
import { showSuccessToast } from '@/shared/lib/toast'

// Errors are handled globally (MutationCache.onError in
// src/providers/query-provider.tsx) - these hooks only wire success feedback
// and cache invalidation, same shape as use-stand-mutations.ts.

export function useCreateStage() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (input: StageInput) => createStage(input),
    onSuccess: () => {
      clearEtags()
      void queryClient.invalidateQueries({ queryKey: stageKeys.allLocales })
      showSuccessToast(t('toasts.stageCreated'))
    },
  })
}

export function useUpdateStage() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: StageInput }) => updateStage(id, input),
    onSuccess: () => {
      clearEtags()
      void queryClient.invalidateQueries({ queryKey: stageKeys.allLocales })
      showSuccessToast(t('toasts.stageUpdated'))
    },
  })
}

export function useUploadStagePicture() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, asset }: { id: string; asset: PickedPicture }) =>
      uploadStagePicture(id, asset),
    onSuccess: () => {
      clearEtags()
      void queryClient.invalidateQueries({ queryKey: stageKeys.allLocales })
      showSuccessToast(t('toasts.stagePictureUploading'))
    },
  })
}

export function useDeleteStage() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => deleteStage(id),
    onSuccess: () => {
      clearEtags()
      void queryClient.invalidateQueries({ queryKey: stageKeys.allLocales })
      showSuccessToast(t('toasts.stageDeleted'))
    },
  })
}
