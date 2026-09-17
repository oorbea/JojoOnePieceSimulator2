package redis

import (
	"encoding/json"
	"fmt"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// envelope is ports.GameInvite's JSON wire form. GameID/UserID travel as
// their string forms and come back through the existing
// game.ParseGameID/user.ParseUserID functions, the same round-trip
// streamticket/redis's wire.go uses for every id it persists.
type envelope struct {
	GameID    string `json:"gameId"`
	Code      string `json:"code"`
	CreatedBy string `json:"createdBy"`
}

func encode(inv ports.GameInvite) ([]byte, error) {
	env := envelope{
		GameID:    inv.GameID.String(),
		Code:      inv.Code,
		CreatedBy: inv.CreatedBy.String(),
	}
	data, err := json.Marshal(env)
	if err != nil {
		return nil, fmt.Errorf("marshaling game invite: %w", err)
	}
	return data, nil
}

func decode(payload []byte) (ports.GameInvite, error) {
	var env envelope
	if err := json.Unmarshal(payload, &env); err != nil {
		return ports.GameInvite{}, fmt.Errorf("unmarshaling game invite: %w", err)
	}
	gameID, err := game.ParseGameID(env.GameID)
	if err != nil {
		return ports.GameInvite{}, fmt.Errorf("parsing game invite game id: %w", err)
	}
	createdBy, err := user.ParseUserID(env.CreatedBy)
	if err != nil {
		return ports.GameInvite{}, fmt.Errorf("parsing game invite created-by: %w", err)
	}
	return ports.GameInvite{
		GameID:    gameID,
		Code:      env.Code,
		CreatedBy: createdBy,
	}, nil
}
