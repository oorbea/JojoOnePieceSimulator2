package user

import (
	"errors"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

type User struct {
	googleSub      string
	email          string
	username       string
	completeName   string
	googlePicture  string
	avatarKey      string
	avatarThumbKey string
	id             UserID
	role           enums.UserRole
	avatarStatus   enums.PictureStatus
	language       enums.Locale
}

// NewUser builds a User from the fields synced from Google. avatar* fields
// always start empty/NONE - a self-uploaded avatar is set afterwards via
// SetAvatarRenditions (mirroring powers.Power's NewPower + SetPictureRenditions
// split), so hydrating a User with an existing avatar from the repository is
// a two-step build just like a Power's picture.
func NewUser(
	id UserID,
	googleSub string,
	email string,
	username string,
	completeName string,
	googlePicture string,
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
		id:            id,
		googleSub:     googleSub,
		email:         email,
		username:      username,
		completeName:  completeName,
		googlePicture: googlePicture,
		role:          role,
		avatarStatus:  enums.PictureNone,
		language:      enums.EnGB,
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

// GooglePicture is the avatar URL synced from the Google account on every
// login. It is never user-editable.
func (u *User) GooglePicture() string {
	return u.googlePicture
}

func (u *User) AvatarKey() string {
	return u.avatarKey
}

func (u *User) AvatarThumbKey() string {
	return u.avatarThumbKey
}

func (u *User) AvatarStatus() enums.PictureStatus {
	return u.avatarStatus
}

func (u *User) Role() enums.UserRole {
	return u.role
}

// Language returns the user's preferred locale, defaulting to en-GB for
// users created before this field existed.
func (u *User) Language() enums.Locale {
	return u.language
}

// ChangeLanguage updates the user's preferred locale, validating it first.
// Also used by the repository to hydrate a User from an already-valid DB
// row.
func (u *User) ChangeLanguage(language enums.Locale) error {
	if !language.IsValid() {
		return enums.ErrInvalidLocale
	}
	u.language = language
	return nil
}

func (u *User) IsAdmin() bool {
	return u.role == enums.Admin
}

// ChangeRole updates the user's role, validating it first. Used to sync the
// ADMIN/USER role against the ADMIN_EMAILS configuration on every login, and
// by an admin changing another user's role.
func (u *User) ChangeRole(role enums.UserRole) error {
	if !role.IsValid() {
		return enums.ErrInvalidUserRole
	}
	u.role = role
	return nil
}

// ChangeUsername validates and updates username. This is the only
// self-service profile mutation besides the avatar - email, role, and
// completeName (Google-owned) are never touched here.
func (u *User) ChangeUsername(username string) error {
	if err := ValidateUsername(username); err != nil {
		return err
	}
	u.username = username
	return nil
}

// SetAvatarRenditions replaces the user-owned avatar's object-storage keys
// together with the pipeline status that produced them, so the three always
// change as one unit - mirrors powers.Power.SetPictureRenditions. Passing
// ("", "", enums.PictureNone) clears the avatar entirely (DELETE
// /users/me/picture), reverting display to GooglePicture.
func (u *User) SetAvatarRenditions(key, thumbKey string, status enums.PictureStatus) {
	u.avatarKey = key
	u.avatarThumbKey = thumbKey
	u.avatarStatus = status
}
