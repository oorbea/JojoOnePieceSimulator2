package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// codeAlphabet excludes visually ambiguous characters (0/O, 1/I) so a join
// code is easy to read aloud or type back in. codeLength/maxCodeAttempts
// bound CreateGame's retry loop against a collision, mirroring
// AuthService.uniqueUsername's suffix-retry pattern.
const (
	codeAlphabet    = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	codeLength      = 6
	maxCodeAttempts = 5
)

// Default Team names/colors CreateGame assigns - purely cosmetic, the host
// cannot rename or recolor them in this tanda.
const (
	squadTeamName  = "Squad"
	squadTeamColor = 0x2E86DE
	teamAName      = "Team A"
	teamAColor     = 0xE74C3C
	teamBName      = "Team B"
	teamBColor     = 0x3498DB
)

var (
	// ErrAlreadyInGame is returned when a user tries to join a Game they are
	// already seated in as a human participant.
	ErrAlreadyInGame = errors.New("user is already in this game")

	// ErrNotABot is returned when RemoveBot targets a participant that is
	// not a bot - LeaveGame removes a human participant instead.
	ErrNotABot = errors.New("participant is not a bot")

	// ErrCodeGenerationFailed is returned when CreateGame could not find an
	// unused join code within maxCodeAttempts tries.
	ErrCodeGenerationFailed = errors.New("could not generate a unique game code")
)

// CreateGameInput carries every field needed to set up a new Game's Config,
// so CreateGame takes one argument instead of a long positional list -
// mirrors StandInput/DevilFruitInput.
type CreateGameInput struct {
	Mode          enums.GameModeKind
	Mangas        []enums.Manga
	AbilitySource enums.AbilitySource
	TeamSize      int
	AllowBots     bool
}

// VotingPolicy configures GameService's voting window.
type VotingPolicy struct {
	// Window is how long a round's Ballot stays open before GameService
	// force-closes it (see game.Game.VotingComplete for the early-close
	// path). Applies identically to the single revote window a tie opens.
	Window time.Duration
}

// GameService coordinates every Gauntlet/Versus use case against the
// game.Game aggregate: creating and joining lobbies, starting a match,
// assigning Loadouts, running the vote/tiebreak/resolve cycle (including
// the 30s window the domain itself knows nothing about - see
// game.Game.VotingComplete), and finalizing a finished or aborted Game.
type GameService struct {
	store    ports.IGameStore
	gameIDs  ports.IIdGenerator[game.GameID]
	partIDs  ports.IIdGenerator[game.ParticipantID]
	teamIDs  ports.IIdGenerator[game.TeamID]
	users    ports.IUserRepository
	stages   ports.IStageCatalog
	powers   ports.IGamePowerPool
	weights  ports.IAssignmentWeights
	tiebreak ports.ITiebreaker
	// history may be nil - tests exercise GameService without a real
	// ports.IGameHistory adapter, and finalizeLocked tolerates that by
	// simply skipping recording. Production always passes one
	// (repositories.NewGameHistory).
	history   ports.IGameHistory
	rng       game.RandomSource
	hub       *GameEventHub
	clock     Clock
	votingCfg VotingPolicy

	// locks serializes every mutation of a given Game by its GameID, so
	// concurrent requests (a vote, a disconnect, a timer firing) against the
	// same lobby never race on the in-memory aggregate.
	locks *gameLocks

	timersMu sync.Mutex
	// timers holds the pending voting-window (or revote-window) timer for
	// each Game currently in VOTING/TIEBREAK, keyed by GameID.
	timers map[game.GameID]Timer
}

// NewGameService builds a GameService. history may be nil until an
// ports.IGameHistory adapter exists.
func NewGameService(
	store ports.IGameStore,
	gameIDs ports.IIdGenerator[game.GameID],
	partIDs ports.IIdGenerator[game.ParticipantID],
	teamIDs ports.IIdGenerator[game.TeamID],
	users ports.IUserRepository,
	stages ports.IStageCatalog,
	powerPool ports.IGamePowerPool,
	weights ports.IAssignmentWeights,
	tiebreak ports.ITiebreaker,
	history ports.IGameHistory,
	rng game.RandomSource,
	hub *GameEventHub,
	clock Clock,
	votingCfg VotingPolicy,
) *GameService {
	return &GameService{
		store: store, gameIDs: gameIDs, partIDs: partIDs, teamIDs: teamIDs,
		users: users, stages: stages, powers: powerPool, weights: weights,
		tiebreak: tiebreak, history: history, rng: rng, hub: hub, clock: clock,
		votingCfg: votingCfg,
		locks:     newGameLocks(),
		timers:    make(map[game.GameID]Timer),
	}
}

