package endpoints_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/endpoints"
)

// fakeUserRepo is an in-memory ports.IUserRepository, this package's own
// copy of the fake following the repo's convention of duplicating small
// fakes per test file/package.
type fakeUserRepo struct {
	mu    sync.Mutex
	users map[user.UserID]*user.User
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{users: make(map[user.UserID]*user.User)}
}

func (f *fakeUserRepo) Save(_ context.Context, u *user.User) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.users[u.ID()] = u
	return nil
}

func (f *fakeUserRepo) FindByID(_ context.Context, id user.UserID) (*user.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.users[id]
	if !ok {
		return nil, ports.ErrUserNotFound
	}
	cp := *u
	return &cp, nil
}

func (f *fakeUserRepo) FindByGoogleSub(_ context.Context, sub string) (*user.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, u := range f.users {
		if u.GoogleSub() == sub {
			return u, nil
		}
	}
	return nil, ports.ErrUserNotFound
}

func (f *fakeUserRepo) FindByEmail(_ context.Context, email string) (*user.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, u := range f.users {
		if u.Email() == email {
			return u, nil
		}
	}
	return nil, ports.ErrUserNotFound
}

func (f *fakeUserRepo) FindByUsername(_ context.Context, username string) (*user.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for id, u := range f.users {
		if u.Username() == username {
			cp := *f.users[id]
			return &cp, nil
		}
	}
	return nil, ports.ErrUserNotFound
}

func (f *fakeUserRepo) UpdateUsername(_ context.Context, id user.UserID, username string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.users[id]
	if !ok {
		return ports.ErrUserNotFound
	}
	for otherID, other := range f.users {
		if otherID != id && other.Username() == username {
			return ports.ErrUserAlreadyExists
		}
	}
	return u.ChangeUsername(username)
}

func (f *fakeUserRepo) UpdateLanguage(_ context.Context, id user.UserID, language enums.Locale) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.users[id]
	if !ok {
		return ports.ErrUserNotFound
	}
	return u.ChangeLanguage(language)
}

func (f *fakeUserRepo) UpdateAvatar(_ context.Context, id user.UserID, main, thumb *string, status enums.PictureStatus) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.users[id]
	if !ok {
		return ports.ErrUserNotFound
	}
	newMain, newThumb := u.AvatarKey(), u.AvatarThumbKey()
	if main != nil {
		newMain = *main
	}
	if thumb != nil {
		newThumb = *thumb
	}
	u.SetAvatarRenditions(newMain, newThumb, status)
	return nil
}

func (f *fakeUserRepo) AvatarKeys(_ context.Context, id user.UserID) (string, string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.users[id]
	if !ok {
		return "", "", ports.ErrUserNotFound
	}
	return u.AvatarKey(), u.AvatarThumbKey(), nil
}

func (f *fakeUserRepo) UpdateRole(_ context.Context, id user.UserID, role enums.UserRole) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.users[id]
	if !ok {
		return ports.ErrUserNotFound
	}
	return u.ChangeRole(role)
}

func (f *fakeUserRepo) Delete(_ context.Context, id user.UserID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.users[id]; !ok {
		return ports.ErrUserNotFound
	}
	delete(f.users, id)
	return nil
}

func (f *fakeUserRepo) List(_ context.Context, limit, offset int32) ([]*user.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	all := make([]*user.User, 0, len(f.users))
	for _, u := range f.users {
		all = append(all, u)
	}
	start := int(offset)
	if start > len(all) {
		start = len(all)
	}
	end := start + int(limit)
	if end > len(all) {
		end = len(all)
	}
	return all[start:end], nil
}

func (f *fakeUserRepo) CountAdmins(_ context.Context) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var count int64
	for _, u := range f.users {
		if u.IsAdmin() {
			count++
		}
	}
	return count, nil
}

var _ ports.IUserRepository = (*fakeUserRepo)(nil)

