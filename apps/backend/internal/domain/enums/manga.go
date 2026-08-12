package enums

import "errors"

// Manga discriminates which source material a Game (or a Stage, or a
// player's ability pool) draws from.
type Manga byte

const (
	Jojo Manga = iota
	OnePiece
)

func (m Manga) String() string {
	switch m {
	case Jojo:
		return "JOJO"
	case OnePiece:
		return "ONE_PIECE"
	default:
		return "UNKNOWN"
	}
}

var ErrInvalidManga = errors.New("invalid manga")

func (m Manga) IsValid() bool {
	switch m {
	case Jojo, OnePiece:
		return true
	default:
		return false
	}
}

func ParseManga(str string) (Manga, error) {
	switch str {
	case "JOJO":
		return Jojo, nil
	case "ONE_PIECE":
		return OnePiece, nil
	default:
		return Jojo, ErrInvalidManga
	}
}

// Mangas returns every valid Manga value, in declaration order.
func Mangas() []Manga {
	return []Manga{Jojo, OnePiece}
}
