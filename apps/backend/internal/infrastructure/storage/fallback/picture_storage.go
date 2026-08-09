// Package fallback implements ports.IPictureStorage by chaining several
// ports.IStorageBackend tiers, each with its own free-tier quota: an upload lands
// on the first tier with room, and falls through to the next tier both when
// a tier is at (a configurable fraction of) its quota and when a tier
// errors at runtime. Once written, an object never moves - only new uploads
// are affected by a tier filling up.
package fallback

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"sync/atomic"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// ErrStorageExhausted is returned when every tier in the chain is at (or a
// new upload would push it past) its quota threshold, or every tier's Put
// errored.
var ErrStorageExhausted = errors.New("no object-storage provider has room for this picture")

// Tier is one ports.IStorageBackend and the byte quota it gets to use before the
// chain falls through to the next tier.
type Tier struct {
	Backend    ports.IStorageBackend
	QuotaBytes int64
}

// usageCounter is an in-memory, atomically-updated byte count per provider,
// so Upload's quota check never has to round-trip to the ledger. It is
// seeded from the ledger at construction and kept in sync by Record/Forget
// and by the reconciler's periodic Replace.
type usageCounter struct {
	bytes atomic.Int64
}

// PictureStorage is the ports.IPictureStorage implementation backed by an
// ordered Tier chain plus a ports.IStorageLedger recording where each key
// actually landed.
type PictureStorage struct {
	tiers        []Tier
	byName       map[string]ports.IStorageBackend
	ledger       ports.IStorageLedger
	thresholdPct int
	usage        map[string]*usageCounter
}

var _ ports.IPictureStorage = (*PictureStorage)(nil)

// New builds a PictureStorage over tiers, in fallback order. thresholdPct
// (1-100) is the fraction of a tier's quota an upload must fit under to be
// allowed there. ctx is used only to seed the in-memory usage counters from
// the ledger's current totals.
func New(ctx context.Context, tiers []Tier, ledger ports.IStorageLedger, thresholdPct int) (*PictureStorage, error) {
	if len(tiers) == 0 {
		return nil, fmt.Errorf("fallback: at least one storage tier is required")
	}
	if thresholdPct < 1 || thresholdPct > 100 {
		return nil, fmt.Errorf("fallback: thresholdPct must be between 1 and 100, got %d", thresholdPct)
	}

	byName := make(map[string]ports.IStorageBackend, len(tiers))
	usage := make(map[string]*usageCounter, len(tiers))
	for _, t := range tiers {
		byName[t.Backend.Name()] = t.Backend
		usage[t.Backend.Name()] = &usageCounter{}
	}

	s := &PictureStorage{tiers: tiers, byName: byName, ledger: ledger, thresholdPct: thresholdPct, usage: usage}
	if err := s.refreshUsage(ctx); err != nil {
		return nil, fmt.Errorf("fallback: seeding usage from ledger: %w", err)
	}
	return s, nil
}

// refreshUsage reloads every provider's byte count from the ledger. Exported
// as RefreshUsage for the reconciler to call after it corrects drift.
func (s *PictureStorage) refreshUsage(ctx context.Context) error {
	usages, err := s.ledger.Usage(ctx)
	if err != nil {
		return err
	}
	seen := make(map[string]bool, len(usages))
	for _, u := range usages {
		seen[u.Provider] = true
		if c, ok := s.usage[u.Provider]; ok {
			c.bytes.Store(u.Bytes)
		}
	}
	for name, c := range s.usage {
		if !seen[name] {
			c.bytes.Store(0)
		}
	}
	return nil
}

// RefreshUsage re-seeds the in-memory usage counters from the ledger. The
// reconciler calls this after Replace-ing a provider's inventory so the
// fallback chain's quota decisions reflect corrected drift immediately,
// rather than waiting on the next Record/Forget.
func (s *PictureStorage) RefreshUsage(ctx context.Context) error {
	return s.refreshUsage(ctx)
}

// fitsQuota reports whether adding size bytes to provider's current usage
// would stay under thresholdPct of tier's quota. A zero quota means
// unlimited.
func (s *PictureStorage) fitsQuota(t Tier, size int64) bool {
	if t.QuotaBytes <= 0 {
		return true
	}
	current := s.usage[t.Backend.Name()].bytes.Load()
	// Multiply before dividing: real quotas are gigabyte-scale so this
	// can't overflow int64, and dividing first would truncate small quotas
	// (e.g. quota=10, pct=95 -> 10/100=0) down to a useless zero threshold.
	limit := t.QuotaBytes * int64(s.thresholdPct) / 100
	return current+size <= limit
}

