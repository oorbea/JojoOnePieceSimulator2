package endpoints_test

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/characters"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/endpoints"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/streamticket"
)

// fakeCharacterIDGenerator returns deterministic, incrementing ids - this
// package's own copy, following the repo's convention of duplicating small
// fakes per test file/package (see fakeIDGenerator above, typed for
// powers.PowerID instead).
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

// fakeJojoCharacterRepository is an in-memory ports.IJojoCharacterRepository,
// this package's own copy - see fakeDevilFruitRepository's doc.
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
	return c, nil
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

func (f *fakeJojoCharacterRepository) Filter(_ context.Context, filters ports.JojoCharacterFilters, _ enums.Locale) ([]*characters.JojoCharacter, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []*characters.JojoCharacter
	for _, c := range f.byID {
		if filters.Rarity != nil && c.Rarity() != *filters.Rarity {
			continue
		}
		out = append(out, c)
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

func validJojoCharacterBody(name string) map[string]any {
	return map[string]any{
		"name": name,
		"translations": map[string]any{
			"en-GB": map[string]any{"description": name + " description"},
		},
		"rarity":   "RARE",
		"hamon":    "ADVANCED",
		"spin":     "GOLDEN",
		"battleIq": 130,
	}
}

// newCharacterTestServer builds a router exposing /jojo-characters and
// /one-piece-characters wired to fresh in-memory repositories, plus bare
// mounts for the other resources so NewRouter's signature is satisfied -
// same shape as newDevilFruitTestServer.
func newCharacterTestServer() (http.Handler, *fakeJojoCharacterRepository, *fakeOnePieceCharacterRepository) {
	jojoRepo := newFakeJojoCharacterRepository()
	opRepo := newFakeOnePieceCharacterRepository()
	characterIDGen := &fakeCharacterIDGenerator{}
	pictures := newFakePictureStorage()
	targets := map[enums.PictureSubjectKind]services.PictureTarget{
		enums.JojoCharacterSubject:     {Publisher: services.NewJojoCharacterPicturePublisher(jojoRepo), KeyPrefix: "jojo-characters"},
		enums.OnePieceCharacterSubject: {Publisher: services.NewOnePieceCharacterPicturePublisher(opRepo), KeyPrefix: "one-piece-characters"},
	}
	// NewPictureWorker's idGen is only used to name temp media objects, a
	// concern orthogonal to which entity kind owns the picture - so it stays
	// typed to powers.PowerID like every other picture target in this test
	// server, rather than needing its own generic parameter per kind.
	worker := services.NewPictureWorker(&fakeImageProcessor{}, pictures, targets, &fakeIDGenerator{}, services.WorkerConfig{
		Workers: 1, QueueSize: 1, JobTimeout: 5 * time.Second, MaxDimension: 1024, ThumbDimension: 256, Quality: 80,
	}, nil)
	policy := services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/png"}}

	jojoSvc := services.NewJojoCharacterService(jojoRepo, characterIDGen, pictures, &fakeImageProcessor{}, syncEnqueuer{worker}, policy)
	opSvc := services.NewOnePieceCharacterService(opRepo, characterIDGen, pictures, &fakeImageProcessor{}, syncEnqueuer{worker}, policy)

	standEndpoints := endpoints.NewStandEndpoints(services.NewStandService(
		newFakeStandRepository(), &fakeIDGenerator{}, newFakePictureStorage(), &fakeImageProcessor{}, syncEnqueuer{},
		policy))
	authEndpoints := endpoints.NewAuthEndpoints(nil, endpoints.CookieConfig{})
	eventsEndpoints := endpoints.NewEventsEndpoints(services.NewPictureEventHub(), fakeTokenIssuer{}, streamticket.NewMemoryStore(streamticket.Config{TTL: 30 * time.Second}), context.Background())
	gameEndpoints := endpoints.NewGameEndpoints(nil, services.NewGameEventHub(), nil, nil, nil, nil, fakeTokenIssuer{}, streamticket.NewMemoryStore(streamticket.Config{TTL: 30 * time.Second}), context.Background(), endpoints.GameWSConfig{})
	stageEndpoints := endpoints.NewStageEndpoints(nil)

	h := endpoints.NewRouter(authEndpoints, standEndpoints, endpoints.NewDevilFruitEndpoints(nil), endpoints.NewUserEndpoints(nil), eventsEndpoints, gameEndpoints, stageEndpoints, nil,
		endpoints.NewJojoCharacterEndpoints(jojoSvc), endpoints.NewOnePieceCharacterEndpoints(opSvc), fakeTokenIssuer{},
		endpoints.CORSConfig{}, endpoints.RateLimitConfig{}, endpoints.CacheConfig{}, 0)
	return h, jojoRepo, opRepo
}

func TestCreateJojoCharacter(t *testing.T) {
	h, _, _ := newCharacterTestServer()
	rec := doRequest(t, h, http.MethodPost, "/api/v1/jojo-characters", validJojoCharacterBody("Jotaro Kujo"))
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
	if resp["battleIq"] != float64(130) {
		t.Errorf("battleIq = %v, want 130", resp["battleIq"])
	}
}

func TestCreateJojoCharacter_InvalidHamon(t *testing.T) {
	h, _, _ := newCharacterTestServer()
	body := validJojoCharacterBody("Bad Character")
	body["hamon"] = "NOT_A_LEVEL"
	rec := doRequest(t, h, http.MethodPost, "/api/v1/jojo-characters", body)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400, body = %s", rec.Code, rec.Body.String())
	}
}

func TestCreateJojoCharacter_DuplicateName(t *testing.T) {
	h, _, _ := newCharacterTestServer()
	body := validJojoCharacterBody("Joseph Joestar")
	if rec := doRequest(t, h, http.MethodPost, "/api/v1/jojo-characters", body); rec.Code != http.StatusCreated {
		t.Fatalf("first create: status = %d, body = %s", rec.Code, rec.Body.String())
	}
	rec := doRequest(t, h, http.MethodPost, "/api/v1/jojo-characters", body)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409, body = %s", rec.Code, rec.Body.String())
	}
}

