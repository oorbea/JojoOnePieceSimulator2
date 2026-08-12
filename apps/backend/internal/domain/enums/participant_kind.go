package enums

import "errors"

// ParticipantKind distinguishes a human-controlled participant from a bot
// filling an empty Versus slot.
type ParticipantKind byte

const (
	Human ParticipantKind = iota
	Bot
)

func (k ParticipantKind) String() string {
	switch k {
	case Human:
		return "HUMAN"
	case Bot:
		return "BOT"
	default:
		return "UNKNOWN"
	}
}

var ErrInvalidParticipantKind = errors.New("invalid participant kind")

func (k ParticipantKind) IsValid() bool {
	switch k {
	case Human, Bot:
		return true
	default:
		return false
	}
}

func ParseParticipantKind(str string) (ParticipantKind, error) {
	switch str {
	case "HUMAN":
		return Human, nil
	case "BOT":
		return Bot, nil
	default:
		return Human, ErrInvalidParticipantKind
	}
}
