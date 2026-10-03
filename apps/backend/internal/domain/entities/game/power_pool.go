package game

import (
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
)

// AvailablePowers is a per-team, mutable view over the power catalog used
// while assigning Loadouts: drawing a Stand or DevilFruit removes it - and the
// rest of its evolution family - from the pool, so the same team never
// receives the same Stand/DevilFruit, nor two stages of one evolution chain,
// within a game (Gauntlet) or a round (Versus). Each team must get
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
// most recent call to Stands()), together with its whole evolution family:
// every Stand sharing its StandFamilyKey leaves the pool, so a team can never
// hold two stages of the same family (no Whitesnake alongside C-MOON). The
// returned family (drawn Stand included) is what an evolution of the drawn
// Stand may pick its target from - it only holds what was still in the pool,
// so a stage the lobby banned is never offered.
func (p *AvailablePowers) DrawStand(i int) (*powers.Stand, []*powers.Stand, error) {
	if i < 0 || i >= len(p.stands) {
		return nil, nil, ErrPowerPoolExhausted
	}
	s := p.stands[i]
	key := StandFamilyKey(s)
	var family, rest []*powers.Stand
	for _, c := range p.stands {
		if StandFamilyKey(c) == key {
			family = append(family, c)
		} else {
			rest = append(rest, c)
		}
	}
	p.stands = rest
	return s, family, nil
}

// DrawDevilFruit removes and returns the DevilFruit at index i (as
// returned by the most recent call to DevilFruits()), plus its evolution
// family - see DrawStand.
func (p *AvailablePowers) DrawDevilFruit(i int) (*powers.DevilFruit, []*powers.DevilFruit, error) {
	if i < 0 || i >= len(p.devilFruits) {
		return nil, nil, ErrPowerPoolExhausted
	}
	d := p.devilFruits[i]
	key := FruitFamilyKey(d)
	var family, rest []*powers.DevilFruit
	for _, c := range p.devilFruits {
		if FruitFamilyKey(c) == key {
			family = append(family, c)
		} else {
			rest = append(rest, c)
		}
	}
	p.devilFruits = rest
	return d, family, nil
}
