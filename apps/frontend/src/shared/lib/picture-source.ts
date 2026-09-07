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

// thumbSource: for grid cards and other small (~100-200px) previews.
export function thumbSource(entity: PictureFields): string | null {
  return entity.pictureThumb || entity.picture || null
}

// fullSource: for detail views, lightboxes, and anything rendered at or near
// the main rendition's own size.
export function fullSource(entity: PictureFields): string | null {
  return entity.picture || null
}
