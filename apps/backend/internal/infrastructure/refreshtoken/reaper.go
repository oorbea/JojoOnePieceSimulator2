package refreshtoken

import (
	"context"
	"log"
	"time"
)

// Reaper periodically removes expired refresh tokens from a MemoryStore so
// a burst of minted-but-never-redeemed tokens doesn't sit in the map until
// process exit - mirrors gamestore.Reaper's Start/ReapOnce split. Only
// meaningful for MemoryStore: the Redis store expires its own keys.
type Reaper struct {
	store    *MemoryStore
	interval time.Duration
}

// NewReaper builds a Reaper over store. interval <= 0 disables it (Start
// becomes a no-op).
func NewReaper(store *MemoryStore, interval time.Duration) *Reaper {
	return &Reaper{store: store, interval: interval}
}

// Start runs ReapOnce every interval until ctx is done. It blocks, so
// callers should run it in its own goroutine.
func (r *Reaper) Start(ctx context.Context) {
	if r.interval <= 0 {
		return
	}
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if n := r.ReapOnce(); n > 0 {
				log.Printf("refresh token reaper: removed %d expired token(s)", n)
			}
		}
	}
}

// ReapOnce removes every expired refresh token, returning how many were
// removed. Its own method so callers (and tests) can trigger one pass
// without waiting on Start's ticker.
func (r *Reaper) ReapOnce() int {
	return r.store.DeleteExpired()
}
