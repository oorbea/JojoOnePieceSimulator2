package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/postgres/db"
)

// characterTranslationQueries is the subset of *db.Queries needed by
// saveCharacterTranslations, satisfied by both a plain *db.Queries and a
// transaction-scoped one (q.WithTx(tx)) - shared by JojoCharacterRepository
// and OnePieceCharacterRepository since character_translations has no
// per-kind column.
type characterTranslationQueries interface {
	UpsertCharacterTranslation(ctx context.Context, arg db.UpsertCharacterTranslationParams) error
	DeleteCharacterTranslations(ctx context.Context, arg db.DeleteCharacterTranslationsParams) error
}

// saveCharacterTranslations replaces characterID's character_translations
// rows with translations wholesale - same semantics as saveTranslations
// (power_translations.go), minus Skills (a Character has none). Callers
// must always include en-GB in translations - deleting it would violate
// the read-side fallback invariant that every character has an en-GB
// translation.
func saveCharacterTranslations(ctx context.Context, q characterTranslationQueries, characterID pgtype.UUID, translations ports.CharacterTranslations) error {
	var toDelete []string
	for _, l := range enums.Locales() {
		if _, ok := translations[l]; !ok {
			toDelete = append(toDelete, l.String())
		}
	}
	if len(toDelete) > 0 {
		if err := q.DeleteCharacterTranslations(ctx, db.DeleteCharacterTranslationsParams{
			CharacterID: characterID,
			Locales:     toDelete,
		}); err != nil {
			return fmt.Errorf("deleting stale translations: %w", err)
		}
	}
	for locale, description := range translations {
		if err := q.UpsertCharacterTranslation(ctx, db.UpsertCharacterTranslationParams{
			CharacterID: characterID,
			Locale:      locale.String(),
			Description: description,
		}); err != nil {
			return fmt.Errorf("upserting %s translation: %w", locale, err)
		}
	}
	return nil
}

// characterTranslationsFromRows converts GetCharacterTranslations's rows
// into ports.CharacterTranslations, keyed by the parsed locale enum.
func characterTranslationsFromRows(rows []db.CharacterTranslation) (ports.CharacterTranslations, error) {
	out := make(ports.CharacterTranslations, len(rows))
	for _, row := range rows {
		locale, err := enums.ParseLocale(row.Locale)
		if err != nil {
			return nil, fmt.Errorf("row locale %q: %w", row.Locale, err)
		}
		out[locale] = row.Description
	}
	return out, nil
}
