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
	PoolFilter          game.PoolFilter
	Mangas              []enums.Manga
	TeamSize            int
	VotingWindowSeconds int
	Mode                enums.GameModeKind
	AbilitySource       enums.AbilitySource
	AllowBots           bool
	// Visibility, VotingWindowSeconds, PoolFilter are all optional: a zero
	// Visibility defaults to enums.Private, a zero VotingWindowSeconds
	// defaults to VotingPolicy.Window, and a zero PoolFilter means "no
	// restriction".
	Visibility enums.LobbyVisibility
}

// ConfigUpdateInput is EditLobbyConfig's whole-replacement input - the same
// shape as CreateGameInput, since an edit sends back the full config the
// client has in hand rather than a sparse patch.
type ConfigUpdateInput = CreateGameInput

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
	// timers holds the pending timer for each Game currently in
	// ASSIGNING/VOTING/TIEBREAK, keyed by GameID - either the reveal-delay
	// timer (see openVotingAfterReveal) or the voting-window/revote timer
	// (see scheduleVotingTimer). The two never coexist for the same GameID:
	// the reveal timer fires, opens voting, and only then does the voting
	// timer take its place in this same map - so cancelTimer's single
	// Stop+delete already covers whichever one is currently pending.
	timers map[game.GameID]Timer
	// revealEnds holds the wall-clock deadline of the in-flight reveal
	// animation for each Game currently in ASSIGNING with a pending reveal
	// timer, keyed by GameID - read by RevealEndsAt so a client
	// joining/reconnecting mid-reveal can resume the countdown instead of
	// restarting it. Process memory only, like timers itself: a backend
	// restart mid-reveal loses this (and its timer), the same known gap
	// scheduleVotingTimer already has.
	revealEnds map[game.GameID]time.Time
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
		votingCfg:  votingCfg,
		locks:      newGameLocks(),
		timers:     make(map[game.GameID]Timer),
		revealEnds: make(map[game.GameID]time.Time),
	}
}

// --- Creation / membership ---

// CreateGame builds a new Game in the LOBBY state, hosted by hostUserID, and
// indexes it under a freshly generated join code.
func (s *GameService) CreateGame(ctx context.Context, hostUserID user.UserID, input CreateGameInput) (*game.Game, string, error) {
	cfg, err := s.buildConfig(input)
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
	host.SetAvatar(hostUser.AvatarThumbKey(), hostUser.GooglePicture())

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

// buildConfig turns a CreateGameInput/ConfigUpdateInput into a validated
// game.Config, substituting service-level defaults for the zero-valued
// optional fields.
func (s *GameService) buildConfig(input CreateGameInput) (game.Config, error) {
	visibility := input.Visibility
	if !visibility.IsValid() {
		visibility = enums.Private
	}
	votingWindowSeconds := input.VotingWindowSeconds
	if votingWindowSeconds == 0 {
		votingWindowSeconds = int(s.votingCfg.Window / time.Second)
	}
	return game.NewConfig(
		input.Mode, input.Mangas, input.AbilitySource, input.TeamSize, input.AllowBots,
		visibility, votingWindowSeconds, input.PoolFilter,
	)
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
		return s.joinLocked(ctx, g, userID)
	})
}

// JoinByID seats userID in the Game identified by gameID, the public lobby
// browser's join path. Unlike JoinByCode, it's reachable without knowing a
// secret, so it's rejected outright for anything but a PUBLIC lobby.
func (s *GameService) JoinByID(ctx context.Context, gameID game.GameID, userID user.UserID) (*game.Game, error) {
	return s.withGame(ctx, gameID, func(g *game.Game) error {
		if g.Config().Visibility() != enums.Public {
			return game.ErrLobbyPrivate
		}
		return s.joinLocked(ctx, g, userID)
	})
}

