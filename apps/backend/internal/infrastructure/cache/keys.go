package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// standsNamespace holds every cached Stand read (FindByID, FindByName,
// GetAll, Filter) and is invalidated as a whole on any write - a Stand
// appears inside "all", inside an unknown set of filter:* entries, and
// nested in the EvolvesFrom chain of any descendant, so enumerating the
// affected keys is not tractable; flush-all is both correct and cheap for an
// admin-write catalogue.
const standsNamespace = "stands"

// presignNamespace holds cached presigned picture URLs, keyed by object
// storage key. Never invalidated wholesale - entries are evicted
// individually (on Delete) or simply expire.
const presignNamespace = "presign"

func idKey(id fmt.Stringer) string {
	return "id:" + id.String()
}

func nameKey(name string) string {
	return "name:" + hashString(name)
}

const allKey = "all"

// filterKey renders filters in a fixed field order (rarity, attackPower,
// speed, attackRange, endurance, precision, potential, evolvesFrom) so two
// requests differing only in query-param order share one cache entry, then
// hashes the result to bound key length.
func filterKey(filters ports.StandFilters) string {
	canonical := stringifyStat(filters.Rarity) + "|" +
		stringifyStat(filters.AttackPower) + "|" +
		stringifyStat(filters.Speed) + "|" +
		stringifyStat(filters.AttackRange) + "|" +
		stringifyStat(filters.Endurance) + "|" +
		stringifyStat(filters.Precision) + "|" +
		stringifyStat(filters.Potential) + "|" +
		derefString(filters.EvolvesFrom)
	return "filter:" + hashString(canonical)
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
