package services_test

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// fakeStageRepository is a minimal in-memory ports.IStageRepository, kept
// local to this package - mirrors fakeStandRepository above.
type fakeStageRepository struct {
	mu           sync.Mutex
	stages       map[game.StageID]*game.Stage
	translations map[game.StageID]ports.StageTranslations
}

func newFakeStageRepository() *fakeStageRepository {
	return &fakeStageRepository{
		stages:       make(map[game.StageID]*game.Stage),
		translations: make(map[game.StageID]ports.StageTranslations),
	}
}

func (f *fakeStageRepository) List(_ context.Context, _ enums.Locale) ([]game.Stage, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	all := make([]game.Stage, 0, len(f.stages))
	for _, s := range f.stages {
		all = append(all, *s)
	}
	return all, nil
}

func (f *fakeStageRepository) ListByManga(_ context.Context, manga enums.Manga, _ enums.Locale) ([]game.Stage, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []game.Stage
	for _, s := range f.stages {
		if s.Manga() == manga {
			out = append(out, *s)
		}
	}
	return out, nil
}

func (f *fakeStageRepository) FindByID(_ context.Context, id game.StageID, _ enums.Locale) (game.Stage, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	s, ok := f.stages[id]
	if !ok {
		return game.Stage{}, ports.ErrStageNotFound
	}
	// Return a copy, like the real (query-backed) repository does - two
	// FindByID calls must never alias the same *game.Stage.
	return *s, nil
}

func (f *fakeStageRepository) Save(_ context.Context, s game.Stage, translations ports.StageTranslations) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := s
	f.stages[s.ID()] = &cp
	f.translations[s.ID()] = translations
	return nil
}

func (f *fakeStageRepository) Delete(_ context.Context, id game.StageID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.stages[id]; !ok {
		return ports.ErrStageNotFound
	}
	delete(f.stages, id)
	return nil
}

func (f *fakeStageRepository) Translations(_ context.Context, id game.StageID) (ports.StageTranslations, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	t, ok := f.translations[id]
	if !ok {
		return nil, ports.ErrStageNotFound
	}
	return t, nil
}

