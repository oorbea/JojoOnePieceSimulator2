import { useTranslation } from 'react-i18next'
import { YStack } from 'tamagui'

import { PlayerRow } from '@/features/game/components/presentational/player-row'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlowText } from '@/shared/components/presentational/glow-text'
import type { GameParticipant } from '@/features/game/types/game.types'

type Props = {
  participants: GameParticipant[]
  hostId: string
  selfId: string
  capacity: number
  isHost: boolean
  onKick: (participantId: string) => void
  onTransferHost: (participantId: string) => void
}

export function SquadRoster({ participants, hostId, selfId, capacity, isHost, onKick, onTransferHost }: Props) {
  const { t } = useTranslation()

  return (
    <GlassPanel tone="strong" width="100%" p="$4" gap="$3">
      <GlowText level="heading">
        {t('game.lobby.playersCount', { count: participants.length, max: capacity })}
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
    </GlassPanel>
  )
}
