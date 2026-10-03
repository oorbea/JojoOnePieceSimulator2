package powers

import (
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

type DevilFruit struct {
	Power
	fruitType   enums.FruitType
	evolvesFrom *DevilFruit
}

func NewDevilFruit(power Power, fruitType enums.FruitType) (*DevilFruit, error) {
	if !fruitType.IsValid() {
		return nil, enums.ErrInvalidFruitType
	}
	return &DevilFruit{
		Power:     power,
		fruitType: fruitType,
	}, nil
}

func (d *DevilFruit) FruitType() enums.FruitType {
	return d.fruitType
}

// EvolvesFrom is the fruit this one is the awakened/evolved form of, or nil.
// Unlike Stand, it is not persisted: the evolution relation between fruits
// (Gomu Gomu no mi -> Hito Hito no mi: Model Nika) lives in game's rule table
// and is attached at load time via WithEvolvesFrom.
func (d *DevilFruit) EvolvesFrom() *DevilFruit {
	return d.evolvesFrom
}

// WithEvolvesFrom returns a shallow copy of d whose parent is parent, leaving
// d itself untouched - the catalogue's fruits are shared (and cached), so the
// relation must never be written onto them in place.
func (d *DevilFruit) WithEvolvesFrom(parent *DevilFruit) *DevilFruit {
	c := *d
	c.evolvesFrom = parent
	return &c
}

// EvolutionDepth is how many ancestors this fruit has - 0 for a base fruit.
// Same contract as Stand.EvolutionDepth (bounded against cycles).
func (d *DevilFruit) EvolutionDepth() int {
	depth := 0
	seen := map[string]struct{}{}
	for cur := d.evolvesFrom; cur != nil; cur = cur.evolvesFrom {
		id := cur.ID().String()
		if _, ok := seen[id]; ok {
			break
		}
		seen[id] = struct{}{}
		depth++
	}
	return depth
}
