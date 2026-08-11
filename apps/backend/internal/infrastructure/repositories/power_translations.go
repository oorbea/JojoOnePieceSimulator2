package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/postgres/db"
)

// likeEscaper escapes the three characters that are meaningful inside a
// Postgres LIKE/ILIKE pattern (\, %, _), so a user-supplied search term is
// always matched literally. Queries built from escapeLikePattern's output
// must pair it with `ESCAPE '\'` in SQL.
var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// escapeLikePattern escapes s for safe embedding inside a `'%' || ... || '%'`
// ILIKE pattern - without this, a search term containing % or _ would match
// far more than the user typed (see FilterStandRows/FilterDevilFruitRows/
// FilterStageRows).
func escapeLikePattern(s string) string {
	return likeEscaper.Replace(s)
}

// searchPtr escapes and re-wraps an optional search term for use as a
// nullable ILIKE query parameter, leaving nil untouched.
func searchPtr(s *string) *string {
	if s == nil {
		return nil
	}
	escaped := escapeLikePattern(*s)
	return &escaped
}

// fallbackStrings renders enums.FallbackChain(locale) as the []string the
// generated queries expect for their `locales` parameter (most specific
// first, always ending in en-GB).
func fallbackStrings(locale enums.Locale) []string {
	chain := enums.FallbackChain(locale)
	out := make([]string, len(chain))
	for i, l := range chain {
		out[i] = l.String()
	}
	return out
}

// translationQueries is the subset of *db.Queries needed by
// saveTranslations, satisfied by both a plain *db.Queries and a
// transaction-scoped one (q.WithTx(tx)).
type translationQueries interface {
	UpsertPowerTranslation(ctx context.Context, arg db.UpsertPowerTranslationParams) error
	DeletePowerTranslations(ctx context.Context, arg db.DeletePowerTranslationsParams) error
}

// saveTranslations replaces powerID's power_translations rows with
// translations wholesale: every locale present is upserted, every supported
// locale absent from translations is deleted. Callers must always include
// en-GB in translations - deleting it would violate the read-side fallback
// invariant that every power has an en-GB translation.
func saveTranslations(ctx context.Context, q translationQueries, powerID pgtype.UUID, translations ports.PowerTranslations) error {
	var toDelete []string
	for _, l := range enums.Locales() {
		if _, ok := translations[l]; !ok {
			toDelete = append(toDelete, l.String())
		}
	}
	if len(toDelete) > 0 {
		if err := q.DeletePowerTranslations(ctx, db.DeletePowerTranslationsParams{
			PowerID: powerID,
			Locales: toDelete,
		}); err != nil {
			return fmt.Errorf("deleting stale translations: %w", err)
		}
	}
	for locale, content := range translations {
		if err := q.UpsertPowerTranslation(ctx, db.UpsertPowerTranslationParams{
			PowerID:     powerID,
			Locale:      locale.String(),
			Description: content.Description,
			Skills:      content.Skills,
		}); err != nil {
			return fmt.Errorf("upserting %s translation: %w", locale, err)
		}
	}
	return nil
}

// translationsFromRows converts GetPowerTranslations's rows into
// ports.PowerTranslations, keyed by the parsed locale enum.
func translationsFromRows(rows []db.PowerTranslation) (ports.PowerTranslations, error) {
	out := make(ports.PowerTranslations, len(rows))
	for _, row := range rows {
		locale, err := enums.ParseLocale(row.Locale)
		if err != nil {
			return nil, fmt.Errorf("row locale %q: %w", row.Locale, err)
		}
		out[locale] = ports.PowerContent{Description: row.Description, Skills: row.Skills}
	}
	return out, nil
}
