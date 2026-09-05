package ports

import "errors"

var (
	// ErrStandNotFound is returned when a Stand lookup (by id, by name, or as
	// the EvolvesFrom ancestor of another Stand) finds no matching row.
	ErrStandNotFound = errors.New("stand not found")

	// ErrStandAlreadyExists is returned when saving a Stand would violate the
	// unique name constraint against a different, already-existing Stand.
	ErrStandAlreadyExists = errors.New("stand already exists")

	// ErrDevilFruitNotFound is returned when a DevilFruit lookup (by id or by
	// name) finds no matching row.
	ErrDevilFruitNotFound = errors.New("devil fruit not found")

	// ErrDevilFruitAlreadyExists is returned when saving a DevilFruit would
	// violate the unique name constraint against a different,
	// already-existing DevilFruit.
	ErrDevilFruitAlreadyExists = errors.New("devil fruit already exists")

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

	// ErrTicketInvalid is returned when a stream connection ticket is
	// unknown, expired, already redeemed, or was minted for a different
	// stream/resource. Kept distinct from ErrUnauthenticated so logs and
	// tests can tell "no credential presented" from "a ticket was presented
	// but rejected", even though both map to 401.
	ErrTicketInvalid = errors.New("invalid connection ticket")

	// ErrConstraintViolation is returned when the database rejects a write
	// because it violates a CHECK, NOT NULL, or foreign key constraint - i.e.
	// invalid input that Postgres itself caught, as opposed to a uniqueness
	// conflict (which gets its own, more specific sentinel per entity).
	ErrConstraintViolation = errors.New("data violates a database constraint")

	// ErrRateLimited is returned when a caller has exceeded the request rate
	// allowed for an endpoint.
	ErrRateLimited = errors.New("too many requests")

	// ErrInvalidImage is returned when an uploaded picture's bytes cannot be
	// parsed as an image (corrupt, truncated, or a decompression-bomb sized
	// header) by an IImageProcessor.
	ErrInvalidImage = errors.New("invalid image")

	// ErrGameNotFound is returned when an IGameStore lookup (by GameID or by
	// join code) finds no matching Game.
	ErrGameNotFound = errors.New("game not found")

	// ErrGameCodeTaken is returned when IGameStore.Create is given a join
	// code that already indexes a different Game.
	ErrGameCodeTaken = errors.New("game code already in use")

	// ErrStageNotFound is returned when an IStageRepository lookup by id
	// finds no matching Stage.
	ErrStageNotFound = errors.New("stage not found")

	// ErrStageAlreadyExists is returned when saving a Stage would violate
	// the unique (manga, name) constraint against a different,
	// already-existing Stage.
	ErrStageAlreadyExists = errors.New("stage already exists")

	// ErrRefreshInvalid is returned when a refresh token is unknown,
	// expired, or belongs to a revoked family. Kept distinct from
	// ErrUnauthenticated so logs and tests can tell "no credential
	// presented" from "a refresh token was presented but rejected", even
	// though both map to 401.
	ErrRefreshInvalid = errors.New("invalid refresh token")

	// ErrRefreshReuse is returned when a refresh token is redeemed a second
	// time - proof the token leaked, since single-use redemption means a
	// legitimate client never presents the same token twice. The caller
	// uses this (as opposed to ErrRefreshInvalid) to know the whole token
	// family must be treated as compromised; the HTTP response must not
	// leak the distinction, only server-side logs should.
	ErrRefreshReuse = errors.New("refresh token reuse detected")
)
