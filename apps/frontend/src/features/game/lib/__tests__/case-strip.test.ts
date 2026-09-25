import { buildCaseStrip, caseStripRestX, type CaseStripCard } from '@/features/game/lib/case-strip'

function pool(size: number, rarity = 'COMMON'): CaseStripCard[] {
  return Array.from({ length: size }, (_, i) => ({ id: `p${i}`, label: `Power ${i}`, rarity }))
}

const WINNER: CaseStripCard = { id: 'winner', label: 'Star Platinum', rarity: 'LEGENDARY' }

describe('buildCaseStrip', () => {
  it('is deterministic - the same seed always produces the same strip', () => {
    const a = buildCaseStrip(pool(20), WINNER, 12345)
    const b = buildCaseStrip(pool(20), WINNER, 12345)
    expect(a).toEqual(b)
  })

  it('different seeds produce different strips', () => {
    const a = buildCaseStrip(pool(20), WINNER, 1)
    const b = buildCaseStrip(pool(20), WINNER, 2)
    expect(a.cards.map((c) => c.id)).not.toEqual(b.cards.map((c) => c.id))
  })

  it('is at least 30 cards long', () => {
    const { cards } = buildCaseStrip(pool(20), WINNER, 1)
    expect(cards.length).toBeGreaterThanOrEqual(30)
  })

  it('places the winner at landingIndex, and ONLY there', () => {
    const { cards, landingIndex } = buildCaseStrip(pool(20), WINNER, 1)
    expect(cards[landingIndex]).toEqual(WINNER)
    const occurrences = cards.filter((c) => c.id === WINNER.id).length
    expect(occurrences).toBe(1)
  })

  it('never lands before 60% of the strip', () => {
    for (let seed = 0; seed < 50; seed++) {
      const { cards, landingIndex } = buildCaseStrip(pool(20), WINNER, seed)
      expect(landingIndex).toBeGreaterThanOrEqual(Math.floor(cards.length * 0.6))
      expect(landingIndex).toBeLessThanOrEqual(cards.length - 2)
    }
  })

  it('never repeats the same card on two adjacent slots', () => {
    for (let seed = 0; seed < 50; seed++) {
      const { cards } = buildCaseStrip(pool(20), WINNER, seed)
      for (let i = 1; i < cards.length; i++) {
        expect(cards[i].id).not.toBe(cards[i - 1].id)
      }
    }
  })

  it('draws from the given pool only (plus the winner) - never a candidate outside it', () => {
    const p = pool(8)
    const allowedIds = new Set([...p.map((c) => c.id), WINNER.id])
    const { cards } = buildCaseStrip(p, WINNER, 7)
    for (const card of cards) {
      expect(allowedIds.has(card.id)).toBe(true)
    }
  })

  it('landingOffset is strictly between 0 and 1, never a flush edge stop', () => {
    for (let seed = 0; seed < 20; seed++) {
      const { landingOffset } = buildCaseStrip(pool(20), WINNER, seed)
      expect(landingOffset).toBeGreaterThan(0)
      expect(landingOffset).toBeLessThan(1)
    }
  })

  it('falls back to a uniform strip of just the winner when the pool has nothing else', () => {
    const { cards, landingIndex } = buildCaseStrip([WINNER], WINNER, 1)
    expect(cards[landingIndex]).toEqual(WINNER)
    expect(cards.every((c) => c.id === WINNER.id)).toBe(true)
  })

  it('excludes the winner from the pool even if the caller passed it in twice', () => {
    const p = [...pool(5), WINNER]
    const { cards } = buildCaseStrip(p, WINNER, 3)
    expect(cards.filter((c) => c.id === WINNER.id)).toHaveLength(1)
  })
})

describe('caseStripRestX', () => {
  const CARD_WIDTH = 100
  const CARD_GAP = 10
  const WINDOW_WIDTH = 400

  it('centres the needle exactly on landingIndex 0 at landingOffset 0.5', () => {
    const restX = caseStripRestX(0, 0.5, CARD_WIDTH, CARD_GAP, WINDOW_WIDTH)
    // Card 0 spans [0, 100); its centre (50) must sit under the needle
    // (windowWidth/2 = 200) once translated.
    expect(restX + 50).toBe(WINDOW_WIDTH / 2)
  })

  it('moves the strip further left for a later landingIndex', () => {
    const early = caseStripRestX(2, 0.5, CARD_WIDTH, CARD_GAP, WINDOW_WIDTH)
    const later = caseStripRestX(10, 0.5, CARD_WIDTH, CARD_GAP, WINDOW_WIDTH)
    expect(later).toBeLessThan(early)
  })

  it('a landingOffset near 0 stops closer to the card start than one near 1', () => {
    const startOffset = caseStripRestX(5, 0.05, CARD_WIDTH, CARD_GAP, WINDOW_WIDTH)
    const endOffset = caseStripRestX(5, 0.95, CARD_WIDTH, CARD_GAP, WINDOW_WIDTH)
    // A larger landingOffset means the stop point is further right within
    // the card, i.e. the whole strip must shift further left (more
    // negative restX) to keep it under the fixed needle.
    expect(endOffset).toBeLessThan(startOffset)
  })
})
