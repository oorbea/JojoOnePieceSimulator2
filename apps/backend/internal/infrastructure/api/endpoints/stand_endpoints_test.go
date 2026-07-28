package endpoints_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
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

// fakeTokenIssuer recognizes two fixed tokens ("user-token"/"admin-token")
// instead of doing real JWT parsing, so endpoint tests can exercise
// RequireAuth/RequireAdmin without any signing/verification machinery.
type fakeTokenIssuer struct{}

func (fakeTokenIssuer) Issue(_ *user.User) (string, time.Time, error) {
	return "", time.Time{}, errors.New("not implemented")
}

// userIDForToken gives each fake token a distinct, deterministic UserID so
// per-user rate-limit tests can tell two callers' buckets apart.
var userIDForToken = map[string]user.UserID{
	"user-token":  {1},
	"admin-token": {2},
	"user2-token": {3},
}

func (fakeTokenIssuer) Parse(token string) (ports.Claims, error) {
	switch token {
	case "user-token", "user2-token":
		return ports.Claims{UserID: userIDForToken[token], Role: enums.Regular}, nil
	case "admin-token":
		return ports.Claims{UserID: userIDForToken[token], Role: enums.Admin}, nil
	default:
		return ports.Claims{}, ports.ErrUnauthenticated
	}
}

var _ ports.ITokenIssuer = fakeTokenIssuer{}

// fakePictureStorage is an in-memory ports.IPictureStorage, so endpoint
// tests can exercise PATCH .../picture without a real R2 bucket.
type fakePictureStorage struct {
	mu         sync.Mutex
	objects    map[string][]byte
	deleted    []string
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
	content, err := io.ReadAll(pic.Content)
	if err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.objects[key] = content
	return nil
}

func (f *fakePictureStorage) PresignGetURL(_ context.Context, key string) (string, error) {
	if f.presignErr != nil {
		return "", f.presignErr
	}
	return fmt.Sprintf("https://r2.test/%s?sig=fake", key), nil
}

func (f *fakePictureStorage) Delete(_ context.Context, key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.deleted = append(f.deleted, key)
	delete(f.objects, key)
	return nil
}

func (f *fakePictureStorage) has(key string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	_, ok := f.objects[key]
	return ok
}

func (f *fakePictureStorage) wasDeleted(key string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, k := range f.deleted {
		if k == key {
			return true
		}
	}
	return false
}

var _ ports.IPictureStorage = (*fakePictureStorage)(nil)

// newTestServer builds a router with rate limiting disabled, so existing
// (non-rate-limit) tests keep exercising exactly what they exercised before
// RateLimitConfig existed.
func newTestServer() http.Handler {
	return newTestServerWithRateLimit(endpoints.RateLimitConfig{})
}

func newTestServerWithRateLimit(rateCfg endpoints.RateLimitConfig) http.Handler {
	return newTestServerWithDeps(rateCfg, newFakePictureStorage())
}

