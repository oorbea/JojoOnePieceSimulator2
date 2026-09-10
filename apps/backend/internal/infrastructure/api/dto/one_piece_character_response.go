package dto

import (
	"context"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/characters"
)

// OnePieceCharacterResponse is the JSON representation of an
// OnePieceCharacter - same convention as JojoCharacterResponse.
type OnePieceCharacterResponse struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	Rarity          string `json:"rarity" ts:"PowerRarity"`
	Picture         string `json:"picture"`
	PictureThumb    string `json:"pictureThumb"`
	PictureCard     string `json:"pictureCard"`
	PictureStatus   string `json:"pictureStatus" ts:"PictureStatus"`
	PictureLqip     string `json:"pictureLqip"`
	PhysicalForm    string `json:"physicalForm" ts:"PhysicalForm"`
	ArmamentHaki    string `json:"armamentHaki" ts:"HakiLevel"`
	ObservationHaki string `json:"observationHaki" ts:"HakiLevel"`
	ConquerorHaki   string `json:"conquerorHaki" ts:"HakiLevel"`
	FruitMastery    string `json:"fruitMastery" ts:"FruitMastery"`
}

// NewOnePieceCharacterResponse builds an OnePieceCharacterResponse from a
// domain OnePieceCharacter, resolving its picture key through resolve.
func NewOnePieceCharacterResponse(ctx context.Context, c *characters.OnePieceCharacter, resolve PictureURLResolver, media MediaURLBuilder) (OnePieceCharacterResponse, error) {
	pictureURL, pictureThumbURL, pictureCardURL, err := resolveCatalogPictures(
		ctx, c.Picture(), c.PictureThumb(), c.PictureCard(), c.PictureMediaID(), resolve, media,
	)
	if err != nil {
		return OnePieceCharacterResponse{}, err
	}

	return OnePieceCharacterResponse{
		ID:              c.ID().String(),
		Name:            c.Name(),
		Description:     c.Description(),
		Rarity:          c.Rarity().String(),
		Picture:         pictureURL,
		PictureThumb:    pictureThumbURL,
		PictureCard:     pictureCardURL,
		PictureLqip:     c.PictureLqip(),
		PictureStatus:   c.PictureStatus().String(),
		PhysicalForm:    c.PhysicalForm().String(),
		ArmamentHaki:    c.ArmamentHaki().String(),
		ObservationHaki: c.ObservationHaki().String(),
		ConquerorHaki:   c.ConquerorHaki().String(),
		FruitMastery:    c.FruitMastery().String(),
	}, nil
}

// NewOnePieceCharacterResponses builds an OnePieceCharacterResponse slice,
// never nil, from a list of domain OnePieceCharacters.
func NewOnePieceCharacterResponses(ctx context.Context, list []*characters.OnePieceCharacter, resolve PictureURLResolver, media MediaURLBuilder) ([]OnePieceCharacterResponse, error) {
	responses := make([]OnePieceCharacterResponse, 0, len(list))
	for _, c := range list {
		resp, err := NewOnePieceCharacterResponse(ctx, c, resolve, media)
		if err != nil {
			return nil, err
		}
		responses = append(responses, resp)
	}
	return responses, nil
}

// OnePieceCharacterPageResponse is GET /one-piece-characters's response
// body when the request opts into pagination.
type OnePieceCharacterPageResponse struct {
	PageInfo
	Items []OnePieceCharacterResponse `json:"items"`
}