func (f *fakeStageRepository) UpdatePicture(_ context.Context, id game.StageID, main, thumb *string, status enums.PictureStatus) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	s, ok := f.stages[id]
	if !ok {
		return ports.ErrStageNotFound
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

var _ ports.IStageRepository = (*fakeStageRepository)(nil)

// fakeStageIDGenerator returns deterministic, incrementing ids.
type fakeStageIDGenerator struct {
	mu   sync.Mutex
	next byte
}

func (g *fakeStageIDGenerator) NewID() game.StageID {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.next++
	var id game.StageID
	id[15] = g.next
	return id
}

func newTestStage(t *testing.T, repo *fakeStageRepository, idGen *fakeStageIDGenerator, name string) game.Stage {
	t.Helper()
	id := idGen.NewID()
	s, err := game.NewStage(id, enums.Jojo, 0, name, name+" description", "")
	if err != nil {
		t.Fatalf("building stage: %v", err)
	}
	translations := ports.StageTranslations{
		enums.EnGB: name + " description",
		enums.EsES: name + " descripcion",
		enums.CaES: name + " descripcio",
	}
	if err := repo.Save(context.Background(), s, translations); err != nil {
		t.Fatalf("saving stage: %v", err)
	}
	return s
}

func newTestStageService(repo *fakeStageRepository, idGen *fakeStageIDGenerator, pictures *fakePictureStorage,
	processor *fakeImageProcessor, enqueuer *fakePictureEnqueuer, policy services.PicturePolicy) *services.StageService {
	return services.NewStageService(repo, idGen, pictures, processor, enqueuer, policy)
}

func TestSetStagePicture_NotFound_DoesNotTouchStorage(t *testing.T) {
	repo := newFakeStageRepository()
	pictures := newFakePictureStorage()
	enqueuer := &fakePictureEnqueuer{}
	svc := newTestStageService(repo, &fakeStageIDGenerator{}, pictures, newFakeImageProcessor(), enqueuer,
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/png"}})

	_, err := svc.SetStagePicture(context.Background(), game.StageID{1}, ports.Picture{
		Content:     bytes.NewReader([]byte("data")),
		ContentType: "image/png",
		Size:        4,
	})
	if !errors.Is(err, ports.ErrStageNotFound) {
		t.Fatalf("err = %v, want ports.ErrStageNotFound", err)
	}
	if len(pictures.objects) != 0 {
		t.Errorf("storage should be untouched, got %d objects", len(pictures.objects))
	}
	if len(enqueuer.jobs) != 0 {
		t.Errorf("no job should be enqueued, got %d", len(enqueuer.jobs))
	}
}

func TestSetStagePicture_RejectsUnsupportedType(t *testing.T) {
	repo := newFakeStageRepository()
	idGen := &fakeStageIDGenerator{}
	pictures := newFakePictureStorage()
	svc := newTestStageService(repo, idGen, pictures, newFakeImageProcessor(), &fakePictureEnqueuer{},
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/png"}})

	stage := newTestStage(t, repo, idGen, "Stardust Crusaders")

	_, err := svc.SetStagePicture(context.Background(), stage.ID(), ports.Picture{
		Content: bytes.NewReader([]byte("data")), ContentType: "text/plain", Size: 4,
	})
	if !errors.Is(err, services.ErrUnsupportedPictureType) {
		t.Fatalf("err = %v, want ErrUnsupportedPictureType", err)
	}
}

func TestSetStagePicture_RejectsTooLarge(t *testing.T) {
	repo := newFakeStageRepository()
	idGen := &fakeStageIDGenerator{}
	pictures := newFakePictureStorage()
	svc := newTestStageService(repo, idGen, pictures, newFakeImageProcessor(), &fakePictureEnqueuer{},
		services.PicturePolicy{MaxBytes: 1, AllowedTypes: []string{"image/png"}})

	stage := newTestStage(t, repo, idGen, "Diamond is Unbreakable")

	_, err := svc.SetStagePicture(context.Background(), stage.ID(), ports.Picture{
		Content: bytes.NewReader([]byte("data")), ContentType: "image/png", Size: 4,
	})
	if !errors.Is(err, services.ErrPictureTooLarge) {
		t.Fatalf("err = %v, want ErrPictureTooLarge", err)
	}
}

func TestSetStagePicture_ProbeFailure_UploadsNothing(t *testing.T) {
	repo := newFakeStageRepository()
	idGen := &fakeStageIDGenerator{}
	pictures := newFakePictureStorage()
	processor := newFakeImageProcessor()
	processor.probeErr = ports.ErrInvalidImage
	enqueuer := &fakePictureEnqueuer{}
	svc := newTestStageService(repo, idGen, pictures, processor, enqueuer,
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/png"}})

	stage := newTestStage(t, repo, idGen, "Golden Wind")

	_, err := svc.SetStagePicture(context.Background(), stage.ID(), ports.Picture{
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
	if stage.PictureStatus() != enums.PictureNone {
		t.Errorf("picture status = %v, want unchanged NONE", stage.PictureStatus())
	}
}

func TestSetStagePicture_Success_MarksPendingAndEnqueues(t *testing.T) {
	repo := newFakeStageRepository()
	idGen := &fakeStageIDGenerator{}
	pictures := newFakePictureStorage()
	enqueuer := &fakePictureEnqueuer{}
	svc := newTestStageService(repo, idGen, pictures, newFakeImageProcessor(), enqueuer,
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/png"}})

	stage := newTestStage(t, repo, idGen, "Stone Ocean")

	updated, err := svc.SetStagePicture(context.Background(), stage.ID(), ports.Picture{
		Content: bytes.NewReader([]byte("first")), ContentType: "image/png", Size: 5,
	})
	if err != nil {
		t.Fatalf("SetStagePicture: %v", err)
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
	if job.SubjectID != stage.ID().String() || job.Kind != enums.StageSubject || job.ContentType != "image/png" || string(job.Content) != "first" {
		t.Errorf("unexpected job: %+v", job)
	}

	persisted, err := repo.FindByID(context.Background(), stage.ID(), enums.EnGB)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if persisted.PictureStatus() != enums.PicturePending {
		t.Errorf("persisted status = %v, want PENDING", persisted.PictureStatus())
	}
}

func TestSetStagePicture_QueueFull_RevertsStatus(t *testing.T) {
	repo := newFakeStageRepository()
	idGen := &fakeStageIDGenerator{}
	pictures := newFakePictureStorage()
	enqueuer := &fakePictureEnqueuer{enqueueErr: services.ErrPictureQueueFull}
	svc := newTestStageService(repo, idGen, pictures, newFakeImageProcessor(), enqueuer,
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/png"}})

	stage := newTestStage(t, repo, idGen, "Vento Aureo")

	_, err := svc.SetStagePicture(context.Background(), stage.ID(), ports.Picture{
		Content: bytes.NewReader([]byte("data")), ContentType: "image/png", Size: 4,
	})
	if !errors.Is(err, services.ErrPictureQueueFull) {
		t.Fatalf("err = %v, want ErrPictureQueueFull", err)
	}

	persisted, err := repo.FindByID(context.Background(), stage.ID(), enums.EnGB)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if persisted.PictureStatus() != enums.PictureNone {
		t.Errorf("status = %v, want reverted to NONE", persisted.PictureStatus())
	}
}

func TestUpdateStage_PreservesExistingPicture(t *testing.T) {
	repo := newFakeStageRepository()
	idGen := &fakeStageIDGenerator{}
	pictures := newFakePictureStorage()
	svc := newTestStageService(repo, idGen, pictures, newFakeImageProcessor(), &fakePictureEnqueuer{},
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/png"}})

	stage := newTestStage(t, repo, idGen, "Jojolion")
	if err := repo.UpdatePicture(context.Background(), stage.ID(), strPtr("stages/x/main.webp"), strPtr("stages/x/main_thumb.webp"), enums.PictureReady); err != nil {
		t.Fatalf("UpdatePicture: %v", err)
	}

	updated, err := svc.UpdateStage(context.Background(), stage.ID(), services.StageInput{
		Manga: enums.Jojo,
		Order: 1,
		Name:  "Jojolion",
		Translations: ports.StageTranslations{
			enums.EnGB: "updated description",
			enums.EsES: "descripcion actualizada",
			enums.CaES: "descripcio actualitzada",
		},
	})
	if err != nil {
		t.Fatalf("UpdateStage: %v", err)
	}
	if updated.Picture() != "stages/x/main.webp" {
		t.Errorf("picture after update = %q, want preserved", updated.Picture())
	}
	if updated.PictureThumb() != "stages/x/main_thumb.webp" {
		t.Errorf("picture thumb after update = %q, want preserved", updated.PictureThumb())
	}
	if updated.PictureStatus() != enums.PictureReady {
		t.Errorf("picture status after update = %v, want preserved READY", updated.PictureStatus())
	}
}
