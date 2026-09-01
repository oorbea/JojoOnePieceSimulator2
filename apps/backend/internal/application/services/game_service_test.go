package services_test

import (
	"context"
	"errors"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// --- fakes ---

// fakeIDGen is a generic ports.IIdGenerator[T] that hands out deterministic,
// distinct ids by incrementing the last byte - mirrors the repo's
// fakeStandIDGenerator convention.
type fakeIDGen[T ~[16]byte] struct {
	mu sync.Mutex
	n  byte
}

func newFakeIDGen[T ~[16]byte]() *fakeIDGen[T] { return &fakeIDGen[T]{} }

func (g *fakeIDGen[T]) NewID() T {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.n++
	var b [16]byte
	b[15] = g.n
	return T(b)
}

type fakeGameStore struct {
	mu     sync.Mutex
	byID   map[game.GameID]*game.Game
	byCode map[string]game.GameID
	// savedTTL records the TTL each game was last saved under (0 = the
	// store's default), so tests can assert a terminal game got the short
	// one instead of lingering for the full lobby TTL.
	savedTTL map[game.GameID]time.Duration
}

func newFakeGameStore() *fakeGameStore {
	return &fakeGameStore{
		byID:     make(map[game.GameID]*game.Game),
		byCode:   make(map[string]game.GameID),
		savedTTL: make(map[game.GameID]time.Duration),
	}
}

// ttlFor reports the TTL id was last saved under (0 = the store's default).
func (s *fakeGameStore) ttlFor(id game.GameID) time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.savedTTL[id]
}

func (s *fakeGameStore) Create(_ context.Context, code string, g *game.Game) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.byCode[code]; ok {
		return ports.ErrGameCodeTaken
	}
	s.byID[g.ID()] = g
	s.byCode[code] = g.ID()
	return nil
}

func (s *fakeGameStore) Get(_ context.Context, id game.GameID) (*game.Game, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	g, ok := s.byID[id]
	if !ok {
		return nil, ports.ErrGameNotFound
	}
	return g, nil
}

func (s *fakeGameStore) GetByCode(_ context.Context, code string) (*game.Game, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id, ok := s.byCode[code]
	if !ok {
		return nil, ports.ErrGameNotFound
	}
	g, ok := s.byID[id]
	if !ok {
		return nil, ports.ErrGameNotFound
	}
	return g, nil
}

func (s *fakeGameStore) Code(_ context.Context, id game.GameID) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for code, gid := range s.byCode {
		if gid == id {
			return code, nil
		}
	}
	return "", ports.ErrGameNotFound
}

func (s *fakeGameStore) Save(ctx context.Context, g *game.Game) error {
	return s.SaveWithTTL(ctx, g, 0)
}

func (s *fakeGameStore) SaveWithTTL(_ context.Context, g *game.Game, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.byID[g.ID()]; !ok {
		return ports.ErrGameNotFound
	}
	s.byID[g.ID()] = g
	s.savedTTL[g.ID()] = ttl
	return nil
}

func (s *fakeGameStore) Delete(_ context.Context, id game.GameID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.byID[id]; ok {
		delete(s.byID, id)
		for code, gid := range s.byCode {
			if gid == id {
				delete(s.byCode, code)
			}
		}
	}
	return nil
}

func (s *fakeGameStore) DeleteExpired(_ context.Context, _ time.Duration) int { return 0 }

func (s *fakeGameStore) ListPublic(_ context.Context, limit int) ([]*game.Game, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]*game.Game, 0)
	for _, g := range s.byID {
		if g.IsPubliclyJoinable() {
			out = append(out, g)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID().String() < out[j].ID().String() })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

var _ ports.IGameStore = (*fakeGameStore)(nil)

type fakeStageCatalog struct {
	stages map[enums.Manga][]game.Stage
	err    error
}

func (f *fakeStageCatalog) Stages(_ context.Context, m enums.Manga) ([]game.Stage, error) {
	if f.err != nil {
		return nil, f.err
	}
	return append([]game.Stage(nil), f.stages[m]...), nil
}

var _ ports.IStageCatalog = (*fakeStageCatalog)(nil)

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

type fakeAssignmentWeights struct{ w game.AssignmentWeights }

func (f fakeAssignmentWeights) Load(context.Context) (game.AssignmentWeights, error) { return f.w, nil }

var _ ports.IAssignmentWeights = fakeAssignmentWeights{}

type fakeTiebreaker struct {
	mu     sync.Mutex
	winner string
	err    error
	calls  int
}

func (f *fakeTiebreaker) Break(_ context.Context, options []string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	if f.err != nil {
		return "", f.err
	}
	if f.winner != "" {
		return f.winner, nil
	}
	return options[0], nil
}

var _ ports.ITiebreaker = (*fakeTiebreaker)(nil)

type fakeGameHistory struct {
	mu      sync.Mutex
	results []game.GameResult
	err     error
}

func (f *fakeGameHistory) Record(_ context.Context, r game.GameResult) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return f.err
	}
	f.results = append(f.results, r)
	return nil
}

func (f *fakeGameHistory) all() []game.GameResult {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]game.GameResult(nil), f.results...)
}

var _ ports.IGameHistory = (*fakeGameHistory)(nil)

// fakeRandom is a deterministic game.RandomSource: it pops values from a
// fixed queue (cycling once exhausted), reduced modulo n.
type fakeRandom struct {
	mu  sync.Mutex
	seq []int
	i   int
}

func (r *fakeRandom) IntN(n int) int {
	if n <= 0 {
		return 0
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.seq) == 0 {
		return 0
	}
	v := r.seq[r.i%len(r.seq)]
	r.i++
	return v % n
}

var _ game.RandomSource = (*fakeRandom)(nil)

// fakeUserRepository is shared with auth_service_test.go (same
// services_test package) - see that file for its definition.

// fakeTimer/fakeClock let voting-window tests advance time deterministically
// instead of racing a real timer.
type fakeTimer struct {
	clock   *fakeClock
	fn      func()
	fired   bool
	stopped bool
}

func (t *fakeTimer) Stop() bool {
	t.clock.mu.Lock()
	defer t.clock.mu.Unlock()
	if t.fired || t.stopped {
		return false
	}
	t.stopped = true
	return true
}

type pendingTimer struct {
	deadline time.Time
	timer    *fakeTimer
}

type fakeClock struct {
	mu     sync.Mutex
	now    time.Time
	timers []*pendingTimer
}

func newFakeClock() *fakeClock { return &fakeClock{now: time.Unix(0, 0)} }

