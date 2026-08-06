package services_test

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// fakeDevilFruitRepository is a minimal in-memory ports.IDevilFruitRepository,
// following this package's convention (see fakeStandRepository above) of
// duplicating small fakes per test file rather than sharing them.
type fakeDevilFruitRepository struct {
	mu           sync.Mutex
	fruits       map[powers.PowerID]*powers.DevilFruit
	translations map[powers.PowerID]ports.PowerTranslations
}

func newFakeDevilFruitRepository() *fakeDevilFruitRepository {
	return &fakeDevilFruitRepository{
		fruits:       make(map[powers.PowerID]*powers.DevilFruit),
		translations: make(map[powers.PowerID]ports.PowerTranslations),
	}
}

func (f *fakeDevilFruitRepository) Save(_ context.Context, fruit *powers.DevilFruit, translations ports.PowerTranslations) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for id, existing := range f.fruits {
		if existing.Name() == fruit.Name() && id != fruit.ID() {
			return ports.ErrDevilFruitAlreadyExists
		}
	}
	f.fruits[fruit.ID()] = fruit
	f.translations[fruit.ID()] = translations
	return nil
}

func (f *fakeDevilFruitRepository) FindByID(_ context.Context, id powers.PowerID, _ enums.Locale) (*powers.DevilFruit, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	fruit, ok := f.fruits[id]
	if !ok {
		return nil, ports.ErrDevilFruitNotFound
	}
	cp := *fruit
	return &cp, nil
}

func (f *fakeDevilFruitRepository) FindByName(_ context.Context, name string, _ enums.Locale) (*powers.DevilFruit, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, fruit := range f.fruits {
		if fruit.Name() == name {
			return fruit, nil
		}
	}
	return nil, ports.ErrDevilFruitNotFound
}

func (f *fakeDevilFruitRepository) GetAll(_ context.Context, _ enums.Locale) ([]*powers.DevilFruit, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	all := make([]*powers.DevilFruit, 0, len(f.fruits))
	for _, fruit := range f.fruits {
		all = append(all, fruit)
	}
	return all, nil
}

func (f *fakeDevilFruitRepository) Filter(_ context.Context, _ ports.DevilFruitFilters, locale enums.Locale) ([]*powers.DevilFruit, error) {
	return f.GetAll(context.Background(), locale)
}

func (f *fakeDevilFruitRepository) Translations(_ context.Context, id powers.PowerID) (ports.PowerTranslations, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	t, ok := f.translations[id]
	if !ok {
		return nil, ports.ErrDevilFruitNotFound
	}
	return t, nil
}

func (f *fakeDevilFruitRepository) Delete(_ context.Context, id powers.PowerID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.fruits[id]; !ok {
		return ports.ErrDevilFruitNotFound
	}
	delete(f.fruits, id)
	return nil
}

func (f *fakeDevilFruitRepository) UpdatePicture(_ context.Context, id powers.PowerID, main, thumb *string, status enums.PictureStatus) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	fruit, ok := f.fruits[id]
	if !ok {
		return ports.ErrDevilFruitNotFound
	}
	newMain, newThumb := fruit.Picture(), fruit.PictureThumb()
	if main != nil {
		newMain = *main
	}
	if thumb != nil {
		newThumb = *thumb
	}
	fruit.SetPictureRenditions(newMain, newThumb, status)
	return nil
}

var _ ports.IDevilFruitRepository = (*fakeDevilFruitRepository)(nil)

func newTestDevilFruit(t *testing.T, repo *fakeDevilFruitRepository, idGen *fakeStandIDGenerator, name string) *powers.DevilFruit {
	t.Helper()
	id := idGen.NewID()
	power, err := powers.NewPower(id, name, name+" description", enums.Rare, &[]string{"skill"}, "")
	if err != nil {
		t.Fatalf("building power: %v", err)
	}
	fruit, err := powers.NewDevilFruit(*power, enums.Zoan)
	if err != nil {
		t.Fatalf("building devil fruit: %v", err)
	}
	translations := ports.PowerTranslations{enums.EnGB: {Description: name + " description", Skills: []string{"skill"}}}
	if err := repo.Save(context.Background(), fruit, translations); err != nil {
		t.Fatalf("saving devil fruit: %v", err)
	}
	return fruit
}

