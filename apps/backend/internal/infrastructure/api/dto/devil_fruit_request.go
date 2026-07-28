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
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Rarity      string   `json:"rarity"`
	Skills      []string `json:"skills"`
	FruitType   string   `json:"fruitType"`
}

// Validate converts the request into a services.DevilFruitInput, collecting
// all field errors before returning.
func (r DevilFruitRequest) Validate() (services.DevilFruitInput, error) {
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

	fruitType, err := enums.ParseFruitType(r.FruitType)
	if err != nil {
		errs = append(errs, fmt.Sprintf("fruitType: %v", err))
	}

	if len(errs) > 0 {
		return services.DevilFruitInput{}, &ValidationError{Errors: errs}
	}

	skills := append([]string(nil), r.Skills...)
	return services.DevilFruitInput{
		Name:        r.Name,
		Description: r.Description,
		Rarity:      rarity,
		Skills:      &skills,
		FruitType:   fruitType,
	}, nil
}
