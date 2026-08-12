package enums

import "errors"

// PhysicalForm is a player's raw physical conditioning tier on the One
// Piece side, on the same 0-3 scale as HakiLevel but kept as its own type
// since the two are conceptually distinct stats.
type PhysicalForm byte

const (
	PhysicalFormPrivate PhysicalForm = iota
	PhysicalFormViceAdmiral
	PhysicalFormYonkoCommander
	PhysicalFormYonkoPlus
)

func (f PhysicalForm) String() string {
	switch f {
	case PhysicalFormPrivate:
		return "PRIVATE"
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
	case PhysicalFormPrivate, PhysicalFormViceAdmiral, PhysicalFormYonkoCommander, PhysicalFormYonkoPlus:
		return true
	default:
		return false
	}
}

func ParsePhysicalForm(str string) (PhysicalForm, error) {
	switch str {
	case "PRIVATE":
		return PhysicalFormPrivate, nil
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
