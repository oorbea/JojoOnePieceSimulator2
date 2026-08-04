package endpoints_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/endpoints"
)

// fakeDevilFruitRepository is an in-memory ports.IDevilFruitRepository, this
// package's own copy of the fake following the repo's convention of
// duplicating small fakes per test file/package.
type fakeDevilFruitRepository struct {
	mu     sync.Mutex
	fruits map[powers.PowerID]*powers.DevilFruit
}

func newFakeDevilFruitRepository() *fakeDevilFruitRepository {
	return &fakeDevilFruitRepository{fruits: make(map[powers.PowerID]*powers.DevilFruit)}
}

func (f *fakeDevilFruitRepository) Save(_ context.Context, fruit *powers.DevilFruit) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for id, existing := range f.fruits {
		if existing.Name() == fruit.Name() && id != fruit.ID() {
			return ports.ErrDevilFruitAlreadyExists
		}
	}
	f.fruits[fruit.ID()] = fruit
	return nil
}

func (f *fakeDevilFruitRepository) FindByID(_ context.Context, id powers.PowerID) (*powers.DevilFruit, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	fruit, ok := f.fruits[id]
	if !ok {
		return nil, ports.ErrDevilFruitNotFound
	}
	cp := *fruit
	return &cp, nil
}

func (f *fakeDevilFruitRepository) FindByName(_ context.Context, name string) (*powers.DevilFruit, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, fruit := range f.fruits {
		if fruit.Name() == name {
			return fruit, nil
		}
	}
	return nil, ports.ErrDevilFruitNotFound
}

func (f *fakeDevilFruitRepository) GetAll(_ context.Context) ([]*powers.DevilFruit, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	all := make([]*powers.DevilFruit, 0, len(f.fruits))
	for _, fruit := range f.fruits {
		all = append(all, fruit)
	}
	return all, nil
}

func (f *fakeDevilFruitRepository) Filter(_ context.Context, filters ports.DevilFruitFilters) ([]*powers.DevilFruit, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var results []*powers.DevilFruit
	for _, fruit := range f.fruits {
		if filters.Rarity != nil && fruit.Rarity() != *filters.Rarity {
			continue
		}
		if filters.FruitType != nil && fruit.FruitType() != *filters.FruitType {
			continue
		}
		results = append(results, fruit)
	}
	return results, nil
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

func validDevilFruitBody(name string) map[string]any {
	return map[string]any{
		"name":        name,
		"description": name + " description",
		"rarity":      "LEGENDARY",
		"skills":      []string{"Gear Second"},
		"fruitType":   "MYTHICAL_ZOAN",
	}
}

// newDevilFruitTestServer builds a router exposing only the pieces
// TestDevilFruit* tests need: /devil-fruits wired to a fresh in-memory
// repository, plus a bare /stands mount so NewRouter's signature is
// satisfied.
func newDevilFruitTestServer() (http.Handler, *fakeDevilFruitRepository, *fakePictureStorage) {
	repo := newFakeDevilFruitRepository()
	idGen := &fakeIDGenerator{}
	pictures := newFakePictureStorage()
	targets := map[enums.PictureSubjectKind]services.PictureTarget{
		enums.DevilFruitSubject: {Publisher: services.NewDevilFruitPicturePublisher(repo), KeyPrefix: "devil-fruits"},
	}
	worker := services.NewPictureWorker(&fakeImageProcessor{}, pictures, targets, idGen, services.WorkerConfig{
		Workers: 1, QueueSize: 1, JobTimeout: 5 * time.Second, MaxDimension: 1024, ThumbDimension: 256, Quality: 80,
	})
	svc := services.NewDevilFruitService(repo, idGen, pictures, &fakeImageProcessor{}, syncEnqueuer{worker},
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/png"}})

	standEndpoints := endpoints.NewStandEndpoints(services.NewStandService(
		newFakeStandRepository(), &fakeIDGenerator{}, newFakePictureStorage(), &fakeImageProcessor{}, syncEnqueuer{},
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/png"}}))
	devilFruitEndpoints := endpoints.NewDevilFruitEndpoints(svc)
	authEndpoints := endpoints.NewAuthEndpoints(nil)

	h := endpoints.NewRouter(authEndpoints, standEndpoints, devilFruitEndpoints, endpoints.NewUserEndpoints(nil), fakeTokenIssuer{},
		endpoints.CORSConfig{}, endpoints.RateLimitConfig{}, endpoints.CacheConfig{})
	return h, repo, pictures
}

func TestCreateDevilFruit(t *testing.T) {
	h, _, _ := newDevilFruitTestServer()
	rec := doRequest(t, h, http.MethodPost, "/api/v1/devil-fruits", validDevilFruitBody("Gomu Gomu no Mi"))
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201, body = %s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("Location") == "" {
		t.Error("missing Location header")
	}
	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decoding body: %v", err)
	}
	if resp["fruitType"] != "MYTHICAL_ZOAN" {
		t.Errorf("fruitType = %v, want MYTHICAL_ZOAN", resp["fruitType"])
	}
}

func TestCreateDevilFruit_InvalidFruitType(t *testing.T) {
	h, _, _ := newDevilFruitTestServer()
	body := validDevilFruitBody("Bad Fruit")
	body["fruitType"] = "NOT_A_TYPE"
	rec := doRequest(t, h, http.MethodPost, "/api/v1/devil-fruits", body)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400, body = %s", rec.Code, rec.Body.String())
	}
}

func TestCreateDevilFruit_InvalidRarity(t *testing.T) {
	h, _, _ := newDevilFruitTestServer()
	body := validDevilFruitBody("Bad Rarity Fruit")
	body["rarity"] = "NOT_A_RARITY"
	rec := doRequest(t, h, http.MethodPost, "/api/v1/devil-fruits", body)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400, body = %s", rec.Code, rec.Body.String())
	}
}

func TestCreateDevilFruit_DuplicateName(t *testing.T) {
	h, _, _ := newDevilFruitTestServer()
	body := validDevilFruitBody("Mera Mera no Mi")
	if rec := doRequest(t, h, http.MethodPost, "/api/v1/devil-fruits", body); rec.Code != http.StatusCreated {
		t.Fatalf("first create: status = %d, body = %s", rec.Code, rec.Body.String())
	}
	rec := doRequest(t, h, http.MethodPost, "/api/v1/devil-fruits", body)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409, body = %s", rec.Code, rec.Body.String())
	}
}

