package cache

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// DevilFruitRepository decorates a ports.IDevilFruitRepository with a
// ports.ICache, mirroring stand_repository.go's read-through/write-invalidate
// shape.
type DevilFruitRepository struct {
	next        ports.IDevilFruitRepository
	cache       ports.ICache
	fruitTTL    time.Duration
	notFoundTTL time.Duration
}

var _ ports.IDevilFruitRepository = (*DevilFruitRepository)(nil)

// NewDevilFruitRepository wraps next with a read-through/write-invalidate
// cache. fruitTTL bounds staleness if an invalidation is ever missed;
// notFoundTTL bounds how long a 404 is cached before a retry reaches next
// again.
func NewDevilFruitRepository(next ports.IDevilFruitRepository, c ports.ICache, fruitTTL, notFoundTTL time.Duration) *DevilFruitRepository {
	return &DevilFruitRepository{next: next, cache: c, fruitTTL: fruitTTL, notFoundTTL: notFoundTTL}
}

// Save delegates, then invalidates the whole devil fruits namespace (every
// locale's entries) on success.
func (r *DevilFruitRepository) Save(ctx context.Context, fruit *powers.DevilFruit, translations ports.PowerTranslations) error {
	if err := r.next.Save(ctx, fruit, translations); err != nil {
		return err
	}
	r.invalidate(ctx)
	return nil
}

// FindByID is read-through: a hit avoids next entirely, including a cached
// ports.ErrDevilFruitNotFound tombstone. Cached per locale - see keys.go.
func (r *DevilFruitRepository) FindByID(ctx context.Context, id powers.PowerID, locale enums.Locale) (*powers.DevilFruit, error) {
	key := idKey(id, locale)
	if data, ok := r.cache.Get(ctx, devilFruitsNamespace, key); ok {
		if isTombstone(data) {
			return nil, ports.ErrDevilFruitNotFound
		}
		if fruit, err := unmarshalDevilFruit(data); err == nil {
			return fruit, nil
		}
	}

	fruit, err := r.next.FindByID(ctx, id, locale)
	if err != nil {
		if errors.Is(err, ports.ErrDevilFruitNotFound) {
			r.cache.Set(ctx, devilFruitsNamespace, key, notFoundTombstone, r.notFoundTTL)
		}
		return nil, err
	}

	if data, err := marshalDevilFruit(fruit); err == nil {
		r.cache.Set(ctx, devilFruitsNamespace, key, data, r.fruitTTL)
	}
	return fruit, nil
}

// FindByName is read-through, same shape as FindByID.
func (r *DevilFruitRepository) FindByName(ctx context.Context, name string, locale enums.Locale) (*powers.DevilFruit, error) {
	key := nameKey(name, locale)
	if data, ok := r.cache.Get(ctx, devilFruitsNamespace, key); ok {
		if isTombstone(data) {
			return nil, ports.ErrDevilFruitNotFound
		}
		if fruit, err := unmarshalDevilFruit(data); err == nil {
			return fruit, nil
		}
	}

	fruit, err := r.next.FindByName(ctx, name, locale)
	if err != nil {
		if errors.Is(err, ports.ErrDevilFruitNotFound) {
			r.cache.Set(ctx, devilFruitsNamespace, key, notFoundTombstone, r.notFoundTTL)
		}
		return nil, err
	}

	if data, err := marshalDevilFruit(fruit); err == nil {
		r.cache.Set(ctx, devilFruitsNamespace, key, data, r.fruitTTL)
	}
	return fruit, nil
}

// GetAll is read-through, cached as a single entry per locale.
func (r *DevilFruitRepository) GetAll(ctx context.Context, locale enums.Locale) ([]*powers.DevilFruit, error) {
	key := allKey(locale)
	if data, ok := r.cache.Get(ctx, devilFruitsNamespace, key); ok {
		if fruits, err := unmarshalDevilFruits(data); err == nil {
			return fruits, nil
		}
	}

	fruits, err := r.next.GetAll(ctx, locale)
	if err != nil {
		return nil, err
	}

	if data, err := marshalDevilFruits(fruits); err == nil {
		r.cache.Set(ctx, devilFruitsNamespace, key, data, r.fruitTTL)
	}
	return fruits, nil
}

// Filter is read-through, keyed by a canonical rendering of filters (plus
// locale) so query param order doesn't fragment the cache.
func (r *DevilFruitRepository) Filter(ctx context.Context, filters ports.DevilFruitFilters, locale enums.Locale) ([]*powers.DevilFruit, error) {
	key := devilFruitFilterKey(filters, locale)
	if data, ok := r.cache.Get(ctx, devilFruitsNamespace, key); ok {
		if fruits, err := unmarshalDevilFruits(data); err == nil {
			return fruits, nil
		}
	}

	fruits, err := r.next.Filter(ctx, filters, locale)
	if err != nil {
		return nil, err
	}

	if data, err := marshalDevilFruits(fruits); err == nil {
		r.cache.Set(ctx, devilFruitsNamespace, key, data, r.fruitTTL)
	}
	return fruits, nil
}

// Translations bypasses the cache: admin edit forms need a fresh read of
// every locale's content, and this path is not part of the hot,
// high-traffic read surface the cache exists for.
func (r *DevilFruitRepository) Translations(ctx context.Context, id powers.PowerID) (ports.PowerTranslations, error) {
	return r.next.Translations(ctx, id)
}

// Delete delegates, then invalidates the whole devil fruits namespace on
// success.
func (r *DevilFruitRepository) Delete(ctx context.Context, id powers.PowerID) error {
	if err := r.next.Delete(ctx, id); err != nil {
		return err
	}
	r.invalidate(ctx)
	return nil
}

// UpdatePicture delegates, then invalidates the whole devil fruits namespace
// on success. Called both by the PATCH .../picture handler (moving a
// DevilFruit to PENDING) and by the background picture worker (publishing
// READY/FAILED).
func (r *DevilFruitRepository) UpdatePicture(ctx context.Context, id powers.PowerID, main, thumb *string, status enums.PictureStatus) error {
	if err := r.next.UpdatePicture(ctx, id, main, thumb, status); err != nil {
		return err
	}
	r.invalidate(ctx)
	return nil
}

// invalidate drops the whole devil fruits namespace, logging (never failing
// the already-committed write) if the cache backend can't be reached.
func (r *DevilFruitRepository) invalidate(ctx context.Context) {
	if err := r.cache.Invalidate(ctx, devilFruitsNamespace); err != nil {
		log.Printf("cache: invalidating devil fruits namespace: %v", err)
	}
}
