package game

import "github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"

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
}
