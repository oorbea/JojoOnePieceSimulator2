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

// fakeStandRepository is a minimal in-memory ports.IStandRepository, kept
// local to this package (the endpoints package has its own, richer, fake
// used for HTTP-level tests).
type fakeStandRepository struct {
	mu     sync.Mutex
	stands map[powers.PowerID]*powers.Stand
}

func newFakeStandRepository() *fakeStandRepository {
	return &fakeStandRepository{stands: make(map[powers.PowerID]*powers.Stand)}
}

func (f *fakeStandRepository) Save(_ context.Context, stand *powers.Stand) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.stands[stand.ID()] = stand
	return nil
}

func (f *fakeStandRepository) FindByID(_ context.Context, id powers.PowerID) (*powers.Stand, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	stand, ok := f.stands[id]
	if !ok {
		return nil, ports.ErrStandNotFound
	}
	// Return a copy, like the real (query-backed) repository does: two
	// FindByID calls must never alias the same *powers.Stand, or a caller
	// mutating its result (e.g. SetStandPicture) could clobber a concurrent
	// mutation from another caller (e.g. the picture worker).
	cp := *stand
	return &cp, nil
}

func (f *fakeStandRepository) FindByName(_ context.Context, name string) (*powers.Stand, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, stand := range f.stands {
		if stand.Name() == name {
			return stand, nil
		}
	}
	return nil, ports.ErrStandNotFound
}

func (f *fakeStandRepository) GetAll(_ context.Context) ([]*powers.Stand, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	all := make([]*powers.Stand, 0, len(f.stands))
	for _, stand := range f.stands {
		all = append(all, stand)
	}
	return all, nil
}

func (f *fakeStandRepository) Filter(_ context.Context, _ ports.StandFilters) ([]*powers.Stand, error) {
	return f.GetAll(context.Background())
}

func (f *fakeStandRepository) Delete(_ context.Context, id powers.PowerID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.stands[id]; !ok {
		return ports.ErrStandNotFound
	}
	delete(f.stands, id)
	return nil
}

func (f *fakeStandRepository) UpdatePicture(_ context.Context, id powers.PowerID, main, thumb *string, status enums.PictureStatus) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	stand, ok := f.stands[id]
	if !ok {
		return ports.ErrStandNotFound
	}
	newMain, newThumb := stand.Picture(), stand.PictureThumb()
	if main != nil {
		newMain = *main
	}
	if thumb != nil {
		newThumb = *thumb
	}
	stand.SetPictureRenditions(newMain, newThumb, status)
	return nil
}

var _ ports.IStandRepository = (*fakeStandRepository)(nil)

// fakeStandIDGenerator returns deterministic, incrementing ids.
type fakeStandIDGenerator struct {
	mu   sync.Mutex
	next byte
}

func (g *fakeStandIDGenerator) NewID() powers.PowerID {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.next++
	var id powers.PowerID
	id[15] = g.next
	return id
}

// fakePictureStorage is an in-memory ports.IPictureStorage.
type fakePictureStorage struct {
	mu         sync.Mutex
	objects    map[string][]byte
	deleted    []string
	deleteErr  error
	uploadErr  error
	presignErr error
}

func newFakePictureStorage() *fakePictureStorage {
	return &fakePictureStorage{objects: make(map[string][]byte)}
}

func (f *fakePictureStorage) Upload(_ context.Context, key string, pic ports.Picture) error {
	if f.uploadErr != nil {
		return f.uploadErr
	}
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(pic.Content); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.objects[key] = buf.Bytes()
	return nil
}

func (f *fakePictureStorage) PresignGetURL(_ context.Context, key string) (string, error) {
	if f.presignErr != nil {
		return "", f.presignErr
	}
	return "https://r2.test/" + key, nil
}

func (f *fakePictureStorage) Delete(_ context.Context, key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.deleted = append(f.deleted, key)
	delete(f.objects, key)
	return f.deleteErr
}

var _ ports.IPictureStorage = (*fakePictureStorage)(nil)

// fakeImageProcessor is an in-memory ports.IImageProcessor: Probe/Transcode
// never inspect their input bytes, they just return the configured
// meta/result/error, so tests can exercise the pipeline without libvips.
type fakeImageProcessor struct {
	probeMeta    ports.ImageMeta
	probeErr     error
	transcodeErr error
	main         ports.EncodedImage
	thumb        ports.EncodedImage
}

func newFakeImageProcessor() *fakeImageProcessor {
	return &fakeImageProcessor{
		probeMeta: ports.ImageMeta{Width: 1, Height: 1, Pages: 1},
		main:      ports.EncodedImage{Bytes: []byte("main-webp"), ContentType: "image/webp"},
		thumb:     ports.EncodedImage{Bytes: []byte("thumb-webp"), ContentType: "image/webp"},
	}
}

