package game

import "github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"

// VersusMode is the 2-team IGameMode: exactly VersusRounds rounds, each
// against a freshly-drawn random Stage and freshly-reassigned Loadouts
// (ReassignsEachRound is true). Every participant may vote for either
// team, including their own - honesty is assumed, there is no self-vote
// restriction. The team with the most round wins takes the match; with an
// odd VersusRounds and a definite (never-tied, since Game always resolves
// ties before recording a Round.Result) winner per round, the match itself
// can never end tied.
type VersusMode struct{}

func (VersusMode) Kind() enums.GameModeKind { return enums.Versus }

func (VersusMode) BallotOptions(g *Game) []OptionID {
	opts := make([]OptionID, len(g.teams))
	for i, t := range g.teams {
		opts[i] = OptionID(t.ID().String())
	}
	return opts
}

func (VersusMode) ReassignsEachRound() bool { return true }

func (VersusMode) StageFor(g *Game, roundIndex int, rng RandomSource) (Stage, error) {
	if len(g.stages) == 0 {
		return Stage{}, ErrNoStagesAvailable
	}
	return g.stages[rng.IntN(len(g.stages))], nil
}

func (VersusMode) ApplyRoundResult(g *Game, round Round) bool {
	return round.Index >= VersusRounds-1
}

func (VersusMode) Outcome(g *Game) (GameResult, error) {
	wins := make(map[OptionID]int, len(g.teams))
	for _, r := range g.rounds {
		if r.Result != nil {
			wins[r.Result.Winner]++
		}
	}
	var winner OptionID
	best := -1
	for _, t := range g.teams {
		opt := OptionID(t.ID().String())
		if w := wins[opt]; w > best {
			best = w
			winner = opt
		}
	}
	return GameResult{
		GameID:       g.id,
		Mode:         enums.Versus,
		Winner:       winner,
		RoundsPlayed: len(g.rounds),
		Aborted:      g.state == enums.Aborted,
	}, nil
}
