package ports

import "errors"

var (
	// ErrStandNotFound is returned when a Stand lookup (by id, by name, or as
	// the EvolvesFrom ancestor of another Stand) finds no matching row.
	ErrStandNotFound = errors.New("stand not found")

	// ErrStandAlreadyExists is returned when saving a Stand would violate the
	// unique name constraint against a different, already-existing Stand.
	ErrStandAlreadyExists = errors.New("stand already exists")
)
