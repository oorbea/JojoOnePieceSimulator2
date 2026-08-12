package enums

import "errors"

// SpinLevel is a player's Ripple/Spin mastery, independent of whether they
// also have a Stand - except for the handful of Stands whose PowerTrait is
// RequiresSpin4 (Tusk ACT4, Ball Breaker, Soft & Wet: Go Beyond), which
// force SpinInfinite.
type SpinLevel byte

const (
	SpinNone SpinLevel = iota
	SpinBasic
	SpinAdvanced
	SpinGolden
	SpinInfinite
)

func (s SpinLevel) String() string {
	switch s {
	case SpinNone:
		return "NONE"
	case SpinBasic:
		return "BASIC"
	case SpinAdvanced:
		return "ADVANCED"
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
	case SpinNone, SpinBasic, SpinAdvanced, SpinGolden, SpinInfinite:
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
	case "ADVANCED":
		return SpinAdvanced, nil
	case "GOLDEN":
		return SpinGolden, nil
	case "INFINITE":
		return SpinInfinite, nil
	default:
		return SpinNone, ErrInvalidSpinLevel
	}
}
