package entities

import "github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"

type Power struct {
	name        string
	description string
	rarity      enums.PowerRarity
}

func (p Power) Name() string {
	return p.name
}

func (p Power) Description() string {
	return p.description
}

func (p Power) Rarity() enums.PowerRarity {
	return p.rarity
}
