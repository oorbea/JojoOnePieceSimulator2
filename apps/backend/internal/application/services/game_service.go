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

	// ErrGameNotOver is returned by Rematch when the source Game has not
	// actually reached FINISHED/ABORTED yet - a rematch of a game still in
	// progress is meaningless.
	ErrGameNotOver = errors.New("game is not over yet")
)

// CreateGameInput carries every field needed to set up a new Game's Config,
// so CreateGame takes one argument instead of a long positional list -
// mirrors StandInput/DevilFruitInput.
type CreateGameInput struct {
	PoolFilter game.PoolFilter
	// StageMangas and PowerMangas are independent - which manga(s) a
	// lobby's Stages come from vs. which manga(s) its abilities/powers come
	// from need not match (e.g. Stages from both mangas with powers from
	// JoJo only).
	StageMangas         []enums.Manga
	PowerMangas         []enums.Manga
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
	// RevealSpeed is optional - its zero value is enums.Normal (see that
	// enum's doc), so an omitted field needs no extra fallback here.
	RevealSpeed enums.RevealSpeed
	// SummaryDurationSeconds is optional: a zero value defaults to
	// game.DefaultSummaryDurationSeconds, same pattern as
	// VotingWindowSeconds.
	SummaryDurationSeconds int
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
	// votingEnds holds the wall-clock deadline of the open voting (or
	// revote) window for each Game currently in VOTING/TIEBREAK with a
	// pending voting timer, keyed by GameID - read by VotingEndsAt so a
	// client joining/reconnecting mid-vote gets the real remaining time
	// instead of a full window (or no countdown at all). The twin of
	// revealEnds, and just as much process memory only: a backend restart
	// mid-vote loses this and the timer that owns it, the same known gap
	// scheduleVotingTimer has always had. The reveal and voting deadlines
	// never coexist for the same GameID (see the timers field doc).
	votingEnds map[game.GameID]time.Time
	// resultEnds holds the wall-clock deadline of the in-flight
	// round-result display for each Game currently in RESOLVING with a
	// pending result timer, keyed by GameID - read by ResultEndsAt so a
	// client joining/reconnecting mid-result gets the real remaining time.
	// Same process-memory-only caveat as revealEnds/votingEnds.
	resultEnds map[game.GameID]time.Time
	// summaryEnds holds the wall-clock deadline of the in-flight
	// loadout-summary screen for each Game currently in SUMMARY with a
	// pending summary timer, keyed by GameID - read by SummaryEndsAt so a
	// client joining/reconnecting mid-summary gets the real remaining time.
	// Same process-memory-only caveat as revealEnds/votingEnds/resultEnds.
	summaryEnds map[game.GameID]time.Time

	finalizedMu sync.Mutex
	// finalized remembers which Games this process has already recorded
	// history for, so a post-finish Reconnect/Disconnect (both legal against
	// a terminal game, both routed through withGame's Finished branch) does
	// not re-issue the inserts on every client reload. Purely an
	// optimization - the inserts themselves are idempotent.
	finalized map[game.GameID]struct{}
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
		votingCfg:   votingCfg,
		locks:       newGameLocks(),
		timers:      make(map[game.GameID]Timer),
		revealEnds:  make(map[game.GameID]time.Time),
		votingEnds:  make(map[game.GameID]time.Time),
		resultEnds:  make(map[game.GameID]time.Time),
		summaryEnds: make(map[game.GameID]time.Time),
		finalized:   make(map[game.GameID]struct{}),
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
	summaryDurationSeconds := input.SummaryDurationSeconds
	if summaryDurationSeconds == 0 {
		summaryDurationSeconds = game.DefaultSummaryDurationSeconds
	}
	return game.NewConfig(
		input.Mode, input.StageMangas, input.PowerMangas, input.AbilitySource, input.TeamSize, input.AllowBots,
		visibility, votingWindowSeconds, input.PoolFilter, input.RevealSpeed, summaryDurationSeconds,
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

// loadStages resolves cfg's selected StageMangas against the stage
// catalog. Gauntlet needs the full, pre-ordered round list (Interleave);
// Versus just needs the pool game.IGameMode.StageFor draws a random Stage
// from each round, so every selected manga's Stages are simply
// concatenated.
func (s *GameService) loadStages(ctx context.Context, cfg game.Config) ([]game.Stage, error) {
	mangas := cfg.StageMangas()
	byManga := make(map[enums.Manga][]game.Stage, len(mangas))
	for _, m := range mangas {
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
	for _, m := range mangas {
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

// Rematch builds a brand new lobby from a finished/aborted one, carrying
// the whole roster over so nobody has to re-enter a join code: same Config,
// the same seated humans on the matching teams, the same bots in the same
// positions, a fresh GameID, a fresh join code, state LOBBY. Host-only, and
// only valid once the source game is actually over.
//
// The source aggregate is left completely untouched - not reset back through
// its own state machine, not re-saved. A terminal Game is terminal; its
// history row and its own result screen must both stay intact, and a
// rematch is a *new* game, not a rewound one.
func (s *GameService) Rematch(ctx context.Context, gameID game.GameID, requesterID game.ParticipantID) (*game.Game, string, error) {
	// A plain read under the source game's lock: nothing here mutates it, so
	// withGame (which would publish and re-save) is the wrong tool.
	old, err := s.GetGame(ctx, gameID)
	if err != nil {
		return nil, "", err
	}
	if requesterID != old.HostID() {
		return nil, "", game.ErrNotHost
	}
	if old.State() != enums.Finished && old.State() != enums.Aborted {
		return nil, "", ErrGameNotOver
	}

	cfg := old.Config()
	teams, err := s.buildTeams(cfg.Mode())
	if err != nil {
		return nil, "", err
	}
	stages, err := s.loadStages(ctx, cfg)
	if err != nil {
		return nil, "", err
	}

	// Old team -> new team by position. Both lists come from the same mode,
	// so they're the same length and the same order (Squad; or Team A/Team
	// B), which is what keeps a Versus rematch's sides intact.
	teamFor := make(map[game.TeamID]game.TeamID, len(teams))
	oldTeams := old.Teams()
	for i, t := range oldTeams {
		if i < len(teams) {
			teamFor[t.ID()] = teams[i].ID()
		}
	}
	newTeamOf := func(p *game.Participant) game.TeamID {
		if id, ok := teamFor[p.TeamID()]; ok {
			return id
		}
		return teams[0].ID()
	}

	oldHost, ok := old.Participant(old.HostID())
	if !ok || oldHost.UserID() == nil {
		return nil, "", game.ErrParticipantNotFound
	}
	host, err := game.NewHumanParticipant(s.partIDs.NewID(), *oldHost.UserID(), oldHost.DisplayName(), newTeamOf(oldHost))
	if err != nil {
		return nil, "", err
	}
	host.SetAvatar(oldHost.AvatarThumbKey(), oldHost.GooglePicture())

	g, err := game.NewGame(s.gameIDs.NewID(), cfg, host, teams, stages)
	if err != nil {
		return nil, "", err
	}

	// Everyone else, in the source game's own join order.
	for _, p := range old.Participants() {
		if p.ID() == oldHost.ID() {
			continue
		}
		if p.IsBot() {
			bot, err := game.NewBotParticipant(s.partIDs.NewID(), p.DisplayName(), newTeamOf(p))
			if err != nil {
				return nil, "", err
			}
			if err := g.AddBot(bot); err != nil {
				return nil, "", err
			}
			continue
		}
		if p.UserID() == nil {
			continue
		}
		seat, err := game.NewHumanParticipant(s.partIDs.NewID(), *p.UserID(), p.DisplayName(), newTeamOf(p))
		if err != nil {
			return nil, "", err
		}
		seat.SetAvatar(p.AvatarThumbKey(), p.GooglePicture())
		if err := g.Join(seat); err != nil {
			return nil, "", err
		}
	}

	for attempt := 0; attempt < maxCodeAttempts; attempt++ {
		code := s.generateCode()
		if err := s.store.Create(ctx, code, g); err != nil {
			if errors.Is(err, ports.ErrGameCodeTaken) {
				continue
			}
			return nil, "", err
		}
		// Announced on the OLD game's hub, so every client still connected
		// to the finished game learns the new id - not just whoever asked.
		evt := game.RematchReady{GameID: g.ID()}
		s.hub.Publish(GameEvent{GameID: gameID, Name: evt.Name(), Event: evt})
		return g, code, nil
	}
	return nil, "", ErrCodeGenerationFailed
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
		if next.Mode() != g.Config().Mode() || !mangasEqual(next.StageMangas(), g.Config().StageMangas()) {
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
	if g.Config().HasPowerManga(enums.Jojo) && len(stands) < needed {
		return game.ErrPoolTooSmall
	}
	if g.Config().HasPowerManga(enums.OnePiece) && len(fruits) < needed {
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

		builder := game.NewLoadoutBuilder(g.Config().PowerMangas(), weights, s.rng)
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

// scheduleRevealDelay holds g in ASSIGNING for game.RevealDuration(...) -
// every client computes that same duration locally (loadout-reveal.ts),
// from the same mangas/participants'-loadouts/revealSpeed it already has in
// its own snapshot, to pace its reveal overlay - so this is what keeps
// "voting opens" from ever racing ahead of "everyone has seen their
// loadout". Once the delay elapses,
// openSummaryAfterReveal moves the Game on to SUMMARY (see that method) -
// the loadout-summary screen the owner added 2026-08-30 between the sorteo
// and the actual vote. Uses the same s.timers map as the voting/summary
// timer (see its field doc) and records the deadline in s.revealEnds for
// RevealEndsAt to serve to (re)connecting clients.
func (s *GameService) scheduleRevealDelay(g *game.Game) {
	s.schedulePhaseTimer(g)
}

// schedulePhaseTimer arms a full fresh window for whichever timed phase g
// has just entered - the body the four schedule* wrappers all share. It is
// the same armPhaseTimer the re-arm path uses, so a scheduled deadline and a
// re-armed one are recorded identically (in-process map + persisted on the
// aggregate).
func (s *GameService) schedulePhaseTimer(g *game.Game) {
	pt, ok := s.phaseTimerFor(g)
	if !ok {
		return
	}
	s.armPhaseTimer(g, pt, s.clock.Now().Add(pt.window))
}

// openSummaryAfterReveal is scheduleRevealDelay's callback: it re-acquires
// g's lock via withGame and only then opens the summary screen, so it never
// races a concurrent AbortGame/CastVote against the same Game.
func (s *GameService) openSummaryAfterReveal(ctx context.Context, id game.GameID) {
	s.timersMu.Lock()
	delete(s.revealEnds, id)
	s.timersMu.Unlock()

	_, err := s.withGame(ctx, id, func(g *game.Game) error {
		if err := g.OpenSummary(); err != nil {
			return err
		}
		s.scheduleSummaryDelay(g)
		return nil
	})
	if err != nil {
		switch {
		case errors.Is(err, ports.ErrGameNotFound), errors.Is(err, game.ErrInvalidStateTransition):
			// Benign: the Game was aborted/removed, or someone else already
			// moved it out of ASSIGNING, before this timer fired.
		default:
			log.Printf("opening summary after reveal for game %s: %v", id, err)
		}
	}
}

// scheduleSummaryDelay holds g in SUMMARY for the lobby's configured
// SummaryDurationSeconds, mirroring scheduleRevealDelay exactly. Once the
// delay elapses, openVotingAfterSummary opens the actual vote.
func (s *GameService) scheduleSummaryDelay(g *game.Game) {
	s.schedulePhaseTimer(g)
}

// openVotingAfterSummary is scheduleSummaryDelay's callback: it re-acquires
// g's lock via withGame and only then opens voting, so it never races a
// concurrent AbortGame/CastVote against the same Game.
func (s *GameService) openVotingAfterSummary(ctx context.Context, id game.GameID) {
	s.timersMu.Lock()
	delete(s.summaryEnds, id)
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
			// moved it out of SUMMARY, before this timer fired.
		default:
			log.Printf("opening voting after summary for game %s: %v", id, err)
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

// VotingEndsAt reports id's open voting-window deadline, if any - the same
// role RevealEndsAt plays for the reveal: a client joining or reconnecting
// mid-vote resumes the real countdown instead of assuming a full window.
// The bool is false once the window has closed (or never opened), and
// false during the reveal phase, since the reveal and voting timers never
// coexist for the same Game (see the timers field doc).
func (s *GameService) VotingEndsAt(id game.GameID) (time.Time, bool) {
	s.timersMu.Lock()
	defer s.timersMu.Unlock()
	t, ok := s.votingEnds[id]
	return t, ok
}

// ResultEndsAt reports id's in-flight round-result display deadline, if
// any - the same role RevealEndsAt/VotingEndsAt play for their own phases:
// a client joining or reconnecting mid-RESOLVING resumes the real
// countdown instead of missing the window entirely. The bool is false once
// the result display has ended (or never started).
func (s *GameService) ResultEndsAt(id game.GameID) (time.Time, bool) {
	s.timersMu.Lock()
	defer s.timersMu.Unlock()
	t, ok := s.resultEnds[id]
	return t, ok
}

// SummaryEndsAt reports id's in-flight loadout-summary deadline, if any -
// the same role RevealEndsAt/VotingEndsAt/ResultEndsAt play for their own
// phases: a client joining or reconnecting mid-SUMMARY resumes the real
// countdown instead of restarting it. The bool is false once the summary
// has ended (or never started).
func (s *GameService) SummaryEndsAt(id game.GameID) (time.Time, bool) {
	s.timersMu.Lock()
	defer s.timersMu.Unlock()
	t, ok := s.summaryEnds[id]
	return t, ok
}

// scheduleResultDelay holds g in RESOLVING for game.ResultDuration so
// clients have time to render the round's outcome (winner, vote
// breakdown, coin flip) before the next round's sorteo starts - the exact
// same pattern scheduleRevealDelay uses for the ASSIGNING pause. Once the
// delay elapses, completeRoundAfterResult does what closeVoting used to do
// immediately: Game.CompleteRound + beginRound (if there's a next round).
func (s *GameService) scheduleResultDelay(g *game.Game) {
	s.schedulePhaseTimer(g)
}

// completeRoundAfterResult is scheduleResultDelay's callback: it
// re-acquires g's lock via withGame and only then advances the Game past
// RESOLVING, so it never races a concurrent AbortGame/CastVote against the
// same Game.
func (s *GameService) completeRoundAfterResult(ctx context.Context, id game.GameID) {
	s.timersMu.Lock()
	delete(s.resultEnds, id)
	s.timersMu.Unlock()

	_, err := s.withGame(ctx, id, func(g *game.Game) error {
		if err := g.CompleteRound(); err != nil {
			return err
		}
		if g.State() == enums.Assigning {
			return s.beginRound(ctx, g)
		}
		return nil
	})
	if err != nil {
		switch {
		case errors.Is(err, ports.ErrGameNotFound), errors.Is(err, game.ErrInvalidStateTransition):
			// Benign: the Game was aborted/removed, or someone else already
			// moved it out of RESOLVING, before this timer fired.
		default:
			log.Printf("completing round for game %s: %v", id, err)
		}
	}
}

// --- Voting ---

// CastVote records participantID's vote and force-closes the window early
// once every connected human has voted (see game.Game.VotingComplete).
// MarkRevealReady records that participantID (a connected human) is done
// watching its own sorteo reveal - the server-side half of the "todos
// pueden saltar" skip (owner decision, 2026-08-30). Once every connected
// human has called this during the lobby's current ASSIGNING window, the
// pending reveal timer is cancelled and the loadout-summary screen opens
// immediately (see OpenSummary) - skipping the sorteo animation lands on
// the summary, not straight into voting, same as letting it play out fully
// would.
func (s *GameService) MarkRevealReady(ctx context.Context, gameID game.GameID, participantID game.ParticipantID) (*game.Game, error) {
	return s.withGame(ctx, gameID, func(g *game.Game) error {
		if err := g.MarkRevealReady(participantID); err != nil {
			return err
		}
		if g.RevealReadyComplete() {
			s.cancelTimer(g.ID())
			if err := g.OpenSummary(); err != nil {
				return err
			}
			s.scheduleSummaryDelay(g)
		}
		return nil
	})
}

// MarkSummaryReady records that participantID (a connected human) is done
// looking at the loadout-summary screen and wants to skip ahead - the
// server-side half of the summary screen's own "todos pueden saltar" skip,
// mirroring MarkRevealReady exactly but for the SUMMARY window. Once every
// connected human has called this, the pending summary timer is cancelled
// and voting opens immediately.
func (s *GameService) MarkSummaryReady(ctx context.Context, gameID game.GameID, participantID game.ParticipantID) (*game.Game, error) {
	return s.withGame(ctx, gameID, func(g *game.Game) error {
		if err := g.MarkSummaryReady(participantID); err != nil {
			return err
		}
		if g.SummaryReadyComplete() {
			s.cancelTimer(g.ID())
			if err := g.OpenVoting(s.rng); err != nil {
				return err
			}
			s.scheduleVotingTimer(g)
		}
		return nil
	})
}

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

	// A clear winner (or a resolved tiebreak) leaves g in RESOLVING - hold
	// it there so clients can see the round's outcome before the next
	// round's sorteo starts (see scheduleResultDelay/CompleteRound).
	if g.State() == enums.Resolving {
		s.scheduleResultDelay(g)
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

// GetGame returns the live Game identified by id. It also re-arms the phase
// timer when this process has none for a game in a timed phase - this read
// path is what a reconnecting client's RESYNC/initial snapshot goes through
// (see pushCurrentState), so connecting to a game stranded by a restart is
// what heals it.
func (s *GameService) GetGame(ctx context.Context, id game.GameID) (*game.Game, error) {
	mu := s.locks.get(id)
	mu.Lock()
	defer mu.Unlock()
	g, err := s.store.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	s.rearmPhaseTimerLocked(g)
	return g, nil
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
		GameID:          g.ID(),
		Mode:            cfg.Mode(),
		HostDisplayName: hostName,
		PlayerCount:     len(g.Participants()),
		MaxPlayers:      maxPlayers,
		// Mangas is the union of StageMangas/PowerMangas - a public lobby
		// listing only needs "which mangas does this lobby touch at all",
		// not the split; the split itself lives in GameConfigResponse for
		// clients that already joined.
		Mangas:              mangaUnion(cfg.StageMangas(), cfg.PowerMangas()),
		AbilitySource:       cfg.AbilitySource(),
		AllowBots:           cfg.AllowBots(),
		Visibility:          cfg.Visibility(),
		VotingWindowSeconds: cfg.VotingWindowSeconds(),
		Locked:              g.Locked(),
	}
}

// mangaUnion returns every manga appearing in either a or b, deduplicated
// and in enums.Mangas() canonical order.
func mangaUnion(a, b []enums.Manga) []enums.Manga {
	seen := make(map[enums.Manga]struct{}, len(a)+len(b))
	for _, m := range a {
		seen[m] = struct{}{}
	}
	for _, m := range b {
		seen[m] = struct{}{}
	}
	union := make([]enums.Manga, 0, len(seen))
	for _, m := range enums.Mangas() {
		if _, ok := seen[m]; ok {
			union = append(union, m)
		}
	}
	return union
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
	// Before any mutation: if this process has no timer for a Game sitting in
	// a timed phase (a restart, or a game this instance has never touched),
	// re-establish it from the persisted deadline.
	s.rearmPhaseTimerLocked(g)

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

// FinishedGameTTL is how long a FINISHED/ABORTED game stays readable in the
// store before it expires. Deliberately short and deliberately NOT the
// lobby TTL (config.Config.GameLobbyTTL, 2h): a terminal game only needs to
// outlive its players reading the result screen, and keeping it around for
// hours would be a pure storage cost with nothing to join.
//
// It lives here rather than next to defaultGameLobbyTTL because that const
// is package-private to internal/config, which the application layer must
// not import.
const FinishedGameTTL = 15 * time.Minute

// finalizeLocked records g's outcome (best-effort, tolerating a nil history
// port) and saves it back under a short TTL. Callers must already hold g's
// lock.
//
// It used to Delete the game outright the instant it reached
// FINISHED/ABORTED, which made a result screen impossible: the terminal
// frame and the client's follow-up read raced, and the read almost always
// lost with ErrGameNotFound. The aggregate now survives for
// FinishedGameTTL, long enough for every client to fetch the final STATE.
// It disappears from the public lobby browser regardless - both store
// implementations re-derive IsPubliclyJoinable on save, and that is false
// the moment state leaves LOBBY.
func (s *GameService) finalizeLocked(ctx context.Context, g *game.Game) {
	id := g.ID()

	// Record history exactly once per game, however many times a post-finish
	// Reconnect/Disconnect re-enters this branch. Both are legal against a
	// finished game and both flow through withGame's Finished branch, so
	// without this a reconnecting client would re-issue the inserts on every
	// reload. Correctness does not depend on it - the inserts are idempotent
	// (see the repository's ON CONFLICT DO NOTHING) - this only avoids the
	// pointless round trips.
	s.finalizedMu.Lock()
	_, already := s.finalized[id]
	if !already {
		s.finalized[id] = struct{}{}
	}
	s.finalizedMu.Unlock()

	if !already {
		if result, err := g.Result(); err != nil {
			log.Printf("computing result for finished game %s: %v", id, err)
		} else if s.history != nil {
			if err := s.history.Record(ctx, result); err != nil {
				log.Printf("recording history for game %s: %v", id, err)
			}
		}
	}

	if err := s.store.SaveWithTTL(ctx, g, FinishedGameTTL); err != nil {
		log.Printf("saving finished game %s: %v", id, err)
	}
	s.cancelTimer(id)
	// Deliberately NOT s.locks.delete(id) here, unlike the pre-result-screen
	// version of this method. The aggregate stays live and readable now, so
	// evicting its mutex would let a later reader and a concurrent
	// Reconnect/Disconnect serialize on two *different* mutex instances for
	// the same logical game - i.e. no serialization at all. The lock map's
	// own entry is a few words; leaking one per finished game until the
	// process restarts is far cheaper than that race.
}

// publish drains every DomainEvent g has accumulated since the last call
// and fans them out over hub.
//
// Timed events (VOTING_OPENED/TIEBREAK_OPENED/SUMMARY_OPENED) are stamped
// with the authoritative deadline this service already computed for that
// phase, so the transport never has to re-synthesize time.Now()+window and
// drift by however long hub delivery took. This is only correct because
// every path that emits one of those events calls its schedule* function
// inside the same withGame closure, BEFORE withGame calls publish: the
// non-reassigning beginRound path, openVotingAfterSummary,
// openSummaryAfterReveal, MarkRevealReady/MarkSummaryReady's skip paths and
// closeVoting's first-tie revote path all do. Any new emit path must keep
// that ordering or its frame silently falls back to the synthesized value.
func (s *GameService) publish(g *game.Game) {
	window := time.Duration(g.Config().VotingWindowSeconds()) * time.Second
	revealMs := revealDurationFor(g)
	summaryWindow := time.Duration(g.Config().SummaryDurationSeconds()) * time.Second

	id := g.ID()
	s.timersMu.Lock()
	votingEndsAt, hasVotingEnd := s.votingEnds[id]
	summaryEndsAt, hasSummaryEnd := s.summaryEnds[id]
	s.timersMu.Unlock()

	for _, e := range g.PullEvents() {
		var closesAt time.Time
		switch e.(type) {
		case game.VotingOpened, game.TiebreakOpened:
			if hasVotingEnd {
				closesAt = votingEndsAt
			}
		case game.SummaryOpened:
			if hasSummaryEnd {
				closesAt = summaryEndsAt
			}
		}
		s.hub.Publish(GameEvent{
			GameID: id, Name: e.Name(), Event: e,
			VotingWindow: window, RevealMs: revealMs, SummaryWindow: summaryWindow,
			ClosesAt: closesAt,
		})
	}
}

// revealDurationFor computes game.RevealDuration for g as it stands right
// now: its own mangas/RevealSpeed, its current round index (len(g.Rounds())
// - the reveal always runs before that round's Round is created, see
// scheduleRevealDelay's doc), and each participant's own already-assigned
// Loadout (nil before AssignLoadouts has run, treated as "landed nothing").
// Shared by scheduleRevealDelay (to size the actual timer) and publish (to
// stamp every GameEvent's RevealMs, most importantly LOADOUTS_ASSIGNED's)
// so the two never compute a different number for the same reveal.
func revealDurationFor(g *game.Game) time.Duration {
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
	roundIndex := len(g.Rounds())
	return game.RevealDuration(g.ID(), roundIndex, g.Config().PowerMangas(), players, g.Config().RevealSpeed())
}

// scheduleVotingTimer starts g's voting-window timer using its own
// per-lobby Config.VotingWindowSeconds() (host-configurable at creation/
// edit time) rather than the service-wide VotingPolicy.Window default,
// which now only seeds new lobbies that don't specify one - see
// buildConfig.
func (s *GameService) scheduleVotingTimer(g *game.Game) {
	s.schedulePhaseTimer(g)
}

// phaseTimerKind names which of the four per-phase deadline maps a pending
// timer belongs to. Only one is ever populated for a given GameID at a time
// (see the timers field doc), which is also why the aggregate persists a
// single phaseEndsAt rather than four.
type phaseTimerKind int

const (
	phaseReveal phaseTimerKind = iota
	phaseSummary
	phaseVoting
	phaseResult
)

// phaseTimer describes the timed phase a Game is currently sitting in: which
// deadline map owns it, how long a *full* window of that phase lasts for
// this particular lobby, and what to run when it expires.
type phaseTimer struct {
	fire   func()
	kind   phaseTimerKind
	window time.Duration
}

// phaseTimerFor derives g's timed phase from its state alone. State is
// authoritative and needs nothing else to exist (no Round, no loadouts), so
// both the schedule path - which calls this right after the transition that
// entered the phase - and the re-arm path, which calls it on a Game freshly
// loaded from the store, agree by construction. ok is false for LOBBY and
// the two terminal states, which have no timer.
func (s *GameService) phaseTimerFor(g *game.Game) (phaseTimer, bool) {
	id := g.ID()
	switch g.State() {
	case enums.Assigning:
		return phaseTimer{kind: phaseReveal, window: revealDurationFor(g), fire: func() {
			s.openSummaryAfterReveal(context.Background(), id)
		}}, true
	case enums.Summary:
		return phaseTimer{kind: phaseSummary, window: time.Duration(g.Config().SummaryDurationSeconds()) * time.Second, fire: func() {
			s.openVotingAfterSummary(context.Background(), id)
		}}, true
	case enums.Voting, enums.Tiebreak:
		return phaseTimer{kind: phaseVoting, window: time.Duration(g.Config().VotingWindowSeconds()) * time.Second, fire: func() {
			if _, err := s.CloseVotingWindow(context.Background(), id); err != nil {
				switch {
				case errors.Is(err, ports.ErrGameNotFound), errors.Is(err, game.ErrVotingClosed),
					errors.Is(err, game.ErrInvalidStateTransition):
					// Benign: the Game was already finished/aborted/removed,
					// or another instance's re-armed timer got there first,
					// before this timer fired.
				default:
					log.Printf("closing voting window for game %s: %v", id, err)
				}
			}
		}}, true
	case enums.Resolving:
		return phaseTimer{kind: phaseResult, window: game.ResultDuration, fire: func() {
			s.completeRoundAfterResult(context.Background(), id)
		}}, true
	case enums.Lobby, enums.Finished, enums.Aborted:
		return phaseTimer{}, false
	default:
		return phaseTimer{}, false
	}
}

// endsMapFor returns the deadline map owning kind. Callers must already hold
// timersMu.
func (s *GameService) endsMapFor(kind phaseTimerKind) map[game.GameID]time.Time {
	switch kind {
	case phaseReveal:
		return s.revealEnds
	case phaseSummary:
		return s.summaryEnds
	case phaseVoting:
		return s.votingEnds
	case phaseResult:
		return s.resultEnds
	default:
		return s.votingEnds
	}
}

// armPhaseTimer is the one place a phase deadline is recorded and its timer
// started, shared by the four schedule* wrappers and by
// rearmPhaseTimerLocked so the two can never drift. It records deadline both
// in this process's own map (for the *EndsAt accessors a connecting client
// reads) and on the aggregate itself via SetPhaseDeadline, so the deadline
// survives persistence and a restart.
//
// Any timer already pending for this Game is stopped first: the schedule
// paths mostly cancelTimer beforehand anyway, but the re-arm path can race
// one in, and two live timers for the same Game would fire the same
// transition twice.
func (s *GameService) armPhaseTimer(g *game.Game, pt phaseTimer, deadline time.Time) {
	id := g.ID()
	g.SetPhaseDeadline(deadline)

	remaining := deadline.Sub(s.clock.Now())
	if remaining < 0 {
		remaining = 0
	}

	s.timersMu.Lock()
	s.endsMapFor(pt.kind)[id] = deadline
	if prev, ok := s.timers[id]; ok {
		prev.Stop()
		delete(s.timers, id)
	}
	s.timersMu.Unlock()

	timer := s.clock.AfterFunc(remaining, pt.fire)

	s.timersMu.Lock()
	s.timers[id] = timer
	s.timersMu.Unlock()
}

// rearmPhaseTimerLocked re-establishes the in-process timer for a Game that
// is sitting in a timed phase but has no live timer in THIS process -
// exactly what a backend restart mid-vote/reveal/summary/result leaves
// behind, since s.timers and the four deadline maps are process memory only.
// Without it such a game wedges in that phase forever: no timer will ever
// fire to close it, and a client gets no deadline to count down against.
//
// Called from two places: inside withGame right after store.Get succeeds
// (before any mutation), and from GetGame's read path - the latter is what
// heals a reconnecting client, since RESYNC/pushCurrentState goes through it.
// Callers must already hold g's lock.
//
// Multi-instance safety: no leader election or distributed lock is needed.
// If two instances both load the same game and both re-arm, both callbacks
// eventually run; the loser's hits ErrInvalidStateTransition/ErrVotingClosed
// (or ErrGameNotFound), which every callback here already swallows as
// benign. The residual gap is accepted: a game nobody ever loads or connects
// to is never re-armed at all and simply expires by its lobby TTL, exactly
// as it does today.
func (s *GameService) rearmPhaseTimerLocked(g *game.Game) {
	pt, ok := s.phaseTimerFor(g)
	if !ok {
		return
	}

	id := g.ID()
	s.timersMu.Lock()
	_, alreadyArmed := s.timers[id]
	s.timersMu.Unlock()
	if alreadyArmed {
		// This instance is already driving the phase - re-arming would stop
		// a perfectly good timer and restart an identical one.
		return
	}

	deadline, ok := g.PhaseEndsAt()
	if !ok {
		// Legacy snapshot written before phaseEndsAt existed (or a phase
		// entered by a build that didn't record one). A full fresh window is
		// generous but bounded - strictly better than leaving the game
		// wedged with no timer at all. The stamped deadline is only
		// persisted if a mutation happens to save the aggregate afterwards;
		// until then every load simply re-derives the same fallback.
		deadline = s.clock.Now().Add(pt.window)
	}
	s.armPhaseTimer(g, pt, deadline)
}

func (s *GameService) cancelTimer(id game.GameID) {
	s.timersMu.Lock()
	defer s.timersMu.Unlock()
	delete(s.revealEnds, id)
	delete(s.votingEnds, id)
	delete(s.resultEnds, id)
	delete(s.summaryEnds, id)
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
