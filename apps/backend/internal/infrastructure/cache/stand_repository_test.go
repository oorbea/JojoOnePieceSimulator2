package cache_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	infracache "github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/cache"
)

// countingStandRepository wraps an in-memory map and counts calls to each
// method, so tests can assert a cache hit never reached the underlying
// repository.
type countingStandRepository struct {
	mu              sync.Mutex
	stands          map[powers.PowerID]*powers.Stand
	findByIDCalls   int
	findByNameCalls int
	getAllCalls     int
	filterCalls     int
	updatePicCalls  int
	notFoundErr     error
}

func newCountingStandRepository() *countingStandRepository {
	return &countingStandRepository{stands: make(map[powers.PowerID]*powers.Stand), notFoundErr: ports.ErrStandNotFound}
}

func (r *countingStandRepository) Save(_ context.Context, stand *powers.Stand) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stands[stand.ID()] = stand
	return nil
}

func (r *countingStandRepository) FindByID(_ context.Context, id powers.PowerID) (*powers.Stand, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.findByIDCalls++
	s, ok := r.stands[id]
	if !ok {
		return nil, r.notFoundErr
	}
	return s, nil
}

func (r *countingStandRepository) FindByName(_ context.Context, name string) (*powers.Stand, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.findByNameCalls++
	for _, s := range r.stands {
		if s.Name() == name {
			return s, nil
		}
	}
	return nil, ports.ErrStandNotFound
}

func (r *countingStandRepository) GetAll(_ context.Context) ([]*powers.Stand, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.getAllCalls++
	all := make([]*powers.Stand, 0, len(r.stands))
	for _, s := range r.stands {
		all = append(all, s)
	}
	return all, nil
}

func (r *countingStandRepository) Filter(_ context.Context, filters ports.StandFilters) ([]*powers.Stand, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.filterCalls++
	var results []*powers.Stand
	for _, s := range r.stands {
		if filters.Rarity != nil && s.Rarity() != *filters.Rarity {
			continue
		}
		results = append(results, s)
	}
	return results, nil
}

func (r *countingStandRepository) Delete(_ context.Context, id powers.PowerID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.stands, id)
	return nil
}

func (r *countingStandRepository) UpdatePicture(_ context.Context, id powers.PowerID, main, thumb *string, status enums.PictureStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.updatePicCalls++
	s, ok := r.stands[id]
	if !ok {
		return ports.ErrStandNotFound
	}
	newMain, newThumb := s.Picture(), s.PictureThumb()
	if main != nil {
		newMain = *main
	}
	if thumb != nil {
		newThumb = *thumb
	}
	s.SetPictureRenditions(newMain, newThumb, status)
	return nil
}

var _ ports.IStandRepository = (*countingStandRepository)(nil)

func newTestStand(t *testing.T, name string) *powers.Stand {
	t.Helper()
	var id powers.PowerID
	id[15] = byte(len(name)) // cheap distinct id per test
	skills := []string{"skill"}
	power, err := powers.NewPower(id, name, "desc", enums.Rare, &skills, "")
	if err != nil {
		t.Fatalf("NewPower: %v", err)
	}
	stand, err := powers.NewStand(*power, enums.A, enums.B, enums.C, enums.D, enums.E, enums.Infinite, nil)
	if err != nil {
		t.Fatalf("NewStand: %v", err)
	}
	return stand
}

func TestStandRepository_FindByID_CachesOnMiss(t *testing.T) {
	next := newCountingStandRepository()
	stand := newTestStand(t, "Star Platinum")
	_ = next.Save(context.Background(), stand)

	repo := infracache.NewStandRepository(next, newFakeCache(), time.Minute, time.Second)
	ctx := context.Background()

	if _, err := repo.FindByID(ctx, stand.ID()); err != nil {
		t.Fatalf("first FindByID: %v", err)
	}
	if _, err := repo.FindByID(ctx, stand.ID()); err != nil {
		t.Fatalf("second FindByID: %v", err)
	}

	if next.findByIDCalls != 1 {
		t.Errorf("underlying FindByID calls = %d, want 1 (second call should hit cache)", next.findByIDCalls)
	}
}

