package cache_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/characters"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	infracache "github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/cache"
)

// countingJojoCharacterRepository mirrors countingDevilFruitRepository, so
// tests can assert a cache hit never reached the underlying repository.
type countingJojoCharacterRepository struct {
	mu            sync.Mutex
	byID          map[characters.CharacterID]*characters.JojoCharacter
	findByIDCalls int
	getAllCalls   int
	filterCalls   int
	notFoundErr   error
}

func newCountingJojoCharacterRepository() *countingJojoCharacterRepository {
	return &countingJojoCharacterRepository{byID: make(map[characters.CharacterID]*characters.JojoCharacter), notFoundErr: ports.ErrJojoCharacterNotFound}
}

func (r *countingJojoCharacterRepository) Save(_ context.Context, c *characters.JojoCharacter, _ ports.CharacterTranslations) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[c.ID()] = c
	return nil
}

func (r *countingJojoCharacterRepository) FindByID(_ context.Context, id characters.CharacterID, _ enums.Locale) (*characters.JojoCharacter, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.findByIDCalls++
	c, ok := r.byID[id]
	if !ok {
		return nil, r.notFoundErr
	}
	return c, nil
}

func (r *countingJojoCharacterRepository) FindByName(_ context.Context, name string, _ enums.Locale) (*characters.JojoCharacter, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, c := range r.byID {
		if c.Name() == name {
			return c, nil
		}
	}
	return nil, ports.ErrJojoCharacterNotFound
}

func (r *countingJojoCharacterRepository) GetAll(_ context.Context, _ enums.Locale) ([]*characters.JojoCharacter, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.getAllCalls++
	all := make([]*characters.JojoCharacter, 0, len(r.byID))
	for _, c := range r.byID {
		all = append(all, c)
	}
	return all, nil
}

func (r *countingJojoCharacterRepository) Filter(_ context.Context, _ ports.JojoCharacterFilters, _ enums.Locale) ([]*characters.JojoCharacter, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.filterCalls++
	all := make([]*characters.JojoCharacter, 0, len(r.byID))
	for _, c := range r.byID {
		all = append(all, c)
	}
	return all, nil
}

func (r *countingJojoCharacterRepository) Page(_ context.Context, _ ports.JojoCharacterFilters, _ enums.Locale, _ *string, _ int) ([]*characters.JojoCharacter, bool, error) {
	return nil, false, nil
}

func (r *countingJojoCharacterRepository) Count(_ context.Context, _ ports.JojoCharacterFilters, _ enums.Locale) (int, error) {
	return len(r.byID), nil
}

func (r *countingJojoCharacterRepository) Translations(_ context.Context, id characters.CharacterID) (ports.CharacterTranslations, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.byID[id]
	if !ok {
		return nil, r.notFoundErr
	}
	return ports.CharacterTranslations{enums.EnGB: c.Description()}, nil
}

func (r *countingJojoCharacterRepository) Delete(_ context.Context, id characters.CharacterID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.byID, id)
	return nil
}

func (r *countingJojoCharacterRepository) UpdatePicture(_ context.Context, id characters.CharacterID, main, thumb, card, lqip *string, status enums.PictureStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.byID[id]
	if !ok {
		return ports.ErrJojoCharacterNotFound
	}
	newMain, newThumb, newCard, newLqip := c.Picture(), c.PictureThumb(), c.PictureCard(), c.PictureLqip()
	if main != nil {
		newMain = *main
	}
	if thumb != nil {
		newThumb = *thumb
	}
	if card != nil {
		newCard = *card
	}
	if lqip != nil {
		newLqip = *lqip
	}
	c.SetPictureRenditions(newMain, newThumb, newCard, newLqip, status)
	return nil
}

func (r *countingJojoCharacterRepository) SetMediaID(_ context.Context, id characters.CharacterID, mediaID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.byID[id]
	if !ok {
		return ports.ErrJojoCharacterNotFound
	}
	c.SetMediaID(mediaID)
	return nil
}

var _ ports.IJojoCharacterRepository = (*countingJojoCharacterRepository)(nil)

func newTestCachedJojoCharacter(t *testing.T, name string, battleIQ byte) *characters.JojoCharacter {
	t.Helper()
	var id characters.CharacterID
	id[15] = byte(len(name)) // cheap distinct id per test
	base, err := characters.NewCharacter(id, enums.Jojo, name, enums.Rare, "desc", "")
	if err != nil {
		t.Fatalf("NewCharacter: %v", err)
	}
	c, err := characters.NewJojoCharacter(base, enums.HamonBasic, enums.SpinBasic, battleIQ)
	if err != nil {
		t.Fatalf("NewJojoCharacter: %v", err)
	}
	return c
}

