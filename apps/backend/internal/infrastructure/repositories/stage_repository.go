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

// Stages implements ports.IStageCatalog.
func (r *StageRepository) Stages(ctx context.Context, manga enums.Manga) ([]game.Stage, error) {
	rows, err := r.queries.ListStagesByManga(ctx, manga.String())
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
func (r *StageRepository) List(ctx context.Context) ([]game.Stage, error) {
	rows, err := r.queries.ListStages(ctx)
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

// FindByID implements ports.IStageRepository.
func (r *StageRepository) FindByID(ctx context.Context, id game.StageID) (game.Stage, error) {
	row, err := r.queries.GetStageByID(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return game.Stage{}, ports.ErrStageNotFound
		}
		return game.Stage{}, fmt.Errorf("getting stage %s: %w", id, err)
	}
	return toStage(fromGetStageByIDRow(row))
}

// Save implements ports.IStageRepository.
func (r *StageRepository) Save(ctx context.Context, s game.Stage) error {
	_, err := r.queries.UpsertStage(ctx, db.UpsertStageParams{
		ID:       pgtype.UUID{Bytes: s.ID(), Valid: true},
		Manga:    s.Manga().String(),
		Position: int32(s.Order()),
		Name:     s.Name(),
	})
	if err != nil {
		return fmt.Errorf("saving stage %s: %w", s.ID(), wrapPgError(err, ports.ErrStageAlreadyExists))
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