// newFakeClockAt starts a clock at an arbitrary instant - used by the
// phase-deadline tests to stand up a "restarted" service whose clock is
// already partway into a window the previous instance armed.
func newFakeClockAt(now time.Time) *fakeClock { return &fakeClock{now: now} }

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) AfterFunc(d time.Duration, f func()) services.Timer {
	c.mu.Lock()
	defer c.mu.Unlock()
	t := &fakeTimer{clock: c, fn: f}
	c.timers = append(c.timers, &pendingTimer{deadline: c.now.Add(d), timer: t})
	return t
}

// Advance moves the clock forward by d and synchronously fires (in deadline
// order) every timer that is now due and hasn't already fired/been
// stopped.
func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	c.now = c.now.Add(d)
	var due []*pendingTimer
	var remaining []*pendingTimer
	for _, pt := range c.timers {
		if pt.timer.stopped || pt.timer.fired {
			continue
		}
		if !pt.deadline.After(c.now) {
			due = append(due, pt)
		} else {
			remaining = append(remaining, pt)
		}
	}
	c.timers = remaining
	c.mu.Unlock()

	sort.Slice(due, func(i, j int) bool { return due[i].deadline.Before(due[j].deadline) })
	for _, pt := range due {
		pt.timer.fired = true
		pt.timer.fn()
	}
}

var _ services.Clock = (*fakeClock)(nil)

// --- fixtures ---

var stageIDCounter byte

func mustStage(t *testing.T, manga enums.Manga, order int, name string) game.Stage {
	t.Helper()
	stageIDCounter++
	var id game.StageID
	id[15] = stageIDCounter
	s, err := game.NewStage(id, manga, order, name, "a test stage", "")
	if err != nil {
		t.Fatalf("mustStage: %v", err)
	}
	return s
}

var powerIDCounter byte

func mustStand(t *testing.T, name string) *powers.Stand {
	t.Helper()
	powerIDCounter++
	var id powers.PowerID
	id[15] = powerIDCounter
	skills := []string{"skill"}
	power, err := powers.NewPower(id, name, "description", enums.Common, &skills, "")
	if err != nil {
		t.Fatalf("mustStand power: %v", err)
	}
	stand, err := powers.NewStand(*power, enums.B, enums.B, enums.B, enums.B, enums.B, enums.B, nil)
	if err != nil {
		t.Fatalf("mustStand: %v", err)
	}
	return stand
}

func mustDevilFruit(t *testing.T, name string) *powers.DevilFruit {
	t.Helper()
	powerIDCounter++
	var id powers.PowerID
	id[15] = powerIDCounter
	skills := []string{"skill"}
	power, err := powers.NewPower(id, name, "description", enums.Common, &skills, "")
	if err != nil {
		t.Fatalf("mustDevilFruit power: %v", err)
	}
	fruit, err := powers.NewDevilFruit(*power, enums.Paramecia)
	if err != nil {
		t.Fatalf("mustDevilFruit: %v", err)
	}
	return fruit
}

var userIDCounter byte

func mustTestUser(t *testing.T, deps *gameTestDeps, username string) user.UserID {
	t.Helper()
	userIDCounter++
	var id user.UserID
	id[15] = userIDCounter
	u, err := user.NewUser(id, "google-sub-"+username, username+"@example.com", username, username, "", enums.Regular)
	if err != nil {
		t.Fatalf("mustTestUser: %v", err)
	}
	if err := deps.users.Save(context.Background(), u); err != nil {
		t.Fatalf("mustTestUser save: %v", err)
	}
	return id
}

// --- test wiring ---

type gameTestDeps struct {
	store    *fakeGameStore
	stages   *fakeStageCatalog
	powers   *fakeGamePowerPool
	weights  fakeAssignmentWeights
	tiebreak *fakeTiebreaker
	history  *fakeGameHistory
	rng      *fakeRandom
	hub      *services.GameEventHub
	clock    *fakeClock
	users    *fakeUserRepository
	// finishedID is set by the finished-game helpers so a test can keep
	// asserting against a game after the aggregate itself has gone terminal.
	finishedID game.GameID
}

// newTestGameService builds a GameService with every port faked, all
// wiring exposed via the returned gameTestDeps for assertions.
func newTestGameService(t *testing.T) (*services.GameService, *gameTestDeps) {
	t.Helper()
	deps := &gameTestDeps{
		store: newFakeGameStore(),
		stages: &fakeStageCatalog{stages: map[enums.Manga][]game.Stage{
			enums.Jojo:     {mustStage(t, enums.Jojo, 0, "Phantom Blood"), mustStage(t, enums.Jojo, 1, "Battle Tendency")},
			enums.OnePiece: {mustStage(t, enums.OnePiece, 0, "East Blue"), mustStage(t, enums.OnePiece, 1, "Alabasta")},
		}},
		powers:   &fakeGamePowerPool{stands: []*powers.Stand{mustStand(t, "Star Platinum"), mustStand(t, "Crazy Diamond")}, fruits: []*powers.DevilFruit{mustDevilFruit(t, "Gomu Gomu no Mi"), mustDevilFruit(t, "Mera Mera no Mi")}},
		weights:  fakeAssignmentWeights{w: game.DefaultAssignmentWeights()},
		tiebreak: &fakeTiebreaker{},
		history:  &fakeGameHistory{},
		rng:      &fakeRandom{seq: []int{1, 2, 3, 0, 1}},
		hub:      services.NewGameEventHub(),
		clock:    newFakeClock(),
		users:    newFakeUserRepository(),
	}
	svc := newGameServiceFromDeps(deps, deps.history)
	return svc, deps
}

func newGameServiceFromDeps(deps *gameTestDeps, history ports.IGameHistory) *services.GameService {
	return services.NewGameService(
		deps.store,
		newFakeIDGen[game.GameID](),
		newFakeIDGen[game.ParticipantID](),
		newFakeIDGen[game.TeamID](),
		deps.users,
		deps.stages,
		deps.powers,
		deps.weights,
		deps.tiebreak,
		history,
		deps.rng,
		deps.hub,
		deps.clock,
		services.VotingPolicy{Window: 30 * time.Second},
	)
}

func gauntletInput() services.CreateGameInput {
	return services.CreateGameInput{
		Mode: enums.Gauntlet, StageMangas: []enums.Manga{enums.Jojo}, PowerMangas: []enums.Manga{enums.Jojo}, AbilitySource: enums.Random,
		TeamSize: 5, AllowBots: false,
	}
}

