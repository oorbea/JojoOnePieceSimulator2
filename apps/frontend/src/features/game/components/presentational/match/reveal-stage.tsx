import { useTranslation } from 'react-i18next'
import { XStack } from 'tamagui'

import { RevealLane } from '@/features/game/components/presentational/match/reveal-lane'
import { useRevealSpinSound } from '@/features/game/hooks/use-reveal-spin-sound'
import type { RevealPhaseKind } from '@/features/game/lib/loadout-reveal'
import { revealSlotKinds } from '@/features/game/lib/match-rules'
import type { GameSnapshot } from '@/features/game/types/game.types'
import { useDevilFruits } from '@/features/devil-fruits'
import { useStands } from '@/features/stands'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'

type Props = {
  snapshot: GameSnapshot
  selfId: string
  phase: RevealPhaseKind
  slotIndex: number
  totalSlots: number
  onSkip: () => void
  reducedMotion: boolean
}

// The sorteo overlay: every participant's carril spins for the SAME slot at
// the same time (poder a poder, not participant a participant), Wii
// Party-style. Fetches the Stand/DevilFruit catalogs itself (rather than
// having them threaded down from the container) purely for the roulette's
// decorative filler names - the actual answer per lane always comes from
// that participant's own loadout (see RevealLane). Also owns the reel-spin
// loop sound (see useRevealSpinSound) - this component is mounted for
// exactly the reveal's duration (match-screen.tsx only renders it while
// isRevealing), so its own mount/unmount lifecycle is the sound's start/stop
// signal.
export function RevealStage({
  snapshot,
  selfId,
  phase,
  slotIndex,
  totalSlots,
  onSkip,
  reducedMotion,
}: Props) {
  const { t } = useTranslation()
  useRevealSpinSound(!reducedMotion)
  const standsQuery = useStands()
  const devilFruitsQuery = useDevilFruits()
  const standNames = (standsQuery.data ?? []).map((s) => s.name)
  const fruitNames = (devilFruitsQuery.data ?? []).map((f) => f.name)

  const slotKinds = revealSlotKinds(snapshot.config.mangas)
  const currentKind = slotIndex >= 0 && slotIndex < slotKinds.length ? slotKinds[slotIndex] : null
  const spinning = phase === 'spin'

  const orderedParticipants =
    snapshot.mode === 'VERSUS'
      ? snapshot.teams.flatMap((team) => snapshot.participants.filter((p) => p.teamId === team.id))
      : snapshot.participants

  const title =
    phase === 'intro' || phase === 'outro' || !currentKind
      ? t('game.match.reveal.title')
      : t(`game.match.trait.${currentKind}`)

  return (
    <GlassPanel tone="strong" width="100%" p="$4" gap="$3" items="center">
      <GlowText level="heading">
        {phase === 'intro' ? t('game.match.reveal.intro') : title}
      </GlowText>
      {currentKind && phase !== 'outro' ? (
        <GlowText level="label" tone="soft">
          {t('game.match.reveal.progress', { current: slotIndex + 1, total: totalSlots })}
        </GlowText>
      ) : null}

      <XStack flexWrap="wrap" gap="$2.5" justify="center" width="100%">
        {orderedParticipants.map((p, i) => (
          <RevealLane
            key={p.id}
            participant={p}
            isSelf={p.id === selfId}
            slotKind={currentKind}
            spinning={spinning}
            standNames={standNames}
            fruitNames={fruitNames}
            reducedMotion={reducedMotion}
            laneIndex={i}
          />
        ))}
      </XStack>

      <GlossButton
        tone="glass"
        btnSize="sm"
        onPress={onSkip}
        accessibilityLabel={t('game.match.reveal.skipA11y')}
        tooltip={t('game.match.reveal.skipA11y')}
      >
        {t('game.match.reveal.skip')}
      </GlossButton>
    </GlassPanel>
  )
}
