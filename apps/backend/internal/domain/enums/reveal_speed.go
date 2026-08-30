package enums

import "errors"

// RevealSpeed controls how fast a lobby's sorteo (the ASSIGNING-state
// reveal) plays out, as a multiplier applied uniformly to every timing
// constant in entities/game/reveal.go. Swift(1.0) reproduces the original
// terminal-based JoJoOnePiece_Simulator's loadingScreen/delay pacing
// exactly; Normal(1.3, the default) and Relaxed(1.6) slow it down so the
// roulette animation and the big power-reveal card have more time to
// register (owner request, 2026-08-30 - see
// ObsidianVault/game-match-assignment-frontend.md).
type RevealSpeed byte

// Normal is deliberately the zero value so an unset RevealSpeed (a legacy
// Snapshot, an omitted field in a CreateGame/ConfigUpdate request) defaults
// to Normal without any extra fallback logic - the same convention as
// LobbyVisibility's zero-value Private.
const (
	Normal RevealSpeed = iota
	Relaxed
	Swift
)

func (s RevealSpeed) String() string {
	switch s {
	case Relaxed:
		return "RELAXED"
	case Normal:
		return "NORMAL"
	case Swift:
		return "SWIFT"
	default:
		return "UNKNOWN"
	}
}

var ErrInvalidRevealSpeed = errors.New("invalid reveal speed")

func (s RevealSpeed) IsValid() bool {
	switch s {
	case Relaxed, Normal, Swift:
		return true
	default:
		return false
	}
}

// Multiplier returns the factor RevealSpeed applies to every reveal timing
// constant. Kept as a method (not a lookup table in game.RevealDuration)
// so the enum and its pacing stay declared in one place.
func (s RevealSpeed) Multiplier() float64 {
	switch s {
	case Relaxed:
		return 1.6
	case Swift:
		return 1.0
	case Normal:
		return 1.3
	default:
		return 1.3
	}
}

func ParseRevealSpeed(str string) (RevealSpeed, error) {
	switch str {
	case "RELAXED":
		return Relaxed, nil
	case "NORMAL":
		return Normal, nil
	case "SWIFT":
		return Swift, nil
	default:
		return Normal, ErrInvalidRevealSpeed
	}
}
