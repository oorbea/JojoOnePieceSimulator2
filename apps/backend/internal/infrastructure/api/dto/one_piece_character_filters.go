package dto

import (
	"fmt"
	"net/url"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// OnePieceCharacterFiltersFromQuery maps the optional ?rarity=
// &physicalForm=&armamentHaki=&observationHaki=&conquerorHaki=
// &fruitMastery=&q= query params onto ports.OnePieceCharacterFilters - same
// convention as DevilFruitFiltersFromQuery.
func OnePieceCharacterFiltersFromQuery(q url.Values) (ports.OnePieceCharacterFilters, bool, error) {
	var filters ports.OnePieceCharacterFilters
	var errs []string
	hasFilters := false

	if v := q.Get("rarity"); v != "" {
		hasFilters = true
		rarity, err := enums.ParsePowerRarity(v)
		if err != nil {
			errs = append(errs, fmt.Sprintf("rarity: %v", err))
		} else {
			filters.Rarity = &rarity
		}
	}
	if v := q.Get("physicalForm"); v != "" {
		hasFilters = true
		physicalForm, err := enums.ParsePhysicalForm(v)
		if err != nil {
			errs = append(errs, fmt.Sprintf("physicalForm: %v", err))
		} else {
			filters.PhysicalForm = &physicalForm
		}
	}
	if v := q.Get("armamentHaki"); v != "" {
		hasFilters = true
		haki, err := enums.ParseHakiLevel(v)
		if err != nil {
			errs = append(errs, fmt.Sprintf("armamentHaki: %v", err))
		} else {
			filters.ArmamentHaki = &haki
		}
	}
	if v := q.Get("observationHaki"); v != "" {
		hasFilters = true
		haki, err := enums.ParseHakiLevel(v)
		if err != nil {
			errs = append(errs, fmt.Sprintf("observationHaki: %v", err))
		} else {
			filters.ObservationHaki = &haki
		}
	}
	if v := q.Get("conquerorHaki"); v != "" {
		hasFilters = true
		haki, err := enums.ParseHakiLevel(v)
		if err != nil {
			errs = append(errs, fmt.Sprintf("conquerorHaki: %v", err))
		} else {
			filters.ConquerorHaki = &haki
		}
	}
	if v := q.Get("fruitMastery"); v != "" {
		hasFilters = true
		fruitMastery, err := enums.ParseFruitMastery(v)
		if err != nil {
			errs = append(errs, fmt.Sprintf("fruitMastery: %v", err))
		} else {
			filters.FruitMastery = &fruitMastery
		}
	}
	if v := q.Get("q"); v != "" {
		hasFilters = true
		filters.Search = &v
	}

	if len(errs) > 0 {
		return ports.OnePieceCharacterFilters{}, hasFilters, &ValidationError{Errors: errs}
	}
	return filters, hasFilters, nil
}
