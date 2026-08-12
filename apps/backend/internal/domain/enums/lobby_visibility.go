package enums

import "errors"

// LobbyVisibility controls whether a lobby appears in the public browser
// (ports.IGameStore's ListPublic index) or is only reachable by knowing its
// join code.
type LobbyVisibility byte

const (
	Private LobbyVisibility = iota
	Public
)

func (v LobbyVisibility) String() string {
	switch v {
	case Private:
		return "PRIVATE"
	case Public:
		return "PUBLIC"
	default:
		return "UNKNOWN"
	}
}

var ErrInvalidLobbyVisibility = errors.New("invalid lobby visibility")

func (v LobbyVisibility) IsValid() bool {
	switch v {
	case Private, Public:
		return true
	default:
		return false
	}
}

func ParseLobbyVisibility(str string) (LobbyVisibility, error) {
	switch str {
	case "PRIVATE":
		return Private, nil
	case "PUBLIC":
		return Public, nil
	default:
		return Private, ErrInvalidLobbyVisibility
	}
}
