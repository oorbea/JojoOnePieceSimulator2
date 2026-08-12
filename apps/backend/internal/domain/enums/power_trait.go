package enums

import "errors"

// PowerTrait marks a non-obvious rule a specific Power's assignment must
// obey. Today there is a single trait: some Stands force the player's Spin
// to Infinite. Traits are derived from a Power's name in
// game.TraitsOf until powers gains a persisted traits column.
type PowerTrait byte

const (
	// RequiresSpin4 forces SpinInfinite whenever the Stand carrying this
	// trait is assigned (Tusk ACT4, Ball Breaker, Soft & Wet: Go Beyond).
	RequiresSpin4 PowerTrait = iota
)

func (t PowerTrait) String() string {
	switch t {
	case RequiresSpin4:
		return "REQUIRES_SPIN_4"
	default:
		return "UNKNOWN"
	}
}

var ErrInvalidPowerTrait = errors.New("invalid power trait")

func (t PowerTrait) IsValid() bool {
	switch t {
	case RequiresSpin4:
		return true
	default:
		return false
	}
}

func ParsePowerTrait(str string) (PowerTrait, error) {
	switch str {
	case "REQUIRES_SPIN_4":
		return RequiresSpin4, nil
	default:
		return RequiresSpin4, ErrInvalidPowerTrait
	}
}