func TestStandRepository_FindByID_NotFoundIsCachedAsTombstone(t *testing.T) {
	next := newCountingStandRepository()
	repo := infracache.NewStandRepository(next, newFakeCache(), time.Minute, time.Second)
	ctx := context.Background()

	var missing powers.PowerID
	missing[15] = 99

	_, err := repo.FindByID(ctx, missing)
	if !errors.Is(err, ports.ErrStandNotFound) {
		t.Fatalf("first FindByID err = %v, want ErrStandNotFound", err)
	}
	_, err = repo.FindByID(ctx, missing)
	if !errors.Is(err, ports.ErrStandNotFound) {
		t.Fatalf("second FindByID err = %v, want ErrStandNotFound", err)
	}

	if next.findByIDCalls != 1 {
		t.Errorf("underlying FindByID calls = %d, want 1 (404 should be cached)", next.findByIDCalls)
	}
}

func TestStandRepository_Save_InvalidatesCache(t *testing.T) {
	next := newCountingStandRepository()
	stand := newTestStand(t, "Silver Chariot")
	_ = next.Save(context.Background(), stand)

	repo := infracache.NewStandRepository(next, newFakeCache(), time.Minute, time.Second)
	ctx := context.Background()

	if _, err := repo.GetAll(ctx); err != nil {
		t.Fatalf("first GetAll: %v", err)
	}
	if _, err := repo.GetAll(ctx); err != nil {
		t.Fatalf("second GetAll: %v", err)
	}
	if next.getAllCalls != 1 {
		t.Fatalf("underlying GetAll calls before Save = %d, want 1", next.getAllCalls)
	}

	other := newTestStand(t, "Star Platinum")
	if err := repo.Save(ctx, other); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if _, err := repo.GetAll(ctx); err != nil {
		t.Fatalf("GetAll after Save: %v", err)
	}
	if next.getAllCalls != 2 {
		t.Errorf("underlying GetAll calls after Save = %d, want 2 (Save should invalidate)", next.getAllCalls)
	}
}

func TestStandRepository_UpdatePicture_InvalidatesCache(t *testing.T) {
	next := newCountingStandRepository()
	stand := newTestStand(t, "Crazy Diamond")
	_ = next.Save(context.Background(), stand)

	repo := infracache.NewStandRepository(next, newFakeCache(), time.Minute, time.Second)
	ctx := context.Background()

	if _, err := repo.FindByID(ctx, stand.ID()); err != nil {
		t.Fatalf("FindByID: %v", err)
	}

	main := "new-key"
	if err := repo.UpdatePicture(ctx, stand.ID(), &main, nil, enums.PictureReady); err != nil {
		t.Fatalf("UpdatePicture: %v", err)
	}

	got, err := repo.FindByID(ctx, stand.ID())
	if err != nil {
		t.Fatalf("FindByID after UpdatePicture: %v", err)
	}
	if got.Picture() != "new-key" {
		t.Errorf("Picture() after UpdatePicture = %q, want %q (cache should have been invalidated)", got.Picture(), "new-key")
	}
	if next.findByIDCalls != 2 {
		t.Errorf("underlying FindByID calls = %d, want 2 (invalidated cache should re-hit next)", next.findByIDCalls)
	}
}

// TestStandRepository_Save_ErrorDoesNotInvalidate proves a failed write never
// invalidates the cache: the fake cache's gen counter for "stands" must stay
// at 0 throughout, since Invalidate is never even attempted.
func TestStandRepository_Save_ErrorDoesNotInvalidate(t *testing.T) {
	c := newFakeCache()
	failing := &failingSaveRepository{countingStandRepository: newCountingStandRepository()}
	repo := infracache.NewStandRepository(failing, c, time.Minute, time.Second)
	ctx := context.Background()

	stand := newTestStand(t, "Gold Experience")
	if err := repo.Save(ctx, stand); err == nil {
		t.Fatal("Save over a failing repository: err = nil, want an error")
	}
	if c.gen["stands"] != 0 {
		t.Errorf("stands generation = %d, want 0 (a failed Save must not invalidate)", c.gen["stands"])
	}
}

// failingSaveRepository makes Save always fail, to prove the decorator
// never invalidates the cache for a write that never committed.
type failingSaveRepository struct {
	*countingStandRepository
}

func (f *failingSaveRepository) Save(context.Context, *powers.Stand) error {
	return errors.New("save failed")
}