// --- Creation / membership ---

// CreateGame builds a new Game in the LOBBY state, hosted by hostUserID, and
// indexes it under a freshly generated join code.
func (s *GameService) CreateGame(ctx context.Context, hostUserID user.UserID, input CreateGameInput) (*game.Game, string, error) {
	cfg, err := game.NewConfig(input.Mode, input.Mangas, input.AbilitySource, input.TeamSize, input.AllowBots)
	if err != nil {
		return nil, "", err
	}

	hostUser, err := s.users.FindByID(ctx, hostUserID)
	if err != nil {
		return nil, "", err
	}

	teams, err := s.buildTeams(cfg.Mode())
	if err != nil {
		return nil, "", err
	}

	stages, err := s.loadStages(ctx, cfg)
	if err != nil {
		return nil, "", err
	}

	host, err := game.NewHumanParticipant(s.partIDs.NewID(), hostUserID, hostUser.Username(), teams[0].ID())
	if err != nil {
		return nil, "", err
	}

	g, err := game.NewGame(s.gameIDs.NewID(), cfg, host, teams, stages)
	if err != nil {
		return nil, "", err
	}

	for attempt := 0; attempt < maxCodeAttempts; attempt++ {
		code := s.generateCode()
		if err := s.store.Create(ctx, code, g); err != nil {
			if errors.Is(err, ports.ErrGameCodeTaken) {
				continue
			}
			return nil, "", err
		}
		return g, code, nil
	}
	return nil, "", ErrCodeGenerationFailed
}

func (s *GameService) buildTeams(mode enums.GameModeKind) ([]*game.Team, error) {
	switch mode {
	case enums.Gauntlet:
		t, err := game.NewTeam(s.teamIDs.NewID(), squadTeamName, squadTeamColor)
		if err != nil {
			return nil, err
		}
		return []*game.Team{t}, nil
	case enums.Versus:
		a, err := game.NewTeam(s.teamIDs.NewID(), teamAName, teamAColor)
		if err != nil {
			return nil, err
		}
		b, err := game.NewTeam(s.teamIDs.NewID(), teamBName, teamBColor)
		if err != nil {
			return nil, err
		}
		return []*game.Team{a, b}, nil
	default:
		return nil, enums.ErrInvalidGameModeKind
	}
}

// loadStages resolves cfg's selected mangas against the stage catalog.
// Gauntlet needs the full, pre-ordered round list (Interleave); Versus just
// needs the pool game.IGameMode.StageFor draws a random Stage from each
// round, so every selected manga's Stages are simply concatenated.
func (s *GameService) loadStages(ctx context.Context, cfg game.Config) ([]game.Stage, error) {
	byManga := make(map[enums.Manga][]game.Stage, len(cfg.Mangas()))
	for _, m := range cfg.Mangas() {
		stages, err := s.stages.Stages(ctx, m)
		if err != nil {
			return nil, err
		}
		byManga[m] = stages
	}

	if cfg.Mode() == enums.Gauntlet {
		return game.Interleave(byManga), nil
	}

	var pool []game.Stage
	for _, m := range cfg.Mangas() {
		pool = append(pool, byManga[m]...)
	}
	return pool, nil
}

func (s *GameService) generateCode() string {
	b := make([]byte, codeLength)
	for i := range b {
		b[i] = codeAlphabet[s.rng.IntN(len(codeAlphabet))]
	}
	return string(b)
}

