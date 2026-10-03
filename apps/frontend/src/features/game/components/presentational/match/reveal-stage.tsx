import { useTranslation } from 'react-i18next'
import { XStack, YStack } from 'tamagui'

import { CaseStripReel } from '@/features/game/components/presentational/match/case-strip-reel'
import { ParticipantAvatar } from '@/features/game/components/presentational/match/participant-avatar'
import {
  PowerRevealCard,
  type EvolvePhase,
} from '@/features/game/components/presentational/match/power-reveal-card'
import { EffectLevelUp } from '@/features/game/components/presentational/match/effect-level-up'
import { PowerRoulette } from '@/features/game/components/presentational/match/power-roulette'
import { RevealNarrator } from '@/features/game/components/presentational/match/reveal-narrator'
import { StageAnnouncement } from '@/features/game/components/presentational/match/stage-announcement'
import { useRevealSpinSound } from '@/features/game/hooks/use-reveal-spin-sound'
import { buildCaseStrip, type CaseStripCard } from '@/features/game/lib/case-strip'
import {
  playerSlots,
  REVEAL_SLOT_ORDINAL,
  REVEAL_SPEED_MULTIPLIER,
  revealSlotSeed,
  spinMsFor,
  type RevealPhaseKind,
} from '@/features/game/lib/loadout-reveal'
import type { LoadoutSlotKind } from '@/features/game/lib/match-rules'
import { applyPoolFilter } from '@/features/game/lib/power-pool'
import {
  EFFECT_SLOT_KIND,
  displayLoadoutAt,
  revealPlayerFor,
  type LoadoutEffect,
} from '@/features/game/lib/power-effects'
import {
  devilFruitEvolutionChain,
  devilFruitEvolutionSteps,
  standEvolutionChain,
  standEvolutionSteps,
} from '@/features/game/lib/stand-evolution'
import type { GameSnapshot } from '@/features/game/types/game.types'
import { useDevilFruits } from '@/features/devil-fruits'
import { useStands } from '@/features/stands'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { formatBattleIQ } from '@/shared/lib/battle-iq'
import { thumbSource } from '@/shared/lib/picture-source'

// The "landed nothing" card - a Stand/DevilFruit slot that rolled NONE
// still gets its own strip (owner decision, 2026-09-25): it reads as a
// deliberately unremarkable grey card among the real candidates, never a
// blank gap in the strip.
const NONE_POWER_CARD_ID = '__none__'

// One representative score per WAIS-IV band, purely cosmetic decoys for the
// battleIQ roulette to spin through before landing on the real score -
// mirrors SCALAR_VALUES' role for the enum-backed scalar slots above.
const BATTLE_IQ_CANDIDATE_SCORES = [55, 75, 85, 100, 115, 125, 150]

type Props = {
  snapshot: GameSnapshot
  selfId: string
  phase: RevealPhaseKind
  /** Index into snapshot.participants (join order) - whose turn is
   * currently playing. -1 during the lobby-wide 'intro'/'outro'. */
  participantIndex: number
  slotIndex: number
  totalSlots: number
  /** Only meaningful during an 'evolveStep'/'effectStep' phase - see
   * useLoadoutReveal's own doc. -1 otherwise. */
  evolveStage: number
  /** Which of the current participant's power effects is playing (only
   * meaningful during the four 'effect*' phases, -1 otherwise) and how many
   * count as already applied - see useLoadoutReveal's identical fields. */
  effectIndex: number
  effectsApplied: number
  /** How much every phase's local duration is being stretched/squeezed to
   * fit the server's actual reveal window - the roulette's own spinMs must
   * scale by the same factor, or its spin drifts out of step with the rest
   * of the timeline (see useLoadoutReveal's identical field). */
  scale: number
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
  physicalForm: [
    'PRIVATE',
    'STRONG_FISHMAN',
    'MARINE_CAPTAIN',
    'VICE_ADMIRAL',
    'YONKO_COMMANDER',
    'YONKO_PLUS',
  ],
  // No 'NONE' here - a haki level slot only ever appears (see playerSlots)
  // for a type the participant actually has, so it can never land on NONE
  // and the roulette shouldn't tease it as a possible outcome either.
  armamentHaki: ['PRIVATE', 'VICE_ADMIRAL', 'YONKO_COMMANDER', 'YONKO_PLUS'],
  observationHaki: ['PRIVATE', 'VICE_ADMIRAL', 'YONKO_COMMANDER', 'YONKO_PLUS'],
  conquerorHaki: ['PRIVATE', 'VICE_ADMIRAL', 'YONKO_COMMANDER', 'YONKO_PLUS'],
}

