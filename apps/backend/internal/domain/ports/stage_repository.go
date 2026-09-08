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

// StagePageCursor is the decoded (manga, position, name) key a Stage page
// cursor carries - see IStageRepository.Page.
type StagePageCursor struct {
	Manga    enums.Manga
	Position int
	Name     string
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
	// Page returns up to limit+1 Stages matching filters, ordered by
	// (manga, position, name) after `after` (nil for the first page), then
	// the caller trims the extra row and reports hasMore. `after` carries all
	// three cursor fields together - a Stage page cursor is only ever issued
	// with all of them set, never partially.
	Page(ctx context.Context, filters StageFilters, locale enums.Locale, after *StagePageCursor, limit int) ([]game.Stage, bool, error)
	// Count returns the total number of Stages matching filters, ignoring
	// pagination.
	Count(ctx context.Context, filters StageFilters, locale enums.Locale) (int, error)
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
	UpdatePicture(ctx context.Context, id game.StageID, main, thumb, card, lqip *string, status enums.PictureStatus) error
	// SetMediaID updates only the content-addressed media group id - see
	// IStandRepository.SetMediaID.
	SetMediaID(ctx context.Context, id game.StageID, mediaID string) error
}
