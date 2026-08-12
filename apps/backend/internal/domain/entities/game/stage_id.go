package game

import (
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/valueobjects"
)

// StageID identifies a Stage (a JoJo part or a One Piece saga) served by
// ports.IStageCatalog.
type StageID [16]byte

// NilStageID is the zero value, used to mean "no id assigned yet".
var NilStageID StageID

func (id StageID) String() string {
	return valueobjects.Format(id)
}

func (id StageID) IsNil() bool {
	return valueobjects.IsNil(id)
}

// ParseStageID parses a canonical (hyphenated or bare-hex) id string.
func ParseStageID(s string) (StageID, error) {
	return valueobjects.Parse[StageID](s)
}
