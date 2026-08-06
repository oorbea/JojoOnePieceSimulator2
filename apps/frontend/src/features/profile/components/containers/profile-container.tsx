import { zodResolver } from '@hookform/resolvers/zod'
import { useRouter } from 'expo-router'
import { useEffect, useState } from 'react'
import { useForm, useController } from 'react-hook-form'
import { useTranslation } from 'react-i18next'

import { ProfileScreen } from '@/features/profile/components/presentational/profile-screen'
import { useAvatarPicker } from '@/features/profile/hooks/use-avatar-picker'
import {
  useDeleteAccount,
  useDeleteAvatar,
  useUpdateLanguage,
  useUpdateUsername,
  useUploadAvatar,
} from '@/features/profile/hooks/use-profile-mutations'
import { useProfile } from '@/features/profile/hooks/use-profile'
import { usernameFormSchema, type UsernameFormValues } from '@/features/profile/types/profile.types'
import { LoadingScreen } from '@/shared/components/presentational/loading-screen'
import type { Locale } from '@/shared/lib/zod'

export function ProfileContainer() {
  const { t } = useTranslation()
  const router = useRouter()
  const { data: profile, isLoading } = useProfile()

  const {
    control,
    handleSubmit,
    reset,
    formState: { errors, isDirty },
  } = useForm<UsernameFormValues>({
    resolver: zodResolver(usernameFormSchema),
    defaultValues: { username: '' },
  })
  const {
    field: { value: username, onChange: onUsernameChange },
  } = useController({ name: 'username', control })

  // Seeds the form once the profile loads, and again if the username
  // changes from elsewhere (e.g. after a save round-trip) while the field
  // itself hasn't been touched since.
  useEffect(() => {
    if (profile && !isDirty) reset({ username: profile.username })
    // Only re-seed when the server value changes or the field becomes clean.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [profile?.username, isDirty])

  const updateUsernameMutation = useUpdateUsername()
  const updateLanguageMutation = useUpdateLanguage()
  const uploadAvatarMutation = useUploadAvatar()
  const deleteAvatarMutation = useDeleteAvatar()
  const deleteAccountMutation = useDeleteAccount()

  const { pickAvatar } = useAvatarPicker()

  const [isRemoveAvatarOpen, setIsRemoveAvatarOpen] = useState(false)
  const [isDeleteAccountOpen, setIsDeleteAccountOpen] = useState(false)

  if (isLoading || !profile) {
    return <LoadingScreen />
  }

  const onSaveUsername = handleSubmit((values) => {
    updateUsernameMutation.mutate(values.username, {
      onSuccess: () => reset({ username: values.username }),
    })
  })

  // Always sends profile.username (the last saved value), never the dirty
  // form field - a language change must never carry along an unsaved
  // username edit, since the backend's PATCH /users/me requires username on
  // every request (see dto.UpdateProfileRequest).
  const onChangeLanguage = (language: Locale) => {
    updateLanguageMutation.mutate({ username: profile.username, language })
  }

  const onPickAvatar = async () => {
    const asset = await pickAvatar()
    if (asset) uploadAvatarMutation.mutate(asset)
  }

  const onConfirmRemoveAvatar = () => {
    deleteAvatarMutation.mutate(undefined, { onSuccess: () => setIsRemoveAvatarOpen(false) })
  }

  const onConfirmDeleteAccount = () => {
    deleteAccountMutation.mutate(undefined, {
      onSuccess: () => {
        setIsDeleteAccountOpen(false)
        router.replace('/login')
      },
    })
  }

  return (
    <ProfileScreen
      profile={profile}
      onPickAvatar={() => void onPickAvatar()}
      isAvatarBusy={uploadAvatarMutation.isPending || profile.avatarStatus === 'PENDING'}
      username={username}
      onUsernameChange={onUsernameChange}
      usernameError={errors.username?.message && t(errors.username.message)}
      onSaveUsername={() => void onSaveUsername()}
      isSavingUsername={updateUsernameMutation.isPending}
      canSaveUsername={isDirty && !errors.username}
      onChangeLanguage={onChangeLanguage}
      isSavingLanguage={updateLanguageMutation.isPending}
      onRequestRemoveAvatar={() => setIsRemoveAvatarOpen(true)}
      onRequestDeleteAccount={() => setIsDeleteAccountOpen(true)}
      removeAvatarConfirm={{
        visible: isRemoveAvatarOpen,
        isConfirming: deleteAvatarMutation.isPending,
        onConfirm: onConfirmRemoveAvatar,
        onCancel: () => setIsRemoveAvatarOpen(false),
      }}
      deleteAccountConfirm={{
        visible: isDeleteAccountOpen,
        isConfirming: deleteAccountMutation.isPending,
        onConfirm: onConfirmDeleteAccount,
        onCancel: () => setIsDeleteAccountOpen(false),
      }}
    />
  )
}
