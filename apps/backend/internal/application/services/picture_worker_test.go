package services

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
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
	mu           sync.Mutex
	stands       map[powers.PowerID]*powers.Stand
	translations map[powers.PowerID]ports.PowerTranslations
}

func newFakeStandRepository() *fakeStandRepository {
	return &fakeStandRepository{
		stands:       make(map[powers.PowerID]*powers.Stand),
		translations: make(map[powers.PowerID]ports.PowerTranslations),
	}
}

func (f *fakeStandRepository) Save(_ context.Context, stand *powers.Stand, translations ports.PowerTranslations) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.stands[stand.ID()] = stand
	f.translations[stand.ID()] = translations
	return nil
}

func (f *fakeStandRepository) FindByID(_ context.Context, id powers.PowerID, _ enums.Locale) (*powers.Stand, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	stand, ok := f.stands[id]
	if !ok {
		return nil, ports.ErrStandNotFound
	}
	cp := *stand
	return &cp, nil
}

func (f *fakeStandRepository) FindByName(_ context.Context, name string, _ enums.Locale) (*powers.Stand, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, stand := range f.stands {
		if stand.Name() == name {
			return stand, nil
		}
	}
	return nil, ports.ErrStandNotFound
}

func (f *fakeStandRepository) GetAll(_ context.Context, _ enums.Locale) ([]*powers.Stand, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	all := make([]*powers.Stand, 0, len(f.stands))
	for _, stand := range f.stands {
		all = append(all, stand)
	}
	return all, nil
}

func (f *fakeStandRepository) Filter(_ context.Context, _ ports.StandFilters, locale enums.Locale) ([]*powers.Stand, error) {
	return f.GetAll(context.Background(), locale)
}