// rewindableReader returns pic.Content ready to be read from its start,
// buffering it once in memory if it isn't already seekable - needed because
// a failed Put on one tier must retry the same content on the next tier, and
// an io.Reader can normally only be consumed once.
func rewindableReader(content io.Reader, size int64) (io.ReadSeeker, error) {
	if seeker, ok := content.(io.ReadSeeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return nil, fmt.Errorf("rewinding picture content: %w", err)
		}
		return seeker, nil
	}
	buf := make([]byte, size)
	if _, err := io.ReadFull(content, buf); err != nil {
		return nil, fmt.Errorf("buffering picture content: %w", err)
	}
	return bytes.NewReader(buf), nil
}

// Upload implements ports.IPictureStorage. It tries pic.PreferProvider
// first (if set and known), then every tier in order, skipping tiers over
// quota and falling through to the next tier on a Put error. The first
// tier's Put may error even with plenty of quota left (a transient network
// issue, say) and the chain still tries the rest before giving up.
func (s *PictureStorage) Upload(ctx context.Context, key string, pic ports.Picture) (ports.StoredPicture, error) {
	content, err := rewindableReader(pic.Content, pic.Size)
	if err != nil {
		return ports.StoredPicture{}, err
	}

	order := s.uploadOrder(pic.PreferProvider)

	var errs []error
	for _, t := range order {
		if pic.PreferProvider != t.Backend.Name() && !s.fitsQuota(t, pic.Size) {
			continue
		}
		if _, err := content.Seek(0, io.SeekStart); err != nil {
			return ports.StoredPicture{}, fmt.Errorf("rewinding picture content for %s: %w", t.Backend.Name(), err)
		}
		if err := t.Backend.Put(ctx, key, content, pic.ContentType, pic.Size); err != nil {
			log.Printf("uploading %q to %s failed, trying next tier: %v", key, t.Backend.Name(), err)
			errs = append(errs, err)
			continue
		}

		stored := ports.StoredPicture{Provider: t.Backend.Name(), Key: key}
		if err := s.ledger.Record(ctx, ports.StorageObject{Key: key, Provider: t.Backend.Name(), Bytes: pic.Size}); err != nil {
			// The object is safely stored; only its accounting is stale
			// until the next reconciliation. Never fail the upload for this.
			log.Printf("recording %q on %s in the storage ledger: %v", key, t.Backend.Name(), err)
		} else {
			s.usage[t.Backend.Name()].bytes.Add(pic.Size)
		}
		return stored, nil
	}

	if len(errs) > 0 {
		return ports.StoredPicture{}, fmt.Errorf("%w: %w", ErrStorageExhausted, errors.Join(errs...))
	}
	return ports.StoredPicture{}, ErrStorageExhausted
}

// uploadOrder returns the tiers in the order Upload should try them:
// preferred first (if it's a known tier), then the configured chain order,
// skipping the preferred tier the second time it would appear.
func (s *PictureStorage) uploadOrder(preferred string) []Tier {
	if preferred == "" || s.byName[preferred] == nil {
		return s.tiers
	}
	order := make([]Tier, 0, len(s.tiers))
	for _, t := range s.tiers {
		if t.Backend.Name() == preferred {
			order = append([]Tier{t}, order...)
			continue
		}
		order = append(order, t)
	}
	return order
}

// backendFor resolves which ports.IStorageBackend key lives on: whatever the
// ledger recorded, or the first tier if key predates the ledger (every
// object was on that provider before this chain existed).
func (s *PictureStorage) backendFor(ctx context.Context, key string) (ports.IStorageBackend, error) {
	obj, ok, err := s.ledger.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("looking up provider for %q: %w", key, err)
	}
	if !ok {
		return s.tiers[0].Backend, nil
	}
	backend, ok := s.byName[obj.Provider]
	if !ok {
		return nil, fmt.Errorf("key %q is recorded on unknown provider %q", key, obj.Provider)
	}
	return backend, nil
}

// PresignGetURL implements ports.IPictureStorage.
func (s *PictureStorage) PresignGetURL(ctx context.Context, key string) (string, error) {
	backend, err := s.backendFor(ctx, key)
	if err != nil {
		return "", err
	}
	return backend.PresignGet(ctx, key)
}

// Delete implements ports.IPictureStorage.
func (s *PictureStorage) Delete(ctx context.Context, key string) error {
	obj, tracked, err := s.ledger.Get(ctx, key)
	if err != nil {
		return fmt.Errorf("looking up provider for %q: %w", key, err)
	}
	backend := s.tiers[0].Backend
	if tracked {
		found, ok := s.byName[obj.Provider]
		if !ok {
			return fmt.Errorf("key %q is recorded on unknown provider %q", key, obj.Provider)
		}
		backend = found
	}

	if err := backend.Del(ctx, key); err != nil {
		return err
	}
	if err := s.ledger.Forget(ctx, key); err != nil {
		log.Printf("forgetting %q in the storage ledger: %v", key, err)
		return nil
	}
	if tracked {
		if c, ok := s.usage[backend.Name()]; ok {
			c.bytes.Add(-obj.Bytes)
		}
	}
	return nil
}
