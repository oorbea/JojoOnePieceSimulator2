import { useTranslation } from 'react-i18next'
import { View } from 'react-native'
import { YStack } from 'tamagui'

import { PlayerRow } from '@/features/game/components/presentational/player-row'
import type { DragEndInfo } from '@/features/game/hooks/use-player-drag'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { a11yProps } from '@/shared/lib/a11y'
import { teamToneColor, type TeamTone } from '@/features/game/lib/lobby-rules'
import type { GameParticipant, GameTeam } from '@/features/game/types/game.types'

type Props = {
  team: GameTeam
  tone: TeamTone
  participants: GameParticipant[]
  hostId: string
  selfId: string
  capacity: number
  canJoin: boolean
  isHost: boolean
  onJoin: () => void
  onKick: (participantId: string) => void
  onTransferHost: (participantId: string) => void
  /** Drop-zone registration for drag-to-move (`use-drop-zones.ts`, owned by
   * the parent that renders every `TeamColumn` since resolving a drop needs
   * to see all of them, not just this one). Optional so this component
   * still works standalone (e.g. Storybook/tests) without a live registry. */
  zoneRef?: (node: View | null) => void
  onZoneLayout?: () => void
  /** Called with a participant's release point; the parent resolves it
   * against every registered zone and issues switchTeam/movePlayer. */
  onDragEndAt?: (participantId: string, info: DragEndInfo) => void
}

export function TeamColumn({
  team,
  tone,
  participants,
  hostId,
  selfId,
  capacity,
  canJoin,
  isHost,
  onJoin,
  onKick,
  onTransferHost,
  zoneRef,
  onZoneLayout,
  onDragEndAt,
}: Props) {
  const { t } = useTranslation()
  const full = participants.length >= capacity

  return (
    <View
      ref={zoneRef}
      collapsable={false}
      onLayout={onZoneLayout}
      // `flex: 1` (not a fixed `width: '100%'`) so this column shares the
      // row equally with its sibling once the parent switches to
      // `flexDirection: 'row'` at $md+ (lobby-room-screen.tsx) - a fixed
      // 100% width here made each column claim the *entire* row width,
      // overflowing the page's 1080px cap and pushing Team B off-screen
      // with no way to reach it (join/kick/drag all became unusable on any
      // desktop-width viewport). `minWidth: 0` overrides flexbox's default
      // content-based minimum, which otherwise refuses to shrink a column
      // below its content's natural width and reproduces the same overflow.
      // Still fills the full width when stacked (single column,
      // flexDirection: 'column') via RN's default `alignItems: 'stretch'`.
      style={{ flex: 1, minWidth: 0 }}
    >
      <GlassPanel tone="strong" width="100%" p="$4" gap="$3" borderColor={teamToneColor(tone) as never}>
        <GlowText level="heading" color={teamToneColor(tone) as never}>
          {team.name} · {participants.length}/{capacity}
        </GlowText>

        <YStack gap="$2">
          {participants.map((p) => {
            // Drag is additive on top of the tap paths below - self can
            // drag to switch team (mirrors tapping an empty slot), host can
            // drag anyone (mirrors a host-only action, net-new: there was
            // no equivalent tap action for moving *another* player before).
            const draggable = p.id === selfId || isHost
            return (
              <PlayerRow
                key={p.id}
                participant={p}
                isHost={p.id === hostId}
                isSelf={p.id === selfId}
                showHostActions={isHost && p.id !== selfId}
                onKick={() => onKick(p.id)}
                onTransferHost={() => onTransferHost(p.id)}
                onDragEnd={draggable && onDragEndAt ? (info) => onDragEndAt(p.id, info) : undefined}
              />
            )
          })}
        </YStack>

        {canJoin && !full ? (
          <YStack
            borderWidth={1.5}
            borderColor="$glassEdge"
            borderStyle="dashed"
            rounded="$card"
            p="$3"
            items="center"
            onPress={onJoin}
            cursor="pointer"
            {...a11yProps(t('game.lobby.switchTeam', { name: team.name }), 'button')}
          >
            <GlowText level="label">{t('game.lobby.emptySlot')}</GlowText>
          </YStack>
        ) : null}

        {full ? (
          <GlossButton tone="glass" btnSize="sm" disabled>
            {t('game.lobby.teamFull')}
          </GlossButton>
        ) : null}
      </GlassPanel>
    </View>
  )
}
