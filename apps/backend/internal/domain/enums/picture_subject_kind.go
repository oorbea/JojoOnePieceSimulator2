package enums

import "errors"

// PictureSubjectKind labels a PictureJob with which kind of subject it
// belongs to, so PictureWorker can route it to the right PictureTarget. It
// is deliberately separate from PowerKind: PowerKind is persisted as the
// `power_kind` column discriminating the Stand/DevilFruit class-table
// hierarchy, while PictureSubjectKind is an in-memory-only routing label
// that also covers non-Power subjects (User avatars).
type PictureSubjectKind byte

const (
	StandSubject PictureSubjectKind = iota
	DevilFruitSubject
	UserSubject
	StageSubject
)

func (k PictureSubjectKind) String() string {
	switch k {
	case StandSubject:
		return "STAND"
	case DevilFruitSubject:
		return "DEVIL_FRUIT"
	case UserSubject:
		return "USER"
	case StageSubject:
		return "STAGE"
	default:
		return "UNKNOWN"
	}
}

var ErrInvalidPictureSubjectKind = errors.New("invalid picture subject kind")

func (k PictureSubjectKind) IsValid() bool {
	switch k {
	case StandSubject, DevilFruitSubject, UserSubject, StageSubject:
		return true
	default:
		return false
	}
}