// seedUser saves a ready-made user with userIDForToken's id for tok (only
// "user-token"/"user2-token"/"admin-token" are recognized by fakeTokenIssuer)
// so handler tests can address a caller whose id matches their bearer token.
func seedUser(t *testing.T, repo *fakeUserRepo, tok, username, email string, role enums.UserRole) *user.User {
	t.Helper()
	id, ok := userIDForToken[tok]
	if !ok {
		t.Fatalf("no id registered for token %q", tok)
	}
	u, err := user.NewUser(id, "google-sub-"+username, email, username, "Display Name", "https://google.example/pic.png", role)
	if err != nil {
		t.Fatalf("building user: %v", err)
	}
	if err := repo.Save(context.Background(), u); err != nil {
		t.Fatalf("saving user: %v", err)
	}
	return u
}

// newUserTestServer builds a router exposing every route (auth/stands/
// devil-fruits are bare, only /users is wired to a real UserService), so
// tests can exercise the full auth/authorization pipeline.
func newUserTestServer(repo *fakeUserRepo) http.Handler {
	pictures := newFakePictureStorage()
	processor := &fakeImageProcessor{}
	targets := map[enums.PictureSubjectKind]services.PictureTarget{
		enums.UserSubject: {Publisher: services.NewUserPicturePublisher(repo), KeyPrefix: "users"},
	}
	worker := services.NewPictureWorker(processor, pictures, targets, &fakeIDGenerator{}, services.WorkerConfig{
		Workers: 1, QueueSize: 1, JobTimeout: 5 * time.Second, MaxDimension: 1024, ThumbDimension: 256, Quality: 80,
	}, nil)
	svc := services.NewUserService(repo, pictures, processor, syncEnqueuer{worker},
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/webp", "image/avif", "image/jpeg", "image/png", "image/gif"}})
	userEndpoints := endpoints.NewUserEndpoints(svc)

	standEndpoints := endpoints.NewStandEndpoints(services.NewStandService(
		newFakeStandRepository(), &fakeIDGenerator{}, newFakePictureStorage(), &fakeImageProcessor{}, syncEnqueuer{},
		services.PicturePolicy{MaxBytes: 1 << 20, AllowedTypes: []string{"image/png"}}))
	authEndpoints := endpoints.NewAuthEndpoints(nil)
	eventsEndpoints := endpoints.NewEventsEndpoints(services.NewPictureEventHub(), fakeTokenIssuer{}, context.Background())

	gameEndpoints := endpoints.NewGameEndpoints(nil, services.NewGameEventHub(), nil, nil, nil, nil, fakeTokenIssuer{}, context.Background(), endpoints.GameWSConfig{})
	stageEndpoints := endpoints.NewStageEndpoints(nil)

	return endpoints.NewRouter(authEndpoints, standEndpoints, endpoints.NewDevilFruitEndpoints(nil), userEndpoints, eventsEndpoints, gameEndpoints, stageEndpoints,
		fakeTokenIssuer{}, endpoints.CORSConfig{}, endpoints.RateLimitConfig{}, endpoints.CacheConfig{})
}

// doRequestAs is like doRequest but with an explicit bearer token instead of
// the hardcoded "admin-token", so tests can address any of
// user-token/user2-token/admin-token.
func doRequestAs(t *testing.T, h http.Handler, method, path, token string, body any) *httptest.ResponseRecorder {
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
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestPatchUsersMe_UpdatesOnlyCallersOwnProfile(t *testing.T) {
	repo := newFakeUserRepo()
	seedUser(t, repo, "user-token", "old_name", "user1@example.com", enums.Regular)
	seedUser(t, repo, "user2-token", "someone_else", "user2@example.com", enums.Regular)
	h := newUserTestServer(repo)

	rec := doRequestAs(t, h, http.MethodPatch, "/api/v1/users/me", "user-token", map[string]any{"username": "new_name"})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &got)
	if got["username"] != "new_name" {
		t.Errorf("username = %v, want new_name", got["username"])
	}

	other, err := repo.FindByID(context.Background(), userIDForToken["user2-token"])
	if err != nil {
		t.Fatalf("FindByID user2: %v", err)
	}
	if other.Username() != "someone_else" {
		t.Errorf("user2's username changed unexpectedly: %q", other.Username())
	}
}

func TestPatchUsersMe_User2CannotTouchUser1(t *testing.T) {
	repo := newFakeUserRepo()
	seedUser(t, repo, "user-token", "user1_name", "user1@example.com", enums.Regular)
	seedUser(t, repo, "user2-token", "user2_name", "user2@example.com", enums.Regular)
	h := newUserTestServer(repo)

	rec := doRequestAs(t, h, http.MethodPatch, "/api/v1/users/me", "user2-token", map[string]any{"username": "hijacked"})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	user1, err := repo.FindByID(context.Background(), userIDForToken["user-token"])
	if err != nil {
		t.Fatalf("FindByID user1: %v", err)
	}
	if user1.Username() != "user1_name" {
		t.Errorf("user1's username was changed by user2's request: %q", user1.Username())
	}
	user2, err := repo.FindByID(context.Background(), userIDForToken["user2-token"])
	if err != nil {
		t.Fatalf("FindByID user2: %v", err)
	}
	if user2.Username() != "hijacked" {
		t.Errorf("user2's own username should have changed, got %q", user2.Username())
	}
}

func TestPatchUsersMe_UnknownFieldInBody_Returns400(t *testing.T) {
	repo := newFakeUserRepo()
	seedUser(t, repo, "user-token", "some_name", "user1@example.com", enums.Regular)
	h := newUserTestServer(repo)

	cases := []map[string]any{
		{"username": "still_valid", "email": "new@example.com"},
		{"username": "still_valid", "role": "ADMIN"},
		{"username": "still_valid", "completeName": "New Name"},
	}
	for _, body := range cases {
		rec := doRequestAs(t, h, http.MethodPatch, "/api/v1/users/me", "user-token", body)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("body %+v: status = %d, want %d, body = %s", body, rec.Code, http.StatusBadRequest, rec.Body.String())
		}
	}

	got, err := repo.FindByID(context.Background(), userIDForToken["user-token"])
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.Role() != enums.Regular || got.Email() != "user1@example.com" {
		t.Errorf("email/role must never change via PATCH /users/me, got email=%q role=%v", got.Email(), got.Role())
	}
}

