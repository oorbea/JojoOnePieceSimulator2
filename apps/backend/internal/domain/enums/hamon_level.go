package enums

import "errors"

// HamonLevel is a player's Hamon mastery, independent of Stand and Spin.
type HamonLevel byte

const (
	HamonNone HamonLevel = iota
	HamonBasic
	HamonAdvanced
	HamonPerfect
)

func (h HamonLevel) String() string {
	switch h {
	case HamonNone:
		return "NONE"
	case HamonBasic:
		return "BASIC"
	case HamonAdvanced:
		return "ADVANCED"
	case HamonPerfect:
		return "PERFECT"
	default:
		return "UNKNOWN"
	}
}

var ErrInvalidHamonLevel = errors.New("invalid hamon level")

func (h HamonLevel) IsValid() bool {
	switch h {
	case HamonNone, HamonBasic, HamonAdvanced, HamonPerfect:
		return true
	default:
		return false
	}
}

func ParseHamonLevel(str string) (HamonLevel, error) {
	switch str {
	case "NONE":
		return HamonNone, nil
	case "BASIC":
		return HamonBasic, nil
	case "ADVANCED":
		return HamonAdvanced, nil
	case "PERFECT":
		return HamonPerfect, nil
	default:
		return HamonNone, ErrInvalidHamonLevel
	}
}
