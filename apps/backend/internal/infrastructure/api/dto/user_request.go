package dto

import (
	"fmt"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// UpdateProfileRequest is the JSON body accepted by PATCH /users/me.
// Username is mandatory; Language is optional (nil means "leave unchanged").
// email, role, and completeName are Google-owned/admin-owned and must never
// be settable here. decode() rejects unknown fields, so sending any of those
// is a 400, not a silent no-op.
type UpdateProfileRequest struct {
	Username string  `json:"username"`
	Language *string `json:"language,omitempty"`
}

// Validate checks Username against the same rule ChangeUsername enforces
// and, if Language is present, parses it into an enums.Locale, so a bad
// value surfaces as a 400 with a clear message instead of a 500 from the
// domain constructor. The returned bool reports whether Language was
// present at all - the zero enums.Locale (en-GB) is a valid choice, so the
// caller can't tell "unset" from "set to en-GB" any other way.
func (r UpdateProfileRequest) Validate() (language enums.Locale, hasLanguage bool, err error) {
	var errs []string
	if err := user.ValidateUsername(r.Username); err != nil {
		errs = append(errs, fmt.Sprintf("username: %v", err))
	}
	if r.Language != nil {
		hasLanguage = true
		parsed, err := enums.ParseLocale(*r.Language)
		if err != nil {
			errs = append(errs, fmt.Sprintf("language: %v", err))
		} else {
			language = parsed
		}
	}
	if len(errs) > 0 {
		return 0, false, &ValidationError{Errors: errs}
	}
	return language, hasLanguage, nil
}

// AdminUpdateUserRequest is the JSON body accepted by admin-only
// PATCH /users/{id}. Only username is moderatable this way.
type AdminUpdateUserRequest struct {
	Username string `json:"username"`
}

func (r AdminUpdateUserRequest) Validate() error {
	if err := user.ValidateUsername(r.Username); err != nil {
		return &ValidationError{Errors: []string{fmt.Sprintf("username: %v", err)}}
	}
	return nil
}

// UpdateRoleRequest is the JSON body accepted by admin-only
// PATCH /users/{id}/role.
type UpdateRoleRequest struct {
	Role string `json:"role"`
}

// Validate parses Role into an enums.UserRole, collecting a clear message on
// an invalid value.
func (r UpdateRoleRequest) Validate() (enums.UserRole, error) {
	role, err := enums.ParseUserRole(r.Role)
	if err != nil {
		return 0, &ValidationError{Errors: []string{fmt.Sprintf("role: %v", err)}}
	}
	return role, nil
}
