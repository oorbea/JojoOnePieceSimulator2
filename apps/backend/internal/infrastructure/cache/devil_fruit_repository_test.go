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

// countingDevilFruitRepository mirrors countingStandRepository above, so
// tests can assert a cache hit never reached the underlying repository.
type countingDevilFruitRepository struct {
	mu              sync.Mutex
	fruits          map[powers.PowerID]*powers.DevilFruit
	findByIDCalls   int
	findByNameCalls int
	getAllCalls     int
	filterCalls     int
	updatePicCalls  int
	notFoundErr     error
}

func newCountingDevilFruitRepository() *countingDevilFruitRepository {
	return &countingDevilFruitRepository{fruits: make(map[powers.PowerID]*powers.DevilFruit), notFoundErr: ports.ErrDevilFruitNotFound}
}

func (r *countingDevilFruitRepository) Save(_ context.Context, fruit *powers.DevilFruit) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.fruits[fruit.ID()] = fruit
	return nil
}

func (r *countingDevilFruitRepository) FindByID(_ context.Context, id powers.PowerID) (*powers.DevilFruit, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.findByIDCalls++
	f, ok := r.fruits[id]
	if !ok {
		return nil, r.notFoundErr
	}
	return f, nil
}

func (r *countingDevilFruitRepository) FindByName(_ context.Context, name string) (*powers.DevilFruit, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.findByNameCalls++
	for _, f := range r.fruits {
		if f.Name() == name {
			return f, nil
		}
	}
	return nil, ports.ErrDevilFruitNotFound
}

func (r *countingDevilFruitRepository) GetAll(_ context.Context) ([]*powers.DevilFruit, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.getAllCalls++
	all := make([]*powers.DevilFruit, 0, len(r.fruits))
	for _, f := range r.fruits {
		all = append(all, f)
	}
	return all, nil
}

func (r *countingDevilFruitRepository) Filter(_ context.Context, filters ports.DevilFruitFilters) ([]*powers.DevilFruit, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.filterCalls++
	var results []*powers.DevilFruit
	for _, f := range r.fruits {
		if filters.Rarity != nil && f.Rarity() != *filters.Rarity {
			continue
		}
		results = append(results, f)
	}
	return results, nil
}

func (r *countingDevilFruitRepository) Delete(_ context.Context, id powers.PowerID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.fruits, id)
	return nil
}

func (r *countingDevilFruitRepository) UpdatePicture(_ context.Context, id powers.PowerID, main, thumb *string, status enums.PictureStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.updatePicCalls++
	f, ok := r.fruits[id]
	if !ok {
		return ports.ErrDevilFruitNotFound
	}
	newMain, newThumb := f.Picture(), f.PictureThumb()
	if main != nil {
		newMain = *main
	}
	if thumb != nil {
		newThumb = *thumb
	}
	f.SetPictureRenditions(newMain, newThumb, status)
	return nil
}

var _ ports.IDevilFruitRepository = (*countingDevilFruitRepository)(nil)

func newTestDevilFruit(t *testing.T, name string) *powers.DevilFruit {
	t.Helper()
	var id powers.PowerID
	id[15] = byte(len(name)) // cheap distinct id per test
	skills := []string{"skill"}
	power, err := powers.NewPower(id, name, "desc", enums.Rare, &skills, "")
	if err != nil {
		t.Fatalf("NewPower: %v", err)
	}
	fruit, err := powers.NewDevilFruit(*power, enums.Zoan)
	if err != nil {
		t.Fatalf("NewDevilFruit: %v", err)
	}
	return fruit
}

func TestDevilFruitRepository_FindByID_CachesOnMiss(t *testing.T) {
	next := newCountingDevilFruitRepository()
	fruit := newTestDevilFruit(t, "Zoan Fruit")
	_ = next.Save(context.Background(), fruit)

	repo := infracache.NewDevilFruitRepository(next, newFakeCache(), time.Minute, time.Second)
	ctx := context.Background()

	if _, err := repo.FindByID(ctx, fruit.ID()); err != nil {
		t.Fatalf("first FindByID: %v", err)
	}
	if _, err := repo.FindByID(ctx, fruit.ID()); err != nil {
		t.Fatalf("second FindByID: %v", err)
	}

	if next.findByIDCalls != 1 {
		t.Errorf("underlying FindByID calls = %d, want 1 (second call should hit cache)", next.findByIDCalls)
	}
}

