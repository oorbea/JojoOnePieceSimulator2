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

// notFoundTombstone is the sentinel value cached in place of a Stand for a
// FindByID/FindByName miss, distinguishing a cached "definitely not found"
// from an ordinary cache miss (empty byte slice, since Get already reports
// the latter via its bool).
var notFoundTombstone = []byte("\x00not-found")

// StandRepository decorates a ports.IStandRepository with a ports.ICache. It
// satisfies ports.IStandRepository itself, so it drops in transparently
// wherever the undecorated repository was used - including the background
// picture worker, which must see the same cache to invalidate it once a
// transcode completes.
type StandRepository struct {
	next        ports.IStandRepository
	cache       ports.ICache
	standTTL    time.Duration
	notFoundTTL time.Duration
}

var _ ports.IStandRepository = (*StandRepository)(nil)

// NewStandRepository wraps next with a read-through/write-invalidate cache.
// standTTL bounds staleness if an invalidation is ever missed; notFoundTTL
// bounds how long a 404 is cached before a retry reaches next again.
func NewStandRepository(next ports.IStandRepository, c ports.ICache, standTTL, notFoundTTL time.Duration) *StandRepository {
	return &StandRepository{next: next, cache: c, standTTL: standTTL, notFoundTTL: notFoundTTL}
}

// Save delegates, then invalidates the whole stands namespace (every
// locale's entries) on success.
func (r *StandRepository) Save(ctx context.Context, stand *powers.Stand, translations ports.PowerTranslations) error {
	if err := r.next.Save(ctx, stand, translations); err != nil {
		return err
	}
	r.invalidate(ctx)
	return nil
}

// FindByID is read-through: a hit avoids next entirely, including a cached
// ports.ErrStandNotFound tombstone. Cached per locale - see keys.go.
func (r *StandRepository) FindByID(ctx context.Context, id powers.PowerID, locale enums.Locale) (*powers.Stand, error) {
	key := idKey(id, locale)
	if data, ok := r.cache.Get(ctx, standsNamespace, key); ok {
		if isTombstone(data) {
			return nil, ports.ErrStandNotFound
		}
		if stand, err := unmarshalStand(data); err == nil {
			return stand, nil
		}
		// A corrupt/incompatible cache entry falls through to next rather
		// than failing the request.
	}

	stand, err := r.next.FindByID(ctx, id, locale)
	if err != nil {
		if errors.Is(err, ports.ErrStandNotFound) {
			r.cache.Set(ctx, standsNamespace, key, notFoundTombstone, r.notFoundTTL)
		}
		return nil, err
	}

	if data, err := marshalStand(stand); err == nil {
		r.cache.Set(ctx, standsNamespace, key, data, r.standTTL)
	}
	return stand, nil
}

// FindByName is read-through, same shape as FindByID.
func (r *StandRepository) FindByName(ctx context.Context, name string, locale enums.Locale) (*powers.Stand, error) {
	key := nameKey(name, locale)
	if data, ok := r.cache.Get(ctx, standsNamespace, key); ok {
		if isTombstone(data) {
			return nil, ports.ErrStandNotFound
		}
		if stand, err := unmarshalStand(data); err == nil {
			return stand, nil
		}
	}

	stand, err := r.next.FindByName(ctx, name, locale)
	if err != nil {
		if errors.Is(err, ports.ErrStandNotFound) {
			r.cache.Set(ctx, standsNamespace, key, notFoundTombstone, r.notFoundTTL)
		}
		return nil, err
	}

	if data, err := marshalStand(stand); err == nil {
		r.cache.Set(ctx, standsNamespace, key, data, r.standTTL)
	}
	return stand, nil
}

// GetAll is read-through, cached as a single entry per locale.
func (r *StandRepository) GetAll(ctx context.Context, locale enums.Locale) ([]*powers.Stand, error) {
	key := allKey(locale)
	if data, ok := r.cache.Get(ctx, standsNamespace, key); ok {
		if stands, err := unmarshalStands(data); err == nil {
			return stands, nil
		}
	}

	stands, err := r.next.GetAll(ctx, locale)
	if err != nil {
		return nil, err
	}

	if data, err := marshalStands(stands); err == nil {
		r.cache.Set(ctx, standsNamespace, key, data, r.standTTL)
	}
	return stands, nil
}

// Filter is read-through, keyed by a canonical rendering of filters (plus
// locale) so query param order doesn't fragment the cache.
func (r *StandRepository) Filter(ctx context.Context, filters ports.StandFilters, locale enums.Locale) ([]*powers.Stand, error) {
	key := standFilterKey(filters, locale)
	if data, ok := r.cache.Get(ctx, standsNamespace, key); ok {
		if stands, err := unmarshalStands(data); err == nil {
			return stands, nil
		}
	}

	stands, err := r.next.Filter(ctx, filters, locale)
	if err != nil {
		return nil, err
	}

	if data, err := marshalStands(stands); err == nil {
		r.cache.Set(ctx, standsNamespace, key, data, r.standTTL)
	}
	return stands, nil
}

// Translations bypasses the cache: admin edit forms need a fresh read of
// every locale's content, and this path is not part of the hot,
// high-traffic read surface the cache exists for.
func (r *StandRepository) Translations(ctx context.Context, id powers.PowerID) (ports.PowerTranslations, error) {
	return r.next.Translations(ctx, id)
}

// Delete delegates, then invalidates the whole stands namespace on success.
func (r *StandRepository) Delete(ctx context.Context, id powers.PowerID) error {
	if err := r.next.Delete(ctx, id); err != nil {
		return err
	}
	r.invalidate(ctx)
	return nil
}

// UpdatePicture delegates, then invalidates the whole stands namespace on
// success. Called both by the PATCH .../picture handler (moving a Stand to
// PENDING) and by the background picture worker (publishing READY/FAILED),
// so a background transcode completing is reflected for readers without
// waiting out standTTL.
func (r *StandRepository) UpdatePicture(ctx context.Context, id powers.PowerID, main, thumb *string, status enums.PictureStatus) error {
	if err := r.next.UpdatePicture(ctx, id, main, thumb, status); err != nil {
		return err
	}
	r.invalidate(ctx)
	return nil
}

// invalidate drops the whole stands namespace, logging (never failing the
// already-committed write) if the cache backend can't be reached - the
// entries left behind are bounded by standTTL/notFoundTTL either way.
func (r *StandRepository) invalidate(ctx context.Context) {
	if err := r.cache.Invalidate(ctx, standsNamespace); err != nil {
		log.Printf("cache: invalidating stands namespace: %v", err)
	}
}

func isTombstone(data []byte) bool {
	return string(data) == string(notFoundTombstone)
}