// JoinByCode seats userID as a new human participant in the Game indexed
// under code, auto-balancing to the emptiest Team in Versus.
func (s *GameService) JoinByCode(ctx context.Context, code string, userID user.UserID) (*game.Game, error) {
	g, err := s.store.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	return s.withGame(ctx, g.ID(), func(g *game.Game) error {
		for _, p := range g.Participants() {
			if p.UserID() != nil && *p.UserID() == userID {
				return ErrAlreadyInGame
			}
		}
		u, err := s.users.FindByID(ctx, userID)
		if err != nil {
			return err
		}
		p, err := game.NewHumanParticipant(s.partIDs.NewID(), userID, u.Username(), s.pickTeam(g))
		if err != nil {
			return err
		}
		return g.Join(p)
	})
}

// pickTeam returns the Team a new joiner should land on: the only Team in
// Gauntlet, or whichever Versus Team currently has fewer members (ties
// favor the first Team, deterministically).
func (s *GameService) pickTeam(g *game.Game) game.TeamID {
	teams := g.Teams()
	best := teams[0]
	for _, t := range teams[1:] {
		if t.Size() < best.Size() {
			best = t
		}
	}
	return best.ID()
}

// LeaveGame removes participantID from the Game entirely.
func (s *GameService) LeaveGame(ctx context.Context, gameID game.GameID, participantID game.ParticipantID) (*game.Game, error) {
	return s.withGame(ctx, gameID, func(g *game.Game) error {
		return g.Leave(participantID, s.rng)
	})
}

// AddBot seats a new bot participant on teamID. Host-only.
func (s *GameService) AddBot(ctx context.Context, gameID game.GameID, callerID game.ParticipantID, teamID game.TeamID) (*game.Game, error) {
	return s.withGame(ctx, gameID, func(g *game.Game) error {
		if callerID != g.HostID() {
			return game.ErrNotHost
		}
		name := fmt.Sprintf("Bot %d", s.countBots(g)+1)
		p, err := game.NewBotParticipant(s.partIDs.NewID(), name, teamID)
		if err != nil {
			return err
		}
		return g.AddBot(p)
	})
}

// RemoveBot removes botID, which must in fact be a bot. Host-only.
func (s *GameService) RemoveBot(ctx context.Context, gameID game.GameID, callerID game.ParticipantID, botID game.ParticipantID) (*game.Game, error) {
	return s.withGame(ctx, gameID, func(g *game.Game) error {
		if callerID != g.HostID() {
			return game.ErrNotHost
		}
		p, ok := g.Participant(botID)
		if !ok {
			return game.ErrParticipantNotFound
		}
		if !p.IsBot() {
			return ErrNotABot
		}
		return g.Leave(botID, s.rng)
	})
}

func (s *GameService) countBots(g *game.Game) int {
	n := 0
	for _, p := range g.Participants() {
		if p.IsBot() {
			n++
		}
	}
	return n
}

// --- Lifecycle ---

// StartGame moves the Game from LOBBY into its first round: it validates
// team sizes (via game.Game.Start), draws the first round's Loadouts, and
// opens voting. Host-only.
func (s *GameService) StartGame(ctx context.Context, gameID game.GameID, callerID game.ParticipantID) (*game.Game, error) {
	return s.withGame(ctx, gameID, func(g *game.Game) error {
		if err := g.Start(callerID); err != nil {
			return err
		}
		return s.beginRound(ctx, g)
	})
}

// AbortGame cancels the Game. Host-only.
func (s *GameService) AbortGame(ctx context.Context, gameID game.GameID, callerID game.ParticipantID) (*game.Game, error) {
	return s.withGame(ctx, gameID, func(g *game.Game) error {
		return g.Abort(callerID)
	})
}

