package endpoints_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/endpoints"
)

// fakeStandRepository is an in-memory ports.IStandRepository, since the
// existing test suite hits a real Postgres for the repository itself and
// has no mock library - these endpoint tests instead exercise the real
// StandService against a handwritten fake, with no DB involved.
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
	for id, existing := range f.stands {
		if existing.Name() == stand.Name() && id != stand.ID() {
			return ports.ErrStandAlreadyExists
		}
	}
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

func (f *fakeStandRepository) Filter(_ context.Context, filters ports.StandFilters) ([]*powers.Stand, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var results []*powers.Stand
	for _, stand := range f.stands {
		if filters.Rarity != nil && stand.Rarity() != *filters.Rarity {
			continue
		}
		if filters.AttackPower != nil && stand.AttackPower() != *filters.AttackPower {
			continue
		}
		if filters.Speed != nil && stand.Speed() != *filters.Speed {
			continue
		}
		results = append(results, stand)
	}
	return results, nil
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

// fakeIDGenerator returns deterministic, incrementing ids so test assertions
// don't need real randomness.
type fakeIDGenerator struct {
	mu   sync.Mutex
	next byte
}

func (g *fakeIDGenerator) NewID() powers.PowerID {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.next++
	var id powers.PowerID
	id[15] = g.next
	return id
}

func newTestServer() http.Handler {
	repo := newFakeStandRepository()
	svc := services.NewStandService(repo, &fakeIDGenerator{})
	return endpoints.NewRouter(endpoints.NewStandEndpoints(svc))
}

func validStandBody(name string) map[string]any {
	return map[string]any{
		"name":        name,
		"description": name + " description",
		"rarity":      "RARE",
		"skills":      []string{"punch", "dash"},
		"picture":     "pic.png",
		"attackPower": "A",
		"speed":       "B",
		"attackRange": "C",
		"endurance":   "D",
		"precision":   "E",
		"potential":   "INFINITE",
	}
}

func doRequest(t *testing.T, h http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var reqBody *bytes.Buffer
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		reqBody = bytes.NewBuffer(b)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}
	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestCreateStand(t *testing.T) {
	h := newTestServer()

	rec := doRequest(t, h, http.MethodPost, "/api/v1/stands", validStandBody("Silver Chariot"))
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if loc := rec.Header().Get("Location"); loc == "" {
		t.Error("Location header not set")
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp["name"] != "Silver Chariot" {
		t.Errorf("name = %v, want Silver Chariot", resp["name"])
	}
	if resp["evolvesFrom"] != nil {
		t.Errorf("evolvesFrom = %v, want nil", resp["evolvesFrom"])
	}
}

func TestCreateStand_InvalidEnum(t *testing.T) {
	h := newTestServer()

	body := validStandBody("Star Platinum")
	body["rarity"] = "NOT_A_RARITY"

	rec := doRequest(t, h, http.MethodPost, "/api/v1/stands", body)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestCreateStand_DuplicateName(t *testing.T) {
	h := newTestServer()

	doRequest(t, h, http.MethodPost, "/api/v1/stands", validStandBody("The World"))
	rec := doRequest(t, h, http.MethodPost, "/api/v1/stands", validStandBody("The World"))
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusConflict, rec.Body.String())
	}
}

func TestGetStand_WithEvolvesFrom(t *testing.T) {
	h := newTestServer()

	parentRec := doRequest(t, h, http.MethodPost, "/api/v1/stands", validStandBody("Silver Chariot"))
	var parent map[string]any
	_ = json.Unmarshal(parentRec.Body.Bytes(), &parent)
	parentID := parent["id"].(string)

	childBody := validStandBody("Silver Chariot Requiem")
	childBody["evolvesFromId"] = parentID
	childRec := doRequest(t, h, http.MethodPost, "/api/v1/stands", childBody)
	var child map[string]any
	_ = json.Unmarshal(childRec.Body.Bytes(), &child)
	childID := child["id"].(string)

	rec := doRequest(t, h, http.MethodGet, "/api/v1/stands/"+childID, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	evolvesFrom, ok := got["evolvesFrom"].(map[string]any)
	if !ok {
		t.Fatalf("evolvesFrom = %v, want nested object", got["evolvesFrom"])
	}
	if evolvesFrom["name"] != "Silver Chariot" {
		t.Errorf("evolvesFrom.name = %v, want Silver Chariot", evolvesFrom["name"])
	}
}

func TestGetStand_NotFound(t *testing.T) {
	h := newTestServer()

	rec := doRequest(t, h, http.MethodGet, "/api/v1/stands/00000000-0000-0000-0000-000000000000", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestGetStand_InvalidID(t *testing.T) {
	h := newTestServer()

	rec := doRequest(t, h, http.MethodGet, "/api/v1/stands/not-a-uuid", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestListAndFilterStands(t *testing.T) {
	h := newTestServer()

	doRequest(t, h, http.MethodPost, "/api/v1/stands", validStandBody("Silver Chariot"))
	doRequest(t, h, http.MethodPost, "/api/v1/stands", validStandBody("Star Platinum"))

	rec := doRequest(t, h, http.MethodGet, "/api/v1/stands", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var all []map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &all)
	if len(all) != 2 {
		t.Fatalf("len(all) = %d, want 2", len(all))
	}

	rec = doRequest(t, h, http.MethodGet, "/api/v1/stands?attackPower=A", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var filtered []map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &filtered)
	if len(filtered) != 2 {
		t.Fatalf("len(filtered) = %d, want 2 (both stands have attackPower=A)", len(filtered))
	}
}

func TestUpdateStand(t *testing.T) {
	h := newTestServer()

	createRec := doRequest(t, h, http.MethodPost, "/api/v1/stands", validStandBody("Silver Chariot"))
	var created map[string]any
	_ = json.Unmarshal(createRec.Body.Bytes(), &created)
	id := created["id"].(string)

	updateBody := validStandBody("Silver Chariot")
	updateBody["description"] = "updated description"
	rec := doRequest(t, h, http.MethodPut, "/api/v1/stands/"+id, updateBody)
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

func TestUpdateStand_NotFound(t *testing.T) {
	h := newTestServer()

	rec := doRequest(t, h, http.MethodPut, "/api/v1/stands/00000000-0000-0000-0000-000000000000", validStandBody("Ghost"))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestDeleteStand(t *testing.T) {
	h := newTestServer()

	createRec := doRequest(t, h, http.MethodPost, "/api/v1/stands", validStandBody("Silver Chariot"))
	var created map[string]any
	_ = json.Unmarshal(createRec.Body.Bytes(), &created)
	id := created["id"].(string)

	rec := doRequest(t, h, http.MethodDelete, "/api/v1/stands/"+id, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusNoContent, rec.Body.String())
	}

	rec = doRequest(t, h, http.MethodGet, "/api/v1/stands/"+id, nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status after delete = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
