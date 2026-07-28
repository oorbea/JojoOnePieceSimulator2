package ports

import (
	"context"
	"time"
)

// ICache is a generic, namespaced byte-cache port. Namespaces let unrelated
// callers (e.g. the Stand repository cache and the picture presign cache)
// share one backend without their keys colliding, and let a whole namespace
// be dropped at once without enumerating its keys.
type ICache interface {
	// Get returns the cached bytes for key inside namespace ns. The bool is
	// false on a miss AND on any backend failure - callers must treat both
	// identically and fall through to the source of truth.
	Get(ctx context.Context, ns, key string) ([]byte, bool)
	// Set stores val under ns/key with the given ttl. Failures are logged,
	// never returned: a cache write is never worth failing a request over.
	Set(ctx context.Context, ns, key string, val []byte, ttl time.Duration)
	// Delete evicts a single ns/key.
	Delete(ctx context.Context, ns, key string)
	// Invalidate drops every entry currently in ns at once. Returns an error
	// so the caller can log loudly - a failed invalidation means stale reads
	// until the affected entries' TTL expires on their own.
	Invalidate(ctx context.Context, ns string) error
	// Close releases the underlying connection(s).
	Close() error
}
