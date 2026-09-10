package services_test

import (
	"bytes"
	"context"
	"errors"
	"sort"
	"sync"
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/characters"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// fakeCharacterIDGenerator returns deterministic, incrementing ids - shared
// with one_piece_character_service_test.go (same package).
type fakeCharacterIDGenerator struct {
	mu   sync.Mutex
	next byte
}

func (g *fakeCharacterIDGenerator) NewID() characters.CharacterID {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.next++
	var id characters.CharacterID
	id[15] = g.next
	return id
}

// fakeJojoCharacterRepository is a minimal in-memory
// ports.IJojoCharacterRepository, following this package's convention of
// duplicating small fakes per test file rather than sharing them.
type fakeJojoCharacterRepository struct {
	mu           sync.Mutex
	byID         map[characters.CharacterID]*characters.JojoCharacter
	translations map[characters.CharacterID]ports.CharacterTranslations
}

func newFakeJojoCharacterRepository() *fakeJojoCharacterRepository {
	return &fakeJojoCharacterRepository{
		byID:         make(map[characters.CharacterID]*characters.JojoCharacter),
		translations: make(map[characters.CharacterID]ports.CharacterTranslations),
	}
}

func (f *fakeJojoCharacterRepository) Save(_ context.Context, c *characters.JojoCharacter, translations ports.CharacterTranslations) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for id, existing := range f.byID {
		if existing.Name() == c.Name() && id != c.ID() {
			return ports.ErrJojoCharacterAlreadyExists
		}
	}
	f.byID[c.ID()] = c
	f.translations[c.ID()] = translations
	return nil
}

func (f *fakeJojoCharacterRepository) FindByID(_ context.Context, id characters.CharacterID, _ enums.Locale) (*characters.JojoCharacter, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	c, ok := f.byID[id]
	if !ok {
		return nil, ports.ErrJojoCharacterNotFound
	}
	cp := *c
	return &cp, nil
}

func (f *fakeJojoCharacterRepository) FindByName(_ context.Context, name string, _ enums.Locale) (*characters.JojoCharacter, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, c := range f.byID {
		if c.Name() == name {
			return c, nil
		}
	}
	return nil, ports.ErrJojoCharacterNotFound
}

func (f *fakeJojoCharacterRepository) GetAll(_ context.Context, _ enums.Locale) ([]*characters.JojoCharacter, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	all := make([]*characters.JojoCharacter, 0, len(f.byID))
	for _, c := range f.byID {
		all = append(all, c)
	}
	return all, nil
}

func (f *fakeJojoCharacterRepository) Filter(_ context.Context, filters ports.JojoCharacterFilters, locale enums.Locale) ([]*characters.JojoCharacter, error) {
	all, _ := f.GetAll(context.Background(), locale)
	if filters.Hamon == nil {
		return all, nil
	}
	var out []*characters.JojoCharacter
	for _, c := range all {
		if c.Hamon() == *filters.Hamon {
			out = append(out, c)
		}
	}
	return out, nil
}

func (f *fakeJojoCharacterRepository) Page(ctx context.Context, filters ports.JojoCharacterFilters, locale enums.Locale, afterName *string, limit int) ([]*characters.JojoCharacter, bool, error) {
	all, err := f.Filter(ctx, filters, locale)
	if err != nil {
		return nil, false, err
	}
	sort.Slice(all, func(i, j int) bool { return all[i].Name() < all[j].Name() })
	start := 0
	if afterName != nil {
		for i, c := range all {
			if c.Name() > *afterName {
				start = i
				break
			}
			start = i + 1
		}
	}
	page := all[start:]
	if len(page) > limit {
		return page[:limit], true, nil
	}
	return page, false, nil
}

func (f *fakeJojoCharacterRepository) Count(ctx context.Context, filters ports.JojoCharacterFilters, locale enums.Locale) (int, error) {
	all, err := f.Filter(ctx, filters, locale)
	return len(all), err
}

func (f *fakeJojoCharacterRepository) Translations(_ context.Context, id characters.CharacterID) (ports.CharacterTranslations, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	t, ok := f.translations[id]
	if !ok {
		return nil, ports.ErrJojoCharacterNotFound
	}
	return t, nil
}

func (f *fakeJojoCharacterRepository) Delete(_ context.Context, id characters.CharacterID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.byID[id]; !ok {
		return ports.ErrJojoCharacterNotFound
	}
	delete(f.byID, id)
	return nil
}

func (f *fakeJojoCharacterRepository) UpdatePicture(_ context.Context, id characters.CharacterID, main, thumb, card, lqip *string, status enums.PictureStatus) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	c, ok := f.byID[id]
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

