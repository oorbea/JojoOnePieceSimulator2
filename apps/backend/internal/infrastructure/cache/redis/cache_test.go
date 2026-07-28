package redis_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/cache/redis"
)

// newTestCache connects to TEST_REDIS_URL, skipping the test entirely when
// it is unset - so `make test` stays runnable with no Redis instance around,
// and only `make test` run with TEST_REDIS_URL set (e.g. after `make db-up`)
// exercises the real adapter.
func newTestCache(t *testing.T) *redis.Cache {
	t.Helper()
	url := os.Getenv("TEST_REDIS_URL")
	if url == "" {
		t.Skip("TEST_REDIS_URL not set, skipping redis-backed cache test")
	}
	c, err := redis.New(context.Background(), redis.Config{
		URL: url, DialTimeout: 2 * time.Second, OpTimeout: time.Second,
	})
	if err != nil {
		t.Fatalf("connecting to redis: %v", err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c
}

func TestCache_SetGet_RoundTrips(t *testing.T) {
	c := newTestCache(t)
	ctx := context.Background()
	ns := "test-roundtrip"
	defer func() { _ = c.Invalidate(ctx, ns) }()

	c.Set(ctx, ns, "key1", []byte("value1"), time.Minute)

	got, ok := c.Get(ctx, ns, "key1")
	if !ok {
		t.Fatal("Get after Set: ok = false, want true")
	}
	if string(got) != "value1" {
		t.Errorf("Get = %q, want %q", got, "value1")
	}
}

func TestCache_Get_MissReturnsFalse(t *testing.T) {
	c := newTestCache(t)
	ctx := context.Background()

	_, ok := c.Get(ctx, "test-miss", "no-such-key")
	if ok {
		t.Error("Get for an unset key: ok = true, want false")
	}
}

func TestCache_Invalidate_OrphansPreviousEntries(t *testing.T) {
	c := newTestCache(t)
	ctx := context.Background()
	ns := "test-invalidate"
	defer func() { _ = c.Invalidate(ctx, ns) }()

	c.Set(ctx, ns, "key1", []byte("value1"), time.Minute)
	if err := c.Invalidate(ctx, ns); err != nil {
		t.Fatalf("Invalidate: %v", err)
	}

	_, ok := c.Get(ctx, ns, "key1")
	if ok {
		t.Error("Get after Invalidate: ok = true, want false (orphaned by generation bump)")
	}

	// A value written after the invalidation is visible again.
	c.Set(ctx, ns, "key2", []byte("value2"), time.Minute)
	got, ok := c.Get(ctx, ns, "key2")
	if !ok || string(got) != "value2" {
		t.Errorf("Get after re-Set post-invalidate = %q, %v, want %q, true", got, ok, "value2")
	}
}

func TestCache_Delete_RemovesSingleKey(t *testing.T) {
	c := newTestCache(t)
	ctx := context.Background()
	ns := "test-delete"
	defer func() { _ = c.Invalidate(ctx, ns) }()

	c.Set(ctx, ns, "key1", []byte("value1"), time.Minute)
	c.Set(ctx, ns, "key2", []byte("value2"), time.Minute)

	c.Delete(ctx, ns, "key1")

	if _, ok := c.Get(ctx, ns, "key1"); ok {
		t.Error("Get for a deleted key: ok = true, want false")
	}
	if got, ok := c.Get(ctx, ns, "key2"); !ok || string(got) != "value2" {
		t.Errorf("Get for an untouched key = %q, %v, want %q, true", got, ok, "value2")
	}
}

func TestCache_Namespaces_DontCollide(t *testing.T) {
	c := newTestCache(t)
	ctx := context.Background()
	defer func() {
		_ = c.Invalidate(ctx, "test-ns-a")
		_ = c.Invalidate(ctx, "test-ns-b")
	}()

	c.Set(ctx, "test-ns-a", "key", []byte("a-value"), time.Minute)
	c.Set(ctx, "test-ns-b", "key", []byte("b-value"), time.Minute)

	if err := c.Invalidate(ctx, "test-ns-a"); err != nil {
		t.Fatalf("Invalidate ns-a: %v", err)
	}

	if _, ok := c.Get(ctx, "test-ns-a", "key"); ok {
		t.Error("test-ns-a: entry survived its own namespace's Invalidate")
	}
	if got, ok := c.Get(ctx, "test-ns-b", "key"); !ok || string(got) != "b-value" {
		t.Errorf("test-ns-b: entry did not survive test-ns-a's Invalidate, got %q, %v", got, ok)
	}
}
