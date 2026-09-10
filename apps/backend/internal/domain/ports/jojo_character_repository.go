package ports

import (
	"context"
	"strconv"
	"strings"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/characters"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// JojoCharacterFilters mirrors StandFilters/DevilFruitFilters' shape -
// every field is optional, Filter applies whichever are set. Kept separate
// from OnePieceCharacterFilters (rather than one union struct) because the
// two genuinely diverge - see the owner's decision recorded in the plan -
// and a union would empty Canonical()'s field-count contract of meaning.
type JojoCharacterFilters struct {
	Rarity   *enums.PowerRarity
	Hamon    *enums.HamonLevel
	Spin     *enums.SpinLevel
	BattleIQ *byte
	// Search matches case-insensitively against name or the
	// locale-resolved description. Unescaped - callers must escape any
	// LIKE metacharacter (%, _, \) before this reaches SQL.
	Search *string
}

// Canonical - see StandFilters.Canonical's doc for why this exists and what
// it's the single source of truth for.
func (f JojoCharacterFilters) Canonical() string {
	battleIQ := ""
	if f.BattleIQ != nil {
		battleIQ = strconv.Itoa(int(*f.BattleIQ))
	}
	return strings.Join([]string{
		optStringer(f.Rarity),
		optStringer(f.Hamon),
		optStringer(f.Spin),
		battleIQ,
		optString(f.Search),
	}, "|")
}

// IJojoCharacterRepository is the admin-facing CRUD for JoJo characters.
type IJojoCharacterRepository interface {
	// Save upserts the given character's characters/jojo_characters rows,
	// then replaces its character_translations rows with translations
	// (en-GB mandatory, es-ES/ca-ES optional) - any locale missing from
	// translations is deleted. Same convention as IStandRepository.Save:
	// the caller already decided the id, so Save can't distinguish create
	// from update and doesn't try to. Returns ErrJojoCharacterAlreadyExists
	// on a (manga, name) conflict with a different character.
	Save(ctx context.Context, c *characters.JojoCharacter, translations CharacterTranslations) error
	FindByID(ctx context.Context, id characters.CharacterID, locale enums.Locale) (*characters.JojoCharacter, error)
	FindByName(ctx context.Context, name string, locale enums.Locale) (*characters.JojoCharacter, error)
	GetAll(ctx context.Context, locale enums.Locale) ([]*characters.JojoCharacter, error)
	Filter(ctx context.Context, filters JojoCharacterFilters, locale enums.Locale) ([]*characters.JojoCharacter, error)
	// Page returns up to limit+1 characters matching filters, ordered by
	// name after afterName, then the caller trims the extra row and reports
	// hasMore - same contract as IDevilFruitRepository.Page.
	Page(ctx context.Context, filters JojoCharacterFilters, locale enums.Locale, afterName *string, limit int) ([]*characters.JojoCharacter, bool, error)
	Count(ctx context.Context, filters JojoCharacterFilters, locale enums.Locale) (int, error)
	Delete(ctx context.Context, id characters.CharacterID) error
	// UpdatePicture updates only a character's picture renditions and
	// pipeline status. A nil main/thumb/card/lqip leaves that column
	// untouched.
	UpdatePicture(ctx context.Context, id characters.CharacterID, main, thumb, card, lqip *string, status enums.PictureStatus) error
	// SetMediaID updates only the content-addressed media group id - see
	// IStandRepository.SetMediaID.
	SetMediaID(ctx context.Context, id characters.CharacterID, mediaID string) error
	// Translations returns every locale's content for id, for the admin
	// edit form.
	Translations(ctx context.Context, id characters.CharacterID) (CharacterTranslations, error)
}