func (f *fakeImageProcessor) Probe(_ []byte) (ports.ImageMeta, error) {
	return f.probeMeta, f.probeErr
}

func (f *fakeImageProcessor) Transcode(_ context.Context, _ []byte, _ ports.TranscodeOptions) (ports.EncodedImage, ports.EncodedImage, error) {
	if f.transcodeErr != nil {
		return ports.EncodedImage{}, ports.EncodedImage{}, f.transcodeErr
	}
	return f.main, f.thumb, nil
}

var _ ports.IImageProcessor = (*fakeImageProcessor)(nil)

// fakePictureEnqueuer records every job handed to it and can be made to fail
// (simulating a full queue) via enqueueErr.
type fakePictureEnqueuer struct {
	mu         sync.Mutex
	jobs       []ports.PictureJob
	enqueueErr error
}

func (f *fakePictureEnqueuer) Enqueue(job ports.PictureJob) error {
	if f.enqueueErr != nil {
		return f.enqueueErr
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.jobs = append(f.jobs, job)
	return nil
}

var _ ports.IPictureEnqueuer = (*fakePictureEnqueuer)(nil)

func newTestStand(t *testing.T, repo *fakeStandRepository, idGen *fakeStandIDGenerator, name string) *powers.Stand {
	t.Helper()
	id := idGen.NewID()
	power, err := powers.NewPower(id, name, name+" description", enums.Rare, &[]string{"skill"}, "")
	if err != nil {
		t.Fatalf("building power: %v", err)
	}
	stand, err := powers.NewStand(*power, enums.A, enums.B, enums.C, enums.D, enums.E, enums.Infinite, nil)
	if err != nil {
		t.Fatalf("building stand: %v", err)
	}
	if err := repo.Save(context.Background(), stand); err != nil {
		t.Fatalf("saving stand: %v", err)
	}
	return stand
}

func newTestStandService(repo *fakeStandRepository, idGen *fakeStandIDGenerator, pictures *fakePictureStorage,
	processor *fakeImageProcessor, enqueuer *fakePictureEnqueuer, policy services.PicturePolicy) *services.StandService {
	return services.NewStandService(repo, idGen, pictures, processor, enqueuer, policy)
}

func TestSetStandPicture_NotFound_DoesNotTouchStorage(t *testing.T) {
	repo := newFakeStandRepository()
	pictures := newFakePictureStorage()
	enqueuer := &fakePictureEnqueuer{}
	svc := newTestStandService(repo, &fakeStandIDGenerator{}, pictures, newFakeImageProcessor(), enqueuer,
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/png"}})

	_, err := svc.SetStandPicture(context.Background(), powers.PowerID{1}, ports.Picture{
		Content:     bytes.NewReader([]byte("data")),
		ContentType: "image/png",
		Size:        4,
	})
	if !errors.Is(err, ports.ErrStandNotFound) {
		t.Fatalf("err = %v, want ports.ErrStandNotFound", err)
	}
	if len(pictures.objects) != 0 {
		t.Errorf("storage should be untouched, got %d objects", len(pictures.objects))
	}
	if len(enqueuer.jobs) != 0 {
		t.Errorf("no job should be enqueued, got %d", len(enqueuer.jobs))
	}
}

func TestSetStandPicture_RejectsUnsupportedType(t *testing.T) {
	repo := newFakeStandRepository()
	idGen := &fakeStandIDGenerator{}
	pictures := newFakePictureStorage()
	svc := newTestStandService(repo, idGen, pictures, newFakeImageProcessor(), &fakePictureEnqueuer{},
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/png"}})

	stand := newTestStand(t, repo, idGen, "The World")

	_, err := svc.SetStandPicture(context.Background(), stand.ID(), ports.Picture{
		Content: bytes.NewReader([]byte("data")), ContentType: "text/plain", Size: 4,
	})
	if !errors.Is(err, services.ErrUnsupportedPictureType) {
		t.Fatalf("err = %v, want ErrUnsupportedPictureType", err)
	}
}

func TestSetStandPicture_RejectsTooLarge(t *testing.T) {
	repo := newFakeStandRepository()
	idGen := &fakeStandIDGenerator{}
	pictures := newFakePictureStorage()
	svc := newTestStandService(repo, idGen, pictures, newFakeImageProcessor(), &fakePictureEnqueuer{},
		services.PicturePolicy{MaxBytes: 1, AllowedTypes: []string{"image/png"}})

	stand := newTestStand(t, repo, idGen, "Killer Queen")

	_, err := svc.SetStandPicture(context.Background(), stand.ID(), ports.Picture{
		Content: bytes.NewReader([]byte("data")), ContentType: "image/png", Size: 4,
	})
	if !errors.Is(err, services.ErrPictureTooLarge) {
		t.Fatalf("err = %v, want ErrPictureTooLarge", err)
	}
}

func TestSetStandPicture_ProbeFailure_UploadsNothing(t *testing.T) {
	repo := newFakeStandRepository()
	idGen := &fakeStandIDGenerator{}
	pictures := newFakePictureStorage()
	processor := newFakeImageProcessor()
	processor.probeErr = ports.ErrInvalidImage
	enqueuer := &fakePictureEnqueuer{}
	svc := newTestStandService(repo, idGen, pictures, processor, enqueuer,
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/png"}})

	stand := newTestStand(t, repo, idGen, "Killer Queen Bites the Dust")

	_, err := svc.SetStandPicture(context.Background(), stand.ID(), ports.Picture{
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
	if stand.PictureStatus() != enums.PictureNone {
		t.Errorf("picture status = %v, want unchanged NONE", stand.PictureStatus())
	}
}

func TestSetStandPicture_Success_MarksPendingAndEnqueues(t *testing.T) {
	repo := newFakeStandRepository()
	idGen := &fakeStandIDGenerator{}
	pictures := newFakePictureStorage()
	enqueuer := &fakePictureEnqueuer{}
	svc := newTestStandService(repo, idGen, pictures, newFakeImageProcessor(), enqueuer,
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/png"}})

	stand := newTestStand(t, repo, idGen, "Crazy Diamond")

	updated, err := svc.SetStandPicture(context.Background(), stand.ID(), ports.Picture{
		Content: bytes.NewReader([]byte("first")), ContentType: "image/png", Size: 5,
	})
	if err != nil {
		t.Fatalf("SetStandPicture: %v", err)
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
	if job.SubjectID != stand.ID().String() || job.Kind != enums.StandSubject || job.ContentType != "image/png" || string(job.Content) != "first" {
		t.Errorf("unexpected job: %+v", job)
	}

	persisted, err := repo.FindByID(context.Background(), stand.ID())
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if persisted.PictureStatus() != enums.PicturePending {
		t.Errorf("persisted status = %v, want PENDING", persisted.PictureStatus())
	}
}

func TestSetStandPicture_QueueFull_RevertsStatus(t *testing.T) {
	repo := newFakeStandRepository()
	idGen := &fakeStandIDGenerator{}
	pictures := newFakePictureStorage()
	enqueuer := &fakePictureEnqueuer{enqueueErr: services.ErrPictureQueueFull}
	svc := newTestStandService(repo, idGen, pictures, newFakeImageProcessor(), enqueuer,
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/png"}})

	stand := newTestStand(t, repo, idGen, "Star Platinum")

	_, err := svc.SetStandPicture(context.Background(), stand.ID(), ports.Picture{
		Content: bytes.NewReader([]byte("data")), ContentType: "image/png", Size: 4,
	})
	if !errors.Is(err, services.ErrPictureQueueFull) {
		t.Fatalf("err = %v, want ErrPictureQueueFull", err)
	}

	persisted, err := repo.FindByID(context.Background(), stand.ID())
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if persisted.PictureStatus() != enums.PictureNone {
		t.Errorf("status = %v, want reverted to NONE", persisted.PictureStatus())
	}
}

func TestUpdateStand_PreservesExistingPicture(t *testing.T) {
	repo := newFakeStandRepository()
	idGen := &fakeStandIDGenerator{}
	pictures := newFakePictureStorage()
	svc := newTestStandService(repo, idGen, pictures, newFakeImageProcessor(), &fakePictureEnqueuer{},
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/png"}})

	stand := newTestStand(t, repo, idGen, "Gold Experience")
	if err := repo.UpdatePicture(context.Background(), stand.ID(), strPtr("stands/x/main.webp"), strPtr("stands/x/main_thumb.webp"), enums.PictureReady); err != nil {
		t.Fatalf("UpdatePicture: %v", err)
	}

	updated, err := svc.UpdateStand(context.Background(), stand.ID(), services.StandInput{
		Name:        "Gold Experience",
		Description: "updated description",
		Rarity:      enums.Rare,
		Skills:      &[]string{"skill"},
		AttackPower: enums.A,
		Speed:       enums.B,
		AttackRange: enums.C,
		Endurance:   enums.D,
		Precision:   enums.E,
		Potential:   enums.Infinite,
	})
	if err != nil {
		t.Fatalf("UpdateStand: %v", err)
	}
	if updated.Picture() != "stands/x/main.webp" {
		t.Errorf("picture after update = %q, want preserved", updated.Picture())
	}
	if updated.PictureThumb() != "stands/x/main_thumb.webp" {
		t.Errorf("picture thumb after update = %q, want preserved", updated.PictureThumb())
	}
	if updated.PictureStatus() != enums.PictureReady {
		t.Errorf("picture status after update = %v, want preserved READY", updated.PictureStatus())
	}
}

func strPtr(s string) *string { return &s }
