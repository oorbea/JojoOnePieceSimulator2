package endpoints_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/endpoints"
)

// fakeStageRepository is an in-memory ports.IStageRepository, mirroring
// fakeStandRepository above (same file's package, no DB involved).
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

func (f *fakeStageRepository) Filter(_ context.Context, filters ports.StageFilters, _ enums.Locale) ([]game.Stage, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []game.Stage
	for _, s := range f.stages {
		if filters.Manga != nil && s.Manga() != *filters.Manga {
			continue
		}
		if filters.Search != nil {
			needle := strings.ToLower(*filters.Search)
			if !strings.Contains(strings.ToLower(s.Name()), needle) &&
				!strings.Contains(strings.ToLower(s.Description()), needle) {
				continue
			}
		}
		out = append(out, *s)
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
	for id, existing := range f.stages {
		if existing.Manga() == s.Manga() && existing.Name() == s.Name() && id != s.ID() {
			return ports.ErrStageAlreadyExists
		}
	}
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

// fakeStageIDGenerator returns deterministic, incrementing ids, kept
// separate from fakeIDGenerator (powers.PowerID) since it targets a
// different id type.
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

// newStageTestServer builds a router with only the stage service backed by
// a real StageService + fake repo; Stand/DevilFruit/User/Game are nil, same
// convention as newTestServer's NewStageEndpoints(nil) - unused routes are
// never exercised.
func newStageTestServer() http.Handler {
	return newStageTestServerWithDeps(endpoints.RateLimitConfig{}, newFakePictureStorage())
}

func newStageTestServerWithDeps(rateCfg endpoints.RateLimitConfig, pictures *fakePictureStorage) http.Handler {
	repo := newFakeStageRepository()
	stageIDGen := &fakeStageIDGenerator{}
	// The worker's idGen only mints the random uuid used to name the
	// uploaded object key (see picture_worker.go) - it's always a
	// powers.PowerID regardless of subject kind, so it's a separate
	// generator from the one StageService uses to mint Stage ids.
	workerIDGen := &fakeIDGenerator{}
	targets := map[enums.PictureSubjectKind]services.PictureTarget{
		enums.StageSubject: {Publisher: services.NewStagePicturePublisher(repo), KeyPrefix: "stages"},
	}
	worker := services.NewPictureWorker(&fakeImageProcessor{}, pictures, targets, workerIDGen, services.WorkerConfig{
		Workers: 1, QueueSize: 1, JobTimeout: 5 * time.Second, MaxDimension: 1024, ThumbDimension: 256, Quality: 80,
	}, nil)
	svc := services.NewStageService(repo, stageIDGen, pictures, &fakeImageProcessor{}, syncEnqueuer{worker},
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/webp", "image/avif", "image/jpeg", "image/png", "image/gif"}})
	stageEndpoints := endpoints.NewStageEndpoints(svc)
	authEndpoints := endpoints.NewAuthEndpoints(nil)
	eventsEndpoints := endpoints.NewEventsEndpoints(services.NewPictureEventHub(), fakeTokenIssuer{}, context.Background())
	return endpoints.NewRouter(authEndpoints, endpoints.NewStandEndpoints(nil), endpoints.NewDevilFruitEndpoints(nil), endpoints.NewUserEndpoints(nil), eventsEndpoints,
		endpoints.NewGameEndpoints(nil, services.NewGameEventHub(), nil, nil, fakeTokenIssuer{}, context.Background(), endpoints.GameWSConfig{}),
		stageEndpoints, fakeTokenIssuer{}, endpoints.CORSConfig{}, rateCfg, endpoints.CacheConfig{})
}

func validStageBody(name string) map[string]any {
	return map[string]any{
		"manga": "JOJO",
		"order": 1,
		"name":  name,
		"translations": map[string]any{
			"en-GB": map[string]any{"description": name + " description"},
			"es-ES": map[string]any{"description": name + " descripcion"},
			"ca-ES": map[string]any{"description": name + " descripcio"},
		},
	}
}

func TestCreateStage(t *testing.T) {
	h := newStageTestServer()

	rec := doRequest(t, h, http.MethodPost, "/api/v1/stages", validStageBody("Stardust Crusaders"))
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp["name"] != "Stardust Crusaders" {
		t.Errorf("name = %v, want Stardust Crusaders", resp["name"])
	}
	if resp["manga"] != "JOJO" {
		t.Errorf("manga = %v, want JOJO", resp["manga"])
	}
}

func TestCreateStage_MissingLocale(t *testing.T) {
	h := newStageTestServer()

	body := validStageBody("Diamond is Unbreakable")
	delete(body["translations"].(map[string]any), "ca-ES")

	rec := doRequest(t, h, http.MethodPost, "/api/v1/stages", body)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	var respBody map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &respBody)
	if respBody["code"] != "VALIDATION_FAILED" {
		t.Errorf("code = %v, want VALIDATION_FAILED", respBody["code"])
	}
}

func TestCreateStage_InvalidManga(t *testing.T) {
	h := newStageTestServer()

	body := validStageBody("Golden Wind")
	body["manga"] = "NOT_A_MANGA"

	rec := doRequest(t, h, http.MethodPost, "/api/v1/stages", body)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestCreateStage_DuplicateNameSameManga(t *testing.T) {
	h := newStageTestServer()

	doRequest(t, h, http.MethodPost, "/api/v1/stages", validStageBody("Stone Ocean"))
	rec := doRequest(t, h, http.MethodPost, "/api/v1/stages", validStageBody("Stone Ocean"))
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusConflict, rec.Body.String())
	}
	var respBody map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &respBody)
	if respBody["code"] != "STAGE_ALREADY_EXISTS" {
		t.Errorf("code = %v, want STAGE_ALREADY_EXISTS", respBody["code"])
	}
}

