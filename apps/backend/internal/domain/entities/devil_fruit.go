package entities

import "github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"

type DevilFruit struct {
	Power
	fruitType enums.FruitType
}

func NewDevilFruit(power Power, fruitType enums.FruitType) *DevilFruit {
	return &DevilFruit{
		Power:     power,
		fruitType: fruitType,
	}
}

func (d *DevilFruit) FruitType() enums.FruitType {
	return d.fruitType
}
