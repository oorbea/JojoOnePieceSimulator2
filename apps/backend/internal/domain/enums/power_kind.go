package enums

import "errors"

// PowerKind discriminates between the Power subtypes in the class-table
// inheritance hierarchy (Stand, DevilFruit, ...). It labels PictureJob so a
// single background worker can route a job to the right subtype repository.
type PowerKind byte

const (
	StandKind PowerKind = iota
	DevilFruitKind
)

func (k PowerKind) String() string {
	switch k {
	case StandKind:
		return "STAND"
	case DevilFruitKind:
		return "DEVIL_FRUIT"
	default:
		return "UNKNOWN"
	}
}

var ErrInvalidPowerKind = errors.New("invalid power kind")

func (k PowerKind) IsValid() bool {
	switch k {
	case StandKind, DevilFruitKind:
		return true
	default:
		return false
	}
}

func ParsePowerKind(str string) (PowerKind, error) {
	switch str {
	case "STAND":
		return StandKind, nil
	case "DEVIL_FRUIT":
		return DevilFruitKind, nil
	default:
		return StandKind, ErrInvalidPowerKind
	}
}
