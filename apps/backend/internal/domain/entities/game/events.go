package game

// DomainEvent is emitted by Game as its state changes. The application
// layer drains them via Game.PullEvents to publish over websockets/SSE and
// to feed ports.IGameHistory - the same pub/sub role as
// application/services.PictureEventHub, but sourced from the aggregate
// itself instead of a background worker.
type DomainEvent interface {
	Name() string
}

type PlayerJoined struct{ ParticipantID ParticipantID }

func (PlayerJoined) Name() string { return "PLAYER_JOINED" }

type PlayerLeft struct{ ParticipantID ParticipantID }

func (PlayerLeft) Name() string { return "PLAYER_LEFT" }

type HostReassigned struct{ NewHostID ParticipantID }

func (HostReassigned) Name() string { return "HOST_REASSIGNED" }

type GameStarted struct{}

func (GameStarted) Name() string { return "GAME_STARTED" }

type LoadoutsAssigned struct{ RoundIndex int }

func (LoadoutsAssigned) Name() string { return "LOADOUTS_ASSIGNED" }

type VotingOpened struct{ RoundIndex int }

func (VotingOpened) Name() string { return "VOTING_OPENED" }

type VoteCast struct {
	RoundIndex    int
	ParticipantID ParticipantID
	Option        OptionID
}

func (VoteCast) Name() string { return "VOTE_CAST" }

type TiebreakOpened struct{ RoundIndex int }

func (TiebreakOpened) Name() string { return "TIEBREAK_OPENED" }

type RoundResolved struct {
	RoundIndex        int
	Winner            OptionID
	DecidedByCoinFlip bool
}

func (RoundResolved) Name() string { return "ROUND_RESOLVED" }

type GameFinished struct{ Result GameResult }

func (GameFinished) Name() string { return "GAME_FINISHED" }

type GameAborted struct{ Reason string }

func (GameAborted) Name() string { return "GAME_ABORTED" }

// TeamChanged is emitted by SwitchTeam whenever a participant actually
// moves teams (not on the already-there no-op).
type TeamChanged struct {
	ParticipantID ParticipantID
	FromTeamID    TeamID
	ToTeamID      TeamID
}

func (TeamChanged) Name() string { return "TEAM_CHANGED" }

// PlayerKicked is emitted by Kick, before the follow-on PlayerLeft, so the
// transport can tell an involuntary removal apart from a voluntary Leave
// and close the victim's socket instead of letting it dangle.
type PlayerKicked struct{ ParticipantID ParticipantID }

func (PlayerKicked) Name() string { return "PLAYER_KICKED" }

// LobbyLockChanged is emitted by SetLocked whenever the lock actually
// toggles (not on a same-value no-op).
type LobbyLockChanged struct{ Locked bool }

func (LobbyLockChanged) Name() string { return "LOBBY_LOCK_CHANGED" }

// ConfigUpdated is emitted by Reconfigure. It carries no fields - the
// mandatory follow-up STATE frame is what the transport uses to deliver the
// new configuration, exactly like every other structural change.
type ConfigUpdated struct{}

func (ConfigUpdated) Name() string { return "CONFIG_UPDATED" }
