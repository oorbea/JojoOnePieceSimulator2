import { ChevronLeft, Globe, Lock, Users } from '@tamagui/lucide-icons-2'
import { useTranslation } from 'react-i18next'
import { ActivityIndicator } from 'react-native'
import { XStack, YStack } from 'tamagui'

import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { PageShell } from '@/shared/components/presentational/page-shell'
import type { InvitePreview } from '@/features/game/types/game.types'

export type JoinInviteStatus = 'checking' | 'error' | 'preview' | 'joining'

type Props = {
  status: JoinInviteStatus
  // i18n key under game.invite.error.* - only meaningful when status ===
  // 'error'.
  errorKey?: string
  preview?: InvitePreview
  onJoin: () => void
  onBack: () => void
}

export function JoinInviteScreen({ status, errorKey, preview, onJoin, onBack }: Props) {
  const { t } = useTranslation()

  if (status === 'checking') {
    return (
      <PageShell align="center" maxWidth={480}>
        <ActivityIndicator />
        <GlowText level="label">{t('game.invite.checking')}</GlowText>
      </PageShell>
    )
  }

  if (status === 'error') {
    return (
      <PageShell align="center" maxWidth={480}>
        <GlowText level="title" align="center">
          {t(errorKey ?? 'game.invite.error.generic')}
        </GlowText>
        <GlossButton
          tone="blue"
          btnSize="lg"
          width="100%"
          onPress={onBack}
          accessibilityLabel={t('game.invite.backToGames')}
        >
          {t('game.invite.backToGames')}
        </GlossButton>
      </PageShell>
    )
  }

  return (
    <PageShell align="center" maxWidth={480}>
      <XStack width="100%" items="center" gap="$3">
        <GlossButton
          tone="glass"
          btnSize="sm"
          shape="circle"
          onPress={onBack}
          accessibilityLabel={t('common.cancel')}
        >
          <ChevronLeft size={18} color="$panelText" />
        </GlossButton>
        <GlowText level="title">{t('game.invite.title')}</GlowText>
      </XStack>

      {preview ? (
        <GlassPanel tone="strong" width="100%" p="$4" gap="$2">
          <XStack items="center" gap="$2">
            {preview.visibility === 'PUBLIC' ? (
              <Globe size={14} color="$panelTextSoft" />
            ) : (
              <Lock size={14} color="$panelTextSoft" />
            )}
            <GlowText level="heading">{t(`enums.gameMode.${preview.mode}`)}</GlowText>
          </XStack>
          <XStack items="center" gap="$1.5">
            <Users size={14} color="$panelTextSoft" />
            <GlowText level="label">
              {preview.playerCount}/{preview.maxPlayers}
            </GlowText>
          </XStack>
          <GlowText level="label">
            {t('game.preview.host', { name: preview.hostDisplayName })}
          </GlowText>
          {preview.locked ? (
            <GlowText level="label">{t('game.preview.lockedBadge')}</GlowText>
          ) : null}
        </GlassPanel>
      ) : (
        <ActivityIndicator />
      )}

      <YStack width="100%">
        <GlossButton
          tone="blue"
          btnSize="lg"
          width="100%"
          disabled={!preview || status === 'joining'}
          onPress={onJoin}
          accessibilityLabel={t('game.invite.join')}
        >
          {status === 'joining' ? t('game.invite.joining') : t('game.invite.join')}
        </GlossButton>
      </YStack>
    </PageShell>
  )
}
