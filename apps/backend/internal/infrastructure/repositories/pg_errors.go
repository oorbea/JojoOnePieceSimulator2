package repositories

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// Postgres SQLSTATE error classes this package translates into domain
// sentinels. See https://www.postgresql.org/docs/current/errcodes-appendix.html
const (
	pgUniqueViolation     = "23505"
	pgForeignKeyViolation = "23503"
	pgNotNullViolation    = "23502"
	pgCheckViolation      = "23514"
)

// wrapPgError translates a Postgres constraint violation into a domain
// sentinel, so every caller ends up 4xx-mappable instead of leaking as a
// generic 500 the first time a constraint we don't already special-case
// fires. alreadyExists is the entity-specific sentinel to use for unique
// violations (e.g. ports.ErrStandAlreadyExists); everything else the
// database can reject on write - CHECK, NOT NULL, and foreign key
// violations - maps to the shared ports.ErrConstraintViolation, since those
// all mean "the input itself was invalid", not "an entity already exists".
// Any error that isn't a recognized *pgconn.PgError passes through
// unchanged.
func wrapPgError(err error, alreadyExists error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}
	switch pgErr.Code {
	case pgUniqueViolation:
		return fmt.Errorf("%w: %s", alreadyExists, pgErr.ConstraintName)
	case pgCheckViolation, pgNotNullViolation, pgForeignKeyViolation:
		return fmt.Errorf("%w: %s", ports.ErrConstraintViolation, pgErr.ConstraintName)
	default:
		return err
	}
}