func versusInput(teamSize int) services.CreateGameInput {
	return services.CreateGameInput{
		Mode: enums.Versus, StageMangas: []enums.Manga{enums.Jojo, enums.OnePiece}, PowerMangas: []enums.Manga{enums.Jojo, enums.OnePiece}, AbilitySource: enums.Random,
		TeamSize: teamSize, AllowBots: true,
	}
}

func hostOf(t *testing.T, g *game.Game) game.ParticipantID {
	t.Helper()
	return g.HostID()
}

// advanceReveal fires the reveal-delay timer scheduleRevealDelay starts
// whenever a round (re)assigns Loadouts - StartGame, and every Versus round
// after it - by advancing deps.clock past the longest possible
// game.RevealDuration for mangas (see maxRevealDuration: since 2026-08-30
// the real duration also depends on which participants actually landed a
// Stand/Devil Fruit and on RevealSpeed, neither of which every caller here
// has in hand, and fakeClock.Advance fires any timer that is due by the new
// "now" regardless of overshoot). Until this timer fires the Game sits in
// ASSIGNING, not VOTING, so any test that needs to actually cast a vote
// must call this (then re-fetch via GetGame, since the timer's own
// openVotingAfterReveal runs asynchronously against the fake clock, not
// inline with the call that scheduled it).
func advanceReveal(deps *gameTestDeps, mangas []enums.Manga) {
	deps.clock.Advance(maxRevealDuration(mangas))
}

// advanceSummary fires the summary-delay timer scheduleSummaryDelay starts
// once the reveal finishes (2026-08-30's loadout-summary screen, between
// ASSIGNING and VOTING) - by advancing deps.clock past the lobby's
// configured SummaryDurationSeconds (game.DefaultSummaryDurationSeconds for
// every test lobby here, since none of them set it explicitly). Until this
// timer fires the Game sits in SUMMARY, not VOTING, so any test that needs
// to actually cast a vote must call this after advanceReveal (then re-fetch
// via GetGame, same async-timer caveat as advanceReveal/advanceResult).
func advanceSummary(deps *gameTestDeps) {
	deps.clock.Advance(time.Duration(game.DefaultSummaryDurationSeconds) * time.Second)
}

// maxRevealDuration upper-bounds game.RevealDuration for mangas across any
// number of players (up to game.MaxGauntletPlayers, the largest a lobby
// this test suite builds ever seats), any draw outcome, and the slowest
// RevealSpeed - enough headroom for advanceReveal to always fire the timer
// it targets. Doubled on top of that because game.RevealSpinCycles picks 1
// or 2 spin cycles deterministically from (gameID, roundIndex, participant,
// slot) - inputs this helper can't reproduce exactly for an arbitrary real
// game, so it pads for the worst case where every slot happens to land 2
// cycles instead of whatever mix the real gameID's hash actually produced.
func maxRevealDuration(mangas []enums.Manga) time.Duration {
	players := make([]game.RevealPlayer, game.MaxGauntletPlayers)
	for i := range players {
		players[i] = game.RevealPlayer{
			HasStand: true, HasDevilFruit: true,
			HasArmamentHaki: true, HasObservationHaki: true, HasConquerorHaki: true,
		}
	}
	return 2 * game.RevealDuration(game.GameID{}, 0, mangas, players, enums.Relaxed)
}

// revealDurationOf mirrors GameService's own (unexported) revealDurationFor
// using only exported *game.Game API, so a test can assert the exact
// deadline the service computed without duplicating its internals: g's own
// mangas/RevealSpeed, current round index, and each participant's own
// already-assigned Loadout.
func revealDurationOf(g *game.Game) time.Duration {
	players := make([]game.RevealPlayer, 0, len(g.Participants()))
	for _, p := range g.Participants() {
		loadout := p.Loadout()
		players = append(players, game.RevealPlayer{
			HasStand:           loadout != nil && loadout.Stand() != nil,
			HasDevilFruit:      loadout != nil && loadout.DevilFruit() != nil,
			HasArmamentHaki:    loadout != nil && loadout.ArmamentHaki() != enums.HakiNone,
			HasObservationHaki: loadout != nil && loadout.ObservationHaki() != enums.HakiNone,
			HasConquerorHaki:   loadout != nil && loadout.ConquerorHaki() != enums.HakiNone,
		})
	}
	return game.RevealDuration(g.ID(), len(g.Rounds()), g.Config().PowerMangas(), players, g.Config().RevealSpeed())
}

// advanceResult fires the result-display timer scheduleResultDelay starts
// once a round resolves (a clear winner, or a tiebreak winner) - by
// advancing deps.clock past game.ResultDuration. Until this timer fires the
// Game sits in RESOLVING, not ASSIGNING/FINISHED/VOTING, so any test that
// needs to observe the round's aftermath (next round opening, or the match
// finishing) must call this (then re-fetch via GetGame, same caveat as
// advanceReveal).
func advanceResult(deps *gameTestDeps) {
	deps.clock.Advance(game.ResultDuration)
}

// assertStillReadableTerminal replaces the "GetGame must now return
// ErrGameNotFound" assertions these tests used to make. finalizeLocked no
// longer deletes a terminal game - it saves it under services.
// FinishedGameTTL so its players can actually read the final result screen
// (the instant delete was the real blocker for one). What these tests were
// ever really pinning is that the game reached a terminal state and no stale
// timer dragged it back out of one, which is what this asserts instead.
func assertStillReadableTerminal(t *testing.T, svc *services.GameService, id game.GameID) {
	t.Helper()
	g, err := svc.GetGame(context.Background(), id)
	if err != nil {
		t.Fatalf("GetGame on a finalized game: err = %v, want it to still be readable", err)
	}
	if g.State() != enums.Finished && g.State() != enums.Aborted {
		t.Fatalf("state after finalize = %v, want FINISHED or ABORTED", g.State())
	}
}

// --- creation / membership ---

func TestCreateGame_Gauntlet_Success(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")

	g, code, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	if len(code) != 6 {
		t.Fatalf("code length = %d, want 6", len(code))
	}
	if len(g.Teams()) != 1 {
		t.Fatalf("teams = %d, want 1", len(g.Teams()))
	}
	if g.State() != enums.Lobby {
		t.Fatalf("state = %v, want LOBBY", g.State())
	}
}

func TestCreateGame_Versus_Success(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")

	g, _, err := svc.CreateGame(context.Background(), hostID, versusInput(2))
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	if len(g.Teams()) != 2 {
		t.Fatalf("teams = %d, want 2", len(g.Teams()))
	}
}

func TestCreateGame_InventoryAbilitySource_Rejected(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")

	input := gauntletInput()
	input.AbilitySource = enums.Inventory
	_, _, err := svc.CreateGame(context.Background(), hostID, input)
	if !errors.Is(err, game.ErrInventoryNotSupported) {
		t.Fatalf("err = %v, want ErrInventoryNotSupported", err)
	}
}

