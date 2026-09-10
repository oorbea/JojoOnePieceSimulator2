package dto

import (
	"context"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/characters"
)

// JojoCharacterResponse is the JSON representation of a JojoCharacter. No
// Skills field (a Character has none) and no Manga field (implied by the
// endpoint) - see DevilFruitResponse for the analogous omission of Kind.
type JojoCharacterResponse struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Rarity        string `json:"rarity" ts:"PowerRarity"`
	Picture       string `json:"picture"`
	PictureThumb  string `json:"pictureThumb"`
	PictureCard   string `json:"pictureCard"`
	PictureStatus string `json:"pictureStatus" ts:"PictureStatus"`
	PictureLqip   string `json:"pictureLqip"`
	Hamon         string `json:"hamon" ts:"HamonLevel"`
	Spin          string `json:"spin" ts:"SpinLevel"`
	BattleIQ      int    `json:"battleIq"`
}

// NewJojoCharacterResponse builds a JojoCharacterResponse from a domain
// JojoCharacter, resolving its picture key through resolve.
func NewJojoCharacterResponse(ctx context.Context, c *characters.JojoCharacter, resolve PictureURLResolver, media MediaURLBuilder) (JojoCharacterResponse, error) {
	pictureURL, pictureThumbURL, pictureCardURL, err := resolveCatalogPictures(
		ctx, c.Picture(), c.PictureThumb(), c.PictureCard(), c.PictureMediaID(), resolve, media,
	)
	if err != nil {
		return JojoCharacterResponse{}, err
	}

	return JojoCharacterResponse{
		ID:            c.ID().String(),
		Name:          c.Name(),
		Description:   c.Description(),
		Rarity:        c.Rarity().String(),
		Picture:       pictureURL,
		PictureThumb:  pictureThumbURL,
		PictureCard:   pictureCardURL,
		PictureLqip:   c.PictureLqip(),
		PictureStatus: c.PictureStatus().String(),
		Hamon:         c.Hamon().String(),
		Spin:          c.Spin().String(),
		BattleIQ:      int(c.BattleIQ()),
	}, nil
}

// NewJojoCharacterResponses builds a JojoCharacterResponse slice, never
// nil, from a list of domain JojoCharacters.
func NewJojoCharacterResponses(ctx context.Context, list []*characters.JojoCharacter, resolve PictureURLResolver, media MediaURLBuilder) ([]JojoCharacterResponse, error) {
	responses := make([]JojoCharacterResponse, 0, len(list))
	for _, c := range list {
		resp, err := NewJojoCharacterResponse(ctx, c, resolve, media)
		if err != nil {
			return nil, err
		}
		responses = append(responses, resp)
	}
	return responses, nil
}

// JojoCharacterPageResponse is GET /jojo-characters's response body when
// the request opts into pagination - see StandPageResponse's doc for the
// wire shape and why this is a concrete per-resource envelope.
type JojoCharacterPageResponse struct {
	PageInfo
	Items []JojoCharacterResponse `json:"items"`
}
