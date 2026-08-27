package enums

import "errors"

// PhysicalForm is a player's raw physical conditioning tier on the One
// Piece side. Six levels, matching the original JoJoOnePiece_Simulator's
// strength table (github.com/oorbea/JoJoOnePiece_Simulator, powers.cc's
// generateStrength), unlike HakiLevel's 4-level scale - the two are kept as
// distinct types since they're conceptually distinct stats.
type PhysicalForm byte

const (
	PhysicalFormPrivate PhysicalForm = iota
	PhysicalFormStrongFishman
	PhysicalFormMarineCaptain
	PhysicalFormViceAdmiral
	PhysicalFormYonkoCommander
	PhysicalFormYonkoPlus
)

func (f PhysicalForm) String() string {
	switch f {
	case PhysicalFormPrivate:
		return "PRIVATE"
	case PhysicalFormStrongFishman:
		return "STRONG_FISHMAN"
	case PhysicalFormMarineCaptain:
		return "MARINE_CAPTAIN"
	case PhysicalFormViceAdmiral:
		return "VICE_ADMIRAL"
	case PhysicalFormYonkoCommander:
		return "YONKO_COMMANDER"
	case PhysicalFormYonkoPlus:
		return "YONKO_PLUS"
	default:
		return "UNKNOWN"
	}
}

var ErrInvalidPhysicalForm = errors.New("invalid physical form")

func (f PhysicalForm) IsValid() bool {
	switch f {
	case PhysicalFormPrivate, PhysicalFormStrongFishman, PhysicalFormMarineCaptain,
		PhysicalFormViceAdmiral, PhysicalFormYonkoCommander, PhysicalFormYonkoPlus:
		return true
	default:
		return false
	}
}

func ParsePhysicalForm(str string) (PhysicalForm, error) {
	switch str {
	case "PRIVATE":
		return PhysicalFormPrivate, nil
	case "STRONG_FISHMAN":
		return PhysicalFormStrongFishman, nil
	case "MARINE_CAPTAIN":
		return PhysicalFormMarineCaptain, nil
	case "VICE_ADMIRAL":
		return PhysicalFormViceAdmiral, nil
	case "YONKO_COMMANDER":
		return PhysicalFormYonkoCommander, nil
	case "YONKO_PLUS":
		return PhysicalFormYonkoPlus, nil
	default:
		return PhysicalFormPrivate, ErrInvalidPhysicalForm
	}
}
