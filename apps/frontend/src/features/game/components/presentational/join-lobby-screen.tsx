import { ChevronLeft, Globe, Lock, Users } from '@tamagui/lucide-icons-2'
import { useTranslation } from 'react-i18next'
import { ActivityIndicator } from 'react-native'
import { XStack, YStack } from 'tamagui'

import { CODE_LENGTH, isCompleteCode } from '@/features/game/lib/game-code'
import { GlassField } from '@/shared/components/presentational/glass-field'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { PageShell } from '@/shared/components/presentational/page-shell'
import type { LobbyPreview } from '@/features/game/types/game.types'

type Props = {
  onBack: () => void
  code: string
  onChangeCode: (code: string) => void
  preview?: LobbyPreview
  previewLoading: boolean
  previewError?: string
  joining: boolean
  onSubmit: () => void
}

export function JoinLobbyScreen({
  onBack,
  code,
  onChangeCode,
  preview,
  previewLoading,
  previewError,
  joining,
  onSubmit,
}: Props) {
  const { t } = useTranslation()

  return (
    <PageShell align="center" maxWidth={480}>
      <XStack width="100%" items="center" gap="$3">
        <GlossButton tone="glass" btnSize="sm" shape="circle" onPress={onBack} accessibilityLabel={t('common.cancel')}>
          <ChevronLeft size={18} color="$panelText" />
        </GlossButton>
        <GlowText level="title">{t('game.join.title')}</GlowText>
      </XStack>

      <GlassField
        label={t('game.join.codeLabel')}
        value={code}
        onChangeText={onChangeCode}
        placeholder={t('game.join.codePlaceholder')}
        autoCapitalize="characters"
        maxLength={CODE_LENGTH}
        error={previewError}
      />

      {previewLoading ? <ActivityIndicator /> : null}

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
          <GlowText level="label">{t('game.preview.host', { name: preview.hostDisplayName })}</GlowText>
          {preview.locked ? <GlowText level="label">{t('game.preview.lockedBadge')}</GlowText> : null}
        </GlassPanel>
      ) : null}

      <YStack width="100%">
        <GlossButton
          tone="blue"
          btnSize="lg"
          width="100%"
          disabled={!isCompleteCode(code) || !preview || joining}
          onPress={onSubmit}
          accessibilityLabel={t('game.join.submit')}
        >
          {joining ? t('game.join.joining') : t('game.join.submit')}
        </GlossButton>
      </YStack>
    </PageShell>
  )
}