func TestGetUsersByID_PublicProjection_OmitsEmailAndRole(t *testing.T) {
	repo := newFakeUserRepo()
	seedUser(t, repo, "user-token", "user1_name", "user1@example.com", enums.Admin)
	seedUser(t, repo, "user2-token", "user2_name", "user2@example.com", enums.Regular)
	h := newUserTestServer(repo)

	rec := doRequestAs(t, h, http.MethodGet, "/api/v1/users/"+userIDForToken["user-token"].String(), "user2-token", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := got["email"]; ok {
		t.Error("public profile response must not include email")
	}
	if _, ok := got["role"]; ok {
		t.Error("public profile response must not include role")
	}
	if got["username"] != "user1_name" {
		t.Errorf("username = %v, want user1_name", got["username"])
	}
}

func TestUsersAdminRoutes_RegularUser_Returns403(t *testing.T) {
	repo := newFakeUserRepo()
	seedUser(t, repo, "user-token", "user1_name", "user1@example.com", enums.Regular)
	seedUser(t, repo, "user2-token", "user2_name", "user2@example.com", enums.Regular)
	h := newUserTestServer(repo)

	target := userIDForToken["user2-token"].String()

	if rec := doRequestAs(t, h, http.MethodGet, "/api/v1/users", "user-token", nil); rec.Code != http.StatusForbidden {
		t.Errorf("GET /users: status = %d, want 403", rec.Code)
	}
	if rec := doRequestAs(t, h, http.MethodPatch, "/api/v1/users/"+target, "user-token", map[string]any{"username": "moderated"}); rec.Code != http.StatusForbidden {
		t.Errorf("PATCH /users/{id}: status = %d, want 403", rec.Code)
	}
	if rec := doRequestAs(t, h, http.MethodPatch, "/api/v1/users/"+target+"/role", "user-token", map[string]any{"role": "ADMIN"}); rec.Code != http.StatusForbidden {
		t.Errorf("PATCH /users/{id}/role: status = %d, want 403", rec.Code)
	}
	if rec := doRequestAs(t, h, http.MethodDelete, "/api/v1/users/"+target, "user-token", nil); rec.Code != http.StatusForbidden {
		t.Errorf("DELETE /users/{id}: status = %d, want 403", rec.Code)
	}
}

func TestDeleteUsersMe_RemovesOnlyCaller(t *testing.T) {
	repo := newFakeUserRepo()
	seedUser(t, repo, "user-token", "user1_name", "user1@example.com", enums.Regular)
	seedUser(t, repo, "user2-token", "user2_name", "user2@example.com", enums.Regular)
	h := newUserTestServer(repo)

	rec := doRequestAs(t, h, http.MethodDelete, "/api/v1/users/me", "user-token", nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusNoContent, rec.Body.String())
	}

	if _, err := repo.FindByID(context.Background(), userIDForToken["user-token"]); err == nil {
		t.Error("user1 should have been deleted")
	}
	if _, err := repo.FindByID(context.Background(), userIDForToken["user2-token"]); err != nil {
		t.Errorf("user2 should be untouched, got err: %v", err)
	}
}

