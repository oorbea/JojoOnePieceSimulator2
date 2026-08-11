package services

import (
	"context"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// StageInput is the mutable content of a Stage - everything NewStage needs
// besides its ID.
type StageInput struct {
	Manga enums.Manga
	Order int
	Name  string
}

// StageService is the admin-facing CRUD service over ports.IStageRepository,
// the counterpart to StandService/DevilFruitService for the game feature's
// round content. It has no picture/imaging concerns - a Stage carries only
// Manga/Order/Name (see entities/game.Stage).
type StageService struct {
	repo    ports.IStageRepository
	catalog ports.IStageCatalog
	ids     ports.IIdGenerator[game.StageID]
}

// NewStageService builds a StageService. repo and catalog are typically the
// same underlying adapter (see repositories.StageRepository), passed
// separately because they're two distinct ports.
func NewStageService(repo ports.IStageRepository, catalog ports.IStageCatalog, ids ports.IIdGenerator[game.StageID]) *StageService {
	return &StageService{repo: repo, catalog: catalog, ids: ids}
}

// ListStages returns every Stage, ordered by manga then position then name.
func (s *StageService) ListStages(ctx context.Context) ([]game.Stage, error) {
	return s.repo.List(ctx)
}

// StagesByManga returns every Stage for manga, ordered by position.
func (s *StageService) StagesByManga(ctx context.Context, manga enums.Manga) ([]game.Stage, error) {
	return s.catalog.Stages(ctx, manga)
}

// GetStage returns the Stage matching id, or ports.ErrStageNotFound.
func (s *StageService) GetStage(ctx context.Context, id game.StageID) (game.Stage, error) {
	return s.repo.FindByID(ctx, id)
}

// CreateStage builds and persists a new Stage with a freshly generated ID.
func (s *StageService) CreateStage(ctx context.Context, input StageInput) (game.Stage, error) {
	st, err := game.NewStage(s.ids.NewID(), input.Manga, input.Order, input.Name)
	if err != nil {
		return game.Stage{}, err
	}
	if err := s.repo.Save(ctx, st); err != nil {
		return game.Stage{}, err
	}
	return st, nil
}

// UpdateStage replaces the content of the Stage matching id, keeping its ID.
func (s *StageService) UpdateStage(ctx context.Context, id game.StageID, input StageInput) (game.Stage, error) {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return game.Stage{}, err
	}
	st, err := game.NewStage(id, input.Manga, input.Order, input.Name)
	if err != nil {
		return game.Stage{}, err
	}
	if err := s.repo.Save(ctx, st); err != nil {
		return game.Stage{}, err
	}
	return st, nil
}

// DeleteStage removes the Stage matching id.
func (s *StageService) DeleteStage(ctx context.Context, id game.StageID) error {
	return s.repo.Delete(ctx, id)
}
