import { useTranslation } from 'react-i18next'
import { XStack, YStack } from 'tamagui'

import { ParticipantAvatar } from '@/features/game/components/presentational/match/participant-avatar'
import { PowerRevealCard } from '@/features/game/components/presentational/match/power-reveal-card'
import { PowerRoulette } from '@/features/game/components/presentational/match/power-roulette'
import { RevealNarrator } from '@/features/game/components/presentational/match/reveal-narrator'
import { useRevealSpinSound } from '@/features/game/hooks/use-reveal-spin-sound'
import {
  playerSlots,
  REVEAL_SLOT_ORDINAL,
  REVEAL_SPEED_MULTIPLIER,
  REVEAL_SPIN_BASE_MS,
  revealSpinCycles,
  type RevealPhaseKind,
  type RevealPlayer,
} from '@/features/game/lib/loadout-reveal'
import type { LoadoutSlotKind } from '@/features/game/lib/match-rules'
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
  /** Index into snapshot.participants (join order) - whose turn is
   * currently playing. -1 during the lobby-wide 'intro'/'outro'. */
  participantIndex: number
  slotIndex: number
  totalSlots: number
  /** REVEAL_READY_CHANGED's own aggregate - how many of how many connected
   * humans have already marked themselves ready to skip. null before the
   * first frame for this ASSIGNING window arrives. */
  readyCount: number | null
  readyTotal: number | null
  onSkip: () => void
  reducedMotion: boolean
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
  // No 'NONE' here - a haki level slot only ever appears (see playerSlots)
  // for a type the participant actually has, so it can never land on NONE
  // and the roulette shouldn't tease it as a possible outcome either.
  armamentHaki: ['PRIVATE', 'VICE_ADMIRAL', 'YONKO_COMMANDER', 'YONKO_PLUS'],
  observationHaki: ['PRIVATE', 'VICE_ADMIRAL', 'YONKO_COMMANDER', 'YONKO_PLUS'],
  conquerorHaki: ['PRIVATE', 'VICE_ADMIRAL', 'YONKO_COMMANDER', 'YONKO_PLUS'],
}

const HAKI_TYPES: { field: 'armamentHaki' | 'observationHaki' | 'conquerorHaki'; i18nKey: string }[] = [
  { field: 'armamentHaki', i18nKey: 'game.match.hakiType.armament' },
  { field: 'observationHaki', i18nKey: 'game.match.hakiType.observation' },
  { field: 'conquerorHaki', i18nKey: 'game.match.hakiType.conqueror' },
]

function hakiSetCombos(t: (key: string) => string): string[] {
  const labels = HAKI_TYPES.map((h) => t(h.i18nKey))
  const combos: string[] = []
  for (let mask = 0; mask < 8; mask++) {
    const present = labels.filter((_, i) => (mask & (1 << i)) !== 0)
    combos.push(present.length === 0 ? t('game.match.hakiType.none') : present.join(', '))
  }
  return combos
}

