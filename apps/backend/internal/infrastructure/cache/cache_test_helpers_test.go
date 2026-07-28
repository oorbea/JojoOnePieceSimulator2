package cache_test

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// fakeCache is an in-memory ports.ICache mirroring the real Redis adapter's
// generation-based invalidation semantics (without any TTL expiry - tests
// don't wait around for entries to expire), so decorator tests can exercise
// Invalidate/Delete without a real Redis instance.
type fakeCache struct {
	mu    sync.Mutex
	gen   map[string]int
	data  map[string]map[string][]byte // ns -> "<gen>:<key>" -> value
	calls int
	// invalidateErr, when set, makes the next Invalidate call fail - for
	// exercising the fail-open "log, don't fail the request" path.
	invalidateErr error
}

func newFakeCache() *fakeCache {
	return &fakeCache{gen: make(map[string]int), data: make(map[string]map[string][]byte)}
}

func (f *fakeCache) fullKey(ns, key string) string {
	return strconv.Itoa(f.gen[ns]) + ":" + key
}

func (f *fakeCache) Get(_ context.Context, ns, key string) ([]byte, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	m, ok := f.data[ns]
	if !ok {
		return nil, false
	}
	v, ok := m[f.fullKey(ns, key)]
	return v, ok
}

func (f *fakeCache) Set(_ context.Context, ns, key string, val []byte, _ time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.data[ns]; !ok {
		f.data[ns] = make(map[string][]byte)
	}
	f.data[ns][f.fullKey(ns, key)] = val
}

func (f *fakeCache) Delete(_ context.Context, ns, key string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if m, ok := f.data[ns]; ok {
		delete(m, f.fullKey(ns, key))
	}
}

func (f *fakeCache) Invalidate(_ context.Context, ns string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.invalidateErr != nil {
		err := f.invalidateErr
		f.invalidateErr = nil
		return err
	}
	f.gen[ns]++
	return nil
}

func (f *fakeCache) Close() error { return nil }

var _ ports.ICache = (*fakeCache)(nil)
