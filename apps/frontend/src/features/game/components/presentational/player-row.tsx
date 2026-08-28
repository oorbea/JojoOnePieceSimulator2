import { Bot, Crown, Move, UserMinus, UserCog } from '@tamagui/lucide-icons-2'
import { useTranslation } from 'react-i18next'
import { View } from 'react-native'
import { Paragraph, XStack, YStack } from 'tamagui'

import { a11yProps } from '@/shared/lib/a11y'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { usePlayerDrag, type DragEndInfo } from '@/features/game/hooks/use-player-drag'
import type { GameParticipant } from '@/features/game/types/game.types'

type Props = {
  participant: GameParticipant
  isHost: boolean
  isSelf: boolean
  showHostActions: boolean
  onKick?: () => void
  onTransferHost?: () => void
  /** Drag-to-move onto another `TeamColumn` - required to work on both
   * desktop (mouse drag) and mobile (touch drag), not optional polish (see
   * game-lobby-todo.md §5). Omit to render the row non-draggable (e.g. a
   * viewer with neither self nor host permission to move this participant -
   * the tap-based `onJoin`/host-action paths stay the sole primary way to
   * move a player regardless, this is an additional interaction layered on
   * top, not a replacement). */
  onDragEnd?: (info: DragEndInfo) => void
}

export function PlayerRow({
  participant,
  isHost,
  isSelf,
  showHostActions,
  onKick,
  onTransferHost,
  onDragEnd,
}: Props) {
  const { t } = useTranslation()
  const draggable = !!onDragEnd
  const { translate, panHandlers } = usePlayerDrag(draggable, onDragEnd ?? (() => {}))

  return (
    <View
      {...panHandlers}
      style={
        draggable
          ? { transform: [{ translateX: translate.x }, { translateY: translate.y }], zIndex: 1 }
          : undefined
      }
    >
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

      {draggable ? (
        // Purely a visual affordance ("this row can be dragged") - the tap
        // path (TeamColumn's empty slot / host row actions) stays the
        // primary, accessible way to move a player; this icon carries no
        // interaction of its own and isn't in the tab order.
        <Move size={14} color="$panelTextSoft" />
      ) : null}
      </XStack>
    </View>
  )
}
