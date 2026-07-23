package entities

import "github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"

type DevilFruit struct {
	Power
	fruitType enums.FruitType
}

func (d *DevilFruit) FruitType() enums.FruitType {
	return d.fruitType
}