func TestGetStage_NotFound(t *testing.T) {
	h := newStageTestServer()

	rec := doRequest(t, h, http.MethodGet, "/api/v1/stages/00000000-0000-0000-0000-000000000000", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
	var body map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["code"] != "STAGE_NOT_FOUND" {
		t.Errorf("code = %v, want STAGE_NOT_FOUND", body["code"])
	}
}

func TestGetStage_InvalidID(t *testing.T) {
	h := newStageTestServer()

	rec := doRequest(t, h, http.MethodGet, "/api/v1/stages/not-a-uuid", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestListAndFilterStagesByManga(t *testing.T) {
	h := newStageTestServer()

	doRequest(t, h, http.MethodPost, "/api/v1/stages", validStageBody("Phantom Blood"))
	body := validStageBody("Alabasta")
	body["manga"] = "ONE_PIECE"
	doRequest(t, h, http.MethodPost, "/api/v1/stages", body)

	rec := doRequest(t, h, http.MethodGet, "/api/v1/stages", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var all []map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &all)
	if len(all) != 2 {
		t.Fatalf("len(all) = %d, want 2", len(all))
	}

	rec = doRequest(t, h, http.MethodGet, "/api/v1/stages?manga=ONE_PIECE", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var filtered []map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &filtered)
	if len(filtered) != 1 || filtered[0]["name"] != "Alabasta" {
		t.Fatalf("filtered = %v, want just Alabasta", filtered)
	}
}

func TestListStages_SearchByNameAndDescription(t *testing.T) {
	h := newStageTestServer()

	body := validStageBody("Phantom Blood")
	body["translations"].(map[string]any)["en-GB"].(map[string]any)["description"] = "the tale begins in Victorian England"
	doRequest(t, h, http.MethodPost, "/api/v1/stages", body)
	alabasta := validStageBody("Alabasta")
	alabasta["manga"] = "ONE_PIECE"
	doRequest(t, h, http.MethodPost, "/api/v1/stages", alabasta)

	rec := doRequest(t, h, http.MethodGet, "/api/v1/stages?q=phantom", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var byName []map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &byName)
	if len(byName) != 1 || byName[0]["name"] != "Phantom Blood" {
		t.Fatalf("q=phantom matched %v, want exactly [Phantom Blood]", byName)
	}

	rec = doRequest(t, h, http.MethodGet, "/api/v1/stages?q=victorian", nil)
	var byDescription []map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &byDescription)
	if len(byDescription) != 1 || byDescription[0]["name"] != "Phantom Blood" {
		t.Fatalf("q=victorian matched %v, want exactly [Phantom Blood]", byDescription)
	}
}

func TestListStages_MangaAndSearchCombined(t *testing.T) {
	h := newStageTestServer()

	doRequest(t, h, http.MethodPost, "/api/v1/stages", validStageBody("Phantom Blood"))
	alabasta := validStageBody("Alabasta")
	alabasta["manga"] = "ONE_PIECE"
	doRequest(t, h, http.MethodPost, "/api/v1/stages", alabasta)

	// "Phantom" only exists on the JOJO side - combined with manga=ONE_PIECE
	// it must match nothing, proving both filters are applied together
	// (AND), not either one alone.
	rec := doRequest(t, h, http.MethodGet, "/api/v1/stages?manga=ONE_PIECE&q=phantom", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var none []map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &none)
	if len(none) != 0 {
		t.Fatalf("len(none) = %d, want 0", len(none))
	}

	rec = doRequest(t, h, http.MethodGet, "/api/v1/stages?manga=ONE_PIECE&q=alabasta", nil)
	var matched []map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &matched)
	if len(matched) != 1 || matched[0]["name"] != "Alabasta" {
		t.Fatalf("matched = %v, want exactly [Alabasta]", matched)
	}
}

func TestListStages_InvalidManga(t *testing.T) {
	h := newStageTestServer()

	rec := doRequest(t, h, http.MethodGet, "/api/v1/stages?manga=NOT_A_MANGA", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestUpdateStage(t *testing.T) {
	h := newStageTestServer()

	createRec := doRequest(t, h, http.MethodPost, "/api/v1/stages", validStageBody("Vento Aureo"))
	var created map[string]any
	_ = json.Unmarshal(createRec.Body.Bytes(), &created)
	id := created["id"].(string)

	updateBody := validStageBody("Vento Aureo")
	updateBody["translations"].(map[string]any)["en-GB"].(map[string]any)["description"] = "updated description"
	rec := doRequest(t, h, http.MethodPut, "/api/v1/stages/"+id, updateBody)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var got map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &got)
	if got["description"] != "updated description" {
		t.Errorf("description = %v, want %q", got["description"], "updated description")
	}
	if got["id"] != id {
		t.Errorf("id changed on update: got %v, want %v", got["id"], id)
	}
}

func TestUpdateStage_NotFound(t *testing.T) {
	h := newStageTestServer()

	rec := doRequest(t, h, http.MethodPut, "/api/v1/stages/00000000-0000-0000-0000-000000000000", validStageBody("Ghost"))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestDeleteStage(t *testing.T) {
	h := newStageTestServer()

	createRec := doRequest(t, h, http.MethodPost, "/api/v1/stages", validStageBody("Jojolion"))
	var created map[string]any
	_ = json.Unmarshal(createRec.Body.Bytes(), &created)
	id := created["id"].(string)

	rec := doRequest(t, h, http.MethodDelete, "/api/v1/stages/"+id, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusNoContent, rec.Body.String())
	}

	rec = doRequest(t, h, http.MethodGet, "/api/v1/stages/"+id, nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status after delete = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestStageTranslations(t *testing.T) {
	h := newStageTestServer()

	createRec := doRequest(t, h, http.MethodPost, "/api/v1/stages", validStageBody("Battle Tendency"))
	var created map[string]any
	_ = json.Unmarshal(createRec.Body.Bytes(), &created)
	id := created["id"].(string)

	rec := doRequest(t, h, http.MethodGet, "/api/v1/stages/"+id+"/translations", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &got)
	translations, ok := got["translations"].(map[string]any)
	if !ok || len(translations) != 3 {
		t.Fatalf("translations = %v, want all 3 locales", got["translations"])
	}
}

func TestPatchStagePicture(t *testing.T) {
	pictures := newFakePictureStorage()
	h := newStageTestServerWithDeps(endpoints.RateLimitConfig{}, pictures)

	createRec := doRequest(t, h, http.MethodPost, "/api/v1/stages", validStageBody("Steel Ball Run"))
	var created map[string]any
	_ = json.Unmarshal(createRec.Body.Bytes(), &created)
	id := created["id"].(string)
	if created["picture"] != "" {
		t.Errorf("picture on create = %v, want empty", created["picture"])
	}

	rec := doMultipartRequest(t, h, http.MethodPatch, "/api/v1/stages/"+id+"/picture", "picture", "stage.png", pngBytes, "admin-token")
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusAccepted, rec.Body.String())
	}

	getRec := doRequest(t, h, http.MethodGet, "/api/v1/stages/"+id, nil)
	var getGot map[string]any
	_ = json.Unmarshal(getRec.Body.Bytes(), &getGot)
	if getGot["pictureStatus"] != "READY" {
		t.Fatalf("pictureStatus after worker ran = %v, want READY", getGot["pictureStatus"])
	}
	if picture, _ := getGot["picture"].(string); picture == "" {
		t.Error("picture should be set once READY")
	}
}

func TestPatchStagePicture_UnsupportedType(t *testing.T) {
	h := newStageTestServer()
	createRec := doRequest(t, h, http.MethodPost, "/api/v1/stages", validStageBody("Thus Spoke Kishibe Rohan"))
	var created map[string]any
	_ = json.Unmarshal(createRec.Body.Bytes(), &created)
	id := created["id"].(string)

	rec := doMultipartRequest(t, h, http.MethodPatch, "/api/v1/stages/"+id+"/picture", "picture", "notes.txt", []byte("plain text content"), "admin-token")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestPatchStagePicture_RequiresAdmin(t *testing.T) {
	h := newStageTestServer()
	createRec := doRequest(t, h, http.MethodPost, "/api/v1/stages", validStageBody("Baroque Works"))
	var created map[string]any
	_ = json.Unmarshal(createRec.Body.Bytes(), &created)
	id := created["id"].(string)

	rec := doMultipartRequest(t, h, http.MethodPatch, "/api/v1/stages/"+id+"/picture", "picture", "stage.png", pngBytes, "user-token")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestPatchStagePicture_NotFound(t *testing.T) {
	h := newStageTestServer()

	rec := doMultipartRequest(t, h, http.MethodPatch, "/api/v1/stages/00000000-0000-0000-0000-000000000000/picture", "picture", "stage.png", pngBytes, "admin-token")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestStageRoutes_RequireAuth(t *testing.T) {
	h := newStageTestServer()

	rec := noAuthRequest(t, h, http.MethodGet, "/api/v1/stages")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("GET / without token: status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	rec = noAuthRequest(t, h, http.MethodPost, "/api/v1/stages")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("POST / without token: status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestStageRoutes_RegularUserCanRead(t *testing.T) {
	h := newStageTestServer()
	doRequest(t, h, http.MethodPost, "/api/v1/stages", validStageBody("Water Seven"))

	rec := userAuthRequest(t, h, http.MethodGet, "/api/v1/stages", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestStageRoutes_RegularUserCannotWrite(t *testing.T) {
	h := newStageTestServer()

	rec := userAuthRequest(t, h, http.MethodPost, "/api/v1/stages", validStageBody("Marineford"))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("POST: status = %d, want %d, body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}

	createRec := doRequest(t, h, http.MethodPost, "/api/v1/stages", validStageBody("Dressrosa"))
	var created map[string]any
	_ = json.Unmarshal(createRec.Body.Bytes(), &created)
	id := created["id"].(string)

	rec = userAuthRequest(t, h, http.MethodPut, "/api/v1/stages/"+id, validStageBody("Dressrosa"))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("PUT: status = %d, want %d, body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}

	rec = userAuthRequest(t, h, http.MethodDelete, "/api/v1/stages/"+id, nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("DELETE: status = %d, want %d, body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}
