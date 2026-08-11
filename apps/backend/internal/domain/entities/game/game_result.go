package game

import (
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// ParticipantOutcome snapshots one Participant's identity at the moment a
// Game finished/aborted, so ports.IGameHistory can record who played
// without holding a live reference into the Game (which is deleted from
// the store right after). UserID is nil for a bot.
type ParticipantOutcome struct {
	ParticipantID ParticipantID
	UserID        *user.UserID
	DisplayName   string
	TeamID        TeamID
	Bot           bool
}

// GameResult is the final outcome of a finished or aborted Game, handed to
// ports.IGameHistory.Record by the application layer.
type GameResult struct {
	GameID GameID
	Mode   enums.GameModeKind
	// Winner is enums.Survive/enums.Fall's string form for Gauntlet, or the
	// winning TeamID's string form for Versus. Empty if Aborted before any
	// round resolved.
	Winner       OptionID
	RoundsPlayed int
	Aborted      bool
	// Participants is every seat in the Game at the moment it ended, in
	// join order - added so IGameHistory can answer "what did I play",
	// not just "what happened".
	Participants []ParticipantOutcome
}

// participantOutcomes builds every ParticipantOutcome for g, in join order.
func participantOutcomes(g *Game) []ParticipantOutcome {
	participants := g.Participants()
	out := make([]ParticipantOutcome, 0, len(participants))
	for _, p := range participants {
		out = append(out, ParticipantOutcome{
			ParticipantID: p.ID(),
			UserID:        p.UserID(),
			DisplayName:   p.DisplayName(),
			TeamID:        p.TeamID(),
			Bot:           p.IsBot(),
		})
	}
	return out
}
