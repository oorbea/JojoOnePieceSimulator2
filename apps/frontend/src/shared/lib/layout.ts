// Pure layout math shared between PageShell and (indirectly) AppShell/
// ChannelBar. Pulled out of page-shell.tsx so the numbers a page uses to
// clear the floating nav bars, and the column widths it grows through at
// each breakpoint, are named, unit-tested constants instead of magic
// numbers re-derived by hand in JSX.

/** Height of a ChannelBar pill (top bar or bottom dock) — see
 * channel-bar.tsx. Kept in one place so PageShell's clearance can never
 * drift out of sync with the bar it's clearing. */
export const NAV_BAR_HEIGHT = 64
export const NAV_BAR_TOP_OFFSET = 8
export const NAV_BAR_CLEARANCE = 16

export const DOCK_BAR_HEIGHT = 64
export const DOCK_BAR_BOTTOM_OFFSET = 8
export const DOCK_BAR_CLEARANCE = 24

type Insets = { top: number; bottom: number }

/** Top padding a page needs to clear the floating top ChannelBar without
 * overlapping it. `hasNav` mirrors PageShell's `navPadding` prop. */
export function topClearance(insets: Insets, hasNav: boolean) {
  return hasNav
    ? insets.top + NAV_BAR_TOP_OFFSET + NAV_BAR_HEIGHT + NAV_BAR_CLEARANCE
    : insets.top + 16
}

/** Bottom padding for the mobile layout, where the floating bottom dock is
 * actually rendered (it's `display:none` from `$md` up — see
 * `desktopBottomClearance` for that tier). */
export function bottomClearance(insets: Insets, hasNav: boolean) {
  return hasNav
    ? insets.bottom + DOCK_BAR_BOTTOM_OFFSET + DOCK_BAR_HEIGHT + DOCK_BAR_CLEARANCE
    : insets.bottom + 16
}

/** Bottom padding once the dock is hidden (`$md` and up) — plain breathing
 * room, not sized to a bar that isn't on screen. Without this, desktop pages
 * kept the mobile dock's ~96px reservation as dead space at the bottom. */
export function desktopBottomClearance(insets: Insets) {
  return insets.bottom + 24
}

export type ColumnTier = 'base' | 'md' | 'lg' | 'xl'

/** Column max-width for a page's content at each breakpoint. `base` is the
 * mobile-first width a screen passes to PageShell; it grows through
 * `$md`/`$lg`/`$xl` so a wide page's column keeps expanding on desktop
 * instead of stopping after a single `$md` step, capped at `alignTo` — the
 * floating nav bars' own `maxW` (channel-bar.tsx) — so the content column
 * and the nav bar above it share the same right edge on very wide screens. */
export function columnMaxWidth(base: number, tier: ColumnTier, alignTo = 1080) {
  switch (tier) {
    case 'md':
      return Math.min(base * 1.18, alignTo)
    case 'lg':
      return Math.min(base * 1.5, alignTo)
    case 'xl':
      return Math.min(base * 1.8, alignTo)
    default:
      return base
  }
}
