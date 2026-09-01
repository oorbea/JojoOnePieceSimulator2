import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { XStack, YStack } from 'tamagui'

import { LoadoutModal } from '@/features/game/components/presentational/match/loadout-modal'
import { RosterParticipant } from '@/features/game/components/presentational/match/roster-participant'
import { teamTone, teamToneColor } from '@/features/game/lib/lobby-rules'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { useRovingGroup } from '@/shared/hooks/use-roving-group'
import type { GameParticipant, GameSnapshot } from '@/features/game/types/game.types'

type Props = {
  snapshot: GameSnapshot
  selfId: string
  /** Fired whenever the loadout modal opens or closes. This component keeps
   * owning that state (it's the one with `mangas` in hand for the modal's
   * content), but the container needs to know an overlay is up so it can
   * suppress the single-key match hotkeys - see use-match-hotkeys' `blocked`.
   * Optional: a caller that has no hotkeys to suppress just omits it. */
  onModalOpenChange?: (open: boolean) => void
}

// VERSUS: two tone-colored columns (reusing the lobby's team tone mapping).
// GAUNTLET: one wrapped row. The roster itself only shows avatar + username
// per participant (RosterParticipant/ParticipantTile) - the full loadout
// breakdown that used to render inline here (LoadoutCard) is now a 0.5s
// hover card or a tap-opened LoadoutModal, both driven by this component
// since it's the one that owns "which participant's modal is open" and has
// `mangas` in hand. Only rendered once the sorteo overlay (RevealStage) has
// finished, so every participant here already has their full loadout.
export function MatchRoster({ snapshot, selfId, onModalOpenChange }: Props) {
  const { t } = useTranslation()
  const mangas = snapshot.config.powerMangas
  const [modalParticipant, setModalParticipant] = useState<GameParticipant | null>(null)

  // Single funnel for every open/close, so the notification can never drift
  // out of sync with the state itself (both the roving-group activation and
  // the tile press go through here).
  const openModal = (participant: GameParticipant | null) => {
    setModalParticipant(participant)
    onModalOpenChange?.(participant !== null)
  }

  // Render order (teams flattened in VERSUS, snapshot order in GAUNTLET) is
  // the same order arrow-key roving moves through - one flat group across
  // the whole roster rather than a per-team one, simple and predictable
  // even though VERSUS renders it as two visual columns.
  const orderedParticipants =
    snapshot.mode === 'VERSUS'
      ? snapshot.teams.flatMap((team) => snapshot.participants.filter((p) => p.teamId === team.id))
      : snapshot.participants
  const { getItemProps } = useRovingGroup({
    groupId: 'roster-tile',
    count: orderedParticipants.length,
    onActivate: (index) => openModal(orderedParticipants[index]),
  })

  const renderTile = (p: GameParticipant) => {
    const index = orderedParticipants.findIndex((op) => op.id === p.id)
    return (
      <RosterParticipant
        key={p.id}
        participant={p}
        isSelf={p.id === selfId}
        mangas={mangas}
        onOpenModal={openModal}
        itemProps={getItemProps(index)}
      />
    )
  }

  const modal = (
    <LoadoutModal
      visible={modalParticipant !== null}
      participant={modalParticipant}
      isSelf={modalParticipant?.id === selfId}
      mangas={mangas}
      onClose={() => openModal(null)}
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