func (f *fakeJojoCharacterRepository) SetMediaID(_ context.Context, id characters.CharacterID, mediaID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	c, ok := f.byID[id]
	if !ok {
		return ports.ErrJojoCharacterNotFound
	}
	c.SetMediaID(mediaID)
	return nil
}

var _ ports.IJojoCharacterRepository = (*fakeJojoCharacterRepository)(nil)

func newTestJojoCharacterService(repo *fakeJojoCharacterRepository, idGen *fakeCharacterIDGenerator, pictures *fakePictureStorage,
	processor *fakeImageProcessor, enqueuer *fakePictureEnqueuer, policy services.PicturePolicy) *services.JojoCharacterService {
	return services.NewJojoCharacterService(repo, idGen, pictures, processor, enqueuer, policy)
}

func newTestJojoCharacterViaService(t *testing.T, svc *services.JojoCharacterService, name string) *characters.JojoCharacter {
	t.Helper()
	c, err := svc.CreateJojoCharacter(context.Background(), services.JojoCharacterInput{
		Name:         name,
		Translations: ports.CharacterTranslations{enums.EnGB: name + " description"},
		Rarity:       enums.Rare,
		Hamon:        enums.HamonAdvanced,
		Spin:         enums.SpinGolden,
		BattleIQ:     130,
	})
	if err != nil {
		t.Fatalf("CreateJojoCharacter: %v", err)
	}
	return c
}

