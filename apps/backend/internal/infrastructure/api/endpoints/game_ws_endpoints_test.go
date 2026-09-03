// This file deliberately lives in package endpoints (white-box), not
// endpoints_test - dispatch and forwardEvents are unexported methods of
// *GameEndpoints, only reachable this way. Every other endpoints_test.go
// file in this package tests through the public HTTP surface instead; this
// is the one exception, because §2 of game-lobby-todo.md specifically asks
// for direct dispatch/forwardEvents coverage of the WS command surface.
package endpoints

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/dto"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/gamestore"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/idgen"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/streamticket"
)

// --- local fakes for GameService's own dependencies, mirroring
// game_service_test.go's (services_test package) fakeStageCatalog/
// fakeGamePowerPool/fakeAssignmentWeights/fakeTiebreaker/fakeUserRepository -
// re-declared here (package endpoints, not services_test) since Go doesn't
// let a test file import another package's test-only types. Named with a
// wsFake* prefix to avoid clashing with game_endpoints_test.go's own
// fakeGame*-named types in the sibling endpoints_test package (different
// package, so no real collision risk, but kept distinct for clarity). ---

type wsFakeUserRepository struct {
	users map[user.UserID]*user.User
}

func newWSFakeUserRepository() *wsFakeUserRepository {
	return &wsFakeUserRepository{users: make(map[user.UserID]*user.User)}
}

func (f *wsFakeUserRepository) Save(_ context.Context, u *user.User) error {
	f.users[u.ID()] = u
	return nil
}

func (f *wsFakeUserRepository) FindByID(_ context.Context, id user.UserID) (*user.User, error) {
	u, ok := f.users[id]
	if !ok {
		return nil, ports.ErrUserNotFound
	}
	return u, nil
}

func (f *wsFakeUserRepository) FindByGoogleSub(context.Context, string) (*user.User, error) {
	return nil, ports.ErrUserNotFound
}

func (f *wsFakeUserRepository) FindByEmail(context.Context, string) (*user.User, error) {
	return nil, ports.ErrUserNotFound
}

func (f *wsFakeUserRepository) FindByUsername(context.Context, string) (*user.User, error) {
	return nil, ports.ErrUserNotFound
}

func (f *wsFakeUserRepository) UpdateUsername(context.Context, user.UserID, string) error {
	return nil
}

func (f *wsFakeUserRepository) UpdateLanguage(context.Context, user.UserID, enums.Locale) error {
	return nil
}

func (f *wsFakeUserRepository) UpdateAvatar(context.Context, user.UserID, *string, *string, enums.PictureStatus) error {
	return nil
}

func (f *wsFakeUserRepository) AvatarKeys(context.Context, user.UserID) (string, string, error) {
	return "", "", nil
}

func (f *wsFakeUserRepository) UpdateRole(context.Context, user.UserID, enums.UserRole) error {
	return nil
}

func (f *wsFakeUserRepository) Delete(context.Context, user.UserID) error { return nil }

func (f *wsFakeUserRepository) List(context.Context, int32, int32) ([]*user.User, error) {
	return nil, nil
}

func (f *wsFakeUserRepository) CountAdmins(context.Context) (int64, error) { return 0, nil }

var _ ports.IUserRepository = (*wsFakeUserRepository)(nil)

// wsFakeStageCatalog mirrors game_service_test.go's fakeStageCatalog.
type wsFakeStageCatalog struct {
	stages map[enums.Manga][]game.Stage
}

func (f *wsFakeStageCatalog) Stages(_ context.Context, m enums.Manga) ([]game.Stage, error) {
	return append([]game.Stage(nil), f.stages[m]...), nil
}

var _ ports.IStageCatalog = (*wsFakeStageCatalog)(nil)

// wsFakeGamePowerPool mirrors game_service_test.go's fakeGamePowerPool.
type wsFakeGamePowerPool struct {
	stands []*powers.Stand
	fruits []*powers.DevilFruit
}

func (f *wsFakeGamePowerPool) Stands(context.Context) ([]*powers.Stand, error) {
	return append([]*powers.Stand(nil), f.stands...), nil
}

func (f *wsFakeGamePowerPool) DevilFruits(context.Context) ([]*powers.DevilFruit, error) {
	return append([]*powers.DevilFruit(nil), f.fruits...), nil
}