func TestDevilFruitRepository_FindByID_NotFoundIsCachedAsTombstone(t *testing.T) {
	next := newCountingDevilFruitRepository()
	repo := infracache.NewDevilFruitRepository(next, newFakeCache(), time.Minute, time.Second)
	ctx := context.Background()

	var missing powers.PowerID
	missing[15] = 99

	_, err := repo.FindByID(ctx, missing)
	if !errors.Is(err, ports.ErrDevilFruitNotFound) {
		t.Fatalf("first FindByID err = %v, want ErrDevilFruitNotFound", err)
	}
	_, err = repo.FindByID(ctx, missing)
	if !errors.Is(err, ports.ErrDevilFruitNotFound) {
		t.Fatalf("second FindByID err = %v, want ErrDevilFruitNotFound", err)
	}

	if next.findByIDCalls != 1 {
		t.Errorf("underlying FindByID calls = %d, want 1 (404 should be cached)", next.findByIDCalls)
	}
}

func TestDevilFruitRepository_Save_InvalidatesCache(t *testing.T) {
	next := newCountingDevilFruitRepository()
	fruit := newTestDevilFruit(t, "Paramecia Fruit")
	_ = next.Save(context.Background(), fruit)

	repo := infracache.NewDevilFruitRepository(next, newFakeCache(), time.Minute, time.Second)
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

	other := newTestDevilFruit(t, "Logia Fruit")
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

func TestDevilFruitRepository_UpdatePicture_InvalidatesCache(t *testing.T) {
	next := newCountingDevilFruitRepository()
	fruit := newTestDevilFruit(t, "Ancient Zoan Fruit")
	_ = next.Save(context.Background(), fruit)

	repo := infracache.NewDevilFruitRepository(next, newFakeCache(), time.Minute, time.Second)
	ctx := context.Background()

	if _, err := repo.FindByID(ctx, fruit.ID()); err != nil {
		t.Fatalf("FindByID: %v", err)
	}

	main := "new-key"
	if err := repo.UpdatePicture(ctx, fruit.ID(), &main, nil, enums.PictureReady); err != nil {
		t.Fatalf("UpdatePicture: %v", err)
	}

	got, err := repo.FindByID(ctx, fruit.ID())
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

// TestDevilFruitRepository_Save_ErrorDoesNotInvalidate proves a failed write
// never invalidates the cache: the fake cache's gen counter for
// "devil_fruits" must stay at 0 throughout, since Invalidate is never even
// attempted.
func TestDevilFruitRepository_Save_ErrorDoesNotInvalidate(t *testing.T) {
	c := newFakeCache()
	failing := &failingSaveDevilFruitRepository{countingDevilFruitRepository: newCountingDevilFruitRepository()}
	repo := infracache.NewDevilFruitRepository(failing, c, time.Minute, time.Second)
	ctx := context.Background()

	fruit := newTestDevilFruit(t, "Special Paramecia Fruit")
	if err := repo.Save(ctx, fruit); err == nil {
		t.Fatal("Save over a failing repository: err = nil, want an error")
	}
	if c.gen["devil_fruits"] != 0 {
		t.Errorf("devil_fruits generation = %d, want 0 (a failed Save must not invalidate)", c.gen["devil_fruits"])
	}
}

// failingSaveDevilFruitRepository makes Save always fail, to prove the
// decorator never invalidates the cache for a write that never committed.
type failingSaveDevilFruitRepository struct {
	*countingDevilFruitRepository
}

func (f *failingSaveDevilFruitRepository) Save(context.Context, *powers.DevilFruit) error {
	return errors.New("save failed")
}
