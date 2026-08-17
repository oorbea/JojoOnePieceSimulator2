import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'

import { PowerRoulette } from '@/features/game/components/presentational/match/power-roulette'
import { REVEAL_SPIN_MS } from '@/features/game/lib/loadout-reveal'
import type { LoadoutSlotKind } from '@/features/game/lib/match-rules'
import type { GameLoadout, GameParticipant } from '@/features/game/types/game.types'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlowText } from '@/shared/components/presentational/glow-text'

type Props = {
  participant: GameParticipant
  isSelf: boolean
  /** The slot kind currently spinning/landing, shared by every lane - null
   * during 'intro'/'outro', when nothing has a value to show yet. */
  slotKind: LoadoutSlotKind | null
  spinning: boolean
  /** Decorative filler for the Stand/DevilFruit reels - real catalog names,
   * never the actual answer (that always comes from participant.loadout). */
  standNames: string[]
  fruitNames: string[]
  reducedMotion: boolean
  /** This lane's position among its siblings - drives PowerRoulette's
   * landing stagger only, never which value is drawn. */
  laneIndex: number
}

const SCALAR_NAMESPACE: Record<string, string> = {
  spin: 'spinLevel',
  hamon: 'hamonLevel',
  fruitMastery: 'fruitMastery',
  physicalForm: 'physicalForm',
  armamentHaki: 'hakiLevel',
  observationHaki: 'hakiLevel',
  conquerorHaki: 'hakiLevel',
}

const SCALAR_VALUES: Record<string, string[]> = {
  spin: ['NONE', 'BASIC', 'ADVANCED', 'GOLDEN', 'INFINITE'],
  hamon: ['NONE', 'BASIC', 'ADVANCED', 'PERFECT'],
  fruitMastery: ['NONE', 'REGULAR', 'ADVANCED', 'AWAKENED'],
  physicalForm: ['PRIVATE', 'VICE_ADMIRAL', 'YONKO_COMMANDER', 'YONKO_PLUS'],
  armamentHaki: ['PRIVATE', 'VICE_ADMIRAL', 'YONKO_COMMANDER', 'YONKO_PLUS'],
  observationHaki: ['PRIVATE', 'VICE_ADMIRAL', 'YONKO_COMMANDER', 'YONKO_PLUS'],
  conquerorHaki: ['PRIVATE', 'VICE_ADMIRAL', 'YONKO_COMMANDER', 'YONKO_PLUS'],
}

// One participant's carril: name plus a PowerRoulette for whichever slot the
// whole lobby is currently on. The roulette's candidates are always
// decorative (real catalog names for Stand/DevilFruit, every enum member for
// a scalar slot) - the value it lands on always comes straight from this
// participant's own loadout, never from the candidate pool.
export function RevealLane({
  participant,
  isSelf,
  slotKind,
  spinning,
  standNames,
  fruitNames,
  reducedMotion,
  laneIndex,
}: Props) {
  const { t } = useTranslation()
  const loadout = participant.loadout

  // Memoised so PowerRoulette's reel useMemo doesn't invalidate every render
  // (slotFor otherwise returns a fresh candidates array each time) - matters
  // more now that the reel is mid-animation (stagger/overshoot) far more
  // often than the old instant-cut version was.
  const { candidates, finalLabel } = useMemo(
    () => slotFor(t, loadout, slotKind, standNames, fruitNames),
    [t, loadout, slotKind, standNames, fruitNames]
  )

  return (
    <GlassPanel
      tone={isSelf ? 'strong' : 'plastic'}
      px="$2.5"
      py="$2"
      rounded="$card"
      gap="$1"
      minW={130}
      borderColor={isSelf ? ('$wiiBlue' as never) : undefined}
      borderWidth={isSelf ? 2 : undefined}
    >
      <GlowText level="label" fontSize="$1" numberOfLines={1}>
        {participant.displayName}
      </GlowText>
      <PowerRoulette
        candidates={candidates}
        finalLabel={finalLabel}
        spinning={spinning}
        reducedMotion={reducedMotion}
        spinMs={REVEAL_SPIN_MS}
        laneIndex={laneIndex}
      />
    </GlassPanel>
  )
}

function slotFor(
  t: (key: string, opts?: Record<string, unknown>) => string,
  loadout: GameLoadout | undefined,
  slotKind: LoadoutSlotKind | null,
  standNames: string[],
  fruitNames: string[]
): { candidates: string[]; finalLabel: string } {
  if (!slotKind || !loadout) return { candidates: [], finalLabel: '' }

  if (slotKind === 'stand') {
    return { candidates: standNames, finalLabel: loadout.stand?.name ?? t('game.match.noStand') }
  }
  if (slotKind === 'devilFruit') {
    return {
      candidates: fruitNames,
      finalLabel: loadout.devilFruit?.name ?? t('game.match.noFruit'),
    }
  }

  const namespace = SCALAR_NAMESPACE[slotKind]
  const values = SCALAR_VALUES[slotKind] ?? []
  const rawValue = (loadout as unknown as Record<string, string>)[slotKind]
  return {
    candidates: values.map((v) => t(`enums.${namespace}.${v}`)),
    finalLabel: t(`enums.${namespace}.${rawValue}`),
  }
}
