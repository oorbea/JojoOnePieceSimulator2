// The manual's sections, in reading order. The id is the `?section=` deep-link
// value, the `manual.sections.<id>` i18n key, and what any screen passes to
// the manual overlay (see manual-overlay.tsx). Lives in shared/ so a screen of
// another feature can name a section without importing the manual feature.
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
