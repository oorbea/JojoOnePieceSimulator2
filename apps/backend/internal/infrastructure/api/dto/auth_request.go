package dto

// GoogleLoginRequest is the JSON body accepted by POST /auth/google.
type GoogleLoginRequest struct {
	IDToken string `json:"idToken"`
}

// Validate checks that an ID token was actually sent.
func (r GoogleLoginRequest) Validate() error {
	if r.IDToken == "" {
		return &ValidationError{Errors: []string{"idToken is required"}}
	}
	return nil
}
