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
	Mode string `json:"mode"`
	// StageMangas and PowerMangas are independent - see game.Config's doc
	// comment.
	StageMangas            []string           `json:"stageMangas"`
	PowerMangas            []string           `json:"powerMangas"`
	AbilitySource          string             `json:"abilitySource"`
	TeamSize               int                `json:"teamSize"`
	AllowBots              bool               `json:"allowBots"`
	Visibility             string             `json:"visibility,omitempty"`
	VotingWindowSeconds    int                `json:"votingWindowSeconds,omitempty"`
	PoolFilter             *PoolFilterPayload `json:"poolFilter,omitempty"`
	RevealSpeed            string             `json:"revealSpeed,omitempty"`
	SummaryDurationSeconds int                `json:"summaryDurationSeconds,omitempty"`
}

// Validate converts the request into a services.CreateGameInput, collecting
// all field errors before returning.
func (r CreateGameRequest) Validate() (services.CreateGameInput, error) {
	var errs []string

	mode, err := enums.ParseGameModeKind(r.Mode)
	if err != nil {
		errs = append(errs, fmt.Sprintf("mode: %v", err))
	}

	stageMangas := parseMangas(r.StageMangas, "stageMangas", &errs)
	powerMangas := parseMangas(r.PowerMangas, "powerMangas", &errs)

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

	var revealSpeed enums.RevealSpeed
	if r.RevealSpeed != "" {
		revealSpeed, err = enums.ParseRevealSpeed(r.RevealSpeed)
		if err != nil {
			errs = append(errs, fmt.Sprintf("revealSpeed: %v", err))
		}
	}

	if len(errs) > 0 {
		return services.CreateGameInput{}, &ValidationError{Errors: errs}
	}

	return services.CreateGameInput{
		Mode:                   mode,
		StageMangas:            stageMangas,
		PowerMangas:            powerMangas,
		AbilitySource:          abilitySource,
		TeamSize:               r.TeamSize,
		AllowBots:              r.AllowBots,
		Visibility:             visibility,
		VotingWindowSeconds:    r.VotingWindowSeconds,
		PoolFilter:             poolFilter,
		RevealSpeed:            revealSpeed,
		SummaryDurationSeconds: r.SummaryDurationSeconds,
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
