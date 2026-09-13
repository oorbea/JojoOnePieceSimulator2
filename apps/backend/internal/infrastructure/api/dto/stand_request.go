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
	Name          string                        `json:"name"`
	Translations  map[string]TranslationRequest `json:"translations" ts:"map[Locale]"`
	Rarity        string                        `json:"rarity" ts:"PowerRarity"`
	AttackPower   string                        `json:"attackPower" ts:"StandStat"`
	Speed         string                        `json:"speed" ts:"StandStat"`
	AttackRange   string                        `json:"attackRange" ts:"StandStat"`
	Endurance     string                        `json:"endurance" ts:"StandStat"`
	Precision     string                        `json:"precision" ts:"StandStat"`
	Potential     string                        `json:"potential" ts:"StandStat"`
	EvolvesFromID *string                       `json:"evolvesFromId,omitempty"`
}

// ValidationError collects every field error found while validating a
// request, so the caller can report them all at once instead of stopping at
// the first one.
type ValidationError struct {
	Errors []FieldError
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed: %v", e.Errors)
}

// Validate converts the request into a services.StandInput, collecting all
// field errors before returning.
func (r StandRequest) Validate() (services.StandInput, error) {
	var errs []FieldError

	if r.Name == "" {
		errs = append(errs, FieldError{Field: "name", Code: ValNameRequired, Message: "name is required"})
	}
	translations, translationErrs := validateTranslations(r.Translations)
	errs = append(errs, translationErrs...)

	rarity, err := enums.ParsePowerRarity(r.Rarity)
	if err != nil {
		errs = append(errs, FieldError{Field: "rarity", Code: ValInvalidValue, Message: fmt.Sprintf("rarity: %v", err)})
	}

	attackPower, err := enums.ParseStandStat(r.AttackPower)
	if err != nil {
		errs = append(errs, FieldError{Field: "attackPower", Code: ValInvalidValue, Message: fmt.Sprintf("attackPower: %v", err)})
	}
	speed, err := enums.ParseStandStat(r.Speed)
	if err != nil {
		errs = append(errs, FieldError{Field: "speed", Code: ValInvalidValue, Message: fmt.Sprintf("speed: %v", err)})
	}
	attackRange, err := enums.ParseStandStat(r.AttackRange)
	if err != nil {
		errs = append(errs, FieldError{Field: "attackRange", Code: ValInvalidValue, Message: fmt.Sprintf("attackRange: %v", err)})
	}
	endurance, err := enums.ParseStandStat(r.Endurance)
	if err != nil {
		errs = append(errs, FieldError{Field: "endurance", Code: ValInvalidValue, Message: fmt.Sprintf("endurance: %v", err)})
	}
	precision, err := enums.ParseStandStat(r.Precision)
	if err != nil {
		errs = append(errs, FieldError{Field: "precision", Code: ValInvalidValue, Message: fmt.Sprintf("precision: %v", err)})
	}
	potential, err := enums.ParseStandStat(r.Potential)
	if err != nil {
		errs = append(errs, FieldError{Field: "potential", Code: ValInvalidValue, Message: fmt.Sprintf("potential: %v", err)})
	}

	var evolvesFrom *powers.PowerID
	if r.EvolvesFromID != nil {
		id, err := powers.ParsePowerID(*r.EvolvesFromID)
		if err != nil {
			errs = append(errs, FieldError{Field: "evolvesFromId", Code: ValInvalidValue, Message: fmt.Sprintf("evolvesFromId: %v", err)})
		} else {
			evolvesFrom = &id
		}
	}

	if len(errs) > 0 {
		return services.StandInput{}, &ValidationError{Errors: errs}
	}

	return services.StandInput{
		Name:         r.Name,
		Translations: translations,
		Rarity:       rarity,
		AttackPower:  attackPower,
		Speed:        speed,
		AttackRange:  attackRange,
		Endurance:    endurance,
		Precision:    precision,
		Potential:    potential,
		EvolvesFrom:  evolvesFrom,
	}, nil
}