func TestCreateJojoCharacter_RejectsInvalidHamon(t *testing.T) {
	repo := newFakeJojoCharacterRepository()
	svc := newTestJojoCharacterService(repo, &fakeCharacterIDGenerator{}, newFakePictureStorage(), newFakeImageProcessor(), &fakePictureEnqueuer{},
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/png"}})

	_, err := svc.CreateJojoCharacter(context.Background(), services.JojoCharacterInput{
		Name: "Bad Character", Translations: ports.CharacterTranslations{enums.EnGB: "description"}, Rarity: enums.Rare,
		Hamon: enums.HamonLevel(99), Spin: enums.SpinGolden, BattleIQ: 100,
	})
	if !errors.Is(err, enums.ErrInvalidHamonLevel) {
		t.Fatalf("err = %v, want ErrInvalidHamonLevel", err)
	}
}

func TestCreateJojoCharacter_ListGetFilterDelete(t *testing.T) {
	repo := newFakeJojoCharacterRepository()
	svc := newTestJojoCharacterService(repo, &fakeCharacterIDGenerator{}, newFakePictureStorage(), newFakeImageProcessor(), &fakePictureEnqueuer{},
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/png"}})

	created := newTestJojoCharacterViaService(t, svc, "Jotaro Kujo")
	if created.BattleIQ() != 130 {
		t.Errorf("battleIQ = %d, want 130", created.BattleIQ())
	}

	got, err := svc.GetJojoCharacter(context.Background(), created.ID(), enums.EnGB)
	if err != nil {
		t.Fatalf("GetJojoCharacter: %v", err)
	}
	if got.Name() != "Jotaro Kujo" {
		t.Errorf("name = %q, want Jotaro Kujo", got.Name())
	}

	all, err := svc.ListJojoCharacters(context.Background(), enums.EnGB)
	if err != nil {
		t.Fatalf("ListJojoCharacters: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("len(all) = %d, want 1", len(all))
	}

	hamon := enums.HamonAdvanced
	filtered, err := svc.FilterJojoCharacters(context.Background(), ports.JojoCharacterFilters{Hamon: &hamon}, enums.EnGB)
	if err != nil {
		t.Fatalf("FilterJojoCharacters: %v", err)
	}
	if len(filtered) != 1 {
		t.Fatalf("len(filtered) = %d, want 1", len(filtered))
	}

	if err := svc.DeleteJojoCharacter(context.Background(), created.ID()); err != nil {
		t.Fatalf("DeleteJojoCharacter: %v", err)
	}
	if _, err := svc.GetJojoCharacter(context.Background(), created.ID(), enums.EnGB); !errors.Is(err, ports.ErrJojoCharacterNotFound) {
		t.Fatalf("err after delete = %v, want ErrJojoCharacterNotFound", err)
	}
}

func TestUpdateJojoCharacter_PreservesExistingPicture(t *testing.T) {
	repo := newFakeJojoCharacterRepository()
	svc := newTestJojoCharacterService(repo, &fakeCharacterIDGenerator{}, newFakePictureStorage(), newFakeImageProcessor(), &fakePictureEnqueuer{},
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/png"}})

	c := newTestJojoCharacterViaService(t, svc, "Joseph Joestar")
	if err := repo.UpdatePicture(context.Background(), c.ID(), strPtr("jojo-characters/x/main.webp"), strPtr("jojo-characters/x/main_thumb.webp"), nil, nil, enums.PictureReady); err != nil {
		t.Fatalf("UpdatePicture: %v", err)
	}

	updated, err := svc.UpdateJojoCharacter(context.Background(), c.ID(), services.JojoCharacterInput{
		Name: "Joseph Joestar", Translations: ports.CharacterTranslations{enums.EnGB: "updated description"}, Rarity: enums.Epic,
		Hamon: enums.HamonPerfect, Spin: enums.SpinInfinite, BattleIQ: 200,
	})
	if err != nil {
		t.Fatalf("UpdateJojoCharacter: %v", err)
	}
	if updated.Picture() != "jojo-characters/x/main.webp" {
		t.Errorf("picture after update = %q, want preserved", updated.Picture())
	}
	if updated.PictureStatus() != enums.PictureReady {
		t.Errorf("picture status after update = %v, want preserved READY", updated.PictureStatus())
	}
	if updated.Hamon() != enums.HamonPerfect || updated.BattleIQ() != 200 {
		t.Errorf("updated stats = (%v, %d), want (PERFECT, 200)", updated.Hamon(), updated.BattleIQ())
	}
}

func TestSetJojoCharacterPicture_NotFound_DoesNotTouchStorage(t *testing.T) {
	repo := newFakeJojoCharacterRepository()
	pictures := newFakePictureStorage()
	enqueuer := &fakePictureEnqueuer{}
	svc := newTestJojoCharacterService(repo, &fakeCharacterIDGenerator{}, pictures, newFakeImageProcessor(), enqueuer,
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/png"}})

	_, err := svc.SetJojoCharacterPicture(context.Background(), characters.CharacterID{1}, ports.Picture{
		Content: bytes.NewReader([]byte("data")), ContentType: "image/png", Size: 4,
	})
	if !errors.Is(err, ports.ErrJojoCharacterNotFound) {
		t.Fatalf("err = %v, want ports.ErrJojoCharacterNotFound", err)
	}
	if len(pictures.objects) != 0 {
		t.Errorf("storage should be untouched, got %d objects", len(pictures.objects))
	}
	if len(enqueuer.jobs) != 0 {
		t.Errorf("no job should be enqueued, got %d", len(enqueuer.jobs))
	}
}

func TestSetJojoCharacterPicture_Success_MarksPendingAndEnqueues(t *testing.T) {
	repo := newFakeJojoCharacterRepository()
	idGen := &fakeCharacterIDGenerator{}
	pictures := newFakePictureStorage()
	enqueuer := &fakePictureEnqueuer{}
	svc := newTestJojoCharacterService(repo, idGen, pictures, newFakeImageProcessor(), enqueuer,
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/png"}})

	c := newTestJojoCharacterViaService(t, svc, "Giorno Giovanna")

	updated, err := svc.SetJojoCharacterPicture(context.Background(), c.ID(), ports.Picture{
		Content: bytes.NewReader([]byte("first")), ContentType: "image/png", Size: 5,
	})
	if err != nil {
		t.Fatalf("SetJojoCharacterPicture: %v", err)
	}
	if updated.PictureStatus() != enums.PicturePending {
		t.Fatalf("picture status = %v, want PENDING", updated.PictureStatus())
	}

	if len(enqueuer.jobs) != 1 {
		t.Fatalf("jobs enqueued = %d, want 1", len(enqueuer.jobs))
	}
	job := enqueuer.jobs[0]
	if job.SubjectID != c.ID().String() || job.Kind != enums.JojoCharacterSubject || job.ContentType != "image/png" {
		t.Errorf("unexpected job: %+v", job)
	}
}

func TestSetJojoCharacterPicture_RejectsUnsupportedType(t *testing.T) {
	repo := newFakeJojoCharacterRepository()
	svc := newTestJojoCharacterService(repo, &fakeCharacterIDGenerator{}, newFakePictureStorage(), newFakeImageProcessor(), &fakePictureEnqueuer{},
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/png"}})

	c := newTestJojoCharacterViaService(t, svc, "Bruno Bucciarati")

	_, err := svc.SetJojoCharacterPicture(context.Background(), c.ID(), ports.Picture{
		Content: bytes.NewReader([]byte("data")), ContentType: "text/plain", Size: 4,
	})
	if !errors.Is(err, services.ErrUnsupportedPictureType) {
		t.Fatalf("err = %v, want ErrUnsupportedPictureType", err)
	}
}
