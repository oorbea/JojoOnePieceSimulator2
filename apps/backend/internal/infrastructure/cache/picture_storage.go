package cache

import (
	"context"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// PictureStorage decorates a ports.IPictureStorage with a ports.ICache,
// caching presigned GET URLs per object key so repeat reads (and, in turn,
// the response-body ETag layer built on top of them) stay byte-stable
// within presignTTL instead of computing a fresh SigV4 signature every time.
type PictureStorage struct {
	next       ports.IPictureStorage
	cache      ports.ICache
	presignTTL time.Duration
}

var _ ports.IPictureStorage = (*PictureStorage)(nil)

// NewPictureStorage wraps next. presignTTL must be safely shorter than the
// underlying presign's own expiry (config.Load enforces this), so a served
// URL is never handed out close to expiring.
func NewPictureStorage(next ports.IPictureStorage, c ports.ICache, presignTTL time.Duration) *PictureStorage {
	return &PictureStorage{next: next, cache: c, presignTTL: presignTTL}
}

// Upload passes through untouched - caching only applies to reads.
func (s *PictureStorage) Upload(ctx context.Context, key string, pic ports.Picture) error {
	return s.next.Upload(ctx, key, pic)
}

// PresignGetURL is read-through, cached per object key.
func (s *PictureStorage) PresignGetURL(ctx context.Context, key string) (string, error) {
	if data, ok := s.cache.Get(ctx, presignNamespace, key); ok {
		return string(data), nil
	}

	url, err := s.next.PresignGetURL(ctx, key)
	if err != nil {
		return "", err
	}

	s.cache.Set(ctx, presignNamespace, key, []byte(url), s.presignTTL)
	return url, nil
}

// Delete removes the object, then evicts any cached presigned URL for it -
// otherwise a URL for an object the picture worker has just superseded (see
// PictureWorker's cleanup of replaced renditions) could still be served
// until presignTTL.
func (s *PictureStorage) Delete(ctx context.Context, key string) error {
	if err := s.next.Delete(ctx, key); err != nil {
		return err
	}
	s.cache.Delete(ctx, presignNamespace, key)
	return nil
}
