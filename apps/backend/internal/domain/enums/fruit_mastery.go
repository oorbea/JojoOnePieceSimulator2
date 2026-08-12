package enums

import "errors"

// FruitMastery is how well a player controls their Devil Fruit. It is
// coupled to whether the player has a fruit at all: no fruit forces
// FruitMasteryNone, and any fruit forces at least FruitMasteryRegular - see
// game.NewLoadout.
type FruitMastery byte

const (
	FruitMasteryNone FruitMastery = iota
	FruitMasteryRegular
	FruitMasteryAdvanced
	FruitMasteryAwakened
)

func (m FruitMastery) String() string {
	switch m {
	case FruitMasteryNone:
		return "NONE"
	case FruitMasteryRegular:
		return "REGULAR"
	case FruitMasteryAdvanced:
		return "ADVANCED"
	case FruitMasteryAwakened:
		return "AWAKENED"
	default:
		return "UNKNOWN"
	}
}

var ErrInvalidFruitMastery = errors.New("invalid fruit mastery")

func (m FruitMastery) IsValid() bool {
	switch m {
	case FruitMasteryNone, FruitMasteryRegular, FruitMasteryAdvanced, FruitMasteryAwakened:
		return true
	default:
		return false
	}
}

func ParseFruitMastery(str string) (FruitMastery, error) {
	switch str {
	case "NONE":
		return FruitMasteryNone, nil
	case "REGULAR":
		return FruitMasteryRegular, nil
	case "ADVANCED":
		return FruitMasteryAdvanced, nil
	case "AWAKENED":
		return FruitMasteryAwakened, nil
	default:
		return FruitMasteryNone, ErrInvalidFruitMastery
	}
}
