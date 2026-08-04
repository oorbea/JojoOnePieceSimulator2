package dto

import (
	"fmt"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// UpdateProfileRequest is the JSON body accepted by PATCH /users/me. It
// deliberately has exactly one field: email, role, and completeName are
// Google-owned/admin-owned and must never be settable here. decode() rejects
// unknown fields, so sending any of those is a 400, not a silent no-op.
type UpdateProfileRequest struct {
	Username string `json:"username"`
}

// Validate checks Username against the same rule ChangeUsername enforces,
// so a bad value surfaces as a 400 with a clear message instead of a 500
// from the domain constructor.
func (r UpdateProfileRequest) Validate() error {
	if err := user.ValidateUsername(r.Username); err != nil {
		return &ValidationError{Errors: []string{fmt.Sprintf("username: %v", err)}}
	}
	return nil
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