var _ ports.IGamePowerPool = (*wsFakeGamePowerPool)(nil)

// wsFakeAssignmentWeights mirrors game_service_test.go's fakeAssignmentWeights.
type wsFakeAssignmentWeights struct{ w game.AssignmentWeights }

func (f wsFakeAssignmentWeights) Load(context.Context) (game.AssignmentWeights, error) {
	return f.w, nil
}

var _ ports.IAssignmentWeights = wsFakeAssignmentWeights{}

// wsFakeTiebreaker mirrors game_service_test.go's fakeTiebreaker - no test
// here drives an actual tie, so it's never called for real.
type wsFakeTiebreaker struct{}

func (wsFakeTiebreaker) Break(_ context.Context, options []string) (string, error) {
	if len(options) == 0 {
		return "", nil
	}
	return options[0], nil
}

var _ ports.ITiebreaker = wsFakeTiebreaker{}

// wsFakeRandom is a deterministic game.RandomSource - always 0.
type wsFakeRandom struct{}

func (wsFakeRandom) IntN(n int) int {
	if n <= 0 {
		return 0
	}
	return 0
}

var _ game.RandomSource = wsFakeRandom{}

var wsStageIDCounter byte

func mustWSStage(t *testing.T, manga enums.Manga, order int, name string) game.Stage {
	t.Helper()
	wsStageIDCounter++
	var id game.StageID
	id[15] = wsStageIDCounter
	s, err := game.NewStage(id, manga, order, name, "a test stage", "")
	if err != nil {
		t.Fatalf("mustWSStage: %v", err)
	}
	return s
}

// wsLobby bundles everything a dispatch test needs: the GameEndpoints under
// test, the live game's id, the seated host/joiner ParticipantIDs, and both
// Versus TeamIDs.
type wsLobby struct {
	endpoints *GameEndpoints
	gameID    game.GameID
	hostID    game.ParticipantID
	joinerID  game.ParticipantID
	teamA     game.TeamID
	teamB     game.TeamID
}

// newWSLobby builds a fresh Versus lobby (host + one joiner, TeamSize 2,
// bots allowed) with a real GameService (real idgen.UUIDGenerator, real
// gamestore.NewMemoryGameStore) backing a *GameEndpoints, so each dispatch
// table case starts from identical, isolated state. stages/users/issuer
// passed to NewGameEndpoints are nil - dispatch never touches them (only
// e.svc), and forwardEvents' kicked-self path returns before it would need
// them either.
func newWSLobby(t *testing.T) *wsLobby {
	t.Helper()

	users := newWSFakeUserRepository()
	stageCatalog := &wsFakeStageCatalog{stages: map[enums.Manga][]game.Stage{
		enums.Jojo:     {mustWSStage(t, enums.Jojo, 0, "Phantom Blood")},
		enums.OnePiece: {mustWSStage(t, enums.OnePiece, 0, "East Blue")},
	}}
	powerPool := &wsFakeGamePowerPool{
		stands: []*powers.Stand{mustEndpointStandLike(t, "Star Platinum")},
		fruits: []*powers.DevilFruit{mustEndpointFruitLike(t, "Gomu Gomu no Mi")},
	}

	svc := services.NewGameService(
		gamestore.NewMemoryGameStore(),
		idgen.UUIDGenerator[game.GameID]{},
		idgen.UUIDGenerator[game.ParticipantID]{},
		idgen.UUIDGenerator[game.TeamID]{},
		users,
		stageCatalog,
		powerPool,
		wsFakeAssignmentWeights{w: game.DefaultAssignmentWeights()},
		wsFakeTiebreaker{},
		nil,
		wsFakeRandom{},
		services.NewGameEventHub(),
		services.NewSystemClock(),
		services.VotingPolicy{Window: 30 * time.Second},
	)

	var hostUserID, joinerUserID user.UserID
	hostUserID[15] = 1
	joinerUserID[15] = 2
	hostUser, err := user.NewUser(hostUserID, "google-sub-host", "host@example.com", "host", "host", "", enums.Regular)
	if err != nil {
		t.Fatalf("host user: %v", err)
	}
	joinerUser, err := user.NewUser(joinerUserID, "google-sub-joiner", "joiner@example.com", "joiner", "joiner", "", enums.Regular)
	if err != nil {
		t.Fatalf("joiner user: %v", err)
	}
	if err := users.Save(context.Background(), hostUser); err != nil {
		t.Fatalf("save host: %v", err)
	}
	if err := users.Save(context.Background(), joinerUser); err != nil {
		t.Fatalf("save joiner: %v", err)
	}

	g, code, err := svc.CreateGame(context.Background(), hostUserID, services.CreateGameInput{
		Mode: enums.Versus, StageMangas: []enums.Manga{enums.Jojo}, PowerMangas: []enums.Manga{enums.Jojo}, AbilitySource: enums.Random,
		TeamSize: 2, AllowBots: true,
	})
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	g, err = svc.JoinByCode(context.Background(), code, joinerUserID)
	if err != nil {
		t.Fatalf("JoinByCode: %v", err)
	}

	var joinerParticipant game.ParticipantID
	for _, p := range g.Participants() {
		if p.ID() != g.HostID() {
			joinerParticipant = p.ID()
		}
	}

	e := NewGameEndpoints(svc, services.NewGameEventHub(), nil, nil, nil, nil, nil,
		streamticket.NewMemoryStore(streamticket.Config{TTL: 30 * time.Second}), context.Background(), GameWSConfig{})

	return &wsLobby{
		endpoints: e,
		gameID:    g.ID(),
		hostID:    g.HostID(),
		joinerID:  joinerParticipant,
		teamA:     g.Teams()[0].ID(),
		teamB:     g.Teams()[1].ID(),
	}
}

