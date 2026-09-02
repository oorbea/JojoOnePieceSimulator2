import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'

import {
  deleteAccount,
  deleteAvatar,
  updateLanguage,
  updateUsername,
  uploadAvatar,
  type PickedAvatar,
} from '@/features/profile/api/profile.api'
import { profileKeys } from '@/features/profile/api/profile.keys'
import type { ProfileUser } from '@/features/profile/types/profile.types'
import type { Locale } from '@/shared/contracts/enums'
import { showSuccessToast } from '@/shared/lib/toast'
import { useSessionStore } from '@/shared/stores/session.store'

// Errors are handled globally (see MutationCache.onError in
// src/providers/query-provider.tsx) — these hooks only need to wire success
// feedback and keep the session store's copy of the user in sync so the nav
// shell and HomeScreen reflect a profile edit immediately, without reload.
function useSyncSessionOnSuccess() {
  const session = useSessionStore((state) => state.session)
  const setSession = useSessionStore((state) => state.setSession)

  return (user: ProfileUser) => {
    if (!session) return
    void setSession({
      ...session,
      user: {
        ...session.user,
        username: user.username,
        picture: user.avatar || null,
        language: user.language,
      },
    })
  }
}

export function useUpdateUsername() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const syncSession = useSyncSessionOnSuccess()

  return useMutation({
    mutationFn: updateUsername,
    onSuccess: (user) => {
      queryClient.setQueryData(profileKeys.me, user)
      syncSession(user)
      showSuccessToast(t('toasts.usernameUpdated'))
    },
  })
}

export function useUpdateLanguage() {
  const queryClient = useQueryClient()
  const syncSession = useSyncSessionOnSuccess()

  return useMutation({
    mutationFn: ({ username, language }: { username: string; language: Locale }) =>
      updateLanguage(username, language),
    onSuccess: (user) => {
      queryClient.setQueryData(profileKeys.me, user)
      syncSession(user)
      // Not shown as a toast on purpose - the visible language switch across
      // the whole UI (see language.store.ts, driven by session.user.language
      // via app/_layout.tsx) is already all the feedback this needs.
    },
  })
}

export function useUploadAvatar() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const syncSession = useSyncSessionOnSuccess()

  return useMutation({
    mutationFn: (asset: PickedAvatar) => uploadAvatar(asset),
    onSuccess: (user) => {
      queryClient.setQueryData(profileKeys.me, user)
      syncSession(user)
      showSuccessToast(t('toasts.avatarUploading'))
    },
  })
}

export function useDeleteAvatar() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const syncSession = useSyncSessionOnSuccess()

  return useMutation({
    mutationFn: deleteAvatar,
    onSuccess: (user) => {
      queryClient.setQueryData(profileKeys.me, user)
      syncSession(user)
      showSuccessToast(t('toasts.avatarRemoved'))
    },
  })
}

export function useDeleteAccount() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const clearSession = useSessionStore((state) => state.clearSession)

  return useMutation({
    mutationFn: deleteAccount,
    onSuccess: async () => {
      queryClient.removeQueries({ queryKey: profileKeys.me })
      await clearSession()
      showSuccessToast(t('toasts.accountDeleted'))
    },
  })
}