// beginRound resolves fresh Loadouts (when this is the first round, or the
// mode reassigns every round - see game.IGameMode.ReassignsEachRound), opens
// the round's Ballot, and starts its voting-window timer. Callers must
// already hold g's lock (i.e. call this from inside withGame's fn).
func (s *GameService) beginRound(ctx context.Context, g *game.Game) error {
	if len(g.Rounds()) == 0 || g.Mode().ReassignsEachRound() {
		weights, err := s.weights.Load(ctx)
		if err != nil {
			return err
		}
		stands, err := s.powers.Stands(ctx)
		if err != nil {
			return err
		}
		fruits, err := s.powers.DevilFruits(ctx)
		if err != nil {
			return err
		}

		// Each Team gets its own AvailablePowers built from the same
		// catalog snapshot - drawing on one Team's pool must never affect
		// another Team's pool (see game.AvailablePowers).
		poolByTeam := make(map[game.TeamID]*game.AvailablePowers, len(g.Teams()))
		for _, t := range g.Teams() {
			poolByTeam[t.ID()] = game.NewAvailablePowers(stands, fruits)
		}

		builder := game.NewLoadoutBuilder(g.Config().Mangas(), weights, s.rng)
		if err := g.AssignLoadouts(builder, poolByTeam); err != nil {
			return err
		}
	}

	if err := g.OpenVoting(s.rng); err != nil {
		return err
	}
	s.scheduleVotingTimer(g.ID())
	return nil
}

// --- Voting ---

// CastVote records participantID's vote and force-closes the window early
// once every connected human has voted (see game.Game.VotingComplete).
func (s *GameService) CastVote(ctx context.Context, gameID game.GameID, participantID game.ParticipantID, option game.OptionID) (*game.Game, error) {
	return s.withGame(ctx, gameID, func(g *game.Game) error {
		if err := g.CastVote(participantID, option); err != nil {
			return err
		}
		if g.VotingComplete() {
			return s.closeVoting(ctx, g)
		}
		return nil
	})
}

// CloseVotingWindow force-closes the current round's voting window - called
// by the voting-window timer when it expires, or invocable directly.
func (s *GameService) CloseVotingWindow(ctx context.Context, gameID game.GameID) (*game.Game, error) {
	return s.withGame(ctx, gameID, func(g *game.Game) error {
		return s.closeVoting(ctx, g)
	})
}

// closeVoting tallies the current round. A clear winner resolves the round
// (see game.Game.CloseVoting/resolveRound); a first tie opens a single
// revote window; a second tie (the revote window itself came back tied, or
// nobody voted at all) is handed to ports.ITiebreaker. Callers must already
// hold g's lock.
func (s *GameService) closeVoting(ctx context.Context, g *game.Game) error {
	s.cancelTimer(g.ID())

	wasRevote := g.State() == enums.Tiebreak
	tied, err := g.CloseVoting()
	if err != nil {
		return err
	}

	if tied {
		if !wasRevote {
			// First tie for this round: game.Game already opened the
			// revote window (state is now TIEBREAK) - give it its own
			// timer and wait for that vote instead.
			s.scheduleVotingTimer(g.ID())
			return nil
		}

		rounds := g.Rounds()
		round := rounds[len(rounds)-1]
		options := round.Ballot.Options()
		strOptions := make([]string, len(options))
		for i, o := range options {
			strOptions[i] = string(o)
		}

		winner, err := s.tiebreak.Break(ctx, strOptions)
		if err != nil {
			return err
		}
		if err := g.ResolveTiebreak(game.OptionID(winner)); err != nil {
			return err
		}
	}

	if g.State() == enums.Assigning {
		return s.beginRound(ctx, g)
	}
	return nil
}

// --- Presence ---

// Disconnect marks participantID unreachable without removing their seat.
// If that was the last vote the current round was waiting on, the window
// closes immediately - a disconnected participant counts as a null vote,
// never blocking the round.
func (s *GameService) Disconnect(ctx context.Context, gameID game.GameID, participantID game.ParticipantID) (*game.Game, error) {
	return s.withGame(ctx, gameID, func(g *game.Game) error {
		if err := g.Disconnect(participantID, s.rng); err != nil {
			return err
		}
		if (g.State() == enums.Voting || g.State() == enums.Tiebreak) && g.VotingComplete() {
			return s.closeVoting(ctx, g)
		}
		return nil
	})
}

// Reconnect marks a previously disconnected participant reachable again.
func (s *GameService) Reconnect(ctx context.Context, gameID game.GameID, participantID game.ParticipantID) (*game.Game, error) {
	return s.withGame(ctx, gameID, func(g *game.Game) error {
		return g.Reconnect(participantID)
	})
}

// --- Reads ---

// GetGame returns the live Game identified by id.
func (s *GameService) GetGame(ctx context.Context, id game.GameID) (*game.Game, error) {
	mu := s.locks.get(id)
	mu.Lock()
	defer mu.Unlock()
	return s.store.Get(ctx, id)
}

