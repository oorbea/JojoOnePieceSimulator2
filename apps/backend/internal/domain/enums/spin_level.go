package enums

import "errors"

// SpinLevel is a player's Ripple/Spin mastery, independent of whether they
// also have a Stand - except for the handful of Stands whose PowerTrait is
// RequiresSpin4 (Tusk ACT4, Ball Breaker, Soft & Wet: Go Beyond), which
// force SpinInfinite. Four levels, matching the original JoJoOnePiece_
// Simulator's spin table (github.com/oorbea/JoJoOnePiece_Simulator,
// powers.cc's generateSpin) - there is no ADVANCED tier.
type SpinLevel byte

const (
	SpinNone SpinLevel = iota
	SpinBasic
	SpinGolden
	SpinInfinite
)

func (s SpinLevel) String() string {
	switch s {
	case SpinNone:
		return "NONE"
	case SpinBasic:
		return "BASIC"
	case SpinGolden:
		return "GOLDEN"
	case SpinInfinite:
		return "INFINITE"
	default:
		return "UNKNOWN"
	}
}

var ErrInvalidSpinLevel = errors.New("invalid spin level")

func (s SpinLevel) IsValid() bool {
	switch s {
	case SpinNone, SpinBasic, SpinGolden, SpinInfinite:
		return true
	default:
		return false
	}
}

func ParseSpinLevel(str string) (SpinLevel, error) {
	switch str {
	case "NONE":
		return SpinNone, nil
	case "BASIC":
		return SpinBasic, nil
	case "GOLDEN":
		return SpinGolden, nil
	case "INFINITE":
		return SpinInfinite, nil
	default:
		return SpinNone, ErrInvalidSpinLevel
	}
}
