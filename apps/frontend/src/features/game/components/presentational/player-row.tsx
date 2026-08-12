import { Bot, Crown, UserMinus, UserCog } from '@tamagui/lucide-icons-2'
import { useTranslation } from 'react-i18next'
import { Paragraph, XStack, YStack } from 'tamagui'

import { a11yProps } from '@/shared/lib/a11y'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import type { GameParticipant } from '@/features/game/types/game.types'

type Props = {
  participant: GameParticipant
  isHost: boolean
  isSelf: boolean
  showHostActions: boolean
  onKick?: () => void
  onTransferHost?: () => void
}

export function PlayerRow({ participant, isHost, isSelf, showHostActions, onKick, onTransferHost }: Props) {
  const { t } = useTranslation()

  return (
    <XStack
      items="center"
      gap="$3"
      py="$2"
      px="$3"
      rounded="$card"
      bg="$plasticFill"
      {...a11yProps(t('game.a11y.playerRow', { name: participant.displayName }))}
    >
      <YStack
        width={36}
        height={36}
        rounded="$circle"
        bg={participant.kind === 'BOT' ? '$plasticEdge' : '$wiiBlue'}
        items="center"
        justify="center"
      >
        {participant.kind === 'BOT' ? (
          <Bot size={18} color="$panelTextSoft" />
        ) : (
          <Paragraph color="white" fontWeight="800">
            {participant.displayName.charAt(0).toUpperCase()}
          </Paragraph>
        )}
      </YStack>

      <YStack flex={1} gap="$0.5">
        <XStack items="center" gap="$1.5">
          <GlowText level="label" color="$panelText">
            {participant.displayName}
          </GlowText>
          {isHost ? (
            <Crown size={14} color="$sunYellowDeep" {...a11yProps(t('game.lobby.host'))} />
          ) : null}
          {isSelf ? (
            <GlassPanel tone="plastic" px="$2" py="$0.5" rounded="$pill" elevate={0}>
              <GlowText level="label">{t('game.lobby.you')}</GlowText>
            </GlassPanel>
          ) : null}
        </XStack>
        <XStack items="center" gap="$1.5">
          <YStack
            width={8}
            height={8}
            rounded="$circle"
            bg={participant.connected ? '$meadowGreen' : '$plasticEdge'}
            {...a11yProps(
              t(participant.connected ? 'game.lobby.connected' : 'game.lobby.disconnected')
            )}
          />
          <GlowText level="label">
            {t(participant.connected ? 'game.lobby.connected' : 'game.lobby.disconnected')}
          </GlowText>
        </XStack>
      </YStack>

      {showHostActions ? (
        <XStack gap="$1.5">
          {participant.kind === 'HUMAN' && onTransferHost ? (
            <GlossButton
              tone="glass"
              btnSize="sm"
              shape="circle"
              onPress={onTransferHost}
              accessibilityLabel={t('game.transferHost.action', { name: participant.displayName })}
            >
              <UserCog size={16} color="$panelText" />
            </GlossButton>
          ) : null}
          {onKick ? (
            <GlossButton
              tone="glass"
              btnSize="sm"
              shape="circle"
              onPress={onKick}
              accessibilityLabel={t('game.kick.action', { name: participant.displayName })}
            >
              <UserMinus size={16} color="$strawHatRedDeep" />
            </GlossButton>
          ) : null}
        </XStack>
      ) : null}
    </XStack>
  )
}
