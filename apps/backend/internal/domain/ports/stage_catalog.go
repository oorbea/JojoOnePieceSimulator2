package ports

import (
	"context"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// IStageCatalog serves the round content (JoJo parts, One Piece sagas) a
// Game plays through - backed by the stages table (see IStageRepository for
// the admin-facing CRUD over the same data).
//
// Deliberately no locale parameter: the gameplay engine (game.Interleave,
// IGameMode.StageFor) only ever reads Name/Manga/Order, never Description,
// and a live Game is a single instance shared by every participant, so a
// Stage frozen into a Round can't carry an already-resolved description for
// each viewer anyway - see api/dto.NewGameStateResponse, which re-resolves
// each Stage's description per viewer's own configured language at
// serialize time instead. The adapter resolves Description at a fixed
// enums.EnGB here purely so the returned Stage value satisfies its own
// non-empty invariant; nothing on this path ever reads it.
type IStageCatalog interface {
	// Stages returns every Stage for manga, ordered by Stage.Order().
	// Gauntlet feeds the result of every selected manga into
	// game.Interleave to build its full round order up front; Versus
	// treats it as the pool game.IGameMode.StageFor draws a random Stage
	// from each round.
	Stages(ctx context.Context, manga enums.Manga) ([]game.Stage, error)
}
