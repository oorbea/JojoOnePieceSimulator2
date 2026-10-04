package enums

import "errors"

// PowerEffectKind is what a PowerEffect did to a Loadout during assignment:
// raised a stat up to a floor a power (or another stat) demands, or evolved
// the drawn Stand/DevilFruit into a later stage of its family. See
// game.PowerEffect.
type PowerEffectKind byte

const (
	EffectStatFloor PowerEffectKind = iota
	EffectEvolution
)

func (k PowerEffectKind) String() string {
	switch k {
	case EffectStatFloor:
		return "STAT_FLOOR"
	case EffectEvolution:
		return "EVOLUTION"
	default:
		return "UNKNOWN"
	}
}

var ErrInvalidPowerEffectKind = errors.New("invalid power effect kind")

func (k PowerEffectKind) IsValid() bool {
	switch k {
	case EffectStatFloor, EffectEvolution:
		return true
	default:
		return false
	}
}

func ParsePowerEffectKind(str string) (PowerEffectKind, error) {
	switch str {
	case "STAT_FLOOR":
		return EffectStatFloor, nil
	case "EVOLUTION":
		return EffectEvolution, nil
	default:
		return EffectStatFloor, ErrInvalidPowerEffectKind
	}
}
