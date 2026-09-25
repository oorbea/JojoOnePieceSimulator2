import { applyPoolFilter } from '@/features/game/lib/power-pool'
import type { StandResponse } from '@/features/stands'
import type { DevilFruitResponse } from '@/features/devil-fruits'
import type { PoolFilter } from '@/features/game/types/game.types'

function stand(id: string, rarity: StandResponse['rarity']): StandResponse {
  return { id, name: id, rarity } as StandResponse
}

function fruit(
  id: string,
  rarity: DevilFruitResponse['rarity'],
  fruitType: DevilFruitResponse['fruitType']
): DevilFruitResponse {
  return { id, name: id, rarity, fruitType } as DevilFruitResponse
}

function filter(overrides: Partial<PoolFilter> = {}): PoolFilter {
  return { standRarities: [], fruitRarities: [], fruitTypes: [], banned: [], ...overrides }
}

describe('applyPoolFilter', () => {
  const stands = [stand('s1', 'COMMON'), stand('s2', 'RARE'), stand('s3', 'LEGENDARY')]
  const fruits = [fruit('f1', 'COMMON', 'PARAMECIA'), fruit('f2', 'RARE', 'ZOAN')]

  it('with no restrictions, allows everything except banned ids', () => {
    const { stands: s, fruits: f } = applyPoolFilter(stands, fruits, filter())
    expect(s).toEqual(stands)
    expect(f).toEqual(fruits)
  })

  it('restricts stands to the given rarity allowlist', () => {
    const { stands: s } = applyPoolFilter(stands, fruits, filter({ standRarities: ['RARE'] }))
    expect(s.map((x) => x.id)).toEqual(['s2'])
  })

  it('restricts fruits to the given rarity allowlist', () => {
    const { fruits: f } = applyPoolFilter(stands, fruits, filter({ fruitRarities: ['COMMON'] }))
    expect(f.map((x) => x.id)).toEqual(['f1'])
  })

  it('restricts fruits to the given fruitType allowlist', () => {
    const { fruits: f } = applyPoolFilter(stands, fruits, filter({ fruitTypes: ['ZOAN'] }))
    expect(f.map((x) => x.id)).toEqual(['f2'])
  })

  it('excludes banned ids regardless of the allowlists', () => {
    const { stands: s, fruits: f } = applyPoolFilter(stands, fruits, filter({ banned: ['s1', 'f2'] }))
    expect(s.map((x) => x.id)).toEqual(['s2', 's3'])
    expect(f.map((x) => x.id)).toEqual(['f1'])
  })

  it('preserves the original order', () => {
    const { stands: s } = applyPoolFilter(stands, fruits, filter({ standRarities: ['COMMON', 'LEGENDARY'] }))
    expect(s.map((x) => x.id)).toEqual(['s1', 's3'])
  })
})