func TestGetDevilFruit_NotFound(t *testing.T) {
	h, _, _ := newDevilFruitTestServer()
	rec := doRequest(t, h, http.MethodGet, "/api/v1/devil-fruits/00000000-0000-0000-0000-000000000099", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestGetDevilFruit_InvalidID(t *testing.T) {
	h, _, _ := newDevilFruitTestServer()
	rec := doRequest(t, h, http.MethodGet, "/api/v1/devil-fruits/not-a-uuid", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestListAndFilterDevilFruits(t *testing.T) {
	h, _, _ := newDevilFruitTestServer()
	doRequest(t, h, http.MethodPost, "/api/v1/devil-fruits", validDevilFruitBody("Hito Hito no Mi"))

	rec := doRequest(t, h, http.MethodGet, "/api/v1/devil-fruits", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("list status = %d, want 200", rec.Code)
	}
	var listed []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &listed); err != nil {
		t.Fatalf("decoding list body: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("len(listed) = %d, want 1", len(listed))
	}

	rec = doRequest(t, h, http.MethodGet, "/api/v1/devil-fruits?fruitType=MYTHICAL_ZOAN", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("filter status = %d, want 200", rec.Code)
	}
	var filtered []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &filtered); err != nil {
		t.Fatalf("decoding filter body: %v", err)
	}
	if len(filtered) != 1 {
		t.Fatalf("len(filtered) = %d, want 1", len(filtered))
	}

	rec = doRequest(t, h, http.MethodGet, "/api/v1/devil-fruits?fruitType=LOGIA", nil)
	var empty []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &empty); err != nil {
		t.Fatalf("decoding empty-filter body: %v", err)
	}
	if len(empty) != 0 {
		t.Fatalf("len(empty) = %d, want 0", len(empty))
	}
}

func TestUpdateDevilFruit(t *testing.T) {
	h, _, _ := newDevilFruitTestServer()
	createRec := doRequest(t, h, http.MethodPost, "/api/v1/devil-fruits", validDevilFruitBody("Yami Yami no Mi"))
	var created map[string]any
	_ = json.Unmarshal(createRec.Body.Bytes(), &created)
	id := created["id"].(string)

	body := validDevilFruitBody("Yami Yami no Mi")
	body["fruitType"] = "LOGIA"
	rec := doRequest(t, h, http.MethodPut, "/api/v1/devil-fruits/"+id, body)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body = %s", rec.Code, rec.Body.String())
	}
	var updated map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &updated)
	if updated["fruitType"] != "LOGIA" {
		t.Errorf("fruitType = %v, want LOGIA", updated["fruitType"])
	}
}

