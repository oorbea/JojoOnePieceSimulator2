import { clampOverlayPosition } from '@/shared/lib/overlay-position'

const SCREEN = { width: 400, height: 800 }
const MARGIN = 8

// This is exactly the invariant that broke for the hover card: a
// ~450px-tall LoadoutCard is routinely taller than the room above OR below
// a roster tile, and the old logic only ever checked the near edge picked
// by the above/below flip, never the far one - so it ran off whichever
// edge it flipped away from.
describe('clampOverlayPosition', () => {
  it('places a small overlay above the anchor when there is room', () => {
    const anchor = { x: 150, y: 400, width: 88, height: 60 }
    const size = { width: 100, height: 40 }
    const { top, left } = clampOverlayPosition(anchor, size, SCREEN, MARGIN)
    expect(top).toBe(anchor.y - 8 - size.height)
    expect(left).toBe(anchor.x + anchor.width / 2 - size.width / 2)
  })

  it('flips below when there is no room above', () => {
    const anchor = { x: 150, y: 20, width: 88, height: 60 }
    const size = { width: 100, height: 40 }
    const { top } = clampOverlayPosition(anchor, size, SCREEN, MARGIN)
    expect(top).toBe(anchor.y + anchor.height + 8)
  })

  it('a card taller than the room both above AND below stays fully on-screen', () => {
    // Anchor near vertical middle of an 800px screen: ~400px of room on
    // either side, but the card is 600px tall - too tall for both.
    const anchor = { x: 150, y: 400, width: 88, height: 60 }
    const size = { width: 220, height: 600 }
    const { top } = clampOverlayPosition(anchor, size, SCREEN, MARGIN)

    expect(top).toBeGreaterThanOrEqual(MARGIN)
    expect(top + size.height).toBeLessThanOrEqual(SCREEN.height - MARGIN)
  })

  it('a card taller than the entire screen still stays within [margin, screen] as best-effort', () => {
    const anchor = { x: 150, y: 400, width: 88, height: 60 }
    const size = { width: 220, height: 1200 } // taller than SCREEN.height itself
    const { top } = clampOverlayPosition(anchor, size, SCREEN, MARGIN)

    expect(top).toBeGreaterThanOrEqual(MARGIN)
  })

  it('clamps horizontally so a wide overlay near the left edge never runs off', () => {
    const anchor = { x: 4, y: 400, width: 40, height: 40 }
    const size = { width: 220, height: 100 }
    const { left } = clampOverlayPosition(anchor, size, SCREEN, MARGIN)

    expect(left).toBeGreaterThanOrEqual(MARGIN)
  })

  it('clamps horizontally so a wide overlay near the right edge never runs off', () => {
    const anchor = { x: SCREEN.width - 40, y: 400, width: 40, height: 40 }
    const size = { width: 220, height: 100 }
    const { left } = clampOverlayPosition(anchor, size, SCREEN, MARGIN)

    expect(left + size.width).toBeLessThanOrEqual(SCREEN.width - MARGIN + 0.001)
  })
})