func TestGetJojoCharacter_NotFound(t *testing.T) {
	h, _, _ := newCharacterTestServer()
	rec := doRequest(t, h, http.MethodGet, "/api/v1/jojo-characters/00000000-0000-0000-0000-000000000099", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	var body map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["code"] != "JOJO_CHARACTER_NOT_FOUND" {
		t.Errorf("code = %v, want JOJO_CHARACTER_NOT_FOUND", body["code"])
	}
}

func TestUpdateJojoCharacter(t *testing.T) {
	h, _, _ := newCharacterTestServer()
	created := doRequest(t, h, http.MethodPost, "/api/v1/jojo-characters", validJojoCharacterBody("Giorno Giovanna"))
	var createdResp map[string]any
	_ = json.Unmarshal(created.Body.Bytes(), &createdResp)
	id := createdResp["id"].(string)

	body := validJojoCharacterBody("Giorno Giovanna")
	body["battleIq"] = 200
	rec := doRequest(t, h, http.MethodPut, "/api/v1/jojo-characters/"+id, body)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body = %s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["battleIq"] != float64(200) {
		t.Errorf("battleIq after update = %v, want 200", resp["battleIq"])
	}
}

func TestDeleteJojoCharacter(t *testing.T) {
	h, _, _ := newCharacterTestServer()
	created := doRequest(t, h, http.MethodPost, "/api/v1/jojo-characters", validJojoCharacterBody("Rohan Kishibe"))
	var createdResp map[string]any
	_ = json.Unmarshal(created.Body.Bytes(), &createdResp)
	id := createdResp["id"].(string)

	rec := doRequest(t, h, http.MethodDelete, "/api/v1/jojo-characters/"+id, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204, body = %s", rec.Code, rec.Body.String())
	}

	rec = doRequest(t, h, http.MethodGet, "/api/v1/jojo-characters/"+id, nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status after delete = %d, want 404", rec.Code)
	}
}

func TestListJojoCharacters_FiltersByRarity(t *testing.T) {
	h, _, _ := newCharacterTestServer()
	rareBody := validJojoCharacterBody("Josuke Higashikata")
	doRequest(t, h, http.MethodPost, "/api/v1/jojo-characters", rareBody)
	epicBody := validJojoCharacterBody("Jolyne Cujoh")
	epicBody["rarity"] = "EPIC"
	doRequest(t, h, http.MethodPost, "/api/v1/jojo-characters", epicBody)

	rec := doRequest(t, h, http.MethodGet, "/api/v1/jojo-characters?rarity=EPIC", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var listed []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &listed); err != nil {
		t.Fatalf("decoding list body: %v", err)
	}
	if len(listed) != 1 || listed[0]["name"] != "Jolyne Cujoh" {
		t.Errorf("filtered list = %+v, want only Jolyne Cujoh", listed)
	}
}