// joinLocked is the shared body of JoinByCode/JoinByID. Callers must
// already hold g's lock (i.e. call this from inside withGame's fn).
func (s *GameService) joinLocked(ctx context.Context, g *game.Game, userID user.UserID) error {
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
	p.SetAvatar(u.AvatarThumbKey(), u.GooglePicture())
	return g.Join(p)
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

// SwitchTeam moves targetID onto teamID. Any participant may move
// themselves; moving someone else requires being the host.
func (s *GameService) SwitchTeam(ctx context.Context, gameID game.GameID, callerID, targetID game.ParticipantID, teamID game.TeamID) (*game.Game, error) {
	return s.withGame(ctx, gameID, func(g *game.Game) error {
		return g.SwitchTeam(callerID, targetID, teamID)
	})
}

// KickParticipant removes targetID from the Game entirely. Host-only.
func (s *GameService) KickParticipant(ctx context.Context, gameID game.GameID, callerID, targetID game.ParticipantID) (*game.Game, error) {
	return s.withGame(ctx, gameID, func(g *game.Game) error {
		return g.Kick(callerID, targetID, s.rng)
	})
}

// TransferHost hands the host role to targetID. Host-only.
func (s *GameService) TransferHost(ctx context.Context, gameID game.GameID, callerID, targetID game.ParticipantID) (*game.Game, error) {
	return s.withGame(ctx, gameID, func(g *game.Game) error {
		return g.TransferHost(callerID, targetID)
	})
}

// SetLobbyLocked toggles whether new humans may join the lobby. Host-only.
func (s *GameService) SetLobbyLocked(ctx context.Context, gameID game.GameID, callerID game.ParticipantID, locked bool) (*game.Game, error) {
	return s.withGame(ctx, gameID, func(g *game.Game) error {
		return g.SetLocked(callerID, locked)
	})
}

// EditLobbyConfig replaces the lobby's whole Config while it is still in
// LOBBY. Host-only. When the new mode differs from the current one, fresh
// teams and the matching stage list are built first (mirroring CreateGame);
// otherwise the existing teams/stages are reused as-is.
func (s *GameService) EditLobbyConfig(ctx context.Context, gameID game.GameID, callerID game.ParticipantID, input ConfigUpdateInput) (*game.Game, error) {
	next, err := s.buildConfig(input)
	if err != nil {
		return nil, err
	}
	return s.withGame(ctx, gameID, func(g *game.Game) error {
		if callerID != g.HostID() {
			return game.ErrNotHost
		}
		teams := g.Teams()
		stages := g.Stages()
		if next.Mode() != g.Config().Mode() {
			teams, err = s.buildTeams(next.Mode())
			if err != nil {
				return err
			}
		}
		if next.Mode() != g.Config().Mode() || !mangasEqual(next.Mangas(), g.Config().Mangas()) {
			stages, err = s.loadStages(ctx, next)
			if err != nil {
				return err
			}
		}
		return g.Reconfigure(callerID, next, teams, stages)
	})
}

func mangasEqual(a, b []enums.Manga) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
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
		// Checked before g.Start so a too-small filtered pool leaves the
		// Game untouched in LOBBY, instead of stranding it in ASSIGNING
		// with no way back (Game has no "un-start").
		if callerID != g.HostID() {
			return game.ErrNotHost
		}
		if err := s.checkPoolSufficiency(ctx, g); err != nil {
			return err
		}
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

// checkPoolSufficiency reports whether g's filtered power pool has enough
// unique Stands/DevilFruits to seat every currently-seated team member once
// for each selected manga. It checks actual occupancy (the largest current
// Team.Size()), not Config.TeamSize()'s capacity ceiling - a lobby that
// never fills up must not be blocked from starting by a pool sized for its
// maximum, not its actual, headcount. Every round draws an identically-
// sized pool (AvailablePowers is rebuilt from the same catalog each time -
// see beginRound), so a check once at Start provably rules out
// ErrPowerPoolExhausted mid-match rather than letting it surface
// unpredictably on some later round.
func (s *GameService) checkPoolSufficiency(ctx context.Context, g *game.Game) error {
	stands, err := s.powers.Stands(ctx)
	if err != nil {
		return err
	}
	fruits, err := s.powers.DevilFruits(ctx)
	if err != nil {
		return err
	}
	stands, fruits = g.Config().PoolFilter().Apply(stands, fruits)
	needed := 0
	for _, t := range g.Teams() {
		if t.Size() > needed {
			needed = t.Size()
		}
	}
	if g.Config().HasManga(enums.Jojo) && len(stands) < needed {
		return game.ErrPoolTooSmall
	}
	if g.Config().HasManga(enums.OnePiece) && len(fruits) < needed {
		return game.ErrPoolTooSmall
	}
	return nil
}

// beginRound resolves fresh Loadouts (when this is the first round, or the
// mode reassigns every round - see game.IGameMode.ReassignsEachRound) and,
// only when it just did, delays OpenVoting until the reveal animation every
// client plays has had time to finish (see scheduleRevealDelay) - the Game
// stays in ASSIGNING for that stretch. Rounds that don't reassign (Gauntlet
// rounds after the first) have nothing new to reveal, so voting opens
// immediately as before. Callers must already hold g's lock (i.e. call this
// from inside withGame's fn).
func (s *GameService) beginRound(ctx context.Context, g *game.Game) error {
	assigned := len(g.Rounds()) == 0 || g.Mode().ReassignsEachRound()
	if assigned {
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
		stands, fruits = g.Config().PoolFilter().Apply(stands, fruits)

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

	if assigned {
		s.scheduleRevealDelay(g)
		return nil
	}
	if err := g.OpenVoting(s.rng); err != nil {
		return err
	}
	s.scheduleVotingTimer(g)
	return nil
}

// scheduleRevealDelay holds g in ASSIGNING for game.RevealDuration(mangas) -
// every client computes that same duration locally (loadout-reveal.ts) to
// pace its reveal overlay, so this is what keeps "voting opens" from ever
// racing ahead of "everyone has seen their loadout". Once the delay elapses,
// openVotingAfterReveal does what beginRound used to do immediately:
// OpenVoting + scheduleVotingTimer. Uses the same s.timers map as the voting
// timer (see its field doc) and records the deadline in s.revealEnds for
// RevealEndsAt to serve to (re)connecting clients.
func (s *GameService) scheduleRevealDelay(g *game.Game) {
	id := g.ID()
	d := game.RevealDuration(g.Config().Mangas())

	s.timersMu.Lock()
	s.revealEnds[id] = s.clock.Now().Add(d)
	s.timersMu.Unlock()

	timer := s.clock.AfterFunc(d, func() {
		s.openVotingAfterReveal(context.Background(), id)
	})
	s.timersMu.Lock()
	s.timers[id] = timer
	s.timersMu.Unlock()
}

// openVotingAfterReveal is scheduleRevealDelay's callback: it re-acquires
// g's lock via withGame and only then opens voting, so it never races a
// concurrent AbortGame/CastVote against the same Game.
func (s *GameService) openVotingAfterReveal(ctx context.Context, id game.GameID) {
	s.timersMu.Lock()
	delete(s.revealEnds, id)
	s.timersMu.Unlock()

	_, err := s.withGame(ctx, id, func(g *game.Game) error {
		if err := g.OpenVoting(s.rng); err != nil {
			return err
		}
		s.scheduleVotingTimer(g)
		return nil
	})
	if err != nil {
		switch {
		case errors.Is(err, ports.ErrGameNotFound), errors.Is(err, game.ErrInvalidStateTransition):
			// Benign: the Game was aborted/removed, or someone else already
			// moved it out of ASSIGNING, before this timer fired.
		default:
			log.Printf("opening voting after reveal for game %s: %v", id, err)
		}
	}
}

// RevealEndsAt reports id's in-flight reveal deadline, if any - used to
// serve a (re)connecting client the remaining time instead of restarting
// the reveal from zero. The bool is false once the reveal has ended (or
// never started).
func (s *GameService) RevealEndsAt(id game.GameID) (time.Time, bool) {
	s.timersMu.Lock()
	defer s.timersMu.Unlock()
	t, ok := s.revealEnds[id]
	return t, ok
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
			s.scheduleVotingTimer(g)
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

// LobbyListing is the summary shape exposed by the public lobby browser and
// by PreviewByCode - deliberately roster-free and loadout-free, so a
// non-participant can never learn who's in a lobby or what they'll play as,
// only whether it's worth joining.
type LobbyListing struct {
	HostDisplayName     string
	Mangas              []enums.Manga
	PlayerCount         int
	MaxPlayers          int
	VotingWindowSeconds int
	GameID              game.GameID
	Mode                enums.GameModeKind
	AbilitySource       enums.AbilitySource
	AllowBots           bool
	Visibility          enums.LobbyVisibility
	Locked              bool
}

func newLobbyListing(g *game.Game) LobbyListing {
	cfg := g.Config()
	maxPlayers := cfg.TeamSize()
	if cfg.Mode() == enums.Versus {
		maxPlayers = cfg.TeamSize() * game.VersusTeamCount
	}
	hostName := ""
	if host, ok := g.Participant(g.HostID()); ok {
		hostName = host.DisplayName()
	}
	return LobbyListing{
		GameID:              g.ID(),
		Mode:                cfg.Mode(),
		HostDisplayName:     hostName,
		PlayerCount:         len(g.Participants()),
		MaxPlayers:          maxPlayers,
		Mangas:              cfg.Mangas(),
		AbilitySource:       cfg.AbilitySource(),
		AllowBots:           cfg.AllowBots(),
		Visibility:          cfg.Visibility(),
		VotingWindowSeconds: cfg.VotingWindowSeconds(),
		Locked:              g.Locked(),
	}
}

// ListPublicLobbies returns up to limit lobbies currently joinable through
// the public browser - see game.Game.IsPubliclyJoinable.
func (s *GameService) ListPublicLobbies(ctx context.Context, limit int) ([]LobbyListing, error) {
	games, err := s.store.ListPublic(ctx, limit)
	if err != nil {
		return nil, err
	}
	out := make([]LobbyListing, len(games))
	for i, g := range games {
		out[i] = newLobbyListing(g)
	}
	return out, nil
}

// PreviewByCode returns a LobbyListing for the Game indexed under code,
// without requiring the caller to already be a participant - the code
// itself is the credential, so this works for PRIVATE lobbies too, unlike
// GetGameByCode (which 403s a non-participant; see game_endpoints.go).
func (s *GameService) PreviewByCode(ctx context.Context, code string) (LobbyListing, error) {
	g, err := s.store.GetByCode(ctx, code)
	if err != nil {
		return LobbyListing{}, err
	}
	return newLobbyListing(g), nil
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
	window := time.Duration(g.Config().VotingWindowSeconds()) * time.Second
	revealMs := game.RevealDuration(g.Config().Mangas())
	for _, e := range g.PullEvents() {
		s.hub.Publish(GameEvent{GameID: g.ID(), Name: e.Name(), Event: e, VotingWindow: window, RevealMs: revealMs})
	}
}

// scheduleVotingTimer starts g's voting-window timer using its own
// per-lobby Config.VotingWindowSeconds() (host-configurable at creation/
// edit time) rather than the service-wide VotingPolicy.Window default,
// which now only seeds new lobbies that don't specify one - see
// buildConfig.
func (s *GameService) scheduleVotingTimer(g *game.Game) {
	id := g.ID()
	window := time.Duration(g.Config().VotingWindowSeconds()) * time.Second
	timer := s.clock.AfterFunc(window, func() {
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
	delete(s.revealEnds, id)
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