// newTestServerWithDeps is like newTestServerWithRateLimit but also lets
// tests inject their own fakePictureStorage, so PATCH .../picture tests can
// inspect what was uploaded/deleted.
func newTestServerWithDeps(rateCfg endpoints.RateLimitConfig, pictures *fakePictureStorage) http.Handler {
	repo := newFakeStandRepository()
	svc := services.NewStandService(repo, &fakeIDGenerator{}, pictures,
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/jpeg", "image/png", "image/webp"}})
	standEndpoints := endpoints.NewStandEndpoints(svc)
	authEndpoints := endpoints.NewAuthEndpoints(nil)
	return endpoints.NewRouter(authEndpoints, standEndpoints, fakeTokenIssuer{}, endpoints.CORSConfig{}, rateCfg)
}

func validStandBody(name string) map[string]any {
	return map[string]any{
		"name":        name,
		"description": name + " description",
		"rarity":      "RARE",
		"skills":      []string{"punch", "dash"},
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
	req.Header.Set("Authorization", "Bearer admin-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// doMultipartRequest builds and sends a multipart/form-data request with a
// single file field, for PATCH .../picture tests - doRequest hardcodes a
// JSON content type so it can't be reused here.
func doMultipartRequest(t *testing.T, h http.Handler, method, path, fieldName, filename string, content []byte, token string) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile(fieldName, filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// pngBytes is a minimal valid 1x1 PNG, long enough for http.DetectContentType
// to sniff it as image/png.
var pngBytes = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
	0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4, 0x89, 0x00, 0x00, 0x00,
	0x0d, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9c, 0x62, 0x00, 0x01, 0x00, 0x00,
	0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00, 0x00, 0x00, 0x00, 0x49,
	0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
}

func TestPatchStandPicture(t *testing.T) {
	pictures := newFakePictureStorage()
	h := newTestServerWithDeps(endpoints.RateLimitConfig{}, pictures)

	createRec := doRequest(t, h, http.MethodPost, "/api/v1/stands", validStandBody("Crazy Diamond"))
	var created map[string]any
	_ = json.Unmarshal(createRec.Body.Bytes(), &created)
	id := created["id"].(string)
	if created["picture"] != "" {
		t.Errorf("picture on create = %v, want empty", created["picture"])
	}

	rec := doMultipartRequest(t, h, http.MethodPatch, "/api/v1/stands/"+id+"/picture", "picture", "stand.png", pngBytes, "admin-token")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	firstURL, ok := got["picture"].(string)
	if !ok || firstURL == "" {
		t.Fatalf("picture = %v, want a presigned URL", got["picture"])
	}

	// GET returns the same (freshly resolved) URL shape.
	getRec := doRequest(t, h, http.MethodGet, "/api/v1/stands/"+id, nil)
	var getGot map[string]any
	_ = json.Unmarshal(getRec.Body.Bytes(), &getGot)
	if getGot["picture"] != firstURL {
		t.Errorf("GET picture = %v, want %v", getGot["picture"], firstURL)
	}

	// Uploading a second picture replaces the first and deletes its key.
	rec = doMultipartRequest(t, h, http.MethodPatch, "/api/v1/stands/"+id+"/picture", "picture", "stand2.png", pngBytes, "admin-token")
	if rec.Code != http.StatusOK {
		t.Fatalf("second patch: status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &got)
	secondURL, _ := got["picture"].(string)
	if secondURL == firstURL {
		t.Error("second picture URL should differ from the first (new key)")
	}

	firstKey := urlToKey(firstURL)
	secondKey := urlToKey(secondURL)
	if !pictures.has(secondKey) {
		t.Error("second picture's key not found in storage")
	}
	if !pictures.wasDeleted(firstKey) {
		t.Error("first picture's key should have been deleted after replacement")
	}
}

// urlToKey inverts fakePictureStorage.PresignGetURL's "https://r2.test/<key>?sig=fake" shape.
func urlToKey(url string) string {
	url = strings.TrimPrefix(url, "https://r2.test/")
	return strings.TrimSuffix(url, "?sig=fake")
}

func TestPatchStandPicture_MissingField(t *testing.T) {
	h := newTestServer()
	createRec := doRequest(t, h, http.MethodPost, "/api/v1/stands", validStandBody("Star Platinum"))
	var created map[string]any
	_ = json.Unmarshal(createRec.Body.Bytes(), &created)
	id := created["id"].(string)

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	_ = w.Close()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/stands/"+id+"/picture", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", "Bearer admin-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestPatchStandPicture_UnsupportedType(t *testing.T) {
	h := newTestServer()
	createRec := doRequest(t, h, http.MethodPost, "/api/v1/stands", validStandBody("The World"))
	var created map[string]any
	_ = json.Unmarshal(createRec.Body.Bytes(), &created)
	id := created["id"].(string)

	rec := doMultipartRequest(t, h, http.MethodPatch, "/api/v1/stands/"+id+"/picture", "picture", "notes.txt", []byte("plain text content"), "admin-token")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestPatchStandPicture_TooLarge(t *testing.T) {
	pictures := newFakePictureStorage()
	// MaxBytes: 1 forces any real upload over the limit.
	repo := newFakeStandRepository()
	svc := services.NewStandService(repo, &fakeIDGenerator{}, pictures, services.PicturePolicy{MaxBytes: 1, AllowedTypes: []string{"image/png"}})
	standEndpoints := endpoints.NewStandEndpoints(svc)
	authEndpoints := endpoints.NewAuthEndpoints(nil)
	h := endpoints.NewRouter(authEndpoints, standEndpoints, fakeTokenIssuer{}, endpoints.CORSConfig{}, endpoints.RateLimitConfig{})

	createRec := doRequest(t, h, http.MethodPost, "/api/v1/stands", validStandBody("Gold Experience"))
	var created map[string]any
	_ = json.Unmarshal(createRec.Body.Bytes(), &created)
	id := created["id"].(string)

	rec := doMultipartRequest(t, h, http.MethodPatch, "/api/v1/stands/"+id+"/picture", "picture", "stand.png", pngBytes, "admin-token")
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusRequestEntityTooLarge, rec.Body.String())
	}
}

func TestPatchStandPicture_RequiresAdmin(t *testing.T) {
	h := newTestServer()
	createRec := doRequest(t, h, http.MethodPost, "/api/v1/stands", validStandBody("Killer Queen"))
	var created map[string]any
	_ = json.Unmarshal(createRec.Body.Bytes(), &created)
	id := created["id"].(string)

	rec := doMultipartRequest(t, h, http.MethodPatch, "/api/v1/stands/"+id+"/picture", "picture", "stand.png", pngBytes, "user-token")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestPatchStandPicture_NotFound(t *testing.T) {
	h := newTestServer()

	rec := doMultipartRequest(t, h, http.MethodPatch, "/api/v1/stands/00000000-0000-0000-0000-000000000000/picture", "picture", "stand.png", pngBytes, "admin-token")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestPatchStandPicture_InvalidID(t *testing.T) {
	h := newTestServer()

	rec := doMultipartRequest(t, h, http.MethodPatch, "/api/v1/stands/not-a-uuid/picture", "picture", "stand.png", pngBytes, "admin-token")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
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

// noAuthRequest is like doRequest but never attaches an Authorization
// header, for exercising RequireAuth's rejection path.
func noAuthRequest(t *testing.T, h http.Handler, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBuffer(nil))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// userAuthRequest is like doRequest but authenticates as a regular (non
// admin) user, for exercising RequireAdmin's rejection path.
func userAuthRequest(t *testing.T, h http.Handler, method, path string, body any) *httptest.ResponseRecorder {
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
	req.Header.Set("Authorization", "Bearer user-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestStandRoutes_RequireAuth(t *testing.T) {
	h := newTestServer()

	rec := noAuthRequest(t, h, http.MethodGet, "/api/v1/stands")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("GET / without token: status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	rec = noAuthRequest(t, h, http.MethodPost, "/api/v1/stands")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("POST / without token: status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestStandRoutes_RegularUserCanRead(t *testing.T) {
	h := newTestServer()
	doRequest(t, h, http.MethodPost, "/api/v1/stands", validStandBody("Silver Chariot"))

	rec := userAuthRequest(t, h, http.MethodGet, "/api/v1/stands", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestStandRoutes_RegularUserCannotWrite(t *testing.T) {
	h := newTestServer()

	rec := userAuthRequest(t, h, http.MethodPost, "/api/v1/stands", validStandBody("Star Platinum"))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("POST: status = %d, want %d, body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}

	createRec := doRequest(t, h, http.MethodPost, "/api/v1/stands", validStandBody("The World"))
	var created map[string]any
	_ = json.Unmarshal(createRec.Body.Bytes(), &created)
	id := created["id"].(string)

	rec = userAuthRequest(t, h, http.MethodPut, "/api/v1/stands/"+id, validStandBody("The World"))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("PUT: status = %d, want %d, body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}

	rec = userAuthRequest(t, h, http.MethodDelete, "/api/v1/stands/"+id, nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("DELETE: status = %d, want %d, body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}
