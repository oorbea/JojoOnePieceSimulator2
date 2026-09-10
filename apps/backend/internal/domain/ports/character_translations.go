package ports

import "github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"

// CharacterTranslations is every locale's description for one Character, as
// submitted by an admin create/update request or read back for an admin
// edit form. Same shape as StageTranslations (no per-locale Skills - a
// Character has none) but, per the owner's decision, only en-GB is
// mandatory on write - es-ES/ca-ES are optional, same rule as
// PowerTranslations - callers validate this before it ever reaches a
// repository.
type CharacterTranslations map[enums.Locale]string
