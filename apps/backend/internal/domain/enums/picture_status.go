package enums

import "errors"

// PictureStatus tracks where a Power's picture is in the async compression
// pipeline: NONE (never uploaded), PENDING (uploaded, worker hasn't finished
// yet), READY (compressed renditions are in object storage), or FAILED (the
// worker could not produce renditions - the previous renditions, if any,
// are left untouched).
type PictureStatus byte

const (
	PictureNone PictureStatus = iota
	PicturePending
	PictureReady
	PictureFailed
)

func (s PictureStatus) String() string {
	switch s {
	case PictureNone:
		return "NONE"
	case PicturePending:
		return "PENDING"
	case PictureReady:
		return "READY"
	case PictureFailed:
		return "FAILED"
	default:
		return "UNKNOWN"
	}
}

var ErrInvalidPictureStatus = errors.New("invalid picture status")

func (s PictureStatus) IsValid() bool {
	switch s {
	case PictureNone, PicturePending, PictureReady, PictureFailed:
		return true
	default:
		return false
	}
}

func ParsePictureStatus(str string) (PictureStatus, error) {
	switch str {
	case "NONE":
		return PictureNone, nil
	case "PENDING":
		return PicturePending, nil
	case "READY":
		return PictureReady, nil
	case "FAILED":
		return PictureFailed, nil
	default:
		return PictureNone, ErrInvalidPictureStatus
	}
}
