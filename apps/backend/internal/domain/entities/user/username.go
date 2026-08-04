package user

import (
	"errors"
	"regexp"
	"strings"
)

const (
	// MinUsernameLen and MaxUsernameLen bound a username's length, checked
	// by ValidateUsername (self-service PATCH) as well as by the auto-derived
	// candidates services.AuthService generates at registration.
	MinUsernameLen = 3
	MaxUsernameLen = 32
)

// invalidUsernameChars matches anything outside the allowed alphabet, shared
// by SanitizeUsername (auto-derivation from an email's local part) and
// ValidateUsername (self-service PATCH /users/me), so the two can never
// drift apart on what counts as a valid username.
var invalidUsernameChars = regexp.MustCompile(`[^a-z0-9_]`)

// ErrInvalidUsername is returned when a username fails ValidateUsername.
var ErrInvalidUsername = errors.New("username must be 3-32 characters of lowercase letters, digits, or underscores")

// SanitizeUsername lowercases raw and replaces every disallowed character
// with '_'. It does not enforce length - callers deriving a username from
// arbitrary input (e.g. an email's local part) are expected to pad or
// truncate separately.
func SanitizeUsername(raw string) string {
	return invalidUsernameChars.ReplaceAllString(strings.ToLower(raw), "_")
}

// ValidateUsername reports whether username is acceptable as-is: the right
// length and made up only of lowercase letters, digits, and underscores. It
// never mutates or sanitizes - a user submitting a username through
// ChangeUsername must get it exactly right, unlike the auto-derived
// candidates SanitizeUsername produces at registration.
func ValidateUsername(username string) error {
	if len(username) < MinUsernameLen || len(username) > MaxUsernameLen {
		return ErrInvalidUsername
	}
	if invalidUsernameChars.MatchString(username) {
		return ErrInvalidUsername
	}
	return nil
}
