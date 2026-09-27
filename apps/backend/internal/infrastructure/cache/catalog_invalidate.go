package cache

import (
	"context"
	"log"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// InvalidateCatalogNamespaces flushes every catalogue namespace (stands,
// devil fruits, stages, JoJo characters, One Piece characters) as a whole.
// A goose seed migration (see db/migrations/00017_seed_catalog_*.sql) writes
// these tables directly with raw SQL, bypassing every repository's
// invalidate() call - without this, a cache warmed before the migration ran
// would keep serving pre-seed data until its TTL expired. Called once from
// cmd/app/main.go right after postgres.Migrate reports it applied something,
// fail-open like every other cache path here: a flush failure is logged,
// never fatal, since the affected keys are bounded by TTL regardless.
func InvalidateCatalogNamespaces(ctx context.Context, c ports.ICache) {
	for _, ns := range []string{
		standsNamespace,
		devilFruitsNamespace,
		stagesNamespace,
		jojoCharactersNamespace,
		onePieceCharactersNamespace,
	} {
		if err := c.Invalidate(ctx, ns); err != nil {
			log.Printf("cache: invalidating %s namespace after migration: %v", ns, err)
		}
	}
}
