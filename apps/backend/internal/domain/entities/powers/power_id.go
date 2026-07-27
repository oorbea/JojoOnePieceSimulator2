package powers

import (
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/valueobjects"
)

// PowerID identifies a Power (and, by embedding, every subtype: Stand,
// DevilFruit, ...) - it mirrors the shared `powers.id` primary key from the
// class-table-inheritance schema. Other entities (Game, User, ...) declare
// their own distinct [16]byte id type the same way.
type PowerID [16]byte

// NilPowerID is the zero value, used to mean "no id assigned yet".
var NilPowerID PowerID

func (id PowerID) String() string {
	return valueobjects.Format(id)
}

func (id PowerID) IsNil() bool {
	return valueobjects.IsNil(id)
}

// ParsePowerID parses a canonical (hyphenated or bare-hex) id string.
func ParsePowerID(s string) (PowerID, error) {
	return valueobjects.Parse[PowerID](s)
}
