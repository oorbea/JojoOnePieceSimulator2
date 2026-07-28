package services

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// This file is an internal test (package services, not services_test) so it
// can call PictureWorker.process directly - a single synchronous call, no
// goroutines or time.Sleep needed to make the tests deterministic. Its fakes
// are local copies of the ones in stand_service_test.go (package
// services_test, a separate compiled package that isn't visible from here),
// matching this repo's existing convention of duplicating small fakes per
// test file rather than sharing them across packages.

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

type fakePictureStorage struct {
	mu        sync.Mutex
	objects   map[string][]byte
	deleted   []string
	deleteErr error
	uploadErr error
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

func newWorkerTestStand(t *testing.T, repo *fakeStandRepository, idGen *fakeStandIDGenerator, main, thumb string, status enums.PictureStatus) *powers.Stand {
	t.Helper()
	power, err := powers.NewPower(idGen.NewID(), "Worker Stand", "description", enums.Rare, &[]string{"skill"}, "")
	if err != nil {
		t.Fatalf("building power: %v", err)
	}
	power.SetPictureRenditions(main, thumb, status)
	stand, err := powers.NewStand(*power, enums.A, enums.B, enums.C, enums.D, enums.E, enums.Infinite, nil)
	if err != nil {
		t.Fatalf("building stand: %v", err)
	}
	if err := repo.Save(context.Background(), stand); err != nil {
		t.Fatalf("saving stand: %v", err)
	}
	return stand
}

func newTestWorker(processor *fakeImageProcessor, pictures *fakePictureStorage, repo *fakeStandRepository, idGen *fakeStandIDGenerator) *PictureWorker {
	return NewPictureWorker(processor, pictures, repo, idGen, WorkerConfig{
		Workers:        1,
		QueueSize:      1,
		JobTimeout:     time.Second,
		MaxDimension:   1024,
		ThumbDimension: 256,
		Quality:        80,
	})
}

func TestProcess_Success_PublishesKeysAndDeletesOld(t *testing.T) {
	repo := newFakeStandRepository()
	idGen := &fakeStandIDGenerator{}
	pictures := newFakePictureStorage()
	processor := newFakeImageProcessor()
	worker := newTestWorker(processor, pictures, repo, idGen)

	stand := newWorkerTestStand(t, repo, idGen, "stands/x/old.webp", "stands/x/old_thumb.webp", enums.PicturePending)

	worker.process(ports.PictureJob{StandID: stand.ID(), Content: []byte("data"), ContentType: "image/png"})

	updated, err := repo.FindByID(context.Background(), stand.ID())
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if updated.PictureStatus() != enums.PictureReady {
		t.Fatalf("status = %v, want READY", updated.PictureStatus())
	}
	if updated.Picture() == "stands/x/old.webp" || updated.Picture() == "" {
		t.Errorf("picture key not updated: %q", updated.Picture())
	}
	if updated.PictureThumb() == "stands/x/old_thumb.webp" || updated.PictureThumb() == "" {
		t.Errorf("picture thumb key not updated: %q", updated.PictureThumb())
	}

	if _, ok := pictures.objects[updated.Picture()]; !ok {
		t.Error("new main key missing from storage")
	}
	if _, ok := pictures.objects[updated.PictureThumb()]; !ok {
		t.Error("new thumb key missing from storage")
	}

	deletedOld := false
	deletedOldThumb := false
	for _, k := range pictures.deleted {
		if k == "stands/x/old.webp" {
			deletedOld = true
		}
		if k == "stands/x/old_thumb.webp" {
			deletedOldThumb = true
		}
	}
	if !deletedOld || !deletedOldThumb {
		t.Errorf("deleted = %v, want it to contain both old keys", pictures.deleted)
	}
}

func TestProcess_TranscodeFailure_MarksFailedKeepsOldPicture(t *testing.T) {
	repo := newFakeStandRepository()
	idGen := &fakeStandIDGenerator{}
	pictures := newFakePictureStorage()
	processor := newFakeImageProcessor()
	processor.transcodeErr = ports.ErrInvalidImage
	worker := newTestWorker(processor, pictures, repo, idGen)

	stand := newWorkerTestStand(t, repo, idGen, "stands/x/old.webp", "stands/x/old_thumb.webp", enums.PicturePending)

	worker.process(ports.PictureJob{StandID: stand.ID(), Content: []byte("data"), ContentType: "image/png"})

	updated, err := repo.FindByID(context.Background(), stand.ID())
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if updated.PictureStatus() != enums.PictureFailed {
		t.Fatalf("status = %v, want FAILED", updated.PictureStatus())
	}
	if updated.Picture() != "stands/x/old.webp" || updated.PictureThumb() != "stands/x/old_thumb.webp" {
		t.Errorf("old keys should be untouched, got picture=%q thumb=%q", updated.Picture(), updated.PictureThumb())
	}
	if len(pictures.objects) != 0 {
		t.Errorf("nothing should have been uploaded, got %d objects", len(pictures.objects))
	}
}

func TestProcess_UploadThumbFailure_DeletesPartialUploadAndMarksFailed(t *testing.T) {
	repo := newFakeStandRepository()
	idGen := &fakeStandIDGenerator{}
	pictures := newFakePictureStorage()
	processor := newFakeImageProcessor()

	// Fail only the second (thumbnail) upload, so the worker must clean up
	// the main rendition it already uploaded.
	counting := &countingPictureStorage{fakePictureStorage: pictures, failAfter: 1}
	worker := newTestWorker(processor, nil, repo, idGen)
	worker.pictures = counting

	stand := newWorkerTestStand(t, repo, idGen, "", "", enums.PicturePending)

	worker.process(ports.PictureJob{StandID: stand.ID(), Content: []byte("data"), ContentType: "image/png"})

	updated, err := repo.FindByID(context.Background(), stand.ID())
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if updated.PictureStatus() != enums.PictureFailed {
		t.Fatalf("status = %v, want FAILED", updated.PictureStatus())
	}
	if len(counting.objects) != 0 {
		t.Errorf("the main upload should have been deleted after the thumb upload failed, got %d objects left", len(counting.objects))
	}
}

// countingPictureStorage fails every Upload after the first failAfter calls
// succeed, to exercise the "thumbnail upload fails after main succeeded"
// path deterministically.
type countingPictureStorage struct {
	*fakePictureStorage
	failAfter int
	calls     int
}

func (c *countingPictureStorage) Upload(ctx context.Context, key string, pic ports.Picture) error {
	c.calls++
	if c.calls > c.failAfter {
		return errors.New("upload failed")
	}
	return c.fakePictureStorage.Upload(ctx, key, pic)
}

var _ ports.IPictureStorage = (*countingPictureStorage)(nil)

func TestShutdown_WaitsForInFlightJobs(t *testing.T) {
	repo := newFakeStandRepository()
	idGen := &fakeStandIDGenerator{}
	pictures := newFakePictureStorage()
	processor := newFakeImageProcessor()
	worker := newTestWorker(processor, pictures, repo, idGen)

	stand := newWorkerTestStand(t, repo, idGen, "", "", enums.PictureNone)
	worker.Start()
	if err := worker.Enqueue(ports.PictureJob{StandID: stand.ID(), Content: []byte("data"), ContentType: "image/png"}); err != nil {
		t.Fatalf("Enqueue: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := worker.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}

	updated, err := repo.FindByID(context.Background(), stand.ID())
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if updated.PictureStatus() != enums.PictureReady {
		t.Fatalf("status = %v, want READY after shutdown drained the queue", updated.PictureStatus())
	}
}

func TestEnqueue_FullQueueReturnsErrPictureQueueFull(t *testing.T) {
	repo := newFakeStandRepository()
	idGen := &fakeStandIDGenerator{}
	pictures := newFakePictureStorage()
	processor := newFakeImageProcessor()
	worker := NewPictureWorker(processor, pictures, repo, idGen, WorkerConfig{
		Workers: 0, QueueSize: 1, JobTimeout: time.Second, MaxDimension: 1024, ThumbDimension: 256, Quality: 80,
	})
	// No Start(): nothing drains the queue, so the second Enqueue must see it full.
	if err := worker.Enqueue(ports.PictureJob{StandID: powers.PowerID{1}}); err != nil {
		t.Fatalf("first Enqueue: %v", err)
	}
	if err := worker.Enqueue(ports.PictureJob{StandID: powers.PowerID{2}}); !errors.Is(err, ErrPictureQueueFull) {
		t.Fatalf("err = %v, want ErrPictureQueueFull", err)
	}
}
