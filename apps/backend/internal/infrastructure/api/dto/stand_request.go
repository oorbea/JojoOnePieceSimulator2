package dto

import (
	"fmt"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// StandRequest is the JSON body accepted by POST and PUT /stands. Enum
// fields are plain strings so an invalid value becomes a 400 with a clear
// message instead of a JSON decode error.
type StandRequest struct {
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Rarity        string   `json:"rarity"`
	Skills        []string `json:"skills"`
	AttackPower   string   `json:"attackPower"`
	Speed         string   `json:"speed"`
	AttackRange   string   `json:"attackRange"`
	Endurance     string   `json:"endurance"`
	Precision     string   `json:"precision"`
	Potential     string   `json:"potential"`
	EvolvesFromID *string  `json:"evolvesFromId,omitempty"`
}

// ValidationError collects every field error found while validating a
// request, so the caller can report them all at once instead of stopping at
// the first one.
type ValidationError struct {
	Errors []string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed: %v", e.Errors)
}

// Validate converts the request into a services.StandInput, collecting all
// field errors before returning.
func (r StandRequest) Validate() (services.StandInput, error) {
	var errs []string

	if r.Name == "" {
		errs = append(errs, "name is required")
	}
	if r.Description == "" {
		errs = append(errs, "description is required")
	}
	if len(r.Skills) == 0 {
		errs = append(errs, "skills are required")
	}

	rarity, err := enums.ParsePowerRarity(r.Rarity)
	if err != nil {
		errs = append(errs, fmt.Sprintf("rarity: %v", err))
	}

	attackPower, err := enums.ParseStandStat(r.AttackPower)
	if err != nil {
		errs = append(errs, fmt.Sprintf("attackPower: %v", err))
	}
	speed, err := enums.ParseStandStat(r.Speed)
	if err != nil {
		errs = append(errs, fmt.Sprintf("speed: %v", err))
	}
	attackRange, err := enums.ParseStandStat(r.AttackRange)
	if err != nil {
		errs = append(errs, fmt.Sprintf("attackRange: %v", err))
	}
	endurance, err := enums.ParseStandStat(r.Endurance)
	if err != nil {
		errs = append(errs, fmt.Sprintf("endurance: %v", err))
	}
	precision, err := enums.ParseStandStat(r.Precision)
	if err != nil {
		errs = append(errs, fmt.Sprintf("precision: %v", err))
	}
	potential, err := enums.ParseStandStat(r.Potential)
	if err != nil {
		errs = append(errs, fmt.Sprintf("potential: %v", err))
	}

	var evolvesFrom *powers.PowerID
	if r.EvolvesFromID != nil {
		id, err := powers.ParsePowerID(*r.EvolvesFromID)
		if err != nil {
			errs = append(errs, fmt.Sprintf("evolvesFromId: %v", err))
		} else {
			evolvesFrom = &id
		}
	}

	if len(errs) > 0 {
		return services.StandInput{}, &ValidationError{Errors: errs}
	}

	skills := append([]string(nil), r.Skills...)
	return services.StandInput{
		Name:        r.Name,
		Description: r.Description,
		Rarity:      rarity,
		Skills:      &skills,
		AttackPower: attackPower,
		Speed:       speed,
		AttackRange: attackRange,
		Endurance:   endurance,
		Precision:   precision,
		Potential:   potential,
		EvolvesFrom: evolvesFrom,
	}, nil
}
