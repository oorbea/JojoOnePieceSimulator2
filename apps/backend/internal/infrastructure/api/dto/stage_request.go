package dto

import (
	"fmt"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// StageRequest is the JSON body accepted by POST and PUT /stages. Manga is a
// plain string so an invalid value becomes a 400 with a clear message
// instead of a JSON decode error - same convention as every other catalogue
// request DTO.
type StageRequest struct {
	Manga        string                             `json:"manga" ts:"Manga"`
	Order        int                                `json:"order"`
	Name         string                             `json:"name"`
	Translations map[string]StageTranslationRequest `json:"translations" ts:"map[Locale]"`
}

// Validate converts the request into a services.StageInput, collecting all
// field errors before returning.
func (r StageRequest) Validate() (services.StageInput, error) {
	var errs []FieldError

	manga, err := enums.ParseManga(r.Manga)
	if err != nil {
		errs = append(errs, FieldError{Field: "manga", Code: ValInvalidValue, Message: fmt.Sprintf("manga: %v", err)})
	}
	if r.Order < 0 {
		errs = append(errs, FieldError{Field: "order", Code: ValOrderNonNegative, Message: "order must be non-negative"})
	}
	if r.Name == "" {
		errs = append(errs, FieldError{Field: "name", Code: ValNameRequired, Message: "name is required"})
	}
	translations, translationErrs := validateStageTranslations(r.Translations)
	errs = append(errs, translationErrs...)

	if len(errs) > 0 {
		return services.StageInput{}, &ValidationError{Errors: errs}
	}

	return services.StageInput{Manga: manga, Order: r.Order, Name: r.Name, Translations: translations}, nil
}
