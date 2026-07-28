package user

import (
	"errors"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

type User struct {
	id           UserID
	googleSub    string
	email        string
	username     string
	completeName string
	role         enums.UserRole
	picture      string
}

func NewUser(
	id UserID,
	googleSub string,
	email string,
	username string,
	completeName string,
	picture string,
	role enums.UserRole,
) (*User, error) {
	if id.IsNil() {
		return nil, errors.New("id is required")
	}
	if googleSub == "" {
		return nil, errors.New("google sub is required")
	}
	if email == "" {
		return nil, errors.New("email is required")
	}
	if username == "" {
		return nil, errors.New("username is required")
	}
	if !role.IsValid() {
		return nil, enums.ErrInvalidUserRole
	}
	return &User{
		id:           id,
		googleSub:    googleSub,
		email:        email,
		username:     username,
		completeName: completeName,
		picture:      picture,
		role:         role,
	}, nil
}

func (u *User) ID() UserID {
	return u.id
}

func (u *User) GoogleSub() string {
	return u.googleSub
}

func (u *User) Email() string {
	return u.email
}

func (u *User) Username() string {
	return u.username
}

func (u *User) CompleteName() string {
	return u.completeName
}

func (u *User) Picture() string {
	return u.picture
}

func (u *User) Role() enums.UserRole {
	return u.role
}

func (u *User) IsAdmin() bool {
	return u.role == enums.Admin
}

// ChangeRole updates the user's role, validating it first. Used to sync the
// ADMIN/USER role against the ADMIN_EMAILS configuration on every login.
func (u *User) ChangeRole(role enums.UserRole) error {
	if !role.IsValid() {
		return enums.ErrInvalidUserRole
	}
	u.role = role
	return nil
}