func TestJojoCharacterRepository_FindByID_CachesOnMiss(t *testing.T) {
	next := newCountingJojoCharacterRepository()
	c := newTestCachedJojoCharacter(t, "Cached Character", 130)
	_ = next.Save(context.Background(), c, nil)

	repo := infracache.NewJojoCharacterRepository(next, newFakeCache(), time.Minute, time.Second)
	ctx := context.Background()

	if _, err := repo.FindByID(ctx, c.ID(), enums.EnGB); err != nil {
		t.Fatalf("first FindByID: %v", err)
	}
	if _, err := repo.FindByID(ctx, c.ID(), enums.EnGB); err != nil {
		t.Fatalf("second FindByID: %v", err)
	}
	if next.findByIDCalls != 1 {
		t.Errorf("underlying FindByID calls = %d, want 1 (second call should hit cache)", next.findByIDCalls)
	}
}

func TestJojoCharacterRepository_FindByID_NotFoundIsCachedAsTombstone(t *testing.T) {
	next := newCountingJojoCharacterRepository()
	repo := infracache.NewJojoCharacterRepository(next, newFakeCache(), time.Minute, time.Second)
	ctx := context.Background()

	var missing characters.CharacterID
	missing[15] = 99

	_, err := repo.FindByID(ctx, missing, enums.EnGB)
	if !errors.Is(err, ports.ErrJojoCharacterNotFound) {
		t.Fatalf("first FindByID err = %v, want ErrJojoCharacterNotFound", err)
	}
	_, err = repo.FindByID(ctx, missing, enums.EnGB)
	if !errors.Is(err, ports.ErrJojoCharacterNotFound) {
		t.Fatalf("second FindByID err = %v, want ErrJojoCharacterNotFound", err)
	}
	if next.findByIDCalls != 1 {
		t.Errorf("underlying FindByID calls = %d, want 1 (404 should be cached)", next.findByIDCalls)
	}
}

func TestJojoCharacterRepository_Save_InvalidatesCache(t *testing.T) {
	next := newCountingJojoCharacterRepository()
	c := newTestCachedJojoCharacter(t, "First Character", 100)
	_ = next.Save(context.Background(), c, nil)

	repo := infracache.NewJojoCharacterRepository(next, newFakeCache(), time.Minute, time.Second)
	ctx := context.Background()

	if _, err := repo.GetAll(ctx, enums.EnGB); err != nil {
		t.Fatalf("first GetAll: %v", err)
	}
	if _, err := repo.GetAll(ctx, enums.EnGB); err != nil {
		t.Fatalf("second GetAll: %v", err)
	}
	if next.getAllCalls != 1 {
		t.Fatalf("underlying GetAll calls before Save = %d, want 1", next.getAllCalls)
	}

	other := newTestCachedJojoCharacter(t, "Second Character", 110)
	if err := repo.Save(ctx, other, nil); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if _, err := repo.GetAll(ctx, enums.EnGB); err != nil {
		t.Fatalf("GetAll after Save: %v", err)
	}
	if next.getAllCalls != 2 {
		t.Errorf("underlying GetAll calls after Save = %d, want 2 (Save should invalidate)", next.getAllCalls)
	}
}

// TestJojoCharacterRepository_FindByID_NeverCrossesLocales proves the cache
// key includes locale - see the DevilFruit equivalent's doc for why.
func TestJojoCharacterRepository_FindByID_NeverCrossesLocales(t *testing.T) {
	next := newCountingJojoCharacterRepository()
	c := newTestCachedJojoCharacter(t, "Multilingual Character", 150)
	_ = next.Save(context.Background(), c, nil)

	repo := infracache.NewJojoCharacterRepository(next, newFakeCache(), time.Minute, time.Second)
	ctx := context.Background()

	if _, err := repo.FindByID(ctx, c.ID(), enums.EsES); err != nil {
		t.Fatalf("FindByID es-ES: %v", err)
	}
	if _, err := repo.FindByID(ctx, c.ID(), enums.EnGB); err != nil {
		t.Fatalf("FindByID en-GB: %v", err)
	}
	if next.findByIDCalls != 2 {
		t.Errorf("underlying FindByID calls = %d, want 2 (each locale is a distinct cache entry)", next.findByIDCalls)
	}
}
