package dto

import (
	"context"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
)

// StandResponse is the JSON representation of a Stand, recursively nesting
// its evolves_from ancestor chain (which the repository already loads in a
// single round trip).
type StandResponse struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	Rarity        string         `json:"rarity"`
	Skills        []string       `json:"skills"`
	Picture       string         `json:"picture"`
	PictureThumb  string         `json:"pictureThumb"`
	PictureStatus string         `json:"pictureStatus"`
	AttackPower   string         `json:"attackPower"`
	Speed         string         `json:"speed"`
	AttackRange   string         `json:"attackRange"`
	Endurance     string         `json:"endurance"`
	Precision     string         `json:"precision"`
	Potential     string         `json:"potential"`
	EvolvesFrom   *StandResponse `json:"evolvesFrom"`
}

// PictureURLResolver turns a Stand's stored picture key into a URL a client
// can GET, returning "" for a Stand with no picture.
type PictureURLResolver func(ctx context.Context, key string) (string, error)

// NewStandResponse builds a StandResponse from a domain Stand, resolving its
// picture key (and, recursively, its evolves_from chain's) through resolve.
func NewStandResponse(ctx context.Context, stand *powers.Stand, resolve PictureURLResolver) (StandResponse, error) {
	var evolvesFrom *StandResponse
	if parent := stand.EvolvesFrom(); parent != nil {
		resp, err := NewStandResponse(ctx, parent, resolve)
		if err != nil {
			return StandResponse{}, err
		}
		evolvesFrom = &resp
	}

	skills := stand.Skills()
	if skills == nil {
		skills = []string{}
	}

	pictureURL, err := resolve(ctx, stand.Picture())
	if err != nil {
		return StandResponse{}, err
	}
	pictureThumbURL, err := resolve(ctx, stand.PictureThumb())
	if err != nil {
		return StandResponse{}, err
	}

	return StandResponse{
		ID:            stand.ID().String(),
		Name:          stand.Name(),
		Description:   stand.Description(),
		Rarity:        stand.Rarity().String(),
		Skills:        skills,
		Picture:       pictureURL,
		PictureThumb:  pictureThumbURL,
		PictureStatus: stand.PictureStatus().String(),
		AttackPower:   stand.AttackPower().String(),
		Speed:         stand.Speed().String(),
		AttackRange:   stand.AttackRange().String(),
		Endurance:     stand.Endurance().String(),
		Precision:     stand.Precision().String(),
		Potential:     stand.Potential().String(),
		EvolvesFrom:   evolvesFrom,
	}, nil
}

// NewStandResponses builds a StandResponse slice, never nil, from a list of
// domain Stands.
func NewStandResponses(ctx context.Context, stands []*powers.Stand, resolve PictureURLResolver) ([]StandResponse, error) {
	responses := make([]StandResponse, 0, len(stands))
	for _, stand := range stands {
		resp, err := NewStandResponse(ctx, stand, resolve)
		if err != nil {
			return nil, err
		}
		responses = append(responses, resp)
	}
	return responses, nil
}
