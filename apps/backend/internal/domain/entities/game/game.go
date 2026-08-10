package game

import (
	"errors"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// Game is the aggregate root for a single Gauntlet run or Versus match. It
// is a State machine over enums.GameState:
//
//	LOBBY -> ASSIGNING -> VOTING -> [TIEBREAK] -> RESOLVING -> ASSIGNING -> ... -> FINISHED
//	  |                                                                              ^
//	  +------------------------------------------------------------------------> ABORTED
//
// Every mode-specific decision (ballot options, stage selection, whether
// loadouts reassign each round, round-result handling, final outcome) is
// delegated to an IGameMode - Game itself never branches on
// enums.GameModeKind.
type Game struct {
	id           GameID
	config       Config
	mode         IGameMode
	hostID       ParticipantID
	state        enums.GameState
	participants map[ParticipantID]*Participant
	order        []ParticipantID // join order; deterministic host reassignment + vote-completeness scans
	teams        []*Team
	stages       []Stage
	rounds       []Round
	evaluator    LoadoutEvaluator
	events       []DomainEvent
}

// NewGame builds a Game in the LOBBY state with host already seated. teams
// must already exist (built via NewTeam by the caller, which owns id
// generation) and must match cfg.Mode()'s team count: 1 for Gauntlet, 2
// for Versus. stages is the full, already-ordered round list (see
// Interleave for Gauntlet) - Versus instead treats it as the pool StageFor
// draws from each round.
func NewGame(id GameID, cfg Config, host *Participant, teams []*Team, stages []Stage) (*Game, error) {
	if id.IsNil() {
		return nil, errors.New("id is required")
	}
	if host == nil {
		return nil, errors.New("host is required")
	}
	if host.Kind() != enums.Human {
		return nil, ErrBotsNotAllowed
	}

	expectedTeams := 1
	if cfg.Mode() == enums.Versus {
		expectedTeams = VersusTeamCount
	}
	if len(teams) != expectedTeams {
		return nil, ErrTeamSizeMismatch
	}
	if cfg.Mode() == enums.Gauntlet && len(stages) == 0 {
		return nil, ErrNoStagesAvailable
	}

	var mode IGameMode
	switch cfg.Mode() {
	case enums.Gauntlet:
		mode = GauntletMode{}
	case enums.Versus:
		mode = VersusMode{}
	default:
		return nil, enums.ErrInvalidGameModeKind
	}

	g := &Game{
		id:           id,
		config:       cfg,
		mode:         mode,
		hostID:       host.ID(),
		state:        enums.Lobby,
		participants: make(map[ParticipantID]*Participant),
		teams:        append([]*Team(nil), teams...),
		stages:       append([]Stage(nil), stages...),
		evaluator:    DefaultLoadoutEvaluator{},
	}
	if err := g.addParticipant(host); err != nil {
		return nil, err
	}
	return g, nil
}

// --- Getters ---

func (g *Game) ID() GameID             { return g.id }
func (g *Game) Config() Config         { return g.config }
func (g *Game) State() enums.GameState { return g.state }
func (g *Game) HostID() ParticipantID  { return g.hostID }
func (g *Game) Mode() IGameMode        { return g.mode }
func (g *Game) Teams() []*Team         { return append([]*Team(nil), g.teams...) }
func (g *Game) Rounds() []Round        { return append([]Round(nil), g.rounds...) }

func (g *Game) Participant(id ParticipantID) (*Participant, bool) {
	p, ok := g.participants[id]
	return p, ok
}

// Participants returns every participant, in join order.
func (g *Game) Participants() []*Participant {
	out := make([]*Participant, 0, len(g.order))
	for _, id := range g.order {
		out = append(out, g.participants[id])
	}
	return out
}

// SetLoadoutEvaluator overrides the evaluator used for bot voting.
// no-op if evaluator is nil.
func (g *Game) SetLoadoutEvaluator(evaluator LoadoutEvaluator) {
	if evaluator != nil {
		g.evaluator = evaluator
	}
}

func (g *Game) teamByID(id TeamID) *Team {
	for _, t := range g.teams {
		if t.ID() == id {
			return t
		}
	}
	return nil
}

// --- Membership ---

func (g *Game) addParticipant(p *Participant) error {
	if _, exists := g.participants[p.ID()]; exists {
		return ErrDuplicateParticipant
	}
	team := g.teamByID(p.TeamID())
	if team == nil {
		return ErrTeamNotFound
	}
	if team.Size() >= g.config.TeamSize() {
		if g.config.Mode() == enums.Gauntlet {
			return ErrGameFull
		}
		return ErrTeamFull
	}
	g.participants[p.ID()] = p
	g.order = append(g.order, p.ID())
	team.AddMember(p.ID())
	return nil
}

func (g *Game) removeFromOrder(id ParticipantID) {
	for i, pid := range g.order {
		if pid == id {
			g.order = append(g.order[:i:i], g.order[i+1:]...)
			return
		}
	}
}

// Join seats a human participant while the Game is still in the LOBBY.
func (g *Game) Join(p *Participant) error {
	if g.state != enums.Lobby {
		return ErrInvalidStateTransition
	}
	if p.Kind() != enums.Human {
		return ErrBotsNotAllowed
	}
	if err := g.addParticipant(p); err != nil {
		return err
	}
	g.emit(PlayerJoined{ParticipantID: p.ID()})
	return nil
}

// AddBot seats a bot participant. Only allowed in a Versus Game whose
// Config.AllowBots() is true, and only while still in the LOBBY.
func (g *Game) AddBot(p *Participant) error {
	if g.state != enums.Lobby {
		return ErrInvalidStateTransition
	}
	if g.config.Mode() != enums.Versus || !g.config.AllowBots() {
		return ErrBotsNotAllowed
	}
	if p.Kind() != enums.Bot {
		return errors.New("participant must be a bot")
	}
	if err := g.addParticipant(p); err != nil {
		return err
	}
	g.emit(PlayerJoined{ParticipantID: p.ID()})
	return nil
}

// Leave removes a participant entirely (as opposed to Disconnect, which
// keeps their seat for a possible Reconnect). It reassigns the host if
// needed and checks whether the Game must now Abort.
func (g *Game) Leave(id ParticipantID, rng RandomSource) error {
	p, ok := g.participants[id]
	if !ok {
		return ErrParticipantNotFound
	}
	team := g.teamByID(p.TeamID())
	delete(g.participants, id)
	g.removeFromOrder(id)
	if team != nil {
		team.RemoveMember(id)
	}
	g.emit(PlayerLeft{ParticipantID: id})

	if id == g.hostID {
		g.reassignHost(rng)
	}
	g.checkAbortConditions()
	return nil
}

// Disconnect marks a participant unreachable without removing their seat,
// reassigning the host and checking abort conditions exactly like Leave.
func (g *Game) Disconnect(id ParticipantID, rng RandomSource) error {
	p, ok := g.participants[id]
	if !ok {
		return ErrParticipantNotFound
	}
	p.Disconnect()
	if id == g.hostID {
		g.reassignHost(rng)
	}
	g.checkAbortConditions()
	return nil
}

// Reconnect marks a previously disconnected participant reachable again.
func (g *Game) Reconnect(id ParticipantID) error {
	p, ok := g.participants[id]
	if !ok {
		return ErrParticipantNotFound
	}
	p.Reconnect()
	return nil
}

func (g *Game) reassignHost(rng RandomSource) {
	candidates := make([]ParticipantID, 0, len(g.order))
	for _, pid := range g.order {
		if p := g.participants[pid]; p != nil && p.Kind() == enums.Human && p.Connected() {
			candidates = append(candidates, pid)
		}
	}
	if len(candidates) == 0 {
		g.hostID = NilParticipantID
		return
	}
	g.hostID = candidates[rng.IntN(len(candidates))]
	g.emit(HostReassigned{NewHostID: g.hostID})
}

func (g *Game) hasConnectedHuman() bool {
	for _, p := range g.participants {
		if p.Kind() == enums.Human && p.Connected() {
			return true
		}
	}
	return false
}

func (g *Game) checkAbortConditions() {
	if g.state == enums.Finished || g.state == enums.Aborted {
		return
	}
	if !g.hasConnectedHuman() {
		g.abort("no connected humans remain")
		return
	}
	if g.config.Mode() == enums.Versus {
		for _, t := range g.teams {
			if t.Size() == 0 {
				g.abort("a team has no players")
				return
			}
		}
	}
}

func (g *Game) abort(reason string) {
	g.state = enums.Aborted
	g.emit(GameAborted{Reason: reason})
}

// Abort is the host-triggered cancellation.
func (g *Game) Abort(callerID ParticipantID) error {
	if callerID != g.hostID {
		return ErrNotHost
	}
	if g.state == enums.Finished || g.state == enums.Aborted {
		return ErrInvalidStateTransition
	}
	g.abort("cancelled by host")
	return nil
}

// --- Lifecycle ---

// Start moves the Game from LOBBY to ASSIGNING. Only the host may call it,
// and only once every team meets its size requirement (Gauntlet: at least
// one player; Versus: exactly Config.TeamSize() on both teams).
func (g *Game) Start(callerID ParticipantID) error {
	if g.state != enums.Lobby {
		return ErrInvalidStateTransition
	}
	if callerID != g.hostID {
		return ErrNotHost
	}
	switch g.config.Mode() {
	case enums.Gauntlet:
		if g.teams[0].Size() < MinGauntletPlayers {
			return ErrNotEnoughPlayers
		}
	case enums.Versus:
		for _, t := range g.teams {
			if t.Size() != g.config.TeamSize() {
				return ErrTeamSizeMismatch
			}
		}
	}
	g.state = enums.Assigning
	g.emit(GameStarted{})
	return nil
}

// AssignLoadouts draws a fresh Loadout for every participant, using a
// builder and a per-team pool the caller has already prepared (each
// AvailablePowers must not be shared between teams - see AvailablePowers).
// Called once after Start for Gauntlet, and once per round before
// OpenVoting for Versus (see IGameMode.ReassignsEachRound).
func (g *Game) AssignLoadouts(builder *LoadoutBuilder, poolByTeam map[TeamID]*AvailablePowers) error {
	if g.state != enums.Assigning {
		return ErrInvalidStateTransition
	}
	for _, pid := range g.order {
		p := g.participants[pid]
		pool, ok := poolByTeam[p.TeamID()]
		if !ok {
			return ErrTeamNotFound
		}
		loadout, err := builder.Build(pool)
		if err != nil {
			return err
		}
		p.AssignLoadout(loadout)
	}
	g.emit(LoadoutsAssigned{RoundIndex: len(g.rounds)})
	return nil
}

// OpenVoting picks the round's Stage, opens a fresh Ballot, and casts every
// connected bot's vote immediately (so VotingComplete never blocks on a
// bot). Moves the Game from ASSIGNING to VOTING.
func (g *Game) OpenVoting(rng RandomSource) error {
	if g.state != enums.Assigning {
		return ErrInvalidStateTransition
	}
	roundIndex := len(g.rounds)
	stage, err := g.mode.StageFor(g, roundIndex, rng)
	if err != nil {
		return err
	}
	options := g.mode.BallotOptions(g)
	ballot, err := NewBallot(options)
	if err != nil {
		return err
	}
	g.rounds = append(g.rounds, Round{Index: roundIndex, Stage: stage, Ballot: ballot})
	g.state = enums.Voting
	g.emit(VotingOpened{RoundIndex: roundIndex})
	g.castBotVotes(roundIndex)
	return nil
}

func (g *Game) currentRound() *Round {
	if len(g.rounds) == 0 {
		return nil
	}
	return &g.rounds[len(g.rounds)-1]
}

func (g *Game) optionScores(options []OptionID) map[OptionID]int {
	scores := make(map[OptionID]int, len(options))
	if g.config.Mode() != enums.Versus {
		return scores
	}
	for _, t := range g.teams {
		total := 0
		for _, pid := range t.Members() {
			if p := g.participants[pid]; p != nil {
				total += g.evaluator.Score(p.Loadout())
			}
		}
		scores[OptionID(t.ID().String())] = total
	}
	return scores
}

func (g *Game) castBotVotes(roundIndex int) {
	round := &g.rounds[roundIndex]
	options := g.mode.BallotOptions(g)
	scores := g.optionScores(options)
	voter := NewBotVoter(g.evaluator)
	for _, pid := range g.order {
		p := g.participants[pid]
		if p.Kind() != enums.Bot || !p.Connected() {
			continue
		}
		choice := voter.Vote(options, scores)
		if err := round.Ballot.Cast(pid, choice); err == nil {
			g.emit(VoteCast{RoundIndex: roundIndex, ParticipantID: pid, Option: choice})
		}
	}
}

// CastVote records a human's (or, during a re-vote window, anyone's) vote
// for the current round. Votes may be changed freely until the window
// closes - the last cast vote is the one that counts.
func (g *Game) CastVote(id ParticipantID, o OptionID) error {
	if g.state != enums.Voting && g.state != enums.Tiebreak {
		return ErrVotingClosed
	}
	if _, ok := g.participants[id]; !ok {
		return ErrParticipantNotFound
	}
	round := g.currentRound()
	if round == nil {
		return ErrVotingClosed
	}
	if err := round.Ballot.Cast(id, o); err != nil {
		return err
	}
	g.emit(VoteCast{RoundIndex: round.Index, ParticipantID: id, Option: o})
	return nil
}

// VotingComplete reports whether every connected human has cast a vote in
// the current round. The application layer uses this to close the voting
// window as soon as it goes true, instead of always waiting out its 30s
// timer - the domain itself knows nothing about that timer. Bots always
// vote immediately when the window opens (see OpenVoting), so they never
// block this.
func (g *Game) VotingComplete() bool {
	if g.state != enums.Voting && g.state != enums.Tiebreak {
		return false
	}
	round := g.currentRound()
	if round == nil {
		return false
	}
	for _, pid := range g.order {
		p := g.participants[pid]
		if p.Kind() == enums.Human && p.Connected() && !round.Ballot.HasVoted(pid) {
			return false
		}
	}
	return true
}

// CloseVoting tallies the current round's Ballot. A clear winner resolves
// the round immediately (moving to RESOLVING and then ASSIGNING/FINISHED -
// see resolveRound). A tie opens a single revote window (TIEBREAK) the
// first time it happens for this round; if it is still tied when
// CloseVoting is called again, tied=true is returned once more and the
// caller must resolve it externally (a coin flip today, an LLM call later
// - see ports.ITiebreaker) and call ResolveTiebreak with the result.
func (g *Game) CloseVoting() (tied bool, err error) {
	if g.state != enums.Voting && g.state != enums.Tiebreak {
		return false, ErrVotingClosed
	}
	round := g.currentRound()
	if round == nil {
		return false, ErrVotingClosed
	}
	winner, isTied := round.Ballot.Tally()
	if !isTied {
		g.resolveRound(winner, false)
		return false, nil
	}
	if !round.TiebreakUsed {
		round.TiebreakUsed = true
		g.state = enums.Tiebreak
		g.emit(TiebreakOpened{RoundIndex: round.Index})
	}
	return true, nil
}

// ResolveTiebreak applies an externally-decided winner after CloseVoting
// has reported tied=true a second time for the current round (i.e. the
// revote also tied, or nobody voted at all).
func (g *Game) ResolveTiebreak(winner OptionID) error {
	if g.state != enums.Tiebreak {
		return ErrInvalidStateTransition
	}
	round := g.currentRound()
	if round == nil {
		return ErrVotingClosed
	}
	if !round.Ballot.isValidOption(winner) {
		return ErrInvalidBallotOption
	}
	g.resolveRound(winner, true)
	return nil
}

func (g *Game) resolveRound(winner OptionID, coinFlip bool) {
	round := g.currentRound()
	round.Result = &RoundResult{Winner: winner, DecidedByCoinFlip: coinFlip}
	g.state = enums.Resolving
	g.emit(RoundResolved{RoundIndex: round.Index, Winner: winner, DecidedByCoinFlip: coinFlip})

	finished := g.mode.ApplyRoundResult(g, *round)
	if finished {
		g.state = enums.Finished
		result, _ := g.mode.Outcome(g)
		g.emit(GameFinished{Result: result})
		return
	}
	g.state = enums.Assigning
}

// Result computes the final GameResult. Only valid once the Game is
// FINISHED or ABORTED.
func (g *Game) Result() (GameResult, error) {
	if g.state != enums.Finished && g.state != enums.Aborted {
		return GameResult{}, ErrInvalidStateTransition
	}
	return g.mode.Outcome(g)
}

// --- Events ---

func (g *Game) emit(e DomainEvent) {
	g.events = append(g.events, e)
}

// PullEvents drains and returns every DomainEvent accumulated since the
// last call.
func (g *Game) PullEvents() []DomainEvent {
	events := g.events
	g.events = nil
	return events
}
