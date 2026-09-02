package dto

import (
	"fmt"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// DevilFruitRequest is the JSON body accepted by POST and PUT /devil-fruits.
// Enum fields are plain strings so an invalid value becomes a 400 with a
// clear message instead of a JSON decode error.
type DevilFruitRequest struct {
	Name         string                        `json:"name"`
	Translations map[string]TranslationRequest `json:"translations" ts:"map[Locale]"`
	Rarity       string                        `json:"rarity" ts:"PowerRarity"`
	FruitType    string                        `json:"fruitType" ts:"FruitType"`
}

// Validate converts the request into a services.DevilFruitInput, collecting
// all field errors before returning.
func (r DevilFruitRequest) Validate() (services.DevilFruitInput, error) {
	var errs []string

	if r.Name == "" {
		errs = append(errs, "name is required")
	}
	translations, translationErrs := validateTranslations(r.Translations)
	errs = append(errs, translationErrs...)

	rarity, err := enums.ParsePowerRarity(r.Rarity)
	if err != nil {
		errs = append(errs, fmt.Sprintf("rarity: %v", err))
	}

	fruitType, err := enums.ParseFruitType(r.FruitType)
	if err != nil {
		errs = append(errs, fmt.Sprintf("fruitType: %v", err))
	}

	if len(errs) > 0 {
		return services.DevilFruitInput{}, &ValidationError{Errors: errs}
	}

	return services.DevilFruitInput{
		Name:         r.Name,
		Translations: translations,
		Rarity:       rarity,
		FruitType:    fruitType,
	}, nil
}
