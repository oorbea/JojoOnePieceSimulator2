package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// standsNamespace holds every cached Stand read (FindByID, FindByName,
// GetAll, Filter) and is invalidated as a whole on any write - a Stand
// appears inside "all", inside an unknown set of filter:* entries, and
// nested in the EvolvesFrom chain of any descendant, so enumerating the
// affected keys is not tractable; flush-all is both correct and cheap for an
// admin-write catalogue.
const standsNamespace = "stands"

// devilFruitsNamespace holds every cached DevilFruit read (FindByID,
// FindByName, GetAll, Filter) and is invalidated as a whole on any write -
// same reasoning as standsNamespace.
const devilFruitsNamespace = "devil_fruits"

// presignNamespace holds cached presigned picture URLs, keyed by object
// storage key. Never invalidated wholesale - entries are evicted
// individually (on Delete) or simply expire.
const presignNamespace = "presign"

// Every key below is prefixed with locale so a write's whole-namespace
// Invalidate still clears every locale's entries together, while reads for
// different locales never collide - a stand fetched in es-ES must never
// answer a ca-ES or en-GB request from the same cache slot.

func idKey(id fmt.Stringer, locale enums.Locale) string {
	return "id:" + locale.String() + ":" + id.String()
}

func nameKey(name string, locale enums.Locale) string {
	return "name:" + locale.String() + ":" + hashString(name)
}

func allKey(locale enums.Locale) string {
	return "all:" + locale.String()
}

// standFilterKey renders filters in a fixed field order (rarity,
// attackPower, speed, attackRange, endurance, precision, potential,
// evolvesFrom, search) so two requests differing only in query-param order
// share one cache entry, then hashes the result to bound key length.
func standFilterKey(filters ports.StandFilters, locale enums.Locale) string {
	canonical := stringifyStat(filters.Rarity) + "|" +
		stringifyStat(filters.AttackPower) + "|" +
		stringifyStat(filters.Speed) + "|" +
		stringifyStat(filters.AttackRange) + "|" +
		stringifyStat(filters.Endurance) + "|" +
		stringifyStat(filters.Precision) + "|" +
		stringifyStat(filters.Potential) + "|" +
		derefString(filters.EvolvesFrom) + "|" +
		derefString(filters.Search)
	return "filter:" + locale.String() + ":" + hashString(canonical)
}

// devilFruitFilterKey mirrors standFilterKey for ports.DevilFruitFilters
// (rarity, fruitType, search).
func devilFruitFilterKey(filters ports.DevilFruitFilters, locale enums.Locale) string {
	canonical := stringifyStat(filters.Rarity) + "|" + stringifyStat(filters.FruitType) + "|" +
		derefString(filters.Search)
	return "filter:" + locale.String() + ":" + hashString(canonical)
}

// stringifyStat renders an optional fmt.Stringer enum field as its String()
// value, or "" when the pointer is nil (the field was unset).
func stringifyStat[T fmt.Stringer](v *T) string {
	if v == nil {
		return ""
	}
	return (*v).String()
}

func derefString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func hashString(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}
