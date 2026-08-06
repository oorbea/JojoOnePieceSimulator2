import caES from '@/shared/i18n/locales/ca-ES.json'
import enGB from '@/shared/i18n/locales/en-GB.json'
import esES from '@/shared/i18n/locales/es-ES.json'

// Guards against a translation catalog silently drifting out of sync: a key
// added in en-GB (the mandatory baseline, same role as the backend's en-GB
// content fallback) but forgotten in es-ES/ca-ES would otherwise only
// surface at runtime as a raw "namespace.key" string shown to a user.
function flattenKeys(obj: Record<string, unknown>, prefix = ''): string[] {
  return Object.entries(obj).flatMap(([key, value]) => {
    const path = prefix ? `${prefix}.${key}` : key
    if (value && typeof value === 'object' && !Array.isArray(value)) {
      return flattenKeys(value as Record<string, unknown>, path)
    }
    return [path]
  })
}

describe('i18n locale catalogs have the same keys', () => {
  const enKeys = flattenKeys(enGB).sort()

  it.each([
    ['es-ES', esES],
    ['ca-ES', caES],
  ])('%s has exactly the same keys as en-GB', (_locale, catalog) => {
    const keys = flattenKeys(catalog as Record<string, unknown>).sort()
    expect(keys).toEqual(enKeys)
  })
})
