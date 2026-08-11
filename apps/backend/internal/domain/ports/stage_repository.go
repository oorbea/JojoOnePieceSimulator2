package ports

import (
	"context"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
)

// IStageRepository is the admin-facing CRUD counterpart to IStageCatalog -
// one adapter satisfies both, the same relationship IStandRepository has
// with the read side of the Stand catalogue.
type IStageRepository interface {
	// List returns every Stage, ordered by manga then position then name.
	List(ctx context.Context) ([]game.Stage, error)
	// FindByID returns the Stage matching id, or ErrStageNotFound.
	FindByID(ctx context.Context, id game.StageID) (game.Stage, error)
	// Save upserts s by ID - same convention as IStandRepository.Save: the
	// caller (the application layer, via IIdGenerator) already decided s's
	// ID, so Save can't distinguish create from update and doesn't try to.
	// Returns ErrStageAlreadyExists on a (manga, name) conflict with a
	// different Stage.
	Save(ctx context.Context, s game.Stage) error
	// Delete removes the Stage matching id. Returns ErrStageNotFound if
	// there is none.
	Delete(ctx context.Context, id game.StageID) error
}
