package ports

// IIdGenerator produces new identifiers for a [16]byte-backed id type,
// keeping the concrete generation strategy (UUIDv7, ULID, ...) out of the
// domain layer.
type IIdGenerator[T ~[16]byte] interface {
	NewID() T
}