var wsPowerIDCounter byte

func mustEndpointStandLike(t *testing.T, name string) *powers.Stand {
	t.Helper()
	wsPowerIDCounter++
	var id powers.PowerID
	id[15] = wsPowerIDCounter
	skills := []string{"skill"}
	power, err := powers.NewPower(id, name, "description", enums.Common, &skills, "")
	if err != nil {
		t.Fatalf("mustEndpointStandLike power: %v", err)
	}
	stand, err := powers.NewStand(*power, enums.B, enums.B, enums.B, enums.B, enums.B, enums.B, nil)
	if err != nil {
		t.Fatalf("mustEndpointStandLike: %v", err)
	}
	return stand
}

func mustEndpointFruitLike(t *testing.T, name string) *powers.DevilFruit {
	t.Helper()
	wsPowerIDCounter++
	var id powers.PowerID
	id[15] = wsPowerIDCounter
	skills := []string{"skill"}
	power, err := powers.NewPower(id, name, "description", enums.Common, &skills, "")
	if err != nil {
		t.Fatalf("mustEndpointFruitLike power: %v", err)
	}
	fruit, err := powers.NewDevilFruit(*power, enums.Paramecia)
	if err != nil {
		t.Fatalf("mustEndpointFruitLike: %v", err)
	}
	return fruit
}

func mustMarshal(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return b
}

// --- dispatch table tests ---

func TestDispatch_SwitchTeam(t *testing.T) {
	lobby := newWSLobby(t)
	g, err := lobby.endpoints.svc.GetGame(context.Background(), lobby.gameID)
	if err != nil {
		t.Fatalf("GetGame: %v", err)
	}
	joiner, ok := g.Participant(lobby.joinerID)
	if !ok {
		t.Fatalf("joiner not seated")
	}
	target := lobby.teamA
	if joiner.TeamID() == lobby.teamA {
		target = lobby.teamB
	}

	cmd := dto.ClientCommand{Type: dto.CommandSwitchTeam, Payload: mustMarshal(t, dto.SwitchTeamPayload{TeamID: target.String()})}
	got, err := lobby.endpoints.dispatch(context.Background(), lobby.gameID, lobby.joinerID, cmd)
	if err != nil {
		t.Fatalf("dispatch SWITCH_TEAM: %v", err)
	}
	moved, ok := got.Participant(lobby.joinerID)
	if !ok {
		t.Fatalf("joiner missing after switch")
	}
	if moved.TeamID() != target {
		t.Errorf("teamID = %v, want %v", moved.TeamID(), target)
	}
}

