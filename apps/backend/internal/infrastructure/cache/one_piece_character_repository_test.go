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

// countingOnePieceCharacterRepository mirrors countingJojoCharacterRepository -
// see that type's doc.
type countingOnePieceCharacterRepository struct {
	mu            sync.Mutex
	byID          map[characters.CharacterID]*characters.OnePieceCharacter
	findByIDCalls int
	getAllCalls   int
	notFoundErr   error
}

func newCountingOnePieceCharacterRepository() *countingOnePieceCharacterRepository {
	return &countingOnePieceCharacterRepository{byID: make(map[characters.CharacterID]*characters.OnePieceCharacter), notFoundErr: ports.ErrOnePieceCharacterNotFound}
}

func (r *countingOnePieceCharacterRepository) Save(_ context.Context, c *characters.OnePieceCharacter, _ ports.CharacterTranslations) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[c.ID()] = c
	return nil
}

func (r *countingOnePieceCharacterRepository) FindByID(_ context.Context, id characters.CharacterID, _ enums.Locale) (*characters.OnePieceCharacter, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.findByIDCalls++
	c, ok := r.byID[id]
	if !ok {
		return nil, r.notFoundErr
	}
	return c, nil
}

func (r *countingOnePieceCharacterRepository) FindByName(_ context.Context, name string, _ enums.Locale) (*characters.OnePieceCharacter, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, c := range r.byID {
		if c.Name() == name {
			return c, nil
		}
	}
	return nil, ports.ErrOnePieceCharacterNotFound
}

func (r *countingOnePieceCharacterRepository) GetAll(_ context.Context, _ enums.Locale) ([]*characters.OnePieceCharacter, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.getAllCalls++
	all := make([]*characters.OnePieceCharacter, 0, len(r.byID))
	for _, c := range r.byID {
		all = append(all, c)
	}
	return all, nil
}

func (r *countingOnePieceCharacterRepository) Filter(_ context.Context, _ ports.OnePieceCharacterFilters, _ enums.Locale) ([]*characters.OnePieceCharacter, error) {
	all := make([]*characters.OnePieceCharacter, 0, len(r.byID))
	for _, c := range r.byID {
		all = append(all, c)
	}
	return all, nil
}

func (r *countingOnePieceCharacterRepository) Page(_ context.Context, _ ports.OnePieceCharacterFilters, _ enums.Locale, _ *string, _ int) ([]*characters.OnePieceCharacter, bool, error) {
	return nil, false, nil
}

func (r *countingOnePieceCharacterRepository) Count(_ context.Context, _ ports.OnePieceCharacterFilters, _ enums.Locale) (int, error) {
	return len(r.byID), nil
}

func (r *countingOnePieceCharacterRepository) Translations(_ context.Context, id characters.CharacterID) (ports.CharacterTranslations, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.byID[id]
	if !ok {
		return nil, r.notFoundErr
	}
	return ports.CharacterTranslations{enums.EnGB: c.Description()}, nil
}

func (r *countingOnePieceCharacterRepository) Delete(_ context.Context, id characters.CharacterID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.byID, id)
	return nil
}

func (r *countingOnePieceCharacterRepository) UpdatePicture(_ context.Context, id characters.CharacterID, main, thumb, card, lqip *string, status enums.PictureStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.byID[id]
	if !ok {
		return ports.ErrOnePieceCharacterNotFound
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

func (r *countingOnePieceCharacterRepository) SetMediaID(_ context.Context, id characters.CharacterID, mediaID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.byID[id]
	if !ok {
		return ports.ErrOnePieceCharacterNotFound
	}
	c.SetMediaID(mediaID)
	return nil
}

var _ ports.IOnePieceCharacterRepository = (*countingOnePieceCharacterRepository)(nil)

func newTestCachedOnePieceCharacter(t *testing.T, name string) *characters.OnePieceCharacter {
	t.Helper()
	var id characters.CharacterID
	id[15] = byte(len(name)) // cheap distinct id per test
	base, err := characters.NewCharacter(id, enums.OnePiece, name, enums.Epic, "desc", "")
	if err != nil {
		t.Fatalf("NewCharacter: %v", err)
	}
	c, err := characters.NewOnePieceCharacter(base, enums.PhysicalFormPrivate, enums.HakiNone, enums.HakiNone, enums.HakiNone, enums.FruitMasteryNone)
	if err != nil {
		t.Fatalf("NewOnePieceCharacter: %v", err)
	}
	return c
}

func TestOnePieceCharacterRepository_FindByID_CachesOnMiss(t *testing.T) {
	next := newCountingOnePieceCharacterRepository()
	c := newTestCachedOnePieceCharacter(t, "Cached Character")
	_ = next.Save(context.Background(), c, nil)

	repo := infracache.NewOnePieceCharacterRepository(next, newFakeCache(), time.Minute, time.Second)
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

func TestOnePieceCharacterRepository_FindByID_NotFoundIsCachedAsTombstone(t *testing.T) {
	next := newCountingOnePieceCharacterRepository()
	repo := infracache.NewOnePieceCharacterRepository(next, newFakeCache(), time.Minute, time.Second)
	ctx := context.Background()

	var missing characters.CharacterID
	missing[15] = 99

	_, err := repo.FindByID(ctx, missing, enums.EnGB)
	if !errors.Is(err, ports.ErrOnePieceCharacterNotFound) {
		t.Fatalf("first FindByID err = %v, want ErrOnePieceCharacterNotFound", err)
	}
	_, err = repo.FindByID(ctx, missing, enums.EnGB)
	if !errors.Is(err, ports.ErrOnePieceCharacterNotFound) {
		t.Fatalf("second FindByID err = %v, want ErrOnePieceCharacterNotFound", err)
	}
	if next.findByIDCalls != 1 {
		t.Errorf("underlying FindByID calls = %d, want 1 (404 should be cached)", next.findByIDCalls)
	}
}

func TestOnePieceCharacterRepository_Save_InvalidatesCache(t *testing.T) {
	next := newCountingOnePieceCharacterRepository()
	c := newTestCachedOnePieceCharacter(t, "First Character")
	_ = next.Save(context.Background(), c, nil)

	repo := infracache.NewOnePieceCharacterRepository(next, newFakeCache(), time.Minute, time.Second)
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

	other := newTestCachedOnePieceCharacter(t, "Second Character")
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
