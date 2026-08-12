package dto

import (
	"fmt"
	"net/url"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// StandFiltersFromQuery maps the optional ?rarity=&attackPower=&speed=&
// attackRange=&endurance=&precision=&potential=&evolvesFrom=&q= query
// params onto ports.StandFilters. A param that is absent leaves its pointer
// nil, so the filter is skipped. HasFilters reports whether any param was
// set at all.
func StandFiltersFromQuery(q url.Values) (ports.StandFilters, bool, error) {
	var filters ports.StandFilters
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
	if v := q.Get("attackPower"); v != "" {
		hasFilters = true
		stat, err := enums.ParseStandStat(v)
		if err != nil {
			errs = append(errs, fmt.Sprintf("attackPower: %v", err))
		} else {
			filters.AttackPower = &stat
		}
	}
	if v := q.Get("speed"); v != "" {
		hasFilters = true
		stat, err := enums.ParseStandStat(v)
		if err != nil {
			errs = append(errs, fmt.Sprintf("speed: %v", err))
		} else {
			filters.Speed = &stat
		}
	}
	if v := q.Get("attackRange"); v != "" {
		hasFilters = true
		stat, err := enums.ParseStandStat(v)
		if err != nil {
			errs = append(errs, fmt.Sprintf("attackRange: %v", err))
		} else {
			filters.AttackRange = &stat
		}
	}
	if v := q.Get("endurance"); v != "" {
		hasFilters = true
		stat, err := enums.ParseStandStat(v)
		if err != nil {
			errs = append(errs, fmt.Sprintf("endurance: %v", err))
		} else {
			filters.Endurance = &stat
		}
	}
	if v := q.Get("precision"); v != "" {
		hasFilters = true
		stat, err := enums.ParseStandStat(v)
		if err != nil {
			errs = append(errs, fmt.Sprintf("precision: %v", err))
		} else {
			filters.Precision = &stat
		}
	}
	if v := q.Get("potential"); v != "" {
		hasFilters = true
		stat, err := enums.ParseStandStat(v)
		if err != nil {
			errs = append(errs, fmt.Sprintf("potential: %v", err))
		} else {
			filters.Potential = &stat
		}
	}
	if v := q.Get("evolvesFrom"); v != "" {
		hasFilters = true
		filters.EvolvesFrom = &v
	}
	if v := q.Get("q"); v != "" {
		hasFilters = true
		filters.Search = &v
	}

	if len(errs) > 0 {
		return ports.StandFilters{}, hasFilters, &ValidationError{Errors: errs}
	}
	return filters, hasFilters, nil
}