func TestDispatch_MovePlayer(t *testing.T) {
	lobby := newWSLobby(t)
	g, err := lobby.endpoints.svc.GetGame(context.Background(), lobby.gameID)
	if err != nil {
		t.Fatalf("GetGame: %v", err)
	}
	joiner, ok := g.Participant(lobby.joinerID)
	if !ok {
		t.Fatalf("joiner not seated")
	}
	target := lobby.teamA
	if joiner.TeamID() == lobby.teamA {
		target = lobby.teamB
	}

	cmd := dto.ClientCommand{Type: dto.CommandMovePlayer, Payload: mustMarshal(t, dto.MovePlayerPayload{
		ParticipantID: lobby.joinerID.String(), TeamID: target.String(),
	})}
	got, err := lobby.endpoints.dispatch(context.Background(), lobby.gameID, lobby.hostID, cmd)
	if err != nil {
		t.Fatalf("dispatch MOVE_PLAYER: %v", err)
	}
	moved, ok := got.Participant(lobby.joinerID)
	if !ok {
		t.Fatalf("joiner missing after move")
	}
	if moved.TeamID() != target {
		t.Errorf("teamID = %v, want %v", moved.TeamID(), target)
	}
}

func TestDispatch_MovePlayer_NotHost_Forbidden(t *testing.T) {
	lobby := newWSLobby(t)

	cmd := dto.ClientCommand{Type: dto.CommandMovePlayer, Payload: mustMarshal(t, dto.MovePlayerPayload{
		ParticipantID: lobby.hostID.String(), TeamID: lobby.teamA.String(),
	})}
	// Moving the host, dispatched by the joiner (not the host, not moving
	// themselves) - g.SwitchTeam must reject this as game.ErrNotHost.
	_, err := lobby.endpoints.dispatch(context.Background(), lobby.gameID, lobby.joinerID, cmd)
	if err == nil {
		t.Fatal("dispatch MOVE_PLAYER by non-host targeting someone else: want error, got nil")
	}
}

func TestDispatch_Kick(t *testing.T) {
	lobby := newWSLobby(t)

	cmd := dto.ClientCommand{Type: dto.CommandKick, Payload: mustMarshal(t, dto.KickPayload{ParticipantID: lobby.joinerID.String()})}
	got, err := lobby.endpoints.dispatch(context.Background(), lobby.gameID, lobby.hostID, cmd)
	if err != nil {
		t.Fatalf("dispatch KICK: %v", err)
	}
	if _, ok := got.Participant(lobby.joinerID); ok {
		t.Error("joiner still seated after being kicked")
	}
}

func TestDispatch_Kick_NotHost_Forbidden(t *testing.T) {
	lobby := newWSLobby(t)

	cmd := dto.ClientCommand{Type: dto.CommandKick, Payload: mustMarshal(t, dto.KickPayload{ParticipantID: lobby.hostID.String()})}
	_, err := lobby.endpoints.dispatch(context.Background(), lobby.gameID, lobby.joinerID, cmd)
	if err == nil {
		t.Fatal("dispatch KICK by non-host: want error, got nil")
	}
}

func TestDispatch_TransferHost(t *testing.T) {
	lobby := newWSLobby(t)

	cmd := dto.ClientCommand{Type: dto.CommandTransferHost, Payload: mustMarshal(t, dto.TransferHostPayload{ParticipantID: lobby.joinerID.String()})}
	got, err := lobby.endpoints.dispatch(context.Background(), lobby.gameID, lobby.hostID, cmd)
	if err != nil {
		t.Fatalf("dispatch TRANSFER_HOST: %v", err)
	}
	if got.HostID() != lobby.joinerID {
		t.Errorf("hostID = %v, want %v", got.HostID(), lobby.joinerID)
	}
}

func TestDispatch_SetLock(t *testing.T) {
	lobby := newWSLobby(t)

	cmd := dto.ClientCommand{Type: dto.CommandSetLock, Payload: mustMarshal(t, dto.SetLockPayload{Locked: true})}
	got, err := lobby.endpoints.dispatch(context.Background(), lobby.gameID, lobby.hostID, cmd)
	if err != nil {
		t.Fatalf("dispatch SET_LOCK: %v", err)
	}
	if !got.Locked() {
		t.Error("Locked() = false, want true")
	}

	cmd = dto.ClientCommand{Type: dto.CommandSetLock, Payload: mustMarshal(t, dto.SetLockPayload{Locked: false})}
	got, err = lobby.endpoints.dispatch(context.Background(), lobby.gameID, lobby.hostID, cmd)
	if err != nil {
		t.Fatalf("dispatch SET_LOCK (unlock): %v", err)
	}
	if got.Locked() {
		t.Error("Locked() = true after unlock, want false")
	}
}

