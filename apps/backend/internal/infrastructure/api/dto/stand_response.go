package dto

import (
	"context"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// StandResponse is the JSON representation of a Stand, recursively nesting
// its evolves_from ancestor chain (which the repository already loads in a
// single round trip).
type StandResponse struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	Rarity        string         `json:"rarity" ts:"PowerRarity"`
	Skills        []string       `json:"skills"`
	Picture       string         `json:"picture"`
	PictureThumb  string         `json:"pictureThumb"`
	PictureCard   string         `json:"pictureCard"`
	PictureStatus string         `json:"pictureStatus" ts:"PictureStatus"`
	PictureLqip   string         `json:"pictureLqip"`
	AttackPower   string         `json:"attackPower" ts:"StandStat"`
	Speed         string         `json:"speed" ts:"StandStat"`
	AttackRange   string         `json:"attackRange" ts:"StandStat"`
	Endurance     string         `json:"endurance" ts:"StandStat"`
	Precision     string         `json:"precision" ts:"StandStat"`
	Potential     string         `json:"potential" ts:"StandStat"`
	EvolvesFrom   *StandResponse `json:"evolvesFrom"`
}

// PictureURLResolver turns a Stand's stored picture key into a URL a client
// can GET, returning "" for a Stand with no picture. Only used as a fallback
// while PictureMediaID is empty (not backfilled yet) - see
// resolveCatalogPictures.
type PictureURLResolver func(ctx context.Context, key string) (string, error)

// resolveCatalogPictures resolves a public catalogue subject's three
// renditions: pure string building through media (no I/O, no error) when
// mediaID is set, falling back to resolve (a presign, hitting Redis/S3) when
// it's still empty - the two-path branch every NewXResponse in this package
// shares. Shared by Stand/DevilFruit/Stage; User avatars use the private
// (signed) counterpart instead - see resolveAvatar in user_response.go.
func resolveCatalogPictures(ctx context.Context, main, thumb, card, mediaID string, resolve PictureURLResolver, media MediaURLBuilder) (mainURL, thumbURL, cardURL string, err error) {
	if mediaID != "" {
		return media.Public(mediaID, "main"), media.Public(mediaID, "thumb"), media.Public(mediaID, "card"), nil
	}
	mainURL, err = resolve(ctx, main)
	if err != nil {
		return "", "", "", err
	}
	thumbURL, err = resolve(ctx, thumb)
	if err != nil {
		return "", "", "", err
	}
	cardURL, err = resolve(ctx, card)
	if err != nil {
		return "", "", "", err
	}
	return mainURL, thumbURL, cardURL, nil
}

// NewStandResponse builds a StandResponse from a domain Stand, resolving its
// picture key (and, recursively, its evolves_from chain's) through resolve,
// or through media once PictureMediaID is backfilled - see
// resolveCatalogPictures.
func NewStandResponse(ctx context.Context, stand *powers.Stand, resolve PictureURLResolver, media MediaURLBuilder) (StandResponse, error) {
	var evolvesFrom *StandResponse
	if parent := stand.EvolvesFrom(); parent != nil {
		resp, err := NewStandResponse(ctx, parent, resolve, media)
		if err != nil {
			return StandResponse{}, err
		}
		evolvesFrom = &resp
	}

	skills := stand.Skills()
	if skills == nil {
		skills = []string{}
	}

	pictureURL, pictureThumbURL, pictureCardURL, err := resolveCatalogPictures(
		ctx, stand.Picture(), stand.PictureThumb(), stand.PictureCard(), stand.PictureMediaID(), resolve, media,
	)
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
		PictureCard:   pictureCardURL,
		PictureLqip:   stand.PictureLqip(),
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
func NewStandResponses(ctx context.Context, stands []*powers.Stand, resolve PictureURLResolver, media MediaURLBuilder) ([]StandResponse, error) {
	responses := make([]StandResponse, 0, len(stands))
	for _, stand := range stands {
		resp, err := NewStandResponse(ctx, stand, resolve, media)
		if err != nil {
			return nil, err
		}
		responses = append(responses, resp)
	}
	return responses, nil
}

// StandOptionResponse is the id/name pair the evolvesFrom picker needs - no
// picture, no stats, no translations.
type StandOptionResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// NewStandOptionResponses builds a StandOptionResponse slice, never nil,
// from a list of domain StandOptions.
func NewStandOptionResponses(options []ports.StandOption) []StandOptionResponse {
	responses := make([]StandOptionResponse, 0, len(options))
	for _, o := range options {
		responses = append(responses, StandOptionResponse{ID: o.ID.String(), Name: o.Name})
	}
	return responses
}
