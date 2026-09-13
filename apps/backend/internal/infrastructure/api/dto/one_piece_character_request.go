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
	var errs []FieldError

	if r.Name == "" {
		errs = append(errs, FieldError{Field: "name", Code: ValNameRequired, Message: "name is required"})
	}
	translations, translationErrs := validateCharacterTranslations(r.Translations)
	errs = append(errs, translationErrs...)

	rarity, err := enums.ParsePowerRarity(r.Rarity)
	if err != nil {
		errs = append(errs, FieldError{Field: "rarity", Code: ValInvalidValue, Message: fmt.Sprintf("rarity: %v", err)})
	}
	physicalForm, err := enums.ParsePhysicalForm(r.PhysicalForm)
	if err != nil {
		errs = append(errs, FieldError{Field: "physicalForm", Code: ValInvalidValue, Message: fmt.Sprintf("physicalForm: %v", err)})
	}
	armamentHaki, err := enums.ParseHakiLevel(r.ArmamentHaki)
	if err != nil {
		errs = append(errs, FieldError{Field: "armamentHaki", Code: ValInvalidValue, Message: fmt.Sprintf("armamentHaki: %v", err)})
	}
	observationHaki, err := enums.ParseHakiLevel(r.ObservationHaki)
	if err != nil {
		errs = append(errs, FieldError{Field: "observationHaki", Code: ValInvalidValue, Message: fmt.Sprintf("observationHaki: %v", err)})
	}
	conquerorHaki, err := enums.ParseHakiLevel(r.ConquerorHaki)
	if err != nil {
		errs = append(errs, FieldError{Field: "conquerorHaki", Code: ValInvalidValue, Message: fmt.Sprintf("conquerorHaki: %v", err)})
	}
	fruitMastery, err := enums.ParseFruitMastery(r.FruitMastery)
	if err != nil {
		errs = append(errs, FieldError{Field: "fruitMastery", Code: ValInvalidValue, Message: fmt.Sprintf("fruitMastery: %v", err)})
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
