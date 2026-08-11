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
	rows, err := r.queries.ListStagesByManga(ctx, db.ListStagesByMangaParams{
		Manga:   manga.String(),
		Locales: fallbackStrings(enums.EnGB),
	})
	if err != nil {
		return nil, fmt.Errorf("listing stages for manga %s: %w", manga, err)
	}
	stages := make([]game.Stage, 0, len(rows))
	for _, row := range rows {
		st, err := toStage(fromListStagesByMangaRow(row))
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

// ListByManga implements ports.IStageRepository - the locale-aware,
// admin-facing counterpart to Stages (IStageCatalog).
func (r *StageRepository) ListByManga(ctx context.Context, manga enums.Manga, locale enums.Locale) ([]game.Stage, error) {
	rows, err := r.queries.ListStagesByManga(ctx, db.ListStagesByMangaParams{
		Manga:   manga.String(),
		Locales: fallbackStrings(locale),
	})
	if err != nil {
		return nil, fmt.Errorf("listing stages for manga %s: %w", manga, err)
	}
	stages := make([]game.Stage, 0, len(rows))
	for _, row := range rows {
		st, err := toStage(fromListStagesByMangaRow(row))
		if err != nil {
			return nil, err
		}
		stages = append(stages, st)
	}
	return stages, nil
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
		PictureStatus: s.PictureStatus().String(),
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
func (r *StageRepository) UpdatePicture(ctx context.Context, id game.StageID, main, thumb *string, status enums.PictureStatus) error {
	err := r.queries.UpdateStagePicture(ctx, db.UpdateStagePictureParams{
		ID:            pgtype.UUID{Bytes: id, Valid: true},
		Picture:       main,
		PictureThumb:  thumb,
		PictureStatus: status.String(),
	})
	if err != nil {
		return fmt.Errorf("updating picture for stage %s: %w", id, err)
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
