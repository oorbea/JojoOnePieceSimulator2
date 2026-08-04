import { useMutation, useQueryClient } from '@tanstack/react-query'

import {
  deleteAccount,
  deleteAvatar,
  updateUsername,
  uploadAvatar,
  type PickedAvatar,
} from '@/features/profile/api/profile.api'
import { profileKeys } from '@/features/profile/api/profile.keys'
import type { ProfileUser } from '@/features/profile/types/profile.types'
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
      user: { ...session.user, username: user.username, picture: user.avatar || null },
    })
  }
}

export function useUpdateUsername() {
  const queryClient = useQueryClient()
  const syncSession = useSyncSessionOnSuccess()

  return useMutation({
    mutationFn: updateUsername,
    onSuccess: (user) => {
      queryClient.setQueryData(profileKeys.me, user)
      syncSession(user)
      showSuccessToast('Username updated')
    },
  })
}

export function useUploadAvatar() {
  const queryClient = useQueryClient()
  const syncSession = useSyncSessionOnSuccess()

  return useMutation({
    mutationFn: (asset: PickedAvatar) => uploadAvatar(asset),
    onSuccess: (user) => {
      queryClient.setQueryData(profileKeys.me, user)
      syncSession(user)
      showSuccessToast('Avatar uploading — this takes a moment')
    },
  })
}

export function useDeleteAvatar() {
  const queryClient = useQueryClient()
  const syncSession = useSyncSessionOnSuccess()

  return useMutation({
    mutationFn: deleteAvatar,
    onSuccess: (user) => {
      queryClient.setQueryData(profileKeys.me, user)
      syncSession(user)
      showSuccessToast('Avatar removed')
    },
  })
}

export function useDeleteAccount() {
  const queryClient = useQueryClient()
  const clearSession = useSessionStore((state) => state.clearSession)

  return useMutation({
    mutationFn: deleteAccount,
    onSuccess: async () => {
      queryClient.removeQueries({ queryKey: profileKeys.me })
      await clearSession()
      showSuccessToast('Account deleted')
    },
  })
}
