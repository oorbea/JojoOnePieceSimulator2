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
	locked       bool // host-toggled: blocks new Join while true (AddBot is unaffected)
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

	mode, err := modeFor(cfg.Mode())
	if err != nil {
		return nil, err
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

// modeFor resolves the IGameMode strategy for a GameModeKind. Both
// implementations are stateless (empty structs deriving everything from the
// *Game passed to them), so the mode itself is never serialized - Restore
// calls this too, instead of persisting anything mode-related.
func modeFor(k enums.GameModeKind) (IGameMode, error) {
	switch k {
	case enums.Gauntlet:
		return GauntletMode{}, nil
	case enums.Versus:
		return VersusMode{}, nil
	default:
		return nil, enums.ErrInvalidGameModeKind
	}
}

// --- Getters ---

func (g *Game) ID() GameID             { return g.id }
func (g *Game) Config() Config         { return g.config }
func (g *Game) State() enums.GameState { return g.state }
func (g *Game) HostID() ParticipantID  { return g.hostID }
func (g *Game) Mode() IGameMode        { return g.mode }
func (g *Game) Teams() []*Team         { return append([]*Team(nil), g.teams...) }
func (g *Game) Rounds() []Round        { return append([]Round(nil), g.rounds...) }
func (g *Game) Stages() []Stage        { return append([]Stage(nil), g.stages...) }
func (g *Game) Locked() bool           { return g.locked }

// IsPubliclyJoinable reports whether g should appear in the public lobby
// browser: still in LOBBY, marked PUBLIC, and not locked. Shared by both
// ports.IGameStore adapters so the browsing predicate lives in one place.
func (g *Game) IsPubliclyJoinable() bool {
	return g.state == enums.Lobby && g.config.visibility == enums.Public && !g.locked
}

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
	if g.locked {
		return ErrLobbyLocked
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
	// An empty Versus team only aborts once the match is actually under
	// way: in LOBBY it's normal and recoverable (more players can still
	// join, or SwitchTeam can empty a team on the way to an even split) -
	// see errors.go's ErrTeamSizeMismatch doc and Game.Start's own gate.
	if g.config.Mode() == enums.Versus && g.state != enums.Lobby {
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

// SwitchTeam moves targetID onto teamID while the Game is still in LOBBY.
// Any participant may move themselves; moving someone else requires being
// the host. A no-op (target already on teamID) succeeds without emitting an
// event.
func (g *Game) SwitchTeam(callerID, targetID ParticipantID, teamID TeamID) error {
	if g.state != enums.Lobby {
		return ErrInvalidStateTransition
	}
	if callerID != targetID && callerID != g.hostID {
		return ErrNotHost
	}
	target, ok := g.participants[targetID]
	if !ok {
		return ErrParticipantNotFound
	}
	to := g.teamByID(teamID)
	if to == nil {
		return ErrTeamNotFound
	}
	if target.TeamID() == teamID {
		return nil
	}
	if to.Size() >= g.config.TeamSize() {
		return ErrTeamFull
	}
	from := g.teamByID(target.TeamID())
	fromID := target.TeamID()
	if from != nil {
		from.RemoveMember(targetID)
	}
	to.AddMember(targetID)
	target.setTeam(teamID)
	g.emit(TeamChanged{ParticipantID: targetID, FromTeamID: fromID, ToTeamID: teamID})
	return nil
}

// Kick removes targetID from the Game entirely. Host-only, LOBBY-only; the
// host cannot kick themselves (use Leave, or TransferHost first). Emits
// PlayerKicked before delegating the actual removal to Leave, so the
// transport can close the victim's socket before the roster changes
// underneath it.
func (g *Game) Kick(callerID, targetID ParticipantID, rng RandomSource) error {
	if g.state != enums.Lobby {
		return ErrInvalidStateTransition
	}
	if callerID != g.hostID {
		return ErrNotHost
	}
	if callerID == targetID {
		return ErrCannotKickSelf
	}
	if _, ok := g.participants[targetID]; !ok {
		return ErrParticipantNotFound
	}
	g.emit(PlayerKicked{ParticipantID: targetID})
	return g.Leave(targetID, rng)
}

// TransferHost hands the host role to targetID, a connected human
// participant. Host-only.
func (g *Game) TransferHost(callerID, targetID ParticipantID) error {
	if callerID != g.hostID {
		return ErrNotHost
	}
	if g.state == enums.Finished || g.state == enums.Aborted {
		return ErrInvalidStateTransition
	}
	target, ok := g.participants[targetID]
	if !ok {
		return ErrParticipantNotFound
	}
	if target.Kind() != enums.Human {
		return ErrBotsNotAllowed
	}
	if !target.Connected() {
		return ErrParticipantNotFound
	}
	g.hostID = targetID
	g.emit(HostReassigned{NewHostID: targetID})
	return nil
}

// SetLocked toggles whether new humans may Join this lobby. Host-only,
// LOBBY-only. AddBot is unaffected by locking - only human self-service
// joins are gated. A no-op (already at the requested value) succeeds
// without emitting an event.
func (g *Game) SetLocked(callerID ParticipantID, locked bool) error {
	if g.state != enums.Lobby {
		return ErrInvalidStateTransition
	}
	if callerID != g.hostID {
		return ErrNotHost
	}
	if g.locked == locked {
		return nil
	}
	g.locked = locked
	g.emit(LobbyLockChanged{Locked: locked})
	return nil
}

// Reconfigure replaces the lobby's Config while still in LOBBY. Host-only.
// newTeams must already be built by the caller (Game cannot mint TeamIDs):
// pass g.Teams() back unchanged when cfg.Mode() matches the current mode,
// or a freshly-built team set (1 team for Gauntlet, 2 for Versus) when the
// mode is changing. On a mode change, every currently seated human is
// re-seated onto newTeams by alternating join order (Gauntlet -> Versus) or
// merging onto the single team (Versus -> Gauntlet); bots are dropped if
// the new mode is Gauntlet or no longer allows them. Reconfigure never
// evicts a human: if newTeams would be too small to hold every seated human
// (because cfg lowers TeamSize), it fails with ErrConfigWouldEvictPlayers
// and leaves the Game entirely unchanged.
func (g *Game) Reconfigure(callerID ParticipantID, cfg Config, newTeams []*Team, stages []Stage) error {
	if g.state != enums.Lobby {
		return ErrInvalidStateTransition
	}
	if callerID != g.hostID {
		return ErrNotHost
	}
	expectedTeams := 1
	if cfg.Mode() == enums.Versus {
		expectedTeams = VersusTeamCount
	}
	if len(newTeams) != expectedTeams {
		return ErrTeamSizeMismatch
	}
	if cfg.Mode() == enums.Gauntlet && len(stages) == 0 {
		return ErrNoStagesAvailable
	}

	sameMode := cfg.Mode() == g.config.Mode()
	plan := make(map[ParticipantID]TeamID, len(g.order))
	var botsToDrop []ParticipantID

	if sameMode {
		// Team identities are unchanged; just re-check capacity against the
		// (possibly smaller) new TeamSize per existing team.
		counts := make(map[TeamID]int, len(g.teams))
		for _, pid := range g.order {
			p := g.participants[pid]
			if p.Kind() == enums.Bot && (cfg.Mode() != enums.Versus || !cfg.AllowBots()) {
				botsToDrop = append(botsToDrop, pid)
				continue
			}
			counts[p.TeamID()]++
			plan[pid] = p.TeamID()
		}
		for _, t := range newTeams {
			if counts[t.ID()] > cfg.TeamSize() {
				return ErrConfigWouldEvictPlayers
			}
		}
	} else if cfg.Mode() == enums.Versus {
		// Gauntlet -> Versus: alternate join order across the two new teams.
		humans := make([]ParticipantID, 0, len(g.order))
		for _, pid := range g.order {
			if g.participants[pid].Kind() == enums.Human {
				humans = append(humans, pid)
			} else {
				botsToDrop = append(botsToDrop, pid)
			}
		}
		if len(humans) > cfg.TeamSize()*expectedTeams {
			return ErrConfigWouldEvictPlayers
		}
		for i, pid := range humans {
			plan[pid] = newTeams[i%expectedTeams].ID()
		}
	} else {
		// Versus -> Gauntlet: merge every human onto the single team; bots
		// are never allowed in Gauntlet.
		humans := make([]ParticipantID, 0, len(g.order))
		for _, pid := range g.order {
			if g.participants[pid].Kind() == enums.Human {
				humans = append(humans, pid)
			} else {
				botsToDrop = append(botsToDrop, pid)
			}
		}
		if len(humans) > cfg.TeamSize() {
			return ErrConfigWouldEvictPlayers
		}
		for _, pid := range humans {
			plan[pid] = newTeams[0].ID()
		}
	}

	// All checks passed - apply. Bots are dropped via the ordinary Leave
	// path (no rng needed: a bot can never become host), then the new
	// Config/teams/stages replace the old ones and every remaining
	// participant is reseated per plan.
	for _, pid := range botsToDrop {
		delete(g.participants, pid)
		g.removeFromOrder(pid)
		g.emit(PlayerLeft{ParticipantID: pid})
	}

	g.config = cfg
	g.teams = append([]*Team(nil), newTeams...)
	g.stages = append([]Stage(nil), stages...)

	for _, pid := range g.order {
		p := g.participants[pid]
		teamID := plan[pid]
		p.setTeam(teamID)
		if t := g.teamByID(teamID); t != nil {
			t.AddMember(pid)
		}
	}

	g.emit(ConfigUpdated{})
	return nil
}

// --- Lifecycle ---

// Start moves the Game from LOBBY to ASSIGNING. Only the host may call it.
// Gauntlet needs at least one player; Versus needs both teams equal in size
// and non-empty - not necessarily full to Config.TeamSize() (that's only
// the per-team capacity ceiling enforced by addParticipant/SwitchTeam).
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
		first := g.teams[0].Size()
		for _, t := range g.teams {
			if t.Size() == 0 {
				return ErrNotEnoughPlayers
			}
			if t.Size() != first {
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
	// Computed once, outside the loop: a bot ballot never moves the
	// human-only counters (see humanVoteProgress), so every bot's VoteCast
	// emitted in this batch carries the same progress snapshot.
	cast, total := g.humanVoteProgress()
	for _, pid := range g.order {
		p := g.participants[pid]
		if p.Kind() != enums.Bot || !p.Connected() {
			continue
		}
		choice := voter.Vote(options, scores)
		if err := round.Ballot.Cast(pid, choice); err == nil {
			g.emit(VoteCast{
				RoundIndex: roundIndex, ParticipantID: pid, Option: choice,
				HumanVotesCast: cast, HumanVoters: total,
			})
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
	cast, total := g.humanVoteProgress()
	g.emit(VoteCast{
		RoundIndex: round.Index, ParticipantID: id, Option: o,
		HumanVotesCast: cast, HumanVoters: total,
	})
	return nil
}

// humanVoteProgress reports how many connected humans have voted in the
// current round (cast) out of how many are eligible to (total) - the single
// definition of "the population a voting window waits on". VotingComplete
// is exactly cast == total, and every VoteCast event's HumanVotesCast/
// HumanVoters is exactly this pair, so a progress bar built from the event
// can never disagree with the condition that closes the window. Bots are
// excluded even though they cast real ballots (see castBotVotes), as are
// disconnected humans (a disconnect counts as a null vote, never a
// blocker - see checkAbortConditions/Disconnect). Returns 0, 0 when there
// is no current round.
func (g *Game) humanVoteProgress() (cast, total int) {
	round := g.currentRound()
	if round == nil {
		return 0, 0
	}
	for _, pid := range g.order {
		p := g.participants[pid]
		if p == nil || p.Kind() != enums.Human || !p.Connected() {
			continue
		}
		total++
		if round.Ballot.HasVoted(pid) {
			cast++
		}
	}
	return cast, total
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
	if g.currentRound() == nil {
		return false
	}
	cast, total := g.humanVoteProgress()
	return cast == total
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
		// The revote must start from a genuinely empty ballot - without
		// this, every vote from the tied round would still stand, so the
		// window would open already reporting cast==total (VotingComplete
		// true) and the very first changed vote would close it instantly.
		// Bots are re-cast immediately, same as the first OpenVoting, so a
		// Versus revote isn't missing its bot votes.
		round.Ballot.Reset()
		g.emit(TiebreakOpened{RoundIndex: round.Index})
		g.castBotVotes(round.Index)
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
