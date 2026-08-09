package ports

import (
	"context"
	"io"
)

// Picture is the raw bytes and metadata of an image being stored.
type Picture struct {
	Content     io.Reader
	ContentType string
	Size        int64
	// PreferProvider pins Upload to a specific provider (e.g. "r2", "b2",
	// "supabase") instead of letting the storage fallback chain pick one.
	// Empty means "let the chain decide". The picture worker sets this on a
	// thumbnail upload to the provider its main rendition just landed on, so
	// a Stand/DevilFruit/avatar's two renditions never end up split across
	// two different buckets.
	PreferProvider string
}

// StoredPicture is where Upload actually put a picture - which provider it
// landed on (relevant when multiple are chained) and the key it was stored
// under (normally unchanged from what Upload was given).
type StoredPicture struct {
	Provider string
	Key      string
}

// IPictureStorage keeps object-storage details (which cloud, bucket layout,
// how URLs are signed) out of the domain. Keys are opaque to the domain:
// whatever Upload was given back is what a Stand persists in its picture
// field, and is later handed straight back to PresignGetURL/Delete.
type IPictureStorage interface {
	// Upload stores pic under key, overwriting any existing object there,
	// and reports which provider it ended up on.
	Upload(ctx context.Context, key string, pic Picture) (StoredPicture, error)
	// PresignGetURL returns a time-limited URL a client can GET the object
	// under key from, without the bucket having to be public.
	PresignGetURL(ctx context.Context, key string) (string, error)
	// Delete removes the object under key. Deleting a key that doesn't
	// exist is not an error.
	Delete(ctx context.Context, key string) error
}
