package game

import (
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
)

// AvailablePowers is a per-team, mutable view over the power catalog used
// while assigning Loadouts: drawing a Stand or DevilFruit removes it from
// the pool, so the same team never receives the same Stand/DevilFruit
// twice within a game (Gauntlet) or a round (Versus). Each team must get
// its own AvailablePowers built from the same underlying catalog - the
// rival team in Versus may then draw the same powers.
//
// AvailablePowers is never a field of Game and is never persisted: the
// application layer (GameService.beginRound) builds a fresh pool from
// ports.IGamePowerPool on every loadout assignment and discards it once
// AssignLoadouts returns. Only the drawn Loadout survives - see
// Snapshot/Restore.
type AvailablePowers struct {
	stands      []*powers.Stand
	devilFruits []*powers.DevilFruit
}

// NewAvailablePowers builds a pool from the given catalog slices, copying
// them so later mutation of the caller's slices does not affect the pool.
func NewAvailablePowers(stands []*powers.Stand, devilFruits []*powers.DevilFruit) *AvailablePowers {
	return &AvailablePowers{
		stands:      append([]*powers.Stand(nil), stands...),
		devilFruits: append([]*powers.DevilFruit(nil), devilFruits...),
	}
}

// Stands returns a copy of the Stands still available to draw.
func (p *AvailablePowers) Stands() []*powers.Stand {
	return append([]*powers.Stand(nil), p.stands...)
}

// DevilFruits returns a copy of the DevilFruits still available to draw.
func (p *AvailablePowers) DevilFruits() []*powers.DevilFruit {
	return append([]*powers.DevilFruit(nil), p.devilFruits...)
}

// DrawStand removes and returns the Stand at index i (as returned by the
// most recent call to Stands()).
func (p *AvailablePowers) DrawStand(i int) (*powers.Stand, error) {
	if i < 0 || i >= len(p.stands) {
		return nil, ErrPowerPoolExhausted
	}
	s := p.stands[i]
	p.stands = append(p.stands[:i:i], p.stands[i+1:]...)
	return s, nil
}

// DrawDevilFruit removes and returns the DevilFruit at index i (as
// returned by the most recent call to DevilFruits()).
func (p *AvailablePowers) DrawDevilFruit(i int) (*powers.DevilFruit, error) {
	if i < 0 || i >= len(p.devilFruits) {
		return nil, ErrPowerPoolExhausted
	}
	d := p.devilFruits[i]
	p.devilFruits = append(p.devilFruits[:i:i], p.devilFruits[i+1:]...)
	return d, nil
}
