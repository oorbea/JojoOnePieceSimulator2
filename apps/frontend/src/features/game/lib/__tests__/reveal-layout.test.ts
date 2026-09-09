import {
  REVEAL_ART_MAX,
  REVEAL_ART_MIN,
  REVEAL_SCROLL_MIN,
  revealLayout,
} from '@/features/game/lib/reveal-layout'

const ZERO_INSETS = { top: 0, bottom: 0 }

describe('revealLayout', () => {
  it('sizes a portrait phone (390x844) compact with 3 stat columns', () => {
    const layout = revealLayout({ width: 390, height: 844 }, ZERO_INSETS)
    expect(layout.artHeight).toBe(164)
    expect(layout.statColumns).toBe(3)
    expect(layout.scrollMaxHeight).toBeGreaterThan(REVEAL_SCROLL_MIN)
  })

  it('sizes a taller phone (430x932) slightly larger, still 3 columns', () => {
    const layout = revealLayout({ width: 430, height: 932 }, ZERO_INSETS)
    expect(layout.artHeight).toBe(181)
    expect(layout.statColumns).toBe(3)
  })

  it('sizes a desktop viewport (1280x800) with 6 stat columns', () => {
    const layout = revealLayout({ width: 1280, height: 800 }, ZERO_INSETS)
    expect(layout.artHeight).toBe(208)
    expect(layout.statColumns).toBe(6)
  })

  it('caps art height at the max on a large desktop viewport (1920x1080)', () => {
    const layout = revealLayout({ width: 1920, height: 1080 }, ZERO_INSETS)
    expect(layout.artHeight).toBe(REVEAL_ART_MAX)
    expect(layout.statColumns).toBe(6)
  })

  it('floors art height on a short landscape viewport (800x400)', () => {
    const layout = revealLayout({ width: 800, height: 400 }, ZERO_INSETS)
    expect(layout.artHeight).toBe(REVEAL_ART_MIN)
    expect(layout.statColumns).toBe(6)
  })

  it('floors art height and drops to 3 columns on a tiny window (500x400)', () => {
    const layout = revealLayout({ width: 500, height: 400 }, ZERO_INSETS)
    expect(layout.artHeight).toBe(REVEAL_ART_MIN)
    expect(layout.statColumns).toBe(3)
  })

  it('never lets scrollMaxHeight drop below REVEAL_SCROLL_MIN even on a very short viewport', () => {
    const layout = revealLayout({ width: 400, height: 200 }, ZERO_INSETS)
    expect(layout.scrollMaxHeight).toBe(REVEAL_SCROLL_MIN)
  })

  it('accounts for non-zero safe-area insets by shrinking the scroll budget', () => {
    const withoutInsets = revealLayout({ width: 400, height: 900 }, ZERO_INSETS)
    const withInsets = revealLayout({ width: 400, height: 900 }, { top: 44, bottom: 34 })
    expect(withInsets.scrollMaxHeight).toBe(withoutInsets.scrollMaxHeight - 78)
  })
})