// The sorteo overlay: jugador-por-jugador (owner request, 2026-08-30 - see
// ObsidianVault/game-match-assignment-frontend.md for the all-lanes-
// simultaneously design this supersedes), V1-style. One participant's turn
// at a time gets a big roulette plus a narrator line reproducing V1's own
// before/after copy (game.match.reveal.narrator.*); a Stand/Devil Fruit
// landing opens PowerRevealCard full-screen so the art/description/skills
// actually get read (the gap the pre-2026-08-30 reveal never closed). The
// rest of the lobby shows as a strip of avatars (done/current/pending).
export function RevealStage({
  snapshot,
  selfId,
  phase,
  participantIndex,
  slotIndex,
  totalSlots,
  readyCount,
  readyTotal,
  onSkip,
  reducedMotion,
}: Props) {
  const { t } = useTranslation()
  useRevealSpinSound(phase, !reducedMotion)
  const standsQuery = useStands()
  const devilFruitsQuery = useDevilFruits()
  const standNames = (standsQuery.data ?? []).map((s) => s.name)
  const fruitNames = (devilFruitsQuery.data ?? []).map((f) => f.name)

  const participants = snapshot.participants
  const currentParticipant = participantIndex >= 0 ? participants[participantIndex] : null
  const loadout = currentParticipant?.loadout
  // playerSlots(mangas, thisPlayer) - not the lobby-wide revealSlotKinds -
  // since which haki-level slots exist varies per participant (owner
  // request, 2026-08-30: only the haki types they actually have get a
  // roulette at all). slotIndex indexes into THIS list.
  const currentPlayer: RevealPlayer = {
    hasStand: !!loadout?.stand,
    hasDevilFruit: !!loadout?.devilFruit,
    hasArmamentHaki: loadout?.armamentHaki !== undefined && loadout.armamentHaki !== 'NONE',
    hasObservationHaki: loadout?.observationHaki !== undefined && loadout.observationHaki !== 'NONE',
    hasConquerorHaki: loadout?.conquerorHaki !== undefined && loadout.conquerorHaki !== 'NONE',
  }
  const slotKinds = currentParticipant ? playerSlots(snapshot.config.powerMangas, currentPlayer) : []
  const currentSlot: LoadoutSlotKind | null =
    slotIndex >= 0 && slotIndex < slotKinds.length ? slotKinds[slotIndex] : null
  const spinning = phase === 'spin'
  const landed = phase === 'land'

  const { candidates, finalLabel } = slotFor(t, loadout, currentSlot, standNames, fruitNames)

  const speed = snapshot.config.revealSpeed
  const speedMultiplier = REVEAL_SPEED_MULTIPLIER[speed] ?? REVEAL_SPEED_MULTIPLIER.NORMAL
  const cycles =
    currentParticipant && currentSlot
      ? revealSpinCycles(snapshot.id, snapshot.rounds.length, participantIndex, REVEAL_SLOT_ORDINAL[currentSlot])
      : 1
  const spinMs = REVEAL_SPIN_BASE_MS * cycles * speedMultiplier

  const showPowerCard =
    landed && currentParticipant !== null && (currentSlot === 'stand' || currentSlot === 'devilFruit')

  const narratorLine = narratorLineFor(
    t,
    phase,
    currentParticipant?.displayName,
    currentSlot,
    loadout,
    finalLabel
  )

  const title =
    phase === 'intro' || phase === 'outro'
      ? t('game.match.reveal.title')
      : currentSlot
        ? t(`game.match.trait.${currentSlot}`)
        : t('game.match.reveal.title')

  return (
    <GlassPanel tone="strong" width="100%" p="$4" gap="$3" items="center">
      <GlowText level="heading">{title}</GlowText>
      {currentSlot && phase !== 'outro' && phase !== 'intro' ? (
        <GlowText level="label" tone="soft">
          {t('game.match.reveal.progress', { current: slotIndex + 1, total: totalSlots })}
        </GlowText>
      ) : null}

      <RevealNarrator line={narratorLine} reducedMotion={reducedMotion} />

      {currentParticipant ? (
        <GlassPanel
          tone="strong"
          px="$3"
          py="$2.5"
          rounded="$card"
          gap="$2"
          items="center"
          minW={220}
        >
          <XStack items="center" gap="$2">
            <ParticipantAvatar participant={currentParticipant} size={36} isSelf={currentParticipant.id === selfId} />
            <GlowText level="heading" numberOfLines={1}>
              {currentParticipant.displayName}
            </GlowText>
          </XStack>
          {currentSlot && !showPowerCard ? (
            <PowerRoulette
              candidates={candidates}
              finalLabel={finalLabel}
              spinning={spinning}
              reducedMotion={reducedMotion}
              spinMs={spinMs}
            />
          ) : null}
        </GlassPanel>
      ) : null}

      <XStack flexWrap="wrap" gap="$2" justify="center" width="100%">
        {participants.map((p, i) => (
          <YStack
            key={p.id}
            opacity={i === participantIndex ? 1 : i < participantIndex ? 0.55 : 0.35}
          >
            <ParticipantAvatar participant={p} size={28} isSelf={p.id === selfId} />
          </YStack>
        ))}
      </XStack>

      <GlossButton
        tone="glass"
        btnSize="sm"
        onPress={onSkip}
        accessibilityLabel={t('game.match.reveal.skipA11y')}
        tooltip={t('game.match.reveal.skipA11y')}
      >
        {readyTotal ? t('game.match.reveal.readyCount', { ready: readyCount ?? 0, total: readyTotal }) : t('game.match.reveal.skip')}
      </GlossButton>

      {currentParticipant ? (
        <PowerRevealCard
          visible={showPowerCard}
          kind={currentSlot === 'stand' ? 'stand' : 'devilFruit'}
          stand={loadout?.stand}
          devilFruit={loadout?.devilFruit}
          participantName={currentParticipant.displayName}
          onSkip={onSkip}
        />
      ) : null}
    </GlassPanel>
  )
}