func TestCreateGame_NoStagesForManga_ReturnsError(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")
	deps.stages.stages[enums.Jojo] = nil

	_, _, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if !errors.Is(err, game.ErrNoStagesAvailable) {
		t.Fatalf("err = %v, want ErrNoStagesAvailable", err)
	}
}

func TestJoinByCode_Success(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")
	joinerID := mustTestUser(t, deps, "joiner")

	_, code, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	g, err := svc.JoinByCode(context.Background(), code, joinerID)
	if err != nil {
		t.Fatalf("JoinByCode: %v", err)
	}
	if len(g.Participants()) != 2 {
		t.Fatalf("participants = %d, want 2", len(g.Participants()))
	}
}

func TestJoinByCode_UnknownCode_ReturnsNotFound(t *testing.T) {
	svc, deps := newTestGameService(t)
	joinerID := mustTestUser(t, deps, "joiner")

	_, err := svc.JoinByCode(context.Background(), "NOPE12", joinerID)
	if !errors.Is(err, ports.ErrGameNotFound) {
		t.Fatalf("err = %v, want ErrGameNotFound", err)
	}
}

func TestJoinByCode_AlreadyInGame_ReturnsError(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")

	_, code, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	_, err = svc.JoinByCode(context.Background(), code, hostID)
	if !errors.Is(err, services.ErrAlreadyInGame) {
		t.Fatalf("err = %v, want ErrAlreadyInGame", err)
	}
}

func TestJoinByCode_GameFull_ReturnsError(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")
	joinerID := mustTestUser(t, deps, "joiner")

	input := gauntletInput()
	input.TeamSize = 1
	_, code, err := svc.CreateGame(context.Background(), hostID, input)
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	_, err = svc.JoinByCode(context.Background(), code, joinerID)
	if !errors.Is(err, game.ErrGameFull) {
		t.Fatalf("err = %v, want ErrGameFull", err)
	}
}

func TestAddBot_NotHost_ReturnsError(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")
	joinerID := mustTestUser(t, deps, "joiner")

	g, code, err := svc.CreateGame(context.Background(), hostID, versusInput(2))
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	g, err = svc.JoinByCode(context.Background(), code, joinerID)
	if err != nil {
		t.Fatalf("JoinByCode: %v", err)
	}
	var joinerParticipant game.ParticipantID
	for _, p := range g.Participants() {
		if p.ID() != g.HostID() {
			joinerParticipant = p.ID()
		}
	}
	_, err = svc.AddBot(context.Background(), g.ID(), joinerParticipant, g.Teams()[1].ID())
	if !errors.Is(err, game.ErrNotHost) {
		t.Fatalf("err = %v, want ErrNotHost", err)
	}
}

func TestAddBot_GauntletRejected(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")

	g, _, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	_, err = svc.AddBot(context.Background(), g.ID(), g.HostID(), g.Teams()[0].ID())
	if !errors.Is(err, game.ErrBotsNotAllowed) {
		t.Fatalf("err = %v, want ErrBotsNotAllowed", err)
	}
}

func TestRemoveBot_NotABot_ReturnsError(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")

	g, _, err := svc.CreateGame(context.Background(), hostID, versusInput(2))
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	_, err = svc.RemoveBot(context.Background(), g.ID(), g.HostID(), g.HostID())
	if !errors.Is(err, services.ErrNotABot) {
		t.Fatalf("err = %v, want ErrNotABot", err)
	}
}

func TestAddBotThenRemoveBot_Success(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")

	g, _, err := svc.CreateGame(context.Background(), hostID, versusInput(2))
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	g, err = svc.AddBot(context.Background(), g.ID(), g.HostID(), g.Teams()[1].ID())
	if err != nil {
		t.Fatalf("AddBot: %v", err)
	}
	var botID game.ParticipantID
	for _, p := range g.Participants() {
		if p.IsBot() {
			botID = p.ID()
		}
	}
	if botID.IsNil() {
		t.Fatalf("no bot participant found")
	}
	g, err = svc.RemoveBot(context.Background(), g.ID(), g.HostID(), botID)
	if err != nil {
		t.Fatalf("RemoveBot: %v", err)
	}
	if len(g.Participants()) != 1 {
		t.Fatalf("participants = %d, want 1", len(g.Participants()))
	}
}

// --- lifecycle / voting ---

func TestStartGame_NotHost_ReturnsError(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")
	joinerID := mustTestUser(t, deps, "joiner")

	g, code, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	g, err = svc.JoinByCode(context.Background(), code, joinerID)
	if err != nil {
		t.Fatalf("JoinByCode: %v", err)
	}
	var joinerParticipant game.ParticipantID
	for _, p := range g.Participants() {
		if p.ID() != g.HostID() {
			joinerParticipant = p.ID()
		}
	}
	_, err = svc.StartGame(context.Background(), g.ID(), joinerParticipant)
	if !errors.Is(err, game.ErrNotHost) {
		t.Fatalf("err = %v, want ErrNotHost", err)
	}
}

func TestStartGame_VersusEmptyTeam_ReturnsError(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")

	g, _, err := svc.CreateGame(context.Background(), hostID, versusInput(2))
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	// Only the host is seated (team A) - team B is empty.
	_, err = svc.StartGame(context.Background(), g.ID(), g.HostID())
	if !errors.Is(err, game.ErrNotEnoughPlayers) {
		t.Fatalf("err = %v, want ErrNotEnoughPlayers", err)
	}
}

func TestStartGame_VersusUnequalTeams_ReturnsError(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")
	joinerAID := mustTestUser(t, deps, "joinerA")
	joinerBID := mustTestUser(t, deps, "joinerB")

	g, code, err := svc.CreateGame(context.Background(), hostID, versusInput(2))
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	// pickTeam auto-balances: host->A(1), joinerA->B(1v1), joinerB-> a tie
	// favors team A -> A(2) vs B(1). Both teams are non-empty but unequal.
	if _, err := svc.JoinByCode(context.Background(), code, joinerAID); err != nil {
		t.Fatalf("JoinByCode joinerA: %v", err)
	}
	if _, err := svc.JoinByCode(context.Background(), code, joinerBID); err != nil {
		t.Fatalf("JoinByCode joinerB: %v", err)
	}
	_, err = svc.StartGame(context.Background(), g.ID(), g.HostID())
	if !errors.Is(err, game.ErrTeamSizeMismatch) {
		t.Fatalf("err = %v, want ErrTeamSizeMismatch", err)
	}
}