const HAKI_TYPES: {
  field: 'armamentHaki' | 'observationHaki' | 'conquerorHaki'
  i18nKey: string
}[] = [
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
  evolveStage,
  effectIndex,
  effectsApplied,
  scale,
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
  const finalLoadout = currentParticipant?.loadout
  // The loadout as it looks right now: every power effect that has not played
  // yet is undone, so the sorteo reveals the DRAWN values slot by slot and each
  // effect changes them at the moment its trigger shows up (power-effects.ts's
  // displayLoadoutAt).
  const loadout = finalLoadout ? displayLoadoutAt(finalLoadout, effectsApplied) : undefined
  // playerSlots(mangas, thisPlayer) - not the lobby-wide revealSlotKinds -
  // since which haki-level slots exist varies per participant (owner
  // request, 2026-08-30: only the haki types they actually have get a
  // roulette at all). slotIndex indexes into THIS list. revealPlayerFor is
  // the same derivation the hook builds the timeline from.
  const currentPlayer = revealPlayerFor(finalLoadout)
  const slotKinds = currentParticipant
    ? playerSlots(snapshot.config.powerMangas, currentPlayer)
    : []
  const isEffectPhase =
    phase === 'effectIntro' ||
    phase === 'effectEvolving' ||
    phase === 'effectStep' ||
    phase === 'effectLand'
  const effect: LoadoutEffect | undefined = isEffectPhase
    ? finalLoadout?.effects?.[effectIndex]
    : undefined
  // During a power effect's beat the card on screen is the one the effect
  // changes (not the anchor slot it plays after).
  const currentSlot: LoadoutSlotKind | null = effect
    ? EFFECT_SLOT_KIND[effect.slot]
    : slotIndex >= 0 && slotIndex < slotKinds.length
      ? slotKinds[slotIndex]
      : null
  const spinning = phase === 'spin'
  const landed = phase === 'land'
  // The full ancestor chains, root-first (see stand-evolution.ts). A
  // non-evolving power is a single-element chain (itself), so every downstream
  // lookup below works unchanged whether or not this loadout's Stand/Fruit
  // actually evolves from anything. A power effect's evolution walks the FINAL
  // power's chain (its from/to are always on it); the slot's own evolution
  // walks the DRAWN power's chain, up to the stage drawn.
  const evolvesStandEffect = effect?.kind === 'EVOLUTION' && effect.slot === 'STAND'
  const evolvesFruitEffect = effect?.kind === 'EVOLUTION' && effect.slot === 'DEVIL_FRUIT'
  const standSource = evolvesStandEffect ? finalLoadout?.stand : loadout?.stand
  const fruitSource = evolvesFruitEffect ? finalLoadout?.devilFruit : loadout?.devilFruit
  const standChain = standSource ? standEvolutionChain(standSource) : []
  const fruitChain = fruitSource ? devilFruitEvolutionChain(fruitSource) : []
  const standSteps = loadout?.stand ? standEvolutionSteps(loadout.stand) : 0
  const fruitSteps = loadout?.devilFruit ? devilFruitEvolutionSteps(loadout.devilFruit) : 0
  const isEvolvePhase =
    phase === 'evolveBase' ||
    phase === 'evolving' ||
    phase === 'evolveStep' ||
    (effect?.kind === 'EVOLUTION' && isEffectPhase)
  const evolveChain =
    currentSlot === 'stand' ? standChain : currentSlot === 'devilFruit' ? fruitChain : []
  // Which stage of the chain the current phase shows: the root during
  // evolveBase/evolving, chain[evolveStage + 1] for each intermediate
  // evolveStep, and the stage drawn once landed - see reveal.go's evolveMs /
  // loadout-reveal.ts's revealTimeline doc. A power effect's beats do the same
  // between the effect's `from` and `to` stages.
  const effectFrom = effect ? evolveChain.findIndex((p) => p.id === effect.from) : -1
  const effectTo = effect ? evolveChain.findIndex((p) => p.id === effect.to) : -1
  const lastStage = evolveChain.length - 1
  const evolveIndex: number | undefined =
    evolveChain.length === 0
      ? undefined
      : evolvesStandEffect || evolvesFruitEffect
        ? phase === 'effectIntro' || phase === 'effectEvolving'
          ? Math.max(0, effectFrom)
          : phase === 'effectStep'
            ? Math.min(Math.max(0, effectFrom) + 1 + evolveStage, effectTo >= 0 ? effectTo : lastStage)
            : effectTo >= 0
              ? effectTo
              : lastStage
        : phase === 'evolveBase' || phase === 'evolving'
          ? 0
          : phase === 'evolveStep'
            ? Math.min(evolveStage + 1, lastStage)
            : lastStage
  const evolveDisplayStand =
    currentSlot === 'stand' && evolveIndex !== undefined ? standChain[evolveIndex] : undefined
  const evolveDisplayFruit =
    currentSlot === 'devilFruit' && evolveIndex !== undefined ? fruitChain[evolveIndex] : undefined
  const slotEvolves =
    (currentSlot === 'stand' && standSteps > 0) || (currentSlot === 'devilFruit' && fruitSteps > 0)
  const evolvePhase: EvolvePhase | undefined =
    evolvesStandEffect || evolvesFruitEffect
      ? phase === 'effectIntro'
        ? 'base'
        : phase === 'effectEvolving'
          ? 'evolving'
          : phase === 'effectStep'
            ? 'step'
            : phase === 'effectLand'
              ? 'final'
              : undefined
      : slotEvolves
        ? phase === 'evolveBase'
          ? 'base'
          : phase === 'evolving'
            ? 'evolving'
            : phase === 'evolveStep'
              ? 'step'
              : phase === 'land'
                ? 'final'
                : undefined
        : undefined
  const effectLine =
    effect && finalLoadout ? effectLineFor(t, effect, finalLoadout) : ''

  const { candidates, finalLabel } = slotFor(t, loadout, currentSlot, standNames, fruitNames)
  const isPowerSlot = currentSlot === 'stand' || currentSlot === 'devilFruit'

  const speed = snapshot.config.revealSpeed
  const speedMultiplier = REVEAL_SPEED_MULTIPLIER[speed] ?? REVEAL_SPEED_MULTIPLIER.NORMAL
  // Scaled by the SAME factor useLoadoutReveal is stretching/squeezing every
  // other phase's duration by, or the spin finishes out of step with the
  // narrator/land beats around it the moment the two ever disagree (a slow
  // device, a mid-reveal reconnect that seeked into this phase, ...).
  const spinMs =
    currentParticipant && currentSlot
      ? spinMsFor(snapshot.id, snapshot.rounds.length, participantIndex, currentSlot) *
        speedMultiplier *
        scale
      : 0
  const slotSeed =
    currentParticipant && currentSlot
      ? revealSlotSeed(
          snapshot.id,
          snapshot.rounds.length,
          participantIndex,
          REVEAL_SLOT_ORDINAL[currentSlot]
        )
      : 0

  // A plain computed value, not memoized: it's a cheap, pure build of a
  // ~50-card array from data already in hand, and re-deriving it every
  // render is simpler (and lint-cleaner under the React Compiler) than
  // keeping a dependency list in sync with it.
  const caseStrip = (() => {
    if (!isPowerSlot || !loadout || !currentSlot) return null
    const { stands, fruits } = applyPoolFilter(
      standsQuery.data ?? [],
      devilFruitsQuery.data ?? [],
      snapshot.config.poolFilter
    )
    const pool: CaseStripCard[] =
      currentSlot === 'stand'
        ? stands.map((s) => ({
            id: s.id,
            label: s.name,
            rarity: s.rarity,
            picture: thumbSource(s) ?? undefined,
          }))
        : fruits.map((f) => ({
            id: f.id,
            label: f.name,
            rarity: f.rarity,
            picture: thumbSource(f) ?? undefined,
          }))
    // The strip lands on the ROOT base of an evolving stand's chain, never
    // the final landed form directly - the "algo esta pasando" evolution
    // beats (evolveBase/evolving/evolveStep) take over from there to reveal
    // the actual result (2026-09-25, see standChain's doc above).
    const standWinnerSource = standChain[0] ?? loadout.stand
    // Same for a fruit that evolves (Model Nika drawn directly lands on Gomu
    // Gomu no mi, then evolves).
    const fruitWinnerSource = fruitChain[0] ?? loadout.devilFruit
    const winner: CaseStripCard =
      currentSlot === 'stand'
        ? standWinnerSource
          ? {
              id: standWinnerSource.id,
              label: standWinnerSource.name,
              rarity: standWinnerSource.rarity,
              picture: thumbSource(standWinnerSource) ?? undefined,
            }
          : { id: NONE_POWER_CARD_ID, label: t('game.match.noStand'), rarity: 'NONE' }
        : fruitWinnerSource
          ? {
              id: fruitWinnerSource.id,
              label: fruitWinnerSource.name,
              rarity: fruitWinnerSource.rarity,
              picture: thumbSource(fruitWinnerSource) ?? undefined,
            }
          : { id: NONE_POWER_CARD_ID, label: t('game.match.noFruit'), rarity: 'NONE' }
    return buildCaseStrip(pool, winner, slotSeed)
  })()

  // A power effect that changes a stat shows the stat (EffectLevelUp), not a
  // card: only a Stand/DevilFruit evolution takes the card over.
  const showPowerCard =
    (landed || isEvolvePhase) && currentParticipant !== null && isPowerSlot

  const narratorLine = isEffectPhase
    ? effectLine
    : narratorLineFor(t, phase, currentParticipant?.displayName, currentSlot, loadout, finalLabel)

  const title =
    phase === 'intro' || phase === 'outro'
      ? t('game.match.reveal.title')
      : currentSlot
        ? t(`game.match.trait.${currentSlot}`)
        : t('game.match.reveal.title')

  return (
    <GlassPanel tone="strong" width="100%" p="$4" gap="$3" items="center">
      {snapshot.mode === 'VERSUS' && snapshot.upcomingStage ? (
        <StageAnnouncement
          stage={snapshot.upcomingStage}
          compact={phase !== 'intro' && phase !== 'outro'}
        />
      ) : null}
      <GlowText level="heading">{title}</GlowText>
      {currentSlot && phase !== 'outro' && phase !== 'intro' && !isEffectPhase ? (
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
            <ParticipantAvatar
              participant={currentParticipant}
              size={36}
              isSelf={currentParticipant.id === selfId}
            />
            <GlowText level="heading" numberOfLines={1}>
              {currentParticipant.displayName}
            </GlowText>
          </XStack>
          {currentSlot && !showPowerCard && !isEffectPhase && isPowerSlot && caseStrip ? (
            <CaseStripReel
              cards={caseStrip.cards}
              landingIndex={caseStrip.landingIndex}
              landingOffset={caseStrip.landingOffset}
              spinning={spinning}
              landed={landed}
              reducedMotion={reducedMotion}
              spinMs={spinMs}
            />
          ) : effect && effect.kind === 'STAT_FLOOR' && currentSlot ? (
            <EffectLevelUp
              statLabel={t(`game.match.trait.${currentSlot}`)}
              fromLabel={t(`enums.${SCALAR_NAMESPACE[currentSlot]}.${effect.from}`)}
              toLabel={t(`enums.${SCALAR_NAMESPACE[currentSlot]}.${effect.to}`)}
              raised={phase === 'effectLand'}
              stamp={t('game.match.reveal.effect.raised')}
              reducedMotion={reducedMotion}
            />
          ) : currentSlot && !showPowerCard && !isPowerSlot && !isEffectPhase ? (
            <PowerRoulette
              candidates={candidates}
              finalLabel={finalLabel}
              spinning={spinning}
              landed={landed}
              reducedMotion={reducedMotion}
              spinMs={spinMs}
              seed={slotSeed}
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
        {readyTotal
          ? t('game.match.reveal.readyCount', {
              ready: readyCount ?? 0,
              needed: Math.floor(readyTotal / 2) + 1,
            })
          : t('game.match.reveal.skip')}
      </GlossButton>

      {currentParticipant ? (
        <PowerRevealCard
          visible={showPowerCard}
          kind={currentSlot === 'stand' ? 'stand' : 'devilFruit'}
          stand={currentSlot === 'stand' ? (evolveDisplayStand ?? loadout?.stand) : undefined}
          devilFruit={
            currentSlot === 'devilFruit' ? (evolveDisplayFruit ?? loadout?.devilFruit) : undefined
          }
          participantName={currentParticipant.displayName}
          onSkip={onSkip}
          evolvePhase={evolvePhase}
          causeLine={isEffectPhase ? effectLine : undefined}
          reducedMotion={reducedMotion}
        />
      ) : null}
    </GlassPanel>
  )
}

// effectLineFor says what triggered a power effect, in the participant's own
// words: "Con Spin Infinito, Tusk: Acto 1 evoluciona a Tusk: Acto 4" / "King
// Crimson te hace subir Haki de Observación a Comandante de Yonko". Power
// names come off the final power's chain, stat values through the same enums
// namespaces the roulette already uses.
function effectLineFor(
  t: (key: string, opts?: Record<string, unknown>) => string,
  effect: LoadoutEffect,
  loadout: NonNullable<GameSnapshot['participants'][number]['loadout']>
): string {
  const causeKind = EFFECT_SLOT_KIND[effect.causeSlot]
  const causeIsPower = causeKind === 'stand' || causeKind === 'devilFruit'
  const causeValue = causeIsPower
    ? effect.cause
    : t(`enums.${SCALAR_NAMESPACE[causeKind]}.${effect.cause}`)

  if (effect.kind === 'EVOLUTION') {
    const chain =
      effect.slot === 'STAND'
        ? loadout.stand
          ? standEvolutionChain(loadout.stand)
          : []
        : loadout.devilFruit
          ? devilFruitEvolutionChain(loadout.devilFruit)
          : []
    return t(
      effect.slot === 'STAND'
        ? 'game.match.reveal.effect.evolution.stand'
        : 'game.match.reveal.effect.evolution.fruit',
      {
        cause: causeValue,
        from: chain.find((p) => p.id === effect.from)?.name ?? '',
        to: chain.find((p) => p.id === effect.to)?.name ?? '',
      }
    )
  }

  const statKind = EFFECT_SLOT_KIND[effect.slot]
  const stat = t(`game.match.trait.${statKind}`)
  const value = t(`enums.${SCALAR_NAMESPACE[statKind]}.${effect.to}`)
  if (causeIsPower) {
    return t('game.match.reveal.effect.floor.byPower', { cause: causeValue, stat, value })
  }
  return t('game.match.reveal.effect.floor.byStat', {
    causeStat: t(`game.match.trait.${causeKind}`),
    cause: causeValue,
    stat,
    value,
  })
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
    return t(`game.match.reveal.narrator.${narratorKey(slot)}.before`, {
      type: hakiTypeLabel(t, slot),
    })
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
    if (
      (slot === 'spin' && loadout?.spin === 'NONE') ||
      (slot === 'hamon' && loadout?.hamon === 'NONE')
    ) {
      return t(`game.match.reveal.narrator.${slot}.none`, { name })
    }
    if (slot === 'hakiSet') {
      const hasAnyHaki =
        loadout?.armamentHaki !== 'NONE' ||
        loadout?.observationHaki !== 'NONE' ||
        loadout?.conquerorHaki !== 'NONE'
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
    return {
      candidates: fruitNames,
      finalLabel: loadout.devilFruit?.name ?? t('game.match.noFruit'),
    }
  }
  if (slotKind === 'hakiSet') {
    const present = HAKI_TYPES.filter(
      (h) => (loadout as unknown as Record<string, string>)[h.field] !== 'NONE'
    )
    const finalLabel =
      present.length === 0
        ? t('game.match.hakiType.none')
        : present.map((h) => t(h.i18nKey)).join(', ')
    return { candidates: hakiSetCombos(t), finalLabel }
  }

  if (slotKind === 'battleIQ') {
    const score = loadout.battleIQ
    return {
      candidates: BATTLE_IQ_CANDIDATE_SCORES.map((s) => formatBattleIQ(t, s) ?? ''),
      finalLabel: formatBattleIQ(t, score) ?? '',
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
