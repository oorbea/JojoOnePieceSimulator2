package ports

import (
	"context"
	"io"
)

// ObjectInfo is the metadata Get returns alongside an object's bytes.
type ObjectInfo struct {
	ContentType string
	Size        int64
}

// IStorageBackend is one object-storage provider (Cloudflare R2, Backblaze
// B2, Supabase Storage, ...). It is deliberately narrower and more
// mechanical than IPictureStorage: no quota/fallback decisions, no Picture
// wrapper, just the raw operations a single bucket can do.
// infrastructure/storage/s3store.Backend is the S3-API-compatible
// implementation every provider normally uses - only endpoint/region
// differ, never behavior. infrastructure/storage/workerproxy.Backend is a
// second implementation, used only for R2 when R2_WORKER_URL is set, that
// fronts the bucket through a Cloudflare Worker instead of R2's S3 API
// endpoint (see that package's doc comment for why).
type IStorageBackend interface {
	// Name identifies this backend for the ledger and config (e.g. "r2",
	// "b2", "supabase").
	Name() string
	// Put uploads content under key. content must have exactly size bytes
	// left to read.
	Put(ctx context.Context, key string, content io.Reader, contentType string, size int64) error
	// Get downloads key's full bytes. Returns ErrObjectNotFound if key
	// doesn't exist. The caller must Close the returned ReadCloser.
	Get(ctx context.Context, key string) (io.ReadCloser, ObjectInfo, error)
	// PresignGet returns a time-limited GET URL for key.
	PresignGet(ctx context.Context, key string) (string, error)
	// Del deletes key. Deleting a key that doesn't exist is not an error.
	Del(ctx context.Context, key string) error
	// Walk calls fn once per object currently in the bucket, with its key
	// and size in bytes. Used only by the reconciler.
	Walk(ctx context.Context, fn func(key string, bytes int64) error) error
}
