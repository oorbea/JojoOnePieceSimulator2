import { computePoolCounts, poolShortfalls, searchPoolItems, type PoolItem } from '../pool-stats'

const STAR_PLATINUM: PoolItem = { id: 'stand-1', name: 'Star Platinum', kind: 'STAND' }
const CRAZY_DIAMOND: PoolItem = { id: 'stand-2', name: 'Crazy Diamond', kind: 'STAND' }
const HERMES: PoolItem = { id: 'stand-3', name: 'Hermes', kind: 'STAND' }
const GOMU_GOMU: PoolItem = { id: 'fruit-1', name: 'Gomu Gomu no Mi', kind: 'DEVIL_FRUIT' }
const MERA_MERA: PoolItem = { id: 'fruit-2', name: 'Mera Mera no Mi', kind: 'DEVIL_FRUIT' }

const CATALOG: PoolItem[] = [STAR_PLATINUM, CRAZY_DIAMOND, HERMES, GOMU_GOMU, MERA_MERA]

describe('computePoolCounts', () => {
  it('splits totals and remaining per kind', () => {
    const counts = computePoolCounts(CATALOG, [])
    expect(counts.STAND).toEqual({ total: 3, remaining: 3 })
    expect(counts.DEVIL_FRUIT).toEqual({ total: 2, remaining: 2 })
  })

  it('subtracts banned ids from remaining', () => {
    const counts = computePoolCounts(CATALOG, [STAR_PLATINUM.id])
    expect(counts.STAND).toEqual({ total: 3, remaining: 2 })
  })

  it('ignores banned ids absent from the catalog', () => {
    const counts = computePoolCounts(CATALOG, ['not-in-catalog'])
    expect(counts.STAND).toEqual({ total: 3, remaining: 3 })
    expect(counts.DEVIL_FRUIT).toEqual({ total: 2, remaining: 2 })
  })

  it('does not double-count duplicate banned ids', () => {
    const counts = computePoolCounts(CATALOG, [STAR_PLATINUM.id, STAR_PLATINUM.id])
    expect(counts.STAND.remaining).toBe(2)
  })

  it('never returns a negative remaining', () => {
    const allBanned = CATALOG.map((item) => item.id)
    const counts = computePoolCounts(CATALOG, allBanned)
    expect(counts.STAND.remaining).toBe(0)
    expect(counts.DEVIL_FRUIT.remaining).toBe(0)
  })

  it('returns zeroed counts for an empty catalog', () => {
    const counts = computePoolCounts([], [])
    expect(counts.STAND).toEqual({ total: 0, remaining: 0 })
    expect(counts.DEVIL_FRUIT).toEqual({ total: 0, remaining: 0 })
  })
})

describe('poolShortfalls', () => {
  it('is empty when remaining meets or exceeds required', () => {
    const counts = computePoolCounts(CATALOG, [])
    expect(poolShortfalls(counts, ['JOJO', 'ONE_PIECE'], 2, true)).toEqual([])
  })

  it('flags a Stand shortfall for JOJO when remaining Stands are below required', () => {
    const counts = computePoolCounts(CATALOG, [STAR_PLATINUM.id, CRAZY_DIAMOND.id])
    const shortfalls = poolShortfalls(counts, ['JOJO'], 3, true)
    expect(shortfalls).toEqual([{ manga: 'JOJO', kind: 'STAND', remaining: 1, required: 3 }])
  })

  it('flags a Devil Fruit shortfall for ONE_PIECE when remaining fruits are below required', () => {
    const counts = computePoolCounts(CATALOG, [GOMU_GOMU.id])
    const shortfalls = poolShortfalls(counts, ['ONE_PIECE'], 2, true)
    expect(shortfalls).toEqual([{ manga: 'ONE_PIECE', kind: 'DEVIL_FRUIT', remaining: 1, required: 2 }])
  })

  it('only includes selected power mangas even if the other kind is also short', () => {
    const counts = computePoolCounts(CATALOG, [GOMU_GOMU.id, MERA_MERA.id, STAR_PLATINUM.id, CRAZY_DIAMOND.id, HERMES.id])
    // Both kinds are fully exhausted, but only JOJO is selected.
    const shortfalls = poolShortfalls(counts, ['JOJO'], 1, true)
    expect(shortfalls).toEqual([{ manga: 'JOJO', kind: 'STAND', remaining: 0, required: 1 }])
  })

  it('is empty when the catalog is not yet known (still loading)', () => {
    const counts = computePoolCounts([], [])
    expect(poolShortfalls(counts, ['JOJO', 'ONE_PIECE'], 5, false)).toEqual([])
  })
})

describe('searchPoolItems', () => {
  it('is case-insensitive', () => {
    const { results } = searchPoolItems(CATALOG, [], 'star platinum')
    expect(results.map((r) => r.id)).toEqual([STAR_PLATINUM.id])
  })

  it('is diacritic-insensitive', () => {
    const items: PoolItem[] = [{ id: 'x', name: 'Méra Méra no Mi', kind: 'DEVIL_FRUIT' }]
    const { results } = searchPoolItems(items, [], 'mera mera')
    expect(results.map((r) => r.id)).toEqual(['x'])
  })

  it('requires every whitespace-separated token to match (AND), order-independent', () => {
    const { results } = searchPoolItems(CATALOG, [], 'no gomu')
    expect(results.map((r) => r.id)).toEqual([GOMU_GOMU.id])
  })

  it('excludes already-banned items', () => {
    const { results, total } = searchPoolItems(CATALOG, [STAR_PLATINUM.id], 'star')
    expect(results).toEqual([])
    expect(total).toBe(0)
  })

  it('respects limit while reporting the true total', () => {
    const items: PoolItem[] = [
      { id: '1', name: 'Stand Alpha', kind: 'STAND' },
      { id: '2', name: 'Stand Beta', kind: 'STAND' },
      { id: '3', name: 'Stand Gamma', kind: 'STAND' },
    ]
    const { results, total } = searchPoolItems(items, [], 'stand', 2)
    expect(results).toHaveLength(2)
    expect(total).toBe(3)
  })

  it('returns nothing for an empty or whitespace-only query', () => {
    expect(searchPoolItems(CATALOG, [], '')).toEqual({ results: [], total: 0 })
    expect(searchPoolItems(CATALOG, [], '   ')).toEqual({ results: [], total: 0 })
  })
})
