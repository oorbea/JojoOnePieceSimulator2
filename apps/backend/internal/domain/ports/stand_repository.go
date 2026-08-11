package ports

import (
	"context"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

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
	Delete(ctx context.Context, id powers.PowerID) error
	// UpdatePicture updates only a stand's picture renditions and pipeline
	// status. A nil main or thumb leaves that column untouched.
	UpdatePicture(ctx context.Context, id powers.PowerID, main, thumb *string, status enums.PictureStatus) error
	// Translations returns every locale's content for id, for admin edit
	// forms that need all locales at once instead of one resolved locale.
	Translations(ctx context.Context, id powers.PowerID) (PowerTranslations, error)
}
