import {
  REVEAL_SLOT_ORDINAL,
  type RevealEffectPlan,
  type RevealPlayer,
} from '@/features/game/lib/loadout-reveal'
import type { LoadoutSlotKind } from '@/features/game/lib/match-rules'
import {
  devilFruitEvolutionChain,
  devilFruitEvolutionSteps,
  standEvolutionChain,
  standEvolutionSteps,
} from '@/features/game/lib/stand-evolution'
import type { GameLoadout } from '@/features/game/types/game.types'

// A power effect: what the backend's resolver did while assigning a loadout -
// raised a stat up to a floor a power demands (STAT_FLOOR) or evolved the
// drawn Stand/DevilFruit into a later stage (EVOLUTION). The loadout's own
// stand/devilFruit and stats are the FINAL ones, after every effect; the
// sorteo replays the drawn values first and animates each effect at the
// moment its trigger is revealed.
export type LoadoutEffect = GameLoadout['effects'][number]
type EffectSlot = LoadoutEffect['slot']

// The wire's LoadoutSlot (what an effect changes / was caused by) -> the
// sorteo's own slot kind.
export const EFFECT_SLOT_KIND: Record<EffectSlot, LoadoutSlotKind> = {
  PHYSICAL_FORM: 'physicalForm',
  STAND: 'stand',
  DEVIL_FRUIT: 'devilFruit',
  FRUIT_MASTERY: 'fruitMastery',
  HAMON: 'hamon',
  ARMAMENT_HAKI: 'armamentHaki',
  OBSERVATION_HAKI: 'observationHaki',
  CONQUEROR_HAKI: 'conquerorHaki',
  SPIN: 'spin',
}

type ScalarField =
  | 'physicalForm'
  | 'fruitMastery'
  | 'hamon'
  | 'armamentHaki'
  | 'observationHaki'
  | 'conquerorHaki'
  | 'spin'

const SCALAR_FIELD: Partial<Record<EffectSlot, ScalarField>> = {
  PHYSICAL_FORM: 'physicalForm',
  FRUIT_MASTERY: 'fruitMastery',
  HAMON: 'hamon',
  ARMAMENT_HAKI: 'armamentHaki',
  OBSERVATION_HAKI: 'observationHaki',
  CONQUEROR_HAKI: 'conquerorHaki',
  SPIN: 'spin',
}

const HAKI_LEVEL_SLOTS: EffectSlot[] = ['ARMAMENT_HAKI', 'OBSERVATION_HAKI', 'CONQUEROR_HAKI']

// effectTargetSlot is the slot an effect's change is shown on in the sorteo's
// slot order: its own slot, except a haki type granted from nothing - that
// type has no level slot of its own (it was not in the drawn loadout), so the
// grant shows on the 'hakiSet' summary instead. Mirrors the backend's
// Loadout.hakiGranted.
export function effectTargetSlot(effect: LoadoutEffect): LoadoutSlotKind {
  if (
    effect.kind === 'STAT_FLOOR' &&
    HAKI_LEVEL_SLOTS.includes(effect.slot) &&
    effect.from === 'NONE'
  ) {
    return 'hakiSet'
  }
  return EFFECT_SLOT_KIND[effect.slot]
}

// effectAnchors says after which slot each effect plays: the latest, in
// reveal order, of the slot it changes, the slot that triggered it, and the
// anchor of the effect before it. The first two are "the moment its trigger
// has been revealed, and the card it changes exists"; the third keeps anchors
// non-decreasing along the list so the effects applied at any point are
// always a prefix of it (displayLoadoutAt relies on that).
export function effectAnchors(effects: LoadoutEffect[]): LoadoutSlotKind[] {
  const anchors: LoadoutSlotKind[] = []
  let latest = -1
  for (const effect of effects) {
    const candidates = [effectTargetSlot(effect), EFFECT_SLOT_KIND[effect.causeSlot]]
    let best = latest
    let anchor: LoadoutSlotKind | null = anchors.length > 0 ? anchors[anchors.length - 1] : null
    for (const c of candidates) {
      if (REVEAL_SLOT_ORDINAL[c] > best) {
        best = REVEAL_SLOT_ORDINAL[c]
        anchor = c
      }
    }
    latest = best
    anchors.push(anchor ?? candidates[0])
  }
  return anchors
}

