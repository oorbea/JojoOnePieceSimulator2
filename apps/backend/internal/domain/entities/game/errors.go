package game

import "errors"

// Sentinel errors for the Game aggregate's cross-cutting behaviour
// (membership, state transitions, voting, bots). Entity-local validation
// errors (Stage, Loadout, Team, Config) live beside their own file,
// mirroring the convention in entities/user (ErrInvalidUsername) and
// entities/powers.
var (
	// ErrNotHost is returned when a non-host participant attempts a
	// host-only action (Start, Abort).
	ErrNotHost = errors.New("only the host may perform this action")

	// ErrInvalidStateTransition is returned when a method is called while
	// the Game is in a state that does not allow it.
	ErrInvalidStateTransition = errors.New("invalid game state transition")

	// ErrTeamSizeMismatch is returned when Versus teams do not have the
	// same number of players, both at Game construction and at Start.
	ErrTeamSizeMismatch = errors.New("versus teams must have the same number of players")

	// ErrTeamFull is returned when a Versus team already holds
	// Config.TeamSize() members.
	ErrTeamFull = errors.New("team is full")

	// ErrGameFull is returned when a Gauntlet's single team already holds
	// Config.TeamSize() members.
	ErrGameFull = errors.New("game is full")

	// ErrNotEnoughPlayers is returned when Start is called before every
	// team has at least one player.
	ErrNotEnoughPlayers = errors.New("not enough players to start")

	// ErrDuplicateParticipant is returned when a ParticipantID is already
	// present in the Game.
	ErrDuplicateParticipant = errors.New("participant already in game")

	// ErrParticipantNotFound is returned when a ParticipantID has no
	// matching Participant in the Game.
	ErrParticipantNotFound = errors.New("participant not found")

	// ErrTeamNotFound is returned when a Participant references a TeamID
	// that does not belong to this Game.
	ErrTeamNotFound = errors.New("team not found")

	// ErrVotingClosed is returned when CastVote or CloseVoting is called
	// outside the Voting/Tiebreak states, or when there is no current
	// round to vote on.
	ErrVotingClosed = errors.New("voting is not open")

	// ErrInvalidBallotOption is returned when a vote or a tiebreak
	// resolution references an OptionID that is not on the current
	// Ballot.
	ErrInvalidBallotOption = errors.New("invalid ballot option")

	// ErrNoStagesAvailable is returned when a Game (or a round within it)
	// has no Stage to play - an empty Gauntlet stage list at construction,
	// or an empty Versus stage pool when picking a round's Stage.
	ErrNoStagesAvailable = errors.New("no stages available for the selected mangas")

	// ErrBotsNotAllowed is returned when a bot is added to a Gauntlet
	// game, or to a Versus game whose Config.AllowBots() is false, or when
	// a bot Participant is used as a Game's host.
	ErrBotsNotAllowed = errors.New("bots are not allowed in this game mode or configuration")

	// ErrInventoryNotSupported is returned when a Config selects
	// enums.Inventory as its AbilitySource - there is no player inventory
	// yet (see ports.IInventory).
	ErrInventoryNotSupported = errors.New("inventory-based ability assignment is not supported yet")

	// ErrPowerPoolExhausted is returned when a team's AvailablePowers has
	// no more unique Stand/DevilFruit left to draw.
	ErrPowerPoolExhausted = errors.New("no more unique powers available for this team")
)
