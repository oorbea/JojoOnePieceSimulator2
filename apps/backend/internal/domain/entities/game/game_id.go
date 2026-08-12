package game

import (
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/valueobjects"
)

// GameID identifies a Game aggregate.
type GameID [16]byte

// NilGameID is the zero value, used to mean "no id assigned yet".
var NilGameID GameID

func (id GameID) String() string {
	return valueobjects.Format(id)
}

func (id GameID) IsNil() bool {
	return valueobjects.IsNil(id)
}

// ParseGameID parses a canonical (hyphenated or bare-hex) id string.
func ParseGameID(s string) (GameID, error) {
	return valueobjects.Parse[GameID](s)
}
