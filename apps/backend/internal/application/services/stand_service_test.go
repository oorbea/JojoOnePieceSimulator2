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
	return stand, nil
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

func TestSetStandPicture_NotFound_DoesNotTouchStorage(t *testing.T) {
	repo := newFakeStandRepository()
	pictures := newFakePictureStorage()
	svc := services.NewStandService(repo, &fakeStandIDGenerator{}, pictures,
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
}

func TestSetStandPicture_ReplacesAndDeletesOldKey(t *testing.T) {
	repo := newFakeStandRepository()
	idGen := &fakeStandIDGenerator{}
	pictures := newFakePictureStorage()
	svc := services.NewStandService(repo, idGen, pictures,
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/png"}})

	stand := newTestStand(t, repo, idGen, "Crazy Diamond")

	updated, err := svc.SetStandPicture(context.Background(), stand.ID(), ports.Picture{
		Content:     bytes.NewReader([]byte("first")),
		ContentType: "image/png",
		Size:        5,
	})
	if err != nil {
		t.Fatalf("first SetStandPicture: %v", err)
	}
	firstKey := updated.Picture()
	if firstKey == "" {
		t.Fatal("first picture key is empty")
	}

	updated, err = svc.SetStandPicture(context.Background(), stand.ID(), ports.Picture{
		Content:     bytes.NewReader([]byte("second")),
		ContentType: "image/png",
		Size:        6,
	})
	if err != nil {
		t.Fatalf("second SetStandPicture: %v", err)
	}
	secondKey := updated.Picture()
	if secondKey == firstKey {
		t.Fatal("second picture key should differ from the first")
	}

	found := false
	for _, k := range pictures.deleted {
		if k == firstKey {
			found = true
		}
	}
	if !found {
		t.Errorf("deleted = %v, want it to contain the old key %q", pictures.deleted, firstKey)
	}
	if _, ok := pictures.objects[secondKey]; !ok {
		t.Error("new key not present in storage")
	}
}

func TestSetStandPicture_DeleteFailureIsNonFatal(t *testing.T) {
	repo := newFakeStandRepository()
	idGen := &fakeStandIDGenerator{}
	pictures := newFakePictureStorage()
	pictures.deleteErr = errors.New("boom")
	svc := services.NewStandService(repo, idGen, pictures,
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/png"}})

	stand := newTestStand(t, repo, idGen, "Star Platinum")

	if _, err := svc.SetStandPicture(context.Background(), stand.ID(), ports.Picture{
		Content: bytes.NewReader([]byte("first")), ContentType: "image/png", Size: 5,
	}); err != nil {
		t.Fatalf("first SetStandPicture: %v", err)
	}

	if _, err := svc.SetStandPicture(context.Background(), stand.ID(), ports.Picture{
		Content: bytes.NewReader([]byte("second")), ContentType: "image/png", Size: 6,
	}); err != nil {
		t.Fatalf("second SetStandPicture should succeed despite the delete failure: %v", err)
	}
}

func TestSetStandPicture_RejectsUnsupportedType(t *testing.T) {
	repo := newFakeStandRepository()
	idGen := &fakeStandIDGenerator{}
	pictures := newFakePictureStorage()
	svc := services.NewStandService(repo, idGen, pictures,
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
	svc := services.NewStandService(repo, idGen, pictures,
		services.PicturePolicy{MaxBytes: 1, AllowedTypes: []string{"image/png"}})

	stand := newTestStand(t, repo, idGen, "Killer Queen")

	_, err := svc.SetStandPicture(context.Background(), stand.ID(), ports.Picture{
		Content: bytes.NewReader([]byte("data")), ContentType: "image/png", Size: 4,
	})
	if !errors.Is(err, services.ErrPictureTooLarge) {
		t.Fatalf("err = %v, want ErrPictureTooLarge", err)
	}
}

func TestUpdateStand_PreservesExistingPicture(t *testing.T) {
	repo := newFakeStandRepository()
	idGen := &fakeStandIDGenerator{}
	pictures := newFakePictureStorage()
	svc := services.NewStandService(repo, idGen, pictures,
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/png"}})

	stand := newTestStand(t, repo, idGen, "Gold Experience")
	withPic, err := svc.SetStandPicture(context.Background(), stand.ID(), ports.Picture{
		Content: bytes.NewReader([]byte("data")), ContentType: "image/png", Size: 4,
	})
	if err != nil {
		t.Fatalf("SetStandPicture: %v", err)
	}
	key := withPic.Picture()

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
	if updated.Picture() != key {
		t.Errorf("picture after update = %q, want preserved %q", updated.Picture(), key)
	}
}
