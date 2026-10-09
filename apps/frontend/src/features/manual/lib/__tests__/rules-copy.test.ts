import { MANUAL_RULES } from '@/shared/contracts/rules'
import caES from '@/shared/i18n/locales/ca-ES.json'
import enGB from '@/shared/i18n/locales/en-GB.json'
import esES from '@/shared/i18n/locales/es-ES.json'
import { slotLevelKey, slotTraitKey } from '@/shared/lib/loadout-slots'
import type { LoadoutSlot } from '@/shared/contracts/enums'

import { isAbsolute } from '../convention-copy'
import type { ManualChain, ManualConvention, ManualSubject } from '../../types/manual'

// The manual's rules are generated from the backend; its words are not. This is
// what ties the two together: every convention, category, group, tie rule and
// level the generated rules mention must have copy in every language, so a
// rule added in Go without its translation fails here (and in CI) instead of
// reaching a player as a raw "manual.conventions.items.X.title" key.

function has(catalog: unknown, path: string): boolean {
  let node: unknown = catalog
  for (const part of path.split('.')) {
    if (node === null || typeof node !== 'object' || !(part in node)) return false
    node = (node as Record<string, unknown>)[part]
  }
  return typeof node === 'string'
}

const conventions: readonly ManualConvention[] = MANUAL_RULES.conventions
const chains: readonly ManualChain[] = MANUAL_RULES.evolutions
const floorSubjects: readonly ManualSubject[] = MANUAL_RULES.statFloors.map((f) => f.when)
const allSubjects: ManualSubject[] = conventions.flatMap((c) => [
  ...c.actors,
  ...c.targets,
  ...c.immune,
  ...c.attenuated,
])

function requiredKeys(): string[] {
  const keys = new Set<string>()
  for (const c of conventions) {
    keys.add(`manual.conventions.categories.${c.category}`)
    if (!isAbsolute(c)) {
      keys.add(`manual.conventions.items.${c.id}.title`)
      keys.add(`manual.conventions.items.${c.id}.text`)
    }
  }
  for (const s of [...allSubjects, ...floorSubjects]) {
    if (s.group) keys.add(`manual.groups.${s.group}`)
    for (const ft of s.fruitTypes ?? []) keys.add(`enums.fruitType.${ft}`)
    if (s.kind === 'SLOT_MIN') {
      keys.add(slotTraitKey(s.slot as LoadoutSlot))
      keys.add(slotLevelKey(s.slot as LoadoutSlot, s.min ?? ''))
    }
  }
  keys.add(`manual.effects.tie.${MANUAL_RULES.tieRule}`)
  for (const chain of chains) {
    const driver = chain.driver as LoadoutSlot
    keys.add(slotTraitKey(driver))
    for (const level of chain.levels) keys.add(slotLevelKey(driver, level))
    for (const stage of chain.stages) {
      if (stage.tier) keys.add(slotLevelKey(driver, stage.tier))
    }
    for (const row of chain.rows) {
      for (const o of row.outcomes) {
        keys.add(`manual.effects.outcome.${o.change}`)
        keys.add(slotLevelKey(driver, o.levelAfter))
      }
    }
  }
  for (const f of MANUAL_RULES.statFloors) {
    keys.add(slotTraitKey(f.target as LoadoutSlot))
    keys.add(slotLevelKey(f.target as LoadoutSlot, f.floor))
  }
  for (const band of MANUAL_RULES.battleIQ) keys.add(`enums.battleIQCategory.${band.key}`)
  for (const level of [
    ...MANUAL_RULES.odds.spin,
    ...MANUAL_RULES.odds.hamon,
    ...MANUAL_RULES.odds.fruitMastery,
    ...MANUAL_RULES.odds.physicalForm,
    ...MANUAL_RULES.odds.hakiMastery,
  ]) {
    expect(level.level).toBeTruthy()
  }
  return [...keys]
}

describe.each([
  ['en-GB', enGB],
  ['es-ES', esES],
  ['ca-ES', caES],
])('manual copy in %s covers every generated rule', (_locale, catalog) => {
  it('has a string for each key the generated rules need', () => {
    const missing = requiredKeys().filter((key) => !has(catalog, key))
    expect(missing).toEqual([])
  })

  it('has the fixed manual copy (sections, labels, subjects, etiquette)', () => {
    const fixed = [
      'manual.title',
      'manual.intro',
      'manual.index.jump',
      ...['open', 'close', 'openFull', 'openFullHint'].map((k) => `manual.help.${k}`),
      ...[
        'how',
        'modes',
        'abilities',
        'odds',
        'effects',
        'battleIQ',
        'conventions',
        'etiquette',
      ].map((s) => `manual.sections.${s}`),
      'manual.conventions.absolute.title',
      'manual.conventions.absolute.text',
      ...['actors', 'targets', 'overcomers', 'theStand', 'immune', 'attenuated'].map(
        (l) => `manual.conventions.labels.${l}`
      ),
      ...['anyStand', 'anyFruit', 'fruitType', 'slotAny', 'slotMin'].map(
        (k) => `manual.subject.${k}`
      ),
      ...['honest', 'respect', 'host', 'spoilers'].flatMap((g) => [
        `manual.etiquette.${g}.title`,
        `manual.etiquette.${g}.a`,
        `manual.etiquette.${g}.b`,
      ]),
      'nav.manual',
    ]
    expect(fixed.filter((key) => !has(catalog, key))).toEqual([])
  })
})

// A translation that renames or drops a {{parameter}} would show the player a
// literal "{{ext}}" (or silently lose a number), and the key-parity test cannot
// see that: the key exists, only its placeholders differ.
function manualStrings(node: unknown, prefix: string, out: Record<string, string>) {
  if (typeof node === 'string') out[prefix] = node
  else if (node && typeof node === 'object') {
    for (const [k, v] of Object.entries(node)) manualStrings(v, `${prefix}.${k}`, out)
  }
  return out
}

const placeholders = (s: string) => [...s.matchAll(/\{\{(\w+)\}\}/g)].map((m) => m[1]).sort()

describe.each([
  ['es-ES', esES],
  ['ca-ES', caES],
])('manual copy in %s keeps the placeholders of en-GB', (_locale, catalog) => {
  it('uses the same {{parameters}} in every string', () => {
    const en = manualStrings((enGB as Record<string, unknown>).manual, 'manual', {})
    const other = manualStrings((catalog as Record<string, unknown>).manual, 'manual', {})
    const mismatched = Object.keys(en).filter(
      (key) =>
        JSON.stringify(placeholders(en[key])) !== JSON.stringify(placeholders(other[key] ?? ''))
    )
    expect(mismatched).toEqual([])
  })
})
