package valueobjects

import (
	"encoding/hex"
	"errors"
)

// ErrInvalidID is returned when a string fails to parse as a canonical id.
var ErrInvalidID = errors.New("invalid id")

// Format renders any [16]byte-backed id in canonical 8-4-4-4-12 hex form.
func Format[T ~[16]byte](id T) string {
	b := [16]byte(id)
	var buf [36]byte
	hex.Encode(buf[0:8], b[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], b[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], b[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], b[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:36], b[10:16])
	return string(buf[:])
}

// Parse accepts either the canonical hyphenated form or the bare 32-hex-digit
// form and returns the corresponding id, wrapping ErrInvalidID on failure.
func Parse[T ~[16]byte](s string) (T, error) {
	var zero T

	var compact string
	switch len(s) {
	case 36:
		if s[8] != '-' || s[13] != '-' || s[18] != '-' || s[23] != '-' {
			return zero, ErrInvalidID
		}
		compact = s[0:8] + s[9:13] + s[14:18] + s[19:23] + s[24:36]
	case 32:
		compact = s
	default:
		return zero, ErrInvalidID
	}

	var b [16]byte
	if _, err := hex.Decode(b[:], []byte(compact)); err != nil {
		return zero, ErrInvalidID
	}
	return T(b), nil
}

// IsNil reports whether id is the zero value.
func IsNil[T ~[16]byte](id T) bool {
	var zero T
	return id == zero
}