type Chained = { id: string; evolvesFrom: Chained | null }

// findInChain returns the entry of power's evolution chain (power itself, or
// one of its ancestors) with the given id - an effect's `from`/`to` are
// always on the FINAL power's chain. Falls back to power itself rather than
// failing: it only decides which card an animation shows.
function findInChain<T extends Chained>(power: T | undefined, id: string): T | undefined {
  for (let cur: Chained | null | undefined = power; cur; cur = cur.evolvesFrom) {
    if (cur.id === id) return cur as T
  }
  return power
}

// displayLoadoutAt is the loadout as it looks after the first `applied`
// effects only: every later effect is undone, restoring the values it
// replaced. applied = 0 is the loadout as DRAWN (what the sorteo reveals
// slot by slot); applied >= effects.length is the loadout itself.
export function displayLoadoutAt(loadout: GameLoadout, applied: number): GameLoadout {
  const effects = loadout.effects ?? []
  if (applied >= effects.length) return loadout
  const out: GameLoadout = { ...loadout }
  for (let i = effects.length - 1; i >= Math.max(0, applied); i--) {
    const effect = effects[i]
    if (effect.kind === 'EVOLUTION') {
      if (effect.slot === 'STAND') out.stand = findInChain(loadout.stand, effect.from)
      else if (effect.slot === 'DEVIL_FRUIT') {
        out.devilFruit = findInChain(loadout.devilFruit, effect.from)
      }
      continue
    }
    const field = SCALAR_FIELD[effect.slot]
    if (field) (out as unknown as Record<string, string>)[field] = effect.from
  }
  return out
}

// effectSteps is how many evolution stages an effect spans along the final
// power's chain (to's depth minus from's), at least 1; 0 for a stat floor.
// Mirrors the backend's Loadout.effectSteps.
export function effectSteps(loadout: GameLoadout, effect: LoadoutEffect): number {
  if (effect.kind !== 'EVOLUTION') return 0
  const chain =
    effect.slot === 'STAND'
      ? loadout.stand
        ? standEvolutionChain(loadout.stand)
        : []
      : loadout.devilFruit
        ? devilFruitEvolutionChain(loadout.devilFruit)
        : []
  const from = chain.findIndex((p) => p.id === effect.from)
  const to = chain.findIndex((p) => p.id === effect.to)
  if (from < 0 || to < 0 || to - from < 1) return 1
  return to - from
}

// revealPlayerFor derives the sorteo's per-participant timeline input from a
// loadout, the one place it happens (the hook and the stage both use it, so
// they can never disagree). Mirrors the backend's game.RevealPlayerFor: it
// describes the DRAWN loadout - the evolution depths are the drawn stage's,
// and a haki type an effect granted from nothing has no level slot of its own.
export function revealPlayerFor(loadout: GameLoadout | null | undefined): RevealPlayer {
  if (!loadout) {
    return {
      hasStand: false,
      hasDevilFruit: false,
      hasArmamentHaki: false,
      hasObservationHaki: false,
      hasConquerorHaki: false,
      standEvolutionSteps: 0,
      fruitEvolutionSteps: 0,
      effects: [],
    }
  }
  const drawn = displayLoadoutAt(loadout, 0)
  const effects = loadout.effects ?? []
  const anchors = effectAnchors(effects)
  const plans: RevealEffectPlan[] = effects.map((effect, i) => ({
    kind: effect.kind,
    steps: effectSteps(loadout, effect),
    anchor: anchors[i],
  }))
  return {
    hasStand: !!loadout.stand,
    hasDevilFruit: !!loadout.devilFruit,
    // An absent field counts as "no such haki", like a NONE one.
    hasArmamentHaki: drawn.armamentHaki !== undefined && drawn.armamentHaki !== 'NONE',
    hasObservationHaki: drawn.observationHaki !== undefined && drawn.observationHaki !== 'NONE',
    hasConquerorHaki: drawn.conquerorHaki !== undefined && drawn.conquerorHaki !== 'NONE',
    standEvolutionSteps: standEvolutionSteps(drawn.stand),
    fruitEvolutionSteps: devilFruitEvolutionSteps(drawn.devilFruit),
    effects: plans,
  }
}
