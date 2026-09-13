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
  useUpdateAvatarFocalPoint,
  useUpdateLanguage,
  useUpdateUsername,
  useUploadAvatar,
} from '@/features/profile/hooks/use-profile-mutations'
import { useProfile } from '@/features/profile/hooks/use-profile'
import { usernameFormSchema, type UsernameFormValues } from '@/features/profile/types/profile.types'
import { LoadingScreen } from '@/shared/components/presentational/loading-screen'
import type { Locale } from '@/shared/contracts/enums'

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
  const updateAvatarFocalPointMutation = useUpdateAvatarFocalPoint()
  const uploadAvatarMutation = useUploadAvatar()
  const deleteAvatarMutation = useDeleteAvatar()
  const deleteAccountMutation = useDeleteAccount()

  const { pickAvatar } = useAvatarPicker()

  const [isRemoveAvatarOpen, setIsRemoveAvatarOpen] = useState(false)
  const [isDeleteAccountOpen, setIsDeleteAccountOpen] = useState(false)

  // Local draft so the profile screen's own avatar preview (rendered while
  // the framing modal is open on top of it) never waits on a round-trip -
  // committed once on the modal's Confirm, not per pixel of drag (that
  // replaced an earlier per-drag-tick debounced save - see
  // ObsidianVault/admin-locale-badge-and-focal-point-2026-09-13.md). See
  // stands-container.tsx's focalModal state for the mandatory-vs-reopened
  // distinction.
  const [avatarFocal, setAvatarFocal] = useState({
    x: profile?.avatarFocalX ?? 0.5,
    y: profile?.avatarFocalY ?? 0.5,
  })
  const [focalModal, setFocalModal] = useState<{ visible: boolean; mandatory: boolean }>({
    visible: false,
    mandatory: false,
  })
  // Reseeds avatarFocal whenever the server value changes (a fresh upload,
  // or another device) while the framing modal isn't open editing it.
  // Adjusted directly during render off a "have we seen this server value
  // yet" key (same pattern lobby-room-container.tsx uses) rather than a
  // useEffect, which would call setState synchronously inside the effect
  // body (react-hooks/set-state-in-effect).
  const serverFocalKey = profile ? `${profile.avatarFocalX}:${profile.avatarFocalY}` : null
  const [seenServerFocalKey, setSeenServerFocalKey] = useState(serverFocalKey)
  if (serverFocalKey !== seenServerFocalKey) {
    setSeenServerFocalKey(serverFocalKey)
    if (profile && !focalModal.visible) {
      setAvatarFocal({ x: profile.avatarFocalX, y: profile.avatarFocalY })
    }
  }

  const onAdjustFocal = () => setFocalModal({ visible: true, mandatory: false })

  const onConfirmFocal = (x: number, y: number) => {
    if (!profile) return
    setAvatarFocal({ x, y })
    updateAvatarFocalPointMutation.mutate({ username: profile.username, focalX: x, focalY: y })
    setFocalModal((prev) => ({ ...prev, visible: false }))
  }

  const onCancelFocal = () => setFocalModal((prev) => ({ ...prev, visible: false }))

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
    if (!asset) return
    uploadAvatarMutation.mutate(asset, {
      onSuccess: () => {
        // A newly uploaded avatar has no relationship to whatever focal
        // point the previous one had - reset to center (owner decision,
        // same as the catalogue forms' onPickPicture) and immediately ask
        // for the real framing, mandatory, the same as every other
        // picture-bearing resource.
        setAvatarFocal({ x: 0.5, y: 0.5 })
        setFocalModal({ visible: true, mandatory: true })
      },
    })
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
      avatarFocalX={avatarFocal.x}
      avatarFocalY={avatarFocal.y}
      onAdjustFocal={onAdjustFocal}
      focalModal={{
        visible: focalModal.visible,
        uri: profile.avatar || profile.avatarThumb || null,
        x: avatarFocal.x,
        y: avatarFocal.y,
        onConfirm: onConfirmFocal,
        onCancel: focalModal.mandatory ? undefined : onCancelFocal,
      }}
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
