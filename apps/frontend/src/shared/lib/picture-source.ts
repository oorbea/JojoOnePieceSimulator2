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

// cardSource: the ~128px rendition T1 added, for catalogue grid cells
// (~140px wells) - smaller than pictureThumb's 256px, so it's what LazyImage
// should actually fetch there. Falls down the same ladder as thumbSource
// (thumb, then main) for entities/backends that haven't backfilled a card
// rendition yet - `pictureCard` is optional because avatar-bearing DTOs
// (PublicUserResponse etc.) don't carry it.
export function cardSource(entity: CardPictureFields): string | null {
  return entity.pictureCard || entity.pictureThumb || entity.picture || null
}

// lqipSource: the inline data: URI placeholder for LazyImage's `lqip` prop.
// Empty string means "not backfilled yet" (T1's DEFAULT '' migration), not a
// real value - normalize to null so LazyImage doesn't try to render it.
export function lqipSource(entity: { pictureLqip?: string }): string | null {
  return entity.pictureLqip || null
}
