package ports

import (
	"context"
	"strings"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/characters"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// OnePieceCharacterFilters mirrors JojoCharacterFilters' shape for the One
// Piece side - see that type's doc for why the two are kept separate.
type OnePieceCharacterFilters struct {
	Rarity          *enums.PowerRarity
	PhysicalForm    *enums.PhysicalForm
	ArmamentHaki    *enums.HakiLevel
	ObservationHaki *enums.HakiLevel
	ConquerorHaki   *enums.HakiLevel
	FruitMastery    *enums.FruitMastery
	// Search matches case-insensitively against name or the
	// locale-resolved description. Unescaped - callers must escape any
	// LIKE metacharacter (%, _, \) before this reaches SQL.
	Search *string
}

// Canonical - see StandFilters.Canonical's doc for why this exists and what
// it's the single source of truth for.
func (f OnePieceCharacterFilters) Canonical() string {
	return strings.Join([]string{
		optStringer(f.Rarity),
		optStringer(f.PhysicalForm),
		optStringer(f.ArmamentHaki),
		optStringer(f.ObservationHaki),
		optStringer(f.ConquerorHaki),
		optStringer(f.FruitMastery),
		optString(f.Search),
	}, "|")
}

// IOnePieceCharacterRepository is the admin-facing CRUD for One Piece
// characters - same shape as IJojoCharacterRepository, see that type's
// method docs.
type IOnePieceCharacterRepository interface {
	Save(ctx context.Context, c *characters.OnePieceCharacter, translations CharacterTranslations) error
	FindByID(ctx context.Context, id characters.CharacterID, locale enums.Locale) (*characters.OnePieceCharacter, error)
	FindByName(ctx context.Context, name string, locale enums.Locale) (*characters.OnePieceCharacter, error)
	GetAll(ctx context.Context, locale enums.Locale) ([]*characters.OnePieceCharacter, error)
	Filter(ctx context.Context, filters OnePieceCharacterFilters, locale enums.Locale) ([]*characters.OnePieceCharacter, error)
	Page(ctx context.Context, filters OnePieceCharacterFilters, locale enums.Locale, afterName *string, limit int) ([]*characters.OnePieceCharacter, bool, error)
	Count(ctx context.Context, filters OnePieceCharacterFilters, locale enums.Locale) (int, error)
	Delete(ctx context.Context, id characters.CharacterID) error
	UpdatePicture(ctx context.Context, id characters.CharacterID, main, thumb, card, lqip *string, status enums.PictureStatus) error
	SetMediaID(ctx context.Context, id characters.CharacterID, mediaID string) error
	Translations(ctx context.Context, id characters.CharacterID) (CharacterTranslations, error)
}
