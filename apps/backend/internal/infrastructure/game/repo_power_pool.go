package game

import (
	"context"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// RepoPowerPool is a ports.IGamePowerPool adapter over the existing
// IStandRepository/IDevilFruitRepository - names, not localized
// descriptions, are all Loadout assignment needs, so it always resolves
// enums.EnGB regardless of the caller's own locale.
type RepoPowerPool struct {
	stands      ports.IStandRepository
	devilFruits ports.IDevilFruitRepository
}

// NewRepoPowerPool builds a RepoPowerPool over the given catalog repos.
func NewRepoPowerPool(stands ports.IStandRepository, devilFruits ports.IDevilFruitRepository) *RepoPowerPool {
	return &RepoPowerPool{stands: stands, devilFruits: devilFruits}
}

var _ ports.IGamePowerPool = (*RepoPowerPool)(nil)

func (p *RepoPowerPool) Stands(ctx context.Context) ([]*powers.Stand, error) {
	return p.stands.GetAll(ctx, enums.EnGB)
}

func (p *RepoPowerPool) DevilFruits(ctx context.Context) ([]*powers.DevilFruit, error) {
	return p.devilFruits.GetAll(ctx, enums.EnGB)
}
