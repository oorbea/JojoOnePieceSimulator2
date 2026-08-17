import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { XStack, YStack } from 'tamagui'

import { LoadoutModal } from '@/features/game/components/presentational/match/loadout-modal'
import { RosterParticipant } from '@/features/game/components/presentational/match/roster-participant'
import { teamTone, teamToneColor } from '@/features/game/lib/lobby-rules'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlowText } from '@/shared/components/presentational/glow-text'
import type { GameParticipant, GameSnapshot } from '@/features/game/types/game.types'

type Props = {
  snapshot: GameSnapshot
  selfId: string
}

// VERSUS: two tone-colored columns (reusing the lobby's team tone mapping).
// GAUNTLET: one wrapped row. The roster itself only shows avatar + username
// per participant (RosterParticipant/ParticipantTile) - the full loadout
// breakdown that used to render inline here (LoadoutCard) is now a 0.5s
// hover card or a tap-opened LoadoutModal, both driven by this component
// since it's the one that owns "which participant's modal is open" and has
// `mangas` in hand. Only rendered once the sorteo overlay (RevealStage) has
// finished, so every participant here already has their full loadout.
export function MatchRoster({ snapshot, selfId }: Props) {
  const { t } = useTranslation()
  const mangas = snapshot.config.mangas
  const [modalParticipant, setModalParticipant] = useState<GameParticipant | null>(null)

  const renderTile = (p: GameParticipant) => (
    <RosterParticipant
      key={p.id}
      participant={p}
      isSelf={p.id === selfId}
      mangas={mangas}
      onOpenModal={setModalParticipant}
    />
  )

  const modal = (
    <LoadoutModal
      visible={modalParticipant !== null}
      participant={modalParticipant}
      isSelf={modalParticipant?.id === selfId}
      mangas={mangas}
      onClose={() => setModalParticipant(null)}
    />
  )

  if (snapshot.mode === 'VERSUS') {
    return (
      <>
        <YStack width="100%" gap="$3" $md={{ flexDirection: 'row' }}>
          {snapshot.teams.map((team, index) => {
            const tone = teamTone(index)
            const participants = snapshot.participants.filter((p) => p.teamId === team.id)
            return (
              <GlassPanel
                key={team.id}
                tone="strong"
                flex={1}
                p="$3"
                gap="$2.5"
                borderColor={teamToneColor(tone) as never}
              >
                <GlowText level="heading" color={teamToneColor(tone) as never}>
                  {team.name}
                </GlowText>
                <XStack flexWrap="wrap" gap="$2.5">
                  {participants.map(renderTile)}
                </XStack>
              </GlassPanel>
            )
          })}
        </YStack>
        {modal}
      </>
    )
  }

  return (
    <>
      <GlassPanel tone="strong" width="100%" p="$3" gap="$2.5">
        <GlowText level="heading">{t('game.match.title')}</GlowText>
        <XStack flexWrap="wrap" gap="$2.5">
          {snapshot.participants.map(renderTile)}
        </XStack>
      </GlassPanel>
      {modal}
    </>
  )
}
