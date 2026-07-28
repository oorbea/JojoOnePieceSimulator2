package enums

import "errors"

type PowerRarity byte

const (
	Common PowerRarity = iota
	Rare
	Epic
	Legendary
)

func (p PowerRarity) String() string {
	switch p {
	case Common:
		return "COMMON"
	case Rare:
		return "RARE"
	case Epic:
		return "EPIC"
	case Legendary:
		return "LEGENDARY"
	default:
		return "UNKNOWN"
	}
}

var ErrInvalidRarity = errors.New("invalid power rarity")

func (p PowerRarity) IsValid() bool {
	switch p {
	case Common, Rare, Epic, Legendary:
		return true
	default:
		return false
	}
}

func ParsePowerRarity(str string) (PowerRarity, error) {
	switch str {
	case "COMMON":
		return Common, nil
	case "RARE":
		return Rare, nil
	case "EPIC":
		return Epic, nil
	case "LEGENDARY":
		return Legendary, nil
	default:
		return Common, ErrInvalidRarity
	}
}
