package ports

import (
	"context"
	"strings"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

type DevilFruitFilters struct {
	Rarity    *enums.PowerRarity
	FruitType *enums.FruitType
	// Search matches case-insensitively against name or the
	// locale-resolved description. Unescaped - callers must escape any
	// LIKE metacharacter (%, _, \) before this reaches SQL.
	Search *string
}

// Canonical - see StandFilters.Canonical's doc for why this exists and what
// it's the single source of truth for.
func (f DevilFruitFilters) Canonical() string {
	return strings.Join([]string{
		optStringer(f.Rarity),
		optStringer(f.FruitType),
		optString(f.Search),
	}, "|")
}

type IDevilFruitRepository interface {
	// Save upserts the given fruit's powers/devil_fruits rows, then replaces
	// its power_translations rows with translations (en-GB mandatory,
	// es-ES/ca-ES optional) - any locale missing from translations is
	// deleted.
	Save(ctx context.Context, fruit *powers.DevilFruit, translations PowerTranslations) error
	// FindByID/FindByName/GetAll/Filter resolve description/skills for
	// locale, falling back through enums.FallbackChain(locale) down to
	// en-GB.
	FindByID(ctx context.Context, id powers.PowerID, locale enums.Locale) (*powers.DevilFruit, error)
	FindByName(ctx context.Context, name string, locale enums.Locale) (*powers.DevilFruit, error)
	GetAll(ctx context.Context, locale enums.Locale) ([]*powers.DevilFruit, error)
	Filter(ctx context.Context, filters DevilFruitFilters, locale enums.Locale) ([]*powers.DevilFruit, error)
	// Page returns up to limit+1 devil fruits matching filters, ordered by
	// name after afterName, then the caller trims the extra row and reports
	// hasMore - same contract as IStandRepository.Page, minus the
	// ancestor-truncation concern (DevilFruit has no evolves_from chain).
	Page(ctx context.Context, filters DevilFruitFilters, locale enums.Locale, afterName *string, limit int) ([]*powers.DevilFruit, bool, error)
	// Count returns the total number of devil fruits matching filters,
	// ignoring pagination.
	Count(ctx context.Context, filters DevilFruitFilters, locale enums.Locale) (int, error)
	Delete(ctx context.Context, id powers.PowerID) error
	// UpdatePicture updates only a devil fruit's picture renditions and
	// pipeline status. A nil main/thumb/card/lqip leaves that column untouched.
	UpdatePicture(ctx context.Context, id powers.PowerID, main, thumb, card, lqip *string, status enums.PictureStatus) error
	// SetMediaID updates only the content-addressed media group id - see
	// IStandRepository.SetMediaID.
	SetMediaID(ctx context.Context, id powers.PowerID, mediaID string) error
	// Translations returns every locale's content for id, for admin edit
	// forms that need all locales at once instead of one resolved locale.
	Translations(ctx context.Context, id powers.PowerID) (PowerTranslations, error)
}
