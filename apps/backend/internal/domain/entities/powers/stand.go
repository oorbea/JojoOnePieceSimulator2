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

// EvolutionDepth is how many ancestors this Stand has - 0 for a base stand
// with no evolvesFrom, 1 for a one-step evolution (e.g. Chariot Requiem),
// and so on for a multi-stage family (e.g. Echoes ACT3 has 3). Used by
// game.RevealDuration/RevealPlayer to size the sorteo's "algo esta pasando"
// evolution reveal (2026-09-25) - mirrored on the frontend by
// stand-evolution.ts's standEvolutionSteps, keep both in sync. The
// self-evolution DB constraint (stands_no_self_evolution) and the mapper's
// own cycle check (stand_mapper.go) mean this can't loop in practice; it's
// still bounded defensively rather than trusted blindly.
func (s *Stand) EvolutionDepth() int {
	depth := 0
	seen := map[string]struct{}{}
	for cur := s.evolvesFrom; cur != nil; cur = cur.evolvesFrom {
		id := cur.ID().String()
		if _, ok := seen[id]; ok {
			break
		}
		seen[id] = struct{}{}
		depth++
	}
	return depth
}
