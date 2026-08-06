import {
  columnMaxWidth,
  DOCK_BAR_BOTTOM_OFFSET,
  DOCK_BAR_CLEARANCE,
  NAV_BAR_CLEARANCE,
  NAV_BAR_TOP_OFFSET,
  navBottomInset,
  navTopInset,
} from '../layout'

const ZERO_INSETS = { top: 0, bottom: 0 }

describe('navTopInset / navBottomInset', () => {
  it('reserves the bar\'s real measured height + offset + clearance', () => {
    expect(navTopInset(ZERO_INSETS, 80)).toBe(NAV_BAR_TOP_OFFSET + 80 + NAV_BAR_CLEARANCE)
    expect(navBottomInset(ZERO_INSETS, 96)).toBe(DOCK_BAR_BOTTOM_OFFSET + 96 + DOCK_BAR_CLEARANCE)
  })

  it('tracks a bar that grew past its nominal height (e.g. it wrapped)', () => {
    // The whole point of measuring instead of assuming a fixed constant:
    // a taller-than-usual bar must reserve more, not the same, space.
    expect(navTopInset(ZERO_INSETS, 128)).toBeGreaterThan(navTopInset(ZERO_INSETS, 64))
  })

  it('falls back to plain breathing room when the bar is not rendered at all', () => {
    expect(navTopInset(ZERO_INSETS, null)).toBe(NAV_BAR_CLEARANCE)
    expect(navBottomInset(ZERO_INSETS, null)).toBe(DOCK_BAR_CLEARANCE)
  })

  it('adds device safe-area insets on top of the bar reservation', () => {
    const insets = { top: 44, bottom: 34 }
    expect(navTopInset(insets, 64)).toBe(44 + NAV_BAR_TOP_OFFSET + 64 + NAV_BAR_CLEARANCE)
    expect(navBottomInset(insets, 64)).toBe(34 + DOCK_BAR_BOTTOM_OFFSET + 64 + DOCK_BAR_CLEARANCE)
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
