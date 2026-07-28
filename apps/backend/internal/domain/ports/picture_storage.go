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
}

// IPictureStorage keeps object-storage details (which cloud, bucket layout,
// how URLs are signed) out of the domain. Keys are opaque to the domain:
// whatever Upload was given back is what a Stand persists in its picture
// field, and is later handed straight back to PresignGetURL/Delete.
type IPictureStorage interface {
	// Upload stores pic under key, overwriting any existing object there.
	Upload(ctx context.Context, key string, pic Picture) error
	// PresignGetURL returns a time-limited URL a client can GET the object
	// under key from, without the bucket having to be public.
	PresignGetURL(ctx context.Context, key string) (string, error)
	// Delete removes the object under key. Deleting a key that doesn't
	// exist is not an error.
	Delete(ctx context.Context, key string) error
}
