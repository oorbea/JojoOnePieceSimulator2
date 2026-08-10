package ports

import (
	"context"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
)

// IGamePowerPool is the narrow, role-specific view of the power catalog
// that game assignment needs - deliberately not IStandRepository/
// IDevilFruitRepository themselves, so the game domain never has to know
// about locale or filtering. An adapter wraps
// IStandRepository.GetAll/IDevilFruitRepository.GetAll (any fixed locale -
// names, not descriptions, are what matters for assignment).
type IGamePowerPool interface {
	Stands(ctx context.Context) ([]*powers.Stand, error)
	DevilFruits(ctx context.Context) ([]*powers.DevilFruit, error)
}
