import { useTranslation } from 'react-i18next'
import { YStack } from 'tamagui'

import { PlayerRow } from '@/features/game/components/presentational/player-row'
import type { ChannelTileTone } from '@/shared/components/presentational/channel-tile'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { a11yProps } from '@/shared/lib/a11y'
import type { GameParticipant, GameTeam } from '@/features/game/types/game.types'

const TONE_BG: Record<ChannelTileTone, string> = {
  blue: '$wiiBlue',
  red: '$strawHatRed',
  green: '$meadowGreen',
  grape: '$grapeSoda',
  yellow: '$sunYellow',
  pink: '$bubblegum',
}

type Props = {
  team: GameTeam
  tone: ChannelTileTone
  participants: GameParticipant[]
  hostId: string
  selfId: string
  capacity: number
  canJoin: boolean
  isHost: boolean
  onJoin: () => void
  onKick: (participantId: string) => void
  onTransferHost: (participantId: string) => void
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
}: Props) {
  const { t } = useTranslation()
  const full = participants.length >= capacity

  return (
    <GlassPanel tone="strong" width="100%" p="$4" gap="$3" borderColor={TONE_BG[tone] as never}>
      <GlowText level="heading" color={TONE_BG[tone] as never}>
        {team.name} · {participants.length}/{capacity}
      </GlowText>

      <YStack gap="$2">
        {participants.map((p) => (
          <PlayerRow
            key={p.id}
            participant={p}
            isHost={p.id === hostId}
            isSelf={p.id === selfId}
            showHostActions={isHost && p.id !== selfId}
            onKick={() => onKick(p.id)}
            onTransferHost={() => onTransferHost(p.id)}
          />
        ))}
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
  )
}
