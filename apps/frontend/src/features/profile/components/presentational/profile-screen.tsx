import { Camera } from '@tamagui/lucide-icons-2'
import { Image } from 'react-native'
import { Paragraph, Spinner, XStack, YStack } from 'tamagui'

import { GlassField } from '@/shared/components/presentational/glass-field'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlassSelect } from '@/shared/components/presentational/glass-select'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlossOverlay } from '@/shared/components/presentational/gloss-overlay'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { PageShell } from '@/shared/components/presentational/page-shell'
import { InsetRing } from '@/shared/components/presentational/wii-card'
import { a11yProps } from '@/shared/lib/a11y'
import type { Locale } from '@/shared/lib/zod'
import type { ProfileUser } from '@/features/profile/types/profile.types'

import { ConfirmSheet } from './confirm-sheet'

// Language names stay in their own language regardless of the active UI
// locale (endonyms) - the standard convention for a language picker, so a
// Catalan speaker who ended up on the English UI still recognizes "Català"
// at a glance instead of hunting for "Catalan".
const LANGUAGE_OPTIONS: { value: Locale; label: string }[] = [
  { value: 'en-GB', label: 'English (UK)' },
  { value: 'es-ES', label: 'Español (España)' },
  { value: 'ca-ES', label: 'Català (Catalunya)' },
]

type ConfirmState = {
  visible: boolean
  isConfirming: boolean
  onConfirm: () => void
  onCancel: () => void
}

type Props = {
  profile: ProfileUser
  onPickAvatar: () => void
  isAvatarBusy: boolean
  username: string
  onUsernameChange: (text: string) => void
  usernameError?: string
  onSaveUsername: () => void
  isSavingUsername: boolean
  canSaveUsername: boolean
  onRequestRemoveAvatar: () => void
  onRequestDeleteAccount: () => void
  removeAvatarConfirm: ConfirmState
  deleteAccountConfirm: ConfirmState
  onChangeLanguage: (language: Locale) => void
  isSavingLanguage: boolean
}

