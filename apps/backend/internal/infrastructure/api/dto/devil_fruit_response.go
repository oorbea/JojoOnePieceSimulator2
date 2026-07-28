package dto

import (
	"context"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
)

// DevilFruitResponse is the JSON representation of a DevilFruit.
type DevilFruitResponse struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Rarity        string   `json:"rarity"`
	Skills        []string `json:"skills"`
	Picture       string   `json:"picture"`
	PictureThumb  string   `json:"pictureThumb"`
	PictureStatus string   `json:"pictureStatus"`
	FruitType     string   `json:"fruitType"`
}

// NewDevilFruitResponse builds a DevilFruitResponse from a domain DevilFruit,
// resolving its picture key through resolve.
func NewDevilFruitResponse(ctx context.Context, fruit *powers.DevilFruit, resolve PictureURLResolver) (DevilFruitResponse, error) {
	skills := fruit.Skills()
	if skills == nil {
		skills = []string{}
	}

	pictureURL, err := resolve(ctx, fruit.Picture())
	if err != nil {
		return DevilFruitResponse{}, err
	}
	pictureThumbURL, err := resolve(ctx, fruit.PictureThumb())
	if err != nil {
		return DevilFruitResponse{}, err
	}

	return DevilFruitResponse{
		ID:            fruit.ID().String(),
		Name:          fruit.Name(),
		Description:   fruit.Description(),
		Rarity:        fruit.Rarity().String(),
		Skills:        skills,
		Picture:       pictureURL,
		PictureThumb:  pictureThumbURL,
		PictureStatus: fruit.PictureStatus().String(),
		FruitType:     fruit.FruitType().String(),
	}, nil
}

// NewDevilFruitResponses builds a DevilFruitResponse slice, never nil, from a
// list of domain DevilFruits.
func NewDevilFruitResponses(ctx context.Context, fruits []*powers.DevilFruit, resolve PictureURLResolver) ([]DevilFruitResponse, error) {
	responses := make([]DevilFruitResponse, 0, len(fruits))
	for _, fruit := range fruits {
		resp, err := NewDevilFruitResponse(ctx, fruit, resolve)
		if err != nil {
			return nil, err
		}
		responses = append(responses, resp)
	}
	return responses, nil
}
