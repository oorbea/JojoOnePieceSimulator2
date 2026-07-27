package services

import (
	"context"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

var standRepository ports.StandRepository

func CreateStand(
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
	var evolvesFromStand *powers.Stand = nil
	if evolvesFrom != nil {
		evolvesFromStand, err = standRepository.FindByName(context.Background(), *evolvesFrom)
	}
	if err != nil {
		return nil, err
	}
	stand, err := powers.NewStand(*power, attackPower, speed, attackRange, endurance, precision, potential, evolvesFromStand)
	if err != nil {
		return nil, err
	}
	err = standRepository.Save(context.Background(), stand)
	if err != nil {
		return nil, err
	}
	return stand, nil
}
