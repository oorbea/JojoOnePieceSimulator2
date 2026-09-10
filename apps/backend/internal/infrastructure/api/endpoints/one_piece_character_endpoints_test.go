package endpoints_test

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"sync"
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/characters"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// fakeOnePieceCharacterRepository is an in-memory
// ports.IOnePieceCharacterRepository, this package's own copy - see
// fakeJojoCharacterRepository's doc.
type fakeOnePieceCharacterRepository struct {
	mu           sync.Mutex
	byID         map[characters.CharacterID]*characters.OnePieceCharacter
	translations map[characters.CharacterID]ports.CharacterTranslations
}

func newFakeOnePieceCharacterRepository() *fakeOnePieceCharacterRepository {
	return &fakeOnePieceCharacterRepository{
		byID:         make(map[characters.CharacterID]*characters.OnePieceCharacter),
		translations: make(map[characters.CharacterID]ports.CharacterTranslations),
	}
}

func (f *fakeOnePieceCharacterRepository) Save(_ context.Context, c *characters.OnePieceCharacter, translations ports.CharacterTranslations) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for id, existing := range f.byID {
		if existing.Name() == c.Name() && id != c.ID() {
			return ports.ErrOnePieceCharacterAlreadyExists
		}
	}
	f.byID[c.ID()] = c
	f.translations[c.ID()] = translations
	return nil
}

func (f *fakeOnePieceCharacterRepository) FindByID(_ context.Context, id characters.CharacterID, _ enums.Locale) (*characters.OnePieceCharacter, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	c, ok := f.byID[id]
	if !ok {
		return nil, ports.ErrOnePieceCharacterNotFound
	}
	return c, nil
}

func (f *fakeOnePieceCharacterRepository) FindByName(_ context.Context, name string, _ enums.Locale) (*characters.OnePieceCharacter, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, c := range f.byID {
		if c.Name() == name {
			return c, nil
		}
	}
	return nil, ports.ErrOnePieceCharacterNotFound
}

func (f *fakeOnePieceCharacterRepository) GetAll(_ context.Context, _ enums.Locale) ([]*characters.OnePieceCharacter, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	all := make([]*characters.OnePieceCharacter, 0, len(f.byID))
	for _, c := range f.byID {
		all = append(all, c)
	}
	return all, nil
}

func (f *fakeOnePieceCharacterRepository) Filter(_ context.Context, filters ports.OnePieceCharacterFilters, _ enums.Locale) ([]*characters.OnePieceCharacter, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []*characters.OnePieceCharacter
	for _, c := range f.byID {
		if filters.FruitMastery != nil && c.FruitMastery() != *filters.FruitMastery {
			continue
		}
		out = append(out, c)
	}
	return out, nil
}

func (f *fakeOnePieceCharacterRepository) Page(ctx context.Context, filters ports.OnePieceCharacterFilters, locale enums.Locale, afterName *string, limit int) ([]*characters.OnePieceCharacter, bool, error) {
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

func (f *fakeOnePieceCharacterRepository) Count(ctx context.Context, filters ports.OnePieceCharacterFilters, locale enums.Locale) (int, error) {
	all, err := f.Filter(ctx, filters, locale)
	return len(all), err
}

func (f *fakeOnePieceCharacterRepository) Translations(_ context.Context, id characters.CharacterID) (ports.CharacterTranslations, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	t, ok := f.translations[id]
	if !ok {
		return nil, ports.ErrOnePieceCharacterNotFound
	}
	return t, nil
}

func (f *fakeOnePieceCharacterRepository) Delete(_ context.Context, id characters.CharacterID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.byID[id]; !ok {
		return ports.ErrOnePieceCharacterNotFound
	}
	delete(f.byID, id)
	return nil
}

func (f *fakeOnePieceCharacterRepository) UpdatePicture(_ context.Context, id characters.CharacterID, main, thumb, card, lqip *string, status enums.PictureStatus) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	c, ok := f.byID[id]
	if !ok {
		return ports.ErrOnePieceCharacterNotFound
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

func (f *fakeOnePieceCharacterRepository) SetMediaID(_ context.Context, id characters.CharacterID, mediaID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	c, ok := f.byID[id]
	if !ok {
		return ports.ErrOnePieceCharacterNotFound
	}
	c.SetMediaID(mediaID)
	return nil
}

var _ ports.IOnePieceCharacterRepository = (*fakeOnePieceCharacterRepository)(nil)

func validOnePieceCharacterBody(name string) map[string]any {
	return map[string]any{
		"name": name,
		"translations": map[string]any{
			"en-GB": map[string]any{"description": name + " description"},
		},
		"rarity":          "EPIC",
		"physicalForm":    "MARINE_CAPTAIN",
		"armamentHaki":    "VICE_ADMIRAL",
		"observationHaki": "PRIVATE",
		"conquerorHaki":   "NONE",
		"fruitMastery":    "ADVANCED",
	}
}

func TestCreateOnePieceCharacter(t *testing.T) {
	h, _, _ := newCharacterTestServer()
	rec := doRequest(t, h, http.MethodPost, "/api/v1/one-piece-characters", validOnePieceCharacterBody("Roronoa Zoro"))
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201, body = %s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decoding body: %v", err)
	}
	if resp["fruitMastery"] != "ADVANCED" {
		t.Errorf("fruitMastery = %v, want ADVANCED", resp["fruitMastery"])
	}
}

func TestCreateOnePieceCharacter_InvalidFruitMastery(t *testing.T) {
	h, _, _ := newCharacterTestServer()
	body := validOnePieceCharacterBody("Bad Character")
	body["fruitMastery"] = "NOT_A_MASTERY"
	rec := doRequest(t, h, http.MethodPost, "/api/v1/one-piece-characters", body)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400, body = %s", rec.Code, rec.Body.String())
	}
}

