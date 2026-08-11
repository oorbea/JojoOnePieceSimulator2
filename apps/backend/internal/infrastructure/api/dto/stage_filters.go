package dto

import (
	"fmt"
	"net/url"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// StageFiltersFromQuery maps the optional ?manga=&q= query params onto
// ports.StageFilters. A param that is absent leaves its pointer nil, so the
// filter is skipped. The bool reports whether any param was set at all.
func StageFiltersFromQuery(q url.Values) (ports.StageFilters, bool, error) {
	var filters ports.StageFilters
	var errs []string
	hasFilters := false

	if v := q.Get("manga"); v != "" {
		hasFilters = true
		manga, err := enums.ParseManga(v)
		if err != nil {
			errs = append(errs, fmt.Sprintf("manga: %v", err))
		} else {
			filters.Manga = &manga
		}
	}
	if v := q.Get("q"); v != "" {
		hasFilters = true
		filters.Search = &v
	}

	if len(errs) > 0 {
		return ports.StageFilters{}, hasFilters, &ValidationError{Errors: errs}
	}
	return filters, hasFilters, nil
}