func TestStartGame_Gauntlet_OpensFirstRoundVoting(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")

	g, _, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	g, err = svc.StartGame(context.Background(), g.ID(), hostOf(t, g))
	if err != nil {
		t.Fatalf("StartGame: %v", err)
	}
	if g.State() != enums.Assigning {
		t.Fatalf("state right after StartGame = %v, want ASSIGNING (reveal in progress)", g.State())
	}

	advanceReveal(deps, gauntletInput().PowerMangas)
	advanceSummary(deps)
	g, err = svc.GetGame(context.Background(), g.ID())
	if err != nil {
		t.Fatalf("GetGame after reveal: %v", err)
	}
	if g.State() != enums.Voting {
		t.Fatalf("state = %v, want VOTING", g.State())
	}
	if len(g.Rounds()) != 1 {
		t.Fatalf("rounds = %d, want 1", len(g.Rounds()))
	}
}

// TestStartGame_RevealEndsAt_TracksThenClearsOnVotingOpen checks
// GameService.RevealEndsAt end to end: absent before the reveal starts,
// present with the expected deadline while ASSIGNING, and gone again the
// moment voting actually opens - the seam a (re)connecting client relies on
// (see dto.NewGameStateResponse's revealEndsAt param).
func TestStartGame_RevealEndsAt_TracksThenClearsOnVotingOpen(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")

	g, _, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	if _, ok := svc.RevealEndsAt(g.ID()); ok {
		t.Fatalf("RevealEndsAt before StartGame: want not-ok")
	}

	g, err = svc.StartGame(context.Background(), g.ID(), g.HostID())
	if err != nil {
		t.Fatalf("StartGame: %v", err)
	}
	want := deps.clock.Now().Add(revealDurationOf(g))
	got, ok := svc.RevealEndsAt(g.ID())
	if !ok {
		t.Fatalf("RevealEndsAt during reveal: want ok")
	}
	if !got.Equal(want) {
		t.Fatalf("RevealEndsAt = %v, want %v", got, want)
	}

	advanceReveal(deps, gauntletInput().PowerMangas)
	advanceSummary(deps)
	if _, ok := svc.RevealEndsAt(g.ID()); ok {
		t.Fatalf("RevealEndsAt after voting opens: want not-ok")
	}
}

// TestAbortGame_DuringReveal_NeverOpensVoting guards the exact bug the
// reveal delay could otherwise introduce: aborting a Game while its reveal
// timer is still pending must cancel that timer outright, not just let it
// fire into a Game that no longer exists. Advancing the clock well past the
// reveal duration after the abort proves openVotingAfterReveal's own
// ErrGameNotFound tolerance is never even exercised here - the timer is
// gone, not merely harmless.
func TestAbortGame_DuringReveal_NeverOpensVoting(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")

	g, _, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	g, err = svc.StartGame(context.Background(), g.ID(), g.HostID())
	if err != nil {
		t.Fatalf("StartGame: %v", err)
	}
	if g.State() != enums.Assigning {
		t.Fatalf("state right after StartGame = %v, want ASSIGNING", g.State())
	}

	if _, err := svc.AbortGame(context.Background(), g.ID(), g.HostID()); err != nil {
		t.Fatalf("AbortGame: %v", err)
	}
	if _, ok := svc.RevealEndsAt(g.ID()); ok {
		t.Fatalf("RevealEndsAt after abort: want not-ok, the reveal timer should be cancelled")
	}

	// Advance well past the reveal duration - if the timer had survived the
	// abort, this is where it would fire and try to reopen voting on a Game
	// that finalizeLocked already deleted from the store.
	advanceReveal(deps, gauntletInput().PowerMangas)
	advanceSummary(deps)

	assertStillReadableTerminal(t, svc, g.ID())
}

func TestGauntlet_FallMajority_FinishesAndFinalizes(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")

	g, _, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	g, err = svc.StartGame(context.Background(), g.ID(), g.HostID())
	if err != nil {
		t.Fatalf("StartGame: %v", err)
	}
	advanceReveal(deps, gauntletInput().PowerMangas)
	advanceSummary(deps)
	g, err = svc.GetGame(context.Background(), g.ID())
	if err != nil {
		t.Fatalf("GetGame after reveal: %v", err)
	}

	g, err = svc.CastVote(context.Background(), g.ID(), g.HostID(), game.OptionID("FALL"))
	if err != nil {
		t.Fatalf("CastVote: %v", err)
	}
	// The clear FALL majority leaves the Game in RESOLVING (see
	// GameService.scheduleResultDelay) until the result display's own
	// timer elapses - only then does CompleteRound advance it to FINISHED.
	advanceResult(deps)
	if g.State() != enums.Finished {
		t.Fatalf("state = %v, want FINISHED", g.State())
	}

	results := deps.history.all()
	if len(results) != 1 {
		t.Fatalf("history records = %d, want 1", len(results))
	}
	if results[0].Winner != "FALL" || results[0].Aborted {
		t.Fatalf("result = %+v, want Winner=FALL, Aborted=false", results[0])
	}

	assertStillReadableTerminal(t, svc, g.ID())
}

func TestGauntlet_ClearAllStages_Victory(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")

	g, _, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	g, err = svc.StartGame(context.Background(), g.ID(), g.HostID())
	if err != nil {
		t.Fatalf("StartGame: %v", err)
	}
	advanceReveal(deps, gauntletInput().PowerMangas)
	advanceSummary(deps)
	g, err = svc.GetGame(context.Background(), g.ID())
	if err != nil {
		t.Fatalf("GetGame after reveal: %v", err)
	}

	// gauntletInput selects only Jojo, which has 2 fixture stages - two
	// SURVIVE votes should clear the run.
	g, err = svc.CastVote(context.Background(), g.ID(), g.HostID(), game.OptionID("SURVIVE"))
	if err != nil {
		t.Fatalf("CastVote round 1: %v", err)
	}
	// Gauntlet never reassigns after round 1 (ReassignsEachRound is false),
	// so once the result display elapses, round 2's voting opens
	// immediately - no reveal delay in between.
	advanceResult(deps)
	if g.State() != enums.Voting {
		t.Fatalf("state after round 1 = %v, want VOTING", g.State())
	}

	g, err = svc.CastVote(context.Background(), g.ID(), g.HostID(), game.OptionID("SURVIVE"))
	if err != nil {
		t.Fatalf("CastVote round 2: %v", err)
	}
	advanceResult(deps)
	if g.State() != enums.Finished {
		t.Fatalf("state after round 2 = %v, want FINISHED", g.State())
	}

	results := deps.history.all()
	if len(results) != 1 || results[0].Winner != "SURVIVE" {
		t.Fatalf("results = %+v, want one SURVIVE result", results)
	}
}