func TestCreateOnePieceCharacter_DuplicateName(t *testing.T) {
	h, _, _ := newCharacterTestServer()
	body := validOnePieceCharacterBody("Nami")
	if rec := doRequest(t, h, http.MethodPost, "/api/v1/one-piece-characters", body); rec.Code != http.StatusCreated {
		t.Fatalf("first create: status = %d, body = %s", rec.Code, rec.Body.String())
	}
	rec := doRequest(t, h, http.MethodPost, "/api/v1/one-piece-characters", body)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409, body = %s", rec.Code, rec.Body.String())
	}
}

func TestGetOnePieceCharacter_NotFound(t *testing.T) {
	h, _, _ := newCharacterTestServer()
	rec := doRequest(t, h, http.MethodGet, "/api/v1/one-piece-characters/00000000-0000-0000-0000-000000000099", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	var body map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["code"] != "ONE_PIECE_CHARACTER_NOT_FOUND" {
		t.Errorf("code = %v, want ONE_PIECE_CHARACTER_NOT_FOUND", body["code"])
	}
}

func TestDeleteOnePieceCharacter(t *testing.T) {
	h, _, _ := newCharacterTestServer()
	created := doRequest(t, h, http.MethodPost, "/api/v1/one-piece-characters", validOnePieceCharacterBody("Usopp"))
	var createdResp map[string]any
	_ = json.Unmarshal(created.Body.Bytes(), &createdResp)
	id := createdResp["id"].(string)

	rec := doRequest(t, h, http.MethodDelete, "/api/v1/one-piece-characters/"+id, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204, body = %s", rec.Code, rec.Body.String())
	}

	rec = doRequest(t, h, http.MethodGet, "/api/v1/one-piece-characters/"+id, nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status after delete = %d, want 404", rec.Code)
	}
}

// TestJojoCharacterEndpoints_DoNotLeakIntoOnePieceCatalogue proves the two
// kinds are served from fully independent stores, not a shared "characters"
// list filtered by manga - creating a JoJo character must never appear in
// the One Piece list or vice versa.
func TestJojoCharacterEndpoints_DoNotLeakIntoOnePieceCatalogue(t *testing.T) {
	h, _, _ := newCharacterTestServer()
	doRequest(t, h, http.MethodPost, "/api/v1/jojo-characters", validJojoCharacterBody("Dio Brando"))
	doRequest(t, h, http.MethodPost, "/api/v1/one-piece-characters", validOnePieceCharacterBody("Monkey D. Luffy"))

	rec := doRequest(t, h, http.MethodGet, "/api/v1/jojo-characters", nil)
	var jojoList []map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &jojoList)
	if len(jojoList) != 1 || jojoList[0]["name"] != "Dio Brando" {
		t.Errorf("jojo list = %+v, want only Dio Brando", jojoList)
	}

	rec = doRequest(t, h, http.MethodGet, "/api/v1/one-piece-characters", nil)
	var opList []map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &opList)
	if len(opList) != 1 || opList[0]["name"] != "Monkey D. Luffy" {
		t.Errorf("one piece list = %+v, want only Monkey D. Luffy", opList)
	}
}
