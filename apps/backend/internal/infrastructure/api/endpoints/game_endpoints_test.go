package endpoints_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/endpoints"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/gamestore"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/idgen"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/streamticket"
)

// --- local fakes mirroring game_service_test.go's (services_test package)
// fakeStageCatalog/fakeGamePowerPool/fakeAssignmentWeights/fakeTiebreaker -
// re-declared here since Go doesn't let a non-test package import
// test-only types across packages. fakeTokenIssuer/doRequest/userIDForToken
// are reused as-is from stand_endpoints_test.go (same endpoints_test
// package). ---

// fakeGameUserRepository is an in-memory ports.IUserRepository, just enough
// for GameEndpoints/GameService to resolve a caller's username/locale.
type fakeGameUserRepository struct {
	mu    sync.Mutex
	users map[user.UserID]*user.User
}

func newFakeGameUserRepository() *fakeGameUserRepository {
	return &fakeGameUserRepository{users: make(map[user.UserID]*user.User)}
}

func (f *fakeGameUserRepository) Save(_ context.Context, u *user.User) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.users[u.ID()] = u
	return nil
}

func (f *fakeGameUserRepository) FindByID(_ context.Context, id user.UserID) (*user.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.users[id]
	if !ok {
		return nil, ports.ErrUserNotFound
	}
	return u, nil
}

func (f *fakeGameUserRepository) FindByGoogleSub(_ context.Context, sub string) (*user.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, u := range f.users {
		if u.GoogleSub() == sub {
			return u, nil
		}
	}
	return nil, ports.ErrUserNotFound
}

func (f *fakeGameUserRepository) FindByEmail(_ context.Context, email string) (*user.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, u := range f.users {
		if u.Email() == email {
			return u, nil
		}
	}
	return nil, ports.ErrUserNotFound
}

func (f *fakeGameUserRepository) FindByUsername(_ context.Context, username string) (*user.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, u := range f.users {
		if u.Username() == username {
			return u, nil
		}
	}
	return nil, ports.ErrUserNotFound
}

func (f *fakeGameUserRepository) UpdateUsername(_ context.Context, id user.UserID, username string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.users[id]
	if !ok {
		return ports.ErrUserNotFound
	}
	return u.ChangeUsername(username)
}

func (f *fakeGameUserRepository) UpdateLanguage(_ context.Context, id user.UserID, language enums.Locale) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.users[id]
	if !ok {
		return ports.ErrUserNotFound
	}
	return u.ChangeLanguage(language)
}

func (f *fakeGameUserRepository) UpdateAvatar(_ context.Context, id user.UserID, main, thumb *string, status enums.PictureStatus) error {
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

func (f *fakeGameUserRepository) AvatarKeys(_ context.Context, id user.UserID) (string, string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.users[id]
	if !ok {
		return "", "", ports.ErrUserNotFound
	}
	return u.AvatarKey(), u.AvatarThumbKey(), nil
}

func (f *fakeGameUserRepository) UpdateRole(_ context.Context, id user.UserID, role enums.UserRole) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.users[id]
	if !ok {
		return ports.ErrUserNotFound
	}
	return u.ChangeRole(role)
}

func (f *fakeGameUserRepository) Delete(_ context.Context, id user.UserID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.users[id]; !ok {
		return ports.ErrUserNotFound
	}
	delete(f.users, id)
	return nil
}

