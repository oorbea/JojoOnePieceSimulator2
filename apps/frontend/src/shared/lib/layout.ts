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
 * overlapping it. `barHeight` is the bar's real measured height (from
 * AppShell's onLayout) — pass `null` when the bar isn't rendered at all, or
 * `NAV_BAR_HEIGHT` as a first-render fallback before layout has fired. */
export function navTopInset(insets: Insets, barHeight: number | null) {
  if (barHeight === null) return insets.top + NAV_BAR_CLEARANCE
  return insets.top + NAV_BAR_TOP_OFFSET + barHeight + NAV_BAR_CLEARANCE
}

/** Bottom padding to clear the floating bottom dock. `dockHeight` is the
 * dock's real measured height, or `null` when it isn't rendered (e.g. the
 * desktop tier, where AppShell swaps the dock out for top nav links instead
 * of just hiding it) — in that case this is plain breathing room, not sized
 * to a bar that isn't on screen. */
export function navBottomInset(insets: Insets, dockHeight: number | null) {
  if (dockHeight === null) return insets.bottom + DOCK_BAR_CLEARANCE
  return insets.bottom + DOCK_BAR_BOTTOM_OFFSET + dockHeight + DOCK_BAR_CLEARANCE
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
