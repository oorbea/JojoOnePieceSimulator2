package enums

import "errors"

// InviteStatus is the public answer to "is this lobby share link still
// good" - see services.GameService.InviteStatus's doc for exactly what it
// does and does not check. Deliberately just two members: a link is either
// usable or it isn't, and nothing about why is exposed to an
// unauthenticated caller.
type InviteStatus byte

const (
	InviteValid InviteStatus = iota
	InviteExpired
)

func (s InviteStatus) String() string {
	switch s {
	case InviteValid:
		return "VALID"
	case InviteExpired:
		return "EXPIRED"
	default:
		return "UNKNOWN"
	}
}

var ErrInvalidInviteStatus = errors.New("invalid invite status")

func (s InviteStatus) IsValid() bool {
	switch s {
	case InviteValid, InviteExpired:
		return true
	default:
		return false
	}
}

func ParseInviteStatus(str string) (InviteStatus, error) {
	switch str {
	case "VALID":
		return InviteValid, nil
	case "EXPIRED":
		return InviteExpired, nil
	default:
		return InviteExpired, ErrInvalidInviteStatus
	}
}
