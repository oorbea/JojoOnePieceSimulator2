package characters

import (
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/valueobjects"
)

// CharacterID identifies a Character (and, by embedding, either subtype:
// JojoCharacter, OnePieceCharacter) - it mirrors the shared `characters.id`
// primary key from the class-table-inheritance schema. Deliberately its own
// type, not powers.PowerID: a Character is not a Power (see Character's
// doc), so the two id spaces must never be interchangeable at compile time.
type CharacterID [16]byte

// NilCharacterID is the zero value, used to mean "no id assigned yet".
var NilCharacterID CharacterID

func (id CharacterID) String() string {
	return valueobjects.Format(id)
}

func (id CharacterID) IsNil() bool {
	return valueobjects.IsNil(id)
}

// ParseCharacterID parses a canonical (hyphenated or bare-hex) id string.
func ParseCharacterID(s string) (CharacterID, error) {
	return valueobjects.Parse[CharacterID](s)
}
