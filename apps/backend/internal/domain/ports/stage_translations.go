package ports

import "github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"

// StageTranslations is every locale's description for one Stage, as
// submitted by an admin create/update request or read back for an admin
// edit form. Unlike PowerTranslations, there is no per-locale Skills (a
// Stage has none) and, per the owner's decision, every locale is mandatory
// on write - callers validate this before it ever reaches a repository.
type StageTranslations map[enums.Locale]string
