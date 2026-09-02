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

// stagesNamespace holds every cached Stage read (Stages, List, Filter,
// FindByID) and is invalidated as a whole on any write - same reasoning as
// standsNamespace.
const stagesNamespace = "stages"

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

// stageFilterKey mirrors standFilterKey/devilFruitFilterKey for
// ports.StageFilters (manga, search). ANY field added to StageFilters must
// be added here too, or two different filters silently share one slot.
func stageFilterKey(filters ports.StageFilters, locale enums.Locale) string {
	canonical := stringifyStat(filters.Manga) + "|" + derefString(filters.Search)
	return "filter:" + locale.String() + ":" + hashString(canonical)
}

// stageCatalogKey keys IStageCatalog.Stages, which takes no locale: the
// adapter resolves Description at a fixed enums.EnGB, so the key is
// prefixed with that literal locale rather than skipping the locale
// dimension - every key in this file carries one. Deliberately a different
// key shape from stageFilterKey even when a Filter carries the same Manga
// and nothing else: the admin filter surface and the gameplay catalogue are
// separate contracts (see ports.IStageCatalog's doc on the fixed EnGB
// resolution) and may diverge without one silently answering the other.
func stageCatalogKey(manga enums.Manga) string {
	return "catalog:" + enums.EnGB.String() + ":" + manga.String()
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
