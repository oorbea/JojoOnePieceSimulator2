package ports

import "github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"

// PowerContent is a single locale's translatable content for a Power: its
// description and ordered skills. Power.Name is deliberately not part of
// this - proper nouns like "Star Platinum" read the same in every locale,
// so name stays a single untranslated column on `powers`.
type PowerContent struct {
	Description string
	Skills      []string
}

// PowerTranslations is every locale's content for one Power, as submitted
// by an admin create/update request or read back for an admin edit form.
// EnGB must always be present - it is the final link of the read-side
// fallback chain (ca-ES -> es-ES -> en-GB) - which callers validate before
// this ever reaches a repository.
type PowerTranslations map[enums.Locale]PowerContent
