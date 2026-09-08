package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/postgres/db"
)

// StageRepository is the Postgres-backed adapter for both ports.IStageCatalog
// (the read side game.IGameMode consumes) and ports.IStageRepository (the
// admin CRUD side) - one adapter satisfies both, same relationship
// StandRepository has with IStandRepository.
type StageRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

// NewStageRepository builds a StageRepository over pool.
func NewStageRepository(pool *pgxpool.Pool) *StageRepository {
	return &StageRepository{pool: pool, queries: db.New(pool)}
}

var _ ports.IStageCatalog = (*StageRepository)(nil)
var _ ports.IStageRepository = (*StageRepository)(nil)

// Stages implements ports.IStageCatalog. Description is resolved at a fixed
// enums.EnGB - see IStageCatalog's doc for why that's fine (nothing on the
// gameplay path ever reads it; a live match re-resolves per viewer at the
// transport layer instead).
func (r *StageRepository) Stages(ctx context.Context, manga enums.Manga) ([]game.Stage, error) {
	dbManga := db.Manga(manga.String())
	rows, err := r.queries.FilterStageRows(ctx, db.FilterStageRowsParams{
		Manga:   &dbManga,
		Locales: fallbackStrings(enums.EnGB),
	})
	if err != nil {
		return nil, fmt.Errorf("listing stages for manga %s: %w", manga, err)
	}
	stages := make([]game.Stage, 0, len(rows))
	for _, row := range rows {
		st, err := toStage(fromFilterStageRow(row))
		if err != nil {
			return nil, err
		}
		stages = append(stages, st)
	}
	return stages, nil
}

// List implements ports.IStageRepository.
func (r *StageRepository) List(ctx context.Context, locale enums.Locale) ([]game.Stage, error) {
	rows, err := r.queries.ListStages(ctx, fallbackStrings(locale))
	if err != nil {
		return nil, fmt.Errorf("listing stages: %w", err)
	}
	stages := make([]game.Stage, 0, len(rows))
	for _, row := range rows {
		st, err := toStage(fromListStagesRow(row))
		if err != nil {
			return nil, err
		}
		stages = append(stages, st)
	}
	return stages, nil
}

// Filter implements ports.IStageRepository - the locale-aware, admin-facing
// counterpart to Stages (IStageCatalog), matching whichever of filters'
// fields are set.
func (r *StageRepository) Filter(ctx context.Context, filters ports.StageFilters, locale enums.Locale) ([]game.Stage, error) {
	var dbManga *db.Manga
	if filters.Manga != nil {
		m := db.Manga(filters.Manga.String())
		dbManga = &m
	}
	rows, err := r.queries.FilterStageRows(ctx, db.FilterStageRowsParams{
		Manga:   dbManga,
		Search:  searchPtr(filters.Search),
		Locales: fallbackStrings(locale),
	})
	if err != nil {
		return nil, fmt.Errorf("filtering stages: %w", err)
	}
	stages := make([]game.Stage, 0, len(rows))
	for _, row := range rows {
		st, err := toStage(fromFilterStageRow(row))
		if err != nil {
			return nil, err
		}
		stages = append(stages, st)
	}
	return stages, nil
}

// Page implements ports.IStageRepository. after carries all three cursor
// fields together (nil for the first page) - see PageStageRows's doc for why
// the row-value comparison must not cast manga to ::text.
func (r *StageRepository) Page(ctx context.Context, filters ports.StageFilters, locale enums.Locale, after *ports.StagePageCursor, limit int) ([]game.Stage, bool, error) {
	var dbManga *db.Manga
	if filters.Manga != nil {
		m := db.Manga(filters.Manga.String())
		dbManga = &m
	}
	var afterManga *db.Manga
	var afterPosition *int32
	var afterName *string
	if after != nil {
		m := db.Manga(after.Manga.String())
		afterManga = &m
		pos := int32(after.Position)
		afterPosition = &pos
		afterName = &after.Name
	}
	rows, err := r.queries.PageStageRows(ctx, db.PageStageRowsParams{
		Manga:         dbManga,
		Search:        searchPtr(filters.Search),
		Locales:       fallbackStrings(locale),
		AfterManga:    afterManga,
		AfterPosition: afterPosition,
		AfterName:     afterName,
		PageLimit:     int32(limit + 1),
	})
	if err != nil {
		return nil, false, fmt.Errorf("paging stages: %w", err)
	}
	stages := make([]game.Stage, 0, len(rows))
	for _, row := range rows {
		st, err := toStage(fromPageStageRow(row))
		if err != nil {
			return nil, false, err
		}
		stages = append(stages, st)
	}
	hasMore := len(stages) > limit
	if hasMore {
		stages = stages[:limit]
	}
	return stages, hasMore, nil
}

// Count implements ports.IStageRepository.
func (r *StageRepository) Count(ctx context.Context, filters ports.StageFilters, locale enums.Locale) (int, error) {
	var dbManga *db.Manga
	if filters.Manga != nil {
		m := db.Manga(filters.Manga.String())
		dbManga = &m
	}
	count, err := r.queries.CountStageRows(ctx, db.CountStageRowsParams{
		Manga:   dbManga,
		Search:  searchPtr(filters.Search),
		Locales: fallbackStrings(locale),
	})
	if err != nil {
		return 0, fmt.Errorf("counting stages: %w", err)
	}
	return int(count), nil
}

