package powers

import (
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

type Stand struct {
	Power
	attackPower enums.StandStat
	speed       enums.StandStat
	attackRange enums.StandStat
	endurance   enums.StandStat
	precision   enums.StandStat
	potential   enums.StandStat
	evolvesFrom *Stand
}

func NewStand(
	power Power,
	attackPower enums.StandStat,
	speed enums.StandStat,
	attackRange enums.StandStat,
	endurance enums.StandStat,
	precision enums.StandStat,
	potential enums.StandStat,
	evolvesFrom *Stand,
) (*Stand, error) {
	if !attackPower.IsValid() {
		return nil, enums.ErrInvalidStandStat
	}
	if !speed.IsValid() {
		return nil, enums.ErrInvalidStandStat
	}
	if !attackRange.IsValid() {
		return nil, enums.ErrInvalidStandStat
	}
	if !endurance.IsValid() {
		return nil, enums.ErrInvalidStandStat
	}
	if !precision.IsValid() {
		return nil, enums.ErrInvalidStandStat
	}
	if !potential.IsValid() {
		return nil, enums.ErrInvalidStandStat
	}
	return &Stand{
		Power:       power,
		attackPower: attackPower,
		speed:       speed,
		attackRange: attackRange,
		endurance:   endurance,
		precision:   precision,
		potential:   potential,
		evolvesFrom: evolvesFrom,
	}, nil
}

func (s *Stand) AttackPower() enums.StandStat {
	return s.attackPower
}

func (s *Stand) Speed() enums.StandStat {
	return s.speed
}

func (s *Stand) AttackRange() enums.StandStat {
	return s.attackRange
}

func (s *Stand) Endurance() enums.StandStat {
	return s.endurance
}

func (s *Stand) Precision() enums.StandStat {
	return s.precision
}

func (s *Stand) Potential() enums.StandStat {
	return s.potential
}

func (s *Stand) EvolvesFrom() *Stand {
	return s.evolvesFrom
}
