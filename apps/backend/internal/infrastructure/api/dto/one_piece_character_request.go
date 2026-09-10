package dto

import (
	"fmt"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// OnePieceCharacterRequest is the JSON body accepted by POST and PUT
// /one-piece-characters - same convention as JojoCharacterRequest.
type OnePieceCharacterRequest struct {
	Name            string                                 `json:"name"`
	Translations    map[string]CharacterTranslationRequest `json:"translations" ts:"map[Locale]"`
	Rarity          string                                 `json:"rarity" ts:"PowerRarity"`
	PhysicalForm    string                                 `json:"physicalForm" ts:"PhysicalForm"`
	ArmamentHaki    string                                 `json:"armamentHaki" ts:"HakiLevel"`
	ObservationHaki string                                 `json:"observationHaki" ts:"HakiLevel"`
	ConquerorHaki   string                                 `json:"conquerorHaki" ts:"HakiLevel"`
	FruitMastery    string                                 `json:"fruitMastery" ts:"FruitMastery"`
}

// Validate converts the request into a services.OnePieceCharacterInput,
// collecting all field errors before returning.
func (r OnePieceCharacterRequest) Validate() (services.OnePieceCharacterInput, error) {
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
	physicalForm, err := enums.ParsePhysicalForm(r.PhysicalForm)
	if err != nil {
		errs = append(errs, fmt.Sprintf("physicalForm: %v", err))
	}
	armamentHaki, err := enums.ParseHakiLevel(r.ArmamentHaki)
	if err != nil {
		errs = append(errs, fmt.Sprintf("armamentHaki: %v", err))
	}
	observationHaki, err := enums.ParseHakiLevel(r.ObservationHaki)
	if err != nil {
		errs = append(errs, fmt.Sprintf("observationHaki: %v", err))
	}
	conquerorHaki, err := enums.ParseHakiLevel(r.ConquerorHaki)
	if err != nil {
		errs = append(errs, fmt.Sprintf("conquerorHaki: %v", err))
	}
	fruitMastery, err := enums.ParseFruitMastery(r.FruitMastery)
	if err != nil {
		errs = append(errs, fmt.Sprintf("fruitMastery: %v", err))
	}

	if len(errs) > 0 {
		return services.OnePieceCharacterInput{}, &ValidationError{Errors: errs}
	}

	return services.OnePieceCharacterInput{
		Name:            r.Name,
		Translations:    translations,
		Rarity:          rarity,
		PhysicalForm:    physicalForm,
		ArmamentHaki:    armamentHaki,
		ObservationHaki: observationHaki,
		ConquerorHaki:   conquerorHaki,
		FruitMastery:    fruitMastery,
	}, nil
}
