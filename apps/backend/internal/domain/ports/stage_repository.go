package ports

import (
	"context"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// StageFilters mirrors StandFilters/DevilFruitFilters' shape - every field
// is optional, Filter applies whichever are set.
type StageFilters struct {
	Manga *enums.Manga
	// Search matches case-insensitively against name or the
	// locale-resolved description. Unescaped - callers must escape any
	// LIKE metacharacter (%, _, \) before this reaches SQL.
	Search *string
}

// IStageRepository is the admin-facing CRUD counterpart to IStageCatalog -
// one adapter satisfies both, the same relationship IStandRepository has
// with the read side of the Stand catalogue.
type IStageRepository interface {
	// List returns every Stage, ordered by manga then position then name,
	// description resolved for locale.
	List(ctx context.Context, locale enums.Locale) ([]game.Stage, error)
	// Filter returns every Stage matching the (all-optional) filters,
	// ordered by manga then position, description resolved for locale -
	// the admin-facing, locale-aware counterpart to IStageCatalog.Stages
	// (which is gameplay-facing and always resolves at a fixed
	// enums.EnGB - see that port's doc).
	Filter(ctx context.Context, filters StageFilters, locale enums.Locale) ([]game.Stage, error)
	// FindByID returns the Stage matching id, description resolved for
	// locale, or ErrStageNotFound.
	FindByID(ctx context.Context, id game.StageID, locale enums.Locale) (game.Stage, error)
	// Save upserts s by ID - same convention as IStandRepository.Save: the
	// caller (the application layer, via IIdGenerator) already decided s's
	// ID, so Save can't distinguish create from update and doesn't try to.
	// translations replaces stage_translations wholesale, same semantics as
	// IStandRepository.Save's translations parameter. Returns
	// ErrStageAlreadyExists on a (manga, name) conflict with a different
	// Stage.
	Save(ctx context.Context, s game.Stage, translations StageTranslations) error
	// Delete removes the Stage matching id. Returns ErrStageNotFound if
	// there is none.
	Delete(ctx context.Context, id game.StageID) error
	// Translations returns every locale's content for id, for the admin
	// edit form.
	Translations(ctx context.Context, id game.StageID) (StageTranslations, error)
	// UpdatePicture updates only a Stage's picture renditions and pipeline
	// status - same contract as IStandRepository.UpdatePicture.
	UpdatePicture(ctx context.Context, id game.StageID, main, thumb *string, status enums.PictureStatus) error
}
