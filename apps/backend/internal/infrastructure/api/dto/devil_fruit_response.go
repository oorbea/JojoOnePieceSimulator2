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
	Rarity        string   `json:"rarity" ts:"PowerRarity"`
	Skills        []string `json:"skills"`
	Picture       string   `json:"picture"`
	PictureThumb  string   `json:"pictureThumb"`
	PictureCard   string   `json:"pictureCard"`
	PictureStatus string   `json:"pictureStatus" ts:"PictureStatus"`
	PictureLqip   string   `json:"pictureLqip"`
	FocalX        float64  `json:"focalX"`
	FocalY        float64  `json:"focalY"`
	FruitType     string   `json:"fruitType" ts:"FruitType"`
	// EvolvesFrom is the fruit this one is the evolved form of (Model Nika's
	// parent is Gomu Gomu no mi). Always null from the catalogue endpoints -
	// the relation lives in game's rule table and is only attached to the
	// fruit of a drawn loadout - see game.LinkFruitEvolutions.
	EvolvesFrom *DevilFruitResponse `json:"evolvesFrom"`
}

// NewDevilFruitResponse builds a DevilFruitResponse from a domain DevilFruit,
// resolving its picture key through resolve.
func NewDevilFruitResponse(ctx context.Context, fruit *powers.DevilFruit, resolve PictureURLResolver, media MediaURLBuilder) (DevilFruitResponse, error) {
	var evolvesFrom *DevilFruitResponse
	if parent := fruit.EvolvesFrom(); parent != nil {
		resp, err := NewDevilFruitResponse(ctx, parent, resolve, media)
		if err != nil {
			return DevilFruitResponse{}, err
		}
		evolvesFrom = &resp
	}

	skills := fruit.Skills()
	if skills == nil {
		skills = []string{}
	}

	pictureURL, pictureThumbURL, pictureCardURL, err := resolveCatalogPictures(
		ctx, fruit.Picture(), fruit.PictureThumb(), fruit.PictureCard(), fruit.PictureMediaID(), resolve, media,
	)
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
		PictureCard:   pictureCardURL,
		PictureLqip:   fruit.PictureLqip(),
		PictureStatus: fruit.PictureStatus().String(),
		FocalX:        fruit.FocalX(),
		FocalY:        fruit.FocalY(),
		FruitType:     fruit.FruitType().String(),
		EvolvesFrom:   evolvesFrom,
	}, nil
}

// NewDevilFruitResponses builds a DevilFruitResponse slice, never nil, from a
// list of domain DevilFruits.
func NewDevilFruitResponses(ctx context.Context, fruits []*powers.DevilFruit, resolve PictureURLResolver, media MediaURLBuilder) ([]DevilFruitResponse, error) {
	responses := make([]DevilFruitResponse, 0, len(fruits))
	for _, fruit := range fruits {
		resp, err := NewDevilFruitResponse(ctx, fruit, resolve, media)
		if err != nil {
			return nil, err
		}
		responses = append(responses, resp)
	}
	return responses, nil
}

// DevilFruitPageResponse is GET /devil-fruits's response body when the
// request opts into pagination - see StandPageResponse's doc for the wire
// shape and why this is a concrete per-resource envelope.
type DevilFruitPageResponse struct {
	PageInfo
	Items []DevilFruitResponse `json:"items"`
}
