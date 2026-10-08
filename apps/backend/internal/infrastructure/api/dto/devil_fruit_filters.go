package dto

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// DevilFruitFiltersFromQuery maps the optional ?rarity=&fruitType=&q= query
// params onto ports.DevilFruitFilters. A param that is absent leaves its
// pointer nil, so the filter is skipped. The bool reports whether any param
// was set at all.
func DevilFruitFiltersFromQuery(q url.Values) (ports.DevilFruitFilters, bool, error) {
	var filters ports.DevilFruitFilters
	var errs []FieldError
	hasFilters := false

	if v := q.Get("rarity"); v != "" {
		hasFilters = true
		rarity, err := enums.ParsePowerRarity(v)
		if err != nil {
			errs = append(errs, FieldError{Field: "rarity", Code: ValInvalidValue, Message: fmt.Sprintf("rarity: %v", err)})
		} else {
			filters.Rarity = &rarity
		}
	}
	if v := q.Get("fruitType"); v != "" {
		hasFilters = true
		fruitType, err := enums.ParseFruitType(v)
		if err != nil {
			errs = append(errs, FieldError{Field: "fruitType", Code: ValInvalidValue, Message: fmt.Sprintf("fruitType: %v", err)})
		} else {
			filters.FruitType = &fruitType
		}
	}
	if v := q.Get("q"); v != "" {
		hasFilters = true
		filters.Search = &v
	}
	if v := q.Get("hasPicture"); v != "" {
		hasFilters = true
		b, err := strconv.ParseBool(v)
		if err != nil || (v != "true" && v != "false") {
			errs = append(errs, FieldError{Field: "hasPicture", Code: ValInvalidValue, Message: "hasPicture: must be true or false"})
		} else {
			filters.HasPicture = &b
		}
	}

	if len(errs) > 0 {
		return ports.DevilFruitFilters{}, hasFilters, &ValidationError{Errors: errs}
	}
	return filters, hasFilters, nil
}
