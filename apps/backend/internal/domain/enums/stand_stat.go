package enums

import "errors"

type StandStat byte

const (
	E StandStat = iota
	D
	C
	B
	A
	Infinite
	Null
)

func (p StandStat) String() string {
	switch p {
	case E:
		return "E"
	case D:
		return "D"
	case C:
		return "C"
	case B:
		return "B"
	case A:
		return "A"
	case Infinite:
		return "INFINITE"
	case Null:
		return "NULL"
	default:
		return "UNKNOWN"
	}
}

var ErrInvalidStandStat = errors.New("invalid stand stat")

func (p StandStat) IsValid() bool {
	switch p {
	case E, D, C, B, A, Infinite, Null:
		return true
	default:
		return false
	}
}

func ParseStandStat(str string) (StandStat, error) {
	switch str {
	case "E":
		return E, nil
	case "D":
		return D, nil
	case "C":
		return C, nil
	case "B":
		return B, nil
	case "A":
		return A, nil
	case "INFINITE":
		return Infinite, nil
	case "NULL":
		return Null, nil
	default:
		return E, ErrInvalidStandStat
	}
}
