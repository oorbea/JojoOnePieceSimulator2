package dto

import (
	"fmt"
	"strings"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// CreateGameRequest is the JSON body accepted by POST /games.
// Visibility/VotingWindowSeconds/PoolFilter are all optional: an empty
// Visibility defaults to PRIVATE, a zero VotingWindowSeconds defaults to
// the service's configured window, and an absent PoolFilter means "no
// restriction" - see services.GameService.buildConfig.
type CreateGameRequest struct {
	Mode                string             `json:"mode"`
	Mangas              []string           `json:"mangas"`
	AbilitySource       string             `json:"abilitySource"`
	TeamSize            int                `json:"teamSize"`
	AllowBots           bool               `json:"allowBots"`
	Visibility          string             `json:"visibility,omitempty"`
	VotingWindowSeconds int                `json:"votingWindowSeconds,omitempty"`
	PoolFilter          *PoolFilterPayload `json:"poolFilter,omitempty"`
}

// Validate converts the request into a services.CreateGameInput, collecting
// all field errors before returning.
func (r CreateGameRequest) Validate() (services.CreateGameInput, error) {
	var errs []string

	mode, err := enums.ParseGameModeKind(r.Mode)
	if err != nil {
		errs = append(errs, fmt.Sprintf("mode: %v", err))
	}

	mangas := make([]enums.Manga, 0, len(r.Mangas))
	for _, raw := range r.Mangas {
		m, err := enums.ParseManga(raw)
		if err != nil {
			errs = append(errs, fmt.Sprintf("mangas: %v", err))
			continue
		}
		mangas = append(mangas, m)
	}

	abilitySource, err := enums.ParseAbilitySource(r.AbilitySource)
	if err != nil {
		errs = append(errs, fmt.Sprintf("abilitySource: %v", err))
	}

	if r.TeamSize <= 0 {
		errs = append(errs, "teamSize must be positive")
	}

	var visibility enums.LobbyVisibility
	if r.Visibility != "" {
		visibility, err = enums.ParseLobbyVisibility(r.Visibility)
		if err != nil {
			errs = append(errs, fmt.Sprintf("visibility: %v", err))
		}
	}
	poolFilter := r.PoolFilter.ToPoolFilter(&errs)

	if len(errs) > 0 {
		return services.CreateGameInput{}, &ValidationError{Errors: errs}
	}

	return services.CreateGameInput{
		Mode:                mode,
		Mangas:              mangas,
		AbilitySource:       abilitySource,
		TeamSize:            r.TeamSize,
		AllowBots:           r.AllowBots,
		Visibility:          visibility,
		VotingWindowSeconds: r.VotingWindowSeconds,
		PoolFilter:          poolFilter,
	}, nil
}

// JoinGameRequest is the JSON body accepted by POST /games/join.
type JoinGameRequest struct {
	Code string `json:"code"`
}

// Validate normalizes and checks the join code's shape. The actual
// existence check happens in GameService.JoinByCode.
func (r JoinGameRequest) Validate() (string, error) {
	code := strings.ToUpper(strings.TrimSpace(r.Code))
	if code == "" {
		return "", &ValidationError{Errors: []string{"code is required"}}
	}
	return code, nil
}
