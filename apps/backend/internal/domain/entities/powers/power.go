package powers

import (
	"errors"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

type Power struct {
	id          PowerID
	name        string
	description string
	rarity      enums.PowerRarity
	skills      []string
	picture     string
}

func NewPower(
	id PowerID,
	name string,
	description string,
	rarity enums.PowerRarity,
	skills *[]string,
	picture string,
) (*Power, error) {
	if id.IsNil() {
		return nil, errors.New("id is required")
	}
	if name == "" {
		return nil, errors.New("name is required")
	}
	if description == "" {
		return nil, errors.New("description is required")
	}
	if !rarity.IsValid() {
		return nil, enums.ErrInvalidRarity
	}
	if skills == nil || len(*skills) < 1 {
		return nil, errors.New("skills are required")
	}
	return &Power{
		id:          id,
		name:        name,
		description: description,
		rarity:      rarity,
		skills:      *skills,
		picture:     picture,
	}, nil
}

func (p Power) ID() PowerID {
	return p.id
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

func (p Power) Picture() string {
	return p.picture
}

// SetPicture replaces the stored picture key. The empty string means "no
// picture", which NewPower already allows.
func (p *Power) SetPicture(picture string) {
	p.picture = picture
}
