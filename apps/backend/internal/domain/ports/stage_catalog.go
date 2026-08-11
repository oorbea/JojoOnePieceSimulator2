package ports

import (
	"context"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// IStageCatalog serves the round content (JoJo parts, One Piece sagas) a
// Game plays through - backed by the stages table (see IStageRepository for
// the admin-facing CRUD over the same data).
type IStageCatalog interface {
	// Stages returns every Stage for manga, ordered by Stage.Order().
	// Gauntlet feeds the result of every selected manga into
	// game.Interleave to build its full round order up front; Versus
	// treats it as the pool game.IGameMode.StageFor draws a random Stage
	// from each round.
	Stages(ctx context.Context, manga enums.Manga) ([]game.Stage, error)
}
