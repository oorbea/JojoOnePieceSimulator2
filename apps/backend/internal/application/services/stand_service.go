package services

import (
	"context"
	"errors"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// ErrSelfEvolution is returned when a Stand is asked to evolve from itself.
var ErrSelfEvolution = errors.New("a stand cannot evolve from itself")

// StandInput carries every field needed to create or update a Stand, so
// CreateStand/UpdateStand take one argument instead of a long positional
// list.
type StandInput struct {
	Name        string
	Description string
	Rarity      enums.PowerRarity
	Skills      *[]string
	Picture     string
	AttackPower enums.StandStat
	Speed       enums.StandStat
	AttackRange enums.StandStat
	Endurance   enums.StandStat
	Precision   enums.StandStat
	Potential   enums.StandStat
	EvolvesFrom *powers.PowerID
}

// StandService coordinates Stand use cases against the injected repository.
type StandService struct {
	standRepo ports.IStandRepository
	idGen     ports.IIdGenerator[powers.PowerID]
}

func NewStandService(standRepo ports.IStandRepository, idGen ports.IIdGenerator[powers.PowerID]) *StandService {
	return &StandService{standRepo: standRepo, idGen: idGen}
}

// CreateStand builds a new Stand with a freshly generated id and persists it.
func (s *StandService) CreateStand(ctx context.Context, input StandInput) (*powers.Stand, error) {
	return s.saveStand(ctx, s.idGen.NewID(), input)
}

// UpdateStand rebuilds the Stand identified by id with the given fields and
// persists it, keeping its original id.
func (s *StandService) UpdateStand(ctx context.Context, id powers.PowerID, input StandInput) (*powers.Stand, error) {
	if _, err := s.standRepo.FindByID(ctx, id); err != nil {
		return nil, err
	}
	if input.EvolvesFrom != nil && *input.EvolvesFrom == id {
		return nil, ErrSelfEvolution
	}
	return s.saveStand(ctx, id, input)
}

func (s *StandService) saveStand(ctx context.Context, id powers.PowerID, input StandInput) (*powers.Stand, error) {
	power, err := powers.NewPower(id, input.Name, input.Description, input.Rarity, input.Skills, input.Picture)
	if err != nil {
		return nil, err
	}

	var evolvesFromStand *powers.Stand
	if input.EvolvesFrom != nil {
		evolvesFromStand, err = s.standRepo.FindByID(ctx, *input.EvolvesFrom)
		if err != nil {
			return nil, err
		}
	}

	stand, err := powers.NewStand(*power, input.AttackPower, input.Speed, input.AttackRange, input.Endurance, input.Precision, input.Potential, evolvesFromStand)
	if err != nil {
		return nil, err
	}

	if err := s.standRepo.Save(ctx, stand); err != nil {
		return nil, err
	}
	return stand, nil
}

// GetStand returns the stand identified by id.
func (s *StandService) GetStand(ctx context.Context, id powers.PowerID) (*powers.Stand, error) {
	return s.standRepo.FindByID(ctx, id)
}

// ListStands returns every stand.
func (s *StandService) ListStands(ctx context.Context) ([]*powers.Stand, error) {
	return s.standRepo.GetAll(ctx)
}

// FilterStands returns every stand matching the given filters.
func (s *StandService) FilterStands(ctx context.Context, filters ports.StandFilters) ([]*powers.Stand, error) {
	return s.standRepo.Filter(ctx, filters)
}

// DeleteStand removes the stand identified by id.
func (s *StandService) DeleteStand(ctx context.Context, id powers.PowerID) error {
	return s.standRepo.Delete(ctx, id)
}
