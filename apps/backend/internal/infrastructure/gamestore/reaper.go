package gamestore

import (
	"context"
	"log"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// Reaper periodically removes lobbies/matches nobody has touched in a
// while, so an abandoned Game (browser closed mid-lobby, host never
// started it) does not linger in the store forever - mirrors
// services.StorageReconciler's Start/RunOnce split.
type Reaper struct {
	store    ports.IGameStore
	ttl      time.Duration
	interval time.Duration
}

// NewReaper builds a Reaper over store. interval <= 0 disables it (Start
// becomes a no-op).
func NewReaper(store ports.IGameStore, ttl, interval time.Duration) *Reaper {
	return &Reaper{store: store, ttl: ttl, interval: interval}
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
			if n := r.ReapOnce(ctx); n > 0 {
				log.Printf("game store reaper: removed %d expired game(s)", n)
			}
		}
	}
}

// ReapOnce removes every Game last saved more than ttl ago, returning how
// many were removed. Exists as its own method so callers (and tests) can
// trigger one pass without waiting on Start's ticker.
func (r *Reaper) ReapOnce(ctx context.Context) int {
	return r.store.DeleteExpired(ctx, r.ttl)
}
