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

// fakeOnePieceCharacterRepository mirrors fakeJojoCharacterRepository - see
// that type's doc.
type fakeOnePieceCharacterRepository struct {
	mu           sync.Mutex
	byID         map[characters.CharacterID]*characters.OnePieceCharacter
	translations map[characters.CharacterID]ports.CharacterTranslations
}

func newFakeOnePieceCharacterRepository() *fakeOnePieceCharacterRepository {
	return &fakeOnePieceCharacterRepository{
		byID:         make(map[characters.CharacterID]*characters.OnePieceCharacter),
		translations: make(map[characters.CharacterID]ports.CharacterTranslations),
	}
}

func (f *fakeOnePieceCharacterRepository) Save(_ context.Context, c *characters.OnePieceCharacter, translations ports.CharacterTranslations) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for id, existing := range f.byID {
		if existing.Name() == c.Name() && id != c.ID() {
			return ports.ErrOnePieceCharacterAlreadyExists
		}
	}
	f.byID[c.ID()] = c
	f.translations[c.ID()] = translations
	return nil
}

func (f *fakeOnePieceCharacterRepository) FindByID(_ context.Context, id characters.CharacterID, _ enums.Locale) (*characters.OnePieceCharacter, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	c, ok := f.byID[id]
	if !ok {
		return nil, ports.ErrOnePieceCharacterNotFound
	}
	cp := *c
	return &cp, nil
}

func (f *fakeOnePieceCharacterRepository) FindByName(_ context.Context, name string, _ enums.Locale) (*characters.OnePieceCharacter, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, c := range f.byID {
		if c.Name() == name {
			return c, nil
		}
	}
	return nil, ports.ErrOnePieceCharacterNotFound
}

func (f *fakeOnePieceCharacterRepository) GetAll(_ context.Context, _ enums.Locale) ([]*characters.OnePieceCharacter, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	all := make([]*characters.OnePieceCharacter, 0, len(f.byID))
	for _, c := range f.byID {
		all = append(all, c)
	}
	return all, nil
}

func (f *fakeOnePieceCharacterRepository) Filter(_ context.Context, filters ports.OnePieceCharacterFilters, locale enums.Locale) ([]*characters.OnePieceCharacter, error) {
	all, _ := f.GetAll(context.Background(), locale)
	if filters.FruitMastery == nil {
		return all, nil
	}
	var out []*characters.OnePieceCharacter
	for _, c := range all {
		if c.FruitMastery() == *filters.FruitMastery {
			out = append(out, c)
		}
	}
	return out, nil
}

