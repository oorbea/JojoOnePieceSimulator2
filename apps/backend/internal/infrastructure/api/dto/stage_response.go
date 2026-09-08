package dto

import (
	"context"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
)

// StageResponse is the JSON representation of a Stage.
type StageResponse struct {
	ID            string `json:"id"`
	Manga         string `json:"manga" ts:"Manga"`
	Order         int    `json:"order"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Picture       string `json:"picture"`
	PictureThumb  string `json:"pictureThumb"`
	PictureCard   string `json:"pictureCard"`
	PictureStatus string `json:"pictureStatus" ts:"PictureStatus"`
	PictureLqip   string `json:"pictureLqip"`
}

// NewStageResponse builds a StageResponse from a domain Stage, resolving its
// picture key through resolve - same contract as NewStandResponse.
func NewStageResponse(ctx context.Context, s game.Stage, resolve PictureURLResolver, media MediaURLBuilder) (StageResponse, error) {
	pictureURL, pictureThumbURL, pictureCardURL, err := resolveCatalogPictures(
		ctx, s.Picture(), s.PictureThumb(), s.PictureCard(), s.PictureMediaID(), resolve, media,
	)
	if err != nil {
		return StageResponse{}, err
	}

	return StageResponse{
		ID:            s.ID().String(),
		Manga:         s.Manga().String(),
		Order:         s.Order(),
		Name:          s.Name(),
		Description:   s.Description(),
		Picture:       pictureURL,
		PictureThumb:  pictureThumbURL,
		PictureCard:   pictureCardURL,
		PictureLqip:   s.PictureLqip(),
		PictureStatus: s.PictureStatus().String(),
	}, nil
}

// NewStageResponses builds a StageResponse slice, never nil, from a list of
// domain Stages.
func NewStageResponses(ctx context.Context, stages []game.Stage, resolve PictureURLResolver, media MediaURLBuilder) ([]StageResponse, error) {
	responses := make([]StageResponse, 0, len(stages))
	for _, s := range stages {
		resp, err := NewStageResponse(ctx, s, resolve, media)
		if err != nil {
			return nil, err
		}
		responses = append(responses, resp)
	}
	return responses, nil
}

// StagePageResponse is GET /stages's response body when the request opts
// into pagination - see StandPageResponse's doc for the wire shape and why
// this is a concrete per-resource envelope.
type StagePageResponse struct {
	PageInfo
	Items []StageResponse `json:"items"`
}
