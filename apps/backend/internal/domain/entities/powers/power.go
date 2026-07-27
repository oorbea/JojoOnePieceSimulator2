package powers

import (
	"errors"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

type Power struct {
	name        string
	description string
	rarity      enums.PowerRarity
	skills      []string
}

func NewPower(
	name string,
	description string,
	rarity enums.PowerRarity,
	skills []string,
) (*Power, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}
	if description == "" {
		return nil, errors.New("description is required")
	}
	if !rarity.IsValid() {
		return nil, enums.ErrInvalidRarity
	}
	if len(skills) < 1 {
		return nil, errors.New("skills are required")
	}
	return &Power{
		name:        name,
		description: description,
		rarity:      rarity,
		skills:      skills,
	}, nil
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

func (p Power) Skills() []string {
	return p.skills
}
