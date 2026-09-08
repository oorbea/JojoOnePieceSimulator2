package ports

import "context"

// MediaObject is one variant's row in media_objects - the content-addressed
// index the immutable media proxy reads (see infrastructure/api/endpoints
// media_endpoints.go). GroupID is derived from the main rendition's bytes
// (see dto.MediaURLBuilder's doc), so the same source image always
// resolves to the same group regardless of how many times it's re-uploaded.
type MediaObject struct {
	GroupID     string
	Variant     string
	StorageKey  string
	ContentType string
	Bytes       int64
	// Scope is "public" (Stand/DevilFruit/Stage - readable by anyone who
	// knows the group id, no signature required) or "private" (User
	// avatars - require a valid quantized-exp signature, see dto.MediaURLBuilder.Private).
	Scope string
}

// IMediaRepository persists/reads media_objects rows. Kept separate from
// IPictureStorage (which moves bytes) and from each subject's own repository
// (PutMediaObjects/GetMediaObjectsByGroup are keyed by content hash, not by
// subject id, and are shared across all four picture-bearing subjects).
type IMediaRepository interface {
	// PutMediaObjects upserts one or more variant rows, normally all of a
	// single group's card/thumb/main written together after a transcode.
	PutMediaObjects(ctx context.Context, objects []MediaObject) error
	// GetMediaObjectsByGroup returns every variant row for groupID, for the
	// media handler's card->thumb->main fallback ladder. Returns an empty
	// slice (not an error) if groupID is unknown.
	GetMediaObjectsByGroup(ctx context.Context, groupID string) ([]MediaObject, error)
}
