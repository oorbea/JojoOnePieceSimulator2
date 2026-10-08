// The manual's sections, in reading order. The id is the `?section=` deep-link
// value and the `manual.sections.<id>` i18n key.
export const MANUAL_SECTIONS = [
  'how',
  'modes',
  'abilities',
  'odds',
  'effects',
  'battleIQ',
  'conventions',
  'etiquette',
] as const

export type ManualSectionId = (typeof MANUAL_SECTIONS)[number]

// `?section=` comes from the router as string | string[] | undefined; anything
// that is not a known section is ignored rather than breaking the page.
export function parseSection(raw: unknown): ManualSectionId | null {
  const value = Array.isArray(raw) ? raw[0] : raw
  return typeof value === 'string' && (MANUAL_SECTIONS as readonly string[]).includes(value)
    ? (value as ManualSectionId)
    : null
}
