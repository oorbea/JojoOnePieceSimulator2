package dto

import "github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"

// StageResponse is the JSON representation of a Stage.
type StageResponse struct {
	ID    string `json:"id"`
	Manga string `json:"manga"`
	Order int    `json:"order"`
	Name  string `json:"name"`
}

// NewStageResponse builds a StageResponse from a domain Stage.
func NewStageResponse(s game.Stage) StageResponse {
	return StageResponse{
		ID:    s.ID().String(),
		Manga: s.Manga().String(),
		Order: s.Order(),
		Name:  s.Name(),
	}
}

// NewStageResponses builds a StageResponse slice, never nil, from a list of
// domain Stages.
func NewStageResponses(stages []game.Stage) []StageResponse {
	responses := make([]StageResponse, 0, len(stages))
	for _, s := range stages {
		responses = append(responses, NewStageResponse(s))
	}
	return responses
}
