import { battleIQCategory, battleIQCategoryKey, formatBattleIQ } from '@/shared/lib/battle-iq'

describe('battleIQCategory', () => {
  it.each([
    [0, 'EXTREMELY_LOW'],
    [69, 'EXTREMELY_LOW'],
    [70, 'BORDERLINE'],
    [79, 'BORDERLINE'],
    [80, 'LOW_AVERAGE'],
    [89, 'LOW_AVERAGE'],
    [90, 'AVERAGE'],
    [109, 'AVERAGE'],
    [110, 'HIGH_AVERAGE'],
    [119, 'HIGH_AVERAGE'],
    [120, 'SUPERIOR'],
    [129, 'SUPERIOR'],
    [130, 'VERY_SUPERIOR'],
    [255, 'VERY_SUPERIOR'],
  ] as const)('%d -> %s', (score, want) => {
    expect(battleIQCategory(score)).toBe(want)
  })
})

describe('battleIQCategoryKey', () => {
  it('namespaces under enums.battleIQCategory', () => {
    expect(battleIQCategoryKey(100)).toBe('enums.battleIQCategory.AVERAGE')
  })
})

describe('formatBattleIQ', () => {
  const t = (key: string) => key

  it('returns null when the score is absent', () => {
    expect(formatBattleIQ(t, undefined)).toBeNull()
  })

  it('never collapses a present 0 into null', () => {
    expect(formatBattleIQ(t, 0)).toBe('0 · enums.battleIQCategory.EXTREMELY_LOW')
  })

  it('formats a present score with its category label', () => {
    expect(formatBattleIQ(t, 115)).toBe('115 · enums.battleIQCategory.HIGH_AVERAGE')
  })
})
