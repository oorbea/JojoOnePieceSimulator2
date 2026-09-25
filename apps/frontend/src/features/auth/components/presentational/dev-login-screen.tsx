import { LogIn, ShieldCheck, User } from '@tamagui/lucide-icons-2'
import { useTranslation } from 'react-i18next'
import { Spinner, XStack, YStack } from 'tamagui'

import { GlassField } from '@/shared/components/presentational/glass-field'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { PageShell } from '@/shared/components/presentational/page-shell'
import { SpeechBubble } from '@/shared/components/presentational/speech-bubble'
import type { AppError } from '@/shared/api/errors'
import type { RecentDevAccount } from '@/features/auth/lib/dev-login-recent-accounts'

type Props = {
  name: string
  onNameChange: (name: string) => void
  nameError: string | null
  admin: boolean
  onAdminChange: (admin: boolean) => void
  onSubmit: () => void
  isLoading: boolean
  error: AppError | null
  recentAccounts: RecentDevAccount[]
  onPickRecentAccount: (account: RecentDevAccount) => void
  onBackToGoogle: () => void
}

// Pure UI, same Wii-Party-glass language as login-screen.tsx - this is a
// second entry point into the exact same visual family, not a separate
// "debug tool" look, so it never reads as less trustworthy than the real
// login while still being unmistakably a dev affordance (title/subtitle say
// so outright).
export function DevLoginScreen({
  name,
  onNameChange,
  nameError,
  admin,
  onAdminChange,
  onSubmit,
  isLoading,
  error,
  recentAccounts,
  onPickRecentAccount,
  onBackToGoogle,
}: Props) {
  const { t } = useTranslation()

  return (
    <PageShell align="center" maxWidth={440}>
      <GlassPanel
        glossy
        radiusSize="hero"
        elevate={3}
        width="100%"
        p="$6"
        items="center"
        gap="$4"
        transition="bouncy"
        enterStyle={{ scale: 0.9, y: 20, opacity: 0 }}
      >
        <GlowText level="title" align="center">
          {t('devLogin.title')}
        </GlowText>
        <GlowText level="label" align="center">
          {t('devLogin.subtitle')}
        </GlowText>

        <YStack width="100%" gap="$1.5">
          <GlassField
            label={t('devLogin.nameLabel')}
            placeholder={t('devLogin.namePlaceholder')}
            value={name}
            onChangeText={(text) => onNameChange(text.toLowerCase())}
            error={nameError ? t(nameError) : undefined}
            autoCapitalize="none"
            autoCorrect={false}
          />
          {!nameError ? (
            <GlowText level="label" color="$panelTextSoft">
              {t('devLogin.nameHint')}
            </GlowText>
          ) : null}
        </YStack>

        <YStack width="100%" gap="$1.5">
          <GlowText level="label">{t('devLogin.roleLabel')}</GlowText>
          <XStack width="100%" gap="$2.5" role="radiogroup">
            <GlossButton
              flex={1}
              tone={admin ? 'glass' : 'blue'}
              btnSize="sm"
              icon={() => <User size={16} color={admin ? undefined : 'white'} />}
              a11yRole="radio"
              a11yChecked={!admin}
              accessibilityLabel={t('devLogin.roleRegular')}
              onPress={() => onAdminChange(false)}
            >
              {t('devLogin.roleRegular')}
            </GlossButton>
            <GlossButton
              flex={1}
              tone={admin ? 'yellow' : 'glass'}
              btnSize="sm"
              icon={() => <ShieldCheck size={16} color={admin ? '$inkBlack' : undefined} />}
              a11yRole="radio"
              a11yChecked={admin}
              accessibilityLabel={t('devLogin.roleAdmin')}
              onPress={() => onAdminChange(true)}
            >
              {t('devLogin.roleAdmin')}
            </GlossButton>
          </XStack>
        </YStack>

        <GlossButton
          tone="blue"
          btnSize="lg"
          shape="pill"
          flare
          width="100%"
          disabled={isLoading}
          icon={isLoading ? () => <Spinner color="white" /> : () => <LogIn size={20} color="white" />}
          onPress={onSubmit}
          accessibilityLabel={t('devLogin.submit')}
        >
          {isLoading ? t('devLogin.submitting') : t('devLogin.submit')}
        </GlossButton>

        {error ? (
          <SpeechBubble tailSide="top" tone="strong" width="100%">
            <GlowText level="label" align="center" color="$strawHatRedDeep">
              {error.message}
            </GlowText>
          </SpeechBubble>
        ) : null}

        {recentAccounts.length > 0 ? (
          <YStack width="100%" gap="$2">
            <GlowText level="label">{t('devLogin.recentAccounts')}</GlowText>
            <XStack width="100%" gap="$2" flexWrap="wrap">
              {recentAccounts.map((account) => (
                <GlossButton
                  key={account.name}
                  tone="glass"
                  btnSize="sm"
                  shape="pill"
                  icon={
                    account.admin ? () => <ShieldCheck size={14} color="$sunYellowDeep" /> : undefined
                  }
                  accessibilityLabel={account.name}
                  onPress={() => onPickRecentAccount(account)}
                >
                  {account.name}
                </GlossButton>
              ))}
            </XStack>
          </YStack>
        ) : null}

        <GlossButton
          tone="glass"
          btnSize="sm"
          onPress={onBackToGoogle}
          accessibilityLabel={t('devLogin.backToGoogle')}
        >
          {t('devLogin.backToGoogle')}
        </GlossButton>
      </GlassPanel>
    </PageShell>
  )
}
