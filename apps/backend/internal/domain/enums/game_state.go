package enums

import "errors"

// GameState is the finite set of states a Game aggregate can be in - the
// State pattern driving *game.Game's transitions.
type GameState byte

const (
	Lobby GameState = iota
	Assigning
	Summary
	Voting
	Tiebreak
	Resolving
	Finished
	Aborted
)

func (s GameState) String() string {
	switch s {
	case Lobby:
		return "LOBBY"
	case Assigning:
		return "ASSIGNING"
	case Summary:
		return "SUMMARY"
	case Voting:
		return "VOTING"
	case Tiebreak:
		return "TIEBREAK"
	case Resolving:
		return "RESOLVING"
	case Finished:
		return "FINISHED"
	case Aborted:
		return "ABORTED"
	default:
		return "UNKNOWN"
	}
}

var ErrInvalidGameState = errors.New("invalid game state")

func (s GameState) IsValid() bool {
	switch s {
	case Lobby, Assigning, Summary, Voting, Tiebreak, Resolving, Finished, Aborted:
		return true
	default:
		return false
	}
}

func ParseGameState(str string) (GameState, error) {
	switch str {
	case "LOBBY":
		return Lobby, nil
	case "ASSIGNING":
		return Assigning, nil
	case "SUMMARY":
		return Summary, nil
	case "VOTING":
		return Voting, nil
	case "TIEBREAK":
		return Tiebreak, nil
	case "RESOLVING":
		return Resolving, nil
	case "FINISHED":
		return Finished, nil
	case "ABORTED":
		return Aborted, nil
	default:
		return Lobby, ErrInvalidGameState
	}
}