func TestVersus_ThreeRounds_TeamAWins(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")
	joinerID := mustTestUser(t, deps, "joiner")

	g, code, err := svc.CreateGame(context.Background(), hostID, versusInput(1))
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	g, err = svc.JoinByCode(context.Background(), code, joinerID)
	if err != nil {
		t.Fatalf("JoinByCode: %v", err)
	}
	g, err = svc.StartGame(context.Background(), g.ID(), g.HostID())
	if err != nil {
		t.Fatalf("StartGame: %v", err)
	}
	advanceReveal(deps, versusInput(1).PowerMangas)
	advanceSummary(deps)
	g, err = svc.GetGame(context.Background(), g.ID())
	if err != nil {
		t.Fatalf("GetGame after reveal: %v", err)
	}

	teamA := g.Teams()[0].ID()
	optionA := game.OptionID(teamA.String())

	var joinerParticipant game.ParticipantID
	for _, p := range g.Participants() {
		if p.ID() != g.HostID() {
			joinerParticipant = p.ID()
		}
	}

	for round := 0; round < game.VersusRounds; round++ {
		g, err = svc.CastVote(context.Background(), g.ID(), g.HostID(), optionA)
		if err != nil {
			t.Fatalf("round %d host vote: %v", round, err)
		}
		g, err = svc.CastVote(context.Background(), g.ID(), joinerParticipant, optionA)
		if err != nil {
			t.Fatalf("round %d joiner vote: %v", round, err)
		}
		// The unanimous vote leaves the Game in RESOLVING until the result
		// display's own timer elapses (see scheduleResultDelay).
		advanceResult(deps)
		// Versus reassigns Loadouts every round (see VersusMode.
		// ReassignsEachRound), so every round but the last one just
		// resolved schedules its own reveal delay before voting reopens.
		if round < game.VersusRounds-1 {
			advanceReveal(deps, versusInput(1).PowerMangas)
			advanceSummary(deps)
			g, err = svc.GetGame(context.Background(), g.ID())
			if err != nil {
				t.Fatalf("GetGame after round %d reveal: %v", round, err)
			}
		}
	}

	if g.State() != enums.Finished {
		t.Fatalf("state = %v, want FINISHED", g.State())
	}
	results := deps.history.all()
	if len(results) != 1 || results[0].Winner != optionA || results[0].RoundsPlayed != game.VersusRounds {
		t.Fatalf("results = %+v, want team A to win all %d rounds", results, game.VersusRounds)
	}
}

func TestCloseVoting_Tie_OpensRevoteThenUsesTiebreaker(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")
	joinerID := mustTestUser(t, deps, "joiner")

	g, code, err := svc.CreateGame(context.Background(), hostID, versusInput(1))
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	g, err = svc.JoinByCode(context.Background(), code, joinerID)
	if err != nil {
		t.Fatalf("JoinByCode: %v", err)
	}
	g, err = svc.StartGame(context.Background(), g.ID(), g.HostID())
	if err != nil {
		t.Fatalf("StartGame: %v", err)
	}
	advanceReveal(deps, versusInput(1).PowerMangas)
	advanceSummary(deps)
	g, err = svc.GetGame(context.Background(), g.ID())
	if err != nil {
		t.Fatalf("GetGame after reveal: %v", err)
	}

	teamA := game.OptionID(g.Teams()[0].ID().String())
	teamB := game.OptionID(g.Teams()[1].ID().String())
	deps.tiebreak.winner = string(teamA)

	var joinerParticipant game.ParticipantID
	for _, p := range g.Participants() {
		if p.ID() != g.HostID() {
			joinerParticipant = p.ID()
		}
	}

	sub, unsub := deps.hub.Subscribe(g.ID())
	var mu sync.Mutex
	var events []services.GameEvent
	done := make(chan struct{})
	go func() {
		for e := range sub {
			mu.Lock()
			events = append(events, e)
			mu.Unlock()
		}
		close(done)
	}()

	// First close: a straight tie opens the revote window.
	g, err = svc.CastVote(context.Background(), g.ID(), g.HostID(), teamA)
	if err != nil {
		t.Fatalf("host vote 1: %v", err)
	}
	g, err = svc.CastVote(context.Background(), g.ID(), joinerParticipant, teamB)
	if err != nil {
		t.Fatalf("joiner vote 1: %v", err)
	}
	if g.State() != enums.Tiebreak {
		t.Fatalf("state after first tie = %v, want TIEBREAK", g.State())
	}
	if deps.tiebreak.calls != 0 {
		t.Fatalf("tiebreak calls = %d, want 0 before the revote", deps.tiebreak.calls)
	}

	// Second close (the revote): the tie opening the TIEBREAK window reset
	// the round's Ballot (see game.Ballot.Reset / Game.CloseVoting), so both
	// the host and the joiner must recast for VotingComplete to go true
	// again - a single recast is no longer enough to resolve it.
	g, err = svc.CastVote(context.Background(), g.ID(), g.HostID(), teamA)
	if err != nil {
		t.Fatalf("host vote 2: %v", err)
	}
	if g.State() != enums.Tiebreak {
		t.Fatalf("state after host's lone revote cast = %v, want still TIEBREAK", g.State())
	}
	g, err = svc.CastVote(context.Background(), g.ID(), joinerParticipant, teamB)
	if err != nil {
		t.Fatalf("joiner vote 2: %v", err)
	}
	// The revote just resolved via the tiebreaker, leaving the Game in
	// RESOLVING until the result display's own timer elapses.
	advanceResult(deps)
	// That advances Versus into round 2 - which reassigns Loadouts and so
	// schedules its own reveal delay before voting reopens (see
	// VersusMode.ReassignsEachRound).
	advanceReveal(deps, versusInput(1).PowerMangas)
	advanceSummary(deps)
	g, err = svc.GetGame(context.Background(), g.ID())
	if err != nil {
		t.Fatalf("GetGame after round 2 reveal: %v", err)
	}

	unsub()
	<-done

	if deps.tiebreak.calls != 1 {
		t.Fatalf("tiebreak calls = %d, want 1", deps.tiebreak.calls)
	}
	if g.State() != enums.Voting {
		t.Fatalf("state after tiebreak = %v, want VOTING (round 2)", g.State())
	}
	if len(g.Rounds()) != 2 {
		t.Fatalf("rounds = %d, want 2", len(g.Rounds()))
	}

	mu.Lock()
	defer mu.Unlock()
	found := false
	for _, e := range events {
		if rr, ok := e.Event.(game.RoundResolved); ok {
			if !rr.DecidedByCoinFlip || rr.Winner != teamA {
				t.Fatalf("RoundResolved = %+v, want DecidedByCoinFlip=true Winner=%s", rr, teamA)
			}
			found = true
		}
	}
	if !found {
		t.Fatalf("no RoundResolved event observed")
	}
}