func newTestDevilFruitService(repo *fakeDevilFruitRepository, idGen *fakeStandIDGenerator, pictures *fakePictureStorage,
	processor *fakeImageProcessor, enqueuer *fakePictureEnqueuer, policy services.PicturePolicy) *services.DevilFruitService {
	return services.NewDevilFruitService(repo, idGen, pictures, processor, enqueuer, policy)
}

func TestCreateDevilFruit_RejectsInvalidFruitType(t *testing.T) {
	repo := newFakeDevilFruitRepository()
	idGen := &fakeStandIDGenerator{}
	svc := newTestDevilFruitService(repo, idGen, newFakePictureStorage(), newFakeImageProcessor(), &fakePictureEnqueuer{},
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/png"}})

	_, err := svc.CreateDevilFruit(context.Background(), services.DevilFruitInput{
		Name: "Bad Fruit", Translations: ports.PowerTranslations{enums.EnGB: {Description: "description", Skills: []string{"skill"}}}, Rarity: enums.Rare,
		FruitType: enums.FruitType(99),
	})
	if !errors.Is(err, enums.ErrInvalidFruitType) {
		t.Fatalf("err = %v, want ErrInvalidFruitType", err)
	}
}

func TestCreateDevilFruit_ListGetFilterDelete(t *testing.T) {
	repo := newFakeDevilFruitRepository()
	idGen := &fakeStandIDGenerator{}
	svc := newTestDevilFruitService(repo, idGen, newFakePictureStorage(), newFakeImageProcessor(), &fakePictureEnqueuer{},
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/png"}})

	created, err := svc.CreateDevilFruit(context.Background(), services.DevilFruitInput{
		Name: "Gomu Gomu no Mi", Translations: ports.PowerTranslations{enums.EnGB: {Description: "rubber", Skills: []string{"Gear Second"}}}, Rarity: enums.Legendary,
		FruitType: enums.MythicalZoan,
	})
	if err != nil {
		t.Fatalf("CreateDevilFruit: %v", err)
	}
	if created.FruitType() != enums.MythicalZoan {
		t.Errorf("fruit type = %v, want MYTHICAL_ZOAN", created.FruitType())
	}

	got, err := svc.GetDevilFruit(context.Background(), created.ID(), enums.EnGB)
	if err != nil {
		t.Fatalf("GetDevilFruit: %v", err)
	}
	if got.Name() != "Gomu Gomu no Mi" {
		t.Errorf("name = %q, want Gomu Gomu no Mi", got.Name())
	}

	all, err := svc.ListDevilFruits(context.Background(), enums.EnGB)
	if err != nil {
		t.Fatalf("ListDevilFruits: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("len(all) = %d, want 1", len(all))
	}

	rarity := enums.Legendary
	filtered, err := svc.FilterDevilFruits(context.Background(), ports.DevilFruitFilters{Rarity: &rarity}, enums.EnGB)
	if err != nil {
		t.Fatalf("FilterDevilFruits: %v", err)
	}
	if len(filtered) != 1 {
		t.Fatalf("len(filtered) = %d, want 1", len(filtered))
	}

	if err := svc.DeleteDevilFruit(context.Background(), created.ID()); err != nil {
		t.Fatalf("DeleteDevilFruit: %v", err)
	}
	if _, err := svc.GetDevilFruit(context.Background(), created.ID(), enums.EnGB); !errors.Is(err, ports.ErrDevilFruitNotFound) {
		t.Fatalf("err after delete = %v, want ErrDevilFruitNotFound", err)
	}
}

func TestUpdateDevilFruit_PreservesExistingPicture(t *testing.T) {
	repo := newFakeDevilFruitRepository()
	idGen := &fakeStandIDGenerator{}
	svc := newTestDevilFruitService(repo, idGen, newFakePictureStorage(), newFakeImageProcessor(), &fakePictureEnqueuer{},
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/png"}})

	fruit := newTestDevilFruit(t, repo, idGen, "Mera Mera no Mi")
	if err := repo.UpdatePicture(context.Background(), fruit.ID(), strPtr("devil-fruits/x/main.webp"), strPtr("devil-fruits/x/main_thumb.webp"), enums.PictureReady); err != nil {
		t.Fatalf("UpdatePicture: %v", err)
	}

	updated, err := svc.UpdateDevilFruit(context.Background(), fruit.ID(), services.DevilFruitInput{
		Name: "Mera Mera no Mi", Translations: ports.PowerTranslations{enums.EnGB: {Description: "updated description", Skills: []string{"skill"}}}, Rarity: enums.Epic,
		FruitType: enums.Logia,
	})
	if err != nil {
		t.Fatalf("UpdateDevilFruit: %v", err)
	}
	if updated.Picture() != "devil-fruits/x/main.webp" {
		t.Errorf("picture after update = %q, want preserved", updated.Picture())
	}
	if updated.PictureThumb() != "devil-fruits/x/main_thumb.webp" {
		t.Errorf("picture thumb after update = %q, want preserved", updated.PictureThumb())
	}
	if updated.PictureStatus() != enums.PictureReady {
		t.Errorf("picture status after update = %v, want preserved READY", updated.PictureStatus())
	}
	if updated.FruitType() != enums.Logia {
		t.Errorf("fruit type after update = %v, want LOGIA", updated.FruitType())
	}
}

func TestSetDevilFruitPicture_NotFound_DoesNotTouchStorage(t *testing.T) {
	repo := newFakeDevilFruitRepository()
	pictures := newFakePictureStorage()
	enqueuer := &fakePictureEnqueuer{}
	svc := newTestDevilFruitService(repo, &fakeStandIDGenerator{}, pictures, newFakeImageProcessor(), enqueuer,
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/png"}})

	_, err := svc.SetDevilFruitPicture(context.Background(), powers.PowerID{1}, ports.Picture{
		Content: bytes.NewReader([]byte("data")), ContentType: "image/png", Size: 4,
	})
	if !errors.Is(err, ports.ErrDevilFruitNotFound) {
		t.Fatalf("err = %v, want ports.ErrDevilFruitNotFound", err)
	}
	if len(pictures.objects) != 0 {
		t.Errorf("storage should be untouched, got %d objects", len(pictures.objects))
	}
	if len(enqueuer.jobs) != 0 {
		t.Errorf("no job should be enqueued, got %d", len(enqueuer.jobs))
	}
}

func TestSetDevilFruitPicture_RejectsUnsupportedType(t *testing.T) {
	repo := newFakeDevilFruitRepository()
	idGen := &fakeStandIDGenerator{}
	svc := newTestDevilFruitService(repo, idGen, newFakePictureStorage(), newFakeImageProcessor(), &fakePictureEnqueuer{},
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/png"}})

	fruit := newTestDevilFruit(t, repo, idGen, "Bara Bara no Mi")

	_, err := svc.SetDevilFruitPicture(context.Background(), fruit.ID(), ports.Picture{
		Content: bytes.NewReader([]byte("data")), ContentType: "text/plain", Size: 4,
	})
	if !errors.Is(err, services.ErrUnsupportedPictureType) {
		t.Fatalf("err = %v, want ErrUnsupportedPictureType", err)
	}
}

func TestSetDevilFruitPicture_RejectsTooLarge(t *testing.T) {
	repo := newFakeDevilFruitRepository()
	idGen := &fakeStandIDGenerator{}
	svc := newTestDevilFruitService(repo, idGen, newFakePictureStorage(), newFakeImageProcessor(), &fakePictureEnqueuer{},
		services.PicturePolicy{MaxBytes: 1, AllowedTypes: []string{"image/png"}})

	fruit := newTestDevilFruit(t, repo, idGen, "Suna Suna no Mi")

	_, err := svc.SetDevilFruitPicture(context.Background(), fruit.ID(), ports.Picture{
		Content: bytes.NewReader([]byte("data")), ContentType: "image/png", Size: 4,
	})
	if !errors.Is(err, services.ErrPictureTooLarge) {
		t.Fatalf("err = %v, want ErrPictureTooLarge", err)
	}
}

func TestSetDevilFruitPicture_ProbeFailure_UploadsNothing(t *testing.T) {
	repo := newFakeDevilFruitRepository()
	idGen := &fakeStandIDGenerator{}
	pictures := newFakePictureStorage()
	processor := newFakeImageProcessor()
	processor.probeErr = ports.ErrInvalidImage
	enqueuer := &fakePictureEnqueuer{}
	svc := newTestDevilFruitService(repo, idGen, pictures, processor, enqueuer,
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/png"}})

	fruit := newTestDevilFruit(t, repo, idGen, "Yami Yami no Mi")

	_, err := svc.SetDevilFruitPicture(context.Background(), fruit.ID(), ports.Picture{
		Content: bytes.NewReader([]byte("data")), ContentType: "image/png", Size: 4,
	})
	if !errors.Is(err, ports.ErrInvalidImage) {
		t.Fatalf("err = %v, want ports.ErrInvalidImage", err)
	}
	if len(pictures.objects) != 0 {
		t.Errorf("storage should be untouched, got %d objects", len(pictures.objects))
	}
	if len(enqueuer.jobs) != 0 {
		t.Errorf("no job should be enqueued, got %d", len(enqueuer.jobs))
	}
	if fruit.PictureStatus() != enums.PictureNone {
		t.Errorf("picture status = %v, want unchanged NONE", fruit.PictureStatus())
	}
}

func TestSetDevilFruitPicture_Success_MarksPendingAndEnqueues(t *testing.T) {
	repo := newFakeDevilFruitRepository()
	idGen := &fakeStandIDGenerator{}
	pictures := newFakePictureStorage()
	enqueuer := &fakePictureEnqueuer{}
	svc := newTestDevilFruitService(repo, idGen, pictures, newFakeImageProcessor(), enqueuer,
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/png"}})

	fruit := newTestDevilFruit(t, repo, idGen, "Gura Gura no Mi")

	updated, err := svc.SetDevilFruitPicture(context.Background(), fruit.ID(), ports.Picture{
		Content: bytes.NewReader([]byte("first")), ContentType: "image/png", Size: 5,
	})
	if err != nil {
		t.Fatalf("SetDevilFruitPicture: %v", err)
	}
	if updated.PictureStatus() != enums.PicturePending {
		t.Fatalf("picture status = %v, want PENDING", updated.PictureStatus())
	}
	if updated.Picture() != "" {
		t.Fatalf("picture key should stay empty on a first upload, got %q", updated.Picture())
	}

	if len(enqueuer.jobs) != 1 {
		t.Fatalf("jobs enqueued = %d, want 1", len(enqueuer.jobs))
	}
	job := enqueuer.jobs[0]
	if job.SubjectID != fruit.ID().String() || job.Kind != enums.DevilFruitSubject || job.ContentType != "image/png" || string(job.Content) != "first" {
		t.Errorf("unexpected job: %+v", job)
	}

	persisted, err := repo.FindByID(context.Background(), fruit.ID(), enums.EnGB)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if persisted.PictureStatus() != enums.PicturePending {
		t.Errorf("persisted status = %v, want PENDING", persisted.PictureStatus())
	}
}

func TestSetDevilFruitPicture_QueueFull_RevertsStatus(t *testing.T) {
	repo := newFakeDevilFruitRepository()
	idGen := &fakeStandIDGenerator{}
	pictures := newFakePictureStorage()
	enqueuer := &fakePictureEnqueuer{enqueueErr: services.ErrPictureQueueFull}
	svc := newTestDevilFruitService(repo, idGen, pictures, newFakeImageProcessor(), enqueuer,
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/png"}})

	fruit := newTestDevilFruit(t, repo, idGen, "Hito Hito no Mi")

	_, err := svc.SetDevilFruitPicture(context.Background(), fruit.ID(), ports.Picture{
		Content: bytes.NewReader([]byte("data")), ContentType: "image/png", Size: 4,
	})
	if !errors.Is(err, services.ErrPictureQueueFull) {
		t.Fatalf("err = %v, want ErrPictureQueueFull", err)
	}

	persisted, err := repo.FindByID(context.Background(), fruit.ID(), enums.EnGB)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if persisted.PictureStatus() != enums.PictureNone {
		t.Errorf("status = %v, want reverted to NONE", persisted.PictureStatus())
	}
}
