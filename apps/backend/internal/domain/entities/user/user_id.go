package user

import (
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/valueobjects"
)

type UserID [16]byte

// NilUserID is the zero value, used to mean "no id assigned yet".
var NilUserID UserID

func (id UserID) String() string {
	return valueobjects.Format(id)
}

func (id UserID) IsNil() bool {
	return valueobjects.IsNil(id)
}

// ParseUserID parses a canonical (hyphenated or bare-hex) id string.
func ParseUserID(s string) (UserID, error) {
	return valueobjects.Parse[UserID](s)
}
