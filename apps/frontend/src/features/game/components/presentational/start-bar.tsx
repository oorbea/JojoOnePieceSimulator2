import { useTranslation } from 'react-i18next'
import { YStack } from 'tamagui'

import type { Gate } from '@/features/game/lib/lobby-rules'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'

type Props = {
  isHost: boolean
  gate: Gate
  starting: boolean
  onStart: () => void
  onLeave: () => void
  onAbort: () => void
}

export function StartBar({ isHost, gate, starting, onStart, onLeave, onAbort }: Props) {
  const { t } = useTranslation()

  return (
    <GlassPanel tone="strong" width="100%" p="$4" gap="$3" elevate={2}>
      {isHost ? (
        <YStack gap="$2" items="center">
          <GlossButton
            tone="green"
            btnSize="lg"
            flare
            disabled={!gate.ok || starting}
            onPress={onStart}
            accessibilityLabel={t('game.start.start')}
          >
            {starting ? t('game.start.starting') : t('game.start.start')}
          </GlossButton>
          {!gate.ok ? (
            <GlowText level="label" align="center">
              {t(gate.reasonKey, gate.params)}
            </GlowText>
          ) : null}
          <GlossButton tone="red" btnSize="sm" onPress={onAbort} accessibilityLabel={t('game.abort.action')}>
            {t('game.abort.action')}
          </GlossButton>
        </YStack>
      ) : (
        <YStack gap="$2" items="center">
          <GlowText level="label" align="center">
            {t('game.lobby.waitingForHost')}
          </GlowText>
          <GlossButton tone="glass" btnSize="sm" onPress={onLeave} accessibilityLabel={t('game.leave.action')}>
            {t('game.leave.action')}
          </GlossButton>
        </YStack>
      )}
    </GlassPanel>
  )
}