func TestDispatch_SetLock_NotHost_Forbidden(t *testing.T) {
	lobby := newWSLobby(t)

	cmd := dto.ClientCommand{Type: dto.CommandSetLock, Payload: mustMarshal(t, dto.SetLockPayload{Locked: true})}
	_, err := lobby.endpoints.dispatch(context.Background(), lobby.gameID, lobby.joinerID, cmd)
	if err == nil {
		t.Fatal("dispatch SET_LOCK by non-host: want error, got nil")
	}
}

func TestDispatch_UpdateConfig(t *testing.T) {
	lobby := newWSLobby(t)

	payload := dto.UpdateConfigPayload{
		Mode: "VERSUS", StageMangas: []string{"JOJO"}, PowerMangas: []string{"JOJO"}, AbilitySource: "RANDOM",
		TeamSize: 3, AllowBots: true, Visibility: "PUBLIC",
	}
	cmd := dto.ClientCommand{Type: dto.CommandUpdateConfig, Payload: mustMarshal(t, payload)}
	got, err := lobby.endpoints.dispatch(context.Background(), lobby.gameID, lobby.hostID, cmd)
	if err != nil {
		t.Fatalf("dispatch UPDATE_CONFIG: %v", err)
	}
	if got.Config().TeamSize() != 3 {
		t.Errorf("teamSize = %d, want 3", got.Config().TeamSize())
	}
	if got.Config().Visibility() != enums.Public {
		t.Errorf("visibility = %v, want PUBLIC", got.Config().Visibility())
	}
}

func TestDispatch_UpdateConfig_NotHost_Forbidden(t *testing.T) {
	lobby := newWSLobby(t)

	payload := dto.UpdateConfigPayload{
		Mode: "VERSUS", StageMangas: []string{"JOJO"}, PowerMangas: []string{"JOJO"}, AbilitySource: "RANDOM",
		TeamSize: 3, AllowBots: true, Visibility: "PUBLIC",
	}
	cmd := dto.ClientCommand{Type: dto.CommandUpdateConfig, Payload: mustMarshal(t, payload)}
	_, err := lobby.endpoints.dispatch(context.Background(), lobby.gameID, lobby.joinerID, cmd)
	if err == nil {
		t.Fatal("dispatch UPDATE_CONFIG by non-host: want error, got nil")
	}
}

func TestDispatch_UnknownCommand(t *testing.T) {
	lobby := newWSLobby(t)

	cmd := dto.ClientCommand{Type: "NOT_A_REAL_COMMAND"}
	_, err := lobby.endpoints.dispatch(context.Background(), lobby.gameID, lobby.hostID, cmd)
	if err != errUnknownCommand {
		t.Errorf("err = %v, want errUnknownCommand", err)
	}
}

// --- forwardEvents: kicked participant's own socket closes ---

// TestForwardEvents_KickedParticipant_ClosesOwnSocket exercises the one
// branch of forwardEvents that isn't reachable via dispatch alone: once a
// PlayerKicked event names self as the victim, forwardEvents must close
// that connection with StatusNormalClosure and return, instead of the usual
// resend-STATE-and-keep-going path every other event takes.
func TestForwardEvents_KickedParticipant_ClosesOwnSocket(t *testing.T) {
	e := NewGameEndpoints(nil, nil, nil, nil, nil, nil, nil, nil, context.Background(), GameWSConfig{})

	var self game.ParticipantID
	self[15] = 7
	events := make(chan services.GameEvent, 1)
	done := make(chan struct{})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		outbound := make(chan []byte, wsOutboundBuffer)
		e.forwardEvents(r.Context(), conn, outbound, game.GameID{}, self, events)
		close(done)
	}))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	clientConn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer clientConn.CloseNow()

	events <- services.GameEvent{Event: game.PlayerKicked{ParticipantID: self}}

	// forwardEvents only enqueues the PLAYER_KICKED frame onto outbound
	// (this test never runs the write pump that would actually flush it to
	// the wire); it closes the connection directly via conn.Close, so the
	// very next thing the client observes is the close itself.
	_, _, err = clientConn.Read(ctx)
	if err == nil {
		t.Fatal("Read: want the connection closed, got nil error")
	}
	if got := websocket.CloseStatus(err); got != websocket.StatusNormalClosure {
		t.Errorf("close status = %v, want %v (err = %v)", got, websocket.StatusNormalClosure, err)
	}

	select {
	case <-done:
	case <-ctx.Done():
		t.Fatal("forwardEvents did not return after closing the kicked participant's socket")
	}
}

