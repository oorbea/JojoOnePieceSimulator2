package enums

import "errors"

// SquadVerdict is the ballot option a Gauntlet participant casts each round:
// whether the squad survives the round's Stage or falls to it.
type SquadVerdict byte

const (
	Survive SquadVerdict = iota
	Fall
)

func (v SquadVerdict) String() string {
	switch v {
	case Survive:
		return "SURVIVE"
	case Fall:
		return "FALL"
	default:
		return "UNKNOWN"
	}
}

var ErrInvalidSquadVerdict = errors.New("invalid squad verdict")

func (v SquadVerdict) IsValid() bool {
	switch v {
	case Survive, Fall:
		return true
	default:
		return false
	}
}

func ParseSquadVerdict(str string) (SquadVerdict, error) {
	switch str {
	case "SURVIVE":
		return Survive, nil
	case "FALL":
		return Fall, nil
	default:
		return Survive, ErrInvalidSquadVerdict
	}
}
