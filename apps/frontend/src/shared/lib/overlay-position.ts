// Pure viewport-clamping math for tooltip.tsx's floating overlays (the text
// bubble and the bigger hover card) - split out so "does an oversized
// overlay ever run off the screen" is unit-testable without rendering
// anything. This is exactly the invariant that broke for the hover card: a
// ~450px LoadoutCard is routinely taller than the room above OR below a
// roster tile, and the old logic only ever checked the near edge picked by
// the above/below flip, never the far one.

export type Rect = { x: number; y: number; width: number; height: number }
export type Size = { width: number; height: number }
export type Screen = { width: number; height: number }

export type OverlayPosition = { top: number; left: number }

// clampOverlayPosition centres `size` horizontally on `anchor`'s midpoint,
// and vertically prefers placing it just above the anchor, falling back to
// just below when there isn't room above, and finally clamping to whichever
// side has more room (still touching that edge, never running off it) when
// `size` doesn't fully fit on either side.
export function clampOverlayPosition(
  anchor: Rect,
  size: Size,
  screen: Screen,
  margin: number
): OverlayPosition {
  const { width, height } = size

  const desiredLeft = anchor.x + anchor.width / 2 - width / 2
  const maxLeft = Math.max(screen.width - width - margin, margin)
  const left = Math.min(Math.max(desiredLeft, margin), maxLeft)

  const roomAbove = anchor.y - 8 - margin
  const roomBelow = screen.height - (anchor.y + anchor.height) - 8 - margin
  let top: number
  if (height <= roomAbove) {
    top = anchor.y - 8 - height
  } else if (height <= roomBelow) {
    top = anchor.y + anchor.height + 8
  } else if (roomAbove >= roomBelow) {
    // Doesn't fully fit above either, but there's more room there - clamp
    // to the top edge rather than let it run off.
    top = Math.max(margin, anchor.y - 8 - height)
  } else {
    top = anchor.y + anchor.height + 8
  }
  // Final safety clamp regardless of which branch ran, so the far edge
  // (bottom, in the 'below' branches) can never exceed the viewport either.
  top = Math.min(Math.max(top, margin), Math.max(screen.height - height - margin, margin))

  return { top, left }
}
