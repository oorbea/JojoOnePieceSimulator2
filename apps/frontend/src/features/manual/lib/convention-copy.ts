import type { ManualConvention } from '../types/manual'

type T = (key: string, options?: Record<string, unknown>) => string

// The absolute Stands share one wording ("{{stand}} is absolute"), so only the
// Stand's name changes; every other convention has its own title and text under
// `manual.conventions.items.<ID>`.
const ABSOLUTE_PREFIX = 'ABSOLUTE_'

export function isAbsolute(c: ManualConvention): boolean {
  return c.id.startsWith(ABSOLUTE_PREFIX)
}

function absoluteStand(c: ManualConvention): string {
  return c.targets[0]?.names?.[0] ?? ''
}

export function conventionTitle(t: T, c: ManualConvention): string {
  if (isAbsolute(c)) return t('manual.conventions.absolute.title', { stand: absoluteStand(c) })
  return t(`manual.conventions.items.${c.id}.title`)
}

export function conventionText(t: T, c: ManualConvention): string {
  if (isAbsolute(c)) return t('manual.conventions.absolute.text', { stand: absoluteStand(c) })
  return t(`manual.conventions.items.${c.id}.text`, { gap: c.gap })
}

// What a row of chips means depends on the convention: for an absolute Stand
// the "actors" are the ones who can overcome it, not who does something.
export function actorsLabelKey(c: ManualConvention): string {
  return isAbsolute(c) ? 'manual.conventions.labels.overcomers' : 'manual.conventions.labels.actors'
}

export function targetsLabelKey(c: ManualConvention): string {
  return isAbsolute(c) ? 'manual.conventions.labels.theStand' : 'manual.conventions.labels.targets'
}

// Conventions grouped by category, categories and conventions in the order the
// backend table lists them.
export function groupByCategory(
  conventions: readonly ManualConvention[]
): { category: string; items: ManualConvention[] }[] {
  const groups: { category: string; items: ManualConvention[] }[] = []
  for (const c of conventions) {
    let group = groups.find((g) => g.category === c.category)
    if (!group) {
      group = { category: c.category, items: [] }
      groups.push(group)
    }
    group.items.push(c)
  }
  return groups
}
