import {
  bottomClearance,
  columnMaxWidth,
  desktopBottomClearance,
  DOCK_BAR_BOTTOM_OFFSET,
  DOCK_BAR_CLEARANCE,
  DOCK_BAR_HEIGHT,
  NAV_BAR_CLEARANCE,
  NAV_BAR_HEIGHT,
  NAV_BAR_TOP_OFFSET,
  topClearance,
} from '../layout'

const ZERO_INSETS = { top: 0, bottom: 0 }

describe('topClearance / bottomClearance', () => {
  it('reserves exactly the nav bar height + offset + clearance when docked', () => {
    expect(topClearance(ZERO_INSETS, true)).toBe(
      NAV_BAR_TOP_OFFSET + NAV_BAR_HEIGHT + NAV_BAR_CLEARANCE
    )
    expect(bottomClearance(ZERO_INSETS, true)).toBe(
      DOCK_BAR_BOTTOM_OFFSET + DOCK_BAR_HEIGHT + DOCK_BAR_CLEARANCE
    )
  })

  it('falls back to plain breathing room when not docked', () => {
    expect(topClearance(ZERO_INSETS, false)).toBe(16)
    expect(bottomClearance(ZERO_INSETS, false)).toBe(16)
  })

  it('adds device safe-area insets on top of the bar reservation', () => {
    const insets = { top: 44, bottom: 34 }
    expect(topClearance(insets, true)).toBe(44 + NAV_BAR_TOP_OFFSET + NAV_BAR_HEIGHT + NAV_BAR_CLEARANCE)
    expect(bottomClearance(insets, true)).toBe(
      34 + DOCK_BAR_BOTTOM_OFFSET + DOCK_BAR_HEIGHT + DOCK_BAR_CLEARANCE
    )
  })
})

describe('desktopBottomClearance', () => {
  // The bottom dock is display:none from $md up (channel-bar.tsx /
  // app-shell.tsx) — desktop pages must not carry the mobile dock's ~96px
  // reservation as dead space once it's not actually rendered.
  it('is independent of the dock height', () => {
    expect(desktopBottomClearance(ZERO_INSETS)).toBeLessThan(
      bottomClearance(ZERO_INSETS, true)
    )
  })
})

describe('columnMaxWidth', () => {
  it('returns the base width unchanged at the base tier', () => {
    expect(columnMaxWidth(480, 'base')).toBe(480)
  })

  it('grows monotonically across breakpoints, up to the alignment cap', () => {
    const base = 480
    const md = columnMaxWidth(base, 'md')
    const lg = columnMaxWidth(base, 'lg')
    const xl = columnMaxWidth(base, 'xl')

    expect(md).toBeGreaterThan(base)
    expect(lg).toBeGreaterThan(md)
    expect(xl).toBeGreaterThanOrEqual(lg)
    expect(xl).toBeLessThanOrEqual(1080)
  })

  it('never exceeds the nav bar alignment width, even for a wide base', () => {
    expect(columnMaxWidth(900, 'xl')).toBe(1080)
  })

  it('honors a custom alignment cap', () => {
    expect(columnMaxWidth(900, 'xl', 1200)).toBeLessThanOrEqual(1200)
  })
})
