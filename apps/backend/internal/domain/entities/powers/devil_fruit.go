package powers

import (
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

type DevilFruit struct {
	Power
	fruitType enums.FruitType
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
