import { Crown, Move, UserMinus, UserCog } from '@tamagui/lucide-icons-2'
import { useTranslation } from 'react-i18next'
import { Platform, View, type ViewStyle } from 'react-native'
import { XStack, YStack } from 'tamagui'

import { a11yProps } from '@/shared/lib/a11y'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { ParticipantAvatar } from '@/features/game/components/presentational/match/participant-avatar'
import { usePlayerDrag, type DragEndInfo } from '@/features/game/hooks/use-player-drag'
import type { GameParticipant } from '@/features/game/types/game.types'

type Props = {
  participant: GameParticipant
  isHost: boolean
  isSelf: boolean
  showHostActions: boolean
  onKick?: () => void
  /** Host-only removal of a bot; replaces Kick for bot rows. */
  onRemoveBot?: () => void
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

// Web-only: no text selection / native drag from the handle, and no browser
// scroll/zoom taking over a touch that starts on it.
const DRAG_HANDLE_WEB_STYLE: ViewStyle | undefined =
  Platform.OS === 'web'
    ? ({ cursor: 'grab', touchAction: 'none', userSelect: 'none' } as unknown as ViewStyle)
    : undefined

export function PlayerRow({
  participant,
  isHost,
  isSelf,
  showHostActions,
  onKick,
  onRemoveBot,
  onTransferHost,
  onDragEnd,
}: Props) {
  const { t } = useTranslation()
  const draggable = !!onDragEnd
  const { translate, handleProps } = usePlayerDrag(draggable, onDragEnd ?? (() => {}))

  return (
    <View
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
        <ParticipantAvatar participant={participant} size={36} isSelf={isSelf} />

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
                accessibilityLabel={t('game.transferHost.action', {
                  name: participant.displayName,
                })}
              >
                <UserCog size={16} color="$panelText" />
              </GlossButton>
            ) : null}
            {participant.kind === 'BOT' && onRemoveBot ? (
              <GlossButton
                tone="glass"
                btnSize="sm"
                shape="circle"
                onPress={onRemoveBot}
                accessibilityLabel={t('game.bots.remove', { name: participant.displayName })}
              >
                <UserMinus size={16} color="$strawHatRedDeep" />
              </GlossButton>
            ) : onKick ? (
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
          // The drag handle: the only part of the row that listens for the
          // gesture. Pressing on the row's text instead made the browser
          // start a native text selection/`dragstart` the moment the mouse
          // moved, which cancels the mouse events the PanResponder needs, and
          // on touch the page scroll claimed the gesture. The tap path
          // (TeamColumn's empty slot / host row actions) stays the primary,
          // accessible way to move a player; the handle isn't in the tab
          // order.
          <View
            {...handleProps}
            style={[{ padding: 11, margin: -11 }, DRAG_HANDLE_WEB_STYLE]}
          >
            <Move size={14} color="$panelTextSoft" />
          </View>
        ) : null}
      </XStack>
    </View>
  )
}