func (f *fakeOnePieceCharacterRepository) Page(ctx context.Context, filters ports.OnePieceCharacterFilters, locale enums.Locale, afterName *string, limit int) ([]*characters.OnePieceCharacter, bool, error) {
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

func (f *fakeOnePieceCharacterRepository) Count(ctx context.Context, filters ports.OnePieceCharacterFilters, locale enums.Locale) (int, error) {
	all, err := f.Filter(ctx, filters, locale)
	return len(all), err
}

func (f *fakeOnePieceCharacterRepository) Translations(_ context.Context, id characters.CharacterID) (ports.CharacterTranslations, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	t, ok := f.translations[id]
	if !ok {
		return nil, ports.ErrOnePieceCharacterNotFound
	}
	return t, nil
}

func (f *fakeOnePieceCharacterRepository) Delete(_ context.Context, id characters.CharacterID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.byID[id]; !ok {
		return ports.ErrOnePieceCharacterNotFound
	}
	delete(f.byID, id)
	return nil
}

func (f *fakeOnePieceCharacterRepository) UpdatePicture(_ context.Context, id characters.CharacterID, main, thumb, card, lqip *string, status enums.PictureStatus) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	c, ok := f.byID[id]
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

func (f *fakeOnePieceCharacterRepository) SetMediaID(_ context.Context, id characters.CharacterID, mediaID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	c, ok := f.byID[id]
	if !ok {
		return ports.ErrOnePieceCharacterNotFound
	}
	c.SetMediaID(mediaID)
	return nil
}

var _ ports.IOnePieceCharacterRepository = (*fakeOnePieceCharacterRepository)(nil)

func newTestOnePieceCharacterService(repo *fakeOnePieceCharacterRepository, idGen *fakeCharacterIDGenerator, pictures *fakePictureStorage,
	processor *fakeImageProcessor, enqueuer *fakePictureEnqueuer, policy services.PicturePolicy) *services.OnePieceCharacterService {
	return services.NewOnePieceCharacterService(repo, idGen, pictures, processor, enqueuer, policy)
}

func newTestOnePieceCharacterViaService(t *testing.T, svc *services.OnePieceCharacterService, name string) *characters.OnePieceCharacter {
	t.Helper()
	c, err := svc.CreateOnePieceCharacter(context.Background(), services.OnePieceCharacterInput{
		Name:            name,
		Translations:    ports.CharacterTranslations{enums.EnGB: name + " description"},
		Rarity:          enums.Epic,
		PhysicalForm:    enums.PhysicalFormMarineCaptain,
		ArmamentHaki:    enums.HakiViceAdmiral,
		ObservationHaki: enums.HakiPrivate,
		ConquerorHaki:   enums.HakiNone,
		FruitMastery:    enums.FruitMasteryAdvanced,
	})
	if err != nil {
		t.Fatalf("CreateOnePieceCharacter: %v", err)
	}
	return c
}

func TestCreateOnePieceCharacter_RejectsInvalidFruitMastery(t *testing.T) {
	repo := newFakeOnePieceCharacterRepository()
	svc := newTestOnePieceCharacterService(repo, &fakeCharacterIDGenerator{}, newFakePictureStorage(), newFakeImageProcessor(), &fakePictureEnqueuer{},
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/png"}})

	_, err := svc.CreateOnePieceCharacter(context.Background(), services.OnePieceCharacterInput{
		Name: "Bad Character", Translations: ports.CharacterTranslations{enums.EnGB: "description"}, Rarity: enums.Epic,
		PhysicalForm: enums.PhysicalFormPrivate, ArmamentHaki: enums.HakiNone, ObservationHaki: enums.HakiNone,
		ConquerorHaki: enums.HakiNone, FruitMastery: enums.FruitMastery(99),
	})
	if !errors.Is(err, enums.ErrInvalidFruitMastery) {
		t.Fatalf("err = %v, want ErrInvalidFruitMastery", err)
	}
}

func TestCreateOnePieceCharacter_ListGetFilterDelete(t *testing.T) {
	repo := newFakeOnePieceCharacterRepository()
	svc := newTestOnePieceCharacterService(repo, &fakeCharacterIDGenerator{}, newFakePictureStorage(), newFakeImageProcessor(), &fakePictureEnqueuer{},
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/png"}})

	created := newTestOnePieceCharacterViaService(t, svc, "Roronoa Zoro")
	if created.FruitMastery() != enums.FruitMasteryAdvanced {
		t.Errorf("fruit mastery = %v, want ADVANCED", created.FruitMastery())
	}

	got, err := svc.GetOnePieceCharacter(context.Background(), created.ID(), enums.EnGB)
	if err != nil {
		t.Fatalf("GetOnePieceCharacter: %v", err)
	}
	if got.Name() != "Roronoa Zoro" {
		t.Errorf("name = %q, want Roronoa Zoro", got.Name())
	}

	all, err := svc.ListOnePieceCharacters(context.Background(), enums.EnGB)
	if err != nil {
		t.Fatalf("ListOnePieceCharacters: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("len(all) = %d, want 1", len(all))
	}

	mastery := enums.FruitMasteryAdvanced
	filtered, err := svc.FilterOnePieceCharacters(context.Background(), ports.OnePieceCharacterFilters{FruitMastery: &mastery}, enums.EnGB)
	if err != nil {
		t.Fatalf("FilterOnePieceCharacters: %v", err)
	}
	if len(filtered) != 1 {
		t.Fatalf("len(filtered) = %d, want 1", len(filtered))
	}

	if err := svc.DeleteOnePieceCharacter(context.Background(), created.ID()); err != nil {
		t.Fatalf("DeleteOnePieceCharacter: %v", err)
	}
	if _, err := svc.GetOnePieceCharacter(context.Background(), created.ID(), enums.EnGB); !errors.Is(err, ports.ErrOnePieceCharacterNotFound) {
		t.Fatalf("err after delete = %v, want ErrOnePieceCharacterNotFound", err)
	}
}

func TestUpdateOnePieceCharacter_PreservesExistingPicture(t *testing.T) {
	repo := newFakeOnePieceCharacterRepository()
	svc := newTestOnePieceCharacterService(repo, &fakeCharacterIDGenerator{}, newFakePictureStorage(), newFakeImageProcessor(), &fakePictureEnqueuer{},
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/png"}})

	c := newTestOnePieceCharacterViaService(t, svc, "Nami")
	if err := repo.UpdatePicture(context.Background(), c.ID(), strPtr("one-piece-characters/x/main.webp"), strPtr("one-piece-characters/x/main_thumb.webp"), nil, nil, enums.PictureReady); err != nil {
		t.Fatalf("UpdatePicture: %v", err)
	}

	updated, err := svc.UpdateOnePieceCharacter(context.Background(), c.ID(), services.OnePieceCharacterInput{
		Name: "Nami", Translations: ports.CharacterTranslations{enums.EnGB: "updated description"}, Rarity: enums.Legendary,
		PhysicalForm: enums.PhysicalFormYonkoCommander, ArmamentHaki: enums.HakiYonkoCommander, ObservationHaki: enums.HakiYonkoPlus,
		ConquerorHaki: enums.HakiPrivate, FruitMastery: enums.FruitMasteryAwakened,
	})
	if err != nil {
		t.Fatalf("UpdateOnePieceCharacter: %v", err)
	}
	if updated.Picture() != "one-piece-characters/x/main.webp" {
		t.Errorf("picture after update = %q, want preserved", updated.Picture())
	}
	if updated.PictureStatus() != enums.PictureReady {
		t.Errorf("picture status after update = %v, want preserved READY", updated.PictureStatus())
	}
	if updated.FruitMastery() != enums.FruitMasteryAwakened {
		t.Errorf("fruit mastery after update = %v, want AWAKENED", updated.FruitMastery())
	}
}

func TestSetOnePieceCharacterPicture_NotFound_DoesNotTouchStorage(t *testing.T) {
	repo := newFakeOnePieceCharacterRepository()
	pictures := newFakePictureStorage()
	enqueuer := &fakePictureEnqueuer{}
	svc := newTestOnePieceCharacterService(repo, &fakeCharacterIDGenerator{}, pictures, newFakeImageProcessor(), enqueuer,
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/png"}})

	_, err := svc.SetOnePieceCharacterPicture(context.Background(), characters.CharacterID{1}, ports.Picture{
		Content: bytes.NewReader([]byte("data")), ContentType: "image/png", Size: 4,
	})
	if !errors.Is(err, ports.ErrOnePieceCharacterNotFound) {
		t.Fatalf("err = %v, want ports.ErrOnePieceCharacterNotFound", err)
	}
	if len(pictures.objects) != 0 {
		t.Errorf("storage should be untouched, got %d objects", len(pictures.objects))
	}
	if len(enqueuer.jobs) != 0 {
		t.Errorf("no job should be enqueued, got %d", len(enqueuer.jobs))
	}
}

func TestSetOnePieceCharacterPicture_Success_MarksPendingAndEnqueues(t *testing.T) {
	repo := newFakeOnePieceCharacterRepository()
	idGen := &fakeCharacterIDGenerator{}
	pictures := newFakePictureStorage()
	enqueuer := &fakePictureEnqueuer{}
	svc := newTestOnePieceCharacterService(repo, idGen, pictures, newFakeImageProcessor(), enqueuer,
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/png"}})

	c := newTestOnePieceCharacterViaService(t, svc, "Sanji")

	updated, err := svc.SetOnePieceCharacterPicture(context.Background(), c.ID(), ports.Picture{
		Content: bytes.NewReader([]byte("first")), ContentType: "image/png", Size: 5,
	})
	if err != nil {
		t.Fatalf("SetOnePieceCharacterPicture: %v", err)
	}
	if updated.PictureStatus() != enums.PicturePending {
		t.Fatalf("picture status = %v, want PENDING", updated.PictureStatus())
	}

	if len(enqueuer.jobs) != 1 {
		t.Fatalf("jobs enqueued = %d, want 1", len(enqueuer.jobs))
	}
	job := enqueuer.jobs[0]
	if job.SubjectID != c.ID().String() || job.Kind != enums.OnePieceCharacterSubject || job.ContentType != "image/png" {
		t.Errorf("unexpected job: %+v", job)
	}
}
