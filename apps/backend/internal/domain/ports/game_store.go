package ports

import (
	"context"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
)

// IGameStore holds every live Game aggregate for the duration of its
// lobby/match, indexed both by GameID and by its short join code. There is
// no schema for it yet - an in-memory adapter is enough for a single
// backend instance; a Redis-backed adapter behind this same port is the
// next tanda (see ObsidianVault/ADR.md), so callers must not assume
// anything beyond what this interface promises (in particular: no
// snapshot/rehydration API, since the in-memory adapter needs none).
type IGameStore interface {
	// Create indexes g under both its own GameID and code. Returns
	// ErrGameCodeTaken if code is already indexed by a different Game.
	Create(ctx context.Context, code string, g *game.Game) error
	// Get returns the Game identified by id, or ErrGameNotFound.
	Get(ctx context.Context, id game.GameID) (*game.Game, error)
	// GetByCode returns the Game currently indexed under code, or
	// ErrGameNotFound.
	GetByCode(ctx context.Context, code string) (*game.Game, error)
	// Code returns the join code currently indexed for id, or
	// ErrGameNotFound.
	Code(ctx context.Context, id game.GameID) (string, error)
	// Save persists g's current state and refreshes its TTL. g must have
	// already been Create'd.
	Save(ctx context.Context, g *game.Game) error
	// Delete removes id (and its code index) entirely.
	Delete(ctx context.Context, id game.GameID) error
	// DeleteExpired removes every Game last saved more than olderThan ago,
	// returning how many were removed. Called periodically by a reaper so
	// an abandoned lobby does not linger forever.
	DeleteExpired(ctx context.Context, olderThan time.Duration) int
}
