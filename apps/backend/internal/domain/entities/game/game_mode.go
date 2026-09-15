package game

import "github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"

// IGameMode is the Strategy driving mode-specific rules: which ballot
// options are on offer, whether Loadouts are reassigned each round, which
// Stage backs a round, and how a resolved round affects the Game. Game
// delegates every mode-specific decision here so it never branches on
// enums.GameModeKind itself.
type IGameMode interface {
	Kind() enums.GameModeKind

	// BallotOptions returns the fixed set of options participants vote
	// between for the current round.
	BallotOptions(g *Game) []OptionID

	// ReassignsEachRound reports whether Loadouts must be redrawn before
	// every round (Versus) or only once at game start (Gauntlet).
	ReassignsEachRound() bool

	// StageFor returns the Stage backing round roundIndex (0-based). rng
	// is used by Versus to pick a random Stage from g.stages; Gauntlet
	// ignores it and indexes g.stages directly (its round order was fixed
	// by Interleave at construction).
	StageFor(g *Game, roundIndex int, rng RandomSource) (Stage, error)

	// ApplyRoundResult folds a tallied round's result into the Game and
	// reports whether the Game is now finished.
	ApplyRoundResult(g *Game, round Round) (finished bool)

	// AutoVotesForAbandoned reports whether an abandoned seat (a human whose
	// disconnect grace period elapsed without a Reconnect - see
	// Participant.MarkAbandoned) should be auto-voted for like a bot. Versus
	// says yes (a team must still be able to reach a majority); Gauntlet
	// says no (an abandoned seat's loadout still counts for the squad, but
	// nothing votes on their behalf - see castBotVotes).
	AutoVotesForAbandoned() bool

	// Outcome computes the final GameResult. Only meaningful once the Game
	// is Finished or Aborted.
	Outcome(g *Game) (GameResult, error)
}