func (f *fakeGameUserRepository) List(_ context.Context, limit, offset int32) ([]*user.User, error) {
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

func (f *fakeGameUserRepository) CountAdmins(_ context.Context) (int64, error) {
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

var _ ports.IUserRepository = (*fakeGameUserRepository)(nil)

// fakeGameStageRepository is a minimal in-memory ports.IStageRepository -
// GameEndpoints only ever calls Translations (via stageTextResolver); the
// rest of the interface is unused by these tests but must still be
// satisfied to build a *GameEndpoints.
type fakeGameStageRepository struct{}

func (fakeGameStageRepository) List(context.Context, enums.Locale) ([]game.Stage, error) {
	return nil, nil
}

func (fakeGameStageRepository) Filter(context.Context, ports.StageFilters, enums.Locale) ([]game.Stage, error) {
	return nil, nil
}

func (fakeGameStageRepository) FindByID(context.Context, game.StageID, enums.Locale) (game.Stage, error) {
	return game.Stage{}, ports.ErrStageNotFound
}

func (fakeGameStageRepository) Save(context.Context, game.Stage, ports.StageTranslations) error {
	return nil
}

func (fakeGameStageRepository) Delete(context.Context, game.StageID) error { return nil }

func (fakeGameStageRepository) Translations(context.Context, game.StageID) (ports.StageTranslations, error) {
	return ports.StageTranslations{}, nil
}

func (fakeGameStageRepository) UpdatePicture(context.Context, game.StageID, *string, *string, enums.PictureStatus) error {
	return nil
}

var _ ports.IStageRepository = fakeGameStageRepository{}

// fakeGameStandRepository/fakeGameDevilFruitRepository are minimal
// in-memory ports.IStandRepository/IDevilFruitRepository - GameEndpoints
// only ever calls Translations (via standTextResolver/devilFruitTextResolver);
// the rest of each interface is unused by these tests but must still be
// satisfied to build a *GameEndpoints. translations is nil-safe: an id with
// no entry resolves to ports.PowerTranslations{}, i.e. every locale falls
// back all the way to "" (fine for tests not exercising a specific locale).
type fakeGameStandRepository struct {
	translations map[powers.PowerID]ports.PowerTranslations
}

func (fakeGameStandRepository) FindByID(context.Context, powers.PowerID, enums.Locale) (*powers.Stand, error) {
	return nil, ports.ErrStandNotFound
}

func (fakeGameStandRepository) FindByName(context.Context, string, enums.Locale) (*powers.Stand, error) {
	return nil, ports.ErrStandNotFound
}

func (fakeGameStandRepository) GetAll(context.Context, enums.Locale) ([]*powers.Stand, error) {
	return nil, nil
}

func (fakeGameStandRepository) Filter(context.Context, ports.StandFilters, enums.Locale) ([]*powers.Stand, error) {
	return nil, nil
}

func (fakeGameStandRepository) Save(context.Context, *powers.Stand, ports.PowerTranslations) error {
	return nil
}

func (fakeGameStandRepository) Delete(context.Context, powers.PowerID) error { return nil }

func (fakeGameStandRepository) UpdatePicture(context.Context, powers.PowerID, *string, *string, enums.PictureStatus) error {
	return nil
}

func (f fakeGameStandRepository) Translations(_ context.Context, id powers.PowerID) (ports.PowerTranslations, error) {
	return f.translations[id], nil
}

var _ ports.IStandRepository = fakeGameStandRepository{}

type fakeGameDevilFruitRepository struct {
	translations map[powers.PowerID]ports.PowerTranslations
}

func (fakeGameDevilFruitRepository) FindByID(context.Context, powers.PowerID, enums.Locale) (*powers.DevilFruit, error) {
	return nil, ports.ErrDevilFruitNotFound
}

func (fakeGameDevilFruitRepository) FindByName(context.Context, string, enums.Locale) (*powers.DevilFruit, error) {
	return nil, ports.ErrDevilFruitNotFound
}

func (fakeGameDevilFruitRepository) GetAll(context.Context, enums.Locale) ([]*powers.DevilFruit, error) {
	return nil, nil
}

func (fakeGameDevilFruitRepository) Filter(context.Context, ports.DevilFruitFilters, enums.Locale) ([]*powers.DevilFruit, error) {
	return nil, nil
}

func (fakeGameDevilFruitRepository) Save(context.Context, *powers.DevilFruit, ports.PowerTranslations) error {
	return nil
}

func (fakeGameDevilFruitRepository) Delete(context.Context, powers.PowerID) error { return nil }

func (fakeGameDevilFruitRepository) UpdatePicture(context.Context, powers.PowerID, *string, *string, enums.PictureStatus) error {
	return nil
}

func (f fakeGameDevilFruitRepository) Translations(_ context.Context, id powers.PowerID) (ports.PowerTranslations, error) {
	return f.translations[id], nil
}

var _ ports.IDevilFruitRepository = fakeGameDevilFruitRepository{}

// fakeGameStageCatalog is ports.IStageCatalog, feeding GameService's own
// stage selection at create/reconfigure time - mirrors game_service_test.go's
// fakeStageCatalog.
type fakeGameStageCatalog struct {
	stages map[enums.Manga][]game.Stage
}

func (f *fakeGameStageCatalog) Stages(_ context.Context, m enums.Manga) ([]game.Stage, error) {
	return append([]game.Stage(nil), f.stages[m]...), nil
}

var _ ports.IStageCatalog = (*fakeGameStageCatalog)(nil)

// fakeGamePowerPool mirrors game_service_test.go's fakeGamePowerPool.
type fakeGamePowerPool struct {
	stands []*powers.Stand
	fruits []*powers.DevilFruit
}

func (f *fakeGamePowerPool) Stands(context.Context) ([]*powers.Stand, error) {
	return append([]*powers.Stand(nil), f.stands...), nil
}

func (f *fakeGamePowerPool) DevilFruits(context.Context) ([]*powers.DevilFruit, error) {
	return append([]*powers.DevilFruit(nil), f.fruits...), nil
}

var _ ports.IGamePowerPool = (*fakeGamePowerPool)(nil)

// fakeGameAssignmentWeights mirrors game_service_test.go's fakeAssignmentWeights.
type fakeGameAssignmentWeights struct{ w game.AssignmentWeights }

func (f fakeGameAssignmentWeights) Load(context.Context) (game.AssignmentWeights, error) {
	return f.w, nil
}

var _ ports.IAssignmentWeights = fakeGameAssignmentWeights{}

// fakeGameTiebreaker mirrors game_service_test.go's fakeTiebreaker, always
// picking the first option - no test in this file drives a real tie.
type fakeGameTiebreaker struct{}

func (fakeGameTiebreaker) Break(_ context.Context, options []string) (string, error) {
	if len(options) == 0 {
		return "", nil
	}
	return options[0], nil
}

var _ ports.ITiebreaker = fakeGameTiebreaker{}

// fakeGameRandom is a deterministic game.RandomSource - always 0, fine
// since no test here exercises randomized tie-breaking or team-shuffling.
type fakeGameRandom struct{}

func (fakeGameRandom) IntN(n int) int {
	if n <= 0 {
		return 0
	}
	return 0
}

var _ game.RandomSource = fakeGameRandom{}

// --- fixtures ---

var endpointStageIDCounter byte

func mustEndpointStage(t *testing.T, manga enums.Manga, order int, name string) game.Stage {
	t.Helper()
	endpointStageIDCounter++
	var id game.StageID
	id[15] = endpointStageIDCounter
	s, err := game.NewStage(id, manga, order, name, "a test stage", "")
	if err != nil {
		t.Fatalf("mustEndpointStage: %v", err)
	}
	return s
}

var endpointPowerIDCounter byte

func mustEndpointStand(t *testing.T, name string) *powers.Stand {
	t.Helper()
	endpointPowerIDCounter++
	var id powers.PowerID
	id[15] = endpointPowerIDCounter
	skills := []string{"skill"}
	power, err := powers.NewPower(id, name, "description", enums.Common, &skills, "")
	if err != nil {
		t.Fatalf("mustEndpointStand power: %v", err)
	}
	stand, err := powers.NewStand(*power, enums.B, enums.B, enums.B, enums.B, enums.B, enums.B, nil)
	if err != nil {
		t.Fatalf("mustEndpointStand: %v", err)
	}
	return stand
}

func mustEndpointDevilFruit(t *testing.T, name string) *powers.DevilFruit {
	t.Helper()
	endpointPowerIDCounter++
	var id powers.PowerID
	id[15] = endpointPowerIDCounter
	skills := []string{"skill"}
	power, err := powers.NewPower(id, name, "description", enums.Common, &skills, "")
	if err != nil {
		t.Fatalf("mustEndpointDevilFruit power: %v", err)
	}
	fruit, err := powers.NewDevilFruit(*power, enums.Paramecia)
	if err != nil {
		t.Fatalf("mustEndpointDevilFruit: %v", err)
	}
	return fruit
}

// --- test wiring ---

// gameEndpointsTestDeps exposes the pieces individual tests need: the user
// repo (to seed callers) and the live GameService (to drive host-only
// actions - like SET_LOCK - that have no REST route of their own, so a test
// can still exercise what an HTTP call *does* see afterward).
type gameEndpointsTestDeps struct {
	users   *fakeGameUserRepository
	svc     *services.GameService
	tickets *streamticket.MemoryStore
}

// newGameTestServer builds a full router with a real GameService (real
// idgen.UUIDGenerator, real gamestore.NewMemoryGameStore) wired to local
// fakes for everything else, per game-lobby-todo.md's §2 instructions.
// fakeGameStandRepository/fakeGameDevilFruitRepository carry no translations
// here - GameEndpoints' per-viewer Stand/DevilFruit text resolution
// (standTextResolver/devilFruitTextResolver) is covered directly against
// dto.NewGameStateResponse in dto/game_response_test.go instead of through
// this HTTP harness, since the harness's weighted-random draw (see
// LoadoutBuilder.drawStand) can't be steered to a specific Stand without a
// dedicated RandomSource fake.
func newGameTestServer(t *testing.T) (http.Handler, *gameEndpointsTestDeps) {
	t.Helper()

	users := newFakeGameUserRepository()
	stageCatalog := &fakeGameStageCatalog{stages: map[enums.Manga][]game.Stage{
		enums.Jojo:     {mustEndpointStage(t, enums.Jojo, 0, "Phantom Blood"), mustEndpointStage(t, enums.Jojo, 1, "Battle Tendency")},
		enums.OnePiece: {mustEndpointStage(t, enums.OnePiece, 0, "East Blue"), mustEndpointStage(t, enums.OnePiece, 1, "Alabasta")},
	}}
	powerPool := &fakeGamePowerPool{
		stands: []*powers.Stand{mustEndpointStand(t, "Star Platinum"), mustEndpointStand(t, "Crazy Diamond")},
		fruits: []*powers.DevilFruit{mustEndpointDevilFruit(t, "Gomu Gomu no Mi"), mustEndpointDevilFruit(t, "Mera Mera no Mi")},
	}

	svc := services.NewGameService(
		gamestore.NewMemoryGameStore(),
		idgen.UUIDGenerator[game.GameID]{},
		idgen.UUIDGenerator[game.ParticipantID]{},
		idgen.UUIDGenerator[game.TeamID]{},
		users,
		stageCatalog,
		powerPool,
		fakeGameAssignmentWeights{w: game.DefaultAssignmentWeights()},
		fakeGameTiebreaker{},
		nil,
		fakeGameRandom{},
		services.NewGameEventHub(),
		services.NewSystemClock(),
		services.VotingPolicy{Window: 30_000_000_000},
	)

	tickets := streamticket.NewMemoryStore(streamticket.Config{TTL: 30 * time.Second})
	gameEndpoints := endpoints.NewGameEndpoints(svc, services.NewGameEventHub(), fakeGameStageRepository{},
		fakeGameStandRepository{}, fakeGameDevilFruitRepository{},
		users, fakeTokenIssuer{}, tickets, context.Background(), endpoints.GameWSConfig{})
	authEndpoints := endpoints.NewAuthEndpoints(nil)
	eventsEndpoints := endpoints.NewEventsEndpoints(services.NewPictureEventHub(), fakeTokenIssuer{}, tickets, context.Background())
	h := endpoints.NewRouter(authEndpoints, endpoints.NewStandEndpoints(nil), endpoints.NewDevilFruitEndpoints(nil), endpoints.NewUserEndpoints(nil), eventsEndpoints, gameEndpoints, endpoints.NewStageEndpoints(nil), fakeTokenIssuer{}, endpoints.CORSConfig{}, endpoints.RateLimitConfig{}, endpoints.CacheConfig{})

	return h, &gameEndpointsTestDeps{users: users, svc: svc, tickets: tickets}
}

// mustGameUser creates and saves a user whose UserID matches one of
// fakeTokenIssuer's fixed tokens (userIDForToken, stand_endpoints_test.go),
// so the returned token authenticates as this user for doTokenRequest.
func mustGameUser(t *testing.T, deps *gameEndpointsTestDeps, token, username string) user.UserID {
	t.Helper()
	id := userIDForToken[token]
	u, err := user.NewUser(id, "google-sub-"+username, username+"@example.com", username, username, "", enums.Regular)
	if err != nil {
		t.Fatalf("mustGameUser: %v", err)
	}
	if err := deps.users.Save(context.Background(), u); err != nil {
		t.Fatalf("mustGameUser save: %v", err)
	}
	return id
}

// doTokenRequest is an alias for doRequestAs (user_endpoints_test.go, same
// package) - kept as its own name here so this file reads independently of
// where that helper happens to live.
func doTokenRequest(t *testing.T, h http.Handler, method, path, token string, body any) *httptest.ResponseRecorder {
	return doRequestAs(t, h, method, path, token, body)
}

func decodeGameBody(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal response: %v, body = %s", err, raw)
	}
	return out
}

// gauntletCreateBody sets an explicit visibility even though POST /games
// treats it as optional (defaulting to PRIVATE) - PATCH .../config's
// UpdateConfigPayload.Validate requires it always, and the same body is
// reused for both requests in the config-edit tests below.
func gauntletCreateBody() map[string]any {
	return map[string]any{
		"mode":          "GAUNTLET",
		"stageMangas":   []string{"JOJO"},
		"powerMangas":   []string{"JOJO"},
		"abilitySource": "RANDOM",
		"teamSize":      5,
		"allowBots":     false,
		"visibility":    "PRIVATE",
	}
}

func publicVersusCreateBody(teamSize int) map[string]any {
	return map[string]any{
		"mode":          "VERSUS",
		"stageMangas":   []string{"JOJO", "ONE_PIECE"},
		"powerMangas":   []string{"JOJO", "ONE_PIECE"},
		"abilitySource": "RANDOM",
		"teamSize":      teamSize,
		"allowBots":     true,
		"visibility":    "PUBLIC",
	}
}

func privateVersusCreateBody(teamSize int) map[string]any {
	body := publicVersusCreateBody(teamSize)
	body["visibility"] = "PRIVATE"
	return body
}

// createGameViaAPI POSTs body as hostToken and returns the new game's id and
// join code.
func createGameViaAPI(t *testing.T, h http.Handler, hostToken string, body map[string]any) (id, code string) {
	t.Helper()
	rec := doTokenRequest(t, h, http.MethodPost, "/api/v1/games", hostToken, body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: status = %d, want %d, body = %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	created := decodeGameBody(t, rec.Body.Bytes())
	gameObj, _ := created["game"].(map[string]any)
	id, _ = gameObj["id"].(string)
	code, _ = gameObj["code"].(string)
	if id == "" || code == "" {
		t.Fatalf("create response missing id/code: %v", created)
	}
	return id, code
}

// --- GET /games/public ---

func TestListPublicLobbies_Success(t *testing.T) {
	h, deps := newGameTestServer(t)
	mustGameUser(t, deps, "user-token", "host")
	mustGameUser(t, deps, "user2-token", "viewer")

	createGameViaAPI(t, h, "user-token", publicVersusCreateBody(2))

	rec := doTokenRequest(t, h, http.MethodGet, "/api/v1/games/public", "user2-token", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	body := rec.Body.String()
	if strings.Contains(body, `"code"`) {
		t.Errorf("body must not contain a join code: %s", body)
	}
	if strings.Contains(body, `"participants"`) {
		t.Errorf("body must not contain a roster: %s", body)
	}

	got := decodeGameBody(t, rec.Body.Bytes())
	items, ok := got["items"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("items = %v, want exactly 1", got["items"])
	}
}

func TestListPublicLobbies_RequiresAuth(t *testing.T) {
	h, _ := newGameTestServer(t)
	rec := doTokenRequest(t, h, http.MethodGet, "/api/v1/games/public", "", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

// --- GET /games/preview ---

func TestPreviewByCode_NonParticipant_Success(t *testing.T) {
	h, deps := newGameTestServer(t)
	mustGameUser(t, deps, "user-token", "host")
	mustGameUser(t, deps, "user2-token", "viewer")

	_, code := createGameViaAPI(t, h, "user-token", gauntletCreateBody())

	rec := doTokenRequest(t, h, http.MethodGet, "/api/v1/games/preview?code="+code, "user2-token", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	got := decodeGameBody(t, rec.Body.Bytes())
	if got["code"] != code {
		t.Errorf("preview code = %v, want %v", got["code"], code)
	}
}

func TestPreviewByCode_UnknownCode_NotFound(t *testing.T) {
	h, deps := newGameTestServer(t)
	mustGameUser(t, deps, "user-token", "viewer")

	rec := doTokenRequest(t, h, http.MethodGet, "/api/v1/games/preview?code=NOPE12", "user-token", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestPreviewByCode_WorksForPrivateLobby(t *testing.T) {
	h, deps := newGameTestServer(t)
	mustGameUser(t, deps, "user-token", "host")
	mustGameUser(t, deps, "user2-token", "viewer")

	_, code := createGameViaAPI(t, h, "user-token", privateVersusCreateBody(2))

	rec := doTokenRequest(t, h, http.MethodGet, "/api/v1/games/preview?code="+code, "user2-token", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	got := decodeGameBody(t, rec.Body.Bytes())
	if got["visibility"] != "PRIVATE" {
		t.Errorf("visibility = %v, want PRIVATE", got["visibility"])
	}
}

// --- POST /games/{id}/join ---

func TestJoinByID_PrivateLobby_Forbidden(t *testing.T) {
	h, deps := newGameTestServer(t)
	mustGameUser(t, deps, "user-token", "host")
	mustGameUser(t, deps, "user2-token", "joiner")

	id, _ := createGameViaAPI(t, h, "user-token", privateVersusCreateBody(2))

	rec := doTokenRequest(t, h, http.MethodPost, "/api/v1/games/"+id+"/join", "user2-token", nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
	got := decodeGameBody(t, rec.Body.Bytes())
	if got["code"] != "LOBBY_PRIVATE" {
		t.Errorf("code = %v, want LOBBY_PRIVATE", got["code"])
	}
}

func TestJoinByID_LockedLobby_Conflict(t *testing.T) {
	h, deps := newGameTestServer(t)
	hostUserID := mustGameUser(t, deps, "user-token", "host")
	mustGameUser(t, deps, "user2-token", "joiner")

	id, _ := createGameViaAPI(t, h, "user-token", publicVersusCreateBody(2))

	// SET_LOCK has no REST route (WS-only - see game_ws_endpoints_test.go
	// for direct dispatch coverage of that command), so this test locks the
	// lobby through the live GameService directly, then exercises the HTTP
	// join path against that locked state.
	gid, err := game.ParseGameID(id)
	if err != nil {
		t.Fatalf("ParseGameID: %v", err)
	}
	g, err := deps.svc.GetGame(context.Background(), gid)
	if err != nil {
		t.Fatalf("GetGame: %v", err)
	}
	hostParticipant, err := resolveEndpointParticipant(g, hostUserID)
	if err != nil {
		t.Fatalf("resolve host participant: %v", err)
	}
	if _, err := deps.svc.SetLobbyLocked(context.Background(), gid, hostParticipant, true); err != nil {
		t.Fatalf("SetLobbyLocked: %v", err)
	}

	rec := doTokenRequest(t, h, http.MethodPost, "/api/v1/games/"+id+"/join", "user2-token", nil)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusConflict, rec.Body.String())
	}
	got := decodeGameBody(t, rec.Body.Bytes())
	if got["code"] != "LOBBY_LOCKED" {
		t.Errorf("code = %v, want LOBBY_LOCKED", got["code"])
	}
}

// resolveEndpointParticipant mirrors game_endpoints.go's own
// resolveParticipant (unexported, so it can't be reused directly from this
// package).
func resolveEndpointParticipant(g *game.Game, userID user.UserID) (game.ParticipantID, error) {
	for _, p := range g.Participants() {
		if uid := p.UserID(); uid != nil && *uid == userID {
			return p.ID(), nil
		}
	}
	return game.NilParticipantID, ports.ErrForbidden
}

// --- PATCH /games/{id}/config ---

func TestEditConfig_HostOnly_Success(t *testing.T) {
	h, deps := newGameTestServer(t)
	mustGameUser(t, deps, "user-token", "host")

	id, _ := createGameViaAPI(t, h, "user-token", gauntletCreateBody())

	newBody := gauntletCreateBody()
	newBody["teamSize"] = 8
	rec := doTokenRequest(t, h, http.MethodPatch, "/api/v1/games/"+id+"/config", "user-token", newBody)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	got := decodeGameBody(t, rec.Body.Bytes())
	gameObj, _ := got["game"].(map[string]any)
	cfg, _ := gameObj["config"].(map[string]any)
	if teamSize, _ := cfg["teamSize"].(float64); int(teamSize) != 8 {
		t.Errorf("teamSize = %v, want 8", cfg["teamSize"])
	}
}

func TestEditConfig_NotHost_Forbidden(t *testing.T) {
	h, deps := newGameTestServer(t)
	mustGameUser(t, deps, "user-token", "host")
	mustGameUser(t, deps, "user2-token", "joiner")

	id, code := createGameViaAPI(t, h, "user-token", gauntletCreateBody())

	joinRec := doTokenRequest(t, h, http.MethodPost, "/api/v1/games/join", "user2-token", map[string]any{"code": code})
	if joinRec.Code != http.StatusOK {
		t.Fatalf("join: status = %d, body = %s", joinRec.Code, joinRec.Body.String())
	}

	newBody := gauntletCreateBody()
	newBody["teamSize"] = 8
	rec := doTokenRequest(t, h, http.MethodPatch, "/api/v1/games/"+id+"/config", "user2-token", newBody)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
	got := decodeGameBody(t, rec.Body.Bytes())
	if got["code"] != "NOT_HOST" {
		t.Errorf("code = %v, want NOT_HOST", got["code"])
	}
}

func TestEditConfig_NonParticipant_Forbidden(t *testing.T) {
	h, deps := newGameTestServer(t)
	mustGameUser(t, deps, "user-token", "host")
	mustGameUser(t, deps, "user2-token", "stranger")

	id, _ := createGameViaAPI(t, h, "user-token", gauntletCreateBody())

	newBody := gauntletCreateBody()
	rec := doTokenRequest(t, h, http.MethodPatch, "/api/v1/games/"+id+"/config", "user2-token", newBody)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

// --- POST /games/{id}/ws-ticket ---

func mintWSTicket(t *testing.T, h http.Handler, id, token string) (ticket string, status int) {
	t.Helper()
	rec := doTokenRequest(t, h, http.MethodPost, "/api/v1/games/"+id+"/ws-ticket", token, nil)
	if rec.Code != http.StatusOK {
		return "", rec.Code
	}
	var body struct {
		Ticket    string    `json:"ticket"`
		ExpiresAt time.Time `json:"expiresAt"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decoding ws-ticket response: %v, body = %s", err, rec.Body.String())
	}
	if body.Ticket == "" {
		t.Fatalf("ws-ticket response has an empty ticket: %s", rec.Body.String())
	}
	return body.Ticket, rec.Code
}

func TestMintWSTicket_Unauthenticated(t *testing.T) {
	h, deps := newGameTestServer(t)
	mustGameUser(t, deps, "user-token", "host")
	id, _ := createGameViaAPI(t, h, "user-token", gauntletCreateBody())

	_, status := mintWSTicket(t, h, id, "")
	if status != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", status)
	}
}

func TestMintWSTicket_NonParticipant_Forbidden(t *testing.T) {
	h, deps := newGameTestServer(t)
	mustGameUser(t, deps, "user-token", "host")
	mustGameUser(t, deps, "user2-token", "stranger")
	id, _ := createGameViaAPI(t, h, "user-token", gauntletCreateBody())

	_, status := mintWSTicket(t, h, id, "user2-token")
	if status != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", status)
	}
}

func TestMintWSTicket_UnknownGame_NotFound(t *testing.T) {
	h, deps := newGameTestServer(t)
	mustGameUser(t, deps, "user-token", "host")

	_, status := mintWSTicket(t, h, "00000000-0000-0000-0000-000000000000", "user-token")
	if status != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", status)
	}
}

func TestMintWSTicket_MalformedID_BadRequest(t *testing.T) {
	h, deps := newGameTestServer(t)
	mustGameUser(t, deps, "user-token", "host")

	_, status := mintWSTicket(t, h, "not-a-uuid", "user-token")
	if status != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", status)
	}
}

func TestMintWSTicket_SeatedParticipant_ReturnsTicket(t *testing.T) {
	h, deps := newGameTestServer(t)
	mustGameUser(t, deps, "user-token", "host")
	id, _ := createGameViaAPI(t, h, "user-token", gauntletCreateBody())

	ticket, status := mintWSTicket(t, h, id, "user-token")
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200", status)
	}
	if ticket == "" {
		t.Fatal("expected a non-empty ticket")
	}
}

// --- GET /games/{id}/ws, dialed for real over an httptest.Server ---

// dialGameWS dials the game socket at wsURL, expecting either a successful
// upgrade (wantErr false, in which case the caller is handed the live conn
// and must close it) or a failure (wantErr true).
func dialGameWS(t *testing.T, wsURL string, wantErr bool) *websocket.Conn {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if wantErr {
		if err == nil {
			conn.CloseNow()
			t.Fatal("Dial succeeded, want an error")
		}
		return nil
	}
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	return conn
}

func TestServeWS_ValidTicket_UpgradesAndSendsState(t *testing.T) {
	h, deps := newGameTestServer(t)
	mustGameUser(t, deps, "user-token", "host")
	id, _ := createGameViaAPI(t, h, "user-token", gauntletCreateBody())
	ticket, status := mintWSTicket(t, h, id, "user-token")
	if status != http.StatusOK {
		t.Fatalf("mint status = %d, want 200", status)
	}

	srv := httptest.NewServer(h)
	defer srv.Close()
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/api/v1/games/" + id + "/ws?ticket=" + ticket

	conn := dialGameWS(t, wsURL, false)
	defer conn.CloseNow()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, data, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	var frame struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &frame); err != nil {
		t.Fatalf("unmarshal frame: %v, data = %s", err, data)
	}
	if frame.Type != "STATE" {
		t.Fatalf("first frame type = %q, want STATE", frame.Type)
	}
}

func TestServeWS_ReusedTicket_Fails(t *testing.T) {
	h, deps := newGameTestServer(t)
	mustGameUser(t, deps, "user-token", "host")
	id, _ := createGameViaAPI(t, h, "user-token", gauntletCreateBody())
	ticket, _ := mintWSTicket(t, h, id, "user-token")

	srv := httptest.NewServer(h)
	defer srv.Close()
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/api/v1/games/" + id + "/ws?ticket=" + ticket

	conn := dialGameWS(t, wsURL, false)
	conn.CloseNow()

	dialGameWS(t, wsURL, true)
}

func TestServeWS_TicketForAnotherGame_Fails(t *testing.T) {
	h, deps := newGameTestServer(t)
	mustGameUser(t, deps, "user-token", "host")
	idA, _ := createGameViaAPI(t, h, "user-token", gauntletCreateBody())

	ticketForA, status := mintWSTicket(t, h, idA, "user-token")
	if status != http.StatusOK {
		t.Fatalf("mint status = %d, want 200", status)
	}

	// idB doesn't need to exist as a real game: authenticateStream's
	// resource check (ticketForA's Resource is idA) happens before serveWS
	// ever calls GetGame, so a well-formed but different UUID is enough to
	// prove the ticket is scoped to the game it was minted for. A second
	// real CreateGame here would also collide with fakeGameRandom's
	// always-0 draw, which makes every join code in this file identical.
	idB := "00000000-0000-0000-0000-000000000001"
	srv := httptest.NewServer(h)
	defer srv.Close()
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/api/v1/games/" + idB + "/ws?ticket=" + ticketForA

	dialGameWS(t, wsURL, true)
}

func TestServeWS_EventsTicket_Fails(t *testing.T) {
	h, deps := newGameTestServer(t)
	mustGameUser(t, deps, "user-token", "host")
	id, _ := createGameViaAPI(t, h, "user-token", gauntletCreateBody())

	// Minted directly against the shared store so it carries the events
	// purpose instead of game-ws.
	token, _, err := deps.tickets.Issue(context.Background(), ports.StreamTicket{Purpose: ports.TicketPurposeEvents})
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	srv := httptest.NewServer(h)
	defer srv.Close()
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/api/v1/games/" + id + "/ws?ticket=" + token

	dialGameWS(t, wsURL, true)
}

// TestServeWS_OldTokenQueryParam_Fails is the regression guard for the
// removed ?token=<jwt> fallback on the game socket, mirroring
// TestStream_OldTokenQueryParam_Fails for /events.
func TestServeWS_OldTokenQueryParam_Fails(t *testing.T) {
	h, deps := newGameTestServer(t)
	mustGameUser(t, deps, "user-token", "host")
	id, _ := createGameViaAPI(t, h, "user-token", gauntletCreateBody())

	srv := httptest.NewServer(h)
	defer srv.Close()
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/api/v1/games/" + id + "/ws?token=user-token"

	dialGameWS(t, wsURL, true)
}

func TestServeWS_BearerHeader_StillUpgrades(t *testing.T) {
	h, deps := newGameTestServer(t)
	mustGameUser(t, deps, "user-token", "host")
	id, _ := createGameViaAPI(t, h, "user-token", gauntletCreateBody())

	srv := httptest.NewServer(h)
	defer srv.Close()
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/api/v1/games/" + id + "/ws"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: http.Header{"Authorization": []string{"Bearer user-token"}},
	})
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	conn.CloseNow()
}
