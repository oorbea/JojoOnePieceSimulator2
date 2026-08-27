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
  spin: ['NONE', 'BASIC', 'GOLDEN', 'INFINITE'],
  hamon: ['NONE', 'BASIC', 'ADVANCED', 'PERFECT'],
  fruitMastery: ['NONE', 'REGULAR', 'ADVANCED', 'AWAKENED'],
  physicalForm: ['PRIVATE', 'STRONG_FISHMAN', 'MARINE_CAPTAIN', 'VICE_ADMIRAL', 'YONKO_COMMANDER', 'YONKO_PLUS'],
  armamentHaki: ['NONE', 'PRIVATE', 'VICE_ADMIRAL', 'YONKO_COMMANDER', 'YONKO_PLUS'],
  observationHaki: ['NONE', 'PRIVATE', 'VICE_ADMIRAL', 'YONKO_COMMANDER', 'YONKO_PLUS'],
  conquerorHaki: ['NONE', 'PRIVATE', 'VICE_ADMIRAL', 'YONKO_COMMANDER', 'YONKO_PLUS'],
}

// HAKI_TYPES pairs each haki level field with the i18n key naming that type
// (not its mastery level) - used only by the 'hakiSet' slot below to render
// "which types do you have" before the individual level slots run.
const HAKI_TYPES: { field: 'armamentHaki' | 'observationHaki' | 'conquerorHaki'; i18nKey: string }[] = [
  { field: 'armamentHaki', i18nKey: 'game.match.hakiType.armament' },
  { field: 'observationHaki', i18nKey: 'game.match.hakiType.observation' },
  { field: 'conquerorHaki', i18nKey: 'game.match.hakiType.conqueror' },
]

// Every combination of the 3 haki types, as a localized joined label - the
// roulette's decorative candidate pool for the 'hakiSet' slot. Mirrors the
// backend's game.HakiSet (weights.go): 8 combinations including "none".
function hakiSetCombos(t: (key: string) => string): string[] {
  const labels = HAKI_TYPES.map((h) => t(h.i18nKey))
  const combos: string[] = []
  for (let mask = 0; mask < 8; mask++) {
    const present = labels.filter((_, i) => (mask & (1 << i)) !== 0)
    combos.push(present.length === 0 ? t('game.match.hakiType.none') : present.join(', '))
  }
  return combos
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
  if (slotKind === 'hakiSet') {
    const present = HAKI_TYPES.filter((h) => (loadout as unknown as Record<string, string>)[h.field] !== 'NONE')
    const finalLabel = present.length === 0 ? t('game.match.hakiType.none') : present.map((h) => t(h.i18nKey)).join(', ')
    return { candidates: hakiSetCombos(t), finalLabel }
  }

  const namespace = SCALAR_NAMESPACE[slotKind]
  const values = SCALAR_VALUES[slotKind] ?? []
  const rawValue = (loadout as unknown as Record<string, string>)[slotKind]
  return {
    candidates: values.map((v) => t(`enums.${namespace}.${v}`)),
    finalLabel: t(`enums.${namespace}.${rawValue}`),
  }
}