// TestBuildEventFrame_TerminalFrames_ResendStateAndCarryParticipants pins
// the result-screen contract. Both terminal frames used to skip the STATE
// resend, on the (then-true) reasoning that finalizeLocked had already
// deleted the game, and on the claim that the frame alone carried everything
// a terminal screen needs - it never did. A finished game now survives for a
// short TTL, so the resend is both safe and necessary.
func TestBuildEventFrame_TerminalFrames_ResendStateAndCarryParticipants(t *testing.T) {
	pid := game.ParticipantID{3}
	teamID := game.TeamID{7}
	result := game.GameResult{
		Winner: "SURVIVE", RoundsPlayed: 4, Mode: enums.Gauntlet,
		Participants: []game.ParticipantOutcome{
			{ParticipantID: pid, DisplayName: "jotaro", TeamID: teamID, Bot: false},
			{ParticipantID: game.ParticipantID{4}, DisplayName: "bot", TeamID: teamID, Bot: true},
		},
	}

	frameType, payload, resendState := buildEventFrame(game.GameFinished{Result: result}, 30*time.Second, 0, 0, time.Time{})
	if frameType != dto.FrameGameFinished {
		t.Fatalf("frameType = %q, want %q", frameType, dto.FrameGameFinished)
	}
	if !resendState {
		t.Fatal("GAME_FINISHED resendState = false, want true (the result screen needs the full STATE)")
	}
	got, ok := payload.(dto.GameFinishedPayload)
	if !ok {
		t.Fatalf("payload type = %T, want dto.GameFinishedPayload", payload)
	}
	if len(got.Result.Participants) != 2 {
		t.Fatalf("participants = %+v, want both seats mapped onto the wire", got.Result.Participants)
	}
	first := got.Result.Participants[0]
	if first.ParticipantID != pid.String() || first.DisplayName != "jotaro" ||
		first.TeamID != teamID.String() || first.Bot {
		t.Fatalf("participant[0] = %+v, want the human seat mapped faithfully", first)
	}
	if !got.Result.Participants[1].Bot {
		t.Fatalf("participant[1] = %+v, want it flagged as a bot", got.Result.Participants[1])
	}

	_, _, abortedResend := buildEventFrame(game.GameAborted{Reason: "host left"}, 30*time.Second, 0, 0, time.Time{})
	if !abortedResend {
		t.Fatal("GAME_ABORTED resendState = false, want true (aborted games land on the result screen too)")
	}
}

