package dto

import (
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
)

// StandResponse is the JSON representation of a Stand, recursively nesting
// its evolves_from ancestor chain (which the repository already loads in a
// single round trip).
type StandResponse struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Rarity      string         `json:"rarity"`
	Skills      []string       `json:"skills"`
	Picture     string         `json:"picture"`
	AttackPower string         `json:"attackPower"`
	Speed       string         `json:"speed"`
	AttackRange string         `json:"attackRange"`
	Endurance   string         `json:"endurance"`
	Precision   string         `json:"precision"`
	Potential   string         `json:"potential"`
	EvolvesFrom *StandResponse `json:"evolvesFrom"`
}

// NewStandResponse builds a StandResponse from a domain Stand.
func NewStandResponse(stand *powers.Stand) StandResponse {
	var evolvesFrom *StandResponse
	if parent := stand.EvolvesFrom(); parent != nil {
		resp := NewStandResponse(parent)
		evolvesFrom = &resp
	}

	skills := stand.Skills()
	if skills == nil {
		skills = []string{}
	}

	return StandResponse{
		ID:          stand.ID().String(),
		Name:        stand.Name(),
		Description: stand.Description(),
		Rarity:      stand.Rarity().String(),
		Skills:      skills,
		Picture:     stand.Picture(),
		AttackPower: stand.AttackPower().String(),
		Speed:       stand.Speed().String(),
		AttackRange: stand.AttackRange().String(),
		Endurance:   stand.Endurance().String(),
		Precision:   stand.Precision().String(),
		Potential:   stand.Potential().String(),
		EvolvesFrom: evolvesFrom,
	}
}

// NewStandResponses builds a StandResponse slice, never nil, from a list of
// domain Stands.
func NewStandResponses(stands []*powers.Stand) []StandResponse {
	responses := make([]StandResponse, 0, len(stands))
	for _, stand := range stands {
		responses = append(responses, NewStandResponse(stand))
	}
	return responses
}