func TestUpdateDevilFruit_NotFound(t *testing.T) {
	h, _, _ := newDevilFruitTestServer()
	rec := doRequest(t, h, http.MethodPut, "/api/v1/devil-fruits/00000000-0000-0000-0000-000000000099", validDevilFruitBody("Ghost"))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestDeleteDevilFruit(t *testing.T) {
	h, _, _ := newDevilFruitTestServer()
	createRec := doRequest(t, h, http.MethodPost, "/api/v1/devil-fruits", validDevilFruitBody("Suna Suna no Mi"))
	var created map[string]any
	_ = json.Unmarshal(createRec.Body.Bytes(), &created)
	id := created["id"].(string)

	rec := doRequest(t, h, http.MethodDelete, "/api/v1/devil-fruits/"+id, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
	rec = doRequest(t, h, http.MethodGet, "/api/v1/devil-fruits/"+id, nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status after delete = %d, want 404", rec.Code)
	}
}

func TestPatchDevilFruitPicture(t *testing.T) {
	h, _, pictures := newDevilFruitTestServer()
	createRec := doRequest(t, h, http.MethodPost, "/api/v1/devil-fruits", validDevilFruitBody("Gura Gura no Mi"))
	var created map[string]any
	_ = json.Unmarshal(createRec.Body.Bytes(), &created)
	id := created["id"].(string)

	rec := doMultipartRequest(t, h, http.MethodPatch, "/api/v1/devil-fruits/"+id+"/picture", "picture", "pic.png", pngBytes, "admin-token")
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want 202, body = %s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["pictureStatus"] != "PENDING" {
		t.Errorf("pictureStatus = %v, want PENDING", resp["pictureStatus"])
	}

	getRec := doRequest(t, h, http.MethodGet, "/api/v1/devil-fruits/"+id, nil)
	var got map[string]any
	_ = json.Unmarshal(getRec.Body.Bytes(), &got)
	if got["pictureStatus"] != "READY" {
		t.Fatalf("pictureStatus after sync worker run = %v, want READY", got["pictureStatus"])
	}
	if len(pictures.objects) != 2 {
		t.Errorf("uploaded objects = %d, want 2 (main + thumb)", len(pictures.objects))
	}
}

func TestPatchDevilFruitPicture_UnsupportedType(t *testing.T) {
	h, _, _ := newDevilFruitTestServer()
	createRec := doRequest(t, h, http.MethodPost, "/api/v1/devil-fruits", validDevilFruitBody("Ope Ope no Mi"))
	var created map[string]any
	_ = json.Unmarshal(createRec.Body.Bytes(), &created)
	id := created["id"].(string)

	rec := doMultipartRequest(t, h, http.MethodPatch, "/api/v1/devil-fruits/"+id+"/picture", "picture", "pic.txt", []byte("not an image"), "admin-token")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400, body = %s", rec.Code, rec.Body.String())
	}
}

func TestPatchDevilFruitPicture_RequiresAdmin(t *testing.T) {
	h, _, _ := newDevilFruitTestServer()
	createRec := doRequest(t, h, http.MethodPost, "/api/v1/devil-fruits", validDevilFruitBody("Nikyu Nikyu no Mi"))
	var created map[string]any
	_ = json.Unmarshal(createRec.Body.Bytes(), &created)
	id := created["id"].(string)

	rec := doMultipartRequest(t, h, http.MethodPatch, "/api/v1/devil-fruits/"+id+"/picture", "picture", "pic.png", pngBytes, "user-token")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}

func TestDevilFruitRoutes_RequireAuth(t *testing.T) {
	h, _, _ := newDevilFruitTestServer()
	rec := noAuthRequest(t, h, http.MethodGet, "/api/v1/devil-fruits")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestDevilFruitRoutes_RegularUserCanRead(t *testing.T) {
	h, _, _ := newDevilFruitTestServer()
	rec := userAuthRequest(t, h, http.MethodGet, "/api/v1/devil-fruits", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestDevilFruitRoutes_RegularUserCannotWrite(t *testing.T) {
	h, _, _ := newDevilFruitTestServer()
	rec := userAuthRequest(t, h, http.MethodPost, "/api/v1/devil-fruits", validDevilFruitBody("Forbidden Fruit"))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}

func TestGetDevilFruits_ETag_MatchingIfNoneMatchReturns304(t *testing.T) {
	h, _, _ := newDevilFruitTestServer()
	doRequest(t, h, http.MethodPost, "/api/v1/devil-fruits", validDevilFruitBody("Hie Hie no Mi"))

	first := doRequest(t, h, http.MethodGet, "/api/v1/devil-fruits", nil)
	etag := first.Header().Get("ETag")
	if etag == "" {
		t.Fatal("first GET: no ETag header set")
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/devil-fruits", nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	req.Header.Set("If-None-Match", etag)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotModified {
		t.Fatalf("second GET with If-None-Match: status = %d, want %d", rec.Code, http.StatusNotModified)
	}
}
