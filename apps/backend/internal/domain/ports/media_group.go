package ports

import (
	"crypto/sha256"
	"encoding/hex"
)

// MediaRenditionProfile is folded into every content-addressed group id
// (see MediaGroupID). Public/private URLs under the media proxy are
// immutable forever (Cache-Control: immutable, cached by browsers, the
// service worker, and the backend's on-disk LRU) - so re-uploading
// byte-identical content after a transcode-settings change (a rendition's
// MaxDimension/Quality) MUST still resolve to a new group, or every one of
// those caches keeps serving the stale bytes indefinitely. Bump this string
// any time PICTURE_*_DIMENSION/PICTURE_*_QUALITY (or the set of variants
// itself) changes in a way that alters what a given group's renditions look
// like. See ObsidianVault/media-proxy-content-addressed.md.
const MediaRenditionProfile = "v2"

// MediaGroupID computes the content-addressed group id for a main
// rendition's bytes: hex(sha256(salt || MediaRenditionProfile || mainBytes))[:16]
// (32 hex chars) - short enough for a clean URL path segment, long enough
// that guessing a valid group id is infeasible. Same source bytes + same
// rendition profile always produce the same group (free dedup on
// re-upload); either one changing produces a brand-new group. Shared by
// application/services.PictureWorker (the live transcode path) and
// cmd/mediabackfill (the one-shot backfill) so the two can never drift.
func MediaGroupID(salt string, mainBytes []byte) string {
	h := sha256.New()
	h.Write([]byte(salt))
	h.Write([]byte(MediaRenditionProfile))
	h.Write(mainBytes)
	sum := h.Sum(nil)
	return hex.EncodeToString(sum[:16])
}
