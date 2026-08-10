package ports

import (
	"context"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
)

// IGameHistory records a finished or aborted Game's outcome. There is no
// adapter yet - Games otherwise live only in Redis for the duration of the
// lobby/match and disappear once its TTL expires.
type IGameHistory interface {
	Record(ctx context.Context, result game.GameResult) error
}
