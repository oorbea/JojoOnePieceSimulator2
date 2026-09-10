import type {
  JojoCharacterResponse,
  OnePieceCharacterResponse,
  TaggedCharacter,
} from '@/features/characters/types/characters.types'

// One row of a character's stat block: an i18n key for the label plus how
// to read/format that stat off the response. `enumNamespace` is undefined
// for battleIq (a raw 0-255 number, formatted through shared/lib/battle-iq
// instead of an enums.* lookup - see loadout-modal.tsx's own
// battleIq/enum split for the precedent this mirrors).
export type CharacterStatRow<T> = {
  key: string
  labelKey: string
  enumNamespace?: string
  value: (character: T) => string | number
}

// character-stat-block.tsx is the only leaf that imports both of these -
// everything else (characters-screen.tsx, character-card.tsx,
// character-detail.tsx) is kind-agnostic and just forwards a descriptor.
export const JOJO_STAT_ROWS: CharacterStatRow<JojoCharacterResponse>[] = [
  {
    key: 'hamon',
    labelKey: 'characters.stats.hamon',
    enumNamespace: 'hamonLevel',
    value: (c) => c.hamon,
  },
  {
    key: 'spin',
    labelKey: 'characters.stats.spin',
    enumNamespace: 'spinLevel',
    value: (c) => c.spin,
  },
  { key: 'battleIq', labelKey: 'characters.stats.battleIq', value: (c) => c.battleIq },
]

export const ONE_PIECE_STAT_ROWS: CharacterStatRow<OnePieceCharacterResponse>[] = [
  {
    key: 'physicalForm',
    labelKey: 'characters.stats.physicalForm',
    enumNamespace: 'physicalForm',
    value: (c) => c.physicalForm,
  },
  {
    key: 'armamentHaki',
    labelKey: 'characters.stats.armamentHaki',
    enumNamespace: 'hakiLevel',
    value: (c) => c.armamentHaki,
  },
  {
    key: 'observationHaki',
    labelKey: 'characters.stats.observationHaki',
    enumNamespace: 'hakiLevel',
    value: (c) => c.observationHaki,
  },
  {
    key: 'conquerorHaki',
    labelKey: 'characters.stats.conquerorHaki',
    enumNamespace: 'hakiLevel',
    value: (c) => c.conquerorHaki,
  },
  {
    key: 'fruitMastery',
    labelKey: 'characters.stats.fruitMastery',
    enumNamespace: 'fruitMastery',
    value: (c) => c.fruitMastery,
  },
]

// Picks the right descriptor for a `.kind`-tagged item - the only place
// that needs to know both descriptors at once when the manga filter is
// 'ALL' and a single grid mixes both kinds. Cast through `unknown`: each
// row's `value` closes over its own concrete response type, which is sound
// per-branch (a JOJO-tagged item really is a JojoCharacterResponse) but not
// expressible as a single `CharacterStatRow<TaggedCharacter>[]` without it.
export function rowsForCharacter(character: TaggedCharacter): CharacterStatRow<TaggedCharacter>[] {
  return (character.kind === 'JOJO'
    ? JOJO_STAT_ROWS
    : ONE_PIECE_STAT_ROWS) as unknown as CharacterStatRow<TaggedCharacter>[]
}
