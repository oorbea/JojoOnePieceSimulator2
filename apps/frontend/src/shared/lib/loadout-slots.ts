import type { LoadoutSlot } from '@/shared/contracts/enums'

// Where each wire LoadoutSlot (what a PowerEffect or a manual rule names) finds
// its labels: `trait` is the stat's name under `game.match.trait.*`, `ns` the
// `enums.<ns>.*` namespace its level values translate through (null for the
// slots that hold a power, not a level), and `lowest` the first level that
// counts as "having it at all" (a Haki at PRIVATE, a Spin at BASIC...).
// Record<LoadoutSlot, ...> makes a new backend slot a compile error here.
const SLOT_INFO: Record<LoadoutSlot, { trait: string; ns: string | null; lowest: string | null }> =
  {
    PHYSICAL_FORM: { trait: 'physicalForm', ns: 'physicalForm', lowest: 'PRIVATE' },
    STAND: { trait: 'stand', ns: null, lowest: null },
    DEVIL_FRUIT: { trait: 'devilFruit', ns: null, lowest: null },
    FRUIT_MASTERY: { trait: 'fruitMastery', ns: 'fruitMastery', lowest: 'REGULAR' },
    HAMON: { trait: 'hamon', ns: 'hamonLevel', lowest: 'BASIC' },
    ARMAMENT_HAKI: { trait: 'armamentHaki', ns: 'hakiLevel', lowest: 'PRIVATE' },
    OBSERVATION_HAKI: { trait: 'observationHaki', ns: 'hakiLevel', lowest: 'PRIVATE' },
    CONQUEROR_HAKI: { trait: 'conquerorHaki', ns: 'hakiLevel', lowest: 'PRIVATE' },
    SPIN: { trait: 'spin', ns: 'spinLevel', lowest: 'BASIC' },
  }

export function slotTraitKey(slot: LoadoutSlot): string {
  return `game.match.trait.${SLOT_INFO[slot].trait}`
}

export function slotLevelKey(slot: LoadoutSlot, level: string): string {
  const ns = SLOT_INFO[slot].ns
  return ns ? `enums.${ns}.${level}` : level
}

// True when `level` is the lowest level that still counts as having the stat,
// so "Spin Basic or higher" can be said as "any Spin".
export function isLowestLevel(slot: LoadoutSlot, level: string): boolean {
  return SLOT_INFO[slot].lowest === level
}
