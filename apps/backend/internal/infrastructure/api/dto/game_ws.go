package dto

import (
	"encoding/json"
	"fmt"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

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
	CommandLeave        = "LEAVE"
	CommandAddBot       = "ADD_BOT"
	CommandRemoveBot    = "REMOVE_BOT"
	CommandStart        = "START"
	CommandAbort        = "ABORT"
	CommandVote         = "VOTE"
	CommandResync       = "RESYNC"
	CommandSwitchTeam   = "SWITCH_TEAM"
	CommandMovePlayer   = "MOVE_PLAYER"
	CommandKick         = "KICK"
	CommandTransferHost = "TRANSFER_HOST"
	CommandSetLock      = "SET_LOCK"
	CommandUpdateConfig = "UPDATE_CONFIG"
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

// SwitchTeamPayload is CommandSwitchTeam's payload. ParticipantID is
// optional - an empty value means "move myself".
type SwitchTeamPayload struct {
	ParticipantID string `json:"participantId,omitempty"`
	TeamID        string `json:"teamId"`
}

// MovePlayerPayload is CommandMovePlayer's payload - the host moving
// someone else. Identical shape to SwitchTeamPayload with ParticipantID
// required instead of optional; kept as its own command so a client never
// has to omit a field to express "move myself" vs "move them".
type MovePlayerPayload struct {
	ParticipantID string `json:"participantId"`
	TeamID        string `json:"teamId"`
}

// KickPayload is CommandKick's payload.
type KickPayload struct {
	ParticipantID string `json:"participantId"`
}

// TransferHostPayload is CommandTransferHost's payload.
type TransferHostPayload struct {
	ParticipantID string `json:"participantId"`
}

// SetLockPayload is CommandSetLock's payload.
type SetLockPayload struct {
	Locked bool `json:"locked"`
}

// PoolFilterPayload is the wire form of game.PoolFilter, shared by
// CreateGameRequest and CommandUpdateConfig.
type PoolFilterPayload struct {
	StandRarities []string `json:"standRarities,omitempty"`
	FruitRarities []string `json:"fruitRarities,omitempty"`
	FruitTypes    []string `json:"fruitTypes,omitempty"`
	Banned        []string `json:"banned,omitempty"`
}

// ToPoolFilter validates p (which may be nil, meaning "no restriction")
// into a game.PoolFilter, collecting field errors under the given prefix.
func (p *PoolFilterPayload) ToPoolFilter(errs *[]string) game.PoolFilter {
	if p == nil {
		return game.PoolFilter{}
	}
	standRarities := parseRarities(p.StandRarities, "poolFilter.standRarities", errs)
	fruitRarities := parseRarities(p.FruitRarities, "poolFilter.fruitRarities", errs)
	fruitTypes := make([]enums.FruitType, 0, len(p.FruitTypes))
	for _, raw := range p.FruitTypes {
		t, err := enums.ParseFruitType(raw)
		if err != nil {
			*errs = append(*errs, fmt.Sprintf("poolFilter.fruitTypes: %v", err))
			continue
		}
		fruitTypes = append(fruitTypes, t)
	}
	banned := make([]powers.PowerID, 0, len(p.Banned))
	for _, raw := range p.Banned {
		id, err := powers.ParsePowerID(raw)
		if err != nil {
			*errs = append(*errs, fmt.Sprintf("poolFilter.banned: %v", err))
			continue
		}
		banned = append(banned, id)
	}
	filter, err := game.NewPoolFilter(standRarities, fruitRarities, fruitTypes, banned)
	if err != nil {
		*errs = append(*errs, fmt.Sprintf("poolFilter: %v", err))
	}
	return filter
}

func parseRarities(raw []string, field string, errs *[]string) []enums.PowerRarity {
	out := make([]enums.PowerRarity, 0, len(raw))
	for _, r := range raw {
		parsed, err := enums.ParsePowerRarity(r)
		if err != nil {
			*errs = append(*errs, fmt.Sprintf("%s: %v", field, err))
			continue
		}
		out = append(out, parsed)
	}
	return out
}

// UpdateConfigPayload is CommandUpdateConfig's payload and the JSON body
// accepted by the config-edit REST path - a full replacement of the
// lobby's Config, the same shape CreateGameRequest builds a Config from
// plus the fields the host can only set once a lobby exists.
type UpdateConfigPayload struct {
	Mode                string             `json:"mode"`
	Mangas              []string           `json:"mangas"`
	AbilitySource       string             `json:"abilitySource"`
	TeamSize            int                `json:"teamSize"`
	AllowBots           bool               `json:"allowBots"`
	Visibility          string             `json:"visibility"`
	VotingWindowSeconds int                `json:"votingWindowSeconds"`
	PoolFilter          *PoolFilterPayload `json:"poolFilter,omitempty"`
}

// Validate converts the payload into a services.ConfigUpdateInput,
// collecting all field errors before returning.
func (p UpdateConfigPayload) Validate() (services.ConfigUpdateInput, error) {
	var errs []string

	mode, err := enums.ParseGameModeKind(p.Mode)
	if err != nil {
		errs = append(errs, fmt.Sprintf("mode: %v", err))
	}
	mangas := make([]enums.Manga, 0, len(p.Mangas))
	for _, raw := range p.Mangas {
		m, err := enums.ParseManga(raw)
		if err != nil {
			errs = append(errs, fmt.Sprintf("mangas: %v", err))
			continue
		}
		mangas = append(mangas, m)
	}
	abilitySource, err := enums.ParseAbilitySource(p.AbilitySource)
	if err != nil {
		errs = append(errs, fmt.Sprintf("abilitySource: %v", err))
	}
	if p.TeamSize <= 0 {
		errs = append(errs, "teamSize must be positive")
	}
	visibility, err := enums.ParseLobbyVisibility(p.Visibility)
	if err != nil {
		errs = append(errs, fmt.Sprintf("visibility: %v", err))
	}
	poolFilter := p.PoolFilter.ToPoolFilter(&errs)

	if len(errs) > 0 {
		return services.ConfigUpdateInput{}, &ValidationError{Errors: errs}
	}

	return services.ConfigUpdateInput{
		Mode:                mode,
		Mangas:              mangas,
		AbilitySource:       abilitySource,
		TeamSize:            p.TeamSize,
		AllowBots:           p.AllowBots,
		Visibility:          visibility,
		VotingWindowSeconds: p.VotingWindowSeconds,
		PoolFilter:          poolFilter,
	}, nil
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
	FrameTeamChanged      = "TEAM_CHANGED"
	FramePlayerKicked     = "PLAYER_KICKED"
	FrameLobbyLockChanged = "LOBBY_LOCK_CHANGED"
	FrameConfigUpdated    = "CONFIG_UPDATED"
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

// VoteCastPayload carries the round's human-vote progress and nothing else:
// no participant, no option, even though the domain event has both -
// GameStateResponse's own votes-hidden-until-resolved rule still holds.
// VotesCast/Voters count connected humans only (bots vote instantly when a
// window opens and are never waited on - see game.Game.humanVoteProgress),
// so votesCast == voters is precisely the condition that closes the window
// early. Both are aggregates over the whole lobby, so neither reveals who
// voted or what they voted for.
type VoteCastPayload struct {
	RoundIndex int `json:"roundIndex"`
	VotesCast  int `json:"votesCast"`
	Voters     int `json:"voters"`
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

// LoadoutsAssignedPayload carries revealMs - the transport-computed
// duration (game.RevealDuration for the lobby's own mangas) every client
// should spend on its reveal overlay before voting can open. See
// GameService.scheduleRevealDelay: the server itself waits exactly this
// long before calling OpenVoting, so a client that paces its animation to
// anything else either finishes early (and just waits) or late (and misses
// nothing, since voting genuinely hasn't opened yet).
type LoadoutsAssignedPayload struct {
	RoundIndex int   `json:"roundIndex"`
	RevealMs   int64 `json:"revealMs"`
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

type TeamChangedPayload struct {
	ParticipantID string `json:"participantId"`
	FromTeamID    string `json:"fromTeamId"`
	ToTeamID      string `json:"toTeamId"`
}

type PlayerKickedPayload struct {
	ParticipantID string `json:"participantId"`
}

type LobbyLockChangedPayload struct {
	Locked bool `json:"locked"`
}
