package enums

import "errors"

type FruitType byte

const (
	Paramecia FruitType = iota
	Zoan
	Logia
	SpecialParamecia
	AncientZoan
	MythicalZoan
)

func (ft FruitType) String() string {
	switch ft {
	case Paramecia:
		return "PARAMECIA"
	case Zoan:
		return "ZOAN"
	case Logia:
		return "LOGIA"
	case SpecialParamecia:
		return "SPECIAL_PARAMECIA"
	case AncientZoan:
		return "ANCIENT_ZOAN"
	case MythicalZoan:
		return "MYTHICAL_ZOAN"
	default:
		return "UNKNOWN"
	}
}

var ErrInvalidFruitType = errors.New("invalid devil fruit type")

func (ft FruitType) IsValid() bool {
	switch ft {
	case Paramecia, Zoan, Logia, SpecialParamecia, AncientZoan, MythicalZoan:
		return true
	default:
		return false
	}
}

func ParseFruitType(str string) (FruitType, error) {
	switch str {
	case "PARAMECIA":
		return Paramecia, nil
	case "ZOAN":
		return Zoan, nil
	case "LOGIA":
		return Logia, nil
	case "SPECIAL_PARAMECIA":
		return SpecialParamecia, nil
	case "ANCIENT_ZOAN":
		return AncientZoan, nil
	case "MYTHICAL_ZOAN":
		return MythicalZoan, nil
	default:
		return Paramecia, ErrInvalidFruitType
	}
}
