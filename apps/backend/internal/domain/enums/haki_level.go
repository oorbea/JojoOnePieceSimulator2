package enums

import "errors"

// HakiLevel is the shared 0-3 scale used by all three haki types
// (Armament, Observation, Conqueror). Conqueror is drawn far less often
// than the other two - that skew lives in an IAssignmentWeights adapter,
// not in this type.
type HakiLevel byte

const (
	HakiPrivate HakiLevel = iota
	HakiViceAdmiral
	HakiYonkoCommander
	HakiYonkoPlus
)

func (h HakiLevel) String() string {
	switch h {
	case HakiPrivate:
		return "PRIVATE"
	case HakiViceAdmiral:
		return "VICE_ADMIRAL"
	case HakiYonkoCommander:
		return "YONKO_COMMANDER"
	case HakiYonkoPlus:
		return "YONKO_PLUS"
	default:
		return "UNKNOWN"
	}
}

var ErrInvalidHakiLevel = errors.New("invalid haki level")

func (h HakiLevel) IsValid() bool {
	switch h {
	case HakiPrivate, HakiViceAdmiral, HakiYonkoCommander, HakiYonkoPlus:
		return true
	default:
		return false
	}
}

func ParseHakiLevel(str string) (HakiLevel, error) {
	switch str {
	case "PRIVATE":
		return HakiPrivate, nil
	case "VICE_ADMIRAL":
		return HakiViceAdmiral, nil
	case "YONKO_COMMANDER":
		return HakiYonkoCommander, nil
	case "YONKO_PLUS":
		return HakiYonkoPlus, nil
	default:
		return HakiPrivate, ErrInvalidHakiLevel
	}
}
