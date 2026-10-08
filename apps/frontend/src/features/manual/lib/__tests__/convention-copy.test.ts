import { MANUAL_RULES } from '@/shared/contracts/rules'

import type { ManualConvention } from '../../types/manual'
import {
  actorsLabelKey,
  conventionText,
  conventionTitle,
  groupByCategory,
  isAbsolute,
  targetsLabelKey,
} from '../convention-copy'

const conventions: readonly ManualConvention[] = MANUAL_RULES.conventions
const byId = (id: string) => {
  const found = conventions.find((c) => c.id === id)
  if (!found) throw new Error(`no convention ${id}`)
  return found
}

// A t() that returns the key plus its options, so the test sees exactly which
// key and which parameters a convention asks for.
const t = (key: string, options?: Record<string, unknown>) =>
  options ? `${key}|${JSON.stringify(options)}` : key

describe('convention copy', () => {
  it('titles an absolute Stand by its own name', () => {
    const ger = byId('ABSOLUTE_GER')
    expect(isAbsolute(ger)).toBe(true)
    expect(conventionTitle(t, ger)).toBe(
      'manual.conventions.absolute.title|{"stand":"Gold Experience: Requiem"}'
    )
    expect(conventionText(t, ger)).toContain('"stand":"Gold Experience: Requiem"')
  })

  it('gives every other convention its own copy keys, with the Conqueror gap', () => {
    const knockout = byId('CONQUEROR_KNOCKOUT')
    expect(isAbsolute(knockout)).toBe(false)
    expect(conventionTitle(t, knockout)).toBe('manual.conventions.items.CONQUEROR_KNOCKOUT.title')
    expect(conventionText(t, knockout)).toBe(
      'manual.conventions.items.CONQUEROR_KNOCKOUT.text|{"gap":2}'
    )
  })

  it('reads the chip rows of an absolute Stand as "who can overcome it"', () => {
    expect(actorsLabelKey(byId('ABSOLUTE_GER'))).toBe('manual.conventions.labels.overcomers')
    expect(targetsLabelKey(byId('ABSOLUTE_GER'))).toBe('manual.conventions.labels.theStand')
    expect(actorsLabelKey(byId('SEE_STANDS'))).toBe('manual.conventions.labels.actors')
  })

  it('groups by category in the backend order without losing or duplicating any', () => {
    const groups = groupByCategory(conventions)
    expect(groups.flatMap((g) => g.items.map((c) => c.id))).toEqual(
      [...conventions]
        .sort(
          (a, b) =>
            groups.findIndex((g) => g.category === a.category) -
            groups.findIndex((g) => g.category === b.category)
        )
        .map((c) => c.id)
    )
    expect(new Set(groups.map((g) => g.category)).size).toBe(groups.length)
    expect(groups.reduce((n, g) => n + g.items.length, 0)).toBe(conventions.length)
  })
})
