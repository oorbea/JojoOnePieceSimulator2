package cache_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	infracache "github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/cache"
)

// countingPictureStorage counts PresignGetURL calls and returns a distinct
// URL each time, so tests can tell a cache hit (same URL twice) from a
// fresh presign (different URL, incremented call count).
type countingPictureStorage struct {
	mu           sync.Mutex
	presignCalls int
	deletedKeys  []string
}

func (s *countingPictureStorage) Upload(context.Context, string, ports.Picture) error { return nil }

func (s *countingPictureStorage) PresignGetURL(_ context.Context, key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.presignCalls++
	return fmt.Sprintf("https://r2.test/%s?call=%d", key, s.presignCalls), nil
}

func (s *countingPictureStorage) Delete(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deletedKeys = append(s.deletedKeys, key)
	return nil
}

var _ ports.IPictureStorage = (*countingPictureStorage)(nil)

func TestPictureStorage_PresignGetURL_CachesOnMiss(t *testing.T) {
	next := &countingPictureStorage{}
	storage := infracache.NewPictureStorage(next, newFakeCache(), time.Minute)
	ctx := context.Background()

	first, err := storage.PresignGetURL(ctx, "stands/1/main.webp")
	if err != nil {
		t.Fatalf("first PresignGetURL: %v", err)
	}
	second, err := storage.PresignGetURL(ctx, "stands/1/main.webp")
	if err != nil {
		t.Fatalf("second PresignGetURL: %v", err)
	}

	if first != second {
		t.Errorf("PresignGetURL: first = %q, second = %q, want identical (second should be a cache hit)", first, second)
	}
	if next.presignCalls != 1 {
		t.Errorf("underlying PresignGetURL calls = %d, want 1", next.presignCalls)
	}
}

func TestPictureStorage_PresignGetURL_DifferentKeysDontCollide(t *testing.T) {
	next := &countingPictureStorage{}
	storage := infracache.NewPictureStorage(next, newFakeCache(), time.Minute)
	ctx := context.Background()

	a, err := storage.PresignGetURL(ctx, "stands/1/main.webp")
	if err != nil {
		t.Fatalf("PresignGetURL a: %v", err)
	}
	b, err := storage.PresignGetURL(ctx, "stands/2/main.webp")
	if err != nil {
		t.Fatalf("PresignGetURL b: %v", err)
	}
	if a == b {
		t.Errorf("distinct keys produced the same cached URL: %q", a)
	}
	if next.presignCalls != 2 {
		t.Errorf("underlying PresignGetURL calls = %d, want 2", next.presignCalls)
	}
}

func TestPictureStorage_Delete_EvictsCachedPresign(t *testing.T) {
	next := &countingPictureStorage{}
	storage := infracache.NewPictureStorage(next, newFakeCache(), time.Minute)
	ctx := context.Background()

	if _, err := storage.PresignGetURL(ctx, "stands/1/main.webp"); err != nil {
		t.Fatalf("PresignGetURL: %v", err)
	}
	if err := storage.Delete(ctx, "stands/1/main.webp"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if _, err := storage.PresignGetURL(ctx, "stands/1/main.webp"); err != nil {
		t.Fatalf("PresignGetURL after Delete: %v", err)
	}
	if next.presignCalls != 2 {
		t.Errorf("underlying PresignGetURL calls after Delete = %d, want 2 (Delete should evict the cached URL)", next.presignCalls)
	}
	if len(next.deletedKeys) != 1 || next.deletedKeys[0] != "stands/1/main.webp" {
		t.Errorf("deletedKeys = %v, want [stands/1/main.webp]", next.deletedKeys)
	}
}