func TestCloseVotingWindow_TimerExpiry_ResolvesWithEmittedVotes(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")

	g, _, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	g, err = svc.StartGame(context.Background(), g.ID(), g.HostID())
	if err != nil {
		t.Fatalf("StartGame: %v", err)
	}
	advanceReveal(deps, gauntletInput().PowerMangas)
	advanceSummary(deps)

	// Host never votes - the window expiring must resolve with zero votes,
	// which is a tie (see game.Ballot.Tally), opening a revote instead of
	// getting stuck.
	deps.clock.Advance(30 * time.Second)

	g, err = svc.GetGame(context.Background(), g.ID())
	if err != nil {
		t.Fatalf("GetGame: %v", err)
	}
	if g.State() != enums.Tiebreak {
		t.Fatalf("state = %v, want TIEBREAK", g.State())
	}
}

// --- presence ---

func TestDisconnect_HostReassigned(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")
	joinerID := mustTestUser(t, deps, "joiner")

	g, code, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	g, err = svc.JoinByCode(context.Background(), code, joinerID)
	if err != nil {
		t.Fatalf("JoinByCode: %v", err)
	}
	oldHost := g.HostID()

	g, err = svc.Disconnect(context.Background(), g.ID(), oldHost)
	if err != nil {
		t.Fatalf("Disconnect: %v", err)
	}
	if g.HostID() == oldHost || g.HostID().IsNil() {
		t.Fatalf("host = %v, want reassigned away from %v", g.HostID(), oldHost)
	}
}

func TestDisconnect_LastHuman_AbortsAndFinalizes(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")

	g, _, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}

	g, err = svc.Disconnect(context.Background(), g.ID(), g.HostID())
	if err != nil {
		t.Fatalf("Disconnect: %v", err)
	}
	if g.State() != enums.Aborted {
		t.Fatalf("state = %v, want ABORTED", g.State())
	}

	results := deps.history.all()
	if len(results) != 1 || !results[0].Aborted {
		t.Fatalf("results = %+v, want one aborted result", results)
	}
	assertStillReadableTerminal(t, svc, g.ID())
}

func TestFinalize_NilHistory_DoesNotError(t *testing.T) {
	_, deps := newTestGameService(t)
	svc := newGameServiceFromDeps(deps, nil)
	hostID := mustTestUser(t, deps, "host")

	g, _, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	if _, err := svc.Disconnect(context.Background(), g.ID(), g.HostID()); err != nil {
		t.Fatalf("Disconnect: %v", err)
	}
	assertStillReadableTerminal(t, svc, g.ID())
}

// --- concurrency ---

func TestCastVote_Concurrent_NoRace(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")

	g, code, err := svc.CreateGame(context.Background(), hostID, versusInput(2))
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	for i := 0; i < 3; i++ {
		u := mustTestUser(t, deps, "joiner"+string(rune('A'+i)))
		g, err = svc.JoinByCode(context.Background(), code, u)
		if err != nil {
			t.Fatalf("JoinByCode %d: %v", i, err)
		}
	}
	g, err = svc.StartGame(context.Background(), g.ID(), g.HostID())
	if err != nil {
		t.Fatalf("StartGame: %v", err)
	}
	advanceReveal(deps, versusInput(2).PowerMangas)
	advanceSummary(deps)
	g, err = svc.GetGame(context.Background(), g.ID())
	if err != nil {
		t.Fatalf("GetGame after reveal: %v", err)
	}

	optionA := game.OptionID(g.Teams()[0].ID().String())
	participants := g.Participants()

	var wg sync.WaitGroup
	for _, p := range participants {
		wg.Add(1)
		go func(id game.ParticipantID) {
			defer wg.Done()
			if _, err := svc.CastVote(context.Background(), g.ID(), id, optionA); err != nil {
				t.Errorf("CastVote(%v): %v", id, err)
			}
		}(p.ID())
	}
	wg.Wait()

	// A unanimous vote resolves round 1, leaving the Game in RESOLVING
	// until the result display's own timer elapses.
	advanceResult(deps)
	// That moves Versus into round 2 - which reassigns Loadouts and so
	// schedules its own reveal delay before voting reopens (see
	// VersusMode.ReassignsEachRound).
	advanceReveal(deps, versusInput(2).PowerMangas)
	advanceSummary(deps)

	final, err := svc.GetGame(context.Background(), g.ID())
	if err != nil {
		t.Fatalf("GetGame: %v", err)
	}
	if final.State() != enums.Voting {
		t.Fatalf("state = %v, want VOTING (round 2 opened)", final.State())
	}
	if len(final.Rounds()) != 2 {
		t.Fatalf("rounds = %d, want 2", len(final.Rounds()))
	}
}

// --- lobby management ---

func TestEditLobbyConfig_HostOnly(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")
	joinerID := mustTestUser(t, deps, "joiner")

	g, code, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	g, err = svc.JoinByCode(context.Background(), code, joinerID)
	if err != nil {
		t.Fatalf("JoinByCode: %v", err)
	}
	var joinerParticipant game.ParticipantID
	for _, p := range g.Participants() {
		if p.ID() != g.HostID() {
			joinerParticipant = p.ID()
		}
	}

	input := gauntletInput()
	input.Visibility = enums.Public
	if _, err := svc.EditLobbyConfig(context.Background(), g.ID(), joinerParticipant, input); !errors.Is(err, game.ErrNotHost) {
		t.Fatalf("err = %v, want ErrNotHost", err)
	}
	g, err = svc.EditLobbyConfig(context.Background(), g.ID(), g.HostID(), input)
	if err != nil {
		t.Fatalf("EditLobbyConfig: %v", err)
	}
	if g.Config().Visibility() != enums.Public {
		t.Fatalf("expected the edited Config to be applied")
	}
}

func TestEditLobbyConfig_ModeChangeReloadsStagesAndTeams(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")

	g, _, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}

	input := versusInput(2)
	g, err = svc.EditLobbyConfig(context.Background(), g.ID(), g.HostID(), input)
	if err != nil {
		t.Fatalf("EditLobbyConfig: %v", err)
	}
	if g.Config().Mode() != enums.Versus {
		t.Fatalf("expected the mode to switch to VERSUS")
	}
	if len(g.Teams()) != 2 {
		t.Fatalf("expected 2 teams after the mode switch, got %d", len(g.Teams()))
	}
}