// FindByID implements ports.IStageRepository.
func (r *StageRepository) FindByID(ctx context.Context, id game.StageID, locale enums.Locale) (game.Stage, error) {
	row, err := r.queries.GetStageByID(ctx, db.GetStageByIDParams{
		ID:      pgtype.UUID{Bytes: id, Valid: true},
		Locales: fallbackStrings(locale),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return game.Stage{}, ports.ErrStageNotFound
		}
		return game.Stage{}, fmt.Errorf("getting stage %s: %w", id, err)
	}
	return toStage(fromGetStageByIDRow(row))
}

// Save implements ports.IStageRepository, upserting the stage row and
// replacing its translations wholesale, in one transaction.
func (r *StageRepository) Save(ctx context.Context, s game.Stage, translations ports.StageTranslations) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			fmt.Printf("rolling back transaction for stage %s: %v\n", s.ID(), rollbackErr)
		}
	}()

	q := r.queries.WithTx(tx)

	if _, err := q.UpsertStage(ctx, db.UpsertStageParams{
		ID:            pgtype.UUID{Bytes: s.ID(), Valid: true},
		Manga:         s.Manga().String(),
		Position:      int32(s.Order()),
		Name:          s.Name(),
		Picture:       s.Picture(),
		PictureThumb:  s.PictureThumb(),
		PictureCard:   s.PictureCard(),
		PictureStatus: s.PictureStatus().String(),
		PictureLqip:   s.PictureLqip(),
	}); err != nil {
		return fmt.Errorf("saving stage %s: %w", s.ID(), wrapPgError(err, ports.ErrStageAlreadyExists))
	}

	if err := saveStageTranslations(ctx, q, pgtype.UUID{Bytes: s.ID(), Valid: true}, translations); err != nil {
		return fmt.Errorf("saving translations for stage %s: %w", s.ID(), err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing stage %s: %w", s.ID(), err)
	}
	return nil
}

// Delete implements ports.IStageRepository.
func (r *StageRepository) Delete(ctx context.Context, id game.StageID) error {
	n, err := r.queries.DeleteStageByID(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return fmt.Errorf("deleting stage %s: %w", id, err)
	}
	if n == 0 {
		return ports.ErrStageNotFound
	}
	return nil
}

// Translations implements ports.IStageRepository.
func (r *StageRepository) Translations(ctx context.Context, id game.StageID) (ports.StageTranslations, error) {
	rows, err := r.queries.GetStageTranslations(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("querying translations for stage %s: %w", id, err)
	}
	out := make(ports.StageTranslations, len(rows))
	for _, row := range rows {
		locale, err := enums.ParseLocale(row.Locale)
		if err != nil {
			return nil, fmt.Errorf("row locale %q: %w", row.Locale, err)
		}
		out[locale] = row.Description
	}
	return out, nil
}

// UpdatePicture implements ports.IStageRepository.
func (r *StageRepository) UpdatePicture(ctx context.Context, id game.StageID, main, thumb, card, lqip *string, status enums.PictureStatus) error {
	err := r.queries.UpdateStagePicture(ctx, db.UpdateStagePictureParams{
		ID:            pgtype.UUID{Bytes: id, Valid: true},
		Picture:       main,
		PictureThumb:  thumb,
		PictureCard:   card,
		PictureStatus: status.String(),
		PictureLqip:   lqip,
	})
	if err != nil {
		return fmt.Errorf("updating picture for stage %s: %w", id, err)
	}
	return nil
}

// SetMediaID implements ports.IStageRepository.
func (r *StageRepository) SetMediaID(ctx context.Context, id game.StageID, mediaID string) error {
	if err := r.queries.UpdateStageMediaID(ctx, db.UpdateStageMediaIDParams{
		ID:             pgtype.UUID{Bytes: id, Valid: true},
		PictureMediaID: mediaID,
	}); err != nil {
		return fmt.Errorf("setting media id for stage %s: %w", id, err)
	}
	return nil
}

// stageTranslationQueries is the subset of *db.Queries needed by
// saveStageTranslations, satisfied by both a plain *db.Queries and a
// transaction-scoped one (q.WithTx(tx)).
type stageTranslationQueries interface {
	UpsertStageTranslation(ctx context.Context, arg db.UpsertStageTranslationParams) error
	DeleteStageTranslations(ctx context.Context, arg db.DeleteStageTranslationsParams) error
}

// saveStageTranslations replaces stageID's stage_translations rows with
// translations wholesale - same shape as saveTranslations (power_translations.go),
// without the Skills field a Stage doesn't have.
func saveStageTranslations(ctx context.Context, q stageTranslationQueries, stageID pgtype.UUID, translations ports.StageTranslations) error {
	var toDelete []string
	for _, l := range enums.Locales() {
		if _, ok := translations[l]; !ok {
			toDelete = append(toDelete, l.String())
		}
	}
	if len(toDelete) > 0 {
		if err := q.DeleteStageTranslations(ctx, db.DeleteStageTranslationsParams{
			StageID: stageID,
			Locales: toDelete,
		}); err != nil {
			return fmt.Errorf("deleting stale translations: %w", err)
		}
	}
	for locale, description := range translations {
		if err := q.UpsertStageTranslation(ctx, db.UpsertStageTranslationParams{
			StageID:     stageID,
			Locale:      locale.String(),
			Description: description,
		}); err != nil {
			return fmt.Errorf("upserting %s translation: %w", locale, err)
		}
	}
	return nil
}