// Pure UI — an Aero glass account settings screen. All form state, mutation
// wiring, and confirmation flow live in ProfileContainer; this file only
// turns props into JSX.
export function ProfileScreen({
  profile,
  onPickAvatar,
  isAvatarBusy,
  username,
  onUsernameChange,
  usernameError,
  onSaveUsername,
  isSavingUsername,
  canSaveUsername,
  onRequestRemoveAvatar,
  onRequestDeleteAccount,
  removeAvatarConfirm,
  deleteAccountConfirm,
  onChangeLanguage,
  isSavingLanguage,
}: Props) {
  const avatarUri = profile.avatar || null
  const hasCustomAvatar = profile.avatarStatus !== 'NONE'

  return (
    <YStack flex={1} position="relative">
      <PageShell align="top" scroll maxWidth={640}>
        <GlassPanel glossy elevate={2} width="100%" p="$6" gap="$5" items="center">
          <YStack position="relative" width={112} height={112}>
            <YStack
              width={112}
              height={112}
              rounded="$circle"
              overflow="hidden"
              position="relative"
              onPress={onPickAvatar}
              cursor="pointer"
              transition="bouncy"
              hoverStyle={{ scale: 1.05 }}
              pressStyle={{ scale: 0.95 }}
              {...a11yProps('Change profile picture', 'button', { disabled: isAvatarBusy })}
            >
              <InsetRing rounded="$circle" />
              <GlossOverlay coverage="third" shape="circle" />
              {avatarUri ? (
                <Image source={{ uri: avatarUri }} style={{ width: '100%', height: '100%' }} />
              ) : (
                <YStack flex={1} items="center" justify="center" bg="$grapeSoda">
                  <Paragraph color="white" fontSize="$8" fontWeight="800">
                    {profile.completeName.charAt(0).toUpperCase()}
                  </Paragraph>
                </YStack>
              )}
              {isAvatarBusy ? (
                <YStack
                  position="absolute"
                  t={0}
                  l={0}
                  r={0}
                  b={0}
                  items="center"
                  justify="center"
                  bg="rgba(10,12,20,0.45)"
                  style={{ pointerEvents: 'none' }}
                >
                  <Spinner size="large" color="white" />
                </YStack>
              ) : null}
            </YStack>

            <YStack
              position="absolute"
              b={0}
              r={0}
              width={36}
              height={36}
              rounded="$circle"
              items="center"
              justify="center"
              bg="$wiiBlue"
              borderWidth={1.5}
              borderColor="$glassEdge"
              onPress={onPickAvatar}
              cursor="pointer"
              hitSlop={8}
              transition="bouncy"
              pressStyle={{ scale: 0.9 }}
              {...a11yProps('Change profile picture', 'button', { disabled: isAvatarBusy })}
            >
              <Camera size={18} color="white" strokeWidth={2.5} />
            </YStack>
          </YStack>

          <GlowText level="label" align="center">
            {profile.avatarStatus === 'PENDING'
              ? 'Processing your new picture…'
              : profile.avatarStatus === 'FAILED'
                ? "We couldn't use that picture. Try another one."
                : 'Tap the camera to change your picture'}
          </GlowText>

          <YStack width="100%" gap="$3">
            <GlassField
              label="Username"
              value={username}
              onChangeText={onUsernameChange}
              error={usernameError}
              autoCapitalize="none"
              autoCorrect={false}
            />
            <GlossButton
              tone="blue"
              btnSize="md"
              disabled={!canSaveUsername || isSavingUsername}
              onPress={onSaveUsername}
              accessibilityLabel="Save username"
            >
              {isSavingUsername ? 'Saving…' : 'Save username'}
            </GlossButton>

            <GlassSelect
              label={isSavingLanguage ? 'Language (saving…)' : 'Language'}
              options={LANGUAGE_OPTIONS}
              value={profile.language}
              onChange={(value) => value && onChangeLanguage(value as Locale)}
            />
          </YStack>
        </GlassPanel>

        <GlassPanel tone="plastic" elevate={0} width="100%" p="$5" gap="$4">
          <GlowText level="heading" fontSize="$5">
            Account details
          </GlowText>

          <XStack flexWrap="wrap" gap="$4">
            <YStack gap="$1.5" flexBasis={220} grow={1}>
              <GlowText level="label" tone="soft">
                Full name
              </GlowText>
              <GlowText level="label">{profile.completeName}</GlowText>
              <GlowText level="label" tone="soft" fontSize="$2">
                From your Google account
              </GlowText>
            </YStack>

            <YStack gap="$1.5" flexBasis={220} grow={1}>
              <GlowText level="label" tone="soft">
                Email
              </GlowText>
              <GlowText level="label">{profile.email}</GlowText>
              <GlowText level="label" tone="soft" fontSize="$2">
                From your Google account
              </GlowText>
            </YStack>

            <YStack gap="$1.5" flexBasis={220} grow={1}>
              <GlowText level="label" tone="soft">
                Role
              </GlowText>
              <GlowText level="label">{profile.role}</GlowText>
              <GlowText level="label" tone="soft" fontSize="$2">
                Set by an administrator
              </GlowText>
            </YStack>
          </XStack>
        </GlassPanel>

        <YStack width="100%" my="$2" borderTopWidth={1.5} borderColor="$glassEdge" />

        <GlassPanel tone="strong" elevate={0} width="100%" p="$5" gap="$3">
          <GlowText level="heading" fontSize="$5" color="$strawHatRedDeep">
            Danger zone
          </GlowText>

          <GlossButton
            tone="glass"
            btnSize="md"
            disabled={!hasCustomAvatar}
            onPress={onRequestRemoveAvatar}
            accessibilityLabel="Remove avatar"
          >
            Remove avatar
          </GlossButton>

          <GlossButton
            tone="red"
            btnSize="md"
            onPress={onRequestDeleteAccount}
            accessibilityLabel="Delete account"
          >
            Delete account
          </GlossButton>
        </GlassPanel>
      </PageShell>

      <ConfirmSheet
        visible={removeAvatarConfirm.visible}
        title="Remove avatar?"
        message="Your picture will revert to the one from your Google account."
        confirmLabel="Remove avatar"
        isConfirming={removeAvatarConfirm.isConfirming}
        onConfirm={removeAvatarConfirm.onConfirm}
        onCancel={removeAvatarConfirm.onCancel}
      />
      <ConfirmSheet
        visible={deleteAccountConfirm.visible}
        title="Delete account?"
        message="This permanently deletes your account and everything in it. This can't be undone."
        confirmLabel="Delete account"
        isConfirming={deleteAccountConfirm.isConfirming}
        onConfirm={deleteAccountConfirm.onConfirm}
        onCancel={deleteAccountConfirm.onCancel}
      />
    </YStack>
  )
}
