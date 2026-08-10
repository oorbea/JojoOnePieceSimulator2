package game

import (
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/valueobjects"
)

// TeamID identifies a Team within a Game.
type TeamID [16]byte

// NilTeamID is the zero value, used to mean "no id assigned yet".
var NilTeamID TeamID

func (id TeamID) String() string {
	return valueobjects.Format(id)
}

func (id TeamID) IsNil() bool {
	return valueobjects.IsNil(id)
}

// ParseTeamID parses a canonical (hyphenated or bare-hex) id string.
func ParseTeamID(s string) (TeamID, error) {
	return valueobjects.Parse[TeamID](s)
}
