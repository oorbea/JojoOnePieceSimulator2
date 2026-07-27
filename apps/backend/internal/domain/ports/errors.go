package ports

import "errors"

var (
	// ErrStandNotFound is returned when a Stand lookup (by name, or as the
	// EvolvesFrom ancestor of another Stand) finds no matching row.
	ErrStandNotFound = errors.New("stand not found")
)
