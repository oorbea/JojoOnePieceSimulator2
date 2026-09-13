package dto

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// JojoCharacterFiltersFromQuery maps the optional ?rarity=&hamon=&spin=
// &battleIq=&q= query params onto ports.JojoCharacterFilters - same
// convention as DevilFruitFiltersFromQuery.
func JojoCharacterFiltersFromQuery(q url.Values) (ports.JojoCharacterFilters, bool, error) {
	var filters ports.JojoCharacterFilters
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
	if v := q.Get("hamon"); v != "" {
		hasFilters = true
		hamon, err := enums.ParseHamonLevel(v)
		if err != nil {
			errs = append(errs, FieldError{Field: "hamon", Code: ValInvalidValue, Message: fmt.Sprintf("hamon: %v", err)})
		} else {
			filters.Hamon = &hamon
		}
	}
	if v := q.Get("spin"); v != "" {
		hasFilters = true
		spin, err := enums.ParseSpinLevel(v)
		if err != nil {
			errs = append(errs, FieldError{Field: "spin", Code: ValInvalidValue, Message: fmt.Sprintf("spin: %v", err)})
		} else {
			filters.Spin = &spin
		}
	}
	if v := q.Get("battleIq"); v != "" {
		hasFilters = true
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 || n > 255 {
			errs = append(errs, FieldError{Field: "battleIq", Code: ValBattleIqRange, Message: "battleIq: must be between 0 and 255"})
		} else {
			battleIQ := byte(n)
			filters.BattleIQ = &battleIQ
		}
	}
	if v := q.Get("q"); v != "" {
		hasFilters = true
		filters.Search = &v
	}

	if len(errs) > 0 {
		return ports.JojoCharacterFilters{}, hasFilters, &ValidationError{Errors: errs}
	}
	return filters, hasFilters, nil
}
