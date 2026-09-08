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
// Bumped to v2 when PictureCard/PictureLqip/PictureMediaID were added to
// the snapshot: Redis is appendonly on a persistent volume, so entries
// written by the pre-v2 binary would otherwise survive the deploy and
// deserialize with the new fields zeroed - silently serving stale/empty
// media ids until standTTL expired. The version bump forces a cold cache
// instead, which is cheap for a write-by-admin-only catalogue.
const standsNamespace = "stands:v2"

// devilFruitsNamespace holds every cached DevilFruit read (FindByID,
// FindByName, GetAll, Filter) and is invalidated as a whole on any write -
// same reasoning as standsNamespace.
// Bumped to v2 - see standsNamespace's doc.
const devilFruitsNamespace = "devil_fruits:v2"

// stagesNamespace holds every cached Stage read (Stages, List, Filter,
// FindByID) and is invalidated as a whole on any write - same reasoning as
// standsNamespace.
// Bumped to v2 - see standsNamespace's doc.
const stagesNamespace = "stages:v2"

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

// optionsKey backs StandRepository.Options - locale-free (powers.name is not
// translatable), so unlike every other key here it carries no locale
// component.
func optionsKey() string {
	return "options"
}

// standFilterKey hashes ports.StandFilters.Canonical() (the single source of
// truth for a filter set's canonical string rendering - see that method's
// doc) so two requests differing only in query-param order share one cache
// entry, with the result bounded to a fixed length regardless of filter
// count.
func standFilterKey(filters ports.StandFilters, locale enums.Locale) string {
	return "filter:" + locale.String() + ":" + hashString(filters.Canonical())
}

// devilFruitFilterKey mirrors standFilterKey for ports.DevilFruitFilters.
func devilFruitFilterKey(filters ports.DevilFruitFilters, locale enums.Locale) string {
	return "filter:" + locale.String() + ":" + hashString(filters.Canonical())
}

// stageFilterKey mirrors standFilterKey/devilFruitFilterKey for
// ports.StageFilters.
func stageFilterKey(filters ports.StageFilters, locale enums.Locale) string {
	return "filter:" + locale.String() + ":" + hashString(filters.Canonical())
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

func hashString(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}
