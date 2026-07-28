package ports

import "errors"

var (
	// ErrStandNotFound is returned when a Stand lookup (by id, by name, or as
	// the EvolvesFrom ancestor of another Stand) finds no matching row.
	ErrStandNotFound = errors.New("stand not found")

	// ErrStandAlreadyExists is returned when saving a Stand would violate the
	// unique name constraint against a different, already-existing Stand.
	ErrStandAlreadyExists = errors.New("stand already exists")

	// ErrUserNotFound is returned when a User lookup (by id, google sub,
	// email, or username) finds no matching row.
	ErrUserNotFound = errors.New("user not found")

	// ErrUserAlreadyExists is returned when saving a User would violate a
	// unique constraint (google sub, email, or username) against a
	// different, already-existing User.
	ErrUserAlreadyExists = errors.New("user already exists")

	// ErrInvalidGoogleToken is returned when a Google ID token fails
	// signature, issuer, audience, or expiry verification.
	ErrInvalidGoogleToken = errors.New("invalid google id token")

	// ErrEmailNotVerified is returned when the Google account's email is not
	// verified, so it cannot be trusted as a stable identity.
	ErrEmailNotVerified = errors.New("google email is not verified")

	// ErrUnauthenticated is returned when a request has no valid credentials.
	ErrUnauthenticated = errors.New("unauthenticated")

	// ErrForbidden is returned when a request is authenticated but lacks the
	// role required for the action.
	ErrForbidden = errors.New("forbidden")

	// ErrConstraintViolation is returned when the database rejects a write
	// because it violates a CHECK, NOT NULL, or foreign key constraint - i.e.
	// invalid input that Postgres itself caught, as opposed to a uniqueness
	// conflict (which gets its own, more specific sentinel per entity).
	ErrConstraintViolation = errors.New("data violates a database constraint")

	// ErrRateLimited is returned when a caller has exceeded the request rate
	// allowed for an endpoint.
	ErrRateLimited = errors.New("too many requests")
)
