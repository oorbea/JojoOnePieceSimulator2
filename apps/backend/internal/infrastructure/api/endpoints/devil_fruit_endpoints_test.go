package endpoints_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/endpoints"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/streamticket"
)

// fakeDevilFruitRepository is an in-memory ports.IDevilFruitRepository, this
// package's own copy of the fake following the repo's convention of
// duplicating small fakes per test file/package.
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

func (f *fakeDevilFruitRepository) Filter(_ context.Context, filters ports.DevilFruitFilters, _ enums.Locale) ([]*powers.DevilFruit, error) {
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
		if filters.Search != nil {
			needle := strings.ToLower(*filters.Search)
			if !strings.Contains(strings.ToLower(fruit.Name()), needle) &&
				!strings.Contains(strings.ToLower(fruit.Description()), needle) {
				continue
			}
		}
		results = append(results, fruit)
	}
	return results, nil
}

func (f *fakeDevilFruitRepository) Page(ctx context.Context, filters ports.DevilFruitFilters, locale enums.Locale, afterName *string, limit int) ([]*powers.DevilFruit, bool, error) {
	all, err := f.Filter(ctx, filters, locale)
	if err != nil {
		return nil, false, err
	}
	sort.Slice(all, func(i, j int) bool { return all[i].Name() < all[j].Name() })
	start := 0
	if afterName != nil {
		for i, d := range all {
			if d.Name() > *afterName {
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

func (f *fakeDevilFruitRepository) Count(ctx context.Context, filters ports.DevilFruitFilters, locale enums.Locale) (int, error) {
	all, err := f.Filter(ctx, filters, locale)
	return len(all), err
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

func (f *fakeDevilFruitRepository) UpdatePicture(_ context.Context, id powers.PowerID, main, thumb, card, lqip *string, status enums.PictureStatus) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	fruit, ok := f.fruits[id]
	if !ok {
		return ports.ErrDevilFruitNotFound
	}
	newMain, newThumb, newCard, newLqip := fruit.Picture(), fruit.PictureThumb(), fruit.PictureCard(), fruit.PictureLqip()
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
	fruit.SetPictureRenditions(newMain, newThumb, newCard, newLqip, status)
	return nil
}

func (f *fakeDevilFruitRepository) SetMediaID(_ context.Context, id powers.PowerID, mediaID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	fruit, ok := f.fruits[id]
	if !ok {
		return ports.ErrDevilFruitNotFound
	}
	fruit.SetMediaID(mediaID)
	return nil
}

var _ ports.IDevilFruitRepository = (*fakeDevilFruitRepository)(nil)

func validDevilFruitBody(name string) map[string]any {
	return map[string]any{
		"name": name,
		"translations": map[string]any{
			"en-GB": map[string]any{
				"description": name + " description",
				"skills":      []string{"Gear Second"},
			},
		},
		"rarity":    "LEGENDARY",
		"fruitType": "MYTHICAL_ZOAN",
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
	}, nil)
	svc := services.NewDevilFruitService(repo, idGen, pictures, &fakeImageProcessor{}, syncEnqueuer{worker},
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/png"}})

	standEndpoints := endpoints.NewStandEndpoints(services.NewStandService(
		newFakeStandRepository(), &fakeIDGenerator{}, newFakePictureStorage(), &fakeImageProcessor{}, syncEnqueuer{},
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/png"}}))
	devilFruitEndpoints := endpoints.NewDevilFruitEndpoints(svc)
	authEndpoints := endpoints.NewAuthEndpoints(nil, endpoints.CookieConfig{})
	eventsEndpoints := endpoints.NewEventsEndpoints(services.NewPictureEventHub(), fakeTokenIssuer{}, streamticket.NewMemoryStore(streamticket.Config{TTL: 30 * time.Second}), context.Background())
	gameEndpoints := endpoints.NewGameEndpoints(nil, services.NewGameEventHub(), nil, nil, nil, nil, fakeTokenIssuer{}, streamticket.NewMemoryStore(streamticket.Config{TTL: 30 * time.Second}), context.Background(), endpoints.GameWSConfig{})
	stageEndpoints := endpoints.NewStageEndpoints(nil)

	h := endpoints.NewRouter(authEndpoints, standEndpoints, devilFruitEndpoints, endpoints.NewUserEndpoints(nil), eventsEndpoints, gameEndpoints, stageEndpoints, nil, endpoints.NewJojoCharacterEndpoints(nil), endpoints.NewOnePieceCharacterEndpoints(nil), fakeTokenIssuer{},
		endpoints.CORSConfig{}, endpoints.RateLimitConfig{}, endpoints.CacheConfig{}, 0)
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
	var body map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["code"] != "DEVIL_FRUIT_NOT_FOUND" {
		t.Errorf("code = %v, want DEVIL_FRUIT_NOT_FOUND", body["code"])
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

func TestListDevilFruits_SearchByNameAndDescription(t *testing.T) {
	h, _, _ := newDevilFruitTestServer()

	body := validDevilFruitBody("Gomu Gomu no Mi")
	body["translations"].(map[string]any)["en-GB"].(map[string]any)["description"] = "turns the user into rubber"
	doRequest(t, h, http.MethodPost, "/api/v1/devil-fruits", body)
	doRequest(t, h, http.MethodPost, "/api/v1/devil-fruits", validDevilFruitBody("Mera Mera no Mi"))

	rec := doRequest(t, h, http.MethodGet, "/api/v1/devil-fruits?q=gomu", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body = %s", rec.Code, rec.Body.String())
	}
	var byName []map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &byName)
	if len(byName) != 1 || byName[0]["name"] != "Gomu Gomu no Mi" {
		t.Fatalf("q=gomu matched %v, want exactly [Gomu Gomu no Mi]", byName)
	}

	rec = doRequest(t, h, http.MethodGet, "/api/v1/devil-fruits?q=rubber", nil)
	var byDescription []map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &byDescription)
	if len(byDescription) != 1 || byDescription[0]["name"] != "Gomu Gomu no Mi" {
		t.Fatalf("q=rubber matched %v, want exactly [Gomu Gomu no Mi]", byDescription)
	}

	rec = doRequest(t, h, http.MethodGet, "/api/v1/devil-fruits?q=nonexistent", nil)
	var noMatch []map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &noMatch)
	if len(noMatch) != 0 {
		t.Fatalf("len(noMatch) = %d, want 0", len(noMatch))
	}
}

// TestListDevilFruits_NoPageParams_StaysBareArray mirrors
// TestListStands_NoPageParams_StaysBareArray - see that test's doc.
func TestListDevilFruits_NoPageParams_StaysBareArray(t *testing.T) {
	h, _, _ := newDevilFruitTestServer()
	doRequest(t, h, http.MethodPost, "/api/v1/devil-fruits", validDevilFruitBody("Gomu Gomu no Mi"))

	rec := doRequest(t, h, http.MethodGet, "/api/v1/devil-fruits", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	body := rec.Body.Bytes()
	first := ""
	for _, b := range body {
		if b == ' ' || b == '\n' || b == '\t' || b == '\r' {
			continue
		}
		first = string(b)
		break
	}
	if first != "[" {
		t.Fatalf("first non-whitespace byte = %q, want %q (bare array)", first, "[")
	}
}

// TestListDevilFruits_Page_WalkingCursorToExhaustion_EqualsUnpaginatedList
// mirrors the Stand test of the same shape - DevilFruit has no evolves_from
// chain, so there is no T3.6-style ancestor-truncation risk, but the
// exhaustion/ordering contract is identical.
func TestListDevilFruits_Page_WalkingCursorToExhaustion_EqualsUnpaginatedList(t *testing.T) {
	h, _, _ := newDevilFruitTestServer()

	doRequest(t, h, http.MethodPost, "/api/v1/devil-fruits", validDevilFruitBody("Gomu Gomu no Mi"))
	doRequest(t, h, http.MethodPost, "/api/v1/devil-fruits", validDevilFruitBody("Mera Mera no Mi"))
	doRequest(t, h, http.MethodPost, "/api/v1/devil-fruits", validDevilFruitBody("Hito Hito no Mi"))

	unpagedRec := doRequest(t, h, http.MethodGet, "/api/v1/devil-fruits", nil)
	var unpaged []map[string]any
	if err := json.Unmarshal(unpagedRec.Body.Bytes(), &unpaged); err != nil {
		t.Fatalf("unmarshal unpaged: %v", err)
	}

	var pagedNames []string
	cursor := ""
	pages := 0
	for {
		path := "/api/v1/devil-fruits?limit=1"
		if cursor != "" {
			path += "&cursor=" + url.QueryEscape(cursor)
		}
		rec := doRequest(t, h, http.MethodGet, path, nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("page %d: status = %d, body = %s", pages, rec.Code, rec.Body.String())
		}
		var page struct {
			Items      []map[string]any `json:"items"`
			NextCursor *string          `json:"nextCursor"`
			Total      *int             `json:"total"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
			t.Fatalf("page %d: unmarshal: %v", pages, err)
		}
		if pages == 0 && (page.Total == nil || *page.Total != len(unpaged)) {
			t.Errorf("first page total = %v, want %d", page.Total, len(unpaged))
		}
		if pages > 0 && page.Total != nil {
			t.Errorf("page %d: total = %v, want absent on non-first pages", pages, *page.Total)
		}
		if len(page.Items) != 1 {
			t.Fatalf("page %d: len(items) = %d, want 1", pages, len(page.Items))
		}
		pagedNames = append(pagedNames, page.Items[0]["name"].(string))

		pages++
		if page.NextCursor == nil {
			break
		}
		cursor = *page.NextCursor
		if pages > len(unpaged)+1 {
			t.Fatalf("cursor never reached exhaustion after %d pages", pages)
		}
	}

	if len(pagedNames) != len(unpaged) {
		t.Fatalf("paged through %d items, want %d (every unpaginated item exactly once)", len(pagedNames), len(unpaged))
	}
	wantNames := make(map[string]int, len(unpaged))
	for _, d := range unpaged {
		wantNames[d["name"].(string)]++
	}
	gotNames := make(map[string]int, len(pagedNames))
	for _, name := range pagedNames {
		gotNames[name]++
	}
	for name, want := range wantNames {
		if got := gotNames[name]; got != want {
			t.Errorf("%q appeared %d times across pages, want %d", name, got, want)
		}
	}
	for i := 1; i < len(pagedNames); i++ {
		if pagedNames[i-1] >= pagedNames[i] {
			t.Errorf("paged order not strictly increasing at %d: %q >= %q", i, pagedNames[i-1], pagedNames[i])
		}
	}
}

// TestListDevilFruits_Page_CursorFromDifferentFilters_Returns400 mirrors the
// Stand test of the same name.
func TestListDevilFruits_Page_CursorFromDifferentFilters_Returns400(t *testing.T) {
	h, _, _ := newDevilFruitTestServer()
	doRequest(t, h, http.MethodPost, "/api/v1/devil-fruits", validDevilFruitBody("Gomu Gomu no Mi"))
	doRequest(t, h, http.MethodPost, "/api/v1/devil-fruits", validDevilFruitBody("Mera Mera no Mi"))

	rec := doRequest(t, h, http.MethodGet, "/api/v1/devil-fruits?limit=1", nil)
	var page struct {
		NextCursor *string `json:"nextCursor"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &page)
	if page.NextCursor == nil {
		t.Fatal("expected a nextCursor with 2 devil fruits and limit=1")
	}

	replayed := doRequest(t, h, http.MethodGet,
		"/api/v1/devil-fruits?limit=1&fruitType=LOGIA&cursor="+url.QueryEscape(*page.NextCursor), nil)
	if replayed.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d (cursor replayed under different filters), body = %s",
			replayed.Code, http.StatusBadRequest, replayed.Body.String())
	}
}

// TestListDevilFruits_Page_TamperedCursor_Returns400 mirrors the Stand test
// of the same name.
func TestListDevilFruits_Page_TamperedCursor_Returns400(t *testing.T) {
	h, _, _ := newDevilFruitTestServer()
	doRequest(t, h, http.MethodPost, "/api/v1/devil-fruits", validDevilFruitBody("Gomu Gomu no Mi"))

	for _, cursor := range []string{"not-base64!!!", "AAAA"} {
		rec := doRequest(t, h, http.MethodGet, "/api/v1/devil-fruits?limit=1&cursor="+url.QueryEscape(cursor), nil)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("cursor=%q: status = %d, want %d", cursor, rec.Code, http.StatusBadRequest)
		}
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
	if len(pictures.objects) != 3 {
		t.Errorf("uploaded objects = %d, want 3 (main + thumb + card)", len(pictures.objects))
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
