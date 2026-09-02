package cache

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// StageStore is what the decorator wraps: the concrete adapter
// repositories.StageRepository satisfies both the admin-facing CRUD port
// and the gameplay-facing catalogue port over the same table, and the
// decorator must satisfy both too so a single decorated instance can reach
// every consumer (admin CRUD, the picture worker, and GameService). Two
// separate decorators over the same data would leave one side serving
// entries the other just invalidated.
type StageStore interface {
	ports.IStageRepository
	ports.IStageCatalog
}

// StageRepository decorates a StageStore with a ports.ICache, read-through
// on every read and whole-namespace invalidate on every write - the same
// shape as StandRepository/DevilFruitRepository.
type StageRepository struct {
	next        StageStore
	cache       ports.ICache
	stageTTL    time.Duration
	notFoundTTL time.Duration
}

var (
	_ ports.IStageRepository = (*StageRepository)(nil)
	_ ports.IStageCatalog    = (*StageRepository)(nil)
	_ StageStore             = (*StageRepository)(nil)
)

// NewStageRepository wraps next with a read-through/write-invalidate cache.
// stageTTL bounds staleness if an invalidation is ever missed; notFoundTTL
// bounds how long a 404 is cached before a retry reaches next again.
func NewStageRepository(next StageStore, c ports.ICache, stageTTL, notFoundTTL time.Duration) *StageRepository {
	return &StageRepository{next: next, cache: c, stageTTL: stageTTL, notFoundTTL: notFoundTTL}
}

// Stages is read-through, cached as a single entry per manga. This is the
// gameplay path (GameService.CreateGame builds its whole round order from
// it), and it deliberately does not share a slot with an equivalent
// Filter - see stageCatalogKey.
func (r *StageRepository) Stages(ctx context.Context, manga enums.Manga) ([]game.Stage, error) {
	key := stageCatalogKey(manga)
	if data, ok := r.cache.Get(ctx, stagesNamespace, key); ok {
		if stages, err := unmarshalStages(data); err == nil {
			return stages, nil
		}
		// A corrupt/incompatible cache entry falls through to next rather
		// than failing the request.
	}

	stages, err := r.next.Stages(ctx, manga)
	if err != nil {
		return nil, err
	}

	if data, err := marshalStages(stages); err == nil {
		r.cache.Set(ctx, stagesNamespace, key, data, r.stageTTL)
	}
	return stages, nil
}

// List is read-through, cached as a single entry per locale.
func (r *StageRepository) List(ctx context.Context, locale enums.Locale) ([]game.Stage, error) {
	key := allKey(locale)
	if data, ok := r.cache.Get(ctx, stagesNamespace, key); ok {
		if stages, err := unmarshalStages(data); err == nil {
			return stages, nil
		}
	}

	stages, err := r.next.List(ctx, locale)
	if err != nil {
		return nil, err
	}

	if data, err := marshalStages(stages); err == nil {
		r.cache.Set(ctx, stagesNamespace, key, data, r.stageTTL)
	}
	return stages, nil
}

// Filter is read-through, keyed by a canonical rendering of filters (plus
// locale) so query param order doesn't fragment the cache.
func (r *StageRepository) Filter(ctx context.Context, filters ports.StageFilters, locale enums.Locale) ([]game.Stage, error) {
	key := stageFilterKey(filters, locale)
	if data, ok := r.cache.Get(ctx, stagesNamespace, key); ok {
		if stages, err := unmarshalStages(data); err == nil {
			return stages, nil
		}
	}

	stages, err := r.next.Filter(ctx, filters, locale)
	if err != nil {
		return nil, err
	}

	if data, err := marshalStages(stages); err == nil {
		r.cache.Set(ctx, stagesNamespace, key, data, r.stageTTL)
	}
	return stages, nil
}

// FindByID is read-through: a hit avoids next entirely, including a cached
// ports.ErrStageNotFound tombstone. Cached per locale - see keys.go.
// Unlike the Stand/DevilFruit decorators this returns a value, not a
// pointer, so the tombstone branch must return the zero Stage *with* the
// error, never a zero Stage on its own.
func (r *StageRepository) FindByID(ctx context.Context, id game.StageID, locale enums.Locale) (game.Stage, error) {
	key := idKey(id, locale)
	if data, ok := r.cache.Get(ctx, stagesNamespace, key); ok {
		if isTombstone(data) {
			return game.Stage{}, ports.ErrStageNotFound
		}
		if stage, err := unmarshalStage(data); err == nil {
			return stage, nil
		}
	}

	stage, err := r.next.FindByID(ctx, id, locale)
	if err != nil {
		if errors.Is(err, ports.ErrStageNotFound) {
			r.cache.Set(ctx, stagesNamespace, key, notFoundTombstone, r.notFoundTTL)
		}
		return game.Stage{}, err
	}

	if data, err := marshalStage(stage); err == nil {
		r.cache.Set(ctx, stagesNamespace, key, data, r.stageTTL)
	}
	return stage, nil
}

// Save delegates, then invalidates the whole stages namespace (every
// locale's entries, plus every manga's catalogue entry) on success.
func (r *StageRepository) Save(ctx context.Context, s game.Stage, translations ports.StageTranslations) error {
	if err := r.next.Save(ctx, s, translations); err != nil {
		return err
	}
	r.invalidate(ctx)
	return nil
}

// Delete delegates, then invalidates the whole stages namespace on success.
func (r *StageRepository) Delete(ctx context.Context, id game.StageID) error {
	if err := r.next.Delete(ctx, id); err != nil {
		return err
	}
	r.invalidate(ctx)
	return nil
}

// Translations bypasses the cache: admin edit forms need a fresh read of
// every locale's content, and this path is not part of the hot,
// high-traffic read surface the cache exists for.
func (r *StageRepository) Translations(ctx context.Context, id game.StageID) (ports.StageTranslations, error) {
	return r.next.Translations(ctx, id)
}

// UpdatePicture delegates, then invalidates the whole stages namespace on
// success. Called both by the PATCH .../picture handler (moving a Stage to
// PENDING) and by the background picture worker (publishing READY/FAILED),
// so a background transcode completing is reflected for readers without
// waiting out stageTTL.
func (r *StageRepository) UpdatePicture(ctx context.Context, id game.StageID, main, thumb *string, status enums.PictureStatus) error {
	if err := r.next.UpdatePicture(ctx, id, main, thumb, status); err != nil {
		return err
	}
	r.invalidate(ctx)
	return nil
}

// invalidate drops the whole stages namespace, logging (never failing the
// already-committed write) if the cache backend can't be reached - the
// entries left behind are bounded by stageTTL/notFoundTTL either way.
func (r *StageRepository) invalidate(ctx context.Context) {
	if err := r.cache.Invalidate(ctx, stagesNamespace); err != nil {
		log.Printf("cache: invalidating stages namespace: %v", err)
	}
}