// TestBuildEventFrame_TimedFrames_UseStampedClosesAt pins the B1 contract:
// all five timed frames render the authoritative deadline the service
// stamped on the event, never a freshly synthesized time.Now()+window.
func TestBuildEventFrame_TimedFrames_UseStampedClosesAt(t *testing.T) {
	stamped := time.Now().Add(97 * time.Second).Truncate(time.Second)
	window := 30 * time.Second

	cases := []struct {
		evt game.DomainEvent
		closesAt  func(payload any) string
		name      string
		wantFrame string
		// fallbackWindow is the window buildEventFrame's zero-closesAt
		// fallback actually uses for this case - window for the four
		// frames driven by the votingWindow/revealWindow/summaryWindow
		// args (all passed as `window` below), but game.ResultDuration
		// for ROUND_RESOLVED, which frameDeadline hardcodes rather than
		// reading from the envelope.
		fallbackWindow time.Duration
	}{
		{
			name: "VOTING_OPENED", evt: game.VotingOpened{RoundIndex: 1}, wantFrame: dto.FrameVotingOpened,
			closesAt: func(p any) string { return p.(dto.VotingOpenedPayload).ClosesAt }, fallbackWindow: window,
		},
		{
			name: "TIEBREAK_OPENED", evt: game.TiebreakOpened{RoundIndex: 1}, wantFrame: dto.FrameTiebreakOpened,
			closesAt: func(p any) string { return p.(dto.TiebreakOpenedPayload).ClosesAt }, fallbackWindow: window,
		},
		{
			name: "SUMMARY_OPENED", evt: game.SummaryOpened{RoundIndex: 1}, wantFrame: dto.FrameSummaryOpened,
			closesAt: func(p any) string { return p.(dto.SummaryOpenedPayload).ClosesAt }, fallbackWindow: window,
		},
		{
			name: "LOADOUTS_ASSIGNED", evt: game.LoadoutsAssigned{RoundIndex: 1}, wantFrame: dto.FrameLoadoutsAssigned,
			closesAt: func(p any) string { return p.(dto.LoadoutsAssignedPayload).ClosesAt }, fallbackWindow: window,
		},
		{
			name: "ROUND_RESOLVED", evt: game.RoundResolved{RoundIndex: 1, Winner: "SURVIVE"}, wantFrame: dto.FrameRoundResolved,
			closesAt:       func(p any) string { return p.(dto.RoundResolvedPayload).ClosesAt },
			fallbackWindow: game.ResultDuration,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			frameType, payload, _ := buildEventFrame(tc.evt, window, window, window, stamped)
			if frameType != tc.wantFrame {
				t.Fatalf("frameType = %q, want %q", frameType, tc.wantFrame)
			}
			if got, want := tc.closesAt(payload), stamped.Format(time.RFC3339); got != want {
				t.Fatalf("closesAt = %q, want the stamped deadline %q", got, want)
			}
		})

		t.Run(tc.name+"_zero_falls_back", func(t *testing.T) {
			before := time.Now()
			_, payload, _ := buildEventFrame(tc.evt, window, window, window, time.Time{})
			got, err := time.Parse(time.RFC3339, tc.closesAt(payload))
			if err != nil {
				t.Fatalf("parsing fallback closesAt: %v", err)
			}
			if got.Before(before.Add(tc.fallbackWindow - 2*time.Second)) {
				t.Fatalf("fallback closesAt = %v, want roughly now+%v", got, tc.fallbackWindow)
			}
		})
	}
}

func TestBuildEventFrame_VoteCast_CarriesHumanVoteProgress(t *testing.T) {
	pid := game.ParticipantID{7}
	evt := game.VoteCast{
		RoundIndex: 2, ParticipantID: pid, Option: "SURVIVE",
		HumanVotesCast: 1, HumanVoters: 3,
	}
	frameType, payload, resendState := buildEventFrame(evt, 30*time.Second, 0, 0, time.Time{})

	if frameType != dto.FrameVoteCast {
		t.Fatalf("frameType = %q, want %q", frameType, dto.FrameVoteCast)
	}
	if resendState {
		t.Fatal("resendState = true, want false (VOTE_CAST never triggers a resend)")
	}
	got, ok := payload.(dto.VoteCastPayload)
	if !ok {
		t.Fatalf("payload type = %T, want dto.VoteCastPayload", payload)
	}
	want := dto.VoteCastPayload{RoundIndex: 2, VotesCast: 1, Voters: 3}
	if got != want {
		t.Fatalf("payload = %+v, want %+v", got, want)
	}
}

// TestBuildEventFrame_VoteCast_StaysAnonymous is the executable form of the
// anonymity contract dto.VoteCastPayload's doc comment claims: the marshaled
// frame must reveal neither who voted nor what they voted for, even though
// the domain event carries both.
func TestBuildEventFrame_VoteCast_StaysAnonymous(t *testing.T) {
	pid := game.ParticipantID{9}
	evt := game.VoteCast{
		RoundIndex: 0, ParticipantID: pid, Option: "FALL",
		HumanVotesCast: 1, HumanVoters: 1,
	}
	_, payload, _ := buildEventFrame(evt, 30*time.Second, 0, 0, time.Time{})

	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	s := string(raw)
	if strings.Contains(s, "participantId") || strings.Contains(s, "option") {
		t.Fatalf("VOTE_CAST payload leaked participant/option keys: %s", s)
	}
	if strings.Contains(s, pid.String()) || strings.Contains(s, "FALL") {
		t.Fatalf("VOTE_CAST payload leaked the participant id or option value: %s", s)
	}
}
