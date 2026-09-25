package dto

import "regexp"

// GoogleLoginRequest is the JSON body accepted by POST /auth/google.
type GoogleLoginRequest struct {
	IDToken string `json:"idToken"`
}

// Validate checks that an ID token was actually sent.
func (r GoogleLoginRequest) Validate() error {
	if r.IDToken == "" {
		return &ValidationError{Errors: []FieldError{{Field: "idToken", Code: ValRequired, Message: "idToken is required"}}}
	}
	return nil
}

// devLoginNamePattern bounds DevLoginRequest.Name the same way a Google
// account's derived username is sanitized (see AuthService.uniqueUsername) -
// lowercase ASCII, digits and underscore only, so it always yields a valid
// username and a safe <name>@dev.invalid local part.
var devLoginNamePattern = regexp.MustCompile(`^[a-z0-9_]{1,24}$`)

// DevLoginRequest is the JSON body accepted by POST /auth/dev-login - only
// mounted when DEV_AUTH_BYPASS is set, see endpoints/local_only.go.
type DevLoginRequest struct {
	Name  string `json:"name"`
	Admin bool   `json:"admin"`
}

// Validate checks that Name is non-empty and matches devLoginNamePattern.
func (r DevLoginRequest) Validate() error {
	if r.Name == "" {
		return &ValidationError{Errors: []FieldError{{Field: "name", Code: ValRequired, Message: "name is required"}}}
	}
	if !devLoginNamePattern.MatchString(r.Name) {
		return &ValidationError{Errors: []FieldError{{Field: "name", Code: ValInvalidValue, Message: "name must match ^[a-z0-9_]{1,24}$"}}}
	}
	return nil
}
