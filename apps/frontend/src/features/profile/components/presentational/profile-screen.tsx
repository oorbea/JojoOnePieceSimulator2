import { Camera } from '@tamagui/lucide-icons-2'
import { useTranslation } from 'react-i18next'
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
import { LOCALE_ENDONYMS, SUPPORTED_LOCALES } from '@/shared/i18n'
import type { ProfileUser } from '@/features/profile/types/profile.types'

import { ConfirmSheet } from './confirm-sheet'

// Endonyms shared with the admin forms' LocaleTabs - see shared/i18n/index.ts.
const LANGUAGE_OPTIONS: { value: Locale; label: string }[] = SUPPORTED_LOCALES.map((locale) => ({
  value: locale,
  label: LOCALE_ENDONYMS[locale],
}))

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
  const { t } = useTranslation()
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
              {...a11yProps(t('profile.changePicture'), 'button', { disabled: isAvatarBusy })}
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
              {...a11yProps(t('profile.changePicture'), 'button', { disabled: isAvatarBusy })}
            >
              <Camera size={18} color="white" strokeWidth={2.5} />
            </YStack>
          </YStack>

          <GlowText level="label" align="center">
            {profile.avatarStatus === 'PENDING'
              ? t('profile.avatarProcessing')
              : profile.avatarStatus === 'FAILED'
                ? t('profile.avatarFailed')
                : t('profile.avatarTapHint')}
          </GlowText>

          <YStack width="100%" gap="$3">
            <GlassField
              label={t('profile.username')}
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
              accessibilityLabel={t('profile.saveUsername')}
            >
              {isSavingUsername ? t('common.saving') : t('profile.saveUsername')}
            </GlossButton>

            <GlassSelect
              label={isSavingLanguage ? t('profile.languageSaving') : t('profile.language')}
              options={LANGUAGE_OPTIONS}
              value={profile.language}
              onChange={(value) => value && onChangeLanguage(value as Locale)}
            />
          </YStack>
        </GlassPanel>

        <GlassPanel tone="plastic" elevate={0} width="100%" p="$5" gap="$4">
          <GlowText level="heading" fontSize="$5">
            {t('profile.accountDetails')}
          </GlowText>

          <XStack flexWrap="wrap" gap="$4">
            <YStack gap="$1.5" flexBasis={220} grow={1}>
              <GlowText level="label" tone="soft">
                {t('profile.fullName')}
              </GlowText>
              <GlowText level="label">{profile.completeName}</GlowText>
              <GlowText level="label" tone="soft" fontSize="$2">
                {t('profile.fromGoogleAccount')}
              </GlowText>
            </YStack>

            <YStack gap="$1.5" flexBasis={220} grow={1}>
              <GlowText level="label" tone="soft">
                {t('profile.email')}
              </GlowText>
              <GlowText level="label">{profile.email}</GlowText>
              <GlowText level="label" tone="soft" fontSize="$2">
                {t('profile.fromGoogleAccount')}
              </GlowText>
            </YStack>

            <YStack gap="$1.5" flexBasis={220} grow={1}>
              <GlowText level="label" tone="soft">
                {t('profile.role')}
              </GlowText>
              <GlowText level="label">{t(`enums.role.${profile.role}`)}</GlowText>
              <GlowText level="label" tone="soft" fontSize="$2">
                {t('profile.setByAdmin')}
              </GlowText>
            </YStack>
          </XStack>
        </GlassPanel>

        <YStack width="100%" my="$2" borderTopWidth={1.5} borderColor="$glassEdge" />

        <GlassPanel tone="strong" elevate={0} width="100%" p="$5" gap="$3">
          <GlowText level="heading" fontSize="$5" color="$strawHatRedDeep">
            {t('profile.dangerZone')}
          </GlowText>

          <GlossButton
            tone="glass"
            btnSize="md"
            disabled={!hasCustomAvatar}
            onPress={onRequestRemoveAvatar}
            accessibilityLabel={t('profile.removeAvatar')}
          >
            {t('profile.removeAvatar')}
          </GlossButton>

          <GlossButton
            tone="red"
            btnSize="md"
            onPress={onRequestDeleteAccount}
            accessibilityLabel={t('profile.deleteAccount')}
          >
            {t('profile.deleteAccount')}
          </GlossButton>
        </GlassPanel>
      </PageShell>

      <ConfirmSheet
        visible={removeAvatarConfirm.visible}
        title={t('profile.removeAvatarConfirmTitle')}
        message={t('profile.removeAvatarConfirmMessage')}
        confirmLabel={t('profile.removeAvatar')}
        isConfirming={removeAvatarConfirm.isConfirming}
        onConfirm={removeAvatarConfirm.onConfirm}
        onCancel={removeAvatarConfirm.onCancel}
      />
      <ConfirmSheet
        visible={deleteAccountConfirm.visible}
        title={t('profile.deleteAccountConfirmTitle')}
        message={t('profile.deleteAccountConfirmMessage')}
        confirmLabel={t('profile.deleteAccount')}
        isConfirming={deleteAccountConfirm.isConfirming}
        onConfirm={deleteAccountConfirm.onConfirm}
        onCancel={deleteAccountConfirm.onCancel}
      />
    </YStack>
  )
}