function narratorLineFor(
  t: (key: string, opts?: Record<string, unknown>) => string,
  phase: RevealPhaseKind,
  name: string | undefined,
  slot: LoadoutSlotKind | null,
  loadout: GameSnapshot['participants'][number]['loadout'],
  finalLabel: string
): string {
  if (phase === 'outro') return t('game.match.reveal.outro')
  if (!name) return ''
  if (phase === 'playerIntro') return t('game.match.reveal.narrator.playerTurn', { name })
  if (!slot) return ''
  if (phase === 'narrator' || phase === 'spin') {
    return t(`game.match.reveal.narrator.${narratorKey(slot)}.before`, { type: hakiTypeLabel(t, slot) })
  }
  if (phase === 'land') {
    if (slot === 'devilFruit') {
      return loadout?.devilFruit
        ? t('game.match.reveal.narrator.devilFruit.after', {
            name,
            name2: loadout.devilFruit.name,
            type: t(`enums.fruitType.${loadout.devilFruit.fruitType}`),
          })
        : t('game.match.reveal.narrator.devilFruit.none', { name })
    }
    if (slot === 'stand') {
      return loadout?.stand
        ? t('game.match.reveal.narrator.stand.after', { name, name2: loadout.stand.name })
        : t('game.match.reveal.narrator.stand.none', { name })
    }
    // spin/hamon have an explicit "never learned" line, matching V1's own
    // wording - every other scalar slot always has a value (its floor is
    // never "absent", e.g. physicalForm's weakest tier is still a form).
    if ((slot === 'spin' && loadout?.spin === 'NONE') || (slot === 'hamon' && loadout?.hamon === 'NONE')) {
      return t(`game.match.reveal.narrator.${slot}.none`, { name })
    }
    if (slot === 'hakiSet') {
      const hasAnyHaki =
        loadout?.armamentHaki !== 'NONE' || loadout?.observationHaki !== 'NONE' || loadout?.conquerorHaki !== 'NONE'
      return hasAnyHaki
        ? t('game.match.reveal.narrator.haki.after', { list: finalLabel })
        : t('game.match.reveal.narrator.haki.none', { name })
    }
    return t(`game.match.reveal.narrator.${narratorKey(slot)}.after`, {
      name,
      value: finalLabel,
      type: hakiTypeLabel(t, slot),
    })
  }
  return ''
}

// narratorKey maps a slot to its narrator i18n group - 'hakiSet' summarizes
// WHICH haki types a participant has (V1's own first haki beat); the three
// individual level slots share a separate 'hakiMastery' group for HOW MUCH
// of each, matching the owner's "which before how much" ordering.
function narratorKey(slot: LoadoutSlotKind): string {
  switch (slot) {
    case 'hakiSet':
      return 'haki'
    case 'armamentHaki':
    case 'observationHaki':
    case 'conquerorHaki':
      return 'hakiMastery'
    default:
      return slot
  }
}

// hakiTypeLabel names WHICH haki type a hakiMastery beat is about (owner
// request, 2026-08-30: each individual haki-level slot must say clearly
// which type it is, not just "your mastery is X") - unused (empty string)
// for every other slot, where the narrator text names nothing haki-related.
function hakiTypeLabel(t: (key: string) => string, slot: LoadoutSlotKind): string {
  switch (slot) {
    case 'armamentHaki':
    case 'observationHaki':
    case 'conquerorHaki':
      return t(`game.match.trait.${slot}`)
    default:
      return ''
  }
}

function slotFor(
  t: (key: string, opts?: Record<string, unknown>) => string,
  loadout: GameSnapshot['participants'][number]['loadout'],
  slotKind: LoadoutSlotKind | null,
  standNames: string[],
  fruitNames: string[]
): { candidates: string[]; finalLabel: string } {
  if (!slotKind || !loadout) return { candidates: [], finalLabel: '' }

  if (slotKind === 'stand') {
    return { candidates: standNames, finalLabel: loadout.stand?.name ?? t('game.match.noStand') }
  }
  if (slotKind === 'devilFruit') {
    return { candidates: fruitNames, finalLabel: loadout.devilFruit?.name ?? t('game.match.noFruit') }
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
