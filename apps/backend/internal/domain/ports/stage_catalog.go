package ports

import (
	"context"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// IStageCatalog serves the round content (JoJo parts, One Piece sagas) a
// Game plays through. There is no schema for it yet - stages are meant to
// be admin-managed CRUD data, added once that adapter exists.
type IStageCatalog interface {
	// Stages returns every Stage for manga, ordered by Stage.Order().
	// Gauntlet feeds the result of every selected manga into
	// game.Interleave to build its full round order up front; Versus
	// treats it as the pool game.IGameMode.StageFor draws a random Stage
	// from each round.
	Stages(ctx context.Context, manga enums.Manga) ([]game.Stage, error)
}
