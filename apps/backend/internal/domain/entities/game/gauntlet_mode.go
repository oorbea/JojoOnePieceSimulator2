package game

import "github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"

var (
	optionSurvive = OptionID(enums.Survive.String())
	optionFall    = OptionID(enums.Fall.String())
)

// GauntletMode is the cooperative IGameMode: the single team votes
// SURVIVE/FALL each round; a FALL majority ends the run in defeat
// immediately, a SURVIVE majority advances to the next Stage, and clearing
// every Stage is victory. Loadouts are assigned once at game start
// (ReassignsEachRound is false); afterRound is a deliberate no-op hook
// reserved for future run-based progression between rounds (e.g. leveling
// up a surviving squad), so that feature won't need to touch anything else
// in this file.
type GauntletMode struct{}

func (GauntletMode) Kind() enums.GameModeKind { return enums.Gauntlet }

func (GauntletMode) BallotOptions(g *Game) []OptionID {
	return []OptionID{optionSurvive, optionFall}
}

func (GauntletMode) ReassignsEachRound() bool { return false }

func (GauntletMode) StageFor(g *Game, roundIndex int, rng RandomSource) (Stage, error) {
	if roundIndex < 0 || roundIndex >= len(g.stages) {
		return Stage{}, ErrNoStagesAvailable
	}
	return g.stages[roundIndex], nil
}

func (m GauntletMode) ApplyRoundResult(g *Game, round Round) bool {
	m.afterRound(g, round)
	if round.Result.Winner == optionFall {
		return true
	}
	return round.Index >= len(g.stages)-1
}

// afterRound is a seam for future run-based progression between rounds.
// Intentionally empty for now.
func (GauntletMode) afterRound(g *Game, round Round) {}

func (GauntletMode) Outcome(g *Game) (GameResult, error) {
	result := GameResult{
		GameID:       g.id,
		Mode:         enums.Gauntlet,
		RoundsPlayed: len(g.rounds),
		Aborted:      g.state == enums.Aborted,
		Participants: participantOutcomes(g),
	}
	if len(g.rounds) == 0 {
		return result, nil
	}
	last := g.rounds[len(g.rounds)-1]
	result.Winner = optionSurvive
	if last.Result != nil && last.Result.Winner == optionFall {
		result.Winner = optionFall
	}
	return result, nil
}
