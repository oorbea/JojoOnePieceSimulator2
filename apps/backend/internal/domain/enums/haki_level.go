package enums

import "errors"

// HakiLevel is the shared scale used by all three haki types (Armament,
// Observation, Conqueror). HakiNone means the player doesn't have that haki
// at all - distinct from HakiPrivate, the weakest mastery a haki you *do*
// have can show. This mirrors the original JoJoOnePiece_Simulator's model
// (github.com/oorbea/JoJoOnePiece_Simulator, powers.cc's generateHaki):
// whether a player has a given haki at all is drawn first (as a set, with
// correlations between the three - see game.AssignmentWeights.
// HakiSetWeights), then a mastery level is drawn independently for each
// haki present. Conqueror is drawn far less often than the other two - that
// skew lives in an IAssignmentWeights adapter, not in this type.
type HakiLevel byte

const (
	HakiNone HakiLevel = iota
	HakiPrivate
	HakiViceAdmiral
	HakiYonkoCommander
	HakiYonkoPlus
)

func (h HakiLevel) String() string {
	switch h {
	case HakiNone:
		return "NONE"
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
	case HakiNone, HakiPrivate, HakiViceAdmiral, HakiYonkoCommander, HakiYonkoPlus:
		return true
	default:
		return false
	}
}

func ParseHakiLevel(str string) (HakiLevel, error) {
	switch str {
	case "NONE":
		return HakiNone, nil
	case "PRIVATE":
		return HakiPrivate, nil
	case "VICE_ADMIRAL":
		return HakiViceAdmiral, nil
	case "YONKO_COMMANDER":
		return HakiYonkoCommander, nil
	case "YONKO_PLUS":
		return HakiYonkoPlus, nil
	default:
		return HakiNone, ErrInvalidHakiLevel
	}
}
