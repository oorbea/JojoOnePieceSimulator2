package redis

import (
	"encoding/json"
	"fmt"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// envelope is ports.StreamTicket's JSON wire form. UserID/Role travel as
// their string forms (like jwtClaims does) and come back through the
// existing user.ParseUserID/enums.ParseUserRole functions, the same
// round-trip gamestore/redis's wire.go uses for every id/enum it persists.
type envelope struct {
	UserID   string `json:"userId"`
	Role     string `json:"role"`
	Purpose  string `json:"purpose"`
	Resource string `json:"resource"`
}

func encode(t ports.StreamTicket) ([]byte, error) {
	env := envelope{
		UserID:   t.UserID.String(),
		Role:     t.Role.String(),
		Purpose:  string(t.Purpose),
		Resource: t.Resource,
	}
	data, err := json.Marshal(env)
	if err != nil {
		return nil, fmt.Errorf("marshaling stream ticket: %w", err)
	}
	return data, nil
}

func decode(payload []byte) (ports.StreamTicket, error) {
	var env envelope
	if err := json.Unmarshal(payload, &env); err != nil {
		return ports.StreamTicket{}, fmt.Errorf("unmarshaling stream ticket: %w", err)
	}
	userID, err := user.ParseUserID(env.UserID)
	if err != nil {
		return ports.StreamTicket{}, fmt.Errorf("parsing stream ticket user id: %w", err)
	}
	role, err := enums.ParseUserRole(env.Role)
	if err != nil {
		return ports.StreamTicket{}, fmt.Errorf("parsing stream ticket role: %w", err)
	}
	return ports.StreamTicket{
		UserID:   userID,
		Role:     role,
		Purpose:  ports.TicketPurpose(env.Purpose),
		Resource: env.Resource,
	}, nil
}
