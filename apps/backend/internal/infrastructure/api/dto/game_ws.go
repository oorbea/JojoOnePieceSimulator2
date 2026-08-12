package dto

import "encoding/json"

// ClientCommand is the envelope every game WebSocket command arrives in.
// RequestID is opaque to the server - it exists purely so a client can
// correlate an ERROR frame with the optimistic action that caused it.
type ClientCommand struct {
	Type      string          `json:"type"`
	RequestID string          `json:"requestId,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

// Client command types. Each maps to exactly one GameService method; see
// game_ws_endpoints.go's dispatch. CreateGame/JoinByCode/GetGame/
// GetGameByCode/GameCode are deliberately NOT commands (they're plain HTTP -
// you can't join a game over a socket you can only open by already being a
// participant), and Disconnect/Reconnect are deliberately NOT commands
// either (lifecycle-only, driven by the socket's own open/close, so a
// client can never forge its own or anyone else's presence).
const (
	CommandLeave     = "LEAVE"
	CommandAddBot    = "ADD_BOT"
	CommandRemoveBot = "REMOVE_BOT"
	CommandStart     = "START"
	CommandAbort     = "ABORT"
	CommandVote      = "VOTE"
	CommandResync    = "RESYNC"
)

// AddBotPayload is CommandAddBot's payload.
type AddBotPayload struct {
	TeamID string `json:"teamId"`
}

// RemoveBotPayload is CommandRemoveBot's payload.
type RemoveBotPayload struct {
	BotID string `json:"botId"`
}

// VotePayload is CommandVote's payload.
type VotePayload struct {
	Option string `json:"option"`
}

// ServerFrame is the envelope every server->client frame arrives in.
// RequestID is only set on an ERROR frame replying to a specific command.
type ServerFrame struct {
	Type      string `json:"type"`
	RequestID string `json:"requestId,omitempty"`
	Payload   any    `json:"payload,omitempty"`
}

// Server frame types. The eleven domain-event names are reused verbatim
// from entities/game/events.go's DomainEvent.Name() - the transport never
// invents its own vocabulary for them.
const (
	FrameState            = "STATE"
	FramePlayerJoined     = "PLAYER_JOINED"
	FramePlayerLeft       = "PLAYER_LEFT"
	FrameHostReassigned   = "HOST_REASSIGNED"
	FrameGameStarted      = "GAME_STARTED"
	FrameLoadoutsAssigned = "LOADOUTS_ASSIGNED"
	FrameVotingOpened     = "VOTING_OPENED"
	FrameVoteCast         = "VOTE_CAST"
	FrameTiebreakOpened   = "TIEBREAK_OPENED"
	FrameRoundResolved    = "ROUND_RESOLVED"
	FrameGameFinished     = "GAME_FINISHED"
	FrameGameAborted      = "GAME_ABORTED"
	FrameError            = "ERROR"
	FrameResyncRequired   = "RESYNC_REQUIRED"
)

// VotingOpenedPayload/TiebreakOpenedPayload carry a transport-computed
// closesAt: the domain event itself carries no deadline (the voting window
// lives only in GameService's timer), so the transport approximates it as
// time.Now() + the configured voting window at the moment it forwards the
// event - a few ms of skew against the real timer, acceptable for a
// countdown UI.
type VotingOpenedPayload struct {
	RoundIndex int    `json:"roundIndex"`
	ClosesAt   string `json:"closesAt"`
}

type TiebreakOpenedPayload struct {
	RoundIndex int    `json:"roundIndex"`
	ClosesAt   string `json:"closesAt"`
}

// VoteCastPayload is deliberately anonymized: it carries only that a vote
// was cast, not who cast it or for what option, so the transport doesn't
// leak a live round's votes ahead of GameStateResponse's own votes-hidden-
// until-resolved rule.
type VoteCastPayload struct {
	RoundIndex int `json:"roundIndex"`
	VotesCast  int `json:"votesCast"`
}

type PlayerJoinedPayload struct {
	ParticipantID string `json:"participantId"`
}

type PlayerLeftPayload struct {
	ParticipantID string `json:"participantId"`
}

type HostReassignedPayload struct {
	NewHostID string `json:"newHostId"`
}

type LoadoutsAssignedPayload struct {
	RoundIndex int `json:"roundIndex"`
}

type RoundResolvedPayload struct {
	RoundIndex        int    `json:"roundIndex"`
	Winner            string `json:"winner"`
	DecidedByCoinFlip bool   `json:"decidedByCoinFlip"`
}

type GameFinishedPayload struct {
	Result GameResultResponse `json:"result"`
}

type GameAbortedPayload struct {
	Reason string `json:"reason"`
}

type ResyncRequiredPayload struct {
	Reason string `json:"reason"`
}