func TestSwitchTeam_Service(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")
	joinerID := mustTestUser(t, deps, "joiner")

	g, code, err := svc.CreateGame(context.Background(), hostID, versusInput(2))
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	g, err = svc.JoinByCode(context.Background(), code, joinerID)
	if err != nil {
		t.Fatalf("JoinByCode: %v", err)
	}
	var joinerParticipant game.ParticipantID
	for _, p := range g.Participants() {
		if p.ID() != g.HostID() {
			joinerParticipant = p.ID()
		}
	}
	host, ok := g.Participant(g.HostID())
	if !ok {
		t.Fatalf("expected the host to be seated")
	}
	hostTeam := host.TeamID()
	g, err = svc.SwitchTeam(context.Background(), g.ID(), g.HostID(), joinerParticipant, hostTeam)
	if err != nil {
		t.Fatalf("SwitchTeam: %v", err)
	}
	p, ok := g.Participant(joinerParticipant)
	if !ok || p.TeamID() != hostTeam {
		t.Fatalf("expected the joiner to be seated on the host's team %v", hostTeam)
	}
}

func TestKickParticipant_Service(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")
	joinerID := mustTestUser(t, deps, "joiner")

	g, code, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	g, err = svc.JoinByCode(context.Background(), code, joinerID)
	if err != nil {
		t.Fatalf("JoinByCode: %v", err)
	}
	var joinerParticipant game.ParticipantID
	for _, p := range g.Participants() {
		if p.ID() != g.HostID() {
			joinerParticipant = p.ID()
		}
	}
	g, err = svc.KickParticipant(context.Background(), g.ID(), g.HostID(), joinerParticipant)
	if err != nil {
		t.Fatalf("KickParticipant: %v", err)
	}
	if len(g.Participants()) != 1 {
		t.Fatalf("expected the kicked participant to be removed")
	}
}

func TestTransferHost_Service(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")
	joinerID := mustTestUser(t, deps, "joiner")

	g, code, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	g, err = svc.JoinByCode(context.Background(), code, joinerID)
	if err != nil {
		t.Fatalf("JoinByCode: %v", err)
	}
	var joinerParticipant game.ParticipantID
	for _, p := range g.Participants() {
		if p.ID() != g.HostID() {
			joinerParticipant = p.ID()
		}
	}
	g, err = svc.TransferHost(context.Background(), g.ID(), g.HostID(), joinerParticipant)
	if err != nil {
		t.Fatalf("TransferHost: %v", err)
	}
	if g.HostID() != joinerParticipant {
		t.Fatalf("expected the joiner to be the new host")
	}
}

func TestSetLobbyLocked_Service(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")
	joinerID := mustTestUser(t, deps, "joiner")

	g, _, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	g, err = svc.SetLobbyLocked(context.Background(), g.ID(), g.HostID(), true)
	if err != nil {
		t.Fatalf("SetLobbyLocked: %v", err)
	}
	if !g.Locked() {
		t.Fatalf("expected the lobby to be locked")
	}
	// Locking a lobby is orthogonal to its Visibility - this one is still
	// PRIVATE by default, so JoinByID still 403s on that, not on the lock.
	if _, err := svc.JoinByID(context.Background(), g.ID(), joinerID); !errors.Is(err, game.ErrLobbyPrivate) {
		t.Fatalf("err = %v, want ErrLobbyPrivate", err)
	}
}

func TestJoinByID_RejectsPrivateLobby(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")
	joinerID := mustTestUser(t, deps, "joiner")

	g, _, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	if _, err := svc.JoinByID(context.Background(), g.ID(), joinerID); !errors.Is(err, game.ErrLobbyPrivate) {
		t.Fatalf("err = %v, want ErrLobbyPrivate", err)
	}
}

func TestJoinByID_PublicLobby_Success(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")
	joinerID := mustTestUser(t, deps, "joiner")

	input := gauntletInput()
	input.Visibility = enums.Public
	g, _, err := svc.CreateGame(context.Background(), hostID, input)
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	g, err = svc.JoinByID(context.Background(), g.ID(), joinerID)
	if err != nil {
		t.Fatalf("JoinByID: %v", err)
	}
	if len(g.Participants()) != 2 {
		t.Fatalf("expected 2 participants after JoinByID")
	}
}

func TestListPublicLobbies_OnlyReturnsBrowsableLobbies(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")
	privateHostID := mustTestUser(t, deps, "private-host")

	publicInput := gauntletInput()
	publicInput.Visibility = enums.Public
	pub, _, err := svc.CreateGame(context.Background(), hostID, publicInput)
	if err != nil {
		t.Fatalf("CreateGame (public): %v", err)
	}
	if _, _, err := svc.CreateGame(context.Background(), privateHostID, gauntletInput()); err != nil {
		t.Fatalf("CreateGame (private): %v", err)
	}

	listings, err := svc.ListPublicLobbies(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListPublicLobbies: %v", err)
	}
	if len(listings) != 1 {
		t.Fatalf("expected exactly 1 public lobby listed, got %d", len(listings))
	}
	if listings[0].GameID != pub.ID() {
		t.Fatalf("expected the listed lobby to be the public one")
	}
	if listings[0].HostDisplayName == "" {
		t.Fatalf("expected a host display name on the listing")
	}
}

func TestListPublicLobbies_ExcludesLockedLobbies(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")

	input := gauntletInput()
	input.Visibility = enums.Public
	g, _, err := svc.CreateGame(context.Background(), hostID, input)
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	if _, err := svc.SetLobbyLocked(context.Background(), g.ID(), g.HostID(), true); err != nil {
		t.Fatalf("SetLobbyLocked: %v", err)
	}
	listings, err := svc.ListPublicLobbies(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListPublicLobbies: %v", err)
	}
	if len(listings) != 0 {
		t.Fatalf("expected a locked public lobby to be excluded, got %d", len(listings))
	}
}

func TestPreviewByCode_WorksForPrivateLobby(t *testing.T) {
	svc, deps := newTestGameService(t)
	hostID := mustTestUser(t, deps, "host")

	_, code, err := svc.CreateGame(context.Background(), hostID, gauntletInput())
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	listing, err := svc.PreviewByCode(context.Background(), code)
	if err != nil {
		t.Fatalf("PreviewByCode: %v", err)
	}
	if listing.PlayerCount != 1 {
		t.Fatalf("expected the preview to report 1 seated player, got %d", listing.PlayerCount)
	}
}
