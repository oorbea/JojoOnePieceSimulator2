package entities

type DevilFruit struct {
	Power
	fruitType string
	isAwaken  bool
}

func (d *DevilFruit) FruitType() string {
	return d.fruitType
}

func (d *DevilFruit) IsAwaken() bool {
	return d.isAwaken
}