func (f *fakeStandRepository) Translations(_ context.Context, id powers.PowerID) (ports.PowerTranslations, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	t, ok := f.translations[id]
	if !ok {
		return nil, ports.ErrStandNotFound
	}
	return t, nil
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

func (f *fakePictureStorage) Upload(_ context.Context, key string, pic ports.Picture) (ports.StoredPicture, error) {
	if f.uploadErr != nil {
		return ports.StoredPicture{}, f.uploadErr
	}
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(pic.Content); err != nil {
		return ports.StoredPicture{}, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.objects[key] = buf.Bytes()
	return ports.StoredPicture{Provider: "r2", Key: key}, nil
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
	translations := ports.PowerTranslations{enums.EnGB: {Description: "description", Skills: []string{"skill"}}}
	if err := repo.Save(context.Background(), stand, translations); err != nil {
		t.Fatalf("saving stand: %v", err)
	}
	return stand
}

func newTestWorker(processor *fakeImageProcessor, pictures *fakePictureStorage, repo *fakeStandRepository, idGen *fakeStandIDGenerator) *PictureWorker {
	return NewPictureWorker(processor, pictures, newTestTargets(repo), idGen, WorkerConfig{
		Workers:        1,
		QueueSize:      1,
		JobTimeout:     time.Second,
		MaxDimension:   1024,
		ThumbDimension: 256,
		Quality:        80,
	}, nil)
}

// newTestTargets builds the single-target (Stand only) registry most tests
// need; tests exercising the DevilFruit path build their own.
func newTestTargets(repo *fakeStandRepository) map[enums.PictureSubjectKind]PictureTarget {
	return map[enums.PictureSubjectKind]PictureTarget{
		enums.StandSubject: {Publisher: NewStandPicturePublisher(repo), KeyPrefix: "stands"},
	}
}

func TestProcess_Success_PublishesKeysAndDeletesOld(t *testing.T) {
	repo := newFakeStandRepository()
	idGen := &fakeStandIDGenerator{}
	pictures := newFakePictureStorage()
	processor := newFakeImageProcessor()
	worker := newTestWorker(processor, pictures, repo, idGen)

	stand := newWorkerTestStand(t, repo, idGen, "stands/x/old.webp", "stands/x/old_thumb.webp", enums.PicturePending)

	worker.process(ports.PictureJob{SubjectID: stand.ID().String(), Kind: enums.StandSubject, Content: []byte("data"), ContentType: "image/png"})

	updated, err := repo.FindByID(context.Background(), stand.ID(), enums.EnGB)
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

func TestProcess_Success_PublishesReadyEvent(t *testing.T) {
	repo := newFakeStandRepository()
	idGen := &fakeStandIDGenerator{}
	pictures := newFakePictureStorage()
	processor := newFakeImageProcessor()
	hub := NewPictureEventHub()
	events, unsubscribe := hub.Subscribe()
	defer unsubscribe()

	worker := NewPictureWorker(processor, pictures, newTestTargets(repo), idGen, WorkerConfig{
		Workers:        1,
		QueueSize:      1,
		JobTimeout:     time.Second,
		MaxDimension:   1024,
		ThumbDimension: 256,
		Quality:        80,
	}, hub)

	stand := newWorkerTestStand(t, repo, idGen, "stands/x/old.webp", "stands/x/old_thumb.webp", enums.PicturePending)

	worker.process(ports.PictureJob{SubjectID: stand.ID().String(), Kind: enums.StandSubject, Content: []byte("data"), ContentType: "image/png"})

	select {
	case evt := <-events:
		if evt.Kind != enums.StandSubject || evt.SubjectID != stand.ID().String() || evt.Status != enums.PictureReady {
			t.Fatalf("event = %+v, want {StandSubject, %s, PictureReady}", evt, stand.ID().String())
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for READY event")
	}
}

func TestProcess_TranscodeFailure_MarksFailedKeepsOldPicture(t *testing.T) {
	repo := newFakeStandRepository()
	idGen := &fakeStandIDGenerator{}
	pictures := newFakePictureStorage()
	processor := newFakeImageProcessor()
	processor.transcodeErr = ports.ErrInvalidImage
	hub := NewPictureEventHub()
	events, unsubscribe := hub.Subscribe()
	defer unsubscribe()

	worker := NewPictureWorker(processor, pictures, newTestTargets(repo), idGen, WorkerConfig{
		Workers:        1,
		QueueSize:      1,
		JobTimeout:     time.Second,
		MaxDimension:   1024,
		ThumbDimension: 256,
		Quality:        80,
	}, hub)

	stand := newWorkerTestStand(t, repo, idGen, "stands/x/old.webp", "stands/x/old_thumb.webp", enums.PicturePending)

	worker.process(ports.PictureJob{SubjectID: stand.ID().String(), Kind: enums.StandSubject, Content: []byte("data"), ContentType: "image/png"})

	updated, err := repo.FindByID(context.Background(), stand.ID(), enums.EnGB)
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

	select {
	case evt := <-events:
		if evt.Kind != enums.StandSubject || evt.SubjectID != stand.ID().String() || evt.Status != enums.PictureFailed {
			t.Fatalf("event = %+v, want {StandSubject, %s, PictureFailed}", evt, stand.ID().String())
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for FAILED event")
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

	worker.process(ports.PictureJob{SubjectID: stand.ID().String(), Kind: enums.StandSubject, Content: []byte("data"), ContentType: "image/png"})

	updated, err := repo.FindByID(context.Background(), stand.ID(), enums.EnGB)
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

func (c *countingPictureStorage) Upload(ctx context.Context, key string, pic ports.Picture) (ports.StoredPicture, error) {
	c.calls++
	if c.calls > c.failAfter {
		return ports.StoredPicture{}, errors.New("upload failed")
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
	if err := worker.Enqueue(ports.PictureJob{SubjectID: stand.ID().String(), Kind: enums.StandSubject, Content: []byte("data"), ContentType: "image/png"}); err != nil {
		t.Fatalf("Enqueue: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := worker.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}

	updated, err := repo.FindByID(context.Background(), stand.ID(), enums.EnGB)
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
	worker := NewPictureWorker(processor, pictures, newTestTargets(repo), idGen, WorkerConfig{
		Workers: 0, QueueSize: 1, JobTimeout: time.Second, MaxDimension: 1024, ThumbDimension: 256, Quality: 80,
	}, nil)
	// No Start(): nothing drains the queue, so the second Enqueue must see it full.
	if err := worker.Enqueue(ports.PictureJob{SubjectID: powers.PowerID{1}.String(), Kind: enums.StandSubject}); err != nil {
		t.Fatalf("first Enqueue: %v", err)
	}
	if err := worker.Enqueue(ports.PictureJob{SubjectID: powers.PowerID{2}.String(), Kind: enums.StandSubject}); !errors.Is(err, ErrPictureQueueFull) {
		t.Fatalf("err = %v, want ErrPictureQueueFull", err)
	}
}

// fakeDevilFruitRepository is a minimal in-memory ports.IDevilFruitRepository
// - a local copy following this file's own duplication convention - used to
// exercise the worker's per-Kind routing.
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

func newWorkerTestDevilFruit(t *testing.T, repo *fakeDevilFruitRepository, idGen *fakeStandIDGenerator, main, thumb string, status enums.PictureStatus) *powers.DevilFruit {
	t.Helper()
	power, err := powers.NewPower(idGen.NewID(), "Worker Fruit", "description", enums.Rare, &[]string{"skill"}, "")
	if err != nil {
		t.Fatalf("building power: %v", err)
	}
	power.SetPictureRenditions(main, thumb, status)
	fruit, err := powers.NewDevilFruit(*power, enums.Zoan)
	if err != nil {
		t.Fatalf("building devil fruit: %v", err)
	}
	translations := ports.PowerTranslations{enums.EnGB: {Description: "description", Skills: []string{"skill"}}}
	if err := repo.Save(context.Background(), fruit, translations); err != nil {
		t.Fatalf("saving devil fruit: %v", err)
	}
	return fruit
}

func TestProcess_DevilFruitKind_PublishesUnderDevilFruitsPrefix(t *testing.T) {
	standRepo := newFakeStandRepository()
	fruitRepo := newFakeDevilFruitRepository()
	idGen := &fakeStandIDGenerator{}
	pictures := newFakePictureStorage()
	processor := newFakeImageProcessor()

	targets := map[enums.PictureSubjectKind]PictureTarget{
		enums.StandSubject:      {Publisher: NewStandPicturePublisher(standRepo), KeyPrefix: "stands"},
		enums.DevilFruitSubject: {Publisher: NewDevilFruitPicturePublisher(fruitRepo), KeyPrefix: "devil-fruits"},
	}
	worker := NewPictureWorker(processor, pictures, targets, idGen, WorkerConfig{
		Workers: 1, QueueSize: 1, JobTimeout: time.Second, MaxDimension: 1024, ThumbDimension: 256, Quality: 80,
	}, nil)

	fruit := newWorkerTestDevilFruit(t, fruitRepo, idGen, "devil-fruits/x/old.webp", "devil-fruits/x/old_thumb.webp", enums.PicturePending)

	worker.process(ports.PictureJob{SubjectID: fruit.ID().String(), Kind: enums.DevilFruitSubject, Content: []byte("data"), ContentType: "image/png"})

	updated, err := fruitRepo.FindByID(context.Background(), fruit.ID(), enums.EnGB)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if updated.PictureStatus() != enums.PictureReady {
		t.Fatalf("status = %v, want READY", updated.PictureStatus())
	}
	if !strings.HasPrefix(updated.Picture(), "devil-fruits/") {
		t.Errorf("picture key = %q, want devil-fruits/ prefix", updated.Picture())
	}
	if len(standRepo.stands) != 0 {
		t.Errorf("the stand repository should never have been touched, got %d entries", len(standRepo.stands))
	}
}

func TestProcess_UnknownKind_TouchesNothing(t *testing.T) {
	repo := newFakeStandRepository()
	idGen := &fakeStandIDGenerator{}
	pictures := newFakePictureStorage()
	processor := newFakeImageProcessor()
	worker := newTestWorker(processor, pictures, repo, idGen)

	stand := newWorkerTestStand(t, repo, idGen, "", "", enums.PictureNone)

	worker.process(ports.PictureJob{SubjectID: stand.ID().String(), Kind: enums.PictureSubjectKind(99), Content: []byte("data"), ContentType: "image/png"})

	updated, err := repo.FindByID(context.Background(), stand.ID(), enums.EnGB)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if updated.PictureStatus() != enums.PictureNone {
		t.Errorf("status = %v, want unchanged NONE", updated.PictureStatus())
	}
	if len(pictures.objects) != 0 {
		t.Errorf("nothing should have been uploaded, got %d objects", len(pictures.objects))
	}
}

// fakeStageRepository is a local copy of the one in stage_service_test.go
// (package services_test, not visible from here) - see this file's header
// comment for why fakes are duplicated per test file instead of shared.
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

// fakeStageIDGenerator returns deterministic, incrementing game.StageIDs -
// fakeStandIDGenerator can't be reused since it targets powers.PowerID.
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

func newWorkerTestStage(t *testing.T, repo *fakeStageRepository, idGen *fakeStageIDGenerator, main, thumb string, status enums.PictureStatus) game.Stage {
	t.Helper()
	id := idGen.NewID()
	st, err := game.NewStage(id, enums.Jojo, 0, "Worker Stage", "description", main)
	if err != nil {
		t.Fatalf("building stage: %v", err)
	}
	st.SetPictureRenditions(main, thumb, status)
	translations := ports.StageTranslations{enums.EnGB: "description", enums.EsES: "descripcion", enums.CaES: "descripcio"}
	if err := repo.Save(context.Background(), st, translations); err != nil {
		t.Fatalf("saving stage: %v", err)
	}
	return st
}

func TestProcess_StageKind_PublishesUnderStagesPrefix(t *testing.T) {
	standRepo := newFakeStandRepository()
	stageRepo := newFakeStageRepository()
	// The worker's idGen only mints the random uuid used to name the
	// uploaded object key (see picture_worker.go) - it's always a
	// powers.PowerID regardless of subject kind. Stage ids are minted
	// separately by stageIDGen, below.
	idGen := &fakeStandIDGenerator{}
	stageIDGen := &fakeStageIDGenerator{}
	pictures := newFakePictureStorage()
	processor := newFakeImageProcessor()

	targets := map[enums.PictureSubjectKind]PictureTarget{
		enums.StandSubject: {Publisher: NewStandPicturePublisher(standRepo), KeyPrefix: "stands"},
		enums.StageSubject: {Publisher: NewStagePicturePublisher(stageRepo), KeyPrefix: "stages"},
	}
	worker := NewPictureWorker(processor, pictures, targets, idGen, WorkerConfig{
		Workers: 1, QueueSize: 1, JobTimeout: time.Second, MaxDimension: 1024, ThumbDimension: 256, Quality: 80,
	}, nil)

	stage := newWorkerTestStage(t, stageRepo, stageIDGen, "stages/x/old.webp", "stages/x/old_thumb.webp", enums.PicturePending)

	worker.process(ports.PictureJob{SubjectID: stage.ID().String(), Kind: enums.StageSubject, Content: []byte("data"), ContentType: "image/png"})

	updated, err := stageRepo.FindByID(context.Background(), stage.ID(), enums.EnGB)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if updated.PictureStatus() != enums.PictureReady {
		t.Fatalf("status = %v, want READY", updated.PictureStatus())
	}
	if !strings.HasPrefix(updated.Picture(), "stages/") {
		t.Errorf("picture key = %q, want stages/ prefix", updated.Picture())
	}
	if len(standRepo.stands) != 0 {
		t.Errorf("the stand repository should never have been touched, got %d entries", len(standRepo.stands))
	}
}
