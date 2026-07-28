package enums

import "errors"

type UserRole byte

const (
	Regular UserRole = iota
	Admin
)

func (ur UserRole) String() string {
	switch ur {
	case Regular:
		return "REGULAR"
	case Admin:
		return "ADMIN"
	default:
		return "UNKNOWN"
	}
}

var ErrInvalidUserRole = errors.New("invalid user role")

func (ur UserRole) IsValid() bool {
	switch ur {
	case Regular, Admin:
		return true
	default:
		return false
	}
}

func ParseUserRole(str string) (UserRole, error) {
	switch str {
	case "REGULAR":
		return Regular, nil
	case "ADMIN":
		return Admin, nil
	default:
		return Regular, ErrInvalidUserRole
	}
}
