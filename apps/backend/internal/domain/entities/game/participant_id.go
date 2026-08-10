package game

import (
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/valueobjects"
)

// ParticipantID identifies a Participant within a Game. It is distinct from
// user.UserID: a bot Participant has a ParticipantID but no UserID.
type ParticipantID [16]byte

// NilParticipantID is the zero value, used to mean "no id assigned yet".
var NilParticipantID ParticipantID

func (id ParticipantID) String() string {
	return valueobjects.Format(id)
}

func (id ParticipantID) IsNil() bool {
	return valueobjects.IsNil(id)
}

// ParseParticipantID parses a canonical (hyphenated or bare-hex) id string.
func ParseParticipantID(s string) (ParticipantID, error) {
	return valueobjects.Parse[ParticipantID](s)
}