// GetGameByCode returns the live Game currently indexed under code.
func (s *GameService) GetGameByCode(ctx context.Context, code string) (*game.Game, error) {
	g, err := s.store.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	return s.GetGame(ctx, g.ID())
}

// GameCode returns the join code currently indexed for id.
func (s *GameService) GameCode(ctx context.Context, id game.GameID) (string, error) {
	return s.store.Code(ctx, id)
}

// --- Internal plumbing ---

// withGame serializes access to the Game identified by id: it locks that
// Game's mutex, loads it, runs fn against it, publishes whatever
// DomainEvents fn produced (even on error - a rejected call can still have
// mutated the aggregate up to the point it failed), and either finalizes
// the Game (if fn left it FINISHED/ABORTED) or persists it back to the
// store.
func (s *GameService) withGame(ctx context.Context, id game.GameID, fn func(g *game.Game) error) (*game.Game, error) {
	mu := s.locks.get(id)
	mu.Lock()
	defer mu.Unlock()

	g, err := s.store.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := fn(g); err != nil {
		s.publish(g)
		return nil, err
	}
	s.publish(g)

	if g.State() == enums.Finished || g.State() == enums.Aborted {
		s.finalizeLocked(ctx, g)
		return g, nil
	}
	if err := s.store.Save(ctx, g); err != nil {
		return nil, err
	}
	return g, nil
}

// finalizeLocked records g's outcome (best-effort, tolerating a nil
// history port) and removes it from the store. Callers must already hold
// g's lock.
func (s *GameService) finalizeLocked(ctx context.Context, g *game.Game) {
	if result, err := g.Result(); err != nil {
		log.Printf("computing result for finished game %s: %v", g.ID(), err)
	} else if s.history != nil {
		if err := s.history.Record(ctx, result); err != nil {
			log.Printf("recording history for game %s: %v", g.ID(), err)
		}
	}

	if err := s.store.Delete(ctx, g.ID()); err != nil {
		log.Printf("deleting finished game %s: %v", g.ID(), err)
	}
	s.cancelTimer(g.ID())
	s.locks.delete(g.ID())
}

// publish drains every DomainEvent g has accumulated since the last call
// and fans them out over hub.
func (s *GameService) publish(g *game.Game) {
	for _, e := range g.PullEvents() {
		s.hub.Publish(GameEvent{GameID: g.ID(), Name: e.Name(), Event: e})
	}
}

func (s *GameService) scheduleVotingTimer(id game.GameID) {
	timer := s.clock.AfterFunc(s.votingCfg.Window, func() {
		ctx := context.Background()
		if _, err := s.CloseVotingWindow(ctx, id); err != nil {
			switch {
			case errors.Is(err, ports.ErrGameNotFound), errors.Is(err, game.ErrVotingClosed):
				// Benign: the Game was already finished/aborted/removed
				// before this timer fired.
			default:
				log.Printf("closing voting window for game %s: %v", id, err)
			}
		}
	})
	s.timersMu.Lock()
	s.timers[id] = timer
	s.timersMu.Unlock()
}

func (s *GameService) cancelTimer(id game.GameID) {
	s.timersMu.Lock()
	defer s.timersMu.Unlock()
	if t, ok := s.timers[id]; ok {
		t.Stop()
		delete(s.timers, id)
	}
}

// gameLocks hands out one *sync.Mutex per GameID, created on first use, so
// GameService can serialize access to each Game independently instead of
// behind one global lock.
type gameLocks struct {
	mu sync.Mutex
	m  map[game.GameID]*sync.Mutex
}

func newGameLocks() *gameLocks {
	return &gameLocks{m: make(map[game.GameID]*sync.Mutex)}
}

func (l *gameLocks) get(id game.GameID) *sync.Mutex {
	l.mu.Lock()
	defer l.mu.Unlock()
	mu, ok := l.m[id]
	if !ok {
		mu = &sync.Mutex{}
		l.m[id] = mu
	}
	return mu
}

func (l *gameLocks) delete(id game.GameID) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.m, id)
}
