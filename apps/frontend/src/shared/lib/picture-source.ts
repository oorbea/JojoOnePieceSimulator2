// Every picture-bearing entity (Stand, DevilFruit, Stage, User avatar) wires
// the same trio: picture (main, ~1024px), pictureThumb (~256px), and
// pictureStatus. This is the single place that picks which rendition a call
// site should render - see ObsidianVault/admin-panel-crud-ux-fixes.md: the
// thumb is empty until the transcode worker finishes, so `pictureThumb ||
// picture` (not a blind switch to pictureThumb) is required to avoid a blank
// well during that window. It is also the one file that changes when the
// backend adds size variants/LQIP - every LazyImage call site should read
// through here rather than touching `picture`/`pictureThumb` directly.

type PictureFields = { picture: string; pictureThumb: string }
type CardPictureFields = PictureFields & { pictureCard?: string; pictureLqip?: string }

// thumbSource: for grid cards and other small (~100-200px) previews.
export function thumbSource(entity: PictureFields): string | null {
  return entity.pictureThumb || entity.picture || null
}

// fullSource: for detail views, lightboxes, and anything rendered at or near
// the main rendition's own size.
export function fullSource(entity: PictureFields): string | null {
  return entity.picture || null
}

// cardSource: the ~512px rendition T1 added (bumped from an original 128px -
// see ObsidianVault/media-proxy-content-addressed.md - 128 was too small for
// the ~250-560px wells it actually renders into and looked blurry), for
// catalogue grid cells and other medium (~150-560px) wells: bigger than
// pictureThumb's 256px, smaller than the full ~1024px main rendition. Falls
// down the same ladder as thumbSource (thumb, then main) for entities/
// backends that haven't backfilled a card rendition yet - `pictureCard` is
// optional because avatar-bearing DTOs (PublicUserResponse etc.) don't carry
// it.
export function cardSource(entity: CardPictureFields): string | null {
  return entity.pictureCard || entity.pictureThumb || entity.picture || null
}

// lqipSource: the inline data: URI placeholder for LazyImage's `lqip` prop.
// Empty string means "not backfilled yet" (T1's DEFAULT '' migration), not a
// real value - normalize to null so LazyImage doesn't try to render it.
export function lqipSource(entity: { pictureLqip?: string }): string | null {
  return entity.pictureLqip || null
}

// focalPosition: the admin-chosen focal point (0..1, see the focal-point
// picker), converted to LazyImage's `contentPosition` prop - what a
// contentFit:'cover' well keeps centered instead of always cropping to the
// image's own center. Returns null for the default center (0.5, 0.5), which
// is exactly today's behaviour, so every call site can pass this
// unconditionally without special-casing entities that predate the focal
// point column.
export function focalPosition(
  entity: { focalX?: number; focalY?: number } | null | undefined
): { top: `${number}%`; left: `${number}%` } | null {
  const x = entity?.focalX ?? 0.5
  const y = entity?.focalY ?? 0.5
  if (x === 0.5 && y === 0.5) return null
  return { top: `${y * 100}%`, left: `${x * 100}%` }
}

// focalFromLocation: the inverse of focalPosition's math - turns a
// touch/click location (in the same coordinate space as the rect it landed
// in, e.g. RN's locationX/locationY) into a normalized 0..1 focal pair.
// Callers must measure the *displayed image's* rect, not a letterboxed well
// around it (see focal-point-picker.tsx) - dividing by the wrong rect gives
// a point that doesn't match where the user actually clicked. Clamped so a
// drag that overshoots the rect still lands at a valid edge instead of
// producing an out-of-range focal value the backend would reject. A
// zero-size rect (layout not measured yet) returns the center rather than
// dividing by zero.
export function focalFromLocation(
  locationX: number,
  locationY: number,
  width: number,
  height: number
): { x: number; y: number } {
  if (!width || !height) return { x: 0.5, y: 0.5 }
  const clamp01 = (v: number) => Math.min(1, Math.max(0, v))
  return { x: clamp01(locationX / width), y: clamp01(locationY / height) }
}

// imageRectForWell: the same "fit inside, preserve aspect ratio" math RN's
// `resizeMode="contain"` uses internally, exposed so the focal-point picker
// can measure the image's own *displayed* rect (letterboxed inside its
// well) instead of the well itself - the two only coincide when the well
// happens to share the image's aspect ratio exactly. Falls back to the well
// itself (no letterbox correction) when either size isn't known yet, so the
// gesture still works - just less precisely - until an onLoad measurement
// lands.
export function imageRectForWell(
  natural: { width: number; height: number } | null,
  well: { width: number; height: number }
): { width: number; height: number } {
  if (!natural || !well.width || !well.height) return well
  const scale = Math.min(well.width / natural.width, well.height / natural.height)
  return { width: natural.width * scale, height: natural.height * scale }
}
