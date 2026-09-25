import { mulberry32, seededShuffle } from '@/features/game/lib/prng'

describe('mulberry32', () => {
  it('is deterministic - the same seed always produces the same sequence', () => {
    const a = mulberry32(42)
    const b = mulberry32(42)
    const seqA = Array.from({ length: 10 }, () => a())
    const seqB = Array.from({ length: 10 }, () => b())
    expect(seqA).toEqual(seqB)
  })

  it('different seeds produce different sequences', () => {
    const a = mulberry32(1)
    const b = mulberry32(2)
    expect(a()).not.toBe(b())
  })

  it('always returns a value in [0, 1)', () => {
    const rand = mulberry32(7)
    for (let i = 0; i < 200; i++) {
      const v = rand()
      expect(v).toBeGreaterThanOrEqual(0)
      expect(v).toBeLessThan(1)
    }
  })
})

describe('seededShuffle', () => {
  it('is a permutation - same elements, same length', () => {
    const items = [1, 2, 3, 4, 5]
    const shuffled = seededShuffle(items, mulberry32(1))
    expect(shuffled).toHaveLength(items.length)
    expect(shuffled.slice().sort()).toEqual(items.slice().sort())
  })

  it('does not mutate the input array', () => {
    const items = [1, 2, 3]
    const copy = items.slice()
    seededShuffle(items, mulberry32(1))
    expect(items).toEqual(copy)
  })

  it('is deterministic for a fresh generator with the same seed', () => {
    const items = ['a', 'b', 'c', 'd', 'e']
    const first = seededShuffle(items, mulberry32(99))
    const second = seededShuffle(items, mulberry32(99))
    expect(first).toEqual(second)
  })
})