func TestAdminUpdateRole_LastAdmin_Returns409(t *testing.T) {
	repo := newFakeUserRepo()
	seedUser(t, repo, "admin-token", "sole_admin", "admin@example.com", enums.Admin)
	seedUser(t, repo, "user-token", "user1_name", "user1@example.com", enums.Regular)
	h := newUserTestServer(repo)

	rec := doRequestAs(t, h, http.MethodPatch, "/api/v1/users/"+userIDForToken["admin-token"].String()+"/role", "admin-token", map[string]any{"role": "REGULAR"})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("self-demotion via admin route: status = %d, want 403, body = %s", rec.Code, rec.Body.String())
	}
}

func TestPatchUsersMePicture_QueuesTranscodeAndReturns202(t *testing.T) {
	repo := newFakeUserRepo()
	seedUser(t, repo, "user-token", "user1_name", "user1@example.com", enums.Regular)
	h := newUserTestServer(repo)

	rec := doMultipartRequest(t, h, http.MethodPatch, "/api/v1/users/me/picture", "picture", "avatar.png", pngBytes, "user-token")
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusAccepted, rec.Body.String())
	}

	getRec := doRequestAs(t, h, http.MethodGet, "/api/v1/users/me", "user-token", nil)
	var got map[string]any
	_ = json.Unmarshal(getRec.Body.Bytes(), &got)
	if got["avatarStatus"] != "READY" {
		t.Fatalf("avatarStatus after worker ran = %v, want READY", got["avatarStatus"])
	}
	if avatar, _ := got["avatar"].(string); avatar == "" {
		t.Error("avatar should be a presigned URL once READY")
	}
}

func TestDuplicateUsername_Returns409(t *testing.T) {
	repo := newFakeUserRepo()
	seedUser(t, repo, "user-token", "taken_name", "user1@example.com", enums.Regular)
	seedUser(t, repo, "user2-token", "user2_name", "user2@example.com", enums.Regular)
	h := newUserTestServer(repo)

	rec := doRequestAs(t, h, http.MethodPatch, "/api/v1/users/me", "user2-token", map[string]any{"username": "taken_name"})
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusConflict, rec.Body.String())
	}
}
