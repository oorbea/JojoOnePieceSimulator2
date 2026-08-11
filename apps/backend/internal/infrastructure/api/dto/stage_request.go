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
	Manga string `json:"manga"`
	Order int    `json:"order"`
	Name  string `json:"name"`
}

// Validate converts the request into a services.StageInput, collecting all
// field errors before returning.
func (r StageRequest) Validate() (services.StageInput, error) {
	var errs []string

	manga, err := enums.ParseManga(r.Manga)
	if err != nil {
		errs = append(errs, fmt.Sprintf("manga: %v", err))
	}
	if r.Order < 0 {
		errs = append(errs, "order must be non-negative")
	}
	if r.Name == "" {
		errs = append(errs, "name is required")
	}

	if len(errs) > 0 {
		return services.StageInput{}, &ValidationError{Errors: errs}
	}

	return services.StageInput{Manga: manga, Order: r.Order, Name: r.Name}, nil
}
