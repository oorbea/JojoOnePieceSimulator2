import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'

import {
  createJojoCharacter,
  createOnePieceCharacter,
  deleteJojoCharacter,
  deleteOnePieceCharacter,
  updateJojoCharacter,
  updateOnePieceCharacter,
  uploadJojoCharacterPicture,
  uploadOnePieceCharacterPicture,
} from '@/features/characters/api/characters.api'
import { characterKeys } from '@/features/characters/api/characters.keys'
import type {
  JojoCharacterInput,
  OnePieceCharacterInput,
} from '@/features/characters/types/characters.types'
import { clearEtags } from '@/shared/api/etag'
import type { PickedPicture } from '@/shared/hooks/use-picture-picker'
import { showSuccessToast } from '@/shared/lib/toast'

// Errors are handled globally (MutationCache.onError) - same shape as
// use-stage-mutations.ts, cloned per kind since they hit different
// endpoints/cache branches.

export function useCreateJojoCharacter() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: JojoCharacterInput) => createJojoCharacter(input),
    onSuccess: () => {
      clearEtags()
      void queryClient.invalidateQueries({ queryKey: characterKeys.jojo.allLocales })
      showSuccessToast(t('toasts.characterCreated'))
    },
  })
}

export function useUpdateJojoCharacter() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: JojoCharacterInput }) =>
      updateJojoCharacter(id, input),
    onSuccess: () => {
      clearEtags()
      void queryClient.invalidateQueries({ queryKey: characterKeys.jojo.allLocales })
      showSuccessToast(t('toasts.characterUpdated'))
    },
  })
}

export function useUploadJojoCharacterPicture() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, asset }: { id: string; asset: PickedPicture }) =>
      uploadJojoCharacterPicture(id, asset),
    onSuccess: () => {
      clearEtags()
      void queryClient.invalidateQueries({ queryKey: characterKeys.jojo.allLocales })
      showSuccessToast(t('toasts.characterPictureUploading'))
    },
  })
}

export function useDeleteJojoCharacter() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => deleteJojoCharacter(id),
    onSuccess: () => {
      clearEtags()
      void queryClient.invalidateQueries({ queryKey: characterKeys.jojo.allLocales })
      showSuccessToast(t('toasts.characterDeleted'))
    },
  })
}

export function useCreateOnePieceCharacter() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: OnePieceCharacterInput) => createOnePieceCharacter(input),
    onSuccess: () => {
      clearEtags()
      void queryClient.invalidateQueries({ queryKey: characterKeys.onePiece.allLocales })
      showSuccessToast(t('toasts.characterCreated'))
    },
  })
}

export function useUpdateOnePieceCharacter() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: OnePieceCharacterInput }) =>
      updateOnePieceCharacter(id, input),
    onSuccess: () => {
      clearEtags()
      void queryClient.invalidateQueries({ queryKey: characterKeys.onePiece.allLocales })
      showSuccessToast(t('toasts.characterUpdated'))
    },
  })
}

export function useUploadOnePieceCharacterPicture() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, asset }: { id: string; asset: PickedPicture }) =>
      uploadOnePieceCharacterPicture(id, asset),
    onSuccess: () => {
      clearEtags()
      void queryClient.invalidateQueries({ queryKey: characterKeys.onePiece.allLocales })
      showSuccessToast(t('toasts.characterPictureUploading'))
    },
  })
}

export function useDeleteOnePieceCharacter() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => deleteOnePieceCharacter(id),
    onSuccess: () => {
      clearEtags()
      void queryClient.invalidateQueries({ queryKey: characterKeys.onePiece.allLocales })
      showSuccessToast(t('toasts.characterDeleted'))
    },
  })
}
