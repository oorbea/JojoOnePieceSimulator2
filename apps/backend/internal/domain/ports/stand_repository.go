package ports

import (
	"context"
	"strings"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// StandOption is the id/name pair the evolvesFrom picker needs - no
// picture, no stats, no translations.
type StandOption struct {
	ID   powers.PowerID
	Name string
}

type StandFilters struct {
	Rarity      *enums.PowerRarity
	AttackPower *enums.StandStat
	Speed       *enums.StandStat
	AttackRange *enums.StandStat
	Endurance   *enums.StandStat
	Precision   *enums.StandStat
	Potential   *enums.StandStat
	EvolvesFrom *string
	// Search matches case-insensitively against name or the
	// locale-resolved description. Unescaped - callers must escape any
	// LIKE metacharacter (%, _, \) before this reaches SQL.
	Search *string
}

// Canonical renders every field in a fixed order, joined by "|", so two
// requests differing only in query-param order produce the identical
// string - the single source of truth both cache/keys.go's standFilterKey
// (hashed, for the cache key) and dto/pagination.go's
// StandFiltersFingerprint (hashed again with locale, for the pagination
// cursor) build on. Before this method existed, both call sites duplicated
// this field list separately, and a field added to StandFilters without
// updating both was a silent correctness gap - see
// ObsidianVault/catalogue-pagination.md. A field added to the struct now
// only has one place to add it here; canonical_test.go's field-count
// assertion fails loudly if this method itself falls out of sync with the
// struct instead.
func (f StandFilters) Canonical() string {
	return strings.Join([]string{
		optStringer(f.Rarity),
		optStringer(f.AttackPower),
		optStringer(f.Speed),
		optStringer(f.AttackRange),
		optStringer(f.Endurance),
		optStringer(f.Precision),
		optStringer(f.Potential),
		optString(f.EvolvesFrom),
		optString(f.Search),
	}, "|")
}

type IStandRepository interface {
	// Save upserts the given stand's powers/stands rows, then replaces its
	// power_translations rows with translations (en-GB mandatory, es-ES/ca-ES
	// optional) - any locale missing from translations is deleted.
	Save(ctx context.Context, stand *powers.Stand, translations PowerTranslations) error
	// FindByID/FindByName/GetAll/Filter resolve description/skills for
	// locale, falling back through enums.FallbackChain(locale) down to
	// en-GB.
	FindByID(ctx context.Context, id powers.PowerID, locale enums.Locale) (*powers.Stand, error)
	FindByName(ctx context.Context, name string, locale enums.Locale) (*powers.Stand, error)
	GetAll(ctx context.Context, locale enums.Locale) ([]*powers.Stand, error)
	Filter(ctx context.Context, filters StandFilters, locale enums.Locale) ([]*powers.Stand, error)
	// Page returns up to limit stands matching filters, ordered by name,
	// strictly after afterName (nil means "from the start"). hasMore reports
	// whether more rows exist beyond the returned page - callers ask for
	// limit+1 rows and this reports true iff that extra row came back.
	Page(ctx context.Context, filters StandFilters, locale enums.Locale, afterName *string, limit int) (stands []*powers.Stand, hasMore bool, err error)
	// Count returns the total number of stands matching filters, ignoring
	// pagination - used for a page response's `total`, typically only on the
	// first page.
	Count(ctx context.Context, filters StandFilters, locale enums.Locale) (int, error)
	// Options returns every stand's id/name only, unfiltered and
	// locale-free (powers.name is not translatable) - backs the
	// evolvesFrom picker without the cost of a full catalogue fetch.
	Options(ctx context.Context) ([]StandOption, error)
	Delete(ctx context.Context, id powers.PowerID) error
	// UpdatePicture updates only a stand's picture renditions and pipeline
	// status. A nil main/thumb/card/lqip leaves that column untouched.
	UpdatePicture(ctx context.Context, id powers.PowerID, main, thumb, card, lqip *string, status enums.PictureStatus) error
	// SetMediaID updates only the content-addressed media group id, once the
	// worker has persisted the corresponding media_objects rows - see
	// powers.Power.SetMediaID's doc for why this is separate from
	// UpdatePicture.
	SetMediaID(ctx context.Context, id powers.PowerID, mediaID string) error
	// Translations returns every locale's content for id, for admin edit
	// forms that need all locales at once instead of one resolved locale.
	Translations(ctx context.Context, id powers.PowerID) (PowerTranslations, error)
}
