package dto

import (
	"fmt"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// JojoCharacterRequest is the JSON body accepted by POST and PUT
// /jojo-characters. Enum fields are plain strings - same convention as
// every other catalogue request DTO. BattleIQ is a plain int (not a byte -
// JSON has no byte type) validated into [0, 255].
type JojoCharacterRequest struct {
	Name         string                                 `json:"name"`
	Translations map[string]CharacterTranslationRequest `json:"translations" ts:"map[Locale]"`
	Rarity       string                                 `json:"rarity" ts:"PowerRarity"`
	Hamon        string                                 `json:"hamon" ts:"HamonLevel"`
	Spin         string                                 `json:"spin" ts:"SpinLevel"`
	BattleIQ     int                                    `json:"battleIq"`
}

// Validate converts the request into a services.JojoCharacterInput,
// collecting all field errors before returning.
func (r JojoCharacterRequest) Validate() (services.JojoCharacterInput, error) {
	var errs []string

	if r.Name == "" {
		errs = append(errs, "name is required")
	}
	translations, translationErrs := validateCharacterTranslations(r.Translations)
	errs = append(errs, translationErrs...)

	rarity, err := enums.ParsePowerRarity(r.Rarity)
	if err != nil {
		errs = append(errs, fmt.Sprintf("rarity: %v", err))
	}
	hamon, err := enums.ParseHamonLevel(r.Hamon)
	if err != nil {
		errs = append(errs, fmt.Sprintf("hamon: %v", err))
	}
	spin, err := enums.ParseSpinLevel(r.Spin)
	if err != nil {
		errs = append(errs, fmt.Sprintf("spin: %v", err))
	}
	if r.BattleIQ < 0 || r.BattleIQ > 255 {
		errs = append(errs, "battleIq: must be between 0 and 255")
	}

	if len(errs) > 0 {
		return services.JojoCharacterInput{}, &ValidationError{Errors: errs}
	}

	return services.JojoCharacterInput{
		Name:         r.Name,
		Translations: translations,
		Rarity:       rarity,
		Hamon:        hamon,
		Spin:         spin,
		BattleIQ:     byte(r.BattleIQ),
	}, nil
}
