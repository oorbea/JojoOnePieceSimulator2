// Pure sizing math for PowerRevealCard (the sorteo's full-screen power
// reveal) - split out so "does the card actually fit inside the viewport,
// especially a short landscape one" is unit-testable without rendering
// layout (RNTL doesn't). Same pattern as reel-geometry.ts's landingTiming
// and shared/lib/overlay-position.ts's clampOverlayPosition.

export type Viewport = { width: number; height: number }
export type Insets = { top: number; bottom: number }

export type RevealLayout = {
  /** Height of the art well - shrinks on short/narrow viewports so the
   * description and skills (the actual point of this card) keep their
   * size instead of getting clipped. */
  artHeight: number
  /** Max height for the scrollable body (art + description + skills +
   * stat grid) inside the panel, leaving room for the pinned participant
   * name above and Skip button below. */
  scrollMaxHeight: number
  /** Stat-grid columns: compact 3-wide on narrow phones, full 6-wide once
   * there's room. */
  statColumns: 3 | 6
}

export const REVEAL_ART_MIN = 120
export const REVEAL_ART_MAX = 260

// Participant-name row + Skip row + the panel's own padding/gaps that sit
// outside the ScrollView - kept as a named constant rather than re-derived
// from JSX so a future spacing tweak has one place to update.
export const REVEAL_CHROME_HEIGHT = 168

// Never squeeze the scrollable body below this even on the shortest
// realistic viewport - a ScrollView still works when its content exceeds
// this, it just scrolls more.
export const REVEAL_SCROLL_MIN = 160

// Below this width the stat grid drops to 3 columns (2 rows of 3 instead
// of a cramped single row of 6) - matches the app's own `sm` breakpoint
// (tamagui.config.ts) so it agrees with every other "phone vs. wider"
// decision in the codebase.
const STAT_GRID_WIDE_MIN = 640

function clamp(value: number, min: number, max: number): number {
  return Math.min(Math.max(value, min), max)
}

export function revealLayout(viewport: Viewport, insets: Insets): RevealLayout {
  const artHeight = Math.round(
    clamp(Math.min(viewport.height * 0.26, viewport.width * 0.42), REVEAL_ART_MIN, REVEAL_ART_MAX)
  )

  const scrollMaxHeight = Math.max(
    REVEAL_SCROLL_MIN,
    viewport.height - insets.top - insets.bottom - 24 - REVEAL_CHROME_HEIGHT
  )

  const statColumns: 3 | 6 = viewport.width >= STAT_GRID_WIDE_MIN ? 6 : 3

  return { artHeight, scrollMaxHeight, statColumns }
}
