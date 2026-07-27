package services

import (
	"context"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// StandService coordinates Stand creation against the injected repository.
type StandService struct {
	standRepo ports.IStandRepository
}

func NewStandService(standRepo ports.IStandRepository) *StandService {
	return &StandService{standRepo: standRepo}
}

func (s *StandService) CreateStand(
	ctx context.Context,
	name string,
	description string,
	rarity enums.PowerRarity,
	skills *[]string,
	picture string,
	attackPower enums.StandStat,
	speed enums.StandStat,
	attackRange enums.StandStat,
	endurance enums.StandStat,
	precision enums.StandStat,
	potential enums.StandStat,
	evolvesFrom *string,
) (*powers.Stand, error) {
	power, err := powers.NewPower(name, description, rarity, skills, picture)
	if err != nil {
		return nil, err
	}

	var evolvesFromStand *powers.Stand
	if evolvesFrom != nil {
		evolvesFromStand, err = s.standRepo.FindByName(ctx, *evolvesFrom)
		if err != nil {
			return nil, err
		}
	}

	stand, err := powers.NewStand(*power, attackPower, speed, attackRange, endurance, precision, potential, evolvesFromStand)
	if err != nil {
		return nil, err
	}

	if err := s.standRepo.Save(ctx, stand); err != nil {
		return nil, err
	}
	return stand, nil
}
