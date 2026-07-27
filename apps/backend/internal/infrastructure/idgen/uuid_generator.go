// Package idgen provides infrastructure implementations of
// ports.IIdGenerator. This is the only place identifier generation touches
// google/uuid - the domain layer only ever sees the resulting [16]byte id.
package idgen

import (
	"github.com/google/uuid"
)

// UUIDGenerator generates time-ordered UUIDv7 values for any [16]byte-backed
// id type T.
type UUIDGenerator[T ~[16]byte] struct{}

func (UUIDGenerator[T]) NewID() T {
	return T(uuid.Must(uuid.NewV7()))
}
