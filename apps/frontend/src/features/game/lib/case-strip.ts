import { mulberry32, seededShuffle } from '@/features/game/lib/prng'

// buildCaseStrip lays out a CS:GO-style horizontal case-opening strip for
// the Stand/Devil Fruit sorteo slots (owner request, 2026-09-25 playtest
// feedback: the old vertical roulette gave the answer away immediately and
// only ever showed a handful of repeated candidates). Unlike the vertical
// PowerRoulette's reel (built for a short enum-backed candidate list), this
// strip draws from the lobby's real, possibly large power pool, so it needs
// its own shuffling: the winner spliced in once, deep into the strip (never
// early), never adjacent to itself, and a landing offset that doesn't
// always stop dead-centre on the card (CS's own "near-miss" catch).
export type CaseStripCard = {
  id: string
  label: string
  /** A PowerRarity value, or 'NONE' for the "landed nothing" card - never a
   * duplicate of a real rarity string, so the colour-bar mapping can treat
   * it as its own tier. */
  rarity: string
  picture?: string
}

export type CaseStrip = {
  cards: CaseStripCard[]
  /** Index into `cards` - where the real answer sits. Always in the last
   * 40% of the strip, so glancing at (or skipping past) the strip before
   * the spin settles never previews the answer. */
  landingIndex: number
  /** 0..1 - how far into the landing card's own width the strip actually
   * stops, so it isn't always a dead-centre stop. */
  landingOffset: number
}

// A CS-style strip needs to feel long enough to scroll through - shorter
// and the "many possible outcomes" illusion falls apart immediately.
const STRIP_LENGTH = 50
// The winner never lands before this fraction of the strip.
const MIN_LANDING_FRACTION = 0.6

// pickAvoiding draws from `pool` via `rand`, retrying (bounded) to avoid
// landing on the SAME id as `avoid` - used to keep two adjacent cards from
// ever showing the same power, which would read as "wait, was that also a
// possible answer?".
function pickAvoiding(
  pool: CaseStripCard[],
  avoid: string | null,
  rand: () => number
): CaseStripCard {
  let pick = pool[Math.floor(rand() * pool.length)]
  let guard = 0
  while (pick.id === avoid && pool.length > 1 && guard < 8) {
    pick = pool[Math.floor(rand() * pool.length)]
    guard += 1
  }
  return pick
}

export function buildCaseStrip(
  pool: CaseStripCard[],
  winner: CaseStripCard,
  seed: number,
  length: number = STRIP_LENGTH
): CaseStrip {
  const rand = mulberry32(seed)
  // Excludes the winner from the sampling pool entirely, rather than
  // filtering duplicates out after the fact - the only way to guarantee it
  // never appears anywhere but landingIndex, not just "not adjacent to
  // itself". Falls back to [winner] only in the degenerate case where the
  // pool WAS just the winner (e.g. a lobby pool restricted to one power) -
  // there's nothing else to draw from, so the strip is uniform and the
  // "only at landingIndex" invariant no longer applies.
  const decoys = pool.filter((c) => c.id !== winner.id)
  const source = decoys.length > 0 ? seededShuffle(decoys, rand) : [winner]

  const cards: CaseStripCard[] = []
  let prevId: string | null = null
  for (let i = 0; i < length; i++) {
    const pick = pickAvoiding(source, prevId, rand)
    cards.push(pick)
    prevId = pick.id
  }

  const minLanding = Math.floor(length * MIN_LANDING_FRACTION)
  const maxLanding = length - 2 // never the very last card - always a real neighbour past it
  const landingIndex = minLanding + Math.floor(rand() * Math.max(1, maxLanding - minLanding))
  cards[landingIndex] = winner
  // Re-roll either neighbour if it happens to already equal the winner
  // (only possible in the degenerate source === [winner] case, or a
  // vanishingly unlucky roll right at the pool's own boundary) - a
  // duplicate winner card right next to the real one reads as two
  // candidate landing spots instead of one.
  if (decoys.length > 0) {
    if (cards[landingIndex - 1]?.id === winner.id) {
      cards[landingIndex - 1] = pickAvoiding(source, winner.id, rand)
    }
    if (cards[landingIndex + 1]?.id === winner.id) {
      cards[landingIndex + 1] = pickAvoiding(source, winner.id, rand)
    }
  }

  // Never flush against either edge of the card - a dead-centre stop every
  // single time is exactly the "too clean" tell CS's own near-miss catch
  // avoids.
  const landingOffset = 0.15 + rand() * 0.7

  return { cards, landingIndex, landingOffset }
}

// caseStripRestX is CaseStripReel's own geometry, split out the same way
// reel-geometry.ts's restRows/finalLabelIndex are - so "does the needle
// actually land inside the winning card, never past the strip's end" is
// unit-testable without rendering reanimated. `cardWidth`/`cardGap` are the
// card's own footprint (owner request, 2026-09-25: real card art + a
// rarity colour bar, not just a text row), and the needle is a fixed
// marker at `windowWidth / 2`; the strip itself is what moves under it.
export function caseStripRestX(
  landingIndex: number,
  landingOffset: number,
  cardWidth: number,
  cardGap: number,
  windowWidth: number
): number {
  const step = cardWidth + cardGap
  // Distance from the strip's own left edge to the point inside the
  // landing card the needle should stop on.
  const landingPointPx = landingIndex * step + cardWidth * landingOffset
  return windowWidth / 2 - landingPointPx
}
