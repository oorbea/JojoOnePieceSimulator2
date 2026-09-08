package dto

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// defaultPageSize/maxPageSize are Go-side constants rather than env vars for
// this first pagination adopter (Stand only) - see
// ObsidianVault/catalogue-pagination.md for why: the plan called for
// CATALOGUE_PAGE_SIZE/CATALOGUE_MAX_PAGE_SIZE, but that's another pair of
// env vars the owner would need to add to .env/.env.example/GitHub Variables
// before anything works, for a value with no operational reason to change
// per-deploy. 24 divides evenly into the frontend grid's 2/3/4/6-column
// breakpoints.
const (
	defaultPageSize = 24
	maxPageSize     = 100
)

// PageInfo is the paging half of a resource's page envelope (see
// StandPageResponse) - embedded, not a wire type on its own, so it must be
// in cmd/typegen/registry.go's nonWireTypeNames or the generator's AST scan
// fails. NextCursor absent means end of list; Total is only ever set on the
// first page (omitting it on later pages saves a COUNT(*) most callers
// already have).
type PageInfo struct {
	NextCursor *string `json:"nextCursor,omitempty"`
	Total      *int    `json:"total,omitempty"`
}

func NewPageInfo(nextCursor *string, total *int) PageInfo {
	return PageInfo{NextCursor: nextCursor, Total: total}
}

// StandCursor is the decoded shape of a Stand page cursor's `k` field.
type StandCursor struct {
	Name string `json:"name"`
}

type cursorEnvelope[K any] struct {
	V int    `json:"v"`
	F string `json:"f"`
	K K      `json:"k"`
}

const cursorVersion = 1

// EncodeCursor renders an opaque, base64url cursor carrying key K and a
// fingerprint binding it to the exact filter set it was issued under - a
// cursor reproduced against a different filter combination decodes fine but
// fails DecodeCursor's fingerprint check, returning 400 instead of
// silently-wrong results.
func EncodeCursor[K any](key K, fingerprint string) string {
	env := cursorEnvelope[K]{V: cursorVersion, F: fingerprint, K: key}
	data, err := json.Marshal(env)
	if err != nil {
		// K is always a small struct of strings - Marshal cannot fail for the
		// types this is actually called with.
		panic(fmt.Sprintf("encoding cursor: %v", err))
	}
	return base64.RawURLEncoding.EncodeToString(data)
}

// DecodeCursor reverses EncodeCursor, verifying both the version and the
// fingerprint against the filters the request is actually using.
func DecodeCursor[K any](cursor string, expectedFingerprint string) (K, error) {
	var zero K
	data, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return zero, &ValidationError{Errors: []string{"cursor: malformed"}}
	}
	var env cursorEnvelope[K]
	if err := json.Unmarshal(data, &env); err != nil {
		return zero, &ValidationError{Errors: []string{"cursor: malformed"}}
	}
	if env.V != cursorVersion {
		return zero, &ValidationError{Errors: []string{"cursor: unsupported version"}}
	}
	if env.F != expectedFingerprint {
		return zero, &ValidationError{Errors: []string{"cursor: does not match the current filters"}}
	}
	return env.K, nil
}

// FilterFingerprint hashes a canonical filter rendering down to a short,
// URL-safe token embedded in every cursor for that filter set.
func FilterFingerprint(canonical string) string {
	sum := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(sum[:8])
}

// standFiltersCanonical renders a ports.StandFilters in the same fixed field
// order cache/keys.go's standFilterKey uses. Deliberately duplicated rather
// than importing cache from dto (no cycle either way today, but dto is the
// lower-level package) - see ObsidianVault/catalogue-pagination.md for the
// anti-drift follow-up (a shared Canonical() method on StandFilters itself)
// this duplication should eventually become.
func standFiltersCanonical(f ports.StandFilters) string {
	str := func(v *string) string {
		if v == nil {
			return ""
		}
		return *v
	}
	parts := []string{
		str(f.EvolvesFrom),
		str(f.Search),
	}
	if f.Rarity != nil {
		parts = append(parts, f.Rarity.String())
	} else {
		parts = append(parts, "")
	}
	if f.AttackPower != nil {
		parts = append(parts, f.AttackPower.String())
	} else {
		parts = append(parts, "")
	}
	if f.Speed != nil {
		parts = append(parts, f.Speed.String())
	} else {
		parts = append(parts, "")
	}
	if f.AttackRange != nil {
		parts = append(parts, f.AttackRange.String())
	} else {
		parts = append(parts, "")
	}
	if f.Endurance != nil {
		parts = append(parts, f.Endurance.String())
	} else {
		parts = append(parts, "")
	}
	if f.Precision != nil {
		parts = append(parts, f.Precision.String())
	} else {
		parts = append(parts, "")
	}
	if f.Potential != nil {
		parts = append(parts, f.Potential.String())
	} else {
		parts = append(parts, "")
	}
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += "|"
		}
		out += p
	}
	return out
}

// StandFiltersFingerprint is the fingerprint a Stand page cursor is bound
// to, given the request's current filters and locale.
func StandFiltersFingerprint(filters ports.StandFilters, locale fmt.Stringer) string {
	return FilterFingerprint(standFiltersCanonical(filters) + "|" + locale.String())
}

// PageParams is the parsed, validated ?limit=&cursor=&total= query params
// common to every paginated list endpoint.
type PageParams struct {
	// Requested reports whether the caller opted into pagination at all
	// (either param present) - a request with neither gets the legacy bare
	// array response, unchanged.
	Requested bool
	Limit     int
	Cursor    string
	WithTotal bool
	HasCursor bool
}

// PageParamsFromQuery parses ?limit=&cursor=&total= - limit defaults to
// defaultPageSize when absent, is clamped silently to [1, maxPageSize], and
// a non-numeric limit is a 400 (not silently clamped, since that's a client
// bug worth surfacing rather than masking).
func PageParamsFromQuery(q url.Values) (PageParams, error) {
	var errs []string
	params := PageParams{WithTotal: true}

	limitStr := q.Get("limit")
	cursor := q.Get("cursor")
	params.Cursor = cursor
	params.HasCursor = cursor != ""
	params.Requested = limitStr != "" || params.HasCursor

	if limitStr == "" {
		params.Limit = defaultPageSize
	} else {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			errs = append(errs, "limit: must be a number")
		} else if limit < 1 {
			errs = append(errs, "limit: must be at least 1")
		} else {
			if limit > maxPageSize {
				limit = maxPageSize
			}
			params.Limit = limit
		}
	}

	// total is opt-out (?total=false) since it's a cheap COUNT(*) most
	// callers want on the first page.
	if v := q.Get("total"); v == "false" {
		params.WithTotal = false
	}

	if len(errs) > 0 {
		return PageParams{}, &ValidationError{Errors: errs}
	}
	return params, nil
}
