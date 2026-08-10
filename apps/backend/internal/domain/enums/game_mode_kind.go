package enums

import "errors"

// GameModeKind selects which IGameMode strategy a Game runs: cooperative
// Gauntlet or 2-team Versus.
type GameModeKind byte

const (
	Gauntlet GameModeKind = iota
	Versus
)

func (k GameModeKind) String() string {
	switch k {
	case Gauntlet:
		return "GAUNTLET"
	case Versus:
		return "VERSUS"
	default:
		return "UNKNOWN"
	}
}

var ErrInvalidGameModeKind = errors.New("invalid game mode kind")

func (k GameModeKind) IsValid() bool {
	switch k {
	case Gauntlet, Versus:
		return true
	default:
		return false
	}
}

func ParseGameModeKind(str string) (GameModeKind, error) {
	switch str {
	case "GAUNTLET":
		return Gauntlet, nil
	case "VERSUS":
		return Versus, nil
	default:
		return Gauntlet, ErrInvalidGameModeKind
	}
}
